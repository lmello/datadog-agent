// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package client

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

type stackInitializer interface {
	setStack(t *testing.T, stackResult auto.UpResult) error
}

// CheckEnvStructValid validates an environment struct
func CheckEnvStructValid[Env any]() error {
	var env Env
	_, err := getFields(&env)
	return err
}

// CallStackInitializers validate an environment struct and initialise a stack
func CallStackInitializers[Env any](t *testing.T, env *Env, upResult auto.UpResult) error {
	fields, err := getFields(env)

	for _, field := range fields {
		initializer := field.stackInitializer
		if reflect.TypeOf(initializer).Kind() == reflect.Ptr && reflect.ValueOf(initializer).IsNil() {
			return fmt.Errorf("the field %v of %v is nil", field.name, reflect.TypeOf(env))
		}

		if err = initializer.setStack(t, upResult); err != nil {
			return err
		}
	}

	return err
}

type field struct {
	stackInitializer stackInitializer
	name             string
}

func getFields[Env any](env *Env) ([]field, error) {
	var fields []field
	envValue := reflect.ValueOf(*env)
	envType := reflect.TypeOf(*env)
	exportedFields := make(map[string]struct{})

	for _, f := range reflect.VisibleFields(envType) {
		if f.IsExported() {
			exportedFields[f.Name] = struct{}{}
		}
	}

	stackInitializerType := reflect.TypeOf((*stackInitializer)(nil)).Elem()
	for i := 0; i < envValue.NumField(); i++ {
		fieldName := envValue.Type().Field(i).Name
		if _, found := exportedFields[fieldName]; !found {
			return nil, fmt.Errorf("the field %v in %v is not exported", fieldName, envType)
		}

		initializer, ok := envValue.Field(i).Interface().(stackInitializer)
		if !ok {
			return nil, fmt.Errorf("%v contains %v which doesn't implement %v",
				envType,
				fieldName,
				stackInitializerType,
			)
		}
		fields = append(fields, field{
			stackInitializer: initializer,
			name:             fieldName,
		})
	}
	return fields, nil
}
