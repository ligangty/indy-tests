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

package common

import (
	"testing"
)

func TestIsRegularFile(t *testing.T) {
	type args struct {
		fileLoc string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// Add test cases.
		{name: "meta", args: args{fileLoc: "/tmp/maven-metadata.xml"}, want: false},
		{name: "pom", args: args{fileLoc: "/tmp/pom.xml"}, want: true},
		{name: "jar", args: args{fileLoc: "/tmp/foo.jar"}, want: true},
		{name: "tgz", args: args{fileLoc: "/tmp/foo.tgz"}, want: true},
		{name: "md5", args: args{fileLoc: "/tmp/foo.jar.md5"}, want: false},
		{name: "sha1", args: args{fileLoc: "/tmp/foo.jar.sha1"}, want: false},
		{name: "othertgz", args: args{fileLoc: "/tmp/othertgz"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRegularFile(tt.args.fileLoc); got != tt.want {
				t.Errorf("IsRegularFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAlterUploadPath(t *testing.T) {
	type args struct {
		rawPath          string
		storeKey         string
		newReleaseNumber string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{name: "maven", args: args{
			rawPath:          "/org/apache/kafka/connect-api/2.7.0.redhat-00012/connect-api-2.7.0.redhat-00012-javadoc.jar",
			storeKey:         "maven:hosted:build-101385",
			newReleaseNumber: "94465"},
			want: "/org/apache/kafka/connect-api/2.7.0.redhat-94465/connect-api-2.7.0.redhat-94465-javadoc.jar"},
		{name: "npm1", args: args{
			rawPath: "/@redhat/opossum/-/opossum-6.2.1.tgz", storeKey: "npm:hosted:build-AMRM", newReleaseNumber: "94465"},
			want: "/@redhat/opossum/-/opossum-6.2.1-94465.tgz"},
		{name: "npm2", args: args{
			rawPath: "/@redhat/kogito-tooling-backend/-/kogito-tooling-backend-0.9.0-3.tgz", storeKey: "npm:hosted:build-ALNI", newReleaseNumber: "94465"},
			want: "/@redhat/kogito-tooling-backend/-/kogito-tooling-backend-0.9.0-94465.tgz"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AlterUploadPath(tt.args.rawPath, tt.args.storeKey, tt.args.newReleaseNumber); got != tt.want {
				t.Errorf("AlterUploadPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
