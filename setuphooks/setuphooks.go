package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"text/template"

	"github.com/dustin/go-jsonpointer"
)

const base = "https://api.github.com"

var username = flag.String("user", "", "Your github username")
var password = flag.String("pass", "", "Your github password")
var tmplStr = flag.String("template",
	"http://example.com/gitmirror/{{.Owner.Login}}/{{.Name}}.git",
	"Gitmirror dest url pattern")

var tmpl *template.Template

type Hook struct {
	ID     int                    `json:"id,omitempty"`
	URL    string                 `json:"url,omitempty"`
	Name   string                 `json:"name"`
	Events []string               `json:"events,omitempty"`
	Active bool                   `json:"active"`
	Config map[string]interface{} `json:"config"`
}

type Repo struct {
	Id    int
	Owner struct {
		Login string
		Id    int
	}
	Name     string
	FullName string `json:"full_name"`
	Language string
}

func maybeFatal(m string, err error) {
	if err != nil {
		log.Fatalf("%s: %v", m, err)
	}
}

func maybeHTTPFatal(m string, exp int, res *http.Response) {
	if res.StatusCode != exp {
		log.Fatalf("Expected %v %v, got %v", exp, m, res.Status)
	}
}

func getJSON(name, subu string, out interface{}) {
	req, err := http.NewRequest("GET", base+subu, nil)
	maybeFatal(name, err)

	req.SetBasicAuth(*username, *password)
	res, err := http.DefaultClient.Do(req)
	maybeFatal(name, err)
	maybeHTTPFatal(name, 200, res)
	defer res.Body.Close()

	d := json.NewDecoder(res.Body)

	maybeFatal(name, d.Decode(out))
}

func listRepos() []Repo {
	rv := []Repo{}
	getJSON("repo list", "/user/repos?type=owner", &rv)
	return rv
}

func mirrorFor(repo Repo) string {
	b := bytes.Buffer{}
	maybeFatal("executing template", tmpl.Execute(&b, repo))
	return b.String()
}

func contains(haystack []string, needle string) bool {
	for _, n := range haystack {
		if n == needle {
			return true
		}
	}
	return false
}

func hasMirror(repo Repo, hooks []Hook) bool {
	u := mirrorFor(repo)
	for _, h := range hooks {
		if h.Name == "web" && contains(h.Events, "push") &&
			jsonpointer.Get(h.Config, "/url") == u {
			return true
		}
	}
	return false
}

func createHook(r Repo) {
	h := Hook{
		Name:   "web",
		Active: true,
		Events: []string{"push"},
		Config: map[string]interface{}{"url": mirrorFor(r)},
	}
	body, err := json.Marshal(&h)
	maybeFatal("encoding", err)

	req, err := http.NewRequest("POST",
		base+"/repos/"+r.FullName+"/hooks",
		bytes.NewReader(body))
	maybeFatal("creating hook", err)

	req.SetBasicAuth(*username, *password)
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(body))
	res, err := http.DefaultClient.Do(req)
	maybeFatal("creating hook", err)
	maybeHTTPFatal("creating hook", 201, res)
	defer res.Body.Close()
}

func updateHooks(r Repo) {
	hooks := []Hook{}
	getJSON(r.FullName, "/repos/"+r.FullName+"/hooks", &hooks)

	if hasMirror(r, hooks) {
		return
	}

	log.Printf("Setting up %v", r)
	createHook(r)
}

func main() {
	flag.Parse()

	t, err := template.New("u").Parse(*tmplStr)
	maybeFatal("parsing template", err)
	tmpl = t

	repos := listRepos()

	for _, r := range repos {
		updateHooks(r)
	}
}
