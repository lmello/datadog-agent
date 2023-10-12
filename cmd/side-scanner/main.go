package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	// DataDog agent: config stuffs
	"github.com/DataDog/datadog-agent/cmd/agent/common"
	commonpath "github.com/DataDog/datadog-agent/cmd/agent/common/path"
	"github.com/DataDog/datadog-agent/cmd/internal/runcmd"
	pkgconfig "github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/util/flavor"
	"github.com/DataDog/datadog-agent/pkg/version"

	// DataDog agent: SBOM + proto stuffs
	sbommodel "github.com/DataDog/agent-payload/v5/sbom"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	// DataDog agent: RC stuffs
	"github.com/DataDog/datadog-agent/pkg/config/remote"
	_ "github.com/DataDog/datadog-agent/pkg/config/remote"
	remoteconfig "github.com/DataDog/datadog-agent/pkg/config/remote/service"
	"github.com/DataDog/datadog-agent/pkg/remoteconfig/state"

	// DataDog agent: logs stuffs
	"github.com/DataDog/datadog-agent/pkg/epforwarder"
	"github.com/DataDog/datadog-agent/pkg/logs/message"

	// Trivy stuffs
	"github.com/aquasecurity/trivy-db/pkg/db"
	"github.com/aquasecurity/trivy/pkg/detector/ospkg"
	"github.com/aquasecurity/trivy/pkg/fanal/applier"
	"github.com/aquasecurity/trivy/pkg/fanal/artifact"
	"github.com/aquasecurity/trivy/pkg/fanal/artifact/vm"
	ftypes "github.com/aquasecurity/trivy/pkg/fanal/types"
	"github.com/aquasecurity/trivy/pkg/sbom/cyclonedx"
	"github.com/aquasecurity/trivy/pkg/scanner"
	"github.com/aquasecurity/trivy/pkg/scanner/local"
	"github.com/aquasecurity/trivy/pkg/types"
	"github.com/aquasecurity/trivy/pkg/vulnerability"

	"github.com/spf13/cobra"
)

var (
	globalParams struct {
		ConfigFilePath string
	}
)

func main() {
	flavor.SetFlavor(flavor.SideScannerAgent)
	os.Exit(runcmd.Run(rootCommand()))
}

func rootCommand() *cobra.Command {
	sideScannerCmd := &cobra.Command{
		Use:          "side-scanner [command]",
		Short:        "Datadog Side Scanner at your service.",
		Long:         `Datadog Side Scanner scans your cloud environment for vulnerabilities, compliance and security issues.`,
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			pkgconfig.Datadog.AddConfigPath(commonpath.DefaultConfPath)
			_, err := pkgconfig.Load()
			return err
		},
	}

	sideScannerCmd.AddCommand(runCommand())

	return sideScannerCmd
}

func runCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Runs the side-scanner",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(context.Background())
		},
	}
}

func run(ctx context.Context) error {
	fmt.Println(pkgconfig.Datadog.Get("hostname"))
	common.SetupInternalProfiling(pkgconfig.Datadog, "")
	configService, err := remoteconfig.NewService()
	if err != nil {
		return fmt.Errorf("unable to create remote-config service: %w", err)
	}

	if err := configService.Start(ctx); err != nil {
		return fmt.Errorf("unable to start remote-config service: %w", err)
	}

	rcClient, err := remote.NewClient("side-scanner", configService, version.AgentVersion, nil, time.Second*5)
	if err != nil {
		return fmt.Errorf("unable to create local remote-config client: %w", err)
	}

	scanner := newSideScanner(rcClient)
	scanner.start(ctx)
	return nil
}

type task struct {
	Type  string            `json:"type"`
	Scans []json.RawMessage `json:"scans"`
}

type ebsScan struct {
	Region     string `json:"region"`
	SnapshotID string `json:"snapshotId"`
	VolumeID   string `json:"volumeId"`
	Hostname   string `json:"hostname"`
}

type lambdaScan struct {
	Region       string `json:"region"`
	FunctionName string `json:"function_name"`
}

type sideScanner struct {
	tasks          <-chan []byte
	eventForwarder epforwarder.EventPlatformForwarder
}

func newSideScanner(rcClient *remote.Client) *sideScanner {
	eventForwarder := epforwarder.NewEventPlatformForwarder()
	tasks := make(chan []byte)
	rcClient.Start()
	rcClient.Subscribe(state.ProductDebug, func(update map[string]state.RawConfig, _ func(string, state.ApplyStatus)) {
		fmt.Println("update", update)
		for _, cfg := range update {
			tasks <- cfg.Config
		}
	})
	return &sideScanner{
		tasks:          tasks,
		eventForwarder: eventForwarder,
	}
}

