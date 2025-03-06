package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/pageza/recipe-resolver-ms/generation"
)

func main() {
	// Load environment variables from .env file.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it, ensure environment variables are set.")
	}

	// Define a test query for DeepSeek.
	query := "Test recipe with unique ingredients and flavors"

	// Call the GenerateRecipe function.
	primary, alternatives, err := generation.GenerateRecipe(query)
	if err != nil {
		log.Fatalf("Error generating recipe: %v", err)
	}

	// Combine the results into a single response object.
	response := struct {
		PrimaryRecipe      generation.Recipe   `json:"primary_recipe"`
		AlternativeRecipes []generation.Recipe `json:"alternative_recipes"`
	}{
		PrimaryRecipe:      primary,
		AlternativeRecipes: alternatives,
	}

	// Marshal the response into pretty-printed JSON.
	data, err := json.MarshalIndent(response, "", "    ")
	if err != nil {
		log.Fatalf("Error marshalling JSON: %v", err)
	}

	// Write the JSON data to output.json.
	if err := os.WriteFile("output.json", data, 0644); err != nil {
		log.Fatalf("Error writing output.json: %v", err)
	}

	fmt.Println("Output JSON written to output.json")
}
