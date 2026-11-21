package main

import (
	"encoding/json"
	"testing"

	"github.com/egoist/mygo/i18n"

	"mygo-store/locales"
)

func TestSystemLanguagesMatchTheAppLanguages(t *testing.T) {
	languages := locales.Catalog.Languages()
	cases := map[string]string{
		"en-US": "en",
		"ja-JP": "ja",
		"ko-KR": "ko",
		"zh-CN": "zh-Hans",
		"zh-SG": "zh-Hans",
		"zh-TW": "zh-Hant",
		"zh-HK": "zh-Hant",
	}
	for tag, want := range cases {
		if got := i18n.Match([]string{tag}, languages); got != want {
			t.Errorf("i18n.Match(%q) = %q, want %q", tag, got, want)
		}
	}
}

func TestContentTag(t *testing.T) {
	cases := map[string]string{
		"":           "",
		"en":         "en",
		"en-US":      "en",
		"zh-Hans":    "zh-CN",
		"zh-Hans-HK": "zh-CN",
		"zh-Hant":    "zh-TW",
		"zh-TW":      "zh-TW",
		"zh-HK":      "zh-TW",
		"zh-MO":      "zh-TW",
		"ja":         "ja",
		"ja-JP":      "ja",
		"ko":         "ko",
		"ko-KR":      "ko",
		"fr":         "fr",
	}
	for locale, want := range cases {
		if got := contentTag(locale); got != want {
			t.Errorf("contentTag(%q) = %q, want %q", locale, got, want)
		}
	}
}

func TestTextIn(t *testing.T) {
	text := Text{
		"en":    "English",
		"zh-CN": "简体",
		"zh-TW": "繁體",
		"ja":    "日本語",
		"ko":    "한국어",
	}
	cases := map[string]string{
		"en":      "English",
		"zh-Hans": "简体",
		"zh-Hant": "繁體",
		"ja":      "日本語",
		"ko":      "한국어",
		"fr":      "English", // falls back to English
	}
	for locale, want := range cases {
		if got := text.In(locale); got != want {
			t.Errorf("Text.In(%q) = %q, want %q", locale, got, want)
		}
	}

	partial := Text{"en": "English", "ja": "日本語"}
	if got := partial.In("ko"); got != "English" {
		t.Errorf("a missing language = %q, want English", got)
	}
}

func TestTextKeepsACatalogOfOneLanguage(t *testing.T) {
	var text Text
	if err := json.Unmarshal([]byte(`"只有一种语言"`), &text); err != nil {
		t.Fatalf("unmarshal a one-language text: %v", err)
	}
	for _, locale := range []string{"en", "zh-Hans", "ja", "ko"} {
		if got := text.In(locale); got != "只有一种语言" {
			t.Errorf("Text.In(%q) = %q, want the catalog's one text", locale, got)
		}
	}

	encoded, err := json.Marshal(text)
	if err != nil {
		t.Fatalf("marshal the one-language text: %v", err)
	}
	var decoded Text
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal the encoded text: %v", err)
	}
	if got := decoded.In("en"); got != "只有一种语言" {
		t.Errorf("round trip = %q, want the catalog's one text", got)
	}
}

func TestParseCatalogReadsACatalogOfOneLanguage(t *testing.T) {
	data := []byte(`{"apps":[{"slug":"a","name":"A","url":"https://example.com/a",
		"summary":"简介","description":"描述"}]}`)
	catalog, err := parseCatalog(data)
	if err != nil {
		t.Fatalf("parse an old catalog: %v", err)
	}
	app := catalog.Apps[0]
	if got := app.Summary.In("ja"); got != "简介" {
		t.Errorf("summary = %q, want the catalog's one text", got)
	}
	if got := app.Description.In("ko"); got != "描述" {
		t.Errorf("description = %q, want the catalog's one text", got)
	}
}

