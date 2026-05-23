package market

import (
	"context"
	"time"

	"github.com/CsJsss/CS2Ledger/pkg/model"
	"github.com/CsJsss/CS2Ledger/pkg/orm"
	"github.com/CsJsss/CS2Ledger/pkg/platform"
	"github.com/CsJsss/CS2Ledger/pkg/platform/csqaq"
	"github.com/CsJsss/CS2Ledger/pkg/utils/configfx"
	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
)

// MarketService fetches market prices with two-tier caching:
// L1: in-memory cache (configurable TTL), L2: SQLite market_prices table.
type MarketService struct {
	log         *logfx.Logger
	db          *orm.GormDB
	cache       *PriceCache
	ttlMin      int
	provider    platform.PriceProvider
	stopRefresh chan struct{}
}

func NewMarketService(
	log *logfx.Logger,
	db *orm.GormDB,
	cfg configfx.Config,
) *MarketService {
	ttlMin := cfg.PriceCacheTTL
	if ttlMin < 5 {
		ttlMin = 5
	}
	ttl := time.Duration(ttlMin) * time.Minute
	return &MarketService{
		log:    log,
		db:     db,
		cache:  NewPriceCache(ttl, log),
		ttlMin: ttlMin,
	}
}

// SetProvider sets the price provider lazily (csqaq account may not exist at startup).
func (s *MarketService) SetProvider(p platform.PriceProvider) {
	s.provider = p
}

// SetConfig updates runtime settings (price source + cache TTL).
func (s *MarketService) SetConfig(cfg PriceConfig) {
	s.log.Info("market: config updated", "priceSource", cfg.PriceSource, "cacheTTLMin", cfg.CacheTTLMin)
	s.ttlMin = cfg.CacheTTLMin
	s.cache.SetTTL(time.Duration(cfg.CacheTTLMin) * time.Minute)
	s.StartAutoRefresh(context.Background())
}

// ensureProvider lazily creates the price provider from the first csqaq account in DB.
func (s *MarketService) ensureProvider() {
	if s.provider != nil {
		return
	}
	var acc model.Account
	if err := s.db.Where("platform = ? AND deleted_at IS NULL", platform.PlatformCsqaq).First(&acc).Error; err != nil {
		s.log.Debug("market: no csqaq account found, prices unavailable")
		return
	}
	s.provider = csqaq.New(acc.Cookie) // cookie field stores the API token
	s.log.Info("market: csqaq provider ready", "account", acc.Name)
}

func (s *MarketService) GetPrices(names []string) ([]platform.PriceInfo, error) {
	if len(names) == 0 {
		return nil, nil
	}
	seen := make(map[string]bool)
	unique := make([]string, 0, len(names))
	for _, n := range names {
		if !seen[n] {
			seen[n] = true
			unique = append(unique, n)
		}
	}
	hits, _ := s.cache.Get(unique)
	dbHits, _ := s.queryDB(unique)
	return append(hits, dbHits...), nil
}

// GetAllPrices returns all cached market prices from SQLite. Does NOT call external API.
func (s *MarketService) GetAllPrices() ([]platform.PriceInfo, error) {
	rows, err := s.db.Table("market_prices").
		Select("market_hash_name", "buff_price", "buff_volume", "youpin_price", "youpin_volume", "steam_price", "steam_volume").
		Rows()
	if err != nil {
		s.log.Warn("market: GetAllPrices query failed", "err", err)
		return nil, nil
	}
	defer func() { _ = rows.Close() }()

	var all []platform.PriceInfo
	for rows.Next() {
		var mhn string
		var buffPrice, youpinPrice, steamPrice float64
		var buffVol, youpinVol, steamVol int
		if err := rows.Scan(&mhn, &buffPrice, &buffVol, &youpinPrice, &youpinVol, &steamPrice, &steamVol); err != nil {
			continue
		}
		all = append(all, platform.PriceInfo{
			MarketHashName: mhn,
			BuffPrice:      buffPrice,
			BuffVolume:     buffVol,
			YoupinPrice:    youpinPrice,
			YoupinVolume:   youpinVol,
			SteamPrice:     steamPrice,
			SteamVolume:    steamVol,
		})
	}
	s.log.Debug("market: GetAllPrices returned", "count", len(all))
	return all, nil
}

