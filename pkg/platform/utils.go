package platform

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
)

var exteriorSuffixes = []string{
	"崭新出厂", "略有磨损", "久经沙场", "破损不堪", "战痕累累",
	"Factory New", "Minimal Wear", "Field-Tested", "Well-Worn", "Battle-Scarred",
}

// NormalizeItemName strips known CS2 exterior/wear suffixes from item names
// and returns the cleaned name along with any extracted exterior.
// e.g. "蝴蝶刀（★） | 澄澈之水 (久经沙场)" → "蝴蝶刀（★） | 澄澈之水", "久经沙场"
func NormalizeItemName(name string) (normalized string, exterior string) {
	normalized = strings.TrimSpace(name)
	for _, ext := range exteriorSuffixes {
		for _, pat := range []string{" (" + ext + ")", "(" + ext + ")", " " + ext, ext} {
			if strings.HasSuffix(normalized, pat) {
				normalized = strings.TrimSpace(normalized[:len(normalized)-len(pat)])
				exterior = ext
				return
			}
		}
	}
	return normalized, ""
}

// randomUA returns a randomized Chrome User-Agent string.
func RandomUA() string {
	firstNum := 55 + rand.Intn(8) // 55-62
	thirdNum := rand.Intn(3201)   // 0-3200
	fourthNum := rand.Intn(141)   // 0-140
	osTypes := []string{
		"(Windows NT 6.1; WOW64)",
		"(Windows NT 10.0; WOW64)",
		"(X11; Linux x86_64)",
		"(Macintosh; Intel Mac OS X 10_12_6)",
	}
	return fmt.Sprintf(
		"Mozilla/5.0 %s AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%d.0.%d.%d Safari/537.36",
		osTypes[rand.Intn(len(osTypes))], firstNum, thirdNum, fourthNum,
	)
}

// FetchAllPages calls fetchFn for each page until exhausted or ctx cancelled.
// When limit > 0, pagination stops early once limit records are collected.
func FetchAllPages(
	ctx context.Context,
	log *logfx.Logger,
	name, direction string,
	pageSleep time.Duration,
	limit int,
	fetchFn func(ctx context.Context, page int) (items []TradeRecord, hasMore bool, err error),
) ([]TradeRecord, error) {
	var all []TradeRecord
	page := 1
	for {
		if err := ctx.Err(); err != nil {
			return all, err
		}

		items, hasMore, err := fetchFn(ctx, page)
		if err != nil {
			return all, fmt.Errorf("%s %s history page %d: %w", name, direction, page, err)
		}
		all = append(all, items...)

		if limit > 0 && len(all) >= limit {
			return all[:limit], nil
		}
		if !hasMore {
			return all, nil
		}
		page++

		select {
		case <-time.After(pageSleep):
		case <-ctx.Done():
			return all, ctx.Err()
		}
	}
}
