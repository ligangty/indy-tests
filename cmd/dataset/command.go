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
	"fmt"
	"os"

	"github.com/commonjava/indy-tests/pkg/dataset"
	"github.com/spf13/cobra"
)

func NewDatasetCmd() *cobra.Command {

	exec := &cobra.Command{
		Use:     "dataset $pncBaseUrl $indyBaseUrl $buildId",
		Short:   "To generate test dataset from any PNC successful group build.",
		Example: "dataset https://orch-stage.xyz.com http://indy-admin-stage.xyz.com 2836",
		Run: func(cmd *cobra.Command, args []string) {
			if !validate(args) {
				cmd.Help()
				os.Exit(1)
			}
			groupBuild, _ := cmd.Flags().GetBool("groupBuild")
			dataset.Run(args[0], args[1], args[2], groupBuild)
		},
	}

	exec.Flags().BoolP("groupBuild", "g", false, "Is group build.")
	return exec
}

func validate(args []string) bool {
	if len(args) < 3 {
		fmt.Printf("there are 3 mandatory arguments: pncBaseUrl, indyBaseUrl, buildId!\n\n")
		return false
	}
	return true
}
