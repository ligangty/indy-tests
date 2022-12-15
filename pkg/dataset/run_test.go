/*
 *  Copyright (C) 2021-2023 Red Hat, Inc.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *          http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package dataset

import (
	"reflect"
	"testing"
)

func Test_getMetadataPaths(t *testing.T) {
	type args struct {
		alignLog string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		// TODO: Add test cases.
		{
			name: "normal",
			args: args{alignLog: "REST Client returned: {org.sonatype.oss:oss-parent:9=9.0.0.redhat-2, junit:junit:4.13.1=4.13.1}"},
			want: []string{"org/sonatype/oss/oss-parent/maven-metadata.xml", "junit/junit/maven-metadata.xml"},
		},
		{
			name: "empty",
			args: args{alignLog: "REST Client returned for project versions: {}"},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMavenMetadataPaths(tt.args.alignLog); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getMetadataPaths() = %v, want %v", got, tt.want)
			}
		})
	}
}
