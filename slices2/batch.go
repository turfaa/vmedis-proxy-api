package slices2

// GenerateBatches generates batches of the given size.
func GenerateBatches[T any](slice []T, batchSize int) [][]T {
	if batchSize == 0 {
		return nil
	}

	batches := make([][]T, 0, (len(slice)+batchSize-1)/batchSize)
	for batchSize < len(slice) {
		slice, batches = slice[batchSize:], append(batches, slice[0:batchSize:batchSize])
	}

	return append(batches, slice)
}
