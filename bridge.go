package harmony

import (
	"context"
)

// BridgeWithContext will return chan of generic type `T` used a pipe for the
// values received from the sequence of channels. Close channel (received from
// `incoming`) in order to  switch for a new one. Goroutines exists on close of
// `incoming` or context canceled.
func BridgeWithContext[T any](ctx context.Context, incoming <-chan (<-chan T)) <-chan T {
	outgoing := make(chan T)

	go func() {
		defer close(outgoing)

		for {
			var stream <-chan T

			// Reading new stream on each iteration over the incoming, once incoming
			// drained, close goroutine.
			select {
			case <-ctx.Done():
			case tmp, ok := <-incoming:
				if !ok {
					return
				}
				stream = tmp
			}

			// now we ae trying to actually read from channel, and pass value next to
			// its receive channel, if fail to read back to new iteration of this
			// loop.
			for val := range OrContextDone(ctx, stream) {
				outgoing <- val
			}
		}
	}()

	return outgoing
}
