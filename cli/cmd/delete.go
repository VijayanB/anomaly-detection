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
	handler "esad/internal/handler/ad"
	"fmt"
	"github.com/spf13/cobra"
)

const commandDelete = "delete"

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   commandDelete + " [flags] [list of detectors]",
	Short: "Deletes detectors",
	Long:  `Deletes detectors based on pattern, use "" to make sure the name is not matched with pwd lists'`,
	Run: func(cmd *cobra.Command, args []string) {
		fstatus, _ := cmd.Flags().GetBool("force")
		err := deleteDetectors(args, fstatus)
		if err != nil {
			fmt.Println(commandDelete, "command failed")
			fmt.Println("Reason:", err)
		}
	},
}

func init() {
	esadCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().BoolP("force", "f", false, "Force deletion even if it is running")
}

func deleteDetectors(detectors []string, fstatus bool) error {
	commandHandler, err := getCommandHandler()
	if err != nil {
		return err
	}
	for _, detector := range detectors {
		err = handler.DeleteAnomalyDetector(commandHandler, detector, fstatus)
		if err != nil {
			return err
		}
	}
	return nil
}