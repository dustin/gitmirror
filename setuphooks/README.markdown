setuphooks is a commandline tool to set up github to notify gitmirror
when a repository has changed.

```
Usage:
./setuphooks [opts] template

Options:
  -d=false: Delete, instead of adding a hook.
  -events="push": Comma separated list of events (or default)
  -n=false: If true, don't make any hook changes
  -org="": Organization to check
  -pass="": Your github password
  -repo="": Specific repo (default: all)
  -t=false: Test all hooks
  -user="": Your github username

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
