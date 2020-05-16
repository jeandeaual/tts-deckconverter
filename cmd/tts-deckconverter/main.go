package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	dc "github.com/jeandeaual/tts-deckconverter"
	"github.com/jeandeaual/tts-deckconverter/log"
	"github.com/jeandeaual/tts-deckconverter/plugins"
	"github.com/jeandeaual/tts-deckconverter/tts"
	"github.com/jeandeaual/tts-deckconverter/tts/upload"
)

func handleFolder(config appConfig) []error {
	log.Infof("Processing directory %s", config.target)

	files := []string{}
	errs := []error{}

	err := filepath.Walk(config.target, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == config.target {
			// The WalkFun is first called with the folder itself as argument
			// Skip it
			return nil
		}

		if info.IsDir() {
			log.Infof("Ignoring directory %s", path)
			// Do not process the files in the subfolder
			return filepath.SkipDir
		}

		// Do not process the file inside the WalkFun, overwise if we
		// generate files inside the target directory, these generated
		// files will be picked up by filepath.Walk
		files = append(files, path)

		return nil
	})
	if err != nil {
		log.Error(err)
		errs = append(errs, err)
		return errs
	}

	for _, file := range files {
		fileConfig := config
		fileConfig.target = file
		targetErrs := handleTarget(fileConfig)
		errs = append(errs, targetErrs...)
	}

	return errs
}

func handleTarget(config appConfig) []error {
	errs := []error{}

	var (
		decks []*plugins.Deck
		err   error
	)

	if config.target != "-" {
		log.Infof("Processing %s", config.target)

		decks, err = dc.Parse(config.target, config.mode, config.options)
	} else {
		plugin, found := dc.Plugins[config.mode]
		if !found {
			log.Fatalf("Invalid mode: %s", config.mode)
		}

		handler := plugin.GenericFileHandler()
		if deckTypeHandler, found := plugin.DeckTypeHandlers()[config.deckFormat]; found {
			handler = deckTypeHandler
		} else {
			log.Fatalf("Invalid format: %s", config.deckFormat)
		}

		log.Info("Processing stdin")

		decks, err = handler(os.Stdin, config.deckName, config.options)
	}
	if err != nil {
		errs = append(errs, fmt.Errorf("couldn't parse target: %w", err))
		return errs
	}

	if config.uploader != nil {
		templateErrs := tts.GenerateTemplates([][]*plugins.Deck{decks}, config.outputFolder, *config.uploader)
		if len(templateErrs) > 0 {
			uploadSizeErrsOnly := true
			for _, err := range templateErrs {
				if !errors.Is(err, upload.ErrUploadSize) {
					uploadSizeErrsOnly = false
				}
			}

			// If the only error we got was that the template was too big to be uploaded, continue
			// The user will be able to upload the template manually later on
			if !uploadSizeErrsOnly {
				return templateErrs
			}

			errs = append(errs, templateErrs...)
		}
	}

	generateErrs := tts.Generate(decks, config.backURL, config.outputFolder, !config.compact)
	return append(errs, generateErrs...)
}

func checkCreateDir(path string) error {
	if stat, err := os.Stat(path); os.IsNotExist(err) {
		log.Infof("Output folder %s doesn't exist, creating it", path)
		if plugins.CheckInvalidFolderName(path) {
			log.Fatalf("Invalid folder name: %s", path)
		}
		err = os.MkdirAll(path, 0o755)
		if err != nil {
			return err
		}
	} else if err != nil {
		return fmt.Errorf("invalid path %s: %w", path, err)
	} else if !stat.IsDir() {
		return fmt.Errorf("output folder %s is not a directory", path)
	}

	return nil
}

func checkErrs(errs []error) {
	if len(errs) > 0 {
		for _, err := range errs {
			log.Error(plugins.CapitalizeString(err.Error()))
		}
		os.Exit(1)
	}
}

type appConfig struct {
	target       string
	backURL      string
	back         string
	debug        bool
	mode         string
	deckName     string
	deckFormat   string
	outputFolder string
	chest        string
	templateMode string
	uploader     *upload.TemplateUploader
	compact      bool
	options      options
}

