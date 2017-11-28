package controller

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type Starter interface {
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

func Sync(ctx context.Context, starters ...Starter) error {
	eg, ctx := errgroup.WithContext(ctx)
	for _, starter := range starters {
		func(starter Starter) {
			eg.Go(func() error {
				return starter.Sync(ctx)
			})
		}(starter)
	}
	return eg.Wait()
}

func Start(ctx context.Context, threadiness int, starters ...Starter) error {
	for _, starter := range starters {
		if err := starter.Start(ctx, threadiness); err != nil {
			return err
		}
	}
	return nil
}
