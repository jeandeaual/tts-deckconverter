package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	dc "github.com/jeandeaual/tts-deckconverter"
	"github.com/jeandeaual/tts-deckconverter/log"
	"github.com/jeandeaual/tts-deckconverter/plugins"
	"github.com/jeandeaual/tts-deckconverter/tts"
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

			sb.WriteString("\n")
			sb.WriteString("\t")
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

			sb.WriteString("\n")
			sb.WriteString("\t")
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

func handleFolder(target, mode, outputFolder, backURL string, uploader *upload.TemplateUploader, indent bool, options options) []error {
	log.Infof("Processing directory %s", target)

	files := []string{}
	errs := []error{}

	err := filepath.Walk(target, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == target {
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
		targetErrs := handleTarget(file, mode, outputFolder, backURL, uploader, indent, options)
		errs = append(errs, targetErrs...)
	}

	return errs
}

func handleTarget(target, mode, outputFolder, backURL string, uploader *upload.TemplateUploader, indent bool, options options) []error {
	log.Infof("Processing %s", target)

	errs := []error{}

	decks, err := dc.Parse(target, mode, options)
	if err != nil {
		errs = append(errs, fmt.Errorf("couldn't parse target: %w", err))
		return errs
	}

	if uploader != nil {
		templateErrs := tts.GenerateTemplates([][]*plugins.Deck{decks}, outputFolder, *uploader)
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

	generateErrs := tts.Generate(decks, backURL, outputFolder, indent)
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

func main() {
	var (
		err          error
		backURL      string
		back         string
		debug        bool
		mode         string
		outputFolder string
		chest        string
		templateMode string
		compact      bool
		showVersion  bool
	)

	availableModes := dc.AvailablePlugins()
	availableOptions := getAvailableOptions(availableModes)
	availableBacks := getAvailableBacks(availableModes)
	availableUploaders := getAvailableUploaders()

	options := make(options)

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s TARGET\n\nFlags:\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.StringVar(&back, "back", "", "card back (cannot be used with \"-backURL\"). Choose from:"+availableBacks)
	flag.StringVar(&backURL, "backURL", "", "custom URL for the card backs (cannot be used with \"-back\")")
	flag.StringVar(&mode, "mode", "", "available modes: "+strings.Join(availableModes, ", "))
	flag.StringVar(&outputFolder, "output", "", "destination folder (defaults to the current folder) (cannot be used with \"-chest\")")
	flag.StringVar(&chest, "chest", "", "save to the Tabletop Simulator chest folder (use \"/\" for the root folder) (cannot be used with \"-output\")")
	flag.StringVar(&templateMode, "template", "", "download each images and create a deck template instead of referring to each image individually. Choose from the following uploaders:"+availableUploaders)
	flag.Var(&options, "option", "plugin specific option (can have multiple)"+availableOptions)
	flag.BoolVar(&compact, "compact", false, "don't indent the resulting JSON file")
	if len(version) > 0 {
		flag.BoolVar(&showVersion, "version", false, "display the version information")
	}
	flag.BoolVar(&debug, "debug", false, "enable debug logging")

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

	plugin, found := dc.Plugins[mode]
	if len(mode) > 0 && !found {
		fmt.Fprintf(os.Stderr, "Invalid mode: %s\n\n", mode)
		flag.Usage()
		os.Exit(1)
	}

	if len(outputFolder) > 0 && len(chest) > 0 {
		fmt.Fprint(os.Stderr, "\"-output\" and \"-chest\" cannot be used at the same time\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if len(back) > 0 && len(backURL) > 0 {
		fmt.Fprint(os.Stderr, "\"-back\" and \"-backURL\" cannot be used at the same time\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if len(back) > 0 && plugin == nil {
		fmt.Fprint(os.Stderr, "You need to choose a mode in order to use \"-back\"\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if len(back) > 0 {
		chosenBack, found := plugin.AvailableBacks()[back]
		if !found {
			fmt.Fprintf(os.Stderr, "Invalid back for %s: %s\n\n", mode, back)
			flag.Usage()
			os.Exit(1)
		}
		backURL = chosenBack.URL
	}

	var uploader *upload.TemplateUploader

	if len(templateMode) > 0 {
		var found bool
		uploader, found = upload.TemplateUploaders[templateMode]
		if !found {
			fmt.Fprintf(os.Stderr, "Invalid template uploader: %s\n\n", templateMode)
			flag.Usage()
			os.Exit(1)
		}
	}

	target := flag.Args()[0]

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

	if len(outputFolder) > 0 {
		err = checkCreateDir(outputFolder)
		if err != nil {
			log.Fatal(err)
		}
	} else if len(chest) > 0 {
		chestPath, err := tts.FindChestPath()
		if err != nil {
			log.Fatal(err)
		}
		outputFolder = filepath.Join(chestPath, chest)
		err = checkCreateDir(outputFolder)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// Set the output directory to the current working directory
		outputFolder, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Infof("Generated files will go in %s", outputFolder)

	if info, err := os.Stat(target); err == nil && info.IsDir() {
		errs := handleFolder(target, mode, outputFolder, backURL, uploader, !compact, options)
		checkErrs(errs)
		os.Exit(0)
	}

	errs := handleTarget(target, mode, outputFolder, backURL, uploader, !compact, options)
	checkErrs(errs)
}
