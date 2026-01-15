package main

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"sort"
	"time"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// --- Helper ---
func formatNumber(n int) string {
	if n >= 1000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000.0)
	}
	return fmt.Sprintf("%d", n)
}

// --- Query 1: MAIN STATS (Heavy Data) ---
type queryStats struct {
	User struct {
		Login string
		AvatarURL string `graphql:"avatarUrl(size: 100)"`
		CreatedAt time.Time
		Followers struct {
			TotalCount int
		}
		Following struct {
			TotalCount int
		}
		// 1. Commits
		ContributionsCollection struct {
			TotalCommitContributions int // Accurate "Year" count
			RestrictedContributionsCount int // Private contribs
			ContributionCalendar struct {
				Weeks []struct {
					ContributionDays []struct{ ContributionCount int }
				}
			}
			CommitContributionsByRepository []struct {
				Repository struct{ Name string }
				Contributions struct {
					Nodes []struct{ OccurredAt time.Time }
				} `graphql:"contributions(first: 20)"`
			} `graphql:"commitContributionsByRepository(maxRepositories: 10)"`
		}
		// 2. Top Repos (Limit 10 for Velocity Calculation)
		Repositories struct {
			TotalCount int
			Nodes []struct {
				Name           string
				IsFork         bool
				StargazerCount int
				ForkCount      int
				Stargazers struct {
					Edges []struct{ StarredAt time.Time }
				} `graphql:"stargazers(last: 50)"`
			}
		} `graphql:"repositories(first: 10, ownerAffiliations: [OWNER], orderBy: {field: STARGAZERS, direction: DESC})"`
		// 3. Activity
		PullRequests struct {
            TotalCount int
            Nodes      []struct {
                Repository struct{ Name string }
                CreatedAt  time.Time
                State      string
            }
        } `graphql:"pullRequests(first: 10, orderBy: {field: CREATED_AT, direction: DESC})"`

        Issues struct {
            TotalCount int
            Nodes      []struct {
                Repository struct{ Name string }
                CreatedAt  time.Time
            }
        } `graphql:"issues(first: 10, orderBy: {field: CREATED_AT, direction: DESC})"`

        MergedPRs struct {
            TotalCount int
        } `graphql:"mergedPRs: pullRequests(states: MERGED)"`
	} `graphql:"user(login: $username)"`
}

// --- Query 2: LANGUAGES (Light Data, High Volume) ---
type queryLangs struct {
	User struct {
		// Fetch 100 to get accurate language mix
		Repositories struct {
			Nodes []struct {
				IsFork bool
				PrimaryLanguage struct {
					Name string
				}
			}
		} `graphql:"repositories(first: 100, ownerAffiliations: [OWNER], orderBy: {field: STARGAZERS, direction: DESC})"`
	} `graphql:"user(login: $username)"`
}

func FetchGitHubData(username, token string, location *time.Location) (DashboardData, error) {
	// 1. Client Setup
	baseClient := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{TLSHandshakeTimeout: 15 * time.Second},
	}
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, baseClient)
	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	client := githubv4.NewClient(oauth2.NewClient(ctx, src))

	// 2. Run Query 1 (Stats)
	var qStats queryStats
	variables := map[string]interface{}{"username": githubv4.String(username)}

	if err := client.Query(context.Background(), &qStats, variables); err != nil {
		return DashboardData{}, fmt.Errorf("GitHub Stats API Error: %v", err)
	}

	// 3. Run Query 2 (Languages)
	var qLangs queryLangs
	if err := client.Query(context.Background(), &qLangs, variables); err != nil {
		return DashboardData{}, fmt.Errorf("GitHub Langs API Error: %v", err)
	}
	starSum := 0
	forksSum := 0
	for _, repo := range qStats.User.Repositories.Nodes {
		starSum += repo.StargazerCount
		forksSum += repo.ForkCount
	}
	// 4. Merge Data
	data := DashboardData{Username: qStats.User.Login, AvatarURL: qStats.User.AvatarURL, Followers: formatNumber(qStats.User.Followers.TotalCount), Following: formatNumber(qStats.User.Following.TotalCount),AccountCreated: qStats.User.CreatedAt, 
		RawFollowers:   qStats.User.Followers.TotalCount,
	RawCommits:     qStats.User.ContributionsCollection.TotalCommitContributions,
		RawStars:       starSum,
		RawForks:       forksSum,
		RawRepos:       qStats.User.Repositories.TotalCount,
		RawPRs:         qStats.User.PullRequests.TotalCount,
		RawPRsMerged:   qStats.User.MergedPRs.TotalCount,
		RawIssues:      qStats.User.Issues.TotalCount,
		RawContributed: qStats.User.ContributionsCollection.RestrictedContributionsCount,
	}

	// A. COMMITS
	commitSum:=0
	var allDays []int
	for _, week := range qStats.User.ContributionsCollection.ContributionCalendar.Weeks {
		for _, day := range week.ContributionDays {
			allDays = append(allDays, day.ContributionCount)
			commitSum += day.ContributionCount
		}
	}
	startIdx := len(allDays) - 90
	if startIdx < 0 { startIdx = 0 }
	data.TotalCommits = allDays[startIdx:]
	data.RawCommits = commitSum
