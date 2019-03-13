package tts

import (
	"image"
	"image/color"
	"io"
	"net/http"

	"github.com/disintegration/imaging"
	"go.uber.org/zap"
)

const thumbnailSize = 256

// We want 11 pixel margins on the top and bottom
const topBottomMargin = 11
const innerImageHeight = 256 - topBottomMargin*2

var white color.Color = color.NRGBA{0xff, 0xff, 0xff, 0xff}

func downloadAndCreateThumbnail(url, filename string, log *zap.SugaredLogger) {
	log.Infof("Querying %s", url)
	// Build the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalw("NewRequest", "error", err)
		return
	}

	// For control over HTTP client headers,
	// redirect policy, and other settings,
	// create a Client
	// A Client is an HTTP client
	client := &http.Client{}

	// Send the request via a client
	// Do sends an HTTP request and
	// returns an HTTP response
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalw("Do", "error", err)
		return
	}

	// Callers should close resp.Body
	// when done reading from it
	// Defer the closing of the body
	defer resp.Body.Close()

	generateThumbnail(resp.Body, filename, log)
}
func generateThumbnail(source io.Reader, filename string, log *zap.SugaredLogger) {
	// Open a test image.
	// TODO: Use decode to open the image using an io.Reader instead
	cardThumb, err := imaging.Decode(source)
	if err != nil {
		log.Fatalf("Failed to open image: %v", err)
	}

	cardThumb = imaging.Resize(cardThumb, 0, innerImageHeight, imaging.Lanczos)

	background := imaging.New(thumbnailSize, thumbnailSize, white)
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
		log.Fatalf("Failed to save image: %v", err)
	}
}
