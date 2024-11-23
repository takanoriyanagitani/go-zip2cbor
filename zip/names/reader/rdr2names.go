package rdr2names

import (
	"bufio"
	"io"
	"iter"
)

func ReaderToNames(r io.Reader) iter.Seq[string] {
	return func(yield func(string) bool) {
		var s *bufio.Scanner = bufio.NewScanner(r)
		for s.Scan() {
			var line string = s.Text()
			if !yield(line) {
				return
			}
		}
	}
}
