// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-FileCopyrightText: 2026 The Roster Authors
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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gotunix.net/roster/internal/email"
	"gotunix.net/roster/internal/store"
	"gotunix.net/roster/internal/ui"
)

var (
	exportVars     []string
	exportOutput   string
	exportExclude  []string
	exportEmail    string
	exportGroup    string
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export inventory data to CSV",
	Long:  `Aggregate data from multiple inventories and export to a CSV file. Specific host variables can be included as columns.`,
	Run: func(_ *cobra.Command, args []string) {
		// Clean exportVars: trim spaces and remove empty strings
		var cleanVars []string
		for _, v := range exportVars {
			v = strings.TrimSpace(v)
			if v != "" {
				cleanVars = append(cleanVars, v)
			}
		}
		exportVars = cleanVars

		csvBuffer := bytes.NewBuffer(nil)
		totalHosts := 0

		for _, dir := range inventoryPaths {
			inv, err := store.LoadInventory(dir)
			if err != nil {
				fmt.Fprintln(os.Stderr, ui.ErrorMsg("Skipping %s: %v", dir, err))
				continue
			}

			data, numHosts, err := store.ExportInventory(inv, dir, exportVars, exportExclude, exportGroup)
			if err != nil {
				fmt.Fprintln(os.Stderr, ui.ErrorMsg("Exporting %s: %v", dir, err))
				continue
			}

			csvBuffer.Write(data)
			totalHosts += numHosts
		}

		// Handle file output
		if exportOutput != "" {
			if err := os.WriteFile(exportOutput, csvBuffer.Bytes(), 0644); err != nil {
				fmt.Fprintln(os.Stderr, ui.ErrorMsg("Creating output file: %v", err))
				return
			}
			fmt.Fprintln(os.Stderr, ui.SuccessMsg("Exported %d hosts to %s", totalHosts, exportOutput))
		}

		// Handle email output
		if exportEmail != "" {
			fmt.Fprintln(os.Stderr, ui.BoldStyle.Foreground(ui.Cyan).Render("📧 Sending email to "+exportEmail+"..."))

			subject := "Roster Inventory Export"
			body := fmt.Sprintf("Please find attached the inventory export from Roster.\n\n"+
				"Inventories included: %s\n"+
				"Total hosts: %d", strings.Join(inventoryPaths, ", "), totalHosts)

			filename := "inventory_export.csv"
			if exportOutput != "" {
				filename = filepath.Base(exportOutput)
			}

			smtpCfg := email.LoadSMTPConfig()
			if smtpCfg.Host == "" {
				h, p, f, u, pw := store.LoadSMTPConfig()
				smtpCfg = email.SMTPConfig{Host: h, Port: p, From: f, User: u, Pass: pw}
			}
			if err := email.SendCSV(exportEmail, subject, body, filename, csvBuffer.Bytes(), smtpCfg); err != nil {
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
	exportCmd.Flags().StringVarP(&exportGroup, "group", "g", "", "Only include hosts belonging to this group")
	rootCmd.AddCommand(exportCmd)
}
