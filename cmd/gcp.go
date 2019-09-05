/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

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
package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/reoim/pingcloud-cli/ping"
	"github.com/reoim/pingcloud-cli/ping/gcp"
	"github.com/spf13/cobra"
)

// gcpCmd represents the gcp command
var gcpCmd = &cobra.Command{
	Use:   "gcp",
	Short: "Ping GCP regions",
	Long:  `Ping GCP regions and print median latency.`,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("gcp called")

		// Init tabwriter
		tr := tabwriter.NewWriter(os.Stdout, 40, 8, 2, '\t', 0)
		fmt.Fprintf(tr, "Region\tCity, State, Country\tLatency")
		fmt.Fprintln(tr)
		fmt.Fprintf(tr, "------------------------------\t------------------------------\t------------------------------")
		fmt.Fprintln(tr)
		// Flush tabwriter
		tr.Flush()

		for r, i := range gcp.GCPEndpoints {
			e := ping.Endpoints{
				Region:  r,
				Name:    gcp.GCPEndpointsName[r],
				Address: i,
			}
			// e.TestPrint()
			e.Ping()
			// fmt.Println(out.Name+" "+out.Region+" "+out.Address+"Latency: ", out.Latency)
			//fmt.Println(output.Latency)
		}
	},
}

func init() {
	rootCmd.AddCommand(gcpCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// gcpCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// gcpCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
