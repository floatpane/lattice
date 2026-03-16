package modules

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/floatpane/lattice/pkg/config"
	"github.com/floatpane/lattice/pkg/module"
	"github.com/floatpane/lattice/pkg/registry"
	"github.com/floatpane/lattice/pkg/styles"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func init() {
	registry.Register("github", NewGitHubModule)
}

type GitHubModule struct {
	username string
	token    string
	commits  int
	prs      int
	issues   int
	latest   []string
	status   string
}

type githubDataMsg struct {
	commits int
	prs     int
	issues  int
	latest  []string
	status  string
}

func NewGitHubModule(cfg config.ModuleConfig) module.Module {
	return &GitHubModule{
		username: cfg.Get("username", "GITHUB_USERNAME", ""),
		token:    cfg.Get("token", "GITHUB_TOKEN", ""),
		status:   "Loading…",
	}
}

func (m *GitHubModule) Name() string { return "GITHUB" }

func (m *GitHubModule) Init() tea.Cmd {
	username, token := m.username, m.token
	return func() tea.Msg { return fetchGitHubData(username, token) }
}

func (m *GitHubModule) Update(msg tea.Msg) tea.Cmd {
	if data, ok := msg.(githubDataMsg); ok {
		m.commits = data.commits
		m.prs = data.prs
		m.issues = data.issues
		m.latest = data.latest
		m.status = data.status
		username, token := m.username, m.token
		return tea.Tick(60*time.Second, func(_ time.Time) tea.Msg {
			return fetchGitHubData(username, token)
		})
	}
	return nil
}

func (m *GitHubModule) View(width, height int) string {
	if m.username == "" {
		return lipgloss.NewStyle().Foreground(styles.DimText).Render("Set GITHUB_USERNAME to enable")
	}
	if m.status != "" && m.commits == 0 && m.prs == 0 && m.issues == 0 && len(m.latest) == 0 {
		return lipgloss.NewStyle().Foreground(styles.DimText).Render(m.status)
	}

	stats := lipgloss.JoinVertical(lipgloss.Left,
		styles.RenderStat("Commits", fmt.Sprintf("%d", m.commits)),
		styles.RenderStat("PRs Merged", fmt.Sprintf("%d", m.prs)),
		styles.RenderStat("Issues", fmt.Sprintf("%d", m.issues)),
	)

	latest := ""
	if len(m.latest) > 0 {
		latest = "\n" + lipgloss.NewStyle().Foreground(styles.DimText).Render(
			"RECENT\n"+strings.Join(m.latest, "\n"),
		)
	}

	return stats + latest
}

func (m *GitHubModule) MinSize() (int, int) { return 36, 8 }

func fetchGitHubData(username, token string) githubDataMsg {
	if username == "" {
		return githubDataMsg{status: "No username"}
	}

	url := fmt.Sprintf("https://api.github.com/users/%s/events?per_page=100", username)
	req, _ := http.NewRequest("GET", url, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return githubDataMsg{status: "Network error"}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return githubDataMsg{status: fmt.Sprintf("HTTP %d", resp.StatusCode)}
	}

	type Commit struct {
		Message string `json:"message"`
	}
	type Payload struct {
		Action      string   `json:"action"`
		Size        int      `json:"size"`
		Ref         string   `json:"ref"`
		Head        string   `json:"head"`
		PullRequest struct {
			Merged bool `json:"merged"`
		} `json:"pull_request"`
		Commits []Commit `json:"commits"`
	}
	type Event struct {
		Type      string                `json:"type"`
		CreatedAt time.Time             `json:"created_at"`
		Repo      struct{ Name string } `json:"repo"`
		Payload   Payload               `json:"payload"`
	}

	var events []Event
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return githubDataMsg{status: "Parse error"}
	}

	todayLocal := time.Now().Format("2006-01-02")
	c, p, i := 0, 0, 0
	var latest []string

	for _, e := range events {
		if e.CreatedAt.In(time.Local).Format("2006-01-02") != todayLocal {
			continue
		}

		if e.Type == "PushEvent" && len(latest) < 4 {
			repo := e.Repo.Name
			if len(e.Payload.Commits) > 0 {
				for k := len(e.Payload.Commits) - 1; k >= 0 && len(latest) < 4; k-- {
					msg := styles.Truncate(e.Payload.Commits[k].Message, 30)
					latest = append(latest, fmt.Sprintf("• %s: %s", repo, msg))
				}
			} else {
				msg := "Pushed update"
				if e.Payload.Ref != "" {
					branch := strings.Replace(e.Payload.Ref, "refs/heads/", "", 1)
					msg = fmt.Sprintf("→ %s", branch)
				}
				latest = append(latest, fmt.Sprintf("• %s: %s", repo, msg))
			}
		}

		switch e.Type {
		case "PushEvent":
			n := e.Payload.Size
			if n == 0 {
				n = 1
			}
			c += n
		case "PullRequestEvent":
			if e.Payload.Action == "closed" && e.Payload.PullRequest.Merged {
				p++
			}
		case "IssuesEvent":
			if e.Payload.Action == "closed" {
				i++
			}
		}
	}

	return githubDataMsg{commits: c, prs: p, issues: i, latest: latest}
}
