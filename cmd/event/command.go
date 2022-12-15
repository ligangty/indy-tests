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

package event

import (
	"fmt"
	"os"
	"strconv"

	"github.com/commonjava/indy-tests/pkg/common"
	event "github.com/commonjava/indy-tests/pkg/event"

	"github.com/spf13/cobra"
)

// example: http://orchhost/pnc-rest/v2/builds/97241/logs/build
var targetIndy, repoReplPattern, buildType string
var processNum int
var doRunEnablement bool

const DEFAULT_PROCESS_NUM = 1
const DEFAULT_REPO_REPL_PATTERN = ""
const DEFAULT_BUILD_TYPE = "maven"

func NewEventTestCmd() *cobra.Command {

	exec := &cobra.Command{
		Use:   "event $indy_url $folo_track_id",
		Short: "To do an event test with an existed folo tracking report",
		Run: func(cmd *cobra.Command, args []string) {
			if !validate(args) {
				cmd.Help()
				os.Exit(1)
			}
			// here will use env variables if they are specified for some flags
			checkEnvVars()
			indyURL := args[0]
			foloTrackId := args[1]
			if common.IsEmptyString(targetIndy) {
				fmt.Printf("!!targetIndy is not specified, will use the same one as the $indy_url: %s\n", indyURL)
				targetIndy = indyURL
			}
			doRunEnablement, _ := cmd.Flags().GetBool("doRunEnablement")
			fmt.Printf("doRunEnablement: %t\n", doRunEnablement)
			event.Run(indyURL, foloTrackId, targetIndy, buildType, processNum, doRunEnablement)
		},
	}

	exec.Flags().StringVarP(&targetIndy, "targetIndy", "t", "", "The target indy server to do the testing. Will get from this flag or from env variables 'INDY_TARGET' if flag is not specified. If both are not specified, will use $indy_url.")
	exec.Flags().StringVarP(&buildType, "buildType", "b", DEFAULT_BUILD_TYPE, "The type of the build, should be 'maven' or 'npm'. Default is 'maven'.")
	exec.Flags().IntVarP(&processNum, "processNum", "p", DEFAULT_PROCESS_NUM, "The number of processes to download and upload files in parralel.")
	exec.Flags().BoolP("doRunEnablement", "e", true, "Decide whether to run store enablement validation or not.")
	return exec
}

func validate(args []string) bool {
	if len(args) <= 1 {
		fmt.Printf("indy_url or folo_track_id is not specified!\n\n")
		return false
	}
	if common.IsEmptyString(args[0]) {
		fmt.Printf("$indy_url cannot be empty!\n\n")
		return false
	}
	if common.IsEmptyString(args[1]) {
		fmt.Printf("$folo_track_id cannot be empty!\n\n")
		return false
	}
	if common.IsEmptyString(targetIndy) {
		targetIndy = os.Getenv("INDY_TARGET")
	}

	return true
}

func checkEnvVars() {
	envBuildType := os.Getenv("INDY_BUILD_TYPE")
	if envBuildType == "maven" || envBuildType == "npm" {
		buildType = envBuildType
	}
	envProcNum := os.Getenv("BUILD_PROC_NUM")
	if num, err := strconv.Atoi(envProcNum); err == nil {
		processNum = num
	}
}
