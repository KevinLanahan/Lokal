package runner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"
)

type sharedStep struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Output string `json:"output,omitempty"`
}

type sharedSession struct {
	Slug         string       `json:"slug"`
	WorkflowName string       `json:"workflow_name"`
	Platform     string       `json:"platform"`
	Steps        []sharedStep `json:"steps"`
}

func stepStatusStr(r stepResult) string {
	switch {
	case r.aborted:
		return "aborted"
	case r.warned:
		return "warned"
	case r.skipped:
		return "skipped"
	case r.passed:
		return "passed"
	default:
		return "failed"
	}
}

func ShareSession(wfName, platform string, results []stepResult) (string, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	anonKey := os.Getenv("SUPABASE_ANON_KEY")
	if supabaseURL == "" || anonKey == "" {
		return "", fmt.Errorf("SUPABASE_URL and SUPABASE_ANON_KEY must be set in .env")
	}

	slug := randomSlug(8)
	steps := make([]sharedStep, len(results))
	for i, r := range results {
		steps[i] = sharedStep{Name: r.name, Status: stepStatusStr(r), Output: r.output}
	}

	session := sharedSession{
		Slug:         slug,
		WorkflowName: wfName,
		Platform:     platform,
		Steps:        steps,
	}

	body, err := json.Marshal(session)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", supabaseURL+"/rest/v1/sessions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("apikey", anonKey)
	req.Header.Set("Authorization", "Bearer "+anonKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=minimal")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("upload failed (status %d) — check your Supabase keys", resp.StatusCode)
	}

	return slug, nil
}

func randomSlug(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[rng.Intn(len(chars))]
	}
	return string(b)
}
