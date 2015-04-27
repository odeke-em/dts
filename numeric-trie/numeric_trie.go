package numtrie

import "github.com/odeke-em/dts/trie"

func New() *trie.Trie {
	return trie.New(trie.NumericAlphabet)
}
