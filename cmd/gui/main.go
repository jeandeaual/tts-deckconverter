package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strconv"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"go.uber.org/zap"

	dc "deckconverter"
	"deckconverter/log"
	"deckconverter/plugins"
	"deckconverter/tts"
)

func handleTarget(target, mode, backURL, folder string, optionWidgets map[string]interface{}, win fyne.Window) {
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
			return
		}

		tts.Generate(decks, backURL, folder, true)

		result := "Generated the following files in " + folder + ":\n"
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
			options[name] = uncapitalizeString(w.Selected)
		case *widget.Check:
			options[name] = strconv.FormatBool(w.Checked)
		default:
			log.Errorf("Unknown widget type: %s", reflect.TypeOf(w))
		}
	}

	return options
}

func pluginScreen(win fyne.Window, folderEntry *widget.Entry, plugin plugins.Plugin) fyne.CanvasObject {
	options := plugin.AvailableOptions()

	vbox := widget.NewVBox()

	optionWidgets := make(map[string]interface{})

	for name, option := range options {
		switch option.Type {
		case plugins.OptionTypeEnum:
			label := widget.NewLabel(capitalizeString(option.Description))
			vbox.Append(label)
			radio := widget.NewRadio(capitalizeStrings(option.AllowedValues), nil)
			radio.Required = true
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

	tabItems := make([]*widget.TabItem, 0, 2)

	urlEntry := widget.NewEntry()
	fileEntry := widget.NewEntry()
	fileEntry.Disable()

	if len(plugin.URLHandlers()) > 0 {
		supportedUrls := widget.NewVBox(widget.NewLabel("Supported URLs:"))

		for _, urlHandler := range plugin.URLHandlers() {
			u, err := url.Parse(urlHandler.BasePath)
			if err != nil {
				log.Errorf("Invalid URL found for plugin %s: %v", plugin.PluginID, err)
				continue
			}
			supportedUrls.Append(widget.NewHyperlink(urlHandler.BasePath, u))
		}

		tabItems = append(tabItems, widget.NewTabItem("From URL", widget.NewVBox(
			urlEntry,
			widget.NewButtonWithIcon("Generate", theme.ConfirmIcon(), func() {
				if len(urlEntry.Text) == 0 {
					dialog.ShowError(errors.New("The URL field is empty"), win)
				}
				handleTarget(urlEntry.Text, plugin.PluginID(), "https://gamepedia.cursecdn.com/mtgsalvation_gamepedia/thumb/f/f8/Magic_card_back.jpg/250px-Magic_card_back.jpg?version=56c40a91c76ffdbe89867f0bc5172888", folderEntry.Text, optionWidgets, win)
			}),
			supportedUrls,
		)))
	}

	tabItems = append(tabItems, widget.NewTabItem("From File", widget.NewVBox(
		fileEntry,
		widget.NewButton("File…", func() {
			dialog.ShowFileOpen(
				func(file string) {
					if len(file) == 0 {
						// Cancelled
						return
					}
					log.Infof("Selected %s", file)
					fileEntry.SetText(file)
				},
				win,
			)
		}),
		widget.NewButtonWithIcon("Generate", theme.ConfirmIcon(), func() {
			if len(fileEntry.Text) == 0 {
				dialog.ShowError(errors.New("No file has been selected"), win)
			}
			handleTarget(fileEntry.Text, plugin.PluginID(), "https://gamepedia.cursecdn.com/mtgsalvation_gamepedia/thumb/f/f8/Magic_card_back.jpg/250px-Magic_card_back.jpg?version=56c40a91c76ffdbe89867f0bc5172888", folderEntry.Text, optionWidgets, win)
		}),
	)))

	vbox.Append(widget.NewTabContainer(tabItems...))

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

	// TODO: Remove then upgrading Fyne
	// Temporary fix for OS X (see https://github.com/fyne-io/fyne/issues/824)
	// Manually specify the application theme
	err = os.Setenv("FYNE_THEME", "dark")
	if err != nil {
		log.Errorf("Couldn't set the theme: %v", err.Error())
		os.Exit(1)
	}

	availablePlugins := dc.AvailablePlugins()

	app := app.NewWithID("tts-deckbuilder")

	win := app.NewWindow("Deckbuilder GUI")
	win.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("Menu",
			fyne.NewMenuItem("About", func() {
				aboutWindow := app.NewWindow("About")

				aboutMsg := "Built with Go version " + getGoVersion()
				fyneVersion, err := getModuleVersion("fyne.io/fyne")
				if err != nil {
					log.Error(err)
				} else {
					aboutMsg += " and Fyne version " + fyneVersion
				}

				licenseLabel := widget.NewLabel(aboutMsg)

				okButton := widget.NewButton("OK", func() {
					aboutWindow.Close()
				})

				buttons := fyne.NewContainerWithLayout(layout.NewHBoxLayout(), layout.NewSpacer(), okButton, layout.NewSpacer())
				paragraphContainer := fyne.NewContainerWithLayout(layout.NewVBoxLayout(), licenseLabel)
				content := fyne.NewContainerWithLayout(layout.NewVBoxLayout(), paragraphContainer, layout.NewSpacer(), buttons)

				aboutWindow.SetContent(content)
				aboutWindow.Show()
			}),
		)), // a quit item will be appended to our first menu
	)
	win.SetMaster()

	folderEntry := widget.NewEntry()

	chestPath, err := tts.FindChestPath()
	if err == nil {
		folderEntry.SetText(chestPath)
	} else {
		log.Debugf("Couldn't find chest path: %v", err)
		currentDir, err := os.Getwd()
		if err == nil {
			folderEntry.SetText(currentDir)
		} else {
			log.Errorf("Couldn't get the working directory: %v", err)
		}
	}

	tabItems := make([]*widget.TabItem, 0, len(availablePlugins))

	for _, pluginName := range availablePlugins {
		plugin, found := dc.Plugins[pluginName]
		if !found {
			log.Fatalf("Invalid mode: %s", pluginName)
		}

		tabItems = append(tabItems, widget.NewTabItem(plugin.PluginName(), pluginScreen(win, folderEntry, plugin)))
	}

	tabs := widget.NewTabContainer(tabItems...)
	tabs.SetTabLocation(widget.TabLocationLeading)

	win.SetContent(
		widget.NewVBox(
			widget.NewHBox(
				widget.NewLabel("Output Folder:"),
				folderEntry,
			),
			tabs,
		),
	)

	win.ShowAndRun()
}
