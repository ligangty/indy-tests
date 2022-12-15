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
	"testing"
)

func Test_getBuildQueue(t *testing.T) {
	type args struct {
		edges []Edge
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{name: "test", args: args{edges}, want: "ok"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getBuildQueue(tt.args.edges)
			size := len(got.Builds)
			if len(got.Builds) != 5 {
				t.Errorf("getBuildQueue(), size: %d, want: %d", size, 5)
			}
			var total int
			for _, b := range got.Builds {
				total += len(b.Items)
			}
			if total != 23 {
				t.Errorf("getBuildQueue(), total: %d, want: %d", total, 23)
			}
		})
	}
}

var edges = []Edge{
	{
		Source: "90440",
		Target: "90447",
	},
	{
		Source: "90439",
		Target: "90446",
	},
	{
		Source: "90440",
		Target: "90439",
	},
	{
		Source: "90450",
		Target: "90440",
	},
	{
		Source: "90452",
		Target: "90446",
	},
	{
		Source: "90451",
		Target: "90452",
	},
	{
		Source: "90455",
		Target: "90451",
	},
	{
		Source: "90444",
		Target: "90446",
	},
	{
		Source: "90441",
		Target: "90444",
	},
	{
		Source: "90441",
		Target: "90439",
	},
	{
		Source: "90453",
		Target: "90441",
	},
	{
		Source: "90434",
		Target: "90444",
	},
	{
		Source: "90453",
		Target: "90434",
	},
	{
		Source: "90453",
		Target: "90439",
	},
	{
		Source: "90450",
		Target: "90453",
	},
	{
		Source: "90445",
		Target: "90448",
	},
	{
		Source: "90450",
		Target: "90445",
	},
	{
		Source: "90442",
		Target: "90443",
	},
	{
		Source: "90442",
		Target: "90437",
	},
	{
		Source: "90436",
		Target: "90442",
	},
	{
		Source: "90436",
		Target: "90452",
	},
	{
		Source: "90436",
		Target: "90438",
	},
	{
		Source: "90438",
		Target: "90452",
	},
	{
		Source: "90454",
		Target: "90433",
	},
	{
		Source: "90455",
		Target: "90439",
	},
	{
		Source: "90435",
		Target: "90446",
	},
	{
		Source: "90454",
		Target: "90435",
	},
	{
		Source: "90454",
		Target: "90439",
	},
	{
		Source: "90436",
		Target: "90439",
	},
	{
		Source: "90436",
		Target: "90449",
	},
}
