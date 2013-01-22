setuphooks is a commandline tool to set up github webhooks on
repositories in bulk.

You can use it to add or remove a webhook destination to all of your
personal repositories, repositories in any specific org you have
access to, or individual repositories.

Hooks can be specified statically, or with a pattern (as I have
different individual hooks for gitmirror for every repo, but they're
all built roughly the same way).

# Installation

This tool is written in [go][go].  To install it (assuming you have your go
environment configured properly):

    go get github.com/dustin/gitmirror/setuphooks

# Usage

```
Usage:
./setuphooks/setuphooks.mac [opts] template

Options:
  -T=false: Test all hooks
  -d=false: Delete, instead of adding a hook.
  -events="push": Comma separated list of events
  -n=false: If true, don't make any hook changes
  -org="": Organization to check
  -pass="": Your github password
  -repo="": Specific repo (default: all)
  -t=false: Test hooks when creating them
  -user="": Your github username
  -v=false: Print more stuff

Template parameters:
  {{.FullName}}     - full name of repo (e.g. dustin/gitmirror)
  {{.Id}}           - numeric ID of repo
  {{.Language}}     - repository language (if detected)
  {{.Name}}         - short name of repo (e.g. gitmirror)
  {{.Owner.Id}}     - github numeric id of repo owner
  {{.Owner.Login}}  - github username of repo owner

Example templates:
  http://example.com/gitmirror/{{.FullName}}.git
  http://example.com/gitmirror/{{.Owner.Login}}/{{.Language}}/{{.Name}}.git
  http://example.com/gitmirror/{{.Name}}.git
```

[go]: http://golang.org/
