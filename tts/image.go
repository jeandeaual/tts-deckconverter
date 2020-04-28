package tts

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"net/http"

	"github.com/disintegration/imaging"

	"github.com/jeandeaual/tts-deckconverter/log"
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

func downloadAndCreateThumbnail(url, filename string) (err error) {
	log.Infof("Querying %s", url)

	// Build the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		err = fmt.Errorf("couldn't create request for %s: %w", url, err)
		return
	}

	client := &http.Client{}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("couldn't query %s: %w", url, err)
		return
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	err = generateThumbnail(resp.Body, filename)

	return
}

func generateThumbnail(source io.Reader, filename string) error {
	// Open the source image
	cardThumb, err := imaging.Decode(source)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
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
		return fmt.Errorf("failed to save image: %w", err)
	}

	return nil
}
