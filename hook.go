package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"
	"time"
)

var (
	argGitlab  = flag.String("gitlab", "https://gitlab.com", "gitlab instance URL")
	argProject = flag.String("p", "", "gitlab project ID")
	argToken   = flag.String("token", "", "access token")
)

type Payload struct {
	ObjectKind   string `json:"object_kind"`
	Before       string `json:"before"`
	After        string `json:"after"`
	Ref          string `json:"ref"`
	CheckoutSha  string `json:"checkout_sha"`
	UserID       int    `json:"user_id"`
	UserName     string `json:"user_name"`
	UserUsername string `json:"user_username"`
	UserEmail    string `json:"user_email"`
	UserAvatar   string `json:"user_avatar"`
	ProjectID    int    `json:"project_id"`
	Project      struct {
		ID                int         `json:"id"`
		Name              string      `json:"name"`
		Description       string      `json:"description"`
		WebURL            string      `json:"web_url"`
		AvatarURL         interface{} `json:"avatar_url"`
		GitSSHURL         string      `json:"git_ssh_url"`
		GitHTTPURL        string      `json:"git_http_url"`
		Namespace         string      `json:"namespace"`
		VisibilityLevel   int         `json:"visibility_level"`
		PathWithNamespace string      `json:"path_with_namespace"`
		DefaultBranch     string      `json:"default_branch"`
		Homepage          string      `json:"homepage"`
		URL               string      `json:"url"`
		SSHURL            string      `json:"ssh_url"`
		HTTPURL           string      `json:"http_url"`
	} `json:"project"`
	Repository struct {
		Name            string `json:"name"`
		URL             string `json:"url"`
		Description     string `json:"description"`
		Homepage        string `json:"homepage"`
		GitHTTPURL      string `json:"git_http_url"`
		GitSSHURL       string `json:"git_ssh_url"`
		VisibilityLevel int    `json:"visibility_level"`
	} `json:"repository"`
	Commits []struct {
		ID        string    `json:"id"`
		Message   string    `json:"message"`
		Title     string    `json:"title"`
		Timestamp time.Time `json:"timestamp"`
		URL       string    `json:"url"`
		Author    struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"author"`
		Added    []string      `json:"added"`
		Modified []string      `json:"modified"`
		Removed  []interface{} `json:"removed"`
	} `json:"commits"`
	TotalCommitsCount int `json:"total_commits_count"`
}

type Commit struct {
	ID             string    `json:"id"`
	ShortID        string    `json:"short_id"`
	Title          string    `json:"title"`
	AuthorName     string    `json:"author_name"`
	AuthorEmail    string    `json:"author_email"`
	CommitterName  string    `json:"committer_name"`
	CommitterEmail string    `json:"committer_email"`
	CreatedAt      time.Time `json:"created_at"`
	Message        string    `json:"message"`
	CommittedDate  time.Time `json:"committed_date"`
	AuthoredDate   time.Time `json:"authored_date"`
	ParentIds      []string  `json:"parent_ids"`
	LastPipeline   struct {
		ID     int    `json:"id"`
		Ref    string `json:"ref"`
		Sha    string `json:"sha"`
		Status string `json:"status"`
	} `json:"last_pipeline"`
	Stats struct {
		Additions int `json:"additions"`
		Deletions int `json:"deletions"`
		Total     int `json:"total"`
	} `json:"stats"`
	Status string `json:"status"`
	WebURL string `json:"web_url"`
}

func getCommit(id string) (*Commit, error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v4/projects/%s/repository/commits/%s", *argGitlab, *argProject, id), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("PRIVATE-TOKEN", *argToken)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	commit := Commit{}
	if err := json.Unmarshal(body, &commit); err != nil {
		return nil, err
	}
	return &commit, nil
}

func main() {
	commit := &Commit{}
	flag.Parse()

	commit, _ = getCommit("master")

	http.HandleFunc("/page", func(w http.ResponseWriter, r *http.Request) {
		t, err := template.ParseFiles("hook.html")
		if err != nil {
			log.Printf("Error parsing template: %v", err)
			return
		}
		if err := t.Execute(w, commit); err != nil {
			log.Printf("Error executing template: %v", err)
			return
		}
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			return
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}
		payload := Payload{}
		if err := json.Unmarshal(body, &payload); err != nil {
			log.Printf("Error unmarshaling JSON body: %v", err)
			return
		}
		if len(payload.Commits) == 0 {
			log.Printf("no commits, skipping")
			return
		}

		log.Printf("%v - added: %+v, removed: %+v", payload.Commits[0].Title, payload.Commits[0].Added, payload.Commits[0].Removed)
		commit, err = getCommit(payload.Commits[0].ID)
		if err != nil {
			log.Printf("Error getting commit: %v", err)
			return
		}
		log.Printf("added: %d, removed: %d", commit.Stats.Additions, commit.Stats.Deletions)
	})
	if err := http.ListenAndServe(":8001", nil); err != nil {
		panic(err)
	}
}
