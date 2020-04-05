package pkm

import (
	"strings"
	"sync"

	pokemontcgsdk "github.com/PokemonTCG/pokemon-tcg-sdk-go/src"
	"go.uber.org/zap"
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

func setUp(log *zap.SugaredLogger) bool {
	sets, err := pokemontcgsdk.GetSets(nil)
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

func getSetCode(ptcgoSetCode string, log *zap.SugaredLogger) (string, bool) {
	if strings.HasSuffix(ptcgoSetCode, "Energy") {
		ptcgoSetCode = strings.TrimSuffix(ptcgoSetCode, "Energy")
	}

	if ptcgoSetToStandardSetMap == nil {
		setUp(log)
	}

	return ptcgoSetToStandardSetMap.Load(ptcgoSetCode)
}

func getPTCGOSetCode(setCode string, log *zap.SugaredLogger) (string, bool) {
	if standardSetToPTCGOSetMap == nil {
		setUp(log)
	}

	return standardSetToPTCGOSetMap.Load(strings.ToLower(setCode))
}
