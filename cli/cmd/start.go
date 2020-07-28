/*
 * Copyright 2020 Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * Licensed under the Apache License, Version 2.0 (the "License").
 * You may not use this file except in compliance with the License.
 * A copy of the License is located at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * or in the "license" file accompanying this file. This file is distributed
 * on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
 * express or implied. See the License for the specific language governing
 * permissions and limitations under the License.
 */

package cmd

import (
	"esad/internal/client"
	"esad/internal/handler/ad"
	"fmt"
	"github.com/spf13/cobra"
)

const commandStart = "start"

// createCmd represents the create command
var startCmd = &cobra.Command{
	Use:   commandStart + " [flags] [list of detectors]",
	Short: "Start detectors",
	Long:  `Start detectors based on pattern, use "" to make sure the name is not matched with pwd lists'`,
	Run: func(cmd *cobra.Command, args []string) {
		idStatus, _ := cmd.Flags().GetBool("id")
		action := ad.StartAnomalyDetector
		if idStatus {
			action = ad.StartAnomalyDetectorByID
		}
		err := execute(action, args)
		if err != nil {
			fmt.Println(commandStop, "command failed")
			fmt.Println("Reason:", err)
		}
	},
}

func init() {
	esadCmd.AddCommand(startCmd)
	startCmd.Flags().BoolP("name", "", true, "Input is name or pattern")
	startCmd.Flags().BoolP("id", "", false, "Input is id")
}

func execute(f func(*ad.Handler, string) error, detectors []string) error {
	// iterate over the arguments
	// the first return value is index of fileNames, we can omit it using _
	h, err := getCommandHandler()
	if err != nil {
		return err
	}
	for _, detector := range detectors {
		err := f(h, detector)
		if err != nil {
			return err
		}
	}
	return nil
}

func getCommandHandler() (*ad.Handler, error) {
	c, err := client.NewClient(nil)
	if err != nil {
		return nil, err
	}
	u, err := getUserProfile()
	if err != nil {
		return nil, err
	}
	h := GetHandler(c, u)
	return h, nil
}