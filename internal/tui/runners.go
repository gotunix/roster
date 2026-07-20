package tui

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"gotunix.net/roster/internal/email"
	"gotunix.net/roster/internal/netbox"
	"gotunix.net/roster/internal/store"
)

func (m *MainTUIModel) runNetboxSync(url, token, filter string, syncConfigContexts, syncHosts, syncInterfaces, syncIPs, syncVMDisks bool) tea.Cmd {
	return func() tea.Msg {
		err := store.LockInventory(m.inventoryPath)
		if err != nil {
			return SyncFinishedMsg{err: err}
		}
		defer store.UnlockInventory()

		netbox.LogWriter = m.syncLogBuffer

		opts := netbox.SyncOptions{
			SyncConfigContexts: syncConfigContexts,
			SyncHosts:          syncHosts,
			SyncInterfaces:     syncInterfaces,
			SyncIPs:            syncIPs,
			SyncVMDisks:        syncVMDisks,
		}
		err = netbox.Sync(url, token, m.inventoryPath, filter, opts)
		return SyncFinishedMsg{err: err}
	}
}

func (m *MainTUIModel) runExport(output, emailAddr string, vars []string, exclude []string, groupFilter string) tea.Cmd {
	return func() tea.Msg {
		inv, err := store.LoadInventory(m.inventoryPath)
		if err != nil {
			return ExportFinishedMsg{err: err}
		}

		csvData, numHosts, err := store.ExportInventory(inv, m.inventoryPath, vars, exclude, groupFilter)
		if err != nil {
			return ExportFinishedMsg{err: err}
		}

		if output != "" {
			if err := os.WriteFile(output, csvData, 0644); err != nil {
				return ExportFinishedMsg{err: err}
			}
		}

		if emailAddr != "" {
			subject := "Roster Inventory Export"
			body := fmt.Sprintf("Please find attached the inventory export from Roster.\n\n"+
				"Inventories included: %s\n"+
				"Total hosts: %d", m.inventoryPath, numHosts)
			filename := "inventory_export.csv"
			if output != "" {
				filename = filepath.Base(output)
			}
			smtpCfg := email.LoadSMTPConfig()
			if smtpCfg.Host == "" {
				h, p, f, u, pw := store.LoadSMTPConfig()
				smtpCfg = email.SMTPConfig{Host: h, Port: p, From: f, User: u, Pass: pw}
			}
			if err := email.SendCSV(emailAddr, subject, body, filename, csvData, smtpCfg); err != nil {
				return ExportFinishedMsg{err: err}
			}
		}

		return ExportFinishedMsg{
			err:      nil,
			numHosts: numHosts,
			outFile:  output,
		}
	}
}
