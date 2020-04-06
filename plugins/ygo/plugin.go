package ygo

import (
	"regexp"

	"deckconverter/plugins"
)

type ygoPlugin struct {
	id   string
	name string
}

func (p ygoPlugin) PluginID() string {
	return p.id
}

func (p ygoPlugin) PluginName() string {
	return p.name
}

func (p ygoPlugin) SupportedLanguages() []string {
	return []string{
		"en",
	}
}

func (p ygoPlugin) AvailableOptions() plugins.Options {
	return plugins.Options{}
}

func (p ygoPlugin) URLHandlers() []plugins.URLHandler {
	return []plugins.URLHandler{
		plugins.URLHandler{
			Regex: regexp.MustCompile(`^https://ygoprodeck.com/`),
			Handler: func(url string, options map[string]string) ([]*plugins.Deck, error) {
				return handleLinkWithYDKFile(
					url,
					ygoproDeckTitleXPath,
					ygoproDeckFileXPath,
					"",
					options,
				)
			},
		},
		plugins.URLHandler{
			Regex: regexp.MustCompile(`^http://yugiohtopdecks.com/deck/`),
			Handler: func(url string, options map[string]string) ([]*plugins.Deck, error) {
				return handleLinkWithYDKFile(
					url,
					yugiohTopDecksTitleXPath,
					yugiohTopDecksFileXPath,
					"http://yugiohtopdecks.com",
					options,
				)
			},
		},
	}
}

func (p ygoPlugin) FileExtHandlers() map[string]plugins.FileHandler {
	return map[string]plugins.FileHandler{
		".ydk": fromYDKFile,
	}
}

func (p ygoPlugin) GenericFileHandler() plugins.PathHandler {
	return parseFile
}

func (p ygoPlugin) AvailableBacks() map[string]plugins.Back {
	return map[string]plugins.Back{
		plugins.DefaultBackKey: plugins.Back{
			URL:         defaultBackURL,
			Description: "standard paper back with no logo",
		},
		"tcg": plugins.Back{
			URL:         tcgBackURL,
			Description: "TCG (Western) paper back",
		},
		"ocg": plugins.Back{
			URL:         ocgBackURL,
			Description: "OCG (Japanese) paper back",
		},
		"anime": plugins.Back{
			URL:         animeBackURL,
			Description: "Paper back used in the anime",
		},
	}
}

// YGOPlugin is the exported plugin for this module
var YGOPlugin = ygoPlugin{
	id:   "ygo",
	name: "Yu-Gi-Oh",
}
