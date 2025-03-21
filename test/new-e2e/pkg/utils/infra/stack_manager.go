// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// Package infra implements utilities to interact with a Pulumi infrastructure
package infra

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/runner"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/debug"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optremove"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	nameSep               = "-"
	e2eWorkspaceDirectory = "dd-e2e-workspace"

	stackUpTimeout      = 60 * time.Minute
	stackDestroyTimeout = 60 * time.Minute
	stackDeleteTimeout  = 20 * time.Minute
)

var (
	stackManager     *StackManager
	initStackManager sync.Once
)

// StackManager handles
type StackManager struct {
	stacks map[string]*auto.Stack
	lock   sync.RWMutex
}

// GetStackManager returns a stack manager, initialising on first call
func GetStackManager() *StackManager {
	initStackManager.Do(func() {
		var err error

		stackManager, err = newStackManager()
		if err != nil {
			panic(fmt.Sprintf("Got an error during StackManager singleton init, err: %v", err))
		}
	})

	return stackManager
}

func newStackManager() (*StackManager, error) {
	return &StackManager{
		stacks: make(map[string]*auto.Stack),
	}, nil
}

// GetStack creates or return a stack based on stack name and config, if error occurs during stack creation it destroy all the resources created
func (sm *StackManager) GetStack(ctx context.Context, name string, config runner.ConfigMap, deployFunc pulumi.RunFunc, failOnMissing bool) (*auto.Stack, auto.UpResult, error) {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	stack, upResult, err := sm.getStack(ctx, name, config, deployFunc, failOnMissing)

	if err != nil {
		errDestroy := sm.deleteStack(ctx, name, stack)
		if errDestroy != nil {
			return stack, upResult, errors.Join(err, errDestroy)
		}
	}

	return stack, upResult, err
}

// GetStackNoDeleteOnFailure creates or return a stack based on stack name and config, if error occurs during stack creation, it will not destroy the created resources. Using this can lead to resource leaks.
func (sm *StackManager) GetStackNoDeleteOnFailure(ctx context.Context, name string, config runner.ConfigMap, deployFunc pulumi.RunFunc, failOnMissing bool) (*auto.Stack, auto.UpResult, error) {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	return sm.getStack(ctx, name, config, deployFunc, failOnMissing)
}

// DeleteStack safely deletes a stack
func (sm *StackManager) DeleteStack(ctx context.Context, name string) error {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	stack, ok := sm.stacks[name]
	if !ok {
		// Build configuration from profile
		profile := runner.GetProfile()
		stackName := buildStackName(profile.NamePrefix(), name)
		workspace, err := buildWorkspace(ctx, profile, stackName, func(ctx *pulumi.Context) error { return nil })
		if err != nil {
			return err
		}

		newStack, err := auto.SelectStack(ctx, stackName, workspace)
		if err != nil {
			return err
		}

		stack = &newStack
	}

	return sm.deleteStack(ctx, name, stack)
}

// ForceRemoveStackConfiguration removes the configuration files pulumi creates for managing a stack.
// It DOES NOT perform any cleanup of the resources created by the stack. Call `DeleteStack` for correct cleanup.
func (sm *StackManager) ForceRemoveStackConfiguration(ctx context.Context, name string) error {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	stack, ok := sm.stacks[name]
	if !ok {
		return fmt.Errorf("unable to remove stack %s: stack not present", name)
	}

	deleteContext, cancel := context.WithTimeout(ctx, stackDeleteTimeout)
	defer cancel()
	return stack.Workspace().RemoveStack(deleteContext, stack.Name(), optremove.Force())
}

