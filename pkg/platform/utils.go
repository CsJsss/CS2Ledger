package platform

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"time"
)

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
func FetchAllPages(
	ctx context.Context,
	log *slog.Logger,
	name, direction string,
	pageSleep time.Duration,
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
