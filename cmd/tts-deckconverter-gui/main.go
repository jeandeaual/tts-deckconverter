package main

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/driver/desktop"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	dc "github.com/jeandeaual/tts-deckconverter"
	"github.com/jeandeaual/tts-deckconverter/log"
	"github.com/jeandeaual/tts-deckconverter/plugins"
	"github.com/jeandeaual/tts-deckconverter/tts"
	"github.com/jeandeaual/tts-deckconverter/tts/upload"
)

const (
	appName         = "TTS Deckconverter GUI"
	appID           = "tts-deckconverter-gui"
	customBackLabel = "Custom URL"
)

func checkDir(path string) error {
	if stat, err := os.Stat(path); os.IsNotExist(err) {
		return err
	} else if err != nil {
		return fmt.Errorf("invalid path %s: %w", path, err)
	} else if !stat.IsDir() {
		return fmt.Errorf("output folder %s is not a directory", path)
	}

	return nil
}

func showErrorf(win fyne.Window, format string, args ...interface{}) {
	msg := fmt.Errorf(format, args...)
	log.Info(msg)
	dialog.ShowError(msg, win)
}

func handleTarget(
	target string,
	mode string,
	backURL string,
	outputFolder string,
	uploader *upload.TemplateUploader,
	compact bool,
	optionWidgets map[string]interface{},
	win fyne.Window,
) {
	options := convertOptions(optionWidgets)
	log.Infof("Selected options: %v", options)

	progress := dialog.NewProgressInfinite("Generating", "Generating deck(s)…", win)

	go func() {
		decks, err := dc.Parse(target, mode, options)
		if err != nil {
			progress.Hide()
			showErrorf(win, "Couldn't parse deck(s): %w", err)
			return
		}
		if len(decks) == 0 {
			progress.Hide()
			showErrorf(win, "Couldn't parse deck(s)")
			return
		}

		if uploader != nil {
			errs := tts.GenerateTemplates([][]*plugins.Deck{decks}, outputFolder, *uploader)
			if len(errs) > 0 {
				progress.Hide()
				uploadSizeErrsOnly := true
				msg := "Couldn't generate template(s):\n"
				for _, err := range errs {
					errorMsg := plugins.CapitalizeString(err.Error())
					log.Info(errorMsg)
					msg += "\n" + errorMsg
					if !errors.Is(err, upload.ErrUploadSize) {
						uploadSizeErrsOnly = false
					}
				}
				dialog.ShowError(errors.New(msg), win)
				// If the only error we got was that the template was too big to be uploaded, continue
				// The user will be able to upload the template manually later on
				if !uploadSizeErrsOnly {
					return
				}
			}
		}

		errs := tts.Generate(decks, backURL, outputFolder, !compact)
		if len(errs) > 0 {
			progress.Hide()
			msg := "Couldn't generate deck(s):\n"
			for _, err := range errs {
				errorMsg := plugins.CapitalizeString(err.Error())
				log.Info(msg)
				msg += "\n" + errorMsg
			}
			dialog.ShowError(errors.New(msg), win)
			return
		}

		result := "Generated the following files in\n" + outputFolder + ":\n"
		for _, deck := range decks {
			result += "\n" + deck.Name + ".json"
		}

		progress.Hide()

		dialog.ShowInformation("Success", result, win)
	}()

	progress.Show()
}

func handleText(
	text string,
	deckName string,
	handler plugins.FileHandler,
	backURL string,
	outputFolder string,
	uploader *upload.TemplateUploader,
	compact bool,
	optionWidgets map[string]interface{},
	win fyne.Window,
) {
	options := convertOptions(optionWidgets)
	log.Infof("Selected options: %v", options)

	progress := dialog.NewProgressInfinite("Generating", "Generating deck(s)…", win)

	go func() {
		decks, err := handler(strings.NewReader(text), deckName, options)
		if err != nil {
			progress.Hide()
			showErrorf(win, "Couldn't parse deck: %w", err)
			return
		}
		if len(decks) == 0 {
			progress.Hide()
			showErrorf(win, "Couldn't parse deck")
			return
		}

		if uploader != nil {
			errs := tts.GenerateTemplates([][]*plugins.Deck{decks}, outputFolder, *uploader)
			if len(errs) > 0 {
				progress.Hide()
				uploadSizeErrsOnly := true
				msg := "Couldn't generate template(s):\n"
				for _, err := range errs {
					errorMsg := plugins.CapitalizeString(err.Error())
					log.Info(errorMsg)
					msg += "\n" + errorMsg
					if !errors.Is(err, upload.ErrUploadSize) {
						uploadSizeErrsOnly = false
					}
				}
				dialog.ShowError(errors.New(msg), win)
				// If the only error we got was that the template was too big to be uploaded, continue
				// The user will be able to upload the template manually later on
				if !uploadSizeErrsOnly {
					return
				}
			}
		}

		errs := tts.Generate(decks, backURL, outputFolder, !compact)
		if len(errs) > 0 {
			progress.Hide()
			msg := "Couldn't generate deck:\n"
			for _, err := range errs {
				errorMsg := plugins.CapitalizeString(err.Error())
				log.Info(msg)
				msg += "\n" + errorMsg
			}
			dialog.ShowError(errors.New(msg), win)
			return
		}

		result := "Generated the following files in\n" + outputFolder + ":\n"
		for _, deck := range decks {
			result += "\n" + deck.Name + ".json"
		}

		progress.Hide()

		dialog.ShowInformation("Success", result, win)
	}()

	progress.Show()
}

