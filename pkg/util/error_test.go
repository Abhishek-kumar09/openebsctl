/*
Copyright 2020 The OpenEBS Authors

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

package util

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
)

func TestCheckErr(t *testing.T) {
	type args struct {
		err       error
		handleErr func(string)
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"Error not nil",
			args{
				err: errors.New("Some error occurred"),
				handleErr: func(s string) {
					fmt.Println("Handled")
				},
			},
		},
		{
			"Error nil",
			args{
				err: nil,
				handleErr: func(s string) {
					fmt.Println("Handled")
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CheckErr(tt.args.err, tt.args.handleErr)
		})
	}
}

func TestHandleEmptyTableError(t *testing.T) {
	type args struct {
		resource string
		ns       string
		casType  string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"No Namespace and cas",
			args{
				resource: "ResourceType",
				ns:       "",
				casType:  "",
			},
		},
		{
			"Wrong cas or Namespace",
			args{
				resource: "ResourceType",
				ns:       "InValidNamespace",
				casType:  "jiva",
			},
		},
		{
			"Wrong cas type in all namespace",
			args{
				resource: "ResourceType",
				ns:       "",
				casType:  "invalid",
			},
		},
		{
			"Wrong Namespace and all cas types",
			args{
				resource: "ResourceType",
				ns:       "InValidNamespace",
				casType:  "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandleEmptyTableError(tt.args.resource, tt.args.ns, tt.args.casType)
		})
	}
}
