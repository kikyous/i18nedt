package i18n

import (
	"sort"
	"testing"
)

func TestGetKeysUnderPrefix(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		prefix   string
		expected []string
	}{
		{
			name: "leaf key - return itself",
			data: map[string]interface{}{
				"welcome": "Welcome",
				"home": map[string]interface{}{
					"title": "Home",
				},
			},
			prefix:   "welcome",
			expected: []string{"welcome"},
		},
		{
			name: "parent key - return all children",
			data: map[string]interface{}{
				"home": map[string]interface{}{
					"title": "Home",
					"desc":  "Description",
				},
			},
			prefix:   "home",
			expected: []string{"home.desc", "home.title"},
		},
		{
			name: "deeply nested parent key",
			data: map[string]interface{}{
				"nav": map[string]interface{}{
					"menu": map[string]interface{}{
						"home": map[string]interface{}{
							"title": "Home",
							"desc":  "Description",
						},
						"about": map[string]interface{}{
							"title": "About",
						},
					},
				},
			},
			prefix:   "nav.menu",
			expected: []string{"nav.menu.about.title", "nav.menu.home.desc", "nav.menu.home.title"},
		},
		{
			name: "non-existent key",
			data: map[string]interface{}{
				"welcome": "Welcome",
			},
			prefix:   "nonexistent",
			expected: []string{},
		},
		{
			name: "mixed nested and leaf",
			data: map[string]interface{}{
				"home": map[string]interface{}{
					"title": "Home",
				},
				"simple": "Simple value",
			},
			prefix:   "home",
			expected: []string{"home.title"},
		},
		{
			name: "key with single child",
			data: map[string]interface{}{
				"app": map[string]interface{}{
					"title": "App Title",
				},
			},
			prefix:   "app",
			expected: []string{"app.title"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetKeysUnderPrefix(tt.data, tt.prefix)

			// Sort both slices for comparison
			sortedResult := make([]string, len(result))
			copy(sortedResult, result)
			sort.Strings(sortedResult)

			sortedExpected := make([]string, len(tt.expected))
			copy(sortedExpected, tt.expected)
			sort.Strings(sortedExpected)

			if len(sortedResult) != len(sortedExpected) {
				t.Errorf("GetKeysUnderPrefix() length = %v, want %v", len(sortedResult), len(sortedExpected))
				return
			}

			for i, key := range sortedResult {
				if key != sortedExpected[i] {
					t.Errorf("GetKeysUnderPrefix()[%d] = %v, want %v", i, key, sortedExpected[i])
				}
			}
		})
	}
}