// Calculate Total Stars



	// B. TIME OF DAY
	timeBuckets := make([]int, 8)
	for _, repo := range qStats.User.ContributionsCollection.CommitContributionsByRepository {
		for _, c := range repo.Contributions.Nodes {
			localTime := c.OccurredAt.In(location)
			bucket := localTime.Hour() / 3
			if bucket >= 8 { bucket = 7 }
			timeBuckets[bucket]++
		}
	}
	data.TimeOfDay = timeBuckets

	// C. ACCOUNT VELOCITY (From Top Repos)
	daysMap := make(map[int]int)
	for _, repo := range qStats.User.Repositories.Nodes {
		if repo.IsFork { continue }
		
		// Fill Top Repos List
		if len(data.TopRepos) < 3 {
			data.TopRepos = append(data.TopRepos, RepoItem{
				Name: repo.Name, Stars: formatNumber(repo.StargazerCount), Forks: formatNumber(repo.ForkCount),
			})
		}
		
		// Calculate Velocity
		for _, edge := range repo.Stargazers.Edges {
			daysAgo := int(time.Since(edge.StarredAt).Hours() / 24)
			if daysAgo < 60 { daysMap[daysAgo]++ }
		}
	}
	var starDelta []int
	for i := 60; i >= 0; i-- {
		starDelta = append(starDelta, daysMap[i])
	}
	data.StarHistory = starDelta
	data.ForkHistory = starDelta

	// D. LANGUAGES (From Query 2 - Light Data, Size 100)
	langMap := make(map[string]int)
	totalLang := 0
	
	for _, repo := range qLangs.User.Repositories.Nodes {
		if repo.IsFork { continue } // Correctly ignores forks
		
		if repo.PrimaryLanguage.Name != "" {
			langMap[repo.PrimaryLanguage.Name]++
			totalLang++
		}
	}

	var tempLangs []LanguageItem
	for name, count := range langMap {
		pct := int(math.Round((float64(count) / float64(totalLang)) * 100))
		tempLangs = append(tempLangs, LanguageItem{Name: name, Percent: pct})
	}
	sort.Slice(tempLangs, func(i, j int) bool { return tempLangs[i].Percent > tempLangs[j].Percent })
	
	sum := 0
	for i, l := range tempLangs {
		if i < 3 {
			data.Languages = append(data.Languages, l)
			sum += l.Percent
		}
	}
	if sum < 100 {
		data.Languages = append(data.Languages, LanguageItem{Name: "Other", Percent: 100 - sum})
	}

	// E. RECENT ACTIVITY
	cutoff := time.Now().Add(-24 * time.Hour)
	var rawAct []ActivityItem

	for _, repo := range qStats.User.ContributionsCollection.CommitContributionsByRepository {
		cCount := 0
		for _, c := range repo.Contributions.Nodes {
			if c.OccurredAt.After(cutoff) { cCount++ }
		}
		if cCount > 0 {
			rawAct = append(rawAct, ActivityItem{
				Action: fmt.Sprintf("Pushed %d commits to", cCount), Repo: repo.Repository.Name, Type: "push",
			})
		}
	}
	for _, pr := range qStats.User.PullRequests.Nodes {
		if pr.CreatedAt.After(cutoff) {
			act := "Opened PR in"
			if pr.State == "MERGED" { act = "Merged PR in" }
			rawAct = append(rawAct, ActivityItem{Action: act, Repo: pr.Repository.Name, Type: "pr"})
		}
	}
	for _, issue := range qStats.User.Issues.Nodes {
		if issue.CreatedAt.After(cutoff) {
			rawAct = append(rawAct, ActivityItem{Action: "Opened Issue in", Repo: issue.Repository.Name, Type: "issue"})
		}
	}
	
	if len(rawAct) == 0 {
		rawAct = append(rawAct, ActivityItem{Action: "No public activity", Repo: "", Type: "none"})
	}
	
	for i, act := range rawAct {
		if i >= 4 { break }
		data.RecentActivity = append(data.RecentActivity, act)
	}

	return data, nil
}