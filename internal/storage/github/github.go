package github

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// GitHubClient is a wrapper for the GitHub image hosting client
type GitHubClient struct {
	token        string
	owner        string
	repo         string
	branch       string
	path         string
	customDomain string
}

// GitHubConfig contains configuration for the GitHub image hosting client
type GitHubConfig struct {
	Token        string // GitHub personal access token
	Owner        string // Repository owner
	Repo         string // Repository name
	Branch       string // Branch name, defaults to main
	Path         string // File storage path, e.g. "images/"
	CustomDomain string // Optional, custom domain such as CDN
}

// NewGitHubClient creates a new GitHub image hosting client
func NewGitHubClient(cfg GitHubConfig) (*GitHubClient, error) {
	if cfg.Token == "" {
		return nil, fmt.Errorf("GitHub access token cannot be empty")
	}

	if cfg.Owner == "" || cfg.Repo == "" {
		return nil, fmt.Errorf("repository owner and name cannot be empty")
	}

	// Set default branch
	branch := cfg.Branch
	if branch == "" {
		branch = "main"
	}

	// Ensure path format is correct
	path := cfg.Path
	if path != "" && !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	return &GitHubClient{
		token:        cfg.Token,
		owner:        cfg.Owner,
		repo:         cfg.Repo,
		branch:       branch,
		path:         path,
		customDomain: cfg.CustomDomain,
	}, nil
}

// UploadFile uploads a local file to GitHub and returns the download URL
func (g *GitHubClient) UploadFile(ctx context.Context, _path string, filename string) (string, error) {
	// Read file content
	fileContent, err := os.ReadFile(_path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	if len(filename) == 0 {
		filename = uuid.New().String()
	}

	fullPath := path.Join(g.path, filename)
	uniqueFileName := filepath.Base(fullPath)

	// Encode file content as Base64
	encodedContent := base64.StdEncoding.EncodeToString(fileContent)

	// Build request body
	type RequestContent struct {
		Message string `json:"message"`
		Content string `json:"content"`
		Branch  string `json:"branch"`
	}

	reqContent := RequestContent{
		Message: fmt.Sprintf("Upload %s", uniqueFileName),
		Content: encodedContent,
		Branch:  g.branch,
	}

	reqBody, err := json.Marshal(reqContent)
	if err != nil {
		return "", fmt.Errorf("failed to serialize request body: %w", err)
	}

	// Build API URL
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", g.owner, g.repo, fullPath)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "PUT", apiURL, strings.NewReader(string(reqBody)))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set request headers
	req.Header.Set("Authorization", "token "+g.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("GitHub API returned error (status code: %d): %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	type ResponseContent struct {
		Content struct {
			DownloadURL string `json:"download_url"`
		} `json:"content"`
	}

	var respContent ResponseContent
	if err := json.NewDecoder(resp.Body).Decode(&respContent); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Build file download URL
	var downloadURL string
	if g.customDomain != "" {
		// Use custom domain
		domain := g.customDomain
		if domain[len(domain)-1] == '/' {
			domain = domain[:len(domain)-1]
		}
		downloadURL = fmt.Sprintf("%s/%s", domain, fullPath)
	} else {
		// Use GitHub raw domain
		downloadURL = fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s",
			g.owner, g.repo, g.branch, fullPath)
	}

	return downloadURL, nil
}

// Upload uploads data from an io.Reader to GitHub and returns the download URL
func (g *GitHubClient) Upload(ctx context.Context, body io.Reader, filename string) (string, error) {
	// Read all data from the reader
	fileContent, err := io.ReadAll(body)
	if err != nil {
		return "", fmt.Errorf("failed to read data: %w", err)
	}

	if len(filename) == 0 {
		filename = uuid.New().String()
	}

	fullPath := path.Join(g.path, filename)
	uniqueFileName := filepath.Base(fullPath)

	// Encode file content as Base64
	encodedContent := base64.StdEncoding.EncodeToString(fileContent)

	// Build request body
	type RequestContent struct {
		Message string `json:"message"`
		Content string `json:"content"`
		Branch  string `json:"branch"`
	}

	reqContent := RequestContent{
		Message: fmt.Sprintf("Upload %s", uniqueFileName),
		Content: encodedContent,
		Branch:  g.branch,
	}

	reqBody, err := json.Marshal(reqContent)
	if err != nil {
		return "", fmt.Errorf("failed to serialize request body: %w", err)
	}

	// Build API URL
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", g.owner, g.repo, fullPath)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "PUT", apiURL, strings.NewReader(string(reqBody)))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set request headers
	req.Header.Set("Authorization", "token "+g.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("GitHub API returned error (status code: %d): %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	type ResponseContent struct {
		Content struct {
			DownloadURL string `json:"download_url"`
		} `json:"content"`
	}

	var respContent ResponseContent
	if err := json.NewDecoder(resp.Body).Decode(&respContent); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Build file download URL
	var downloadURL string
	if g.customDomain != "" {
		// Use custom domain
		domain := g.customDomain
		if domain[len(domain)-1] == '/' {
			domain = domain[:len(domain)-1]
		}
		downloadURL = fmt.Sprintf("%s/%s", domain, fullPath)
	} else {
		// Use GitHub raw domain
		downloadURL = fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s",
			g.owner, g.repo, g.branch, fullPath)
	}

	return downloadURL, nil
}
