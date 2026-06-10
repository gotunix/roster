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

package email

import (
	"encoding/base64"
	"fmt"
	"net/smtp"
	"os"
)

// SendCSV sends an email with a CSV attachment using SMTP settings from environment variables
func SendCSV(to, subject, body, filename string, csvData []byte) error {
	host := os.Getenv("ROSTER_SMTP_HOST")
	port := os.Getenv("ROSTER_SMTP_PORT")
	user := os.Getenv("ROSTER_SMTP_USER")
	pass := os.Getenv("ROSTER_SMTP_PASS")
	from := os.Getenv("ROSTER_SMTP_FROM")

	if host == "" || port == "" || user == "" || pass == "" || from == "" {
		return fmt.Errorf("missing SMTP environment variables (ROSTER_SMTP_HOST, ROSTER_SMTP_PORT, ROSTER_SMTP_USER, ROSTER_SMTP_PASS, ROSTER_SMTP_FROM)")
	}

	boundary := "ROSTER_CSV_EXPORT_BOUNDARY"

	header := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: multipart/mixed; boundary=%s\r\n"+
		"\r\n", from, to, subject, boundary)

	bodyPart := fmt.Sprintf("--%s\r\n"+
		"Content-Type: text/plain; charset=\"utf-8\"\r\n"+
		"\r\n"+
		"%s\r\n"+
		"\r\n", boundary, body)

	attachmentPart := fmt.Sprintf("--%s\r\n"+
		"Content-Type: text/csv; name=\"%s\"\r\n"+
		"Content-Transfer-Encoding: base64\r\n"+
		"Content-Disposition: attachment; filename=\"%s\"\r\n"+
		"\r\n", boundary, filename, filename)

	encodedData := base64.StdEncoding.EncodeToString(csvData)

	// Split base64 into lines for better compatibility
	var chunkedData string
	for i := 0; i < len(encodedData); i += 76 {
		end := i + 76
		if end > len(encodedData) {
			end = len(encodedData)
		}
		chunkedData += encodedData[i:end] + "\r\n"
	}

	footer := fmt.Sprintf("--%s--\r\n", boundary)

	fullMessage := header + bodyPart + attachmentPart + chunkedData + "\r\n" + footer

	auth := smtp.PlainAuth("", user, pass, host)
	err := smtp.SendMail(host+":"+port, auth, from, []string{to}, []byte(fullMessage))
	return err
}
