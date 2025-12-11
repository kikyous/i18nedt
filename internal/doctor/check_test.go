package doctor

import (
	"reflect"
	"testing"

	"github.com/kikyous/i18nedt/pkg/types"
)

func TestCheck(t *testing.T) {
	files := []*types.I18nFile{
		{
			Path:      "en.json",
			Locale:    "en",
			Namespace: "",
			Data:      `{"a": "1", "b": "2"}`,
		},
		{
			Path:      "fr.json",
			Locale:    "fr",
			Namespace: "",
			Data:      `{"a": "1"}`, // missing "b"
		},
		{
			Path:      "de.json",
			Locale:    "de",
			Namespace: "",
			Data:      `{"a": "1", "b": ""}`, // empty "b"
		},
	}

	results, err := Check(files, ":")
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	// Check en.json (should have no issues)
	enRes, ok := results["en.json"]
	if !ok {
		t.Fatal("Result for en.json missing")
	}
	if len(enRes.MissingKeys) > 0 || len(enRes.EmptyKeys) > 0 {
		t.Errorf("en.json should have no issues, got missing: %v, empty: %v", enRes.MissingKeys, enRes.EmptyKeys)
	}

	// Check fr.json (should miss "b")
	frRes, ok := results["fr.json"]
	if !ok {
		t.Fatal("Result for fr.json missing")
	}
	if !reflect.DeepEqual(frRes.MissingKeys, []string{"b"}) {
		t.Errorf("fr.json should miss 'b', got %v", frRes.MissingKeys)
	}

	// Check de.json (should have empty "b")
	deRes, ok := results["de.json"]
	if !ok {
		t.Fatal("Result for de.json missing")
	}
	if !reflect.DeepEqual(deRes.EmptyKeys, []string{"b"}) {
		t.Errorf("de.json should have empty 'b', got %v", deRes.EmptyKeys)
	}
}

func TestCheck_Namespaces(t *testing.T) {
	files := []*types.I18nFile{
		{
			Path:      "ns1_en.json",
			Locale:    "en",
			Namespace: "ns1",
			Data:      `{"key1": "val"}`,
		},
		{
			Path:      "ns1_fr.json",
			Locale:    "fr",
			Namespace: "ns1",
			Data:      `{}`, // missing key1
		},
		{
			Path:      "ns2_en.json",
			Locale:    "en",
			Namespace: "ns2",
			Data:      `{"key2": "val"}`,
		},
	}

	results, err := Check(files, ":")
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	// ns1_fr should miss key1 (prefixed by ns1: maybe? Wait, Flatten uses prefix if namespace provided)
	// Let's check logic in FlattenJSON.
	// FlattenJSON(..., namespace) -> returns keys prefixed with "namespace:"?
	// YES.

	frRes := results["ns1_fr.json"]
	expectedKey := "ns1:key1"
	if len(frRes.MissingKeys) != 1 || frRes.MissingKeys[0] != expectedKey {
		t.Errorf("ns1_fr.json should miss %s, got %v", expectedKey, frRes.MissingKeys)
	}

	// ns2_en should be fine. It is in a different namespace, so it shouldn't care about ns1 keys.
	en2Res := results["ns2_en.json"]
	if len(en2Res.MissingKeys) > 0 {
		t.Errorf("ns2_en.json should have no missing keys, got %v", en2Res.MissingKeys)
	}
}
