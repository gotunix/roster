// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-FileCopyrightText: 2026 The MetaBoard authors
// =============================================================================================== //
//                                                                                                 //
//                   /$$$$$$$                        /$$                                           //
//                  | $$__  $$                      | $$                                           //
//                  | $$  \ $$  /$$$$$$   /$$$$$$$ /$$$$$$    /$$$$$$   /$$$$$$                    //
//                  | $$$$$$$/ /$$__  $$ /$$_____/|_  $$_/   /$$__  $$ /$$__  $$                   //
//                  | $$__  $$| $$  \ $$|  $$$$$$   | $$    | $$$$$$$$| $$  \__/                   //
//                  | $$  \ $$| $$  | $$ \____  $$  | $$ /$$| $$_____/| $$                         //
//                  | $$  | $$|  $$$$$$/ /$$$$$$$/  |  $$$$/|  $$$$$$$| $$                         //
//                  |__/  |__/ \______/ |_______/    \___/   \_______/|__/                         //
//                                                                                                 //
// =============================================================================================== //
//              This program is free software: you can redistribute it and/or modify               //
//              it under the terms of the GNU General Public License as published by               //
//              the Free Software Foundation, either version 3 of the License, or                  //
//              (at your option) any later version.                                                //
//                                                                                                 //
//              This program is distributed in the hope that it will be useful,                    //
//              but WITHOUT ANY WARRANTY; without even the implied warranty of                     //
//              MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the                      //
//              GNU General Public License for more details.                                       //
//                                                                                                 //
//              You should have received a copy of the GNU General Public License                  //
//              along with this program.  If not, see <https://www.gnu.org/licenses/>.             //
// =============================================================================================== //

package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"gotunix.net/roster/internal/email"
	"gotunix.net/roster/internal/store"
	"gotunix.net/roster/internal/ui"
)

var (
	exportVars    []string
	exportOutput  string
	exportExclude []string
	exportEmail   string
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export inventory data to CSV",
	Long: `Aggregate data from multiple inventories and export to a CSV file. Specific host variables can be included as columns.`,
	Run: func(cmd *cobra.Command, args []string) {
		var allRows [][]string

		// Clean exportVars: trim spaces and remove empty strings
		var cleanVars []string
		for _, v := range exportVars {
			v = strings.TrimSpace(v)
			if v != "" {
				cleanVars = append(cleanVars, v)
			}
		}
		exportVars = cleanVars

		// Header
		header := []string{"Inventory", "Host", "Groups"}
		header = append(header, exportVars...)
		allRows = append(allRows, header)

		// Prepare exclusion map for fast lookup
		excludeMap := make(map[string]bool)
		for _, g := range exportExclude {
			excludeMap[g] = true
		}

		for _, dir := range inventoryPaths {
			inv, err := store.LoadInventory(dir)
			if err != nil {
				fmt.Fprintln(os.Stderr, ui.ErrorMsg("Skipping %s: %v", dir, err))
				continue
			}

			// Sort hosts for consistent output
			var hostNames []string
			for name := range inv.Hosts {
				hostNames = append(hostNames, name)
			}
			sort.Strings(hostNames)

			for _, hName := range hostNames {
				h := inv.Hosts[hName]

				// Find groups and check for exclusion
				var groups []string
				excluded := false
				for gName, g := range inv.Groups {
					for _, member := range g.Hosts {
						if member == hName {
							if excludeMap[gName] {
								excluded = true
							}
							groups = append(groups, gName)
							break
						}
					}
				}

				if excluded {
					continue
				}

				sort.Strings(groups)

				row := []string{dir, hName, strings.Join(groups, ", ")}

				// Add requested vars
				for _, vName := range exportVars {
					val := ""
					if h.Vars != nil {
						if v, ok := h.Vars[vName]; ok {
							val = fmt.Sprintf("%v", v)
						}
					}
					row = append(row, val)
				}
				allRows = append(allRows, row)
			}
		}

		// Generate CSV in memory
		var csvBuffer bytes.Buffer
		writer := csv.NewWriter(&csvBuffer)
		if err := writer.WriteAll(allRows); err != nil {
			fmt.Fprintln(os.Stderr, ui.ErrorMsg("Writing CSV: %v", err))
			return
		}

		// Handle file output
		if exportOutput != "" {
			if err := os.WriteFile(exportOutput, csvBuffer.Bytes(), 0644); err != nil {
				fmt.Fprintln(os.Stderr, ui.ErrorMsg("Creating output file: %v", err))
				return
			}
			fmt.Fprintln(os.Stderr, ui.SuccessMsg("Exported %d hosts to %s", len(allRows)-1, exportOutput))
		}

		// Handle email output
		if exportEmail != "" {
			fmt.Fprintln(os.Stderr, ui.BoldStyle.Foreground(ui.Cyan).Render("📧 Sending email to "+exportEmail+"..."))

			subject := "Roster Inventory Export"
			body := fmt.Sprintf("Please find attached the inventory export from Roster.\n\n"+
				"Inventories included: %s\n"+
				"Total hosts: %d", strings.Join(inventoryPaths, ", "), len(allRows)-1)

			filename := "inventory_export.csv"
			if exportOutput != "" {
				filename = filepath.Base(exportOutput)
			}

			if err := email.SendCSV(exportEmail, subject, body, filename, csvBuffer.Bytes()); err != nil {
				fmt.Fprintln(os.Stderr, ui.ErrorMsg("Failed to send email: %v", err))
			} else {
				fmt.Fprintln(os.Stderr, ui.SuccessMsg("Email sent successfully to %s", exportEmail))
			}
		}

		// Handle stdout if no file and no email
		if exportOutput == "" && exportEmail == "" {
			fmt.Print(csvBuffer.String())
		}
	},
}

func init() {
	exportCmd.Flags().StringSliceVar(&exportVars, "vars", []string{}, "Host variables to include as columns")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output CSV file path (default: stdout)")
	exportCmd.Flags().StringSliceVarP(&exportExclude, "exclude", "e", []string{}, "Exclude hosts belonging to these groups")
	exportCmd.Flags().StringVar(&exportEmail, "email", "", "Email the CSV export to this address")
	rootCmd.AddCommand(exportCmd)
}
