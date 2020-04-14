package ygo

import (
	"regexp"

	"github.com/jeandeaual/tts-deckconverter/plugins"
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
		{
			BasePath: "https://ygoprodeck.com",
			Regex:    regexp.MustCompile(`^https://ygoprodeck.com/`),
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
		{
			BasePath: "https://yugiohtopdecks.com",
			Regex:    regexp.MustCompile(`^https://yugiohtopdecks.com/deck/`),
			Handler: func(url string, options map[string]string) ([]*plugins.Deck, error) {
				return handleLinkWithYDKFile(
					url,
					yugiohTopDecksTitleXPath,
					yugiohTopDecksFileXPath,
					"https://yugiohtopdecks.com",
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
	// Card backs created using https://www.deviantart.com/holycrapwhitedragon/art/Yu-Gi-Oh-Back-Card-Template-695173962 (© 2017 - 2020 HolyCrapWhiteDragon)
	return map[string]plugins.Back{
		plugins.DefaultBackKey: {
			URL:         defaultBackURL,
			Description: "standard paper back with no logo",
		},
		"tcg": {
			URL:         tcgBackURL,
			Description: "TCG (Western) paper back",
		},
		"ocg": {
			URL:         ocgBackURL,
			Description: "OCG (Japanese) paper back",
		},
		"anime": {
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
