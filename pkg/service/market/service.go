package market

import (
	"context"
	"sync"
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
	startMu     sync.Mutex
	refreshMu   sync.Mutex
	refreshWg   sync.WaitGroup
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

// EnsureProvider lazily creates the price provider from the first csqaq account in DB.
func (s *MarketService) EnsureProvider() {
	s.ensureProvider()
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
	s.provider = csqaq.New(acc.Cookie, s.log) // cookie field stores the API token
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
		Select("market_hash_name", "buff_price", "buff_volume", "youpin_price", "youpin_volume", "steam_price", "steam_volume", "updated_at").
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
		var updatedAt int64
		if err := rows.Scan(&mhn, &buffPrice, &buffVol, &youpinPrice, &youpinVol, &steamPrice, &steamVol, &updatedAt); err != nil {
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
			UpdatedAt:      updatedAt,
		})
	}
	s.log.Debug("market: GetAllPrices returned", "count", len(all))
	return all, nil
}

// StartAutoRefresh begins periodic full refresh of all market prices.
// Safe to call multiple times — waits for previous loop to finish before starting a new one.
func (s *MarketService) StartAutoRefresh(_ context.Context) {
	s.startMu.Lock()
	defer s.startMu.Unlock()
	if s.stopRefresh != nil {
		close(s.stopRefresh)
		s.refreshWg.Wait()
	}
	s.refreshWg.Add(1)
	s.stopRefresh = make(chan struct{})
	go s.refreshLoop(s.stopRefresh)
}

func (s *MarketService) refreshLoop(stop chan struct{}) {
	defer s.refreshWg.Done()
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
	if !s.refreshMu.TryLock() {
		s.log.Debug("market: refresh skipped, already in progress")
		return
	}
	defer s.refreshMu.Unlock()

	s.ensureProvider()
	if s.provider == nil {
		s.log.Debug("market: refresh skipped, no csqaq account")
		return
	}

	s.fetchPrices(ctx)         // fast: price already-known items
	s.resolveMissingGoods(ctx) // slow: resolve missing market_hash_names
	s.fetchPrices(ctx)         // price newly-resolved items
}

func (s *MarketService) fetchPrices(ctx context.Context) {
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

func (s *MarketService) resolveMissingGoods(ctx context.Context) {
	type unresolvedKey struct {
		ItemName string
		Exterior string
	}

	seen := make(map[unresolvedKey]bool)
	var unresolved []unresolvedKey

	for _, table := range []string{"inventory", "trade_records"} {
		rows, err := s.db.Table(table).
			Select("DISTINCT item_name, exterior").
			Where("csqaq_goods_id = 0 AND item_name != ''").
			Rows()
		if err != nil {
			continue
		}
		for rows.Next() {
			var k unresolvedKey
			if err := rows.Scan(&k.ItemName, &k.Exterior); err != nil {
				continue
			}
			if !seen[k] {
				seen[k] = true
				unresolved = append(unresolved, k)
			}
		}
		_ = rows.Close()
	}

	if len(unresolved) == 0 {
		return
	}

	s.log.Info("market: resolving missing csqaq goods", "count", len(unresolved))

	for i, k := range unresolved {
		// Rate-limit: wait before every request (including first) to avoid 429
		if i > 0 {
			select {
			case <-time.After(3000 * time.Millisecond):
			case <-ctx.Done():
				return
			}
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		// If either table already has a good_id for this item+exterior (e.g. resolved
		// in a previous run but new rows appeared in the other table), reuse it.
		if goodID, mhn := s.findExistingGoods(k); goodID != 0 {
			s.log.Debug("market: reused existing goods", "itemName", k.ItemName, "goodID", goodID)
			s.updateGoods(k, goodID, mhn)
			continue
		}

		goodID, mhn, err := s.provider.ResolveGoodsInfo(ctx, k.ItemName, k.Exterior)
		if err != nil {
			s.log.Warn("market: resolve goods failed", "itemName", k.ItemName, "exterior", k.Exterior, "err", err)
			continue
		}
		if goodID == 0 {
			s.log.Debug("market: resolve goods no match", "itemName", k.ItemName, "exterior", k.Exterior)
			continue
		}

		s.log.Debug("market: resolved goods", "itemName", k.ItemName, "goodID", goodID, "mhn", mhn)
		s.updateGoods(k, goodID, mhn)
	}
}

func (s *MarketService) findExistingGoods(k struct {
	ItemName string
	Exterior string
}) (int, string) {
	var goodID int
	var mhn string
	for _, table := range []string{"inventory", "trade_records"} {
		row := s.db.Table(table).
			Select("csqaq_goods_id, market_hash_name").
			Where("item_name = ? AND exterior = ? AND csqaq_goods_id != 0", k.ItemName, k.Exterior).
			Row()
		if row == nil {
			continue
		}
		if err := row.Scan(&goodID, &mhn); err != nil {
			continue
		}
		if goodID != 0 {
			return goodID, mhn
		}
	}
	return 0, ""
}

func (s *MarketService) updateGoods(k struct {
	ItemName string
	Exterior string
}, goodID int, mhn string) {
	for _, table := range []string{"inventory", "trade_records"} {
		if err := s.db.Table(table).
			Where("item_name = ? AND exterior = ? AND csqaq_goods_id = 0", k.ItemName, k.Exterior).
			Update("csqaq_goods_id", goodID).Error; err != nil {
			s.log.Warn("market: update csqaq_goods_id failed", "table", table, "err", err)
		}
		if mhn != "" {
			if err := s.db.Table(table).
				Where("item_name = ? AND exterior = ? AND market_hash_name = ''", k.ItemName, k.Exterior).
				Update("market_hash_name", mhn).Error; err != nil {
				s.log.Warn("market: update market_hash_name failed", "table", table, "err", err)
			}
		}
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
		var updatedAt int64
		if err := rows.Scan(&mhn, &buffPrice, &buffVol, &youpinPrice, &youpinVol, &steamPrice, &steamVol, &updatedAt); err != nil {
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
			UpdatedAt:      updatedAt,
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
