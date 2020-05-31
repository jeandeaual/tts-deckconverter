package ygo

import (
	"time"

	"github.com/jeandeaual/tts-deckconverter/plugins/ygo/api"
)

// See https://db.ygoprodeck.com/api-guide/
var rateLimiter = time.NewTicker(50 * time.Millisecond)

func queryID(id int64, format api.Format) (api.Data, error) {
	<-rateLimiter.C
	return api.QueryID(id, format)
}

func queryName(name string, format api.Format) (api.Data, error) {
	<-rateLimiter.C
	return api.QueryName(name, format)
}
