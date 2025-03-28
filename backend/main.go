package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// OAuth2 config
var githubOAuthConfig *oauth2.Config

func init() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Set up GitHub OAuth2 config
	githubOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GITHUB_CALLBACK_URL"),
		Scopes:       []string{"read:user", "repo"}, // Read user and repo data
		Endpoint:     github.Endpoint,
	}
}

func main() {
	app := fiber.New()

	// GitHub login route
	app.Get("/api/auth/github", func(c *fiber.Ctx) error {
		url := githubOAuthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOnline)
		return c.Redirect(url)
	})

	// GitHub callback route
	app.Get("/api/auth/callback", func(c *fiber.Ctx) error {
		code := c.Query("code")
		if code == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Code not found"})
		}

		// Exchange code for access token
		token, err := githubOAuthConfig.Exchange(context.Background(), code)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "OAuth exchange failed"})
		}

		// Fetch user details
		client := githubOAuthConfig.Client(context.Background(), token)
		resp, err := client.Get("https://api.github.com/user")
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch user"})
		}
		defer resp.Body.Close()

		var user map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to parse user data"})
		}

		return c.JSON(fiber.Map{
			"token": token.AccessToken,
			"user":  user,
		})
	})

	log.Println("🚀 Server running on http://localhost:3000")
	log.Fatal(app.Listen(":3000"))
}
