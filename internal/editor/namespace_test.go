package editor

import (
	"testing"

	"github.com/kikyous/i18nedt/pkg/types"
	"github.com/tidwall/gjson"
)

func TestNamespaceSupport(t *testing.T) {
	files := []*types.I18nFile{
		{
			Path:      "locales/en/common.json",
			Locale:    "en",
			Namespace: "common",
			Data:      `{"hello": "world"}`,
		},
		{
			Path:      "locales/en/auth.json",
			Locale:    "en",
			Namespace: "auth",
			Data:      `{"login": "sign in"}`,
		},
	}

	// 1. Test CreateTempFile
	// We ask for "hello" (ambiguous) and "auth:login" (specific)
	keys := []string{"hello", "auth:login"}
	temp, err := CreateTempFile(files, keys)
	if err != nil {
		t.Fatalf("CreateTempFile failed: %v", err)
	}

	// Check content
	// "hello" -> should match common (found) and auth (not found, so empty)
	
	// Verify common:hello
	if inner, ok := temp.Content["common:hello"]; !ok {
		t.Error("common:hello key missing from temp file")
	} else {
		val := inner["en"]
		if val == nil || val.Value != "world" {
			t.Errorf("Expected common:hello to be 'world', got %v", val)
		}
	}

	// Verify auth:login
	if inner, ok := temp.Content["auth:login"]; !ok {
		t.Error("auth:login key missing from temp file")
	} else {
		val := inner["en"]
		if val == nil || val.Value != "sign in" {
			t.Errorf("Expected auth:login to be 'sign in', got %v", val)
		}
	}

	// 2. Test ApplyChanges
	tempUpdate := &types.TempFile{
		Content: map[string]map[string]*types.Value{
			"common:hello": {
				"en": types.NewStringValue("world updated"),
			},
			"auth:new_key": {
				"en": types.NewStringValue("new value"),
			},
		},
		Deletes: []string{"auth:login"},
	}

	err = ApplyChanges(files, tempUpdate)
	if err != nil {
		t.Fatalf("ApplyChanges failed: %v", err)
	}

	// Verify updates in files[0] (common)
	if res := gjson.Get(files[0].Data, "hello"); res.String() != "world updated" {
		t.Errorf("Failed to update common:hello. Got: %s", res.String())
	}

	// Verify updates in files[1] (auth)
	// "new_key" added
	if res := gjson.Get(files[1].Data, "new_key"); res.String() != "new value" {
		t.Errorf("Failed to add auth:new_key. Got: %s", res.String())
	}
	// "login" deleted
	if res := gjson.Get(files[1].Data, "login"); res.Exists() {
		t.Errorf("Failed to delete auth:login")
	}
}
