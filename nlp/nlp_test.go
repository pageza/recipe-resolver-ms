// Add basic tests for the NLP module functions in nlp/nlp_test.go
package nlp

import (
	"testing"
)

// TestTokenize verifies that Tokenize correctly converts a string to lower-case tokens.
func TestTokenize(t *testing.T) {
	input := "Hello World"
	tokens := Tokenize(input)
	if len(tokens) != 2 {
		t.Errorf("Expected 2 tokens, got %d", len(tokens))
	}
	if tokens[0] != "hello" || tokens[1] != "world" {
		t.Errorf("Expected tokens 'hello' and 'world', got %v", tokens)
	}
}

// TestJaccardSimilarity verifies that JaccardSimilarity returns the expected similarity score.
func TestJaccardSimilarity(t *testing.T) {
	a := "chicken salad"
	b := "salad with chicken"
	// Expected similarity calculation:
	// a -> {"chicken", "salad"}
	// b -> {"salad", "with", "chicken"}
	// Intersection count = 2, Union count = 3, similarity = 2/3 ~ 0.6667
	sim := JaccardSimilarity(a, b)
	expected := 2.0 / 3.0
	if sim < expected-0.01 || sim > expected+0.01 {
		t.Errorf("Expected similarity around %f, got %f", expected, sim)
	}
}
