package tts

import (
	"image"
	"image/color"
	"io"
	"net/http"

	"github.com/disintegration/imaging"

	"deckconverter/log"
)

const (
	thumbnailSize = 256
	// We want 11 pixel margins on the top and bottom
	topBottomMargin  = 11
	innerImageHeight = 256 - topBottomMargin*2
)

var (
	transparent color.Color = color.NRGBA{0, 0, 0, 0}
	white       color.Color = color.NRGBA{0xff, 0xff, 0xff, 0xff}
)

func downloadAndCreateThumbnail(url, filename string) {
	log.Infof("Querying %s", url)

	// Build the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("Couldn't create request for %s: %s", url, err)
		return
	}

	client := &http.Client{}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("Couldn't query %s: %s", url, err)
		return
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Error(err)
		}
	}()

	generateThumbnail(resp.Body, filename)
}

func generateThumbnail(source io.Reader, filename string) {
	// Open the source image
	cardThumb, err := imaging.Decode(source)
	if err != nil {
		log.Errorf("Failed to open image: %v", err)
	}

	cardThumb = imaging.Resize(cardThumb, 0, innerImageHeight, imaging.Lanczos)

	background := imaging.New(thumbnailSize, thumbnailSize, transparent)
	cardThumbSize := cardThumb.Bounds().Size()
	background = imaging.Paste(
		background,
		cardThumb,
		image.Pt(
			thumbnailSize/2-cardThumbSize.X/2,
			(thumbnailSize-cardThumbSize.Y)/2,
		),
	)

	// Save the resulting image as PNG
	err = imaging.Save(background, filename)
	if err != nil {
		log.Errorf("Failed to save image: %v", err)
	}
}
