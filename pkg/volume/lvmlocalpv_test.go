/*
Copyright 2020-2021 The OpenEBS Authors

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
package volume

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/openebs/lvm-localpv/pkg/generated/clientset/internalclientset/fake"
	fakelvm "github.com/openebs/lvm-localpv/pkg/generated/clientset/internalclientset/typed/lvm/v1alpha1/fake"
	"github.com/openebs/openebsctl/pkg/client"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	k8stest "k8s.io/client-go/testing"
)

func TestGetLVMLocalPV(t *testing.T) {
	type args struct {
		c           *client.K8sClient
		lvmReactors func(*client.K8sClient)
		pvList      *corev1.PersistentVolumeList
		openebsNS   string
	}
	tests := []struct {
		name    string
		args    args
		want    []metav1.TableRow
		wantErr bool
	}{
		{
			name: "no lvm volumes present",
			args: args{
				c: &client.K8sClient{
					Ns:        "random-namespace",
					LVMCS:     fake.NewSimpleClientset(),
					K8sCS:     k8sfake.NewSimpleClientset(),
					OpenebsCS: nil,
				},
				pvList:      &corev1.PersistentVolumeList{Items: []corev1.PersistentVolume{jivaPV1, pv2, pv3}},
				lvmReactors: lvmVolNotExists,
				openebsNS:   "openebs",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "only one lvm volume present",
			args: args{
				c: &client.K8sClient{
					Ns:    "lvmlocalpv",
					K8sCS: k8sfake.NewSimpleClientset(&localpvCSICtrlSTS),
					LVMCS: fake.NewSimpleClientset(&lvmVol1),
				},
				pvList:    &corev1.PersistentVolumeList{Items: []corev1.PersistentVolume{jivaPV1, lvmPV1}},
				openebsNS: "lvmlocalpv",
			},
			wantErr: false,
			want: []metav1.TableRow{
				{
					Cells: []interface{}{"lvmlocalpv", "pvc-1", "Ready", "1.9.0", "4.0GiB", "lvm-sc-1", corev1.VolumeBound, corev1.ReadWriteOnce, "node1"},
				},
			},
		},
		{
			name: "only one lvm volume presentm with lvmvol absent",
			args: args{
				c: &client.K8sClient{
					Ns:    "lvmlocalpv",
					K8sCS: k8sfake.NewSimpleClientset(&localpvCSICtrlSTS),
					LVMCS: fake.NewSimpleClientset(),
				},
				pvList:    &corev1.PersistentVolumeList{Items: []corev1.PersistentVolume{lvmPV1}},
				openebsNS: "lvmlocalpv",
			},
			wantErr: false,
			want:    nil,
		},
		{
			name: "only one lvm volume present, namespace conflicts",
			args: args{
				c: &client.K8sClient{
					Ns:    "jiva",
					K8sCS: k8sfake.NewSimpleClientset(&localpvCSICtrlSTS),
					LVMCS: fake.NewSimpleClientset(&lvmVol1),
				},
				pvList:    &corev1.PersistentVolumeList{Items: []corev1.PersistentVolume{jivaPV1, lvmPV1}},
				openebsNS: "lvmlocalpvXYZ",
			},
			wantErr: false,
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Before func
			if tt.args.lvmReactors != nil {
				tt.args.lvmReactors(tt.args.c)
			}
			// 2. Call the code under test
			got, err := GetLVMLocalPV(tt.args.c, tt.args.pvList, tt.args.openebsNS)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLVMLocalPV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// 3. Test for TC pass/fail & display
			gotLen := len(got)
			expectedLen := len(tt.want)
			if gotLen != expectedLen {
				t.Errorf("GetLVMLocalPV() returned %d elements, wanted %d elements", gotLen, expectedLen)
			}
			for i, gotLine := range got {
				if len(gotLine.Cells) != len(tt.want[i].Cells) {
					t.Errorf("Line#%d in output had %d elements, wanted %d elements", i+1, len(gotLine.Cells), len(tt.want[i].Cells))
				}
				if !reflect.DeepEqual(tt.want[i].Cells, gotLine.Cells) {
					t.Errorf("GetLVMLocalPV() line#%d got = %v, want %v", i+1, got, tt.want)
				}
			}
		})
	}
}

func TestDescribeLVMLocalPVs(t *testing.T) {
	type args struct {
		c       *client.K8sClient
		lvmfunc func(sClient *client.K8sClient)
		vol     *corev1.PersistentVolume
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"no lvm volume present",
			args{c: &client.K8sClient{Ns: "lvm", K8sCS: k8sfake.NewSimpleClientset(), LVMCS: fake.NewSimpleClientset()},
				vol:     nil,
				lvmfunc: lvmVolNotExists,
			},
			true,
		},
		{"one lvm volume present and asked for and lvm-controller absent",
			args{c: &client.K8sClient{Ns: "lvmlocalpv", LVMCS: fake.NewSimpleClientset(&lvmVol1), K8sCS: k8sfake.NewSimpleClientset()},
				vol: &lvmPV1},
			false,
		},
		{"one lvm volume present and asked for and lvm-controller present",
			args{c: &client.K8sClient{Ns: "lvmlocalpv",
				K8sCS: k8sfake.NewSimpleClientset(&localpvCSICtrlSTS),
				LVMCS: fake.NewSimpleClientset(&lvmVol1)},
				vol: &lvmPV1},
			false,
		},
		{"one lvm volume present and asked for but namespace wrong",
			args{c: &client.K8sClient{Ns: "lvmlocalpv", LVMCS: fake.NewSimpleClientset(&lvmVol1)},
				vol: &lvmPV1, lvmfunc: lvmVolNotExists},
			true,
		},
		{"one lvm volume present and some other volume asked for",
			args{c: &client.K8sClient{Ns: "lvm", K8sCS: k8sfake.NewSimpleClientset(&lvmPV1), LVMCS: fake.NewSimpleClientset(&lvmVol1)},
				vol:     &cstorPV2,
				lvmfunc: lvmVolNotExists},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.lvmfunc != nil {
				tt.args.lvmfunc(tt.args.c)
			}
			if err := DescribeLVMLocalPVs(tt.args.c, tt.args.vol); (err != nil) != tt.wantErr {
				t.Errorf("DescribeLVMLocalPVs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// lvmVolNotExists makes fakelvmClientSet return error
func lvmVolNotExists(c *client.K8sClient) {
	// NOTE: Set the VERB & Resource correctly & make it work for single resources
	c.LVMCS.LocalV1alpha1().(*fakelvm.FakeLocalV1alpha1).Fake.PrependReactor("*", "*", func(action k8stest.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, fmt.Errorf("failed to list LVMVolumes")
	})
}
