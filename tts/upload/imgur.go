package upload

import (
	"fmt"
	"net/http"
	"os"

	"github.com/koffeinsource/go-imgur"

	"github.com/jeandeaual/tts-deckconverter/log"
)

var imgurClientID string

// kloggerAdapter makes tts-deckconverter/log satisfy the KLogger interface of
// github.com/koffeinsource/go-klogger
type kloggerAdapter struct{}

func (l *kloggerAdapter) Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

func (l *kloggerAdapter) Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func (l *kloggerAdapter) Warningf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

func (l *kloggerAdapter) Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

func (l *kloggerAdapter) Criticalf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

func byteFormatDecimal(b int64) string {
	const unit = 1000

	if b < unit {
		return fmt.Sprintf("%d B", b)
	}

	div, exp := int64(unit), 0

	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.2f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}

// Imgur's max upload size is 20 MB.
// See https://help.imgur.com/hc/en-us/articles/115000083326-What-files-can-I-upload-What-is-the-size-limit-
const imgurMaxUploadSize = 20_000_000

// ImgurUploader allows the user to upload a template file anonymously to Imgur
type ImgurUploader struct{}

// Upload a file to Imgur
func (iu ImgurUploader) Upload(templatePath string, templateName string, httpClient *http.Client) (string, error) {
	fi, err := os.Stat(templatePath)
	if err != nil {
		return "", fmt.Errorf("couldn't find template file %s: %w", templatePath, err)
	}

	templateSize := fi.Size()
	if templateSize > imgurMaxUploadSize {
		return "", fmt.Errorf(
			"%w (%s > %s). "+
				"Try again with lower-quality images.",
			ErrUploadSize,
			byteFormatDecimal(templateSize),
			byteFormatDecimal(imgurMaxUploadSize),
		)
	}

	client := &imgur.Client{
		HTTPClient:    httpClient,
		Log:           &kloggerAdapter{},
		ImgurClientID: imgurClientID,
	}

	img, _, err := client.UploadImageFromFile(templatePath, "", templateName, "TTS template file for "+templateName)
	if err != nil {
		return "", fmt.Errorf("couldn't upload the file to Imgur: %w", err)
	}

	deleteURL := "https://api.imgur.com/3/image/" + img.Deletehash

	log.Infof("Successfully uploaded %s to Imgur as \"%s\"", templatePath, img.Link)
	log.Infof("The uploaded file can be deleted by sending an HTTP DELETE request to %s", deleteURL)

	log.Debugf("Uploaded image data: %+v", img)

	return img.Link, err
}

// UploaderID returns the ID of the uploading service
func (iu ImgurUploader) UploaderID() string {
	return "imgur"
}

// UploaderName returns the name of the uploading service
func (iu ImgurUploader) UploaderName() string {
	return "Imgur"
}

// UploaderDescription returns the description of the uploading service
func (iu ImgurUploader) UploaderDescription() string {
	return "Upload the template(s) anonymously to Imgur."
}
