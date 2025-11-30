package i18n

import (
	"testing"

	"github.com/kikyous/i18nedt/pkg/types"
)

func TestParseLocaleFromPath(t *testing.T) {
	tests := []struct {
		name      string
		filePath  string
		want      string
		wantErr   bool
	}{
		{
			name:     "standard locale format zh-CN",
			filePath: "/path/to/locales/zh-CN.json",
			want:     "zh-CN",
			wantErr:  false,
		},
		{
			name:     "standard locale format en-US",
			filePath: "src/locales/en-US.json",
			want:     "en-US",
			wantErr:  false,
		},
		{
			name:     "simple two-letter locale",
			filePath: "locales/en.json",
			want:     "en",
			wantErr:  false,
		},
		{
			name:     "simple two-letter locale zh",
			filePath: "locales/zh.json",
			want:     "zh",
			wantErr:  false,
		},
		{
			name:     "locale with additional suffix",
			filePath: "locales/zh-CN.messages.json",
			want:     "zh-CN",
			wantErr:  false,
		},
		{
			name:     "locale directory structure",
			filePath: "src/locales/zh-CN/app.json",
			want:     "zh-CN",
			wantErr:  false,
		},
		{
			name:     "simple locale directory structure",
			filePath: "locales/en/messages.json",
			want:     "en",
			wantErr:  false,
		},
		{
			name:     "complex path with locale directory",
			filePath: "/home/user/project/src/i18n/locales/ja-JP/translation.json",
			want:     "ja-JP",
			wantErr:  false,
		},
		{
			name:     "non-standard locale in filename",
			filePath: "locales/unknown.json",
			want:     "unknown",
			wantErr:  false,
		},
		{
			name:     "non-standard locale in path",
			filePath: "locales/custom-name/file.json",
			want:     "file",
			wantErr:  false,
		},
		{
			name:     "filename without extension",
			filePath: "locales/zh-CN",
			want:     "zh-CN",
			wantErr:  false,
		},
		{
			name:     "uppercase extension",
			filePath: "locales/fr-FR.JSON",
			want:     "fr-FR",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLocaleFromPath(tt.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLocaleFromPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseLocaleFromPath() got = %v, want %v", got, tt.want)
			}
		})
	}
}


func TestGetLocaleList(t *testing.T) {
	tests := []struct {
		name      string
		filePaths []string
		want      []string
		wantErr   bool
	}{
		{
			name:      "standard locales",
			filePaths: []string{"zh-CN.json", "en-US.json"},
			want:      []string{"zh-CN", "en-US"},
			wantErr:   false,
		},
		{
			name:      "simple two-letter locales",
			filePaths: []string{"zh.json", "en.json"},
			want:      []string{"zh", "en"},
			wantErr:   false,
		},
		{
			name:      "mixed locale formats",
			filePaths: []string{"zh-CN.json", "en.json", "ja-JP.messages.json"},
			want:      []string{"zh-CN", "en", "ja-JP"},
			wantErr:   false,
		},
		{
			name:      "duplicate locales",
			filePaths: []string{"zh-CN.json", "zh-CN.messages.json"},
			want:      []string{"zh-CN", "zh-CN"},
			wantErr:   false,
		},
		{
			name:      "non-standard locales",
			filePaths: []string{"custom.json", "unknown.txt"},
			want:      []string{"custom", "unknown"},
			wantErr:   false,
		},
		{
			name:      "empty list",
			filePaths: []string{},
			want:      []string{},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create I18nFile structs first
			files, err := LoadAllFiles(tt.filePaths)
			if err != nil {
				t.Errorf("LoadAllFiles() error = %v", err)
				return
			}

			got, err := GetLocaleList(files)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLocaleList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("GetLocaleList() length = %v, want %v", len(got), len(tt.want))
				return
			}
			for i, locale := range got {
				if locale != tt.want[i] {
					t.Errorf("GetLocaleList()[%d] = %v, want %v", i, locale, tt.want[i])
				}
			}
		})
	}
}

func TestFindFileByLocale(t *testing.T) {
	files := []*types.I18nFile{
		{Path: "zh-CN.json", Data: "{}", Locale: "zh-CN"},
		{Path: "en-US.json", Data: "{}", Locale: "en-US"},
		{Path: "ja-JP.json", Data: "{}", Locale: "ja-JP"},
	}

	tests := []struct {
		name  string
		locale string
		want  *types.I18nFile
	}{
		{
			name:   "find existing locale zh-CN",
			locale: "zh-CN",
			want:   files[0],
		},
		{
			name:   "find existing locale en-US",
			locale: "en-US",
			want:   files[1],
		},
		{
			name:   "find existing locale ja-JP",
			locale: "ja-JP",
			want:   files[2],
		},
		{
			name:   "find non-existent locale",
			locale: "fr-FR",
			want:   nil,
		},
		{
			name:   "find empty locale",
			locale: "",
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindFileByLocale(files, tt.locale)
			if got != tt.want {
				t.Errorf("FindFileByLocale() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindFileByLocaleComplex(t *testing.T) {
	files := []*types.I18nFile{
		{Path: "src/locales/zh-CN/app.json", Data: "{}", Locale: "zh-CN"},
		{Path: "src/locales/en-US/app.json", Data: "{}", Locale: "en-US"},
		{Path: "src/locales/zh-CN.messages.json", Data: "{}", Locale: "zh-CN"},
		{Path: "locales/en.json", Data: "{}", Locale: "en"},
	}

	tests := []struct {
		name  string
		locale string
		want  *types.I18nFile
	}{
		{
			name:   "find zh-CN in complex path",
			locale: "zh-CN",
			want:   files[0], // Should find the first match
		},
		{
			name:   "find en-US in complex path",
			locale: "en-US",
			want:   files[1],
		},
		{
			name:   "find en from simple locale",
			locale: "en",
			want:   files[3],
		},
		{
			name:   "find non-existent locale",
			locale: "ko-KR",
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindFileByLocale(files, tt.locale)
			if got != tt.want {
				t.Errorf("FindFileByLocale() = %v, want %v", got, tt.want)
			}
		})
	}
}