package plugins

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// OptionType is the type of a plugin option.
type OptionType int

const (
	// OptionTypeEnum represents an option which has a list of possible values.
	OptionTypeEnum OptionType = iota
	// OptionTypeBool represents a boolean option.
	OptionTypeBool
	// OptionTypeInt represents an integer option.
	OptionTypeInt
)

// String representation of an OptionType.
func (ot OptionType) String() string {
	switch ot {
	case OptionTypeEnum:
		return "enum"
	case OptionTypeBool:
		return "bool"
	case OptionTypeInt:
		return "int"
	default:
		return "unknown"
	}
}

// Option of a deckconverter plugin.
type Option struct {
	// Type of the option.
	Type OptionType
	// Description of the option.
	Description string
	// DefaultValue is the value the option is set to if no value is provided
	// by the user.
	DefaultValue interface{}
	// AllowedValues should only be set when Type is OptionTypeEnum
	AllowedValues []string
}

// Options is a map of option IDs to option.
type Options map[string]Option

// ValidateNormalize validates plugin options entered by the user and normalizes
// their values.
func (o Options) ValidateNormalize(options map[string]string) (map[string]interface{}, error) {
	output := make(map[string]interface{})

	for key, value := range options {
		option, found := o[key]
		if !found {
			return output, fmt.Errorf("invalid option: %s", key)
		}

		switch option.Type {
		case OptionTypeBool:
			// Try to convert to bool
			if _, err := strconv.ParseBool(value); err == nil {
				output[key] = true
				continue
			}
			lower := strings.ToLower(value)
			output[key] = lower == "on" || lower == "yes" || lower == "y"
		case OptionTypeInt:
			// Try to convert to int
			parsed, err := strconv.Atoi(value)
			if err != nil {
				return output, fmt.Errorf("couldn't convert option %s value (%s) to int", key, value)
			}
			output[key] = parsed
		case OptionTypeEnum:
			// Try to convert to int
			if option.AllowedValues == nil {
				return output, fmt.Errorf("no allowed values set for option %s", key)
			}
			if IndexOf(value, option.AllowedValues) < 0 {
				return output, fmt.Errorf(
					"invalid value set for option %s: %s (allowed values are %s)",
					key,
					value,
					strings.Join(option.AllowedValues, ", "),
				)
			}
			output[key] = value
		}
	}

	return output, nil
}

// Back represents the back of a card.
type Back struct {
	// URL of the card back.
	URL string
	// Description of the card back.
	Description string
}

// DefaultBackKey if the key use for the default card back
// in Plugin.AvailableBacks.
const DefaultBackKey = "default"

// FileHandler is a function used to parse a deck file for a specific file
// extension.
type FileHandler func(io.Reader, string, map[string]string) ([]*Deck, error)

// PathHandler is a function used to parse deck files.
type PathHandler func(string, map[string]string) ([]*Deck, error)

// URLHandler contains the information and function used for parsing a deck
// located at an URL.
type URLHandler struct {
	// BasePath is the main page of the supported website.
	BasePath string
	// Regex used to recognize supported URLs.
	Regex *regexp.Regexp
	// Handler function used to parse the deck.
	Handler PathHandler
}

// Plugin represents a deckconverted plugin.
type Plugin interface {
	// PluginID returns the ID of the plugin.
	PluginID() string
	// PluginName returns the name of the plugin.
	PluginName() string
	// SupportedLanguages returns the list of languages supported by the plugin.
	SupportedLanguages() []string
	// URLHandlers returns the list of URLs supported by the plugin and their
	// parsing functions.
	URLHandlers() []URLHandler
	// FileExtHandlers returns the list of file extensions supported by the
	// plugins and their parsing functions.
	FileExtHandlers() map[string]FileHandler
	// GenericFileHandler returns the default file handler for the plugin.
	GenericFileHandler() PathHandler
	// AvailableOptions returns the list of options that can be set for the
	// plugin.
	AvailableOptions() Options
	// AvailableBacks lists the default card backs available for the plugin.
	AvailableBacks() map[string]Back
}

// Template represents a TTS file template.
// See https://berserk-games.com/knowledgebase/custom-decks/.
type Template struct {
	URL     string
	NumCols int
	NumRows int
}

// TemplateInfo maps card image URLs to templates.
type TemplateInfo struct {
	// ImageURLCardIDMap is a map between card image URLs and card IDs.
	ImageURLCardIDMap map[string]int
	// Templates is a map of template ID to template.
	Templates map[int]*Template
}

// GetAssociatedTemplate returns the template containing the image of the
// supplied card ID.
func (t *TemplateInfo) GetAssociatedTemplate(cardID int) (*Template, int, error) {
	var templateID int

	if t.Templates == nil {
		return nil, templateID, errors.New("no template found")
	}

	templateID = cardID / 100

	template, found := t.Templates[templateID]
	if !found {
		return nil, templateID, fmt.Errorf("template %d not found", templateID)
	}

	return template, templateID, nil
}

// CardInfo contains the information about a card used to build a TTS deck.
type CardInfo struct {
	// Name of the card
	Name string
	// Description of the card
	Description string
	// ImageURL is the URL of the card image
	ImageURL string
	// Count is the amount of this card in the current deck
	Count int
	// AlternativeState is used for double-faced cards (transforms and melds
	// in Magic). The back of the card will be represented as a second state
	// of the card object.
	AlternativeState *CardInfo
	// Oversized card
	// Used for plane, scheme or meld results in MTG
	Oversized bool
}

// CardSize is the size format of a card
type CardSize int

const (
	// CardSizeStandard is the size of a Magic or Pokémon card (Poker size)
	CardSizeStandard CardSize = iota
	// CardSizeSmall is the size of a Yu-Gi-Oh or Cardfight!! Vanguard card
	CardSizeSmall
)

// Deck contains the information about a deck used to build it in TTS.
type Deck struct {
	Name         string
	Cards        []CardInfo
	BackURL      string
	TemplateInfo *TemplateInfo
	CardSize     CardSize
}
