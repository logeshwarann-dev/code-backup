/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1

// ExecActionApplyConfiguration represents a declarative configuration of the ExecAction type for use
// with apply.
type ExecActionApplyConfiguration struct {
	Command []string `json:"command,omitempty"`
}

// ExecActionApplyConfiguration constructs a declarative configuration of the ExecAction type for use with
// apply.
func ExecAction() *ExecActionApplyConfiguration {
	return &ExecActionApplyConfiguration{}
}

// WithCommand adds the given value to the Command field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Command field.
func (b *ExecActionApplyConfiguration) WithCommand(values ...string) *ExecActionApplyConfiguration {
	for i := range values {
		b.Command = append(b.Command, values[i])
	}
	return b
}
