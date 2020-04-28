package upload

import (
	"errors"
	"net/http"
)

// ErrUploadSize is the error returned when the template is too large to be uploaded to the image uploading service
var ErrUploadSize = errors.New("the template is larger than the maximum upload size supported by the service")

// TemplateUploaders is the list of registered template uploader
// implementations.
var TemplateUploaders map[string]*TemplateUploader

// TemplateUploader is a generic interface implemented by template image
// uploaders.
type TemplateUploader interface {
	// Upload a file
	Upload(templatePath string, templateName string, httpClient *http.Client) (string, error)
	// UploaderID returns the ID of the uploading service
	UploaderID() string
	// UploaderName returns the name of the uploading service
	UploaderName() string
	// UploaderDescription returns the description of the uploading service
	UploaderDescription() string
}

func init() {
	TemplateUploaders = make(map[string]*TemplateUploader)

	registerTemplateUploaders(ManualUploader{})

	if len(imgurClientID) > 0 {
		registerTemplateUploaders(ImgurUploader{})
	}
}

func registerTemplateUploaders(uploaders ...TemplateUploader) {
	for _, uploader := range uploaders {
		TemplateUploaders[uploader.UploaderID()] = &uploader
	}
}
