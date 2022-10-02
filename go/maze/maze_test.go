package maze

import "testing"

func BenchmarkGenMaze(b *testing.B) {
	for n := 0; n < b.N; n++ {
		GenMaze(60, 60)
	}
}
