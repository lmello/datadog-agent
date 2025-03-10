// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessage(t *testing.T) {

	message := Message{Content: []byte("hello")}
	assert.Equal(t, "hello", string(message.Content))

	message.Content = []byte("world")
	assert.Equal(t, "world", string(message.Content))
	assert.Equal(t, StatusInfo, message.GetStatus())

}

func TestGetHostnameLambda(t *testing.T) {
	message := Message{
		ServerlessExtra: ServerlessExtra{
			Lambda: &Lambda{
				ARN: "testHostName",
			},
		},
	}
	assert.Equal(t, "testHostName", message.GetHostname())
}

func TestGetHostname(t *testing.T) {
	t.Setenv("DD_HOSTNAME", "testHostnameFromEnvVar")
	message := Message{Content: []byte("hello")}
	assert.Equal(t, "testHostnameFromEnvVar", message.GetHostname())
}
