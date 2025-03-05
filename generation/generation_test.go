// cursor--Add tests for the generation module in generation/generation_test.go.
// These tests use a mock HTTP server to simulate the LLM provider endpoint.
package generation

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

// mockLLMResponse creates a mock response that the LLM endpoint might return.
func mockLLMResponse() LLMResponse {
	return LLMResponse{
		PrimaryRecipe: Recipe{
			ID:                "mock-id-123",
			Title:             "Mock Recipe (Generated)",
			Ingredients:       []string{"ingredient1", "ingredient2"},
			Steps:             []string{"step1", "step2"},
			NutritionalInfo:   map[string]int{"calories": 500},
			AllergyDisclaimer: "None",
			Appliances:        []string{"oven"},
			CreatedAt:         time.Now().Format(time.RFC3339),
			UpdatedAt:         time.Now().Format(time.RFC3339),
		},
		AlternativeRecipes: []Recipe{
			{
				ID:                "mock-id-456",
				Title:             "Alternative Mock Recipe",
				Ingredients:       []string{"ingredientA", "ingredientB"},
				Steps:             []string{"stepA", "stepB"},
				NutritionalInfo:   map[string]int{"calories": 400},
				AllergyDisclaimer: "None",
				Appliances:        []string{"stove"},
				CreatedAt:         time.Now().Format(time.RFC3339),
				UpdatedAt:         time.Now().Format(time.RFC3339),
			},
		},
	}
}

// TestGenerateRecipe verifies that GenerateRecipe correctly calls the LLM endpoint and parses its response.
func TestGenerateRecipe(t *testing.T) {
	// Create a mock LLM endpoint using httptest.
	mockResponse := mockLLMResponse()
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request method and content type.
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Optionally, you can decode the request payload and check the prompt.
		var reqPayload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&reqPayload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// Return the mock response.
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer mockServer.Close()

	// Set the LLM_ENDPOINT environment variable to point to the mock server.
	os.Setenv("LLM_ENDPOINT", mockServer.URL)

	// Call GenerateRecipe with a test query.
	query := "Generate a recipe for a unique test dish"
	primary, alternatives, err := GenerateRecipe(query)
	if err != nil {
		t.Fatalf("GenerateRecipe returned error: %v", err)
	}

	// Verify that the primary recipe matches the mock data.
	if primary.ID != mockResponse.PrimaryRecipe.ID {
		t.Errorf("Expected primary recipe ID %s, got %s", mockResponse.PrimaryRecipe.ID, primary.ID)
	}
	if len(alternatives) != len(mockResponse.AlternativeRecipes) {
		t.Errorf("Expected %d alternative recipes, got %d", len(mockResponse.AlternativeRecipes), len(alternatives))
	}
}