// StartAutoRefresh begins periodic full refresh of all market prices.
// Safe to call multiple times — stops previous loop before starting a new one.
func (s *MarketService) StartAutoRefresh(_ context.Context) {
	if s.stopRefresh != nil {
		close(s.stopRefresh)
	}
	s.stopRefresh = make(chan struct{})
	go s.refreshLoop(s.stopRefresh)
}

func (s *MarketService) refreshLoop(stop chan struct{}) {
	s.log.Info("market: auto-refresh started", "intervalMin", s.ttlMin)
	s.refreshAll(context.Background())

	ticker := time.NewTicker(time.Duration(s.ttlMin) * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.refreshAll(context.Background())
		case <-stop:
			s.log.Info("market: auto-refresh restarted")
			return
		}
	}
}

func (s *MarketService) refreshAll(ctx context.Context) {
	s.ensureProvider()
	if s.provider == nil {
		s.log.Debug("market: refresh skipped, no csqaq account")
		return
	}

	rows, err := s.db.Table("inventory").
		Select("DISTINCT market_hash_name").
		Where("market_hash_name != ''").
		Where("account_id IN (SELECT id FROM accounts WHERE deleted_at IS NULL)").
		Rows()
	if err != nil {
		s.log.Warn("market: refresh query failed", "err", err)
		return
	}
	defer func() { _ = rows.Close() }()

	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			continue
		}
		if n != "" {
			names = append(names, n)
		}
	}
	if len(names) == 0 {
		s.log.Debug("market: refresh found no items in inventory")
		return
	}

	s.log.Info("market: refreshing all prices", "items", len(names))
	infos, err := s.provider.GetPrices(ctx, names)
	if err != nil {
		s.log.Warn("market: refresh API failure", "err", err)
	}
	if len(infos) > 0 {
		s.cache.Set(infos)
		s.upsertDB(infos)
		s.log.Info("market: refresh complete", "fetched", len(infos))
	} else {
		s.log.Warn("market: refresh got no prices from API")
	}
}

func (s *MarketService) queryDB(names []string) (hits []platform.PriceInfo, misses []string) {
	if len(names) == 0 {
		return nil, nil
	}
	cutoff := time.Now().Add(-time.Duration(s.ttlMin) * time.Minute).Unix()

	rows, err := s.db.Table("market_prices").
		Select("market_hash_name", "buff_price", "buff_volume", "youpin_price", "youpin_volume", "steam_price", "steam_volume", "updated_at").
		Where("market_hash_name IN ? AND updated_at >= ?", names, cutoff).
		Rows()
	if err != nil {
		return nil, names
	}
	defer func() { _ = rows.Close() }()

	found := make(map[string]bool)
	for rows.Next() {
		var mhn string
		var buffPrice, youpinPrice, steamPrice float64
		var buffVol, youpinVol, steamVol int
		var _updatedAt int64
		if err := rows.Scan(&mhn, &buffPrice, &buffVol, &youpinPrice, &youpinVol, &steamPrice, &steamVol, &_updatedAt); err != nil {
			continue
		}
		hits = append(hits, platform.PriceInfo{
			MarketHashName: mhn,
			BuffPrice:      buffPrice,
			BuffVolume:     buffVol,
			YoupinPrice:    youpinPrice,
			YoupinVolume:   youpinVol,
			SteamPrice:     steamPrice,
			SteamVolume:    steamVol,
		})
		found[mhn] = true
	}
	for _, n := range names {
		if !found[n] {
			misses = append(misses, n)
		}
	}
	if len(hits) > 0 {
		s.cache.Set(hits)
	}
	return
}

func (s *MarketService) upsertDB(infos []platform.PriceInfo) {
	now := time.Now().Unix()
	for _, info := range infos {
		err := s.db.Exec(
			`INSERT INTO market_prices (market_hash_name, buff_price, buff_volume, youpin_price, youpin_volume, steam_price, steam_volume, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			 ON CONFLICT(market_hash_name) DO UPDATE SET
			 buff_price=excluded.buff_price, buff_volume=excluded.buff_volume,
			 youpin_price=excluded.youpin_price, youpin_volume=excluded.youpin_volume,
			 steam_price=excluded.steam_price, steam_volume=excluded.steam_volume,
			 updated_at=excluded.updated_at`,
			info.MarketHashName, info.BuffPrice, info.BuffVolume,
			info.YoupinPrice, info.YoupinVolume,
			info.SteamPrice, info.SteamVolume, now,
		).Error
		if err != nil {
			s.log.Warn("market: upsert price failed", "name", info.MarketHashName, "err", err)
		}
	}
}
