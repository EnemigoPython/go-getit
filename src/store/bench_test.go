package store

import "testing"

func BenchmarkHashFunction(b *testing.B) {
	key := "test_key_for_hashing"
	limit := int64(1000)

	for b.Loop() {
		hashKey(key, limit)
	}
}
