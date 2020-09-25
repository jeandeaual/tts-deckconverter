package vanguard

import (
	"regexp"

	"github.com/jeandeaual/tts-deckconverter/plugins"
)

type vanguardPlugin struct {
	id   string
	name string
}

func (p vanguardPlugin) PluginID() string {
	return p.id
}

func (p vanguardPlugin) PluginName() string {
	return p.name
}

func (p vanguardPlugin) AvailableOptions() plugins.Options {
	return plugins.Options{
		"lang": plugins.Option{
			Type:        plugins.OptionTypeEnum,
			Description: "Language of the cards",
			AllowedValues: []string{
				"en",
				"ja",
			},
			DefaultValue: "en",
		},
		"vanguard-first": plugins.Option{
			Type:         plugins.OptionTypeBool,
			Description:  "Put the first vanguard on top of the deck",
			DefaultValue: true,
		},
		"prefer-premium": plugins.Option{
			Type:         plugins.OptionTypeBool,
			Description:  "Use premium cards instead of V series cards if available",
			DefaultValue: false,
		},
	}
}

func (p vanguardPlugin) URLHandlers() []plugins.URLHandler {
	return []plugins.URLHandler{
		{
			BasePath: "https://en.cf-vanguard.com/deckrecipe",
			Regex:    regexp.MustCompile(`^https://en.cf-vanguard.com/deckrecipe/detail/`),
			Handler:  handleENCFVanguardLink,
		},
		{
			BasePath: "https://cf-vanguard.com/deckrecipe",
			Regex:    regexp.MustCompile(`^https://cf-vanguard.com/deckrecipe/(:?detail|events)/`),
			Handler:  handleCFVanguardLink,
		},
		{
			BasePath: "https://cardfight.fandom.com",
			Regex:    regexp.MustCompile(`^https://cardfight.fandom.com/wiki/`),
			Handler:  handleCFVWikiLink,
		},
	}
}

func (p vanguardPlugin) FileExtHandlers() map[string]plugins.FileHandler {
	return map[string]plugins.FileHandler{}
}

func (p vanguardPlugin) DeckTypeHandlers() map[string]plugins.DeckType {
	return map[string]plugins.DeckType{}
}

func (p vanguardPlugin) GenericFileHandler() plugins.DeckType {
	return plugins.DeckType{
		FileHandler: fromDeckFile,
		Example: `4x Alfred Early
4x Stardrive Dragon
4x Knight of Silence, Gallatin
4x Blaster Blade
3x Sage of the Arts, Jauron
3x Wingal`,
	}
}

func (p vanguardPlugin) AvailableBacks() map[string]plugins.Back {
	return map[string]plugins.Back{
		plugins.DefaultBackKey: {
			URL:         defaultBackURL,
			Description: "standard paper card back",
		},
	}
}

// VanguardPlugin is the exported plugin for this package
var VanguardPlugin = vanguardPlugin{
	id:   "cfv",
	name: "Vanguard",
}