// Cleanup delete any existing stack
func (sm *StackManager) Cleanup(ctx context.Context) []error {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	var errors []error

	for stackID, stack := range sm.stacks {
		err := sm.deleteStack(ctx, stackID, stack)
		if err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

func (sm *StackManager) deleteStack(ctx context.Context, stackID string, stack *auto.Stack) error {
	if stack == nil {
		return fmt.Errorf("unable to find stack, skipping deletion of: %s", stackID)
	}

	destroyContext, cancel := context.WithTimeout(ctx, stackDestroyTimeout)
	_, err := stack.Destroy(destroyContext, optdestroy.ProgressStreams(os.Stdout))
	cancel()
	if err != nil {
		return err
	}

	deleteContext, cancel := context.WithTimeout(ctx, stackDeleteTimeout)
	defer cancel()
	err = stack.Workspace().RemoveStack(deleteContext, stack.Name())
	return err
}

func (sm *StackManager) getStack(ctx context.Context, name string, config runner.ConfigMap, deployFunc pulumi.RunFunc, failOnMissing bool) (*auto.Stack, auto.UpResult, error) {
	// Build configuration from profile
	profile := runner.GetProfile()
	stackName := buildStackName(profile.NamePrefix(), name)
	deployFunc = runFuncWithRecover(deployFunc)

	// Inject common/managed parameters
	cm, err := runner.BuildStackParameters(profile, config)
	if err != nil {
		return nil, auto.UpResult{}, err
	}

	stack := sm.stacks[name]
	if stack == nil {
		workspace, err := buildWorkspace(ctx, profile, stackName, deployFunc)
		if err != nil {
			return nil, auto.UpResult{}, err
		}

		newStack, err := auto.SelectStack(ctx, stackName, workspace)
		if auto.IsSelectStack404Error(err) && !failOnMissing {
			newStack, err = auto.NewStack(ctx, stackName, workspace)
		}
		if err != nil {
			return nil, auto.UpResult{}, err
		}

		stack = &newStack
		sm.stacks[name] = stack
	} else {
		stack.Workspace().SetProgram(deployFunc)
	}

	err = stack.SetAllConfig(ctx, cm.ToPulumi())
	if err != nil {
		return nil, auto.UpResult{}, err
	}

	upCtx, cancel := context.WithTimeout(ctx, stackUpTimeout)
	var loglevel uint = 1
	defer cancel()
	upResult, err := stack.Up(upCtx, optup.ProgressStreams(os.Stderr), optup.DebugLogging(debug.LoggingOptions{
		LogToStdErr:   true,
		FlowToPlugins: true,
		LogLevel:      &loglevel,
	}))

	return stack, upResult, err

}

func buildWorkspace(ctx context.Context, profile runner.Profile, stackName string, runFunc pulumi.RunFunc) (auto.Workspace, error) {
	project := workspace.Project{
		Name:           tokens.PackageName(profile.ProjectName()),
		Runtime:        workspace.NewProjectRuntimeInfo("go", nil),
		Description:    pulumi.StringRef("E2E Test inline project"),
		StackConfigDir: stackName,
		Config: map[string]workspace.ProjectConfigType{
			// We should always disable default providers
			// Disabling all known except AWS due to https://github.com/pulumi/pulumi-eks/pull/886
			"pulumi:disable-default-providers": {
				Value: []string{"kubernetes", "azure-native", "awsx", "eks"},
			},
			// Required in CI due to https://github.com/pulumi/pulumi-eks/pull/886
			"aws:skipMetadataApiCheck": {
				Value: "false",
			},
		},
	}

	return auto.NewLocalWorkspace(ctx, auto.Project(project), auto.Program(runFunc), auto.WorkDir(profile.RootWorkspacePath()))
}

func buildStackName(namePrefix, stackName string) string {
	stackName = namePrefix + nameSep + stackName
	return strings.ToLower(strings.ReplaceAll(stackName, "_", "-"))
}

func runFuncWithRecover(f pulumi.RunFunc) pulumi.RunFunc {
	return func(ctx *pulumi.Context) (err error) {
		defer func() {
			if r := recover(); r != nil {
				stackDump := make([]byte, 4096)
				stackSize := runtime.Stack(stackDump, false)
				err = fmt.Errorf("panic in run function, stack:\n %s\n\nerror: %v", stackDump[:stackSize], r)
			}
		}()

		return f(ctx)
	}
}
