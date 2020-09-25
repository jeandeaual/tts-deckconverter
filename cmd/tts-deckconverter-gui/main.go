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
	defaultDeckType = "Generic"
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

func createURLTab(
	win fyne.Window,
	folderEntry *widget.Entry,
	uploaderSelect *widget.Select,
	compactCheck *widget.Check,
	backSelect *widget.Select,
	customBack *widget.Entry,
	optionWidgets map[string]interface{},
	plugin plugins.Plugin,
) *widget.TabItem {
	supportedUrls := widget.NewVBox(widget.NewLabel("Supported URLs:"))

	for _, urlHandler := range plugin.URLHandlers() {
		u, err := url.Parse(urlHandler.BasePath)
		if err != nil {
			log.Errorf("Invalid URL found for plugin %s: %v", plugin.PluginID, err)
			continue
		}
		supportedUrls.Append(widget.NewHyperlink(urlHandler.BasePath, u))
	}

	urlEntry := widget.NewEntry()

	return widget.NewTabItem("From URL", widget.NewVBox(
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
	))
}

func createTextTab(
	win fyne.Window,
	folderEntry *widget.Entry,
	uploaderSelect *widget.Select,
	compactCheck *widget.Check,
	backSelect *widget.Select,
	customBack *widget.Entry,
	optionWidgets map[string]interface{},
	plugin plugins.Plugin,
) *widget.TabItem {
	textInput := widget.NewMultiLineEntry()
	deckNameInput := widget.NewEntry()
	deckTypes := make([]string, 0, len(plugin.DeckTypeHandlers())+1)
	deckTypes = append(deckTypes, defaultDeckType)
	for deckType := range plugin.DeckTypeHandlers() {
		deckTypes = append(deckTypes, deckType)
	}
	deckTypeSelect := widget.NewSelect(deckTypes, nil)
	deckTypeSelect.OnChanged = func(selected string) {
		if selected == defaultDeckType {
			textInput.SetPlaceHolder(plugin.GenericFileHandler().Example)
			return
		}
		if deckTypeHandler, found := plugin.DeckTypeHandlers()[selected]; found {
			textInput.SetPlaceHolder(deckTypeHandler.Example)
		}
	}
	deckTypeSelect.SetSelected(defaultDeckType)

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
	)

	if len(deckTypes) > 1 {
		textInputButtons.Append(widget.NewLabel("Deck type:"))
		textInputButtons.Append(deckTypeSelect)
	}

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
				handler := plugin.GenericFileHandler().FileHandler
				if deckTypeSelect.Selected != defaultDeckType {
					if deckTypeHandler, found := plugin.DeckTypeHandlers()[deckTypeSelect.Selected]; found {
						handler = deckTypeHandler.FileHandler
					}
				}
				mode := plugin.PluginID()
				back := selectedBackURL(backSelect, customBack, plugin)
				if len(back) == 0 {
					showErrorf(win, "You need to set a card back")
					return
				}
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

	return widget.NewTabItem("From text", fyne.NewContainerWithLayout(
		layout.NewBorderLayout(
			nil,
			textInputButtons,
			nil,
			nil,
		),
		textInputScrollContainer,
		textInputButtons,
	))
}

