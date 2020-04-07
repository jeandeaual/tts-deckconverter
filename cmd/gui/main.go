package main

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"unicode"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	nativedialog "github.com/sqweek/dialog"
	"go.uber.org/zap"

	dc "deckconverter"
	"deckconverter/log"
	"deckconverter/plugins"
	"deckconverter/tts"
)

func capitalizeString(s string) string {
	a := []rune(s)
	a[0] = unicode.ToUpper(a[0])
	return string(a)
}

func uncapitalizeString(s string) string {
	a := []rune(s)
	a[0] = unicode.ToLower(a[0])
	return string(a)
}

func capitalizeStrings(s []string) []string {
	new := make([]string, len(s))

	for i, el := range s {
		new[i] = capitalizeString(el)
	}

	return new
}

func handleTarget(target, mode, backURL string, optionWidgets map[string]interface{}, win fyne.Window) {
	log.Infof("Processing %s", target)

	options := convertOptions(optionWidgets)
	log.Infof("Selected options: %v", options)

	progress := dialog.NewProgressInfinite("Generating", "Generating deck(s)…", win)

	go func() {
		decks, err := dc.Parse(target, mode, options)
		if err != nil {
			msg := fmt.Errorf("Couldn't parse deck(s): %v", err)
			log.Error(msg)
			dialog.ShowError(msg, win)
		}

		dir, err := os.Getwd()
		if err != nil {
			msg := fmt.Errorf("Couldn't get the current directory: %v", err)
			log.Error(msg)
			dialog.ShowError(msg, win)
		}

		tts.Generate(decks, backURL, dir, true)

		result := "Generated the following files in " + dir + ":\n"
		for _, deck := range decks {
			result += "\n" + deck.Name + ".json"
		}

		progress.Hide()

		dialog.ShowInformation("Success", result, win)
	}()

	progress.Show()
}

func convertOptions(optionWidgets map[string]interface{}) map[string]string {
	options := make(map[string]string)

	for name, optionWidget := range optionWidgets {
		switch w := optionWidget.(type) {
		case *widget.Entry:
			options[name] = w.Text
		case *widget.Radio:
			// TODO: Remove this if statement when upgrading to fyne 1.2.4
			if len(w.Selected) > 0 {
				options[name] = uncapitalizeString(w.Selected)
			}
		case *widget.Check:
			options[name] = strconv.FormatBool(w.Checked)
		default:
			log.Errorf("Unknown widget type: %s", reflect.TypeOf(w))
		}
	}

	return options
}

func pluginScreen(win fyne.Window, plugin plugins.Plugin) fyne.CanvasObject {
	options := plugin.AvailableOptions()

	vbox := widget.NewVBox()

	optionWidgets := make(map[string]interface{})

	for name, option := range options {
		switch option.Type {
		case plugins.OptionTypeEnum:
			label := widget.NewLabel(capitalizeString(option.Description))
			vbox.Append(label)
			radio := widget.NewRadio(capitalizeStrings(option.AllowedValues), nil)
			// TODO: Set radio to Required when fyne 1.2.4 is released
			if option.DefaultValue != nil {
				radio.SetSelected(capitalizeString(option.DefaultValue.(string)))
			}
			optionWidgets[name] = radio
			vbox.Append(radio)
		case plugins.OptionTypeInt:
			label := widget.NewLabel(capitalizeString(option.Description))
			vbox.Append(label)
			entry := widget.NewEntry()
			entry.SetPlaceHolder(capitalizeString(option.DefaultValue.(string)))
			optionWidgets[name] = entry
			vbox.Append(entry)
		case plugins.OptionTypeBool:
			check := widget.NewCheck(capitalizeString(option.Description), nil)
			check.Checked = option.DefaultValue.(bool)
			optionWidgets[name] = check
			vbox.Append(check)
		default:
			log.Warnf("Unknown option type: %s", option.Type)
			continue
		}
	}

	urlEntry := widget.NewEntry()
	fileEntry := widget.NewEntry()
	fileEntry.Disable()

	vbox.Append(
		widget.NewTabContainer(
			widget.NewTabItem("From URL", widget.NewVBox(
				urlEntry,
				widget.NewButtonWithIcon("Generate", theme.ConfirmIcon(), func() {
					if len(urlEntry.Text) == 0 {
						dialog.ShowError(errors.New("The URL field is empty"), win)
					}
					handleTarget(urlEntry.Text, plugin.PluginID(), "https://gamepedia.cursecdn.com/mtgsalvation_gamepedia/thumb/f/f8/Magic_card_back.jpg/250px-Magic_card_back.jpg?version=56c40a91c76ffdbe89867f0bc5172888", optionWidgets, win)
				}),
			)),
			widget.NewTabItem("From File", widget.NewVBox(
				fileEntry,
				widget.NewButton("File…", func() {
					filename, err := nativedialog.File().Load()
					if err == nativedialog.ErrCancelled {
						log.Debug("Cancelled file selection")
						return
					} else if err != nil {
						log.Errorf("Error when selecting file: %s", err)
						return
					}
					log.Debugf("Selected %s", filename)
					fileEntry.SetText(filename)
				}),
				widget.NewButtonWithIcon("Generate", theme.ConfirmIcon(), func() {
					if len(fileEntry.Text) == 0 {
						dialog.ShowError(errors.New("No file has been selected"), win)
					}
					handleTarget(fileEntry.Text, plugin.PluginID(), "https://gamepedia.cursecdn.com/mtgsalvation_gamepedia/thumb/f/f8/Magic_card_back.jpg/250px-Magic_card_back.jpg?version=56c40a91c76ffdbe89867f0bc5172888", optionWidgets, win)
				}),
			)),
		),
	)

	return vbox
}

func main() {
	// Skip 1 caller, since all log calls will be done from deckconverter/log
	logger, err := zap.NewDevelopment(zap.AddCallerSkip(2))
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}
	// Don't check for errors since logger.Sync() can sometimes fail
	// even if the logs were properly displayed
	defer logger.Sync()

	log.SetLogger(logger.Sugar())

	availablePlugins := dc.AvailablePlugins()

	app := app.NewWithID("tts-deckbuilder")

	win := app.NewWindow("Deckbuilder GUI")
	win.SetMaster()

	tabs := widget.NewTabContainer()

	for _, pluginName := range availablePlugins {
		plugin, found := dc.Plugins[pluginName]
		if !found {
			log.Fatalf("Invalid mode: %s", pluginName)
		}

		tabs.Append(widget.NewTabItem(plugin.PluginName(), pluginScreen(win, plugin)))
	}

	tabs.SetTabLocation(widget.TabLocationTop)

	win.SetContent(tabs)
	win.ShowAndRun()
}
