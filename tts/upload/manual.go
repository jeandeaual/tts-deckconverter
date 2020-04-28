package upload

import (
	"net/http"
)

// ManualUploader will not upload the template.
// The user will have to do it manually, then replace the image URL in the generated deck(s).
type ManualUploader struct{}

// Upload just returns the template path, without uploading anything
func (mu ManualUploader) Upload(templatePath string, _templateName string, _client *http.Client) (string, error) {
	// Don't upload anything, just put "{{ TEMPLATE_PATH }}" as the image URL
	return "{{ " + templatePath + " }}", nil
}

// UploaderID returns the ID of the uploading service
func (mu ManualUploader) UploaderID() string {
	return "manual"
}

// UploaderName returns the name of the uploading service
func (mu ManualUploader) UploaderName() string {
	return "Manual"
}

// UploaderDescription returns the description of the uploading service
func (mu ManualUploader) UploaderDescription() string {
	return "Let the user manually upload the template."
}
