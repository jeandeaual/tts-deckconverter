package vanguard

import (
	"regexp"

	"github.com/jeandeaual/tts-deckconverter/plugins"
)

type vanguardPlugin struct {
	id   string
	name string
}

func (v vanguardPlugin) PluginID() string {
	return v.id
}

func (v vanguardPlugin) PluginName() string {
	return v.name
}

func (v vanguardPlugin) AvailableOptions() plugins.Options {
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
	}
}

func (v vanguardPlugin) URLHandlers() []plugins.URLHandler {
	return []plugins.URLHandler{
		{
			BasePath: "https://en.cf-vanguard.com",
			Regex:    regexp.MustCompile(`^https://en.cf-vanguard.com/deckrecipe/detail/`),
			Handler:  handleENCFVanguardLink,
		},
		{
			BasePath: "https://cf-vanguard.com",
			Regex:    regexp.MustCompile(`^https://cf-vanguard.com/deckrecipe/(:?detail|events)/`),
			Handler:  handleCFVanguardLink,
		},
	}
}

func (v vanguardPlugin) FileExtHandlers() map[string]plugins.FileHandler {
	return map[string]plugins.FileHandler{}
}

func (v vanguardPlugin) GenericFileHandler() plugins.PathHandler {
	return parseFile
}

func (v vanguardPlugin) AvailableBacks() map[string]plugins.Back {
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
