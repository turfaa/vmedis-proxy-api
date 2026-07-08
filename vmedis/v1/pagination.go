package vmedisv1

import (
	"context"
	"fmt"
	"log"
	"sync"

	"golang.org/x/sync/errgroup"
)

// lastPageProbe is a page number high enough that vmedis responds with the
// last available page, whose pagination links reveal the total number of pages.
const lastPageProbe = 9999999

// pageFetcher fetches one page of items and returns the items together with
// the other page numbers found in the page's pagination links.
type pageFetcher[T any] func(ctx context.Context, page int) (items []T, otherPages []int, err error)

// getAllPages fetches every page concurrently and returns the combined items.
// It fails on the first page that cannot be fetched; item order across pages
// is not guaranteed. name is only used in log messages.
func getAllPages[T any](ctx context.Context, name string, concurrency int, fetchPage pageFetcher[T]) ([]T, error) {
	log.Printf("Getting number of pages of %s", name)

	_, otherPages, err := fetchPage(ctx, lastPageProbe)
	if err != nil {
		return nil, fmt.Errorf("get number of pages of %s: %w", name, err)
	}

	lastPage := 1
	for _, p := range otherPages {
		if p > lastPage {
			lastPage = p
		}
	}

	log.Printf("Number of %s pages: %d", name, lastPage)

	if concurrency < 1 {
		concurrency = 1
	}

	var (
		items []T
		lock  sync.Mutex
		pages = make(chan int)
	)

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		defer close(pages)

		for page := 1; page <= lastPage; page++ {
			select {
			case pages <- page:
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		return nil
	})

	for i := 0; i < concurrency; i++ {
		eg.Go(func() error {
			for page := range pages {
				log.Printf("Getting %s page %d", name, page)

				pageItems, _, err := fetchPage(ctx, page)
				if err != nil {
					return fmt.Errorf("get %s page %d: %w", name, page, err)
				}

				log.Printf("Got %d %s from page %d", len(pageItems), name, page)

				lock.Lock()
				items = append(items, pageItems...)
				lock.Unlock()
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return items, nil
}
