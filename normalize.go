package main

import (
	snowballEnglish "github.com/kljensen/snowball/english"
	"strings"
	"unicode"
)

var stopwordStrings = []string{
	"a", "an", "and", "are", "as", "at", "be", "but", "by", "for", "if", "in",
	"into", "is", "it", "no", "not", "of", "on", "or", "such", "that", "the",
	"their", "then", "there", "these", "they", "this", "to", "was", "will", "with",
}
var stopwords = make(map[string]struct{}, len(stopwordStrings))

// initStopwordsList populates the stopwords list for fast presence testing.
// Go doesn't have a set type, use a map's key hash for quick lookups and
// don't use the values. Use empty struct values, which allocate zero bytes.
func initStopwordsList() {
	for _, word := range stopwordStrings {
		stopwords[word] = struct{}{}
	}
}

// stopwordFilter returns a slice of tokens with stop words removed.
func stopwordFilter(tokens []string) []string {
	r := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if _, ok := stopwords[token]; !ok {
			r = append(r, token)
		}
	}
	return r
}

// tokenize returns a slice of tokens for the given text.
// It splits simply on any character that is not a letter or a number.
func tokenize(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
}

// lowercaseFilter returns a slice of tokens normalized to lower case.
func lowercaseFilter(tokens []string) []string {
	r := make([]string, len(tokens))
	for i, token := range tokens {
		r[i] = strings.ToLower(token)
	}
	return r
}

// stemmerFilter returns a slice of stemmed tokens using the Snowball stemmer
// for English.
func stemmerFilter(tokens []string) []string {
	r := make([]string, len(tokens))
	for i, token := range tokens {
		r[i] = snowballEnglish.Stem(token, false)
	}
	return r
}

// normalize processes an arbitrary English string, returning a slice of tokens
func normalize(text string) []string {
	return stemmerFilter(stopwordFilter(lowercaseFilter(tokenize(text))))
}
