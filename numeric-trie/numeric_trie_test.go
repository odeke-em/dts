package numtrie

import (
	"fmt"
	"testing"
)

func TestNumericTrie(t *testing.T) {
	mappings := []struct {
		key   string
		value interface{}
	}{
		{
			key: "1430116589.772026", value: "flux",
		},
		{
			key: "1430114589.772026", value: t,
		},
	}

	trie := New()
	for _, item := range mappings {
		prev := trie.Set(item.key, item.value)
		if prev != nil {
			t.Errorf("not expecting any clashes/vacating, instead got %v", prev)
		}
	}

	walk := trie.Walk()
	for value := range walk {
		fmt.Println("# value", value)
	}
}
