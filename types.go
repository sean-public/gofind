package main

import "sort"

type document struct {
	ID    int    `json:"-"`
	URL   string `json:"url"`
	Title string `json:"title"`
	Text  string `json:"-"`
}

type tokenFrequency struct {
	ID    int
	count int
}

// docCache is a temporary store of documents accessible by URL
type docCache map[string]document

// index is an inverted index mapping tokens to slices of tokens with counts
type index map[string][]tokenFrequency

func (idx index) add(doc document) {
	indexLock.Lock()
	defer indexLock.Unlock()

	for _, token := range normalize(doc.Text) {
		freqs, ok := idx[token]
		if ok && freqs[len(freqs)-1].ID > doc.ID {
			continue // skip out-of-order [links repeated by async scrapers]
		} else if ok && freqs[len(freqs)-1].ID == doc.ID {
			freqs[len(freqs)-1].count++
		} else if ok {
			idx[token] = append(freqs, tokenFrequency{doc.ID, 1})
		} else {
			idx[token] = []tokenFrequency{{doc.ID, 1}}
		}
	}
	// Let GC collect the body text, we don't need to keep a copy for
	// search results (unless you want to add highlighted previews)
	doc.Text = ""
	docs[doc.URL] = doc
}

// search returns documents that match the tokens in the provided text
// in descending order by how many total "hits" the document has for the
// tokens in the query
func (idx index) search(text string) (matches []document) {
	var matchHits = make(map[int]int)
	for _, token := range normalize(text) {
		if tokenFreqs, ok := idx[token]; ok {
			for _, tokenFreq := range tokenFreqs {
				matchHits[tokenFreq.ID] += tokenFreq.count
			}
		}
	}
	return sortByScore(matchHits)
}

// sortedMapKeys is necessary because Go randomizes the order of
// items when using range to iterate a map and we want a stable
// result set because many pages have the same score for common terms
func sortedMapKeys(m map[int]int) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

// sortByScore accepts a mapping of Document IDs to their scores
// and sorts them, returning the complete Documents ordered by
// highest match first
func sortByScore(matchHits map[int]int) (d []document) {
	type docScore struct {
		ID    int
		score int
	}
	ss := make([]docScore, 0, len(matchHits))
	for k := range sortedMapKeys(matchHits) {
		ss = append(ss, docScore{k, matchHits[k]})
	}
	sort.SliceStable(ss, func(i, j int) bool {
		return ss[i].score > ss[j].score
	})
	for i := 0; i < len(ss); i++ {
		if doc, ok := getDocByID(ss[i].ID); ok {
			d = append(d, doc)
		}
	}
	return
}

func getDocByID(ID int) (document, bool) {
	for _, doc := range docs {
		if ID == doc.ID {
			return doc, true
		}
	}
	return document{}, false
}
