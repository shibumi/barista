// Copyright 2018 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package multicast provides a method to convert any bar.Module into one that
// can be added to the bar multiple times. When combined with group.Simple, this
// allows multiple combinations of modules without creating extra instances.
package multicast

import (
	"sync"

	"github.com/shibumi/barista/bar"
	"github.com/shibumi/barista/base/value"
	"github.com/shibumi/barista/core"
	"github.com/shibumi/barista/sink"
)

type module struct {
	*value.Value
	start func() // called on Stream(), to ensure backing module is started
}

// Stream starts the module, and tries to stream the original module as well.
func (m module) Stream(sink bar.Sink) {
	go m.start()
	for {
		next := m.Next()
		s, _ := m.Get().(bar.Segments)
		sink.Output(s)
		<-next
	}
}

// New creates a multicast module that can be added to the bar any number of
// times, and mirrors the output of the original module at each location.
// IMPORTANT: The original module must not be added to the bar.
func New(original bar.Module) bar.Module {
	output, sink := sink.Value()
	coreModule := core.NewModule(original)
	var once sync.Once
	start := func() {
		once.Do(func() {
			coreModule.Stream(sink)
		})
	}
	return module{output, start}
}
