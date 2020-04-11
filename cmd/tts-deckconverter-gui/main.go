package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	dc "deckconverter"
	"deckconverter/log"
	"deckconverter/plugins"
	"deckconverter/tts"
)

const (
	appName         = "TTS Deckconverter GUI"
	customBackLabel = "Custom URL"
)

func showError(win fyne.Window, format string, args ...interface{}) {
	msg := fmt.Errorf(format, args...)
	log.Info(msg)
	dialog.ShowError(msg, win)
}

func handleTarget(target, mode, backURL, outputFolder string, templateMode bool, compact bool, optionWidgets map[string]interface{}, win fyne.Window) {
	log.Infof("Processing %s", target)

	if len(outputFolder) == 0 {
		showError(win, "Output folder is empty")
		return
	}

	if !filepath.IsAbs(outputFolder) {
		showError(win, "The output folder must be an absolute path")
		return
	}

	options := convertOptions(optionWidgets)
	log.Infof("Selected options: %v", options)

	progress := dialog.NewProgressInfinite("Generating", "Generating deck(s)…", win)

	go func() {
		decks, err := dc.Parse(target, mode, options)
		if err != nil {
			progress.Hide()
			showError(win, "Couldn't parse deck(s): %w", err)
			return
		}

		if templateMode {
			err := tts.GenerateTemplates([][]*plugins.Deck{decks}, outputFolder)
			if err != nil {
				progress.Hide()
				showError(win, "Couldn't generate template: %w", err)
				return
			}
		}

		tts.Generate(decks, backURL, outputFolder, !compact)

		result := "Generated the following files in\n" + outputFolder + ":\n"
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
			options[name] = plugins.UncapitalizeString(w.Selected)
		case *widget.Check:
			options[name] = strconv.FormatBool(w.Checked)
		default:
			log.Errorf("Unknown widget type: %s", reflect.TypeOf(w))
		}
	}

	return options
}

func selectedBackURL(backSelect *widget.Select, customBack *widget.Entry, plugin plugins.Plugin) string {
	if backSelect.Selected == customBackLabel {
		return customBack.Text
	}
	for _, back := range plugin.AvailableBacks() {
		if plugins.CapitalizeString(back.Description) == backSelect.Selected {
			return back.URL
		}
	}
	return ""
}

func pluginScreen(win fyne.Window, folderEntry *widget.Entry, templateCheck *widget.Check, compactCheck *widget.Check, plugin plugins.Plugin) fyne.CanvasObject {
	options := plugin.AvailableOptions()

	vbox := widget.NewVBox()

	optionWidgets := make(map[string]interface{})

	for name, option := range options {
		switch option.Type {
		case plugins.OptionTypeEnum:
			vbox.Append(widget.NewLabel(plugins.CapitalizeString(option.Description)))
			radio := widget.NewRadio(plugins.CapitalizeStrings(option.AllowedValues), nil)
			radio.Required = true
			if option.DefaultValue != nil {
				radio.SetSelected(plugins.CapitalizeString(option.DefaultValue.(string)))
			}
			optionWidgets[name] = radio
			vbox.Append(radio)
		case plugins.OptionTypeInt:
			vbox.Append(widget.NewLabel(plugins.CapitalizeString(option.Description)))
			entry := widget.NewEntry()
			entry.SetPlaceHolder(plugins.CapitalizeString(option.DefaultValue.(string)))
			optionWidgets[name] = entry
			vbox.Append(entry)
		case plugins.OptionTypeBool:
			check := widget.NewCheck(plugins.CapitalizeString(option.Description), nil)
			check.Checked = option.DefaultValue.(bool)
			optionWidgets[name] = check
			vbox.Append(check)
		default:
			log.Warnf("Unknown option type: %s", option.Type)
			continue
		}
	}

	vbox.Append(widget.NewLabel("Card back"))

	availableBacks := plugin.AvailableBacks()
	backs := make([]string, 0, len(availableBacks))

	for _, back := range availableBacks {
		backs = append(backs, plugins.CapitalizeString(back.Description))
	}
	backs = append(backs, customBackLabel)

	customBack := widget.NewEntry()
	customBack.Hide()
	lastSelected := plugins.CapitalizeString(availableBacks[plugins.DefaultBackKey].Description)

	backSelect := widget.NewSelect(backs, func(selected string) {
		if selected == customBackLabel {
			customBack.Show()
		} else if lastSelected == customBackLabel {
			customBack.Hide()
		}
		lastSelected = selected
	})
	backSelect.SetSelected(lastSelected)

	vbox.Append(backSelect)
	vbox.Append(customBack)

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
					showError(win, "The URL field is empty")
					return
				}
				handleTarget(
					urlEntry.Text,
					plugin.PluginID(),
					selectedBackURL(backSelect, customBack, plugin),
					folderEntry.Text,
					templateCheck.Checked,
					compactCheck.Checked,
					optionWidgets,
					win,
				)
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
				showError(win, "No file has been selected")
			}
			handleTarget(
				fileEntry.Text,
				plugin.PluginID(),
				selectedBackURL(backSelect, customBack, plugin),
				folderEntry.Text,
				templateCheck.Checked,
				compactCheck.Checked,
				optionWidgets,
				win,
			)
		}),
	)))

	vbox.Append(widget.NewTabContainer(tabItems...))

	return vbox
}

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "enable debug logging")

	flag.Parse()

	var config zap.Config

	if debug {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	} else {
		config = zap.NewProductionConfig()
		config.Encoding = "console"
		config.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
		config.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
		config.EncoderConfig.EncodeCaller = nil
		config.OutputPaths = append(config.OutputPaths, "tts-deckbuilder-gui.log")
	}

	// Skip 1 caller, since all log calls will be done from deckconverter/log
	logger, err := config.Build(zap.AddCallerSkip(2))
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}
	// Don't check for errors since logger.Sync() can sometimes fail
	// even if the logs were properly displayed
	defer logger.Sync()

	log.SetLogger(logger.Sugar())

	// TODO: Remove when upgrading Fyne
	// Temporary fix for OS X (see https://github.com/fyne-io/fyne/issues/824)
	// Manually specify the application theme
	err = os.Setenv("FYNE_THEME", "dark")
	if err != nil {
		log.Errorf("Couldn't set the theme: %v", err.Error())
		os.Exit(1)
	}

	availablePlugins := dc.AvailablePlugins()

	app := app.NewWithID("tts-deckbuilder")

	win := app.NewWindow(appName)
	win.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("Menu",
			fyne.NewMenuItem("About", func() {
				showAboutWindow(app)
			}),
		)), // a quit item will be appended to our first menu
	)
	win.SetMaster()

	folderEntry := widget.NewEntry()
	templateCheck := widget.NewCheck("Generate a template", nil)
	compactCheck := widget.NewCheck("Compact file", nil)

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

		tabItems = append(tabItems, widget.NewTabItem(plugin.PluginName(), pluginScreen(win, folderEntry, templateCheck, compactCheck, plugin)))
	}

	tabs := widget.NewTabContainer(tabItems...)
	tabs.SetTabLocation(widget.TabLocationLeading)

	win.SetContent(
		widget.NewVBox(
			widget.NewHBox(
				widget.NewLabel("Output folder:"),
				folderEntry,
			),
			widget.NewHBox(
				templateCheck,
				compactCheck,
			),
			tabs,
		),
	)

	win.ShowAndRun()
}
