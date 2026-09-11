package generator

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func workerJob(ctx context.Context, workerID int, client *http.Client, baseURL *url.URL, interval time.Duration, stats *Stats) {
	for ctx.Err() == nil {
		num := rand.IntN(201) - 100
		err := sendRequest(ctx, client, baseURL, num)
		stats.record(err)
		if err != nil {
			slog.Error("request failed", "worker", workerID, "error", err)
		}

		if interval > 0 {
			timer := time.NewTimer(interval)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
		}
	}
}

func sendRequest(ctx context.Context, client *http.Client, baseURL *url.URL, num int) error {
	requestURL := *baseURL
	query := requestURL.Query()
	query.Set("num", strconv.Itoa(num))
	requestURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL.String(), http.NoBody)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		return err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return errors.New(resp.Status)
	}
	return nil
}
