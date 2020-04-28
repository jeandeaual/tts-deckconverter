package upload

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jeandeaual/tts-deckconverter/log"
)

func setupImgurTestServer(json string) (*http.Client, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-RateLimit-UserLimit", "10")
		w.Header().Set("X-RateLimit-UserRemaining", "2")
		w.Header().Set("X-RateLimit-UserReset", "3")
		w.Header().Set("X-RateLimit-ClientLimit", "40")
		w.Header().Set("X-RateLimit-ClientRemaining", "5")
		w.WriteHeader(200)

		fmt.Fprintln(w, json)
	}))

	u, err := url.Parse(server.URL)
	if err != nil {
		log.Fatal("failed to parse httptest.Server URL:", err)
	}

	http.DefaultClient.Transport = rewriteTransport{URL: u}
	return http.DefaultClient, server
}

func setupImgurErrorTestServer(code int, json string) (*http.Client, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)

		fmt.Fprintln(w, json)
	}))

	u, err := url.Parse(server.URL)
	if err != nil {
		log.Fatal("failed to parse httptest.Server URL:", err)
	}

	http.DefaultClient.Transport = rewriteTransport{URL: u}
	return http.DefaultClient, server
}

type rewriteTransport struct {
	Transport http.RoundTripper
	URL       *url.URL
}

func (t rewriteTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.URL.Scheme = t.URL.Scheme
	r.URL.Host = t.URL.Host
	r.URL.Path = path.Join(t.URL.Path, r.URL.Path)
	rt := t.Transport
	if rt == nil {
		rt = http.DefaultTransport
	}
	return rt.RoundTrip(r)
}

func TestImgurUpload(t *testing.T) {
	uploader := ImgurUploader{}

	assert.Equal(t, "Imgur", uploader.UploaderName())

	httpClient, ts := setupImgurTestServer(`{
  "data": {
    "id": "orunSTu",
    "title": null,
    "description": null,
    "datetime": 1495556889,
    "type": "image/png",
    "animated": false,
    "width": 1,
    "height": 1,
    "size": 42,
    "views": 0,
    "bandwidth": 0,
    "vote": null,
    "favorite": false,
    "nsfw": null,
    "section": null,
    "account_url": null,
    "account_id": 0,
    "is_ad": false,
    "in_most_viral": false,
    "tags": [],
    "ad_type": 0,
    "ad_url": "",
    "in_gallery": false,
    "deletehash": "x70po4w7BVvSUzZ",
    "name": "",
    "link": "http://i.imgur.com/orunSTu.png"
  },
  "success": true,
  "status": 200
}
`)
	defer ts.Close()

	tmpFile := createTempFile(t, 2_000_000)
	defer removeFile(tmpFile)

	// Successfully upload an image
	url, err := uploader.Upload(tmpFile, "Test", httpClient)
	assert.Nil(t, err)
	assert.Equal(t, "http://i.imgur.com/orunSTu.png", url)

	tmpFile = createTempFile(t, 0)
	removeFile(tmpFile)

	// Upload an non-existing file
	_, err = uploader.Upload(tmpFile, "Test", httpClient)
	assert.NotNil(t, err)
	log.Error(err)

	tmpFile = createTempFile(t, 25_000_000)
	defer removeFile(tmpFile)

	// Try to upload an image that is too large
	_, err = uploader.Upload(tmpFile, "Test", httpClient)
	assert.NotNil(t, err)
	log.Error(err)

	httpClient, ts = setupImgurErrorTestServer(500, `{
  "data": {},
  "success": false,
  "status": 500
}
`)
	defer ts.Close()

	tmpFile = createTempFile(t, 200)
	defer removeFile(tmpFile)

	// Image upload with client error
	_, err = uploader.Upload(tmpFile, "Test", httpClient)
	assert.NotNil(t, err)
	log.Error(err)
}
