package trie

import (
	"sync"
)

type keyTranslator func(key string) []int
type indexTranslator func(b byte) int

func indexResolver(f indexTranslator) keyTranslator {
	var cacheMutex sync.Mutex
	cache := map[string][]int{}

	return func(key string) []int {
		cacheMutex.Lock()
		defer cacheMutex.Unlock()

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

var (
	NumericStart = 0
	NumericEnd   = 9
	NumericOther = (NumericEnd - NumericStart) + 1
)

func numericAlphabetizer(b byte) int {
	if b >= '0' && b <= '9' {
		return int(b - '0')
	}
	return NumericOther
}

var NumericAlphabet = &alphabet{
	min:          NumericStart,
	max:          NumericOther + 1,
	alphabetizer: indexResolver(numericAlphabetizer),
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

func (tn *TrieNode) EOS() bool {
	return tn.Eos
}

func (tn *TrieNode) Match(pass func(*TrieNode) bool) (matches chan *TrieNode) {
	return tn.match(pass)
}

func (tn *TrieNode) matchAndHarvest(pass func(*TrieNode) bool) (unravelled chan *TrieNode) {
	unravelled = make(chan *TrieNode)

	go func() {
		matches := tn.Match(pass)

		pops := uint(0)
		done := make(chan bool)

		for ch := range matches {
			unravelled <- ch
			expChan := ch.explore()
			pops += 1
			go func(ccChan *chan *TrieNode) {
				chChan := *ccChan
				for child := range chChan {
					unravelled <- child
				}
				done <- true
			}(&expChan)
		}

		for i := uint(0); i < pops; i += 1 {
			<-done
		}

		close(unravelled)
	}()

	return unravelled
}

func (tn *TrieNode) explore() (chChan chan *TrieNode) {
	chChan = make(chan *TrieNode)

	go func() {
		defer close(chChan)
		if tn == nil || tn.Children == nil {
			return
		}

		children := *(tn.Children)

		pops := uint(0)
		done := make(chan bool)

		for _, child := range children {
			pops += 1
			go func(ctnptr **TrieNode) {
				ctn := *ctnptr
				cchChan := ctn.explore()
				for ch := range cchChan {
					chChan <- ch
				}
				done <- true
			}(&child)
		}

		for i := uint(0); i < pops; i += 1 {
			<-done
		}
	}()

	return
}

func (tn *TrieNode) match(pass func(*TrieNode) bool) (matches chan *TrieNode) {
	matches = make(chan *TrieNode)

	go func() {
		defer func() {
			if pass(tn) {
				matches <- tn
			}
			close(matches)
		}()

		if tn == nil || tn.Children == nil {
			return
		}

		children := *tn.Children
		ticks := make(chan bool)
		spins := uint64(0)

		chanOChan := make(chan chan *TrieNode)

		for _, child := range children {
			if child == nil {
				continue
			}

			spins += 1
			go func(results *chan chan *TrieNode) {
				*results <- child.match(pass)
				ticks <- true
			}(&chanOChan)
		}

		for i := uint64(0); i < spins; i += 1 {
			cchan := <-chanOChan
			<-ticks

			for cch := range cchan {
				matches <- cch
			}
		}
	}()

	return matches
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
			if child.Eos {
				results <- child.Data
			}
			childChan := child.walk()
			for res := range childChan {
				results <- res
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
	cur := tn

	for _, curIndex := range alphaIndices {
		if cur.Children == nil {
			cch := make([]*TrieNode, maxLen)
			cur.Children = &cch
		}

		children := *cur.Children
		mod := curIndex % maxLen

		if children[mod] == nil {
			children[mod] = newTrieNode(nil)
		}

		cur = children[mod]
	}

	prev = cur.Data
	cur.Data = data
	cur.Eos = true
	return prev, cur
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
	return t.root.match(pass)
}

func (t *Trie) MatchAndHarvest(pass func(*TrieNode) bool) (matches chan *TrieNode) {
	return t.root.matchAndHarvest(pass)
}

func potentialDir(t *TrieNode, onTerminal bool) bool {
	if t == nil || t.Children == nil || len(*t.Children) < 1 {
		return false
	}
	if onTerminal && !t.Eos {
		return false
	}
	nonNilCount := 0
	children := *(t.Children)
	for _, child := range children {
		if child != nil {
			nonNilCount += 1
		}
	}
	return nonNilCount >= 2

}

func PotentialTerminalDir(t *TrieNode) bool {
	return potentialDir(t, true)
}

func PotentialDir(t *TrieNode) bool {
	return potentialDir(t, false)
}

func HasEOS(t *TrieNode) bool {
	return t != nil && t.Eos
}

func New(alphabetizer *alphabet) *Trie {
	t := &Trie{
		root:       newTrieNode(""),
		translator: alphabetizer,
	}
	return t
}
