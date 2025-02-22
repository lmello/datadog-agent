// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package tcp

import "bytes"

// prefixer prepends a prefix to a message.
type prefixer struct {
	prefix string
	buffer bytes.Buffer
}

// newPrefixer returns a prefixer that prepends the given prefix to a message.
func newPrefixer(prefix string) *prefixer {
	return &prefixer{
		prefix: prefix,
	}
}

// apply prepends the prefix to the message.
func (p *prefixer) apply(content []byte) []byte {
	p.buffer.Reset()
	p.buffer.WriteString(p.prefix)
	p.buffer.Write(content)
	return p.buffer.Bytes()
}
