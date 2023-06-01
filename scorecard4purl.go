package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Package struct {
	Name          string `json:"name"`
	Ecosystem     string `json:"ecosystem"`
	Description   string `json:"description"`
	Homepage      string `json:"homepage"`
	RepositoryURL string `json:"repository_url"`
}

type ScorecardCheck struct {
	Name   string `json:"name"`
	Score  int    `json:"score"`
	Reason string `json:"reason"`
}

type Repo struct {
	Name   string `json:"name"`
	Commit string `json:"commit"`
}

type ScorecardVersion struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

type ScorecardResult struct {
	Date     string           `json:"date"`
	Repo     Repo             `json:"repo"`
	Score    float64          `json:"score"`
	Checks   []ScorecardCheck `json:"checks"`
	Scorecard ScorecardVersion `json:"scorecard"`
	Metadata string           `json:"metadata"`
}

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s <purl>", os.Args[0])
	}

	purl := os.Args[1]

	// Step 1: Retrieve the GitHub URL from Ecosyste.ms API
	githubURL, err := getGitHubURL(purl)
	if err != nil {
		log.Fatal("Failed to retrieve GitHub URL:", err)
	}

	fmt.Println("GitHub URL:", githubURL)

	// Step 2: Retrieve the scorecard data from Security Scorecards API
	platform, org, repo, err := parseGitHubURL(githubURL)
	if err != nil {
		log.Fatal(err)
	}

	result, err := getScorecard(platform, org, repo)
	if err != nil {
		log.Fatal(err)
	}

	for _, check := range result.Checks {
		fmt.Printf("%s: %d\n", check.Name, check.Score)
	}
}

func getGitHubURL(purl string) (string, error) {
	apiURL := "https://packages.ecosyste.ms/api/v1/packages/lookup"

	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return "", err
	}

	q := req.URL.Query()
	q.Add("purl", purl)
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var lookupResults []Package
	if err := json.NewDecoder(resp.Body).Decode(&lookupResults); err != nil {
		return "", err
	}

	if len(lookupResults) == 0 {
		return "", fmt.Errorf("no GitHub URL found for the provided PURL")
	}

	return lookupResults[0].RepositoryURL, nil
}

func getScorecard(platform, org, repo string) (*ScorecardResult, error) {
	url := fmt.Sprintf("https://api.securityscorecards.dev/projects/%s/%s/%s", platform, org, repo)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code %d", resp.StatusCode)
	}

	var result ScorecardResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func parseGitHubURL(input string) (platform, org, repo string, err error) {
	u, err := url.Parse(input)
	if err != nil {
		return "", "", "", err
	}

	if u.Host != "github.com" {
		return "", "", "", fmt.Errorf("only GitHub URLs are supported")
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("invalid GitHub URL")
	}

	return u.Host, parts[0], parts[1], nil
}
