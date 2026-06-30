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

package netbox

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"gotunix.net/roster/internal/models"
	"gotunix.net/roster/internal/store"
	"gotunix.net/roster/internal/ui"
)

type NetBoxResponse struct {
	Count   int            `json:"count"`
	Next    *string        `json:"next"`
	Results []NetBoxObject `json:"results"`
}

type NetBoxObject struct {
	ID           int                    `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	LocalContext map[string]interface{} `json:"local_context_data"`
	PrimaryIP    *struct {
		Address string `json:"address"`
	} `json:"primary_ip"`
	Tags []struct {
		Slug string `json:"slug"`
		Name string `json:"name"`
	} `json:"tags"`
	Status     interface{} `json:"status"`
	Site       interface{} `json:"site"`
	DeviceRole interface{} `json:"device_role"`
	Role       interface{} `json:"role"`
	Platform   interface{} `json:"platform"`
	DeviceType interface{} `json:"device_type"`
	Cluster    interface{} `json:"cluster"`
}

func Sync(baseURL, token, dir, filter string) error {
	fmt.Fprintln(os.Stderr, ui.BoldStyle.Foreground(ui.Cyan).Render("🔄 Syncing from NetBox: "+baseURL))

	// 1. Sync Config Contexts (Groups)
	if err := syncConfigContexts(baseURL, token, dir); err != nil {
		return err
	}

	// 2. Sync Hosts (Devices & VMs)
	fmt.Fprintln(os.Stderr, ui.BoldStyle.Foreground(ui.Cyan).Render("🖥  Syncing Hosts..."))
	deviceIDs, vmIDs, totalSynced, err := syncHosts(baseURL, token, dir, filter)
	if err != nil {
		return err
	}

	// 3. Sync Interfaces
	fmt.Fprintln(os.Stderr, ui.BoldStyle.Foreground(ui.Cyan).Render("🔌 Syncing Interfaces..."))
	intfToHost, intfToName, isVMIntf, err := syncInterfaces(baseURL, token, dir, deviceIDs, vmIDs)
	if err != nil {
		return err
	}

	// 4. Sync IP Addresses
	fmt.Fprintln(os.Stderr, ui.BoldStyle.Foreground(ui.Cyan).Render("🌐 Syncing IP Addresses..."))
	if err := syncIPAddresses(baseURL, token, dir, intfToHost, intfToName, isVMIntf); err != nil {
		return err
	}

	// 5. Sync VM Disks
	if len(vmIDs) > 0 {
		fmt.Fprintln(os.Stderr, ui.BoldStyle.Foreground(ui.Cyan).Render("💾 Syncing VM Disks..."))
		if err := syncVMDisks(baseURL, token, dir, vmIDs); err != nil {
			return err
		}
	}

	fmt.Fprintln(os.Stderr, ui.SuccessMsg("Successfully synced %d hosts from NetBox", totalSynced))
	return nil
}

func syncConfigContexts(baseURL, token, dir string) error {
	fmt.Fprintln(os.Stderr, ui.BoldStyle.Foreground(ui.Cyan).Render("📦 Syncing Group Config Contexts..."))
	contextURL, _ := url.Parse(baseURL)
	contextURL.Path = strings.TrimSuffix(contextURL.Path, "/") + "/api/extras/config-contexts/"

	cURL := contextURL.String()
	for cURL != "" {
		resp, err := authenticatedGet(cURL, token)
		if err != nil {
			return fmt.Errorf("context request failed: %v", err)
		}

		var ctxResp struct {
			Next    *string `json:"next"`
			Results []struct {
				Name         string                 `json:"name"`
				Data         map[string]interface{} `json:"data"`
				Roles        []interface{}          `json:"roles"`
				DeviceGroups []interface{}          `json:"device_groups"`
				Tags         []interface{}          `json:"tags"`
				Platforms    []interface{}          `json:"platforms"`
				Sites        []interface{}          `json:"sites"`
			} `json:"results"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&ctxResp); err != nil {
			resp.Body.Close()
			return fmt.Errorf("failed to decode contexts: %v", err)
		}
		resp.Body.Close()

		for _, ctx := range ctxResp.Results {
			targetGroups := make(map[string]bool)
			for _, s := range extractSlugs(ctx.Roles) {
				targetGroups[s] = true
			}
			for _, s := range extractSlugs(ctx.DeviceGroups) {
				targetGroups[s] = true
			}
			for _, s := range extractSlugs(ctx.Tags) {
				targetGroups[s] = true
			}
			for _, s := range extractSlugs(ctx.Platforms) {
				targetGroups[s] = true
			}
			for _, s := range extractSlugs(ctx.Sites) {
				targetGroups[s] = true
			}

			if len(targetGroups) == 0 {
				continue
			}

			fmt.Fprintln(os.Stderr, ui.DescriptionStyle.Render(fmt.Sprintf("  • Context: %s (mapping to %d groups)", ctx.Name, len(targetGroups))))

			for gName := range targetGroups {
				inv, _ := store.LoadInventory(dir)
				if _, ok := inv.Groups[gName]; !ok {
					if err := store.SaveGroup(dir, gName, &models.Group{Name: gName}); err != nil {
						fmt.Fprintln(os.Stderr, ui.ErrorMsg("  ! Failed to create group %s: %v", gName, err))
						continue
					}
				}
				if err := store.MergeGroupVars(dir, gName, ctx.Data); err != nil {
					fmt.Fprintln(os.Stderr, ui.ErrorMsg("  ! Merging context to group %s: %v", gName, err))
				}
			}
		}
		if ctxResp.Next != nil {
			cURL = *ctxResp.Next
		} else {
			cURL = ""
		}
	}
	return nil
}

