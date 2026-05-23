package market

import (
	"sync"
	"time"

	"github.com/CsJsss/CS2Ledger/pkg/platform"
	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
)

type priceEntry struct {
	info   platform.PriceInfo
	expiry time.Time
}

// PriceCache is an in-memory cache for market prices with configurable TTL.
type PriceCache struct {
	mu   sync.RWMutex
	data map[string]priceEntry
	ttl  time.Duration
	log  *logfx.Logger
}

func NewPriceCache(ttl time.Duration, log *logfx.Logger) *PriceCache {
	return &PriceCache{
		data: make(map[string]priceEntry),
		ttl:  ttl,
		log:  log,
	}
}

func (pc *PriceCache) SetTTL(ttl time.Duration) {
	pc.ttl = ttl
}

// Get returns cached PriceInfo for names that are present and not expired.
// Returns the hits and the list of names that missed.
func (pc *PriceCache) Get(names []string) (hits []platform.PriceInfo, misses []string) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	now := time.Now()
	for _, name := range names {
		e, ok := pc.data[name]
		if ok && now.Before(e.expiry) {
			hits = append(hits, e.info)
		} else {
			misses = append(misses, name)
		}
	}
	return
}

// Set stores PriceInfo in the cache with the configured TTL.
func (pc *PriceCache) Set(infos []platform.PriceInfo) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	expiry := time.Now().Add(pc.ttl)
	for _, info := range infos {
		pc.data[info.MarketHashName] = priceEntry{info: info, expiry: expiry}
	}
}
