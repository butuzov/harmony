package harmony

import "context"

// TeeWithContext will return two channels of generic type `T` used to fan-out
// data from the incoming channel. Channels needs to be read in order next
// iteration over incoming chanel happen.
func TeeWithContext[T any](ctx context.Context, incoming <-chan T) (<-chan T, <-chan T) {
	ch1, ch2 := make(chan T), make(chan T)

	go func() {
		defer close(ch1)
		defer close(ch2)

		for val := range OrContextDone(ctx, incoming) {
			ch1, ch2 := ch1, ch2
			for i := 0; i < 2; i++ {
				select {
				// this case statement can add issue with disproportional ch1, ch2 sends
				// e.g. first chan got val, second didn't got chance due context.Done.
				case <-ctx.Done():
					return
				case ch1 <- val:
					ch1 = nil
				case ch2 <- val:
					ch2 = nil
				}
			}
		}
	}()

	return ch1, ch2
}

// TeeWithContext will return two channels of generic type `T` used to fan-out
// data from the incoming channel. Channels needs to be read in order next
// iteration over incoming chanel happen.
func TeeWithDone[T any](done chan struct{}, incoming <-chan T) (<-chan T, <-chan T) {
	ch1, ch2 := make(chan T), make(chan T)

	go func() {
		defer close(ch1)
		defer close(ch2)

		for val := range OrDone(done, incoming) {
			ch1, ch2 := ch1, ch2
			for i := 0; i < 2; i++ {
				select {
				// This case statement can add issue with disproportional ch1, ch2 sends
				// e.g. first chan got val, second didn't got chance due closed `done`.
				case <-done:
					return
				case ch1 <- val:
					ch1 = nil
				case ch2 <- val:
					ch2 = nil
				}
			}
		}
	}()

	return ch1, ch2
}
