package calculator

import (
	"context"
	"time"
)

func periodicPrinterJob(ctx context.Context, state *State, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if ctx.Err() != nil {
				return
			}
			state.PrintTotals("periodic")
		}
	}
}
