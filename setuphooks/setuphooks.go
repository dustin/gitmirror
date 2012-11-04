package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"text/template"

	"github.com/dustin/go-jsonpointer"
)

const base = "https://api.github.com"

var username = flag.String("user", "", "Your github username")
var password = flag.String("pass", "", "Your github password")
var org = flag.String("org", "", "Organization to check")
var noop = flag.Bool("n", false, "If true, don't make any hook changes")
var test = flag.Bool("t", false, "Test all hooks")

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

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n%s [opts]\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nTemplate parameters:\n"+
			" {{.Id}} - Numeric ID of repository\n"+
			" {{.Owner.Login}} - Github username of the repo owner\n"+
			" {{.Owner.Id}} - Numeric ID of the repo owner\n"+
			" {{.Name}} - short name of the repo (e.g. gitmirror)\n"+
			" {{.FullName}} - full name of the repo (e.g. dustin/gitmirror)\n"+
			" {{.Language}} - repository language (if detected)\n")
	}
}

func (h Hook) Test(r Repo) {
	log.Printf("Testing %v -> %v", r.FullName,
		jsonpointer.Get(h.Config, "/url"))
	u := base + "/repos/" + r.FullName + "/hooks/" +
		strconv.Itoa(h.ID) + "/test"

	req, err := http.NewRequest("POST", u, nil)
	maybeFatal("hook test", err)

	req.SetBasicAuth(*username, *password)
	res, err := http.DefaultClient.Do(req)
	maybeFatal("hook test", err)
	maybeHTTPFatal("hook test", 204, res)
	defer res.Body.Close()
}

type Repo struct {
	Id    int
	Owner struct {
		Login string
		Id    int
	}
	Name     string
	FullName string `json:"full_name"`
	Language *string
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
	u := "/user/repos?type=owner"
	if *org != "" {
		u = "/orgs/" + *org + "/repos"
	}
	getJSON("repo list", u, &rv)
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
			h.Test(repo)
			return true
		}
	}
	return false
}

func createHook(r Repo) Hook {
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

	rv := Hook{}
	d := json.NewDecoder(res.Body)
	maybeFatal("creating hook", d.Decode(&rv))
	return rv
}

func updateHooks(r Repo) {
	hooks := []Hook{}
	getJSON(r.FullName, "/repos/"+r.FullName+"/hooks", &hooks)

	if hasMirror(r, hooks) {
		return
	}

	log.Printf("Setting up %v", r.FullName)
	if !*noop {
		createHook(r).Test(r)
	}
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