func checkInput(target, mode, backURL, outputFolder string, callback func(), win fyne.Window) {
	log.Infof("Processing %s", target)

	if len(outputFolder) == 0 {
		showErrorf(win, "Output folder is empty")
		return
	}

	if !filepath.IsAbs(outputFolder) {
		showErrorf(win, "The output folder must be an absolute path")
		return
	}

	err := checkDir(outputFolder)
	if os.IsNotExist(err) {
		if plugins.CheckInvalidFolderName(outputFolder) {
			log.Info("Invalid folder name: %s", outputFolder)
			dialog.ShowError(fmt.Errorf("Invalid folder name:\n%s", outputFolder), win)
			return
		}
		dialog.ShowConfirm(
			"Folder creation",
			fmt.Sprintf("This will create the following folder:\n%s\n\nContinue?", outputFolder),
			func(ok bool) {
				if !ok {
					log.Debug("Folder creation cancelled by user")
					return
				}

				log.Infof("Output folder %s doesn't exist, creating it", outputFolder)
				err = os.MkdirAll(outputFolder, 0o755)
				if err != nil {
					showErrorf(win, "Couldn't create folder %s: %w", outputFolder, err)
					return
				}

				callback()
			},
			win,
		)
		return
	} else if err != nil {
		showErrorf(win, plugins.CapitalizeString(err.Error()))
		return
	}

	callback()
}