func createFileTab(
	win fyne.Window,
	folderEntry *widget.Entry,
	uploaderSelect *widget.Select,
	compactCheck *widget.Check,
	backSelect *widget.Select,
	customBack *widget.Entry,
	optionWidgets map[string]interface{},
	plugin plugins.Plugin,
) *widget.TabItem {
	fileEntry := widget.NewEntry()
	fileEntry.Disable()

	return widget.NewTabItem("From file", widget.NewVBox(
		fileEntry,
		widget.NewHBox(
			widget.NewButtonWithIcon("File…", theme.DocumentSaveIcon(), func() {
				dialog.ShowFileOpen(
					func(file fyne.URIReadCloser, err error) {
						if err != nil {
							showErrorf(win, "Error when trying to select file: %v", err)
							return
						}
						if file == nil {
							// Cancelled
							return
						}
						defer func() {
							if cerr := file.Close(); cerr != nil {
								log.Errorf("Error when trying to close file %s: %v", file.URI().String(), cerr)
							}
						}()

						if file.URI().Scheme() != "file" {
							showErrorf(win, "Only local files are supported")
							return
						}

						filepath := strings.TrimPrefix(file.URI().String(), "file://")
						log.Infof("Selected %s", filepath)
						fileEntry.SetText(filepath)
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
				if len(back) == 0 {
					showErrorf(win, "You need to set a card back")
					return
				}
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
	))
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
			check.SetChecked(option.DefaultValue.(bool))
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
	backPreview := widget.NewHyperlink("Preview", nil)
	_ = backPreview.SetURLFromString(availableBacks[plugins.DefaultBackKey].URL)

	var lastSelected string
	if defaultBack, found := availableBacks[plugins.DefaultBackKey]; found {
		lastSelected = plugins.CapitalizeString(defaultBack.Description)
	} else {
		lastSelected = customBackLabel
	}

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

	if len(plugin.URLHandlers()) > 0 {
		tabItems = append(tabItems, createURLTab(
			win,
			folderEntry,
			uploaderSelect,
			compactCheck,
			backSelect,
			customBack,
			optionWidgets,
			plugin,
		))
	}

	tabItems = append(tabItems, createTextTab(
		win,
		folderEntry,
		uploaderSelect,
		compactCheck,
		backSelect,
		customBack,
		optionWidgets,
		plugin,
	))

	tabItems = append(tabItems, createFileTab(
		win,
		folderEntry,
		uploaderSelect,
		compactCheck,
		backSelect,
		customBack,
		optionWidgets,
		plugin,
	))

	tabContainer := widget.NewTabContainer(tabItems...)

	backSelect.SetSelected(lastSelected)

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
	flag.StringVar(&setTheme, "theme", "", "application theme (\"light\" or \"dark\")")

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

	availablePlugins := dc.AvailablePlugins()

	application := app.NewWithID(appID)

	switch setTheme {
	case "light":
		application.Settings().SetTheme(theme.LightTheme())
	case "dark":
		application.Settings().SetTheme(theme.DarkTheme())
	case "":
		// Do nothing
	default:
		log.Errorf("Invalid theme: %s", setTheme)
		os.Exit(1)
	}

	win := application.NewWindow(appName)
	win.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("Menu",
			fyne.NewMenuItem("About", func() {
				showAboutWindow(application)
			}),
		)), // a quit item will be appended to our first menu
	)
	win.SetMaster()
	win.CenterOnScreen()

	uploaders := make([]string, 0, len(upload.TemplateUploaders))
	uploaders = append(uploaders, "No")
	for _, uploader := range upload.TemplateUploaders {
		uploaders = append(uploaders, (*uploader).UploaderName())
	}

	folderLabel := widget.NewLabel("Output folder:")
	folderEntry := widget.NewEntry()

	var (
		chestPath     string
		currentFolder string
	)

	chestFolderButton := widget.NewButton("Chest folder", func() {
		folderEntry.SetText(chestPath)
	})
	currentFolderButton := widget.NewButton("Current folder", func() {
		folderEntry.SetText(currentFolder)
	})
	folderEntry.OnChanged = func(text string) {
		switch text {
		case "":
			chestFolderButton.Enable()
			currentFolderButton.Enable()
		case chestPath:
			chestFolderButton.Disable()
			currentFolderButton.Enable()
		case currentFolder:
			chestFolderButton.Enable()
			currentFolderButton.Disable()
		default:
			chestFolderButton.Enable()
			currentFolderButton.Enable()
		}
	}
	chestPath, err = tts.FindChestPath()
	if err != nil {
		log.Debugf("Couldn't find chest path: %v", err)
	}
	currentFolder, err = os.Getwd()
	if err != nil {
		log.Errorf("Couldn't get the working directory: %v", err)
	}
	if len(chestPath) > 0 {
		folderEntry.SetText(chestPath)
	} else {
		folderEntry.SetText(currentFolder)
	}
	folderButtons := widget.NewHBox(chestFolderButton, currentFolderButton)
	if len(chestPath) == 0 {
		chestFolderButton.Disable()
	}
	templateLabel := widget.NewLabel("Create a template file:")
	uploaderSelect := widget.NewSelect(uploaders, nil)
	uploaderSelect.Selected = "No"

	compactCheck := widget.NewCheck("Compact file", nil)

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
		fyne.NewContainerWithLayout(
			layout.NewBorderLayout(
				nil,
				nil,
				folderLabel,
				folderButtons,
			),
			folderLabel,
			folderEntry,
			folderButtons,
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

	if _, ok := application.Driver().(desktop.Driver); ok {
		// Desktop only
		quitShortcut := desktop.CustomShortcut{KeyName: fyne.KeyQ, Modifier: desktop.ControlModifier}
		win.Canvas().AddShortcut(&quitShortcut, func(shortcut fyne.Shortcut) {
			application.Quit()
		})
	}

	win.ShowAndRun()
}
