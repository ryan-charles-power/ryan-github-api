package main

import (
	"encoding/json"
	"net/http"
	"os"
	"sort"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

const username = "ryan-charles-power"

type Repo struct {
	Name     string `json:"name"`
	Language string `json:"language"`
	Stars    int    `json:"stargazers_count"`
}

type User struct {
	Login       string `json:"login"`
	PublicRepos int    `json:"public_repos"`
	Followers   int    `json:"followers"`
}

type Dashboard struct {
	Username     string            `json:"username"`
	Repos        int               `json:"repos"`
	Followers    int               `json:"followers"`
	TopLanguages map[string]int    `json:"top_languages"`
	TopRepos     []Repo            `json:"top_repos"`
}

func getDashboard(c *fiber.Ctx) error {

	// Get user
	userResp, _ := http.Get("https://api.github.com/users/" + username)
	defer userResp.Body.Close()

	var user User
	json.NewDecoder(userResp.Body).Decode(&user)

	// Get repos
	repoResp, _ := http.Get("https://api.github.com/users/" + username + "/repos")
	defer repoResp.Body.Close()

	var repos []Repo
	json.NewDecoder(repoResp.Body).Decode(&repos)

	// Count languages
	languages := map[string]int{}

	for _, repo := range repos {
		if repo.Language != "" {
			languages[repo.Language]++
		}
	}

	// Sort repos by stars
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Stars > repos[j].Stars
	})

	topRepos := repos
	if len(topRepos) > 5 {
		topRepos = topRepos[:5]
	}

	dashboard := Dashboard{
		Username:     user.Login,
		Repos:        user.PublicRepos,
		Followers:    user.Followers,
		TopLanguages: languages,
		TopRepos:     topRepos,
	}

	return c.JSON(dashboard)
}

func main() {

	app := fiber.New()

	app.Use(cors.New())

	app.Get("/api/github/dashboard", getDashboard)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	app.Listen(":" + port)
}