func convertOptions(optionWidgets map[string]interface{}) map[string]string {
	options := make(map[string]string)

	for name, optionWidget := range optionWidgets {
		switch w := optionWidget.(type) {
		case *widget.Entry:
			options[name] = w.Text
		case *widget.Radio:
			options[name] = w.Selected
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

func pluginScreen(win fyne.Window, folderEntry *widget.Entry, uploaderSelect *widget.Select, compactCheck *widget.Check, plugin plugins.Plugin) fyne.CanvasObject {
	options := plugin.AvailableOptions()
	optionsVBox := widget.NewVBox()
	optionWidgets := make(map[string]interface{})

	for name, option := range options {
		switch option.Type {
		case plugins.OptionTypeEnum:
			optionsVBox.Append(widget.NewLabel(plugins.CapitalizeString(option.Description)))
			radio := widget.NewRadio(option.AllowedValues, nil)
			radio.Required = true
			if option.DefaultValue != nil {
				radio.SetSelected(option.DefaultValue.(string))
			}
			optionWidgets[name] = radio
			optionsVBox.Append(radio)
		case plugins.OptionTypeInt:
			optionsVBox.Append(widget.NewLabel(plugins.CapitalizeString(option.Description)))
			entry := widget.NewEntry()
			entry.SetPlaceHolder(plugins.CapitalizeString(option.DefaultValue.(string)))
			optionWidgets[name] = entry
			optionsVBox.Append(entry)
		case plugins.OptionTypeBool:
			check := widget.NewCheck(plugins.CapitalizeString(option.Description), nil)
			check.Checked = option.DefaultValue.(bool)
			optionWidgets[name] = check
			optionsVBox.Append(check)
		default:
			log.Warnf("Unknown option type: %s", option.Type)
			continue
		}
	}

	optionsVBox.Append(widget.NewLabel("Card back"))

	availableBacks := plugin.AvailableBacks()
	backs := make([]string, 0, len(availableBacks))

	for _, back := range availableBacks {
		backs = append(backs, plugins.CapitalizeString(back.Description))
	}
	backs = append(backs, customBackLabel)

	customBack := widget.NewEntry()
	customBack.Hide()
	lastSelected := plugins.CapitalizeString(availableBacks[plugins.DefaultBackKey].Description)

	backPreview := widget.NewHyperlink("Preview", nil)
	_ = backPreview.SetURLFromString(availableBacks[plugins.DefaultBackKey].URL)
	backSelect := widget.NewSelect(backs, func(selected string) {
		if selected == customBackLabel {
			customBack.Show()
			backPreview.Hide()
		} else if lastSelected == customBackLabel {
			customBack.Hide()
			backPreview.Show()
		}
		if selected != customBackLabel {
			// Update the preview link
			var backURL string
			for _, back := range plugin.AvailableBacks() {
				if plugins.CapitalizeString(back.Description) == selected {
					backURL = back.URL
				}
			}
			err := backPreview.SetURLFromString(backURL)
			if err != nil {
				log.Errorf("Invalid URL found for back %s: %v", backURL, err)
			}
		}
		lastSelected = selected
	})
	backSelect.SetSelected(lastSelected)

	optionsVBox.Append(fyne.NewContainerWithLayout(
		layout.NewBorderLayout(
			nil,
			nil,
			nil,
			backPreview,
		),
		backSelect,
		backPreview,
	))
	optionsVBox.Append(customBack)

	tabItems := make([]*widget.TabItem, 0, 2)

	urlEntry := widget.NewEntry()

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
			widget.NewHBox(
				widget.NewButtonWithIcon("Generate", theme.ConfirmIcon(), func() {
					if len(urlEntry.Text) == 0 {
						showErrorf(win, "The URL field is empty")
						return
					}

					var selectedUploader *upload.TemplateUploader
					for _, uploader := range upload.TemplateUploaders {
						if (*uploader).UploaderName() == uploaderSelect.Selected {
							selectedUploader = uploader
						}
					}

					target := urlEntry.Text
					mode := plugin.PluginID()
					back := selectedBackURL(backSelect, customBack, plugin)
					output := folderEntry.Text

					checkInput(
						target,
						mode,
						back,
						output,
						func() {
							handleTarget(target, mode, back, output, selectedUploader, compactCheck.Checked, optionWidgets, win)
						},
						win,
					)
				}),
			),
			supportedUrls,
		)))
	}

	fileEntry := widget.NewEntry()
	fileEntry.Disable()

	tabItems = append(tabItems, widget.NewTabItem("From file", widget.NewVBox(
		fileEntry,
		widget.NewHBox(
			widget.NewButtonWithIcon("File…", theme.DocumentSaveIcon(), func() {
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
					showErrorf(win, "No file has been selected")
					return
				}

				var selectedUploader *upload.TemplateUploader
				for _, uploader := range upload.TemplateUploaders {
					if (*uploader).UploaderName() == uploaderSelect.Selected {
						selectedUploader = uploader
					}
				}

				target := fileEntry.Text
				mode := plugin.PluginID()
				back := selectedBackURL(backSelect, customBack, plugin)
				output := folderEntry.Text

				checkInput(
					target,
					mode,
					back,
					output,
					func() {
						handleTarget(target, mode, back, output, selectedUploader, compactCheck.Checked, optionWidgets, win)
					},
					win,
				)
			}),
		),
	)))

	textInput := widget.NewMultiLineEntry()
	deckNameInput := widget.NewEntry()
	deckTypes := make([]string, 0, len(plugin.DeckTypeHandlers())+1)
	deckTypes = append(deckTypes, "Generic")
	for deckType := range plugin.DeckTypeHandlers() {
		deckTypes = append(deckTypes, deckType)
	}
	deckTypeSelect := widget.NewSelect(deckTypes, nil)
	deckTypeSelect.SetSelected("Generic")

	textInputScrollContainer := widget.NewVScrollContainer(
		textInput,
	)
	textInputButtons := widget.NewVBox(
		widget.NewHBox(
			widget.NewButtonWithIcon("Paste", theme.ContentPasteIcon(), func() {
				textInput.SetText(win.Clipboard().Content())
			}),
			widget.NewButtonWithIcon("Clear", theme.ContentClearIcon(), func() {
				textInput.SetText("")
			}),
		),
		widget.NewLabel("Deck name:"),
		deckNameInput,
		widget.NewLabel("Deck type:"),
		deckTypeSelect,
	)

	textInputButtons.Append(
		widget.NewHBox(
			widget.NewButtonWithIcon("Generate", theme.ConfirmIcon(), func() {
				if len(textInput.Text) == 0 {
					showErrorf(win, "The input is empty")
					return
				}

				if len(deckNameInput.Text) == 0 {
					showErrorf(win, "No deck name has been provided")
					return
				}

				var selectedUploader *upload.TemplateUploader
				for _, uploader := range upload.TemplateUploaders {
					if (*uploader).UploaderName() == uploaderSelect.Selected {
						selectedUploader = uploader
					}
				}

				text := textInput.Text
				deckName := deckNameInput.Text
				handler := plugin.GenericFileHandler()
				if deckTypeHandler, found := plugin.DeckTypeHandlers()[deckTypeSelect.Selected]; found {
					handler = deckTypeHandler
				}
				mode := plugin.PluginID()
				back := selectedBackURL(backSelect, customBack, plugin)
				output := folderEntry.Text

				checkInput(
					text,
					mode,
					back,
					output,
					func() {
						handleText(text, deckName, handler, back, output, selectedUploader, compactCheck.Checked, optionWidgets, win)
					},
					win,
				)
			}),
		),
	)

	tabItems = append(tabItems, widget.NewTabItem("From text", fyne.NewContainerWithLayout(
		layout.NewBorderLayout(
			nil,
			textInputButtons,
			nil,
			nil,
		),
		textInputScrollContainer,
		textInputButtons,
	)))

	tabContainer := widget.NewTabContainer(tabItems...)

	return fyne.NewContainerWithLayout(
		layout.NewBorderLayout(
			optionsVBox,
			nil,
			nil,
			nil,
		),
		optionsVBox,
		tabContainer,
	)
}

