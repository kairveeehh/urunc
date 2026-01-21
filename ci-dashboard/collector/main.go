package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// Config: Change these to match your repo
const (
	RepoOwner = "urunc-dev"
	RepoName  = "urunc"
	APIURL    = "https://api.github.com/repos/" + RepoOwner + "/" + RepoName
)

type GitHubRun struct {
	ID           int64     `json:"id"`
	Status       string    `json:"status"` // "completed"
	Conclusion   string    `json:"conclusion"` // "success", "failure"
	CreatedAt    time.Time `json:"created_at"`
	HeadCommit   struct {
		ID      string `json:"id"`
		Message string `json:"message"`
	} `json:"head_commit"`
}

type GitHubJobsResponse struct {
	Jobs []struct {
		Name       string `json:"name"`
		Conclusion string `json:"conclusion"`
		HTMLURL    string `json:"html_url"`
	} `json:"jobs"`
}

type RunListResponse struct {
	WorkflowRuns []GitHubRun `json:"workflow_runs"`
}

// Output Data Structure
type DashboardData struct {
	GeneratedAt  string             `json:"generated_at"`
	LatestCommit CommitInfo         `json:"latest_commit"`
	Jobs         map[string]JobData `json:"jobs"`
}

type CommitInfo struct {
	SHA     string `json:"sha"`
	Message string `json:"message"`
	Date    string `json:"date"`
}

type JobData struct {
	PassRate float64      `json:"pass_rate"`
	History  []RunResult  `json:"history"`
}

type RunResult struct {
	Status string `json:"status"` // "success", "failure", "skipped"
	Link   string `json:"link"`
}

func main() {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		fmt.Println("Error: GITHUB_TOKEN is required")
		os.Exit(1)
	}

	client := &http.Client{}

	// 1. Get Latest Merged PR Info (Run on main)
	fmt.Println("Fetching latest runs...")
	req, _ := http.NewRequest("GET", APIURL+"/actions/runs?branch=main&event=push&per_page=1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil { panic(err) }
	defer resp.Body.Close()

	var latestRunData RunListResponse
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &latestRunData)

	var latestCommit CommitInfo
	if len(latestRunData.WorkflowRuns) > 0 {
		run := latestRunData.WorkflowRuns[0]
		latestCommit = CommitInfo{
			SHA:     run.HeadCommit.ID[:7],
			Message: run.HeadCommit.Message,
			Date:    run.CreatedAt.Format("2006-01-02 15:04"),
		}
	}

	// 2. Get Last 20 Runs for Analysis
	// We filter for a specific workflow if needed, or get all. 
	// Here we get recent runs to extract jobs.
	req, _ = http.NewRequest("GET", APIURL+"/actions/runs?per_page=20", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = client.Do(req)
	body, _ = ioutil.ReadAll(resp.Body)
	var recentRuns RunListResponse
	json.Unmarshal(body, &recentRuns)

	// 3. Aggregate Job Data
	jobStats := make(map[string]*JobData)

	fmt.Printf("Analyzing %d runs for job details...\n", len(recentRuns.WorkflowRuns))

	for _, run := range recentRuns.WorkflowRuns {
		// Fetch jobs for this run
		jobsReq, _ := http.NewRequest("GET", fmt.Sprintf("%s/actions/runs/%d/jobs", APIURL, run.ID), nil)
		jobsReq.Header.Set("Authorization", "Bearer "+token)
		jobsResp, _ := client.Do(jobsReq)
		jobsBody, _ := ioutil.ReadAll(jobsResp.Body)
		
		var jobsData GitHubJobsResponse
		json.Unmarshal(jobsBody, &jobsData)
		jobsResp.Body.Close()

		for _, job := range jobsData.Jobs {
			if _, exists := jobStats[job.Name]; !exists {
				jobStats[job.Name] = &JobData{History: []RunResult{}}
			}
			
			// Map GitHub conclusion to our status
			status := "skipped"
			if job.Conclusion == "success" { status = "pass" }
			if job.Conclusion == "failure" { status = "fail" }

			jobStats[job.Name].History = append(jobStats[job.Name].History, RunResult{
				Status: status,
				Link:   job.HTMLURL,
			})
		}
	}

	// 4. Calculate Pass Rates
	finalJobs := make(map[string]JobData)
	for name, data := range jobStats {
		successCount := 0.0
		totalCount := 0.0
		
		// Limit to last 10 runs for display
		if len(data.History) > 10 {
			data.History = data.History[:10]
		}

		for _, h := range data.History {
			if h.Status != "skipped" {
				totalCount++
				if h.Status == "pass" {
					successCount++
				}
			}
		}

		rate := 0.0
		if totalCount > 0 {
			rate = (successCount / totalCount) * 100
		}
		
		finalJobs[name] = JobData{
			PassRate: rate,
			History:  data.History,
		}
	}

	// 5. Save to JSON
	output := DashboardData{
		GeneratedAt:  time.Now().Format("2006-01-02 15:04:05 UTC"),
		LatestCommit: latestCommit,
		Jobs:         finalJobs,
	}

	file, _ := json.MarshalIndent(output, "", "  ")
// Write to the sibling directory "../web"
	err := ioutil.WriteFile("../web/data.json", file, 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		os.Exit(1)
	}	