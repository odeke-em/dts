package trie

import (
	"fmt"
	"reflect"
	"testing"
)

func TestInit(t *testing.T) {
	tr := New(AsciiAlphabet)
	if tr == nil {
		t.Errorf("Expecting non nil creation")
	}
}

func TestGetAndAdd(t *testing.T) {
	tr := New(AsciiAlphabet)
	value, ok := tr.Get("flux")
	if ok {
		t.Errorf("Not expecting a successful retrieval from this empty map")
	}
	if value != nil {
		t.Errorf("Expecting a nil value as the retrieved result!")
	}

	key, saved := "こんにいちは", "konichiwa"
	evicted := tr.Set(key, saved)
	if evicted != nil {
		t.Errorf("Evicted value should be nil since nothing was set in there before!")
	}

	retr, rOk := tr.Get(key)
	if !rOk {
		t.Errorf("A successful retrieval expected!")
	}
	if retr == nil {
		t.Errorf("Retrieved value cannot be nil!")
	}
	if retr != saved {
		t.Errorf("Expected %v got %v", saved, retr)
	}

	popd, pOk := tr.Pop(key)
	if !pOk {
		t.Errorf("Expected a successful pop!")
	}
	if popd != saved {
		t.Errorf("Pop: Expected %v instead got %v", saved, popd)
	}

	retr2, retr2Ok := tr.Get(key)
	if retr2Ok {
		t.Errorf("Already popped that key was not expecting a successful retrieval!")
	}
	if retr2 != nil {
		t.Errorf("Expected nil, instead got %v", retr2)
	}

	setValue := "кофе"
	evicted1 := tr.Set(key, setValue)
	if evicted1 != nil {
		t.Errorf("Expected no eviction instead got %v", evicted1)
	}

	retr3, retr3Ok := tr.Get(key)
	if !retr3Ok {
		t.Errorf("Expected a successful retrieval!")
	}
	if retr3 == nil {
		t.Errorf("Expected %v instead got nil", setValue)
	}
	if retr3 != setValue {
		t.Errorf("Expected %v instead got %s", setValue, retr3)
	}
}

func TestExpectedWalkOrder(t *testing.T) {
	contentMap := map[string]interface{}{
		"mnt":   t.Errorf,
		"ghost": New,
		"break": TestInit,
	}

	tr := New(AsciiAlphabet)
	for key, value := range contentMap {
		evicted := tr.Set(key, value)
		if evicted != nil {
			t.Errorf("Clash detected after inserting %v, expected nil as eviction result, got %v",
				value, evicted)
		}
	}
	resultsChan := tr.Walk()
	results := make([]interface{}, 0)
	for res := range resultsChan {
		results = append(results, res)
	}
	resLen, expectedLen := len(results), len(contentMap)
	if resLen != expectedLen {
		t.Errorf("Expected full traversal of the map, hence %d values expected got %d",
			resLen, expectedLen)
	}

	ptr := func(v interface{}) uintptr {
		return reflect.ValueOf(v).Pointer()
	}
	ptrEqual := func(a, b interface{}) bool {
		return ptr(a) == ptr(b)
	}

	// Checking the ordering manually! TODO: Organize the expectations and value as kv struct list
	// ie break, ghost, mnt
	if !ptrEqual(results[2], t.Errorf) {
		t.Errorf("Expected %v got %v", t.Errorf, results[2])
	}
	if !ptrEqual(results[1], New) {
		t.Errorf("Expected %v got %v", New, results[1])
	}
	if !ptrEqual(results[0], TestInit) {
		t.Errorf("Expected %v got %v", TestInit, results[0])
	}
}

func TestSet(t *testing.T) {
	sets := map[string][]string{
		"flu":              []string{"patron", "symbolic"},
		"angle":            []string{"vim", "emacs", "sublime"},
		"los angeles":      []string{"ample", "sample"},
		"edmonton":         []string{"connectionz", "first"},
		"frontPcxWUy92kml": []string{"ambelic", ""},
	}

	tr := New(AsciiAlphabet)
	for k, v := range sets {
		tr.Set(k, v[0])
		vLen := len(v)

		var prev interface{}
		for i := 1; i < vLen; i += 1 {
			prev = v[i-1]
			evicted := tr.Set(k, v[i])
			if evicted != prev {
				t.Errorf("expected '%v' got '%v'", prev, evicted)
			}
			prev = evicted
		}
	}
}

func TestTagging(t *testing.T) {
	targets := []string{
		"/usr/", "/mnt/", "/mnt/", "/mnt/ch/", "/mnt/ch/gm", "/mnt/ch/px",
		"/usr/bin", "/usr/lib", "/etc/", "/etc/ssh/", "/etc/ssl",
		"/etc/ssl/certs/own", "/usr/sbin", "/usr/exec",
	}
	tr := New(AsciiAlphabet)
	for _, path := range targets {
		tr.Set(path, path)
	}

	dir := "dir"

	divergentPaths := tr.Tag(PotentialDir, dir)
	if divergentPaths != 4 {
		t.Errorf("Expected 4 divergent paths instead got: %d", divergentPaths)
	}

	matchesChan := tr.Match(PotentialDir)
	for match := range matchesChan {
		fmt.Println(match.Tag, match.Data)
	}

	markedDir := func(tn *TrieNode) bool {
		if tn == nil {
			return false
		}
		cast, ok := tn.Tag.(string)
		return ok && cast == dir
	}

	tagCount := tr.Tag(markedDir, dir)
	fmt.Println("tagCount", tagCount)

	// dsp := tr.Match(markedDir)
	// fmt.Println("dsp", len(dsp), markedDir)

	extracts := tr.MatchAndHarvest(markedDir)
	for extract := range extracts {
		wkc := extract.Walk()

		for wChild := range wkc {
			fmt.Println("curPar", extract.Data, "wChild", wChild)
		}
		// fmt.Println("extracted", extract.Data, "bx",  extract.Eos)
		fmt.Println("done\n")
	}

	wk := tr.Walk()
	for ch := range wk {
		fmt.Println(ch)
	}
}
