// cursor--Add tests for resolveRecipe function and the /resolve HTTP handler in main_test.go
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestResolveRecipeExact verifies that an exact query returns the expected recipe.
func TestResolveRecipeExact(t *testing.T) {
	// Query exactly matches "Spaghetti Bolognese" in recipesDB.
	primary, alternatives := resolveRecipe("Spaghetti Bolognese")
	if !strings.EqualFold(primary.Title, "Spaghetti Bolognese") {
		t.Errorf("Expected primary title 'Spaghetti Bolognese', got '%s'", primary.Title)
	}
	if len(alternatives) != 0 {
		t.Errorf("Expected no alternatives, got %d", len(alternatives))
	}
}

// TestResolveRecipeNoMatch verifies that a query with low similarity generates a new recipe.
func TestResolveRecipeNoMatch(t *testing.T) {
	// "chicken noodle soup" does not sufficiently match any recipe in recipesDB.
	primary, alternatives := resolveRecipe("chicken noodle soup")
	if primary.Title != "chicken noodle soup" {
		t.Errorf("Expected new generated recipe with title 'chicken noodle soup', got '%s'", primary.Title)
	}
	if len(alternatives) != 0 {
		t.Errorf("Expected no alternatives for a new generated recipe, got %d", len(alternatives))
	}
}

// TestResolveRecipeNLP verifies that a loosely matching query returns a close match.
func TestResolveRecipeNLP(t *testing.T) {
	// "Salad with chicken" should closely match "Chicken Salad" in recipesDB.
	primary, alternatives := resolveRecipe("Salad with chicken")
	if !strings.Contains(primary.Title, "Chicken Salad") || !strings.Contains(primary.Title, "(Close Match)") {
		t.Errorf("Expected primary title to contain 'Chicken Salad (Close Match)', got '%s'", primary.Title)
	}
	// There should be no alternative recipes in this simple test scenario.
	if len(alternatives) != 0 {
		t.Errorf("Expected no alternatives, got %d", len(alternatives))
	}
}

// TestResolveHandler verifies the behavior of the /resolve HTTP endpoint.
func TestResolveHandler(t *testing.T) {
	// Prepare a JSON payload with a valid query.
	reqBody, err := json.Marshal(ResolveRequest{Query: "Spaghetti Bolognese"})
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	// Create a new HTTP POST request to the /resolve endpoint.
	req, err := http.NewRequest(http.MethodPost, "/resolve", bytes.NewReader(reqBody))
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}

	// Use httptest to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(resolveHandler)
	handler.ServeHTTP(rr, req)

	// Verify the HTTP status code.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected HTTP status %d, got %d", http.StatusOK, status)
	}

	// Verify the Content-Type header.
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", ct)
	}

	// Decode the JSON response.
	var res ResolveResponse
	if err := json.NewDecoder(rr.Body).Decode(&res); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	// Verify that the primary recipe matches the expected title.
	if !strings.EqualFold(res.PrimaryRecipe.Title, "Spaghetti Bolognese") {
		t.Errorf("Expected primary recipe 'Spaghetti Bolognese', got '%s'", res.PrimaryRecipe.Title)
	}
}