func syncHosts(baseURL, token, dir, filter string) (map[int]string, map[int]string, int, error) {
	endpoints := []string{"/api/dcim/devices/", "/api/virtualization/virtual-machines/"}
	totalSynced := 0
	deviceIDs := make(map[int]string)
	vmIDs := make(map[int]string)

	for _, endpoint := range endpoints {
		isVM := strings.Contains(endpoint, "virtual-machines")
		apiURL, err := url.Parse(baseURL)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("invalid baseURL: %v", err)
		}
		apiURL.Path = strings.TrimSuffix(apiURL.Path, "/") + endpoint

		q := apiURL.Query()
		q.Add("exclude", "config_context")
		if filter != "" {
			filters := strings.Split(filter, ",")
			for _, f := range filters {
				parts := strings.SplitN(f, "=", 2)
				if len(parts) == 2 {
					q.Add(parts[0], parts[1])
				}
			}
		}
		apiURL.RawQuery = q.Encode()
		currentURL := apiURL.String()

		for currentURL != "" {
			resp, err := authenticatedGet(currentURL, token)
			if err != nil {
				return nil, nil, 0, fmt.Errorf("host request failed: %v", err)
			}

			if resp.StatusCode != http.StatusOK {
				resp.Body.Close()
				return nil, nil, 0, fmt.Errorf("API returned status %s for %s", resp.Status, currentURL)
			}

			var nbResp NetBoxResponse
			if err := json.NewDecoder(resp.Body).Decode(&nbResp); err != nil {
				resp.Body.Close()
				return nil, nil, 0, fmt.Errorf("failed to decode response: %v", err)
			}
			resp.Body.Close()

			for _, obj := range nbResp.Results {
				name := obj.Name
				if name == "" {
					continue
				}

				if isVM {
					vmIDs[obj.ID] = name
				} else {
					deviceIDs[obj.ID] = name
				}

				if err := store.AddHostToMain(dir, name); err != nil {
					fmt.Fprintln(os.Stderr, ui.ErrorMsg("Adding host %s: %v", name, err))
					continue
				}

				if obj.PrimaryIP != nil && obj.PrimaryIP.Address != "" {
					ip := strings.Split(obj.PrimaryIP.Address, "/")[0]
					if err := store.SetHostVar(dir, name, "ansible_host", ip); err != nil {
						fmt.Fprintln(os.Stderr, ui.ErrorMsg("Setting ansible_host for %s: %v", name, err))
					}
				}

				if obj.Description != "" {
					if err := store.SetHostVar(dir, name, "description", obj.Description); err != nil {
						fmt.Fprintln(os.Stderr, ui.ErrorMsg("Setting description for %s: %v", name, err))
					}
				}

				if len(obj.LocalContext) > 0 {
					if err := store.MergeHostVars(dir, name, obj.LocalContext); err != nil {
						fmt.Fprintln(os.Stderr, ui.ErrorMsg("Merging local_context for %s: %v", name, err))
					}
				}

				netboxVars := make(map[string]interface{})
				if v := extractVal(obj.Status); v != nil {
					netboxVars["status"] = v
				}
				if v := extractVal(obj.Site); v != nil {
					netboxVars["site"] = v
				}
				if v := extractVal(obj.Platform); v != nil {
					netboxVars["platform"] = v
				}
				if v := extractVal(obj.DeviceType); v != nil {
					netboxVars["device_type"] = v
				}
				if v := extractVal(obj.Cluster); v != nil {
					netboxVars["cluster"] = v
				}

				role := extractVal(obj.DeviceRole)
				if role == nil {
					role = extractVal(obj.Role)
				}
				if role != nil {
					netboxVars["role"] = role
				}

				if len(netboxVars) > 0 {
					if err := store.MergeHostVars(dir, name, map[string]interface{}{"netbox": netboxVars}); err != nil {
						fmt.Fprintln(os.Stderr, ui.ErrorMsg("Merging netbox metadata for %s: %v", name, err))
					}
				}

				for _, tag := range obj.Tags {
					groupName := tag.Slug
					if groupName == "" {
						continue
					}
					inv, _ := store.LoadInventory(dir)
					if _, ok := inv.Groups[groupName]; !ok {
						if err := store.SaveGroup(dir, groupName, &models.Group{Name: groupName}); err != nil {
							fmt.Fprintln(os.Stderr, ui.ErrorMsg("Creating group %s from tag: %v", groupName, err))
							continue
						}
					}
					if err := store.AssignHostToGroup(dir, name, groupName); err != nil {
						fmt.Fprintln(os.Stderr, ui.ErrorMsg("Assigning %s to group %s: %v", name, groupName, err))
					}
				}

				totalSynced++
			}
			if nbResp.Next != nil {
				currentURL = *nbResp.Next
			} else {
				currentURL = ""
			}
		}
	}
	return deviceIDs, vmIDs, totalSynced, nil
}

