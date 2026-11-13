package exception

import (
	"embed"
	"encoding/json"

	"golang.org/x/text/language"
)

//go:embed locales/*.json
var tenantLocales embed.FS

var tenantMessages = loadTenantMessages()

func loadTenantMessages() map[string]map[string]string {
	messages := map[string]map[string]string{}
	for _, lang := range []string{"en", "vi"} {
		data, err := tenantLocales.ReadFile("locales/" + lang + ".json")
		if err != nil {
			panic(err)
		}
		if err := json.Unmarshal(data, &messages); err != nil {
			panic(err)
		}
	}
	return messages
}

func TenantLanguage(query, accept string) (string, bool) {
	if query != "" {
		if query == "en" || query == "vi" {
			return query, true
		}
		return "en", false
	}
	tags, _, err := language.ParseAcceptLanguage(accept)
	if err != nil {
		return "en", true
	}
	_, index, _ := language.NewMatcher([]language.Tag{language.English, language.Vietnamese}).Match(tags...)
	return []string{"en", "vi"}[index], true
}

func TenantMessage(code, lang string) string {
	if value := tenantMessages[lang][code]; value != "" {
		return value
	}
	if value := tenantMessages["en"][code]; value != "" {
		return value
	}
	return tenantMessages["en"]["INTERNAL_ERROR"]
}
