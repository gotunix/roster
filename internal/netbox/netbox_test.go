package netbox

import (
	"reflect"
	"testing"
)

func TestMergeNetboxVars(t *testing.T) {
	local := map[string]interface{}{
		"ansible_host": "192.168.1.50",
		"ansible_user": "custom_user",
		"netbox": map[string]interface{}{
			"status": "active",
			"site":   "hq",
		},
	}

	netboxData := map[string]interface{}{
		"ansible_host": "192.168.1.1", // should NOT overwrite local
		"description":  "New Description", // should be added since it doesn't exist
		"netbox": map[string]interface{}{
			"site":    "remote", // should overwrite inner netbox.site
			"tags":    []interface{}{"prod"}, // should be added
		},
	}

	MergeNetboxVars(local, netboxData)

	expected := map[string]interface{}{
		"ansible_host": "192.168.1.50",   // preserved
		"ansible_user": "custom_user",    // preserved
		"description":  "New Description", // added
		"netbox": map[string]interface{}{
			"status": "active",
			"site":   "remote",
			"tags":    []interface{}{"prod"},
		},
	}

	if !reflect.DeepEqual(local["ansible_host"], expected["ansible_host"]) {
		t.Errorf("Expected ansible_host %v, got %v", expected["ansible_host"], local["ansible_host"])
	}
	if !reflect.DeepEqual(local["ansible_user"], expected["ansible_user"]) {
		t.Errorf("Expected ansible_user %v, got %v", expected["ansible_user"], local["ansible_user"])
	}
	if !reflect.DeepEqual(local["description"], expected["description"]) {
		t.Errorf("Expected description %v, got %v", expected["description"], local["description"])
	}

	localNetbox := local["netbox"].(map[string]interface{})
	expectedNetbox := expected["netbox"].(map[string]interface{})
	if !reflect.DeepEqual(localNetbox, expectedNetbox) {
		t.Errorf("Expected nested netbox map %v, got %v", expectedNetbox, localNetbox)
	}
}