func syncInterfaces(baseURL, token, dir string, deviceIDs, vmIDs map[int]string) (map[int]string, map[int]string, map[int]bool, error) {
	intfToHost := make(map[int]string)
	intfToName := make(map[int]string)
	isVMIntf := make(map[int]bool)

	interfaceEndpoints := []struct {
		path string
		ids  map[int]string
		isVM bool
	}{
		{"/api/dcim/interfaces/", deviceIDs, false},
		{"/api/virtualization/interfaces/", vmIDs, true},
	}

	for _, ie := range interfaceEndpoints {
		if len(ie.ids) == 0 {
			continue
		}

		apiURL, _ := url.Parse(baseURL)
		apiURL.Path = strings.TrimSuffix(apiURL.Path, "/") + ie.path
		currentURL := apiURL.String()

		for currentURL != "" {
			resp, err := authenticatedGet(currentURL, token)
			if err != nil {
				break
			}

			var intfResp struct {
				Next    *string `json:"next"`
				Results []struct {
					ID      int              `json:"id"`
					Name    string           `json:"name"`
					MAC     string           `json:"mac_address"`
					MTU     int              `json:"mtu"`
					Enabled bool             `json:"enabled"`
					Device  struct{ ID int } `json:"device"`
					VM      struct{ ID int } `json:"virtual_machine"`
				} `json:"results"`
			}
			json.NewDecoder(resp.Body).Decode(&intfResp)
			resp.Body.Close()

			for _, intf := range intfResp.Results {
				hostID := intf.Device.ID
				if hostID == 0 {
					hostID = intf.VM.ID
				}

				hName, ok := ie.ids[hostID]
				if !ok {
					continue
				}

				intfToHost[intf.ID] = hName
				intfToName[intf.ID] = intf.Name
				isVMIntf[intf.ID] = ie.isVM

				intfData := map[string]interface{}{
					"name":    intf.Name,
					"enabled": intf.Enabled,
				}
				if intf.MAC != "" {
					intfData["mac"] = intf.MAC
				}
				if intf.MTU != 0 {
					intfData["mtu"] = intf.MTU
				}

				interfaceMap := map[string]interface{}{
					"interfaces": map[string]interface{}{
						intf.Name: intfData,
					},
				}
				store.MergeHostVars(dir, hName, map[string]interface{}{"netbox": interfaceMap})
			}
			if intfResp.Next != nil {
				currentURL = *intfResp.Next
			} else {
				currentURL = ""
			}
		}
	}
	return intfToHost, intfToName, isVMIntf, nil
}

