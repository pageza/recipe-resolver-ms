package nlp

import (
	"strings"
)

// Tokenize converts a string to lowercase and splits it into words.
// This simple NLP step helps in comparing the similarity between queries and recipe titles.
func Tokenize(s string) []string {
	return strings.Fields(strings.ToLower(s))
}

// JaccardSimilarity computes the Jaccard similarity coefficient between two strings.
// It tokenizes both strings and calculates the ratio of the size of the intersection
// to the size of the union of the token sets.
func JaccardSimilarity(a, b string) float64 {
	tokensA := Tokenize(a)
	tokensB := Tokenize(b)

	setA := make(map[string]bool)
	setB := make(map[string]bool)

	for _, token := range tokensA {
		setA[token] = true
	}
	for _, token := range tokensB {
		setB[token] = true
	}

	intersectionCount := 0
	unionSet := make(map[string]bool)

	// Count the intersection and build the union set.
	for token := range setA {
		unionSet[token] = true
		if setB[token] {
			intersectionCount++
		}
	}
	for token := range setB {
		unionSet[token] = true
	}

	unionCount := len(unionSet)
	if unionCount == 0 {
		return 0.0
	}
	return float64(intersectionCount) / float64(unionCount)
}