func main() {
	var (
		debug    bool
		setTheme string
	)

	flag.BoolVar(&debug, "debug", false, "enable debug logging")
	// TODO: Use an empty default value when upgrading Fyne
	flag.StringVar(&setTheme, "theme", "dark", "application theme (\"light\" or \"dark\")")

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
		config.OutputPaths = append(config.OutputPaths, appID+".log")
	}

	// Skip 1 caller, since all log calls will be done from deckconverter/log
	logger, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}
	defer func() {
		// Don't check for errors since logger.Sync() can sometimes fail
		// even if the logs were properly displayed
		// See https://github.com/uber-go/zap/issues/328
		_ = logger.Sync()
	}()

	log.SetLogger(logger.Sugar())

	// TODO: Remove when upgrading Fyne
	// Temporary fix for OS X (see https://github.com/fyne-io/fyne/issues/824)
	// Manually specify the application theme
	if setTheme == "light" || setTheme == "dark" {
		err = os.Setenv("FYNE_THEME", setTheme)
		if err != nil {
			log.Errorf("Couldn't set the theme to \"%s\": %v", setTheme, err)
			os.Exit(1)
		}
	} else {
		log.Errorf("Invalid theme: %s", setTheme)
		os.Exit(1)
	}

	availablePlugins := dc.AvailablePlugins()

	application := app.NewWithID(appID)

	// TODO: Uncomment when upgrading Fyne
	// switch setTheme {
	// case "light":
	// 	app.Settings().SetTheme(theme.LightTheme())
	// case "dark":
	// 	app.Settings().SetTheme(theme.DarkTheme())
	// case "":
	// default:
	// 	log.Errorf("Invalid theme: %s", setTheme)
	// 	os.Exit(1)
	// }

	win := application.NewWindow(appName)
	win.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("Menu",
			fyne.NewMenuItem("About", func() {
				showAboutWindow(application)
			}),
		)), // a quit item will be appended to our first menu
	)
	win.SetMaster()

	uploaders := make([]string, 0, len(upload.TemplateUploaders))
	uploaders = append(uploaders, "No")
	for _, uploader := range upload.TemplateUploaders {
		uploaders = append(uploaders, (*uploader).UploaderName())
	}

	folderEntry := widget.NewEntry()
	templateLabel := widget.NewLabel("Create a template file:")
	uploaderSelect := widget.NewSelect(uploaders, nil)
	uploaderSelect.Selected = "No"

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

		tabItems = append(tabItems, widget.NewTabItem(plugin.PluginName(), pluginScreen(win, folderEntry, uploaderSelect, compactCheck, plugin)))
	}

	tabs := widget.NewTabContainer(tabItems...)
	tabs.SetTabLocation(widget.TabLocationLeading)

	generalOptions := widget.NewVBox(
		widget.NewHBox(
			widget.NewLabel("Output folder:"),
			folderEntry,
		),
		widget.NewHBox(
			templateLabel,
			uploaderSelect,
			compactCheck,
		),
	)

	win.SetContent(
		fyne.NewContainerWithLayout(
			layout.NewBorderLayout(generalOptions, nil, nil, nil),
			generalOptions,
			tabs,
		),
	)

	quitShortcut := desktop.CustomShortcut{KeyName: fyne.KeyQ, Modifier: desktop.ControlModifier}
	win.Canvas().AddShortcut(&quitShortcut, func(shortcut fyne.Shortcut) {
		application.Quit()
	})

	win.ShowAndRun()
}
