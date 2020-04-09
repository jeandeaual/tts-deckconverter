package mtg

import (
	"net/url"
	"regexp"
	"strings"

	scryfall "github.com/BlueMonday/go-scryfall"

	"deckconverter/plugins"
)

type imageQuality string

const (
	small  imageQuality = "small"
	normal imageQuality = "normal"
	large  imageQuality = "large"
	png    imageQuality = "png"
)

type magicPlugin struct {
	id   string
	name string
}

func (p magicPlugin) PluginID() string {
	return p.id
}

func (p magicPlugin) PluginName() string {
	return p.name
}

func (p magicPlugin) SupportedLanguages() []string {
	return []string{
		string(scryfall.LangEnglish),
		string(scryfall.LangSpanish),
		string(scryfall.LangFrench),
		string(scryfall.LangGerman),
		string(scryfall.LangItalian),
		string(scryfall.LangPortuguese),
		string(scryfall.LangJapanese),
		string(scryfall.LangKorean),
		string(scryfall.LangRussian),
		string(scryfall.LangSimplifiedChinese),
		string(scryfall.LangTraditionalChinese),
	}
}

func (p magicPlugin) AvailableOptions() plugins.Options {
	return plugins.Options{
		"quality": plugins.Option{
			Type:        plugins.OptionTypeEnum,
			Description: "image quality",
			AllowedValues: []string{
				string(small),
				string(normal),
				string(large),
				string(png),
			},
			DefaultValue: string(large),
		},
		"rulings": plugins.Option{
			Type:         plugins.OptionTypeBool,
			Description:  "add the rulings to each card description",
			DefaultValue: false,
		},
	}
}

func (p magicPlugin) URLHandlers() []plugins.URLHandler {
	return []plugins.URLHandler{
		plugins.URLHandler{
			BasePath: "https://deckstats.net",
			Regex:    regexp.MustCompile(`^https://deckstats.net/decks/`),
			Handler: func(baseURL string, options map[string]string) ([]*plugins.Deck, error) {
				fileURL, err := url.Parse(baseURL)
				if err != nil {
					return nil, err
				}
				q := fileURL.Query()
				q.Set("include_comments", "1")
				q.Set("export_mtgarena", "1")
				fileURL.RawQuery = q.Encode()

				return handleLink(
					baseURL,
					`//h2[@id='subtitle']`,
					fileURL.String(),
					options,
				)
			},
		},
		plugins.URLHandler{
			BasePath: "https://tappedout.net",
			Regex:    regexp.MustCompile(`^https://tappedout.net/mtg-decks/`),
			Handler: func(baseURL string, options map[string]string) ([]*plugins.Deck, error) {
				fileURL, err := url.Parse(baseURL)
				if err != nil {
					return nil, err
				}
				q := fileURL.Query()
				q.Set("fmt", "dec")
				fileURL.RawQuery = q.Encode()

				return handleLink(
					baseURL,
					`//div[contains(@class,'well')]/h2/text()`,
					fileURL.String(),
					options,
				)
			},
		},
		plugins.URLHandler{
			BasePath: "https://deckbox.org",
			Regex:    regexp.MustCompile(`^https://deckbox.org/sets/`),
			Handler: func(baseURL string, options map[string]string) ([]*plugins.Deck, error) {
				var fileURL string
				if strings.HasSuffix(baseURL, "/") {
					fileURL = baseURL + "export"
				} else {
					fileURL = baseURL + "/export"
				}

				return handleHTMLLink(
					baseURL,
					`//span[@id='deck_built_title']/following-sibling::text()`,
					fileURL,
					options,
				)
			},
		},
		plugins.URLHandler{
			BasePath: "https://www.mtggoldfish.com",
			Regex:    regexp.MustCompile(`^https://www.mtggoldfish.com/(?:archetype|deck)`),
			Handler: func(baseURL string, options map[string]string) ([]*plugins.Deck, error) {
				return handleLinkWithDownloadLink(
					baseURL,
					`//h1[contains(@class,'deck-view-title')]/text()`,
					`//a[contains(text(),'Download')]/@href`,
					"https://www.mtggoldfish.com",
					options,
				)
			},
		},
	}
}

func (p magicPlugin) FileExtHandlers() map[string]plugins.FileHandler {
	return map[string]plugins.FileHandler{
		".dec": fromDeckFile,
		".cod": fromCockatriceDeckFile,
	}
}

func (p magicPlugin) GenericFileHandler() plugins.PathHandler {
	return parseFile
}

func (p magicPlugin) AvailableBacks() map[string]plugins.Back {
	return map[string]plugins.Back{
		plugins.DefaultBackKey: plugins.Back{
			URL:         defaultBackURL,
			Description: "standard paper card back",
		},
		"planechase": plugins.Back{
			URL:         planechaseBackURL,
			Description: "Planechase Plane / Phenomenon card back",
		},
		"archenemy": plugins.Back{
			URL:         archenemyBackURL,
			Description: "Archenemy Scheme card back",
		},
		"m_filler": plugins.Back{
			URL:         mFillerBackURL,
			Description: "filler card back with a white M in the middle",
		},
	}
}

// MagicPlugin is the exported plugin for this package
var MagicPlugin = magicPlugin{
	id:   "mtg",
	name: "Magic",
}