func TestBundledCatalogHasEveryLanguage(t *testing.T) {
	catalog := loadCatalog()
	for i := range catalog.Apps {
		app := &catalog.Apps[i]
		if app.Summary.In("en") == "" || app.Description.In("en") == "" {
			t.Errorf("%s has no English summary or description", app.Slug)
		}
	}

	var found *App
	for i := range catalog.Apps {
		if catalog.Apps[i].Slug == "2048" {
			found = &catalog.Apps[i]
			break
		}
	}
	if found == nil {
		t.Skip("2048 is not in the catalog")
	}
	for _, tag := range []string{"en", "zh-CN", "zh-TW", "ja", "ko"} {
		if found.Summary[tag] == "" || found.Description[tag] == "" {
			t.Errorf("the bundled catalog has no %s text of 2048", tag)
		}
	}
}

func TestSearchMatchesEveryLanguage(t *testing.T) {
	app := &App{
		Name:    "Term",
		Summary: Text{"en": "A terminal emulator", "ja": "端末エミュレーター"},
	}
	if !app.matches("terminal", "ja") {
		t.Error("an English query does not match with a Japanese interface")
	}
	if !app.matches("端末", "en") {
		t.Error("a Japanese query does not match with an English interface")
	}
	if app.matches("spreadsheet", "en") {
		t.Error("an unrelated query matches")
	}
}

func TestLanguageChoicesCoverTheCatalogLanguages(t *testing.T) {
	choices := languageChoices(locales.In("en"))
	tags := make(map[string]string, len(choices))
	for _, choice := range choices {
		tags[choice.tag] = choice.label
	}
	for _, tag := range []string{"", "en", "zh-Hans", "zh-Hant", "ja", "ko"} {
		label, ok := tags[tag]
		if !ok {
			t.Errorf("the language menu has no %q", tag)
			continue
		}
		if label == "" {
			t.Errorf("the language %q has no label", tag)
		}
	}
}

func TestLanguageMenuChoosesJapanese(t *testing.T) {
	s, tt := newTestStore(t)
	var applied []string
	previous := applyLocale
	applyLocale = func(tag string) { applied = append(applied, tag) }
	t.Cleanup(func() {
		applyLocale = previous
		languageChoice = ""
	})

	if err := tt.Click("语言"); err != nil {
		t.Fatalf("open the language menu: %v", err)
	}
	tt.Frame()
	if err := tt.Click("日本語"); err != nil {
		t.Fatalf("choose Japanese: %v", err)
	}
	if len(applied) != 1 || applied[0] != "ja" {
		t.Fatalf("applied languages %q, want ja", applied)
	}
	if s.prefs.Locale != "ja" {
		t.Errorf("saved locale = %q, want ja", s.prefs.Locale)
	}

	tt.SetLocale("ja")
	tt.Frame()
	if !tt.HasText("言語") || !tt.HasText("見つける") {
		t.Errorf("the interface did not switch to Japanese: %q", tt.Texts())
	}
}

func TestAppTextFollowsTheInterfaceLanguage(t *testing.T) {
	s, tt := newTestStore(t)
	app := s.appBySlug("2048")
	if app == nil {
		t.Skip("2048 is not in the catalog")
	}
	cases := []struct{ locale, about string }{
		{"en", "About"},
		{"zh-Hans", "关于"},
		{"zh-Hant", "關於"},
		{"ja", "概要"},
		{"ko", "소개"},
	}
	for _, tc := range cases {
		t.Run(tc.locale, func(t *testing.T) {
			tt.SetLocale(tc.locale)
			s.openDetail("2048")
			tt.Frame()
			if !tt.HasText(tc.about) {
				t.Errorf("the %s interface does not show %q: %q", tc.locale, tc.about, tt.Texts())
			}
			description := app.Description.In(tc.locale)
			if !tt.HasText(description) {
				t.Errorf("the %s detail page does not show its description %q", tc.locale, description)
			}
			if tc.locale != "en" && description == app.Description.In("en") {
				t.Errorf("the %s detail page fell back to English", tc.locale)
			}
			s.detail = ""
		})
	}
}
