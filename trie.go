package main

import (
	"fmt"

	"github.com/odeke-em/dts/trie"
)

func main() {
	t := trie.New(trie.AsciiAlphabet)
	fmt.Println(t)
}
