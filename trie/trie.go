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
	max:          255, // TODO: Perform a sizeof(byte) op
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

func (tn *trieNode) findNode(alphaIndices []int) (cur *trieNode, ok bool) {
	cur = tn
	i, max := 0, len(alphaIndices)
	for {
		if cur == nil || i >= max {
			break
		}
		next := cur.children
		if next == nil {
			return nil, false
		}
		children := *next
		first := alphaIndices[i] % len(children)
		cur = children[first]
		i += 1
	}
	if cur == nil || !cur.eos {
		return nil, false
	}

	return cur, true
}

func (tn *trieNode) pop(alphaIndices []int) (popd interface{}, ok bool) {
	location, ok := tn.findNode(alphaIndices)
	if !ok || location == nil {
		return nil, false
	}

	popd = location.data
	location.data = nil
	location.eos = false
	// TODO: Perform a check on whether all the children are set to
	// nil, if so, clean up the array memory, free and set it to nil.

	return popd, true
}

func (tn *trieNode) walk() chan interface{} {
	results := make(chan interface{})
	go func() {
		defer func() {
			close(results)
		}()

		if tn.children == nil {
			return
		}

		children := *tn.children
		for _, child := range children {
			if child == nil {
				continue
			}
			childChan := child.walk()
			for res := range childChan {
				results <- res
			}
			if child.eos {
				results <- child.data
			}
		}
	}()
	return results
}

func (tn *trieNode) get(alphaIndices []int) (value interface{}, ok bool) {
	location, ok := tn.findNode(alphaIndices)
	if !ok || location == nil {
		return nil, false
	}
	return location.data, true
}

func (tn *trieNode) set(alphaIndices []int, data interface{}, maxLen int) (prev interface{}, inserted *trieNode) {
	indicesLen := len(alphaIndices)
	if indicesLen < 1 {
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
	return child.set(alphaIndices[1:], data, maxLen)
}

func (t *Trie) Set(key string, value interface{}) (prev interface{}) {
	indices := t.translator.alphabetizer(key)
	prev, _ = t.root.set(indices, value, t.translator.max)
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

func (t *Trie) Walk() chan interface{} {
	return t.root.walk()
}

func New(alphabetizer *alphabet) *Trie {
	t := &Trie{
		root:       newTrieNode(""),
		translator: alphabetizer,
	}
	return t
}