func (s *sideScanner) start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case rawTask := <-s.tasks:
			if err := s.handleTask(ctx, rawTask); err != nil {
				fmt.Printf("%s\n", err)
			}
		}
	}
}

func (s *sideScanner) handleTask(ctx context.Context, rawTask []byte) error {
	var task task
	if err := json.Unmarshal(rawTask, &task); err != nil {
		return fmt.Errorf("could not parse side-scanner task: %w", err)
	}

	var err error
	for _, rawScan := range task.Scans {
		switch task.Type {
		case "ebs-scan":
			var scan ebsScan
			if err := json.Unmarshal(rawScan, &scan); err != nil {
				return err
			}
			err = s.processEBS(ctx, &scan)
		case "lambda-scan":
			var scan lambdaScan
			if err := json.Unmarshal(rawScan, &scan); err != nil {
				return err
			}
			err = s.processLambda(ctx, &scan)
		default:
			return fmt.Errorf("unknown scan type: %s", task.Type)
		}
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "oops: %#v: %v\n", task, err)
	}

	return nil
}

func (s *sideScanner) processEBS(ctx context.Context, task *ebsScan) error {
	region := task.Region
	target := "ebs:" + task.SnapshotID

	artifactOptions := artifact.Option{
		Offline:           true,
		NoProgress:        true,
		DisabledAnalyzers: nil,
		Slow:              true,
		SBOMSources:       []string{},
		DisabledHandlers:  []ftypes.HandlerType{ftypes.UnpackagedPostHandler},
		OnlyDirs:          []string{"etc", "var/lib/dpkg", "var/lib/rpm", "lib/apk"},
		AWSRegion:         region,
	}

	trivyCache := newMemoryCache()
	trivyVMArtifact, err := vm.NewArtifact(target, trivyCache, artifactOptions)
	if err != nil {
		return fmt.Errorf("unable to create artifact from image, err: %w", err)
	}

	startedAt := time.Now()
	trivyDetector := ospkg.Detector{}
	trivyVulnClient := vulnerability.NewClient(db.Config{})
	trivyApplier := applier.NewApplier(trivyCache)
	trivyLocalScanner := local.NewScanner(trivyApplier, trivyDetector, trivyVulnClient)
	trivyScanner := scanner.NewScanner(trivyLocalScanner, trivyVMArtifact)
	fmt.Println("starting scanning")
	trivyReport, err := trivyScanner.ScanArtifact(ctx, types.ScanOptions{
		VulnType:            []string{},
		SecurityChecks:      []string{},
		ScanRemovedPackages: false,
		ListAllPackages:     true,
	})
	fmt.Println("scanning done", trivyReport, err)
	if err != nil {
		return fmt.Errorf("unable to marshal report to sbom format, err: %w", err)
	}

	sourceAgent := "sidescanner"
	envVarEnv := pkgconfig.Datadog.GetString("env")
	createdAt := time.Now()
	duration := time.Since(startedAt)
	marshaler := cyclonedx.NewMarshaler("")
	bom, err := marshaler.Marshal(trivyReport)
	if err != nil {
		return err
	}

	entity := &sbommodel.SBOMEntity{
		Status:             sbommodel.SBOMStatus_SUCCESS, //  sbommodel.SBOMSourceType_EBS
		Type:               sbommodel.SBOMSourceType_HOST_FILE_SYSTEM,
		Id:                 task.Hostname,
		InUse:              true,
		GeneratedAt:        timestamppb.New(createdAt),
		GenerationDuration: convertDuration(duration),
		Hash:               "pierrot",
		Sbom: &sbommodel.SBOMEntity_Cyclonedx{
			Cyclonedx: convertBOM(bom),
		},
	}

	rawEvent, err := proto.Marshal(&sbommodel.SBOMPayload{
		Version:  1,
		Source:   &sourceAgent,
		Entities: []*sbommodel.SBOMEntity{entity},
		DdEnv:    &envVarEnv,
	})
	if err != nil {
		return fmt.Errorf("unable to proto marhsal sbom: %w", err)
	}

	m := &message.Message{Content: rawEvent}
	return s.eventForwarder.SendEventPlatformEvent(m, epforwarder.EventTypeContainerSBOM)
}

func (s *sideScanner) processLambda(ctx context.Context, task *lambdaScan) error {
	panic("TODO")
}
