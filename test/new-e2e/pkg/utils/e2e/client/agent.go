// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package client

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/runner"
	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/runner/parameters"
	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/utils/e2e/client/agentclientparams"
	"github.com/DataDog/test-infra-definitions/common/utils"
	"github.com/DataDog/test-infra-definitions/components/datadog/agent"
	e2eOs "github.com/DataDog/test-infra-definitions/components/os"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

var _ stackInitializer = (*Agent)(nil)

// An Agent that is connected to an [agent.Installer].
//
// [agent.Installer]: https://pkg.go.dev/github.com/DataDog/test-infra-definitions@main/components/datadog/agent#Installer
type Agent struct {
	deserializer utils.RemoteServiceDeserializer[agent.ClientData]
	os           e2eOs.OS
	*AgentCommandRunner
	vmClient           *VMClient
	shouldWaitForReady bool
}

// NewAgent creates a new instance of an Agent connected to an [agent.Installer].
//
// [agent.Installer]: https://pkg.go.dev/github.com/DataDog/test-infra-definitions@main/components/datadog/agent#Installer
func NewAgent(installer *agent.Installer, agentClientOptions ...agentclientparams.Option) *Agent {
	agentClientParams := agentclientparams.NewParams(agentClientOptions...)
	agentInstance := &Agent{
		os:                 installer.VM().GetOS(),
		shouldWaitForReady: agentClientParams.ShouldWaitForReady,
		deserializer:       installer,
	}
	return agentInstance
}

//lint:ignore U1000 Ignore unused function as this function is called using reflection
func (agent *Agent) setStack(t *testing.T, stackResult auto.UpResult) error {
	clientData, err := agent.deserializer.Deserialize(stackResult)
	if err != nil {
		return err
	}

	var privateSSHKey []byte

	privateKeyPath, err := runner.GetProfile().ParamStore().GetWithDefault(parameters.PrivateKeyPath, "")
	if err != nil {
		return err
	}

	if privateKeyPath != "" {
		privateSSHKey, err = os.ReadFile(privateKeyPath)
		if err != nil {
			return err
		}
	}

	agent.vmClient, err = newVMClient(t, privateSSHKey, &clientData.Connection, agent.os)
	if err != nil {
		return err
	}
	agent.AgentCommandRunner = newAgentCommandRunner(t, agent.executeAgentCmdWithError)
	if !agent.shouldWaitForReady {
		return nil
	}
	return agent.waitForReadyTimeout(1 * time.Minute)
}

func (agent *Agent) executeAgentCmdWithError(arguments []string) (string, error) {
	parameters := ""
	if len(arguments) > 0 {
		parameters = `"` + strings.Join(arguments, `" "`) + `"`
	}
	cmd := agent.os.GetRunAgentCmd(parameters)
	return agent.vmClient.ExecuteWithError(cmd)
}
