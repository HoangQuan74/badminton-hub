package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"

	"github.com/gin-gonic/gin"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

const localizerKey = "i18n_localizer"

//go:embed locales/*.json
var localeFiles embed.FS

type Service struct {
	bundle *goi18n.Bundle
}

func New() (*Service, error) {
	bundle := goi18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	paths, err := fs.Glob(localeFiles, "locales/*.json")
	if err != nil {
		return nil, fmt.Errorf("list locale files: %w", err)
	}
	for _, path := range paths {
		content, err := localeFiles.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read locale %s: %w", path, err)
		}
		if _, err := bundle.ParseMessageFileBytes(content, path); err != nil {
			return nil, fmt.Errorf("parse locale %s: %w", path, err)
		}
	}
	return &Service{bundle: bundle}, nil
}

func (s *Service) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		preferred := c.Query("lang")
		acceptLanguage := c.GetHeader("Accept-Language")
		c.Set(localizerKey, goi18n.NewLocalizer(s.bundle, preferred, acceptLanguage))
		c.Next()
	}
}

func T(c *gin.Context, messageID string) string {
	value, ok := c.Get(localizerKey)
	localizer, valid := value.(*goi18n.Localizer)
	if !ok || !valid {
		return messageID
	}
	message, err := localizer.Localize(&goi18n.LocalizeConfig{MessageID: messageID})
	if err != nil {
		return messageID
	}
	return message
}

func (s *Service) Translate(languageTag, messageID string) (string, error) {
	return goi18n.NewLocalizer(s.bundle, languageTag).Localize(&goi18n.LocalizeConfig{MessageID: messageID})
}
