package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

// NotifyScanCancelled emits a NOTIFY on channel 'scan_cancelled' with the scan id as payload.
func (r *Repository) NotifyScanCancelled(ctx context.Context, id uint) error {
	if r == nil || r.listener == nil {
		return fmt.Errorf("notify: listener not initialized")
	}
	payload := strconv.FormatUint(uint64(id), 10)
	// Use pg_notify to safely deliver payload.
	_, err := r.listener.Exec(ctx, "SELECT pg_notify($1, $2)", "scan_cancelled", payload)
	return err
}

// SubscribeScanCancels returns a channel of cancelled scan IDs and a stop function.
// Caller should call the returned stop function to end the subscription.
func (r *Repository) SubscribeScanCancels(ctx context.Context) (<-chan uint, func(), error) {
	if r == nil || r.listener == nil {
		return nil, nil, fmt.Errorf("subscribe: listener not initialized")
	}

	// Create a derived context we can cancel to stop.
	ctx, cancel := context.WithCancel(ctx)
	out := make(chan uint)

	// Ensure we LISTEN on the channel.
	if _, err := r.listener.Exec(ctx, "LISTEN scan_cancelled"); err != nil {
		cancel()
		return nil, nil, fmt.Errorf("subscribe: listen: %w", err)
	}

	go func() {
		defer func() {
			// Unlisten and cleanup.
			r.listener.Exec(context.Background(), "UNLISTEN scan_cancelled")
			close(out)
		}()

		for {
			notif, err := r.listener.WaitForNotification(ctx)
			if err != nil {
				// If context cancelled, exit silently.
				if ctx.Err() != nil {
					return
				}
				// On transient errors, backoff and retry.
				select {
				case <-time.After(500 * time.Millisecond):
					continue
				case <-ctx.Done():
					return
				}
			}
			if notif == nil {
				continue
			}
			// Payload is the scan id.
			if id, err := strconv.ParseUint(notif.Payload, 10, 64); err == nil {
				select {
				case out <- uint(id):
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	stop := func() { cancel() }
	return out, stop, nil
}
