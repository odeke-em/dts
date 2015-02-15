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

type TrieNode struct {
	Data interface{}
	// Tag should help to annotate a specific feature e.g color, discovered, isDir etc
	Tag      interface{}
	Children *[]*TrieNode
	// Eos(End-Of-Sequence) is designated to terminate a sequence.
	Eos bool
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
	root       *TrieNode
	translator *alphabet
}

func newTrieNode(data interface{}) *TrieNode {
	return &TrieNode{
		Data:     data,
		Children: nil,
	}
}

func (tn *TrieNode) findNode(alphaIndices []int) (cur *TrieNode, ok bool) {
	cur = tn
	i, max := 0, len(alphaIndices)
	for {
		if cur == nil || i >= max {
			break
		}
		next := cur.Children
		if next == nil {
			return nil, false
		}
		children := *next
		first := alphaIndices[i] % len(children)
		cur = children[first]
		i += 1
	}
	if cur == nil || !cur.Eos {
		return nil, false
	}

	return cur, true
}

func (tn *TrieNode) pop(alphaIndices []int) (popd interface{}, ok bool) {
	location, ok := tn.findNode(alphaIndices)
	if !ok || location == nil {
		return nil, false
	}

	popd = location.Data
	location.Data = nil
	location.Eos = false
	// TODO: Perform a check on whether all the Children are set to
	// nil, if so, clean up the array memory, free and set it to nil.

	return popd, true
}

func (tn *TrieNode) tagOn(pass func(*TrieNode) bool, tag interface{}) (count int) {
	count = 0
	defer func() {
		if pass(tn) {
			tn.Tag = tag
			count += 1
		}
	}()

	if tn.Children == nil {
		return
	}

	Children := *tn.Children
	for _, child := range Children {
		if child == nil {
			continue
		}
		count += child.tagOn(pass, tag)
	}

	return count
}

func (tn *TrieNode) match(pass func(*TrieNode) bool, matches *chan *TrieNode) {
	defer func() {
		if pass(tn) {
			*matches <- tn
		}
	}()

	if tn.Children == nil {
		return
	}

	Children := *tn.Children
	for _, child := range Children {
		if child == nil {
			continue
		}
		child.match(pass, &*matches)
	}
	return
}

func (tn *TrieNode) applyOnEos(f func(*TrieNode)) {
	defer func() {
		if tn.Eos {
			f(tn)
		}
	}()

	if tn.Children == nil {
		return
	}

	Children := *tn.Children
	for _, child := range Children {
		if child == nil {
			continue
		}
		child.applyOnEos(f)
	}
	return
}

func (tn *TrieNode) walk() chan interface{} {
	results := make(chan interface{})
	go func() {
		defer func() {
			close(results)
		}()

		if tn.Children == nil {
			return
		}

		Children := *tn.Children
		for _, child := range Children {
			if child == nil {
				continue
			}
			childChan := child.walk()
			for res := range childChan {
				results <- res
			}
			if child.Eos {
				results <- child.Data
			}
		}
	}()
	return results
}

func (tn *TrieNode) get(alphaIndices []int) (value interface{}, ok bool) {
	location, ok := tn.findNode(alphaIndices)
	if !ok || location == nil {
		return nil, false
	}
	return location.Data, true
}

func (tn *TrieNode) set(alphaIndices []int, data interface{}, maxLen int) (prev interface{}, inserted *TrieNode) {
	indicesLen := len(alphaIndices)
	if indicesLen < 1 {
		prev = tn.Data
		tn.Data = data
		tn.Eos = true
		return prev, tn
	}

	var children []*TrieNode
	if tn.Children == nil {
		children = make([]*TrieNode, maxLen)
		tn.Children = &children
	}

	children = *tn.Children
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

func (t *Trie) Apply(f func(*TrieNode)) {
	t.root.applyOnEos(f)
}

func (t *Trie) Tag(pass func(*TrieNode) bool, tag interface{}) int {
	return t.root.tagOn(pass, tag)
}

func (t *Trie) Match(pass func(*TrieNode) bool) (matches chan *TrieNode) {
	matches = make(chan *TrieNode)
	go func() {
		t.root.match(pass, &matches)
		close(matches)
	}()
	return matches
}

func New(alphabetizer *alphabet) *Trie {
	t := &Trie{
		root:       newTrieNode(""),
		translator: alphabetizer,
	}
	return t
}
