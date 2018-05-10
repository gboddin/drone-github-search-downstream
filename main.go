package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	version = "0.0.0"
	build   = "0"
)

func main() {
	app := cli.NewApp()
	app.Name = "github search downstream plugin"
	app.Usage = "github search downstream plugin"
	app.Version = fmt.Sprintf("%s+%s", version, build)
	app.Action = run
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "github-query",
			Usage:  "Github repo search query",
			EnvVar: "GITHUB_SEARCH_DOWNSTREAM_GITHUB_QUERY,PLUGIN_GITHUB_QUERY",
		},
		cli.StringFlag{
			Name:   "github-token",
			Usage:  "Github auth token",
			EnvVar: "GITHUB_TOKEN,GITHUB_SEARCH_DOWNSTREAM_GITHUB_TOKEN,PLUGIN_GITHUB_TOKEN",
		},
		cli.StringFlag{
			Name:   "branch",
			Usage:  "Remote branch to trigger",
			EnvVar: "GITHUB_SEARCH_DOWNSTREAM_BRANCH,PLUGIN_BRANCH",
		},
		cli.StringFlag{
			Name:   "drone-server",
			Usage:  "Trigger a drone build on a custom server",
			EnvVar: "GITHUB_SEARCH_DOWNSTREAM_DRONE_SERVER,PLUGIN_DRONE_SERVER",
		},
		cli.StringFlag{
			Name:   "drone-token",
			Usage:  "Drone API token from your user settings",
			EnvVar: "DRONE_TOKEN,GITHUB_SEARCH_DOWNSTREAM_DRONE_TOKEN,PLUGIN_DRONE_TOKEN",
		},
		cli.BoolFlag{
			Name:   "fork",
			Usage:  "Trigger a new build for a repository",
			EnvVar: "PLUGIN_FORK",
		},
		cli.BoolFlag{
			Name:   "wait",
			Usage:  "Wait for any currently running builds to finish",
			EnvVar: "PLUGIN_WAIT",
		},
		cli.DurationFlag{
			Name:   "timeout",
			Value:  time.Duration(60) * time.Second,
			Usage:  "How long to wait on any currently running builds",
			EnvVar: "PLUGIN_WAIT_TIMEOUT",
		},
		cli.BoolFlag{
			Name:   "last-successful",
			Usage:  "Trigger last successful build",
			EnvVar: "PLUGIN_LAST_SUCCESSFUL",
		},
		cli.StringSliceFlag{
			Name:   "params",
			Usage:  "List of params (key=value or file paths of params) to pass to triggered builds",
			EnvVar: "PLUGIN_PARAMS",
		},
		cli.StringSliceFlag{
			Name:   "params-from-env",
			Usage:  "List of environment variables to pass to triggered builds",
			EnvVar: "PLUGIN_PARAMS_FROM_ENV",
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(c *cli.Context) error {
	plugin := Plugin{
		GithubQuery:    c.String("github-query"),
		GithubToken:    c.String("github-token"),
		DroneServer:         c.String("drone-server"),
		DroneToken:          c.String("drone-token"),
		Fork:           c.Bool("fork"),
		Wait:           c.Bool("wait"),
		Timeout:        c.Duration("timeout"),
		LastSuccessful: c.Bool("last-successful"),
		Params:         c.StringSlice("params"),
		ParamsEnv:      c.StringSlice("params-from-env"),
	}

	return plugin.Exec()
}
