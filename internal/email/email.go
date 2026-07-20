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

package email

import (
	"encoding/base64"
	"fmt"
	"net/smtp"
	"os"
	"strings"
)

// SMTPConfig holds SMTP server settings for sending emails
type SMTPConfig struct {
	Host string
	Port string
	From string
	User string
	Pass string
}

// LoadSMTPConfig returns SMTP settings from environment variables (preferred)
// or falls back to os.Getenv for each individual value. Callers can also
// populate this struct from a roster.conf file before calling SendCSV.
func LoadSMTPConfig() SMTPConfig {
	return SMTPConfig{
		Host: os.Getenv("ROSTER_SMTP_HOST"),
		Port: os.Getenv("ROSTER_SMTP_PORT"),
		From: os.Getenv("ROSTER_SMTP_FROM"),
		User: os.Getenv("ROSTER_SMTP_USER"),
		Pass: os.Getenv("ROSTER_SMTP_PASS"),
	}
}

// SendCSV sends an email with a CSV attachment using the given SMTP config
func SendCSV(to, subject, body, filename string, csvData []byte, smtpCfg SMTPConfig) error {
	required := []struct {
		value string
		name  string
	}{
		{smtpCfg.Host, "SMTP host"},
		{smtpCfg.Port, "SMTP port"},
		{smtpCfg.From, "SMTP from address"},
	}

	var missing []string
	for _, r := range required {
		if r.value == "" {
			missing = append(missing, r.name)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing SMTP settings (set in .roster.conf or %s environment variables): %s",
			"ROSTER_SMTP_*", strings.Join(missing, ", "))
	}

	host := smtpCfg.Host
	port := smtpCfg.Port
	from := smtpCfg.From
	user := smtpCfg.User
	pass := smtpCfg.Pass

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

	var auth smtp.Auth
	if user != "" || pass != "" {
		auth = smtp.PlainAuth("", user, pass, host)
	}

	err := smtp.SendMail(host+":"+port, auth, from, []string{to}, []byte(fullMessage))
	return err
}
