package ygo

import (
	"fmt"
	"regexp"

	"github.com/antchfx/htmlquery"
	"github.com/jeandeaual/tts-deckconverter/plugins"
	"github.com/jeandeaual/tts-deckconverter/plugins/ygo/api"
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

func (p ygoPlugin) AvailableOptions() plugins.Options {
	return plugins.Options{
		"format": plugins.Option{
			Type:        plugins.OptionTypeEnum,
			Description: "duel format",
			AllowedValues: []string{
				string(api.FormatStandard),
				string(api.FormatRushDuel),
			},
			DefaultValue: string(api.FormatStandard),
		},
	}
}

func (p ygoPlugin) URLHandlers() []plugins.URLHandler {
	return []plugins.URLHandler{
		{
			BasePath: "https://ygoprodeck.com",
			Regex:    regexp.MustCompile(`^https://ygoprodeck\.com/`),
			Handler: func(url string, options map[string]string) ([]*plugins.Deck, error) {
				doc, err := htmlquery.LoadURL(url)
				if err != nil {
					return nil, fmt.Errorf("couldn't query %s: %w", url, err)
				}

				deckTypeTab := htmlquery.FindOne(
					doc,
					`//td/b[normalize-space(text())='Deck Type:']/parent::node()/following-sibling::td`,
				)
				if deckTypeTab != nil {
					switch htmlquery.InnerText(deckTypeTab) {
					case "Rush Duel Decks":
						options["format"] = string(api.FormatRushDuel)
					}
				}

				return handleLinkWithYDKFile(
					url,
					doc,
					ygoproDeckTitleXPath,
					ygoproDeckFileXPath,
					"",
					options,
				)
			},
		},
		{
			BasePath: "https://yugiohtopdecks.com",
			Regex:    regexp.MustCompile(`^https://yugiohtopdecks\.com/deck/`),
			Handler: func(url string, options map[string]string) ([]*plugins.Deck, error) {
				doc, err := htmlquery.LoadURL(url)
				if err != nil {
					return nil, fmt.Errorf("couldn't query %s: %w", url, err)
				}

				return handleLinkWithYDKFile(
					url,
					doc,
					yugiohTopDecksTitleXPath,
					yugiohTopDecksFileXPath,
					"https://yugiohtopdecks.com",
					options,
				)
			},
		},
		{
			BasePath: "https://yugioh.fandom.com",
			Regex:    regexp.MustCompile(`^https://yugioh.fandom.com/wiki/`),
			Handler:  handleYGOWikiLink,
		},
	}
}

func (p ygoPlugin) FileExtHandlers() map[string]plugins.FileHandler {
	return map[string]plugins.FileHandler{
		".ydk": fromYDKFile,
	}
}

func (p ygoPlugin) DeckTypeHandlers() map[string]plugins.FileHandler {
	return map[string]plugins.FileHandler{
		"YGOPRODeck": fromYDKFile,
	}
}

func (p ygoPlugin) GenericFileHandler() plugins.FileHandler {
	return fromDeckFile
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
