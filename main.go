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
	Avatar      string `json:"avatar_url"`
	Bio         string `json:"bio"`
}

type Dashboard struct {
	Username     string             `json:"username"`
	Repos        int                `json:"repos"`
	Followers    int                `json:"followers"`
	TopLanguages map[string]float32 `json:"top_languages"`
	TopRepos     []Repo             `json:"top_repos"`
	Avatar       string             `json:"avatar"`
	Bio          string             `json:"bio"`
}

var err error

func getDashboard(c *fiber.Ctx) error {
	// Get user
	userResp, err := http.Get("https://api.github.com/users/" + username)
	if err != nil {
		return err
	}
	defer userResp.Body.Close()

	var user User
	json.NewDecoder(userResp.Body).Decode(&user)

	// Get repos
	repoResp, err := http.Get("https://api.github.com/users/" + username + "/repos")
	if err != nil {
		return err
	}
	defer repoResp.Body.Close()

	var repos []Repo
	json.NewDecoder(repoResp.Body).Decode(&repos)

	// Count languages
	languages := map[string]int{}

	client := &http.Client{}
	totalBytes := 0

	for _, repo := range repos {

		url := "https://api.github.com/repos/" + username + "/" + repo.Name + "/languages"
		req, _ := http.NewRequest("GET", url, nil)
		resp, err := client.Do(req)
		if err != nil {
			continue
		}

		var repoLangs map[string]int
		json.NewDecoder(resp.Body).Decode(&repoLangs)
		resp.Body.Close()

		for lang, bytes := range repoLangs {
			languages[lang] += bytes
			totalBytes += bytes
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

	// Find the max bytes used in a language
	max := 0
	for _, b := range languages {
		if b > max {
			max = b
		}
	}

	// Normalize so the top language is 100
	normalized := map[string]float32{}
	for lang, b := range languages {
		normalized[lang] = float32(b) / float32(totalBytes) * 100
	}

	// Replace TopLanguages with normalized
	dashboard := Dashboard{
		Username:     user.Login,
		Repos:        user.PublicRepos,
		Followers:    user.Followers,
		TopLanguages: normalized,
		TopRepos:     topRepos,
		Avatar:       user.Avatar,
		Bio:          user.Bio,
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
