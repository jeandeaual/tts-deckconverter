package tts

import (
	"errors"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"

	"github.com/jeandeaual/tts-deckconverter/log"
	"github.com/jeandeaual/tts-deckconverter/plugins"
	"github.com/jeandeaual/tts-deckconverter/tts/upload"
)

const charsToRemove = `/<>:"\|?*`

var (
	startingID                = 100
	maxTemplateCols      uint = 10
	maxTemplateRows      uint = 7
	maxTemplateCount          = maxTemplateCols * maxTemplateRows
	errTooManyInTemplate      = errors.New("too many elements in template (should be less than " + fmt.Sprint(maxTemplateCount) + ")")
	errAlreadyExists          = errors.New("target file already exists")
)

func findTemplateSize(count uint) (uint, uint, error) {
	if count > maxTemplateCount {
		return 0, 0, errTooManyInTemplate
	}

	if count > maxTemplateCount-maxTemplateRows {
		return maxTemplateCols, maxTemplateRows, nil
	}

	sqrt := math.Sqrt(float64(count))
	integer, fraction := math.Modf(sqrt)
	if fraction == 0 {
		return uint(sqrt), uint(sqrt), nil
	}
	if fraction > 0.5 {
		return uint(math.Ceil(sqrt) + 1), uint(integer), nil
	}
	return uint(math.Ceil(sqrt)), uint(integer), nil
}

