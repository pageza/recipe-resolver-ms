package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/pageza/recipe-resolver-ms/nlp"

	"github.com/google/uuid"
)

// Recipe defines the structure for a recipe including basic attributes and metadata.
// This structure models the recipes used for matching and is returned in the API response.
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

// newRecipe creates a new Recipe object with the provided details.
// It sets a unique ID (via uuid) and the current UTC timestamps for both creation and update.
func newRecipe(title string, ingredients, steps []string, nutritionalInfo interface{}, allergyDisclaimer string, appliances []string) Recipe {
	now := time.Now().UTC().Format(time.RFC3339)
	return Recipe{
		ID:                uuid.New().String(),
		Title:             title,
		Ingredients:       ingredients,
		Steps:             steps,
		NutritionalInfo:   nutritionalInfo,
		AllergyDisclaimer: allergyDisclaimer,
		Appliances:        appliances,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

// recipesDB simulates an in-memory database of recipes.
// This sample database is used to perform matching based on the incoming query.
var recipesDB = []Recipe{
	newRecipe(
		"Spaghetti Bolognese",
		[]string{"spaghetti", "tomato sauce", "ground beef", "onion", "garlic"},
		[]string{"Boil pasta", "Cook sauce", "Mix and serve"},
		map[string]int{"calories": 400},
		"Contains gluten",
		[]string{"stove"},
	),
	newRecipe(
		"Chicken Salad",
		[]string{"chicken", "lettuce", "tomatoes", "cucumber", "dressing"},
		[]string{"Grill chicken", "Mix vegetables", "Add dressing"},
		map[string]int{"calories": 300},
		"None",
		[]string{"grill"},
	),
}

// resolveRecipe processes the incoming query and determines the best matching recipe.
// The function follows three logical steps:
//
// 1. Exact Match:
//   - It iterates over all recipes and checks for an exact match (case-insensitive)
//     between the recipe title and the queried string.
//   - If found, that recipe is selected as the primary recipe.
//   - Additionally, it collects other recipes whose titles contain the query string
//     as alternative suggestions.
//   - The function then returns the primary recipe along with these alternatives.
//
// 2. Close Match:
//   - If no exact match is found, it searches for recipes where the title contains
//     the query substring (case-insensitive).
//   - If one or more matches are found, the first match is chosen as the primary recipe.
//   - To indicate it is a close match and not an exact one, " (Close Match)" is appended
//     to its title.
//   - Any further close matches are returned as alternative recipes.
//
// 3. No Match Found:
//   - If neither an exact nor a close match is identified, the function generates a new recipe.
//   - The new recipe uses the query as its title and all other fields are initialized as empty or default.
//   - In this case, alternative recipes remain empty.
func resolveRecipe(query string) (Recipe, []Recipe) {
	// Define a local structure to hold each potential match along with its similarity score.
	type match struct {
		recipe     Recipe
		similarity float64
	}
	var matches []match

	// The threshold determines the minimum similarity score required for considering a recipe a match.
	threshold := 0.3

	// Iterate over all recipes in the database and calculate the similarity between the query and each recipe title.
	for _, recipe := range recipesDB {
		// Utilize the JaccardSimilarity function from the nlp module.
		similarity := nlp.JaccardSimilarity(query, recipe.Title)
		// If similarity meets or exceeds the threshold, include this recipe as a potential match.
		if similarity >= threshold {
			matches = append(matches, match{recipe: recipe, similarity: similarity})
		}
	}

	// If no recipes meet the threshold, generate a new Recipe using the query as its title.
	if len(matches) == 0 {
		return newRecipe(query, []string{}, []string{}, map[string]int{}, "", []string{}), nil
	}

	// Sort the collected matches in descending order based on their similarity scores.
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].similarity > matches[j].similarity
	})

	// Select the best matching recipe as the primary match.
	primary := matches[0].recipe
	// Append a " (Close Match)" marker if the best match isn't a perfect (exact) match.
	if matches[0].similarity < 1.0 {
		primary.Title = primary.Title + " (Close Match)"
	}

	// Any additional matches are gathered as alternative recipe suggestions.
	var alternatives []Recipe
	for i := 1; i < len(matches); i++ {
		alternatives = append(alternatives, matches[i].recipe)
	}
	return primary, alternatives
}

// ResolveRequest defines the structure for the incoming JSON payload.
// It represents the user's recipe query.
type ResolveRequest struct {
	Query string `json:"query"`
}

// ResolveResponse defines the structure for the JSON response.
// It includes the primary matching recipe and any alternative suggestions.
type ResolveResponse struct {
	PrimaryRecipe      Recipe   `json:"primary_recipe"`
	AlternativeRecipes []Recipe `json:"alternative_recipes"`
}

// resolveHandler handles POST requests to the /resolve endpoint.
// It validates the request, decodes the JSON payload, applies the recipe resolution logic,
// and returns the matching recipes in the structured JSON response.
func resolveHandler(w http.ResponseWriter, r *http.Request) {
	// Confirm that the request method is POST; otherwise, return a 405 error.
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// Decode the JSON request into a ResolveRequest struct.
	var req ResolveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Query) == "" {
		// If decoding fails or the query is empty, respond with a 400 Bad Request.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request. 'query' field is required and must be a non-empty string."})
		return
	}

	// Use the resolveRecipe function to find the best matching recipe(s) based on the query.
	primary, alternatives := resolveRecipe(req.Query)
	response := ResolveResponse{
		PrimaryRecipe:      primary,
		AlternativeRecipes: alternatives,
	}

	// Set the response headers and send back the JSON-encoded response with a 200 OK status.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Log any error encountered during the encoding process.
		log.Printf("Error encoding response: %v", err)
	}
}

// main initializes the HTTP server, registers the /resolve endpoint handler,
// and starts listening on the port specified by the PORT environment variable (defaults to 3000 if not set).
func main() {
	http.HandleFunc("/resolve", resolveHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Printf("Resolver microservice listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		// If the server cannot start, log the error and terminate the application.
		log.Fatalf("Server failed to start: %v", err)
	}
}
