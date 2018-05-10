# drone-github-search-downstream

[![Build Status](https://hold-on.nobody.run/api/badges/gboddin/drone-github-search-downstream/status.svg)](http://hold-on.nobody.run/drone-github-search-downstream)

Drone plugin to trigger downstream repository builds based on github search. For the usage information and a listing of the available options please take a look at [the docs](README.md).

## Credits

The plugin is based on [plugins/downstream](https://github.com/drone-plugins/drone-downstream/).

Thanks should be givent to @tboerger and its contributors.

## Build

Build the binary with the following commands:

```
go build
```

## Docker

Build the Docker image with the following commands:

```
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -tags netgo -o release/linux/amd64/github-search-downstream
docker build --rm -t gboo/github-search-downstream .
```

## Usage

Execute from the working directory:

```sh
docker run --rm \
  -e PLUGIN_GITHUB_QUERY="org:drone-plugins topic:drone-plugin" \
  -e PLUGIN_DRONE_TOKEN=eyJhbFciHiJISzI1EiIsUnR5cCW6IkpXQCJ9.ezH0ZXh0LjoidGJvZXJnZXIiLCJ0eXBlIjoidXNlciJ9.1m_3QFA6eA7h4wrBby2aIRFAEhQWPrlj4dsO_Gfchtc \
  -v $(pwd):$(pwd) \
  -w $(pwd) \
  gboo/github-search-downstream
```

From Drone:

```yaml
pipeline:
  trigger-downstream:
    image: gboo/github-search-downstream
    github_query: "org:drone-plugins topic:drone-plugin"
    branch: master
    drone_server: https://hold-on.nobody.run
    secrets: [ github_token, drone_token ]
```