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

package statictest

import (
	"fmt"
	"os"
	"strconv"

	"github.com/commonjava/indy-tests/pkg/common"
	static "github.com/commonjava/indy-tests/pkg/statictest"

	"github.com/spf13/cobra"
)

// example: http://orchhost/pnc-rest/v2/builds/97241/logs/build
var originalIndy, staticIndy, foloTrackId string
var processNum int

const DEFAULT_PROCESS_NUM = 1
const DEFAULT_REPO_REPL_PATTERN = ""
const DEFAULT_BUILD_TYPE = "maven"

func NewStaticTestCmd() *cobra.Command {

	exec := &cobra.Command{
		Use:   "static-test",
		Short: "To test a static proxy server with a indy tracking download entry",
		Run: func(cmd *cobra.Command, args []string) {
			if !validate() {
				cmd.Help()
				os.Exit(1)
			}
			// here will use env variables if they are specified for some flags
			checkEnvVars()
			static.Run(originalIndy, foloTrackId, staticIndy, processNum)
		},
	}

	exec.Flags().StringVarP(&originalIndy, "originalIndy", "o", "", "The original indy server to get the folo tracking Id.")
	exec.Flags().StringVarP(&staticIndy, "staticIndy", "s", "", "The static indy server to do the testing.")
	exec.Flags().StringVarP(&foloTrackId, "floloTrackId", "f", "", "The folo tracking id in the original indy server to get download entries.")
	exec.Flags().IntVarP(&processNum, "processNum", "p", DEFAULT_PROCESS_NUM, "The number of processes to download files in parralel.")

	exec.MarkFlagRequired("originalIndy")
	exec.MarkFlagRequired("staticIndy")
	exec.MarkFlagRequired("floloTrackId")

	return exec
}

func validate() bool {
	if common.IsEmptyString(foloTrackId) {
		fmt.Printf("$folo_track_id cannot be empty!\n\n")
		return false
	}
	if common.IsEmptyString(originalIndy) {
		originalIndy = os.Getenv("INDY_TARGET")
	}

	return true
}

func checkEnvVars() {
	envProcNum := os.Getenv("BUILD_PROC_NUM")
	if num, err := strconv.Atoi(envProcNum); err == nil {
		processNum = num
	}
}
