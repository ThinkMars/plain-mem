package internal

import (
	"iter"
	"strings"
)

func SplitSeq(s, sep string) iter.Seq[string] {
	return func(yield func(string) bool) {
		n := len(sep)
		for {
			i := strings.Index(s, sep)
			if i < 0 {
				yield(s)
				return
			}
			if !yield(s[:i]) {
				return
			}
			s = s[i+n:]
		}
	}
}