func syncIPAddresses(baseURL, token, dir string, intfToHost, intfToName map[int]string, isVMIntf map[int]bool) error {
	ipURL, _ := url.Parse(baseURL)
	ipURL.Path = strings.TrimSuffix(ipURL.Path, "/") + "/api/ipam/ip-addresses/"
	currentIPURL := ipURL.String()

	for currentIPURL != "" {
		resp, err := authenticatedGet(currentIPURL, token)
		if err != nil {
			break
		}

		var ipResp struct {
			Next    *string `json:"next"`
			Results []struct {
				Address            string                 `json:"address"`
				Status             struct{ Value string } `json:"status"`
				AssignedObjectID   int                    `json:"assigned_object_id"`
				AssignedObjectType string                 `json:"assigned_object_type"`
			} `json:"results"`
		}
		json.NewDecoder(resp.Body).Decode(&ipResp)
		resp.Body.Close()

		for _, ip := range ipResp.Results {
			if ip.AssignedObjectID == 0 {
				continue
			}

			hName, ok := intfToHost[ip.AssignedObjectID]
			if !ok {
				continue
			}

			isVM := strings.Contains(ip.AssignedObjectType, "vminterface")
			if isVM != isVMIntf[ip.AssignedObjectID] {
				continue
			}

			intfName := intfToName[ip.AssignedObjectID]
			ipMap := map[string]interface{}{
				"interfaces": map[string]interface{}{
					intfName: map[string]interface{}{
						"ips": map[string]interface{}{
							ip.Address: map[string]interface{}{
								"status": ip.Status.Value,
							},
						},
					},
				},
			}
			store.MergeHostVars(dir, hName, map[string]interface{}{"netbox": ipMap})
		}
		if ipResp.Next != nil {
			currentIPURL = *ipResp.Next
		} else {
			currentIPURL = ""
		}
	}
	return nil
}

func syncVMDisks(baseURL, token, dir string, vmIDs map[int]string) error {
	apiURL, _ := url.Parse(baseURL)
	apiURL.Path = strings.TrimSuffix(apiURL.Path, "/") + "/api/virtualization/disks/"
	currentURL := apiURL.String()

	for currentURL != "" {
		resp, err := authenticatedGet(currentURL, token)
		if err != nil {
			break
		}

		var diskResp struct {
			Next    *string `json:"next"`
			Results []struct {
				Name string           `json:"name"`
				Size int              `json:"size"`
				VM   struct{ ID int } `json:"virtual_machine"`
			} `json:"results"`
		}
		json.NewDecoder(resp.Body).Decode(&diskResp)
		resp.Body.Close()

		for _, disk := range diskResp.Results {
			hName, ok := vmIDs[disk.VM.ID]
			if !ok {
				continue
			}
			diskData := map[string]interface{}{
				"name": disk.Name,
				"size": disk.Size,
			}
			diskMap := map[string]interface{}{
				"disks": map[string]interface{}{
					disk.Name: diskData,
				},
			}
			store.MergeHostVars(dir, hName, map[string]interface{}{"netbox": diskMap})
		}
		if diskResp.Next != nil {
			currentURL = *diskResp.Next
		} else {
			currentURL = ""
		}
	}
	return nil
}

func authenticatedGet(url, token string) (*http.Response, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Token "+token)
	req.Header.Add("Accept", "application/json")
	return client.Do(req)
}

func extractSlugs(items []interface{}) []string {
	var slugs []string
	for _, item := range items {
		switch v := item.(type) {
		case string:
			slugs = append(slugs, v)
		case map[string]interface{}:
			if slug, ok := v["slug"].(string); ok {
				slugs = append(slugs, slug)
			}
		}
	}
	return slugs
}

func extractVal(item interface{}) interface{} {
	if item == nil {
		return nil
	}
	if m, ok := item.(map[string]interface{}); ok {
		if slug, ok := m["slug"].(string); ok {
			return slug
		}
		if val, ok := m["value"].(string); ok {
			return val
		}
	}
	return item
}
