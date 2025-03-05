// cursor--Update generation module to support DeepSeek API for recipe generation.
// This module encapsulates the logic to call an external LLM provider to generate a recipe
// when no exact or close match is found in the local database. It allows switching between
// different providers (e.g., OpenAI, DeepSeek) based on environment configuration.
package generation

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"
)

// Recipe defines the structure for a recipe. This structure must be consistent with
// the model expected by the main application. In a production system, consider placing
// this definition into a shared module (e.g., a 'model' package) to avoid duplication.
type Recipe struct {
	ID                string      `json:"id"`
	Title             string      `json:"title"`
	Ingredients       []string    `json:"ingredients"`
	Steps             []string    `json:"steps"`
	NutritionalInfo   interface{} `json:"nutritional_info"`
	AllergyDisclaimer string      `json:"allergy_disclaimer"`
	Appliances        []string    `json:"appliances"`
	CreatedAt         string      `json:"created_at"`
	UpdatedAt         string      `json:"updated_at"`
}

// llmRequest defines the payload for non-DeepSeek API calls.
// For providers compatible with our simple prompt model.
type llmRequest struct {
	Prompt string `json:"prompt"`
}

// LLMResponse defines the structure of the expected response from the LLM endpoint.
// It must include a primary_recipe and an array of alternative_recipes.
type LLMResponse struct {
	PrimaryRecipe      Recipe   `json:"primary_recipe"`
	AlternativeRecipes []Recipe `json:"alternative_recipes"`
}

// GenerateRecipe calls the configured LLM provider endpoint with a structured prompt based
// on the user's recipe query. If the DEEPEEK_API_KEY environment variable is set, it uses DeepSeek's
// API format. Otherwise, it falls back to a default format.
func GenerateRecipe(query string) (Recipe, []Recipe, error) {
	// Retrieve the LLM endpoint URL from environment variables.
	llmEndpoint := os.Getenv("LLM_ENDPOINT")
	if llmEndpoint == "" {
		return Recipe{}, nil, errors.New("LLM_ENDPOINT environment variable not set")
	}

	// Set up an HTTP client with a timeout.
	client := &http.Client{Timeout: 10 * time.Second}

	// Define a structured prompt instructing the LLM to generate a recipe.
	prompt := "Generate a recipe based on the following query: \"" + query + "\". " +
		"Return a JSON object with two keys: 'primary_recipe' and 'alternative_recipes'. " +
		"The 'primary_recipe' should be a JSON object representing the main recipe with keys: " +
		"id, title, ingredients, steps, nutritional_info, allergy_disclaimer, appliances, created_at, and updated_at. " +
		"The 'alternative_recipes' should be an array of recipe objects following the same structure."

	var reqBody []byte
	var err error
	var req *http.Request

	// Check if the DEEPEEK_API_KEY is set to determine if DeepSeek integration should be used.
	deepSeekKey := os.Getenv("DEEPSEEK_API_KEY")
	if deepSeekKey != "" {
		// For DeepSeek, the API expects a payload with "model", "messages", and "stream".
		model := os.Getenv("DEEPSEEK_MODEL")
		if model == "" {
			model = "deepseek-chat"
		}
		// Construct the messages payload.
		payload := struct {
			Model    string              `json:"model"`
			Messages []map[string]string `json:"messages"`
			Stream   bool                `json:"stream"`
		}{
			Model: model,
			Messages: []map[string]string{
				{"role": "system", "content": "You are a helpful assistant."},
				{"role": "user", "content": prompt},
			},
			Stream: false,
		}
		reqBody, err = json.Marshal(payload)
		if err != nil {
			return Recipe{}, nil, err
		}
		// Create the HTTP request for the DeepSeek API.
		req, err = http.NewRequest(http.MethodPost, llmEndpoint, bytes.NewReader(reqBody))
		if err != nil {
			return Recipe{}, nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+deepSeekKey)
	} else {
		// Fallback: use the default API call structure with a simple prompt.
		reqPayload := llmRequest{
			Prompt: prompt,
		}
		reqBody, err = json.Marshal(reqPayload)
		if err != nil {
			return Recipe{}, nil, err
		}
		req, err = http.NewRequest(http.MethodPost, llmEndpoint, bytes.NewReader(reqBody))
		if err != nil {
			return Recipe{}, nil, err
		}
		req.Header.Set("Content-Type", "application/json")
	}

	// Send the HTTP request.
	resp, err := client.Do(req)
	if err != nil {
		return Recipe{}, nil, err
	}
	defer resp.Body.Close()

	// Check the response status.
	if resp.StatusCode != http.StatusOK {
		return Recipe{}, nil, errors.New("LLM endpoint returned non-200 status: " + resp.Status)
	}

	// Decode the JSON response.
	var llmResp LLMResponse
	if err := json.NewDecoder(resp.Body).Decode(&llmResp); err != nil {
		return Recipe{}, nil, err
	}

	// Return the primary recipe and alternative recipes.
	return llmResp.PrimaryRecipe, llmResp.AlternativeRecipes, nil
}