func getImageSize(filepath string) (width int, height int, err error) {
	file, err := os.Open(filepath)
	if err != nil {
		log.Errorf("Couldn't open file %s: %s", filepath, err)
		return
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	image, _, err := image.DecodeConfig(file)
	if err != nil {
		log.Errorf("Couldn't decode image %s: %s", filepath, err)
		return
	}

	return image.Width, image.Height, nil
}

func downloadFile(url string, filepath string) (err error) {
	if _, err = os.Stat(filepath); err == nil {
		err = errAlreadyExists
		return
	}
	output, err := os.Create(filepath)
	if err != nil {
		log.Errorf("Error while creating %s: %s", filepath, err)
		return
	}
	defer func() {
		if cerr := output.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	resp, err := http.Get(url)
	if err != nil {
		log.Errorf("Error while downloading %s: %s", url, err)
		return
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", resp.Status)
		log.Error(err)
		return
	}

	n, err := io.Copy(output, resp.Body)
	if err != nil {
		log.Errorf("Error while downloading %s: %s", url, err)
		return
	}

	log.Debugf("Downloaded file %s to %s (%d bytes)", url, filepath, n)
	return nil
}

func generateTemplate(cards []plugins.CardInfo, tmpDir, outputPath string, count int) (urlIDMap map[string]int, numCols, numRows uint, err error) {
	idFilePathMap := make(map[int]string)
	urlIDMap = make(map[string]int)

	id := startingID * count
	for _, card := range cards {
		filename := filepath.Join(tmpDir, strings.Trim(card.Name, charsToRemove))
		err = downloadFile(card.ImageURL, filename)
		if err != nil && err == errAlreadyExists {
			log.Debugf("File %s already exists, reusing it (card %s)", filename, card.Name)
			err = nil
		} else if err != nil {
			return
		}

		idFilePathMap[id] = filename
		urlIDMap[card.ImageURL] = id

		id++

		if card.AlternativeState != nil {
			filename := filepath.Join(tmpDir, strings.Trim(card.AlternativeState.Name, charsToRemove))
			err = downloadFile(card.AlternativeState.ImageURL, filename)
			if err != nil && err == errAlreadyExists {
				log.Debugf("File %s already exists, reusing it (card %s)", filename, card.AlternativeState.Name)
				err = nil
			} else if err != nil {
				return
			}

			idFilePathMap[id] = filename
			urlIDMap[card.AlternativeState.ImageURL] = id

			id++
		}
	}

	var (
		width     int
		height    int
		maxWidth  int
		maxHeight int
		ratio     float64
	)

	for _, filepath := range idFilePathMap {
		width, height, err = getImageSize(filepath)
		if err != nil {
			return
		}

		if maxWidth == 0 && maxHeight == 0 {
			maxWidth = width
			maxHeight = height
			ratio = float64(width) / float64(height)
		} else if width != maxWidth || height != maxHeight {
			if float64(width)/float64(height) != ratio {
				err = fmt.Errorf("the images don't all have the same ratio")
				return
			}
			if width > maxWidth || height > maxHeight {
				maxWidth = width
				maxHeight = height
			}
		}
	}

	imageCount := len(idFilePathMap)

	log.Debugw(
		"Image parsing done",
		"maxWidth", maxWidth,
		"maxHeight", maxHeight,
		"ratio", ratio,
		"count", imageCount,
	)

	numCols, numRows, err = findTemplateSize(uint(imageCount))
	if err != nil {
		log.Errorw(
			err.Error(),
			"imageCount", imageCount,
			"card length", len(cards),
			"idFilePathMap", idFilePathMap,
		)
		return
	}
	templateWidth := int(numCols) * maxWidth
	templateHeight := int(numRows) * maxHeight
	log.Infof(
		"We have %d items, so create a %d×%d template (%d×%d pixels)",
		imageCount,
		numCols,
		numRows,
		templateWidth,
		templateHeight,
	)

	template := imaging.New(templateWidth, templateHeight, white)
	var (
		curRow uint = 1
		curCol uint = 1
	)

	for i := 0; i < imageCount; i++ {
		filepath, found := idFilePathMap[startingID*count+i]
		if !found {
			err = fmt.Errorf("image for ID %d not found", startingID+i)
			return
		}
		var source *os.File
		source, err = os.Open(filepath)
		if err != nil {
			return
		}
		defer func() {
			if cerr := source.Close(); cerr != nil && err == nil {
				err = cerr
			}
		}()

		var cardImage image.Image
		cardImage, err = imaging.Decode(source)
		if err != nil {
			return
		}

		if cardImage.Bounds().Max.X != maxWidth || cardImage.Bounds().Max.Y != maxHeight {
			// Resize the image so it fits the template
			cardImage = imaging.Resize(cardImage, maxWidth, maxHeight, imaging.Lanczos)
		}

		template = imaging.Paste(
			template,
			cardImage,
			image.Pt((int(curCol)-1)*maxWidth, (int(curRow)-1)*maxHeight),
		)

		if curCol == numCols {
			curCol = 1
			curRow++
		} else {
			curCol++
		}
	}

	// Save the resulting image
	err = imaging.Save(template, outputPath, imaging.JPEGQuality(100))

	return
}

func generateTemplatesForRelatedDecks(decks []*plugins.Deck, tmpDir, outputFolder string, uploader upload.TemplateUploader) []error {
	var (
		urlIDMap   map[string]int
		outputPath string
		numCols    uint
		numRows    uint
		err        error
	)

	errs := []error{}
	uniqueCards := make(map[string]struct{})

	for _, deck := range decks {
		for _, card := range deck.Cards {
			uniqueCards[card.ImageURL] = struct{}{}
			if card.AlternativeState != nil {
				uniqueCards[card.AlternativeState.ImageURL] = struct{}{}
			}
		}
	}

	totalCount := len(uniqueCards)

	if totalCount > int(maxTemplateCount) {
		totalTemplateCount := 1
		for _, deck := range decks {
			log.Debugw(
				"Parsing cards to generate template(s)",
				"card count", len(deck.Cards),
				"cards", deck.Cards,
			)

			// Set used for counting unique cards
			uniqueCards = make(map[string]struct{})
			// Set used for counting alternative state cards
			alts := make(map[string]struct{})

			templateStarts := []int{0}
			templateEnds := []int{}

			for _, card := range deck.Cards {
				uniqueCards[card.ImageURL] = struct{}{}
				if len(uniqueCards)%int(maxTemplateCount) == 0 {
					if card.AlternativeState != nil {
						alts[card.AlternativeState.ImageURL] = struct{}{}
					}

					log.Debugf("Cut template number %d at %d", len(templateStarts), len(uniqueCards)-len(alts))
					log.Debugf("Found %d cards with an alternative state", len(alts))

					templateEnds = append(templateEnds, len(uniqueCards)-len(alts))
					templateStarts = append(templateStarts, len(uniqueCards)-len(alts))

					alts = make(map[string]struct{})
					continue
				}
				if card.AlternativeState != nil {
					alts[card.AlternativeState.ImageURL] = struct{}{}
					if len(uniqueCards)%int(maxTemplateCount) == 0 {
						log.Debugf("Cut template number %d at %d", len(templateStarts), len(uniqueCards)-len(alts))
						log.Debugf("Found %d cards with an alternative state", len(alts))

						templateEnds = append(templateEnds, len(uniqueCards)-len(alts))
						templateStarts = append(templateStarts, len(uniqueCards)-len(alts))

						alts = make(map[string]struct{})
					}
				}
			}

			log.Debugf("Cut template number %d at %d", len(templateStarts), len(uniqueCards)-len(alts))
			log.Debugf("Found %d cards with an alternative state", len(alts))
			templateEnds = append(templateEnds, len(uniqueCards)-len(alts))

			for templateCount := 0; templateCount < len(templateStarts); templateCount++ {
				var suffix string
				if templateCount > 0 {
					suffix = fmt.Sprintf(" %d", templateCount+1)
				}
				templateName := filepathReplacer.Replace(deck.Name) + " - Template" + suffix

				if uploader.UploaderID() != "manual" {
					outputPath = filepath.Join(os.TempDir(), templateName+".jpg")
				} else if outputFolder == "" {
					outputPath = templateName + ".jpg"
				} else {
					outputPath = filepath.Join(outputFolder, templateName+".jpg")
				}

				start := templateStarts[templateCount]
				end := templateEnds[templateCount]
				log.Debugw(
					"Generating new template",
					"start", start,
					"end", end,
					"card count", len(deck.Cards),
				)

				urlIDMap, numCols, numRows, err = generateTemplate(
					deck.Cards[start:end],
					tmpDir,
					outputPath,
					totalTemplateCount,
				)
				if err != nil {
					errs = append(errs, fmt.Errorf("couldn't save template to %s: %w", outputPath, err))
					totalTemplateCount++
					continue
				}

				url, err := uploader.Upload(outputPath, templateName, http.DefaultClient)
				if err != nil {
					err = fmt.Errorf(
						"couldn't upload %s: %v\n"+
							"Try to upload %s manually, and update the URL in the deck file(s) manually",
						outputPath,
						err,
						outputPath,
					)
					errs = append(errs, err)
				} else if uploader.UploaderID() != "manual" {
					log.Debugf("Deleting template file %s", outputPath)
					err = os.Remove(outputPath)
					if err != nil {
						errs = append(errs, fmt.Errorf("Couldn't remove %s: %v", outputPath, err))
					}
				}

				template := &plugins.Template{
					URL:     url,
					NumCols: int(numCols),
					NumRows: int(numRows),
				}
				if deck.TemplateInfo == nil {
					deck.TemplateInfo = &plugins.TemplateInfo{
						ImageURLCardIDMap: urlIDMap,
						Templates: map[int]*plugins.Template{
							totalTemplateCount: template,
						},
					}
				} else {
					for cardURL, cardID := range urlIDMap {
						deck.TemplateInfo.ImageURLCardIDMap[cardURL] = cardID
					}
					deck.TemplateInfo.Templates[totalTemplateCount] = template
				}
				totalTemplateCount++
			}
		}

		return errs
	}

	cards := []plugins.CardInfo{}
	templateName := ""

	for _, deck := range decks {
		if len(templateName) == 0 {
			templateName = filepathReplacer.Replace(deck.Name) + " - Template"
		}
		cards = append(cards, deck.Cards...)
	}

	if uploader.UploaderID() != "manual" {
		outputPath = filepath.Join(os.TempDir(), templateName+".jpg")
	} else if outputFolder == "" {
		outputPath = templateName + ".jpg"
	} else {
		outputPath = filepath.Join(outputFolder, templateName+".jpg")
	}

	log.Debug("Generating new template")

	urlIDMap, numCols, numRows, err = generateTemplate(cards, tmpDir, outputPath, 1)
	if err != nil {
		errs = append(errs, fmt.Errorf("couldn't save template to %s: %w", outputPath, err))
		return errs
	}

	url, err := uploader.Upload(outputPath, templateName, http.DefaultClient)
	if err != nil {
		err = fmt.Errorf(
			"couldn't upload %s: %v\n"+
				"Try to upload %s manually, and update the URL in the deck file(s) manually",
			outputPath,
			err,
			outputPath,
		)
		errs = append(errs, err)
	} else if uploader.UploaderID() != "manual" {
		log.Debugf("Deleting template file %s", outputPath)
		err = os.Remove(outputPath)
		if err != nil {
			errs = append(errs, fmt.Errorf("Couldn't remove %s: %v", outputPath, err))
		}
	}

	template := &plugins.Template{
		URL:     url,
		NumCols: int(numCols),
		NumRows: int(numRows),
	}

	for _, deck := range decks {
		deck.TemplateInfo = &plugins.TemplateInfo{
			ImageURLCardIDMap: urlIDMap,
			Templates: map[int]*plugins.Template{
				1: template,
			},
		}
	}

	return errs
}

// GenerateTemplates generates one or several template files, similar to the
// files generated by the TTS Deck Editor.
// All the images required to display a deck are ordered in several rows and
// columns, to be later displayed by TTS when loading the deck.
// See https://berserk-games.com/knowledgebase/custom-decks/.
func GenerateTemplates(decks [][]*plugins.Deck, outputFolder string, uploader upload.TemplateUploader) (errs []error) {
	tmpDir, err := ioutil.TempDir("", "template")
	if err != nil {
		errs = append(errs, err)
		return
	}
	log.Debugf("Created temporary directory %s", tmpDir)
	// Remove the download folder when done
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			errs = append(errs, fmt.Errorf("couldn't remove template download directory: %w", err))
		}
	}()

	for _, relatedDecks := range decks {
		generateErrs := generateTemplatesForRelatedDecks(relatedDecks, tmpDir, outputFolder, uploader)
		errs = append(errs, generateErrs...)
	}

	return
}
