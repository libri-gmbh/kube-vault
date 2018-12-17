// Copyright © 2018 Alexander Pinnecke <alexander.pinnecke@googlemail.com>
//

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// renewCmd represents the renew command
var renewCmd = &cobra.Command{
	Use:   "renew",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here
		fmt.Println("renew called")
	},
}

func init() {
	RootCmd.AddCommand(renewCmd)
}