func parseFlags() appConfig {
	var (
		config      appConfig
		showVersion bool
	)

	availableModes := dc.AvailablePlugins()
	availableOptions := getAvailableOptions(availableModes)
	availableDeckFormats := getAvailableDeckFormats(availableModes)
	availableBacks := getAvailableBacks(availableModes)
	availableUploaders := getAvailableUploaders()

	config.options = make(options)

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s TARGET\n\nFlags:\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.StringVar(&config.back, "back", "", "card back (cannot be used with \"-backURL\"). Choose from:"+availableBacks)
	flag.StringVar(&config.backURL, "backURL", "", "custom URL for the card backs (cannot be used with \"-back\")")
	flag.StringVar(&config.mode, "mode", "", "available modes: "+strings.Join(availableModes, ", "))
	flag.StringVar(&config.deckName, "name", "", "name of the deck (usually inferred from the input file name or URL, but required with stdin)")
	flag.StringVar(&config.deckFormat, "format", "", "format of the deck (usually inferred from the input file name or URL, but required with stdin)"+availableDeckFormats)
	flag.StringVar(&config.outputFolder, "output", "", "destination folder (defaults to the current folder) (cannot be used with \"-chest\")")
	flag.StringVar(&config.chest, "chest", "", "save to the Tabletop Simulator chest folder (use \"/\" for the root folder) (cannot be used with \"-output\")")
	flag.StringVar(&config.templateMode, "template", "", "download each images and create a deck template instead of referring to each image individually. Choose from the following uploaders:"+availableUploaders)
	flag.Var(&config.options, "option", "plugin specific option (can have multiple)"+availableOptions)
	flag.BoolVar(&config.compact, "compact", false, "don't indent the resulting JSON file")
	if len(version) > 0 {
		flag.BoolVar(&showVersion, "version", false, "display the version information")
	}
	flag.BoolVar(&config.debug, "debug", false, "enable debug logging")

	flag.Parse()

	if showVersion {
		displayBuildInformation()
		os.Exit(0)
	}

	if flag.NArg() == 0 || flag.NArg() > 1 {
		fmt.Fprint(os.Stderr, "A target is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	plugin, found := dc.Plugins[config.mode]
	if len(config.mode) > 0 && !found {
		fmt.Fprintf(os.Stderr, "Invalid mode: %s\n\n", config.mode)
		flag.Usage()
		os.Exit(1)
	}

	if len(config.outputFolder) > 0 && len(config.chest) > 0 {
		fmt.Fprint(os.Stderr, "\"-output\" and \"-chest\" cannot be used at the same time\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if len(config.back) > 0 && len(config.backURL) > 0 {
		fmt.Fprint(os.Stderr, "\"-back\" and \"-backURL\" cannot be used at the same time\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if len(config.back) > 0 && plugin == nil {
		fmt.Fprint(os.Stderr, "You need to choose a mode in order to use \"-back\"\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if len(config.back) > 0 {
		chosenBack, found := plugin.AvailableBacks()[config.back]
		if !found {
			fmt.Fprintf(os.Stderr, "Invalid back for %s: %s\n\n", config.mode, config.back)
			flag.Usage()
			os.Exit(1)
		}
		config.backURL = chosenBack.URL
	}

	if len(config.templateMode) > 0 {
		var found bool
		config.uploader, found = upload.TemplateUploaders[config.templateMode]
		if !found {
			fmt.Fprintf(os.Stderr, "Invalid template uploader: %s\n\n", config.templateMode)
			flag.Usage()
			os.Exit(1)
		}
	}

	config.target = flag.Args()[0]

	if config.target == "-" {
		if len(config.mode) == 0 {
			fmt.Fprintln(os.Stderr, "-mode is required when parsing stdin")
			flag.Usage()
			os.Exit(1)
		}

		if len(config.deckName) == 0 {
			fmt.Fprintln(os.Stderr, "-name is required when parsing stdin")
			flag.Usage()
			os.Exit(1)
		}
	} else if len(config.deckName) > 0 {
		fmt.Fprintln(os.Stderr, "You can only set the deck name when parsing stdin")
		flag.Usage()
		os.Exit(1)
	}

	return config
}

func main() {
	config := parseFlags()

	var zapConf zap.Config

	if config.debug {
		zapConf = zap.NewDevelopmentConfig()
		zapConf.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	} else {
		zapConf = zap.NewProductionConfig()
		zapConf.Encoding = "console"
		zapConf.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
		zapConf.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
		zapConf.EncoderConfig.EncodeCaller = nil
	}

	// Skip 1 caller, since all log calls will be done from deckconverter/log
	logger, err := zapConf.Build(zap.AddCallerSkip(1))
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

	if len(config.outputFolder) > 0 {
		err = checkCreateDir(config.outputFolder)
		if err != nil {
			log.Fatal(err)
		}
	} else if len(config.chest) > 0 {
		var chestPath string
		chestPath, err = tts.FindChestPath()
		if err != nil {
			log.Fatal(err)
		}
		config.outputFolder = filepath.Join(chestPath, config.chest)
		err = checkCreateDir(config.outputFolder)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// Set the output directory to the current working directory
		config.outputFolder, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Infof("Generated files will go in %s", config.outputFolder)

	if info, err := os.Stat(config.target); err == nil && info.IsDir() {
		errs := handleFolder(config)
		checkErrs(errs)
		os.Exit(0)
	}

	errs := handleTarget(config)
	checkErrs(errs)
}
