package trie

type keyTranslator func(key string) []int
type indexTranslator func(b byte) int

func indexResolver(f indexTranslator) keyTranslator {
	cache := map[string][]int{}
	return func(key string) []int {
		retr, ok := cache[key]
		if ok {
			return retr
		}

		retr = make([]int, len(key))
		for i, ch := range key {
			retr[i] = f(byte(ch))
		}
		cache[key] = retr
		return retr
	}
}

type trieNode struct {
	data     interface{}
	children *[]*trieNode
	// eos is designated to terminate a sequence. Stands for End-Of-Sequence
	eos bool
}

type alphabet struct {
	max          int
	min          int
	alphabetizer keyTranslator
}

func asciiAlphabetizer(ch byte) int {
	return int(ch)
}

var AsciiAlphabet = &alphabet{
	min:          0,
	max:          255, // TODO: Perform a sizeof(like) op
	alphabetizer: indexResolver(asciiAlphabetizer),
}

type Trie struct {
	root       *trieNode
	translator *alphabet
}

func newTrieNode(data interface{}) *trieNode {
	return &trieNode{
		data:     data,
		children: nil,
	}
}

func (tn *trieNode) pop(alphaIndices []int) (popd interface{}, ok bool) {
	if len(alphaIndices) < 1 {
		if !tn.eos {
			return nil, false
		}
		popd := tn.data
		tn.data = nil
		tn.eos = false

		return popd, true
	}
	if tn.children == nil {
		return nil, false
	}
	children := *tn.children
	first := alphaIndices[0] % len(children)
	if children[first] == nil {
		return nil, false
	}
	return children[first].get(alphaIndices[1:])

}

func (tn *trieNode) get(alphaIndices []int) (value interface{}, ok bool) {
	if len(alphaIndices) < 1 {
		if !tn.eos {
			return nil, false
		}
		return tn.data, true
	}
	if tn.children == nil {
		return nil, false
	}
	children := *tn.children
	first := alphaIndices[0] % len(children)
	if children[first] == nil {
		return nil, false
	}
	return children[first].get(alphaIndices[1:])
}

func (tn *trieNode) add(alphaIndices []int, data interface{}, maxLen int) (prev interface{}, inserted *trieNode) {
	indicesLen := len(alphaIndices)
	if indicesLen < 0 {
		prev = tn.data
		tn.data = data
		tn.eos = true
		return prev, tn
	}

	var children []*trieNode
	if tn.children == nil {
		children = make([]*trieNode, maxLen)
		tn.children = &children
	}

	children = *tn.children
	first := alphaIndices[0] % maxLen
	if children[first] == nil {
		children[first] = newTrieNode(nil)
	}

	child := children[first]
	return child.add(alphaIndices[1:], data, maxLen)
}

func (t *Trie) Add(key string, value interface{}) (prev interface{}) {
	indices := t.translator.alphabetizer(key)
	prev, _ = t.root.add(indices, value, t.translator.max)
	return prev
}

func (t *Trie) Get(key string) (value interface{}, ok bool) {
	indices := t.translator.alphabetizer(key)
	return t.root.get(indices)
}

func (t *Trie) Pop(key string) (popd interface{}, ok bool) {
	indices := t.translator.alphabetizer(key)
	return t.root.pop(indices)
}

func New(alphabetizer *alphabet) *Trie {
	t := &Trie{
		root:       nil,
		translator: alphabetizer,
	}
	return t
}