func TestExpandKeys(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		keys     []string
		expected []string
	}{
		{
			name: "all leaf keys - no expansion",
			data: map[string]interface{}{
				"welcome": "Welcome",
				"goodbye": "Goodbye",
			},
			keys:     []string{"welcome", "goodbye"},
			expected: []string{"goodbye", "welcome"},
		},
		{
			name: "one parent key and one leaf key",
			data: map[string]interface{}{
				"home": map[string]interface{}{
					"title": "Home",
					"desc":  "Description",
				},
				"simple": "Simple value",
			},
			keys:     []string{"home", "simple"},
			expected: []string{"home.desc", "home.title", "simple"},
		},
		{
			name: "multiple parent keys",
			data: map[string]interface{}{
				"home": map[string]interface{}{
					"title": "Home",
				},
				"nav": map[string]interface{}{
					"menu": map[string]interface{}{
						"home": "Nav Home",
						"about": "About",
					},
				},
			},
			keys:     []string{"home", "nav.menu"},
			expected: []string{"home.title", "nav.menu.about", "nav.menu.home"},
		},
		{
			name: "duplicate keys after expansion",
			data: map[string]interface{}{
				"home": map[string]interface{}{
					"title": "Home",
					"desc":  "Description",
				},
			},
			keys:     []string{"home", "home.title"},
			expected: []string{"home.desc", "home.title"},
		},
		{
			name: "non-existent keys",
			data: map[string]interface{}{
				"welcome": "Welcome",
			},
			keys:     []string{"nonexistent", "another.nonexistent"},
			expected: []string{"another.nonexistent", "nonexistent"},
		},
		{
			name: "deeply nested expansion",
			data: map[string]interface{}{
				"app": map[string]interface{}{
					"nav": map[string]interface{}{
						"header": map[string]interface{}{
							"title": "Header",
							"menu": map[string]interface{}{
								"home": "Home",
								"about": "About",
							},
						},
					},
				},
			},
			keys:     []string{"app.nav.header"},
			expected: []string{"app.nav.header.menu.about", "app.nav.header.menu.home", "app.nav.header.title"},
		},
		{
			name: "empty data",
			data: map[string]interface{}{},
			keys: []string{"home", "welcome"},
			expected: []string{"home", "welcome"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandKeys(tt.data, tt.keys)

			// Sort both slices for comparison
			sortedResult := make([]string, len(result))
			copy(sortedResult, result)
			sort.Strings(sortedResult)

			sortedExpected := make([]string, len(tt.expected))
			copy(sortedExpected, tt.expected)
			sort.Strings(sortedExpected)

			if len(sortedResult) != len(sortedExpected) {
				t.Errorf("ExpandKeys() length = %v, want %v", len(sortedResult), len(sortedExpected))
				t.Errorf("ExpandKeys() result = %v", sortedResult)
				t.Errorf("ExpandKeys() expected = %v", sortedExpected)
				return
			}

			for i, key := range sortedResult {
				if key != sortedExpected[i] {
					t.Errorf("ExpandKeys()[%d] = %v, want %v", i, key, sortedExpected[i])
				}
			}
		})
	}
}

func TestExpandKeysRealWorld(t *testing.T) {
	// Test with a realistic i18n data structure
	data := map[string]interface{}{
		"home": map[string]interface{}{
			"welcome": "Welcome",
			"start":    "Get Started",
		},
		"nav": map[string]interface{}{
			"header": map[string]interface{}{
				"menu": map[string]interface{}{
					"home":   "Home",
					"about":  "About",
					"contact": "Contact",
				},
			},
		},
		"footer": map[string]interface{}{
			"copyright": "© 2024",
			"links": map[string]interface{}{
				"privacy":  "Privacy Policy",
				"terms":    "Terms of Service",
			},
		},
		"simple": "Simple Value",
	}

	tests := []struct {
		name     string
		keys     []string
		expected []string
	}{
		{
			name:     "expand nav.header",
			keys:     []string{"nav.header"},
			expected: []string{"nav.header.menu.about", "nav.header.menu.contact", "nav.header.menu.home"},
		},
		{
			name:     "expand footer",
			keys:     []string{"footer"},
			expected: []string{"footer.copyright", "footer.links.privacy", "footer.links.terms"},
		},
		{
			name:     "mixed parent and leaf keys",
			keys:     []string{"nav", "simple"},
			expected: []string{"nav.header.menu.about", "nav.header.menu.contact", "nav.header.menu.home", "simple"},
		},
		{
			name:     "all individual keys - no expansion",
			keys:     []string{"simple", "home.welcome"},
			expected: []string{"home.welcome", "simple"},
		},
		{
			name:     "mix of existing and non-existing keys",
			keys:     []string{"simple", "nonexistent.key", "home"},
			expected: []string{"home.start", "home.welcome", "nonexistent.key", "simple"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandKeys(data, tt.keys)

			// Sort both slices for comparison
			sortedResult := make([]string, len(result))
			copy(sortedResult, result)
			sort.Strings(sortedResult)

			sortedExpected := make([]string, len(tt.expected))
			copy(sortedExpected, tt.expected)
			sort.Strings(sortedExpected)

			if len(sortedResult) != len(sortedExpected) {
				t.Errorf("ExpandKeys() length = %v, want %v", len(sortedResult), len(sortedExpected))
				return
			}

			for i, key := range sortedResult {
				if key != sortedExpected[i] {
					t.Errorf("ExpandKeys()[%d] = %v, want %v", i, key, sortedExpected[i])
				}
			}
		})
	}
}