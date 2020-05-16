package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	dc "github.com/jeandeaual/tts-deckconverter"
	"github.com/jeandeaual/tts-deckconverter/plugins"
	"github.com/jeandeaual/tts-deckconverter/tts/upload"
)

type options map[string]string

func (o *options) String() string {
	options := make([]string, 0, len(*o))

	for k, v := range *o {
		options = append(options, k+"="+v)
	}

	return strings.Join(options, ",")
}

func (o *options) Set(value string) error {
	kv := strings.Split(value, "=")

	if len(kv) != 2 {
		return errors.New("invalid option value: " + value)
	}

	k := kv[0]
	v := kv[1]

	(*o)[k] = v

	return nil
}

func getAvailableOptions(pluginNames []string) string {
	var sb strings.Builder

	for _, pluginName := range pluginNames {
		plugin, found := dc.Plugins[pluginName]
		if !found {
			fmt.Fprintf(os.Stderr, "Invalid mode: %s\n", pluginName)
			flag.Usage()
			os.Exit(1)
		}

		sb.WriteString("\n")
		sb.WriteString(pluginName)
		sb.WriteString(":")

		options := plugin.AvailableOptions()

		if len(options) == 0 {
			sb.WriteString(" no option available")
			continue
		}

		optionKeys := make([]string, 0, len(options))
		for key := range options {
			optionKeys = append(optionKeys, key)
		}
		sort.Strings(optionKeys)

		for _, key := range optionKeys {
			option := options[key]

			sb.WriteString("\n\t")
			sb.WriteString(key)
			sb.WriteString(" (")
			sb.WriteString(option.Type.String())
			sb.WriteString("): ")
			sb.WriteString(option.Description)

			if option.DefaultValue != nil {
				sb.WriteString(" (default: ")
				sb.WriteString(fmt.Sprintf("%v", option.DefaultValue))
				sb.WriteString(")")
			}
		}
	}

	return sb.String()
}

func getAvailableDeckFormats(pluginNames []string) string {
	var sb strings.Builder

	for _, pluginName := range pluginNames {
		plugin, found := dc.Plugins[pluginName]
		if !found {
			fmt.Fprintf(os.Stderr, "Invalid mode: %s\n", pluginName)
			flag.Usage()
			os.Exit(1)
		}

		sb.WriteString("\n")
		sb.WriteString(pluginName)
		sb.WriteString(":\n")

		deckTypeHandlers := plugin.DeckTypeHandlers()

		sb.WriteString("\tgeneric")

		if len(deckTypeHandlers) == 0 {
			continue
		}

		deckTypes := make([]string, 0, len(deckTypeHandlers))
		for deckType := range deckTypeHandlers {
			deckTypes = append(deckTypes, deckType)
		}
		sort.Strings(deckTypes)

		for _, deckType := range deckTypes {
			sb.WriteString("\n\t")
			sb.WriteString(deckType)
		}
	}

	return sb.String()
}

func getAvailableBacks(pluginNames []string) string {
	var sb strings.Builder

	for _, pluginName := range pluginNames {
		plugin, found := dc.Plugins[pluginName]
		if !found {
			fmt.Fprintf(os.Stderr, "Invalid mode: %s\n", pluginName)
			flag.Usage()
			os.Exit(1)
		}

		sb.WriteString("\n")
		sb.WriteString(pluginName)
		sb.WriteString(":")

		backs := plugin.AvailableBacks()

		if len(backs) == 0 {
			sb.WriteString(" no card back available")
			continue
		}

		backKeys := make([]string, 0, len(backs))
		for key := range backs {
			if key != plugins.DefaultBackKey {
				backKeys = append(backKeys, key)
			}
		}
		sort.Strings(backKeys)

		// Make sure "default" is first
		if _, found := backs[plugins.DefaultBackKey]; found {
			backKeys = append([]string{plugins.DefaultBackKey}, backKeys...)
		}

		for _, key := range backKeys {
			back := backs[key]

			sb.WriteString("\n\t")
			sb.WriteString(key)
			sb.WriteString(": ")
			sb.WriteString(back.Description)
		}
	}

	return sb.String()
}

func getAvailableUploaders() string {
	var sb strings.Builder

	uploaderKeys := make([]string, 0, len(upload.TemplateUploaders))
	for key := range upload.TemplateUploaders {
		if key != plugins.DefaultBackKey {
			uploaderKeys = append(uploaderKeys, key)
		}
	}
	sort.Strings(uploaderKeys)

	for _, key := range uploaderKeys {
		uploader := upload.TemplateUploaders[key]

		sb.WriteString("\n")
		sb.WriteString("\t")
		sb.WriteString(key)
		sb.WriteString(": ")
		sb.WriteString((*uploader).UploaderDescription())
	}

	return sb.String()
}
