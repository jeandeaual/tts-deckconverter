package pkm

import (
	"strings"
	"sync"

	"github.com/jeandeaual/tts-deckconverter/log"
)

type setMap struct {
	sync.RWMutex
	internal map[string]string
}

func newSetMap() *setMap {
	return &setMap{
		internal: make(map[string]string),
	}
}

func (sm *setMap) Load(key string) (value string, ok bool) {
	sm.RLock()
	result, ok := sm.internal[key]
	sm.RUnlock()
	return result, ok
}

func (sm *setMap) Delete(key string) {
	sm.Lock()
	delete(sm.internal, key)
	sm.Unlock()
}

func (sm *setMap) Store(key, value string) {
	sm.Lock()
	sm.internal[key] = value
	sm.Unlock()
}

var (
	ptcgoSetToStandardSetMap *setMap
	standardSetToPTCGOSetMap *setMap
)

func setUp() bool {
	sets, err := getSets()
	if err != nil {
		log.Errorf("Couldn't retrieve sets: %s", err)
		return false
	}

	ptcgoSetToStandardSetMap = newSetMap()
	standardSetToPTCGOSetMap = newSetMap()

	for _, set := range sets {
		ptcgoSetToStandardSetMap.Store(set.PtcgoCode, set.Code)
		standardSetToPTCGOSetMap.Store(set.Code, set.PtcgoCode)
	}

	return true
}

func getSetCode(ptcgoSetCode string) (string, bool) {
	if strings.HasSuffix(ptcgoSetCode, "Energy") {
		ptcgoSetCode = strings.TrimSuffix(ptcgoSetCode, "Energy")
	}

	if ptcgoSetToStandardSetMap == nil {
		setUp()
	}

	return ptcgoSetToStandardSetMap.Load(ptcgoSetCode)
}

func getPTCGOSetCode(setCode string) (string, bool) {
	if standardSetToPTCGOSetMap == nil {
		setUp()
	}

	return standardSetToPTCGOSetMap.Load(strings.ToLower(setCode))
}
