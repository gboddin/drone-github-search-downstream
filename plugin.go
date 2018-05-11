package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/drone/drone-go/drone"
	"github.com/google/go-github/github"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"log"
	"math"
	"net/http"
)

// Plugin defines the Downstream plugin parameters.
type Plugin struct {
	Repos          []string
	GithubQuery    string
	GithubToken    string
	DroneServer    string
	DroneToken     string
	Branch         string
	Wait           bool
	IgnoreMissing  bool
	Timeout        time.Duration
	LastSuccessful bool
	Params         []string
	ParamsEnv      []string
}

// Exec runs the plugin
func (p *Plugin) Exec() error {
	if len(p.GithubQuery) == 0 {
		return fmt.Errorf("Error: you must provide a Github repo search query.")
	}

	if len(p.DroneToken) == 0 {
		return fmt.Errorf("Error: you must provide your Drone access token.")
	}

	if len(p.DroneServer) == 0 {
		return fmt.Errorf("Error: you must provide your Drone server.")
	}

	if p.Wait && p.LastSuccessful {
		return fmt.Errorf("Error: only one of wait and last_successful can be true; choose one")
	}

	params, err := parseParams(p.Params)
	if err != nil {
		return fmt.Errorf("Error: unable to parse params: %s.\n", err)
	}

	for _, k := range p.ParamsEnv {
		v, exists := os.LookupEnv(k)
		if !exists {
			return fmt.Errorf("Error: param_from_env %s is not set.\n", k)
		}

		params[k] = v
	}

	populateGithubRepos(p)

	config := new(oauth2.Config)

	auther := config.Client(
		oauth2.NoContext,
		&oauth2.Token{
			AccessToken: p.DroneToken,
		},
	)

	client := drone.NewClient(p.DroneServer, auther)

	for _, entry := range p.Repos {

		// parses the repository name in owner/name@branch format
		owner, name, branch := parseRepoBranch(entry)
		if len(owner) == 0 || len(name) == 0 {
			return fmt.Errorf("Error: unable to parse repository name %s.\n", entry)
		}
		waiting := false

		timeout := time.After(p.Timeout)
		tick := time.Tick(1 * time.Second)

		// Keep trying until we're timed out, successful or got an error
		// Tagged with "I" due to break nested in select
	I:
		for {
			select {
			// Got a timeout! fail with a timeout error
			case <-timeout:
				return fmt.Errorf("Error: timed out waiting on a build for %s.\n", entry)
			// Got a tick, we should check on the build status
			case <-tick:
				// get the latest build for the specified repository
				build, err := client.BuildLast(owner, name, branch)
				if err != nil {
					if waiting {
						continue
					}
					if p.IgnoreMissing {
						fmt.Printf("Error: unable to get latest build for %s, skipping\n", entry)
						break I
					}
					return fmt.Errorf("Error: unable to get latest build for %s.\n", entry)
				}
				if p.Wait && !waiting && (build.Status == drone.StatusRunning || build.Status == drone.StatusPending) {
					fmt.Printf("BuildLast for repository: %s, returned build number: %v with a status of %s. Will retry for %v.\n", entry, build.Number, build.Status, p.Timeout)
					waiting = true
					continue
				} else if p.LastSuccessful && build.Status != drone.StatusSuccess {
					builds, err := client.BuildList(owner, name)
					if err != nil {
						return fmt.Errorf("Error: unable to get build list for %s.\n", entry)
					}

					build = nil
					for _, b := range builds {
						if b.Branch == branch && b.Status == drone.StatusSuccess {
							build = b
							break
						}
					}
					if build == nil {
						return fmt.Errorf("Error: unable to get last successful build for %s.\n", entry)
					}
				}

				if (build.Status != drone.StatusRunning && build.Status != drone.StatusPending) || !p.Wait {
					// start a new  build
					_, err = client.BuildFork(owner, name, build.Number, params)
					if err != nil {
						if waiting {
							continue
						}
						return fmt.Errorf("Error: unable to trigger a new build for %s.\n", entry)
					}
					fmt.Printf("Starting new build %d for %s.\n", build.Number, entry)
					logParams(params, p.ParamsEnv)
					break I
				}
			}
		}
	}
	return nil
}

func parseRepoBranch(repo string) (string, string, string) {
	var (
		owner  string
		name   string
		branch string
	)

	parts := strings.Split(repo, "@")
	if len(parts) == 2 {
		branch = parts[1]
		repo = parts[0]
	}

	parts = strings.Split(repo, "/")
	if len(parts) == 2 {
		owner = parts[0]
		name = parts[1]
	}
	return owner, name, branch
}

func parseParams(paramList []string) (map[string]string, error) {
	params := make(map[string]string)
	for _, p := range paramList {
		parts := strings.SplitN(p, "=", 2)
		if len(parts) == 2 {
			params[parts[0]] = parts[1]
		} else if _, err := os.Stat(parts[0]); os.IsNotExist(err) {
			return nil, fmt.Errorf(
				"invalid param '%s'; must be KEY=VALUE or file path",
				parts[0],
			)
		} else {
			fileParams, err := godotenv.Read(parts[0])
			if err != nil {
				return nil, err
			}

			for k, v := range fileParams {
				params[k] = v
			}
		}
	}

	return params, nil
}

func logParams(params map[string]string, paramsEnv []string) {
	if len(params) > 0 {
		fmt.Println("  with params:")
		for k, v := range params {
			fromEnv := false
			for _, e := range paramsEnv {
				if k == e {
					fromEnv = true
					break
				}
			}
			if fromEnv {
				v = "[from-environment]"
			}
			fmt.Printf("  - %s: %s\n", k, v)
		}
	}
}

func populateGithubRepos(p *Plugin) {
	var tc *http.Client
	if p.GithubToken != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: p.GithubToken},
		)
		tc = oauth2.NewClient(oauth2.NoContext, ts)
	}
	client := github.NewClient(tc)

	query := fmt.Sprintf(p.GithubQuery)

	page := 1
	maxPage := math.MaxInt32

	opts := &github.SearchOptions{
		Sort:  "updated",
		Order: "desc",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for ; page <= maxPage; page++ {
		opts.Page = page
		result, response, err := client.Search.Repositories(oauth2.NoContext, query, opts)
		wait(response)
		if err != nil {
			log.Fatal("FindRepos:", err)
			break
		}
		maxPage = response.LastPage
		for _, repo := range result.Repositories {
			RepoName := *repo.FullName
			if len(p.Branch) > 0 {
				RepoName = fmt.Sprintf("%s@%s", *repo.FullName, p.Branch)
			}
			p.Repos = append(p.Repos, RepoName)
			fmt.Println("Added", RepoName, "to the downstream list.")
		}
	}

}

func wait(response *github.Response) {
	if response != nil && response.Remaining <= 1 {
		gap := time.Duration(response.Reset.Local().Unix() - time.Now().Unix())
		sleep := gap * time.Second
		if sleep < 0 {
			sleep = -sleep
		}

		time.Sleep(sleep)
	}
}
