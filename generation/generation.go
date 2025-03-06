// cursor--Update generation module to log and verify the Authorization header setting for DeepSeek API calls.
// This updated version adds debug logs to confirm that the Authorization header is correctly attached.
package generation

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
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

// DeepSeekMessage represents the message part of DeepSeek's chat response.
type DeepSeekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// DeepSeekChoice represents a choice in DeepSeek's response.
type DeepSeekChoice struct {
	Index        int             `json:"index"`
	Message      DeepSeekMessage `json:"message"`
	FinishReason string          `json:"finish_reason"`
}

// DeepSeekResponse represents the overall response structure from DeepSeek's API.
type DeepSeekResponse struct {
	ID      string           `json:"id"`
	Object  string           `json:"object"`
	Created int64            `json:"created"`
	Choices []DeepSeekChoice `json:"choices"`
	Usage   interface{}      `json:"usage"`
}

// HTTPClient is a package-level HTTP client which can be overridden in tests.
var HTTPClient = &http.Client{Timeout: 90 * time.Second}

// stripCodeFences removes markdown code fence markers from a string if present.
func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		// Remove the first line containing the opening code fence.
		if i := strings.Index(s, "\n"); i != -1 {
			s = s[i+1:]
		}
		// Remove trailing code fence.
		if i := strings.LastIndex(s, "```"); i != -1 {
			s = s[:i]
		}
	}
	return strings.TrimSpace(s)
}

// GenerateRecipe calls the configured LLM provider endpoint with a structured prompt based
// on the user's recipe query. If the DEEPEEK_API_KEY environment variable is set, it uses DeepSeek's
// API format. Otherwise, it falls back to a default format. It logs the request headers for debugging.
func GenerateRecipe(query string) (Recipe, []Recipe, error) {
	// Retrieve the LLM endpoint URL from environment variables.
	llmEndpoint := os.Getenv("LLM_ENDPOINT")
	if llmEndpoint == "" {
		return Recipe{}, nil, errors.New("LLM_ENDPOINT environment variable not set")
	}

	// Construct the prompt.
	prompt := "Generate a recipe based on the following query: \"" + query + "\". " +
		"Return a JSON object with two keys: 'primary_recipe' and 'alternative_recipes'. " +
		"The 'primary_recipe' should be a JSON object representing the main recipe with keys: " +
		"id, title, ingredients, steps, nutritional_info, allergy_disclaimer, appliances, created_at, and updated_at. " +
		"The 'alternative_recipes' should be an array of recipe objects following the same structure."

	var reqBody []byte
	var err error
	var req *http.Request

	// Check if DEEPEEK_API_KEY is provided to use DeepSeek API.
	deepseekKey := os.Getenv("DEEPSEEK_API_KEY")
	if deepseekKey != "" {
		// Use DeepSeek's expected payload format.
		model := os.Getenv("DEEPSEEK_MODEL")
		if model == "" {
			model = "deepseek-chat"
		}
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
		req, err = http.NewRequest(http.MethodPost, llmEndpoint, bytes.NewReader(reqBody))
		if err != nil {
			return Recipe{}, nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+deepseekKey)
		// Debug log to verify headers are set.
		log.Printf("DeepSeek Request Headers: %+v", req.Header)
	} else {
		// Default API call structure.
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

	start := time.Now()
	resp, err := HTTPClient.Do(req)
	elapsed := time.Since(start)
	log.Printf("DeepSeek API call took %v", elapsed)

	if err != nil {
		return Recipe{}, nil, err
	}
	defer resp.Body.Close()

	// Check if response status is 200 OK.
	if resp.StatusCode != http.StatusOK {
		return Recipe{}, nil, errors.New("LLM endpoint returned non-200 status: " + resp.Status)
	}

	// If using DeepSeek, its response is nested inside a "choices" array.
	if deepseekKey != "" {
		var dsResp DeepSeekResponse
		if err := json.NewDecoder(resp.Body).Decode(&dsResp); err != nil {
			return Recipe{}, nil, err
		}
		if len(dsResp.Choices) == 0 {
			return Recipe{}, nil, errors.New("no choices in DeepSeek response")
		}
		content := dsResp.Choices[0].Message.Content
		cleanContent := stripCodeFences(content)
		log.Printf("Extracted content: %s", cleanContent)

		var llmResp LLMResponse
		if err := json.Unmarshal([]byte(cleanContent), &llmResp); err != nil {
			return Recipe{}, nil, err
		}
		return llmResp.PrimaryRecipe, llmResp.AlternativeRecipes, nil
	} else {
		// Decode the response.
		var llmResp LLMResponse
		if err := json.NewDecoder(resp.Body).Decode(&llmResp); err != nil {
			return Recipe{}, nil, err
		}
		return llmResp.PrimaryRecipe, llmResp.AlternativeRecipes, nil
	}
}
