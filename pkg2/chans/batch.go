package chans

// GenerateBatches generates batches of the given size.
func GenerateBatches[T any](ch <-chan T, batchSize int) <-chan []T {
	batches := make(chan []T, 1)

	go func() {
		batch := make([]T, 0, batchSize)
		for t := range ch {
			batch = append(batch, t)
			if len(batch) == batchSize {
				batches <- batch
				batch = make([]T, 0, batchSize)
			}
		}

		if len(batch) > 0 {
			batches <- batch
		}

		close(batches)
	}()

	return batches
}
