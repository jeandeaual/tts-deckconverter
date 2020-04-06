package plugins

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type OptionType int

const (
	OptionTypeEnum OptionType = iota
	OptionTypeBool
	OptionTypeInt
)

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

type Option struct {
	Type         OptionType
	Description  string
	DefaultValue interface{}
	// AllowedValues should only be set when Type is OptionTypeEnum
	AllowedValues []string
}

type Options map[string]Option

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

type Back struct {
	URL         string
	Description string
}

const DefaultBackKey = "default"

type FileHandler func(io.Reader, string, map[string]string) ([]*Deck, error)
type PathHandler func(string, map[string]string) ([]*Deck, error)

type Plugin interface {
	PluginName() string
	SupportedLanguages() []string
	URLHandlers() []URLHandler
	FileExtHandlers() map[string]FileHandler
	GenericFileHandler() PathHandler
	AvailableOptions() Options
	AvailableBacks() map[string]Back
}

type Template struct {
	URL     string
	NumCols int
	NumRows int
}

type TemplateInfo struct {
	ImageURLCardIDMap map[string]int
	Templates         map[int]*Template
}

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

type Deck struct {
	Name         string
	Cards        []CardInfo
	BackURL      string
	TemplateInfo *TemplateInfo
	CardSize     CardSize
}

type URLHandler struct {
	Regex   *regexp.Regexp
	Handler PathHandler
}
