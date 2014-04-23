package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var thePath = flag.String("dir", "/tmp", "working directory")
var git = flag.String("git", "/usr/bin/git", "path to git")
var shell = flag.String("shell", "/bin/bash", "path to shell")
var gitCommand = flag.String("command", "git remote update -p", "command to run")
var port = flag.String("port", ":8124", "port to listen on")

type commandRequest struct {
	w       http.ResponseWriter
	abspath string
	bg      bool
	after   time.Time
	cmds    []*exec.Cmd
	ch      chan bool
}

var reqch = make(chan commandRequest, 100)
var updates = map[string]time.Time{}

func exists(path string) (rv bool) {
	rv = true
	if _, err := os.Stat(path); os.IsNotExist(err) {
		rv = false
	}
	return
}

func maybePanic(err error) {
	if err != nil {
		panic(err)
	}
}

func runCommands(w http.ResponseWriter, bg bool,
	abspath string, cmds []*exec.Cmd) {

	stderr := io.Writer(ioutil.Discard)
	stdout := io.Writer(ioutil.Discard)

	if !bg {
		stderr = &bytes.Buffer{}
		stdout = &bytes.Buffer{}
	}

	for _, cmd := range cmds {
		if exists(cmd.Path) {
			log.Printf("Running %v in %v", cmd.Args, abspath)
			fmt.Fprintf(stdout, "# Running %v\n", cmd.Args)
			fmt.Fprintf(stderr, "# Running %v\n", cmd.Args)

			cmd.Stdout = stdout
			cmd.Stderr = stderr
			cmd.Dir = abspath
			err := cmd.Run()

			if err != nil {
				log.Printf("Error running %v in %v:  %v",
					cmd.Args, abspath, err)
				if !bg {
					fmt.Fprintf(stderr,
						"\n[gitmirror internal error:  %v]\n", err)
				}
			}

		}
	}

	if !bg {
		fmt.Fprintf(w, "---- stdout ----\n")
		_, err := stdout.(*bytes.Buffer).WriteTo(w)
		maybePanic(err)
		fmt.Fprintf(w, "\n----\n\n\n---- stderr ----\n")
		_, err = stderr.(*bytes.Buffer).WriteTo(w)
		maybePanic(err)
		fmt.Fprintf(w, "\n----\n")
	}
}

func shouldRun(path string, after time.Time) bool {
	if path == "/tmp" {
		return true
	}
	lastRun := updates[path]
	return lastRun.Before(after)
}

func didRun(path string, t time.Time) {
	updates[path] = t
}

func pathRunner(ch chan commandRequest) {
	for r := range ch {
		if shouldRun(r.abspath, r.after) {
			t := time.Now()
			runCommands(r.w, r.bg, r.abspath, r.cmds)
			didRun(r.abspath, t)
		} else {
			log.Printf("Skipping redundant update: %v", r.abspath)
			if !r.bg {
				fmt.Fprintf(r.w, "Redundant request.")
			}
		}
		select {
		case r.ch <- true:
		default:
		}
	}
}

func commandRunner() {
	m := map[string]chan commandRequest{}

	for r := range reqch {
		ch, running := m[r.abspath]
		if !running {
			ch = make(chan commandRequest, 10)
			m[r.abspath] = ch
			go pathRunner(ch)
		}
		ch <- r
	}
}

func queueCommand(w http.ResponseWriter, bg bool,
	abspath string, cmds []*exec.Cmd) chan bool {
	req := commandRequest{w, abspath, bg, time.Now(),
		cmds, make(chan bool)}
	reqch <- req
	return req.ch
}

func updateGit(w http.ResponseWriter, section string,
	bg bool, payload []byte) bool {

	time.Sleep(5 * time.Second)

	abspath := filepath.Join(*thePath, section)

	if !exists(abspath) {
		if !bg {
			http.Error(w, "Not found", http.StatusNotFound)
		}
		return false
	}

	cmds := []*exec.Cmd{
		exec.Command(*shell, "-c", *gitCommand),
		exec.Command(*git, "gc", "--auto"),
		exec.Command(filepath.Join(abspath, "hooks/post-fetch")),
		exec.Command(filepath.Join(*thePath, "bin/post-fetch")),
	}

	cmds[2].Stdin = bytes.NewBuffer(payload)
	cmds[3].Stdin = bytes.NewBuffer(payload)

	return <-queueCommand(w, bg, abspath, cmds)
}

func getPath(req *http.Request) string {
	return filepath.Clean(filepath.FromSlash(req.URL.Path))[1:]
}

func createRepo(w http.ResponseWriter, section string,
	bg bool, payload []byte) {

	p := struct {
		Repository struct {
			Owner   interface{}
			Private bool
			Name    string
		}
	}{}

	err := json.Unmarshal(payload, &p)
	if err != nil {
		log.Printf("Error unmarshalling data: %v", err)
		http.Error(w, "Error parsing JSON", http.StatusInternalServerError)
		return
	}

	var ownerName string
	switch i := p.Repository.Owner.(type) {
	case string:
		ownerName = i
	case map[string]interface{}:
		ownerName = fmt.Sprintf("%v", i["name"])
	}

	repo := fmt.Sprintf("git://github.com/%v/%v.git",
		ownerName, p.Repository.Name)
	if p.Repository.Private {
		repo = fmt.Sprintf("git@github.com:%v/%v.git",
			ownerName, p.Repository.Name)
	}

	cmds := []*exec.Cmd{
		exec.Command(*git, "clone", "--mirror", "--bare", repo,
			filepath.Join(*thePath, section)),
	}

	if bg {
		w.WriteHeader(201)
	}
	queueCommand(w, true, "/tmp", cmds)
}

func doUpdate(w http.ResponseWriter, path string,
	bg bool, payload []byte) {
	if bg {
		go updateGit(w, path, bg, payload)
		w.WriteHeader(201)
	} else {
		updateGit(w, path, bg, payload)
	}
}

func handleGet(w http.ResponseWriter, req *http.Request, bg bool) {
	path := getPath(req)
	doUpdate(w, path, bg, nil)
}

func handlePost(w http.ResponseWriter, req *http.Request, bg bool) {
	b := []byte(req.FormValue("payload"))

	path := getPath(req)
	abspath := filepath.Join(*thePath, path)

	if exists(abspath) {
		doUpdate(w, path, bg, b)
	} else {
		createRepo(w, path, bg, b)
	}
}

func handleReq(w http.ResponseWriter, req *http.Request) {
	backgrounded := req.FormValue("bg") != "false"

	log.Printf("Handling %v %v", req.Method, req.URL.Path)

	switch req.Method {
	case "GET":
		handleGet(w, req, backgrounded)
	case "POST":
		handlePost(w, req, backgrounded)
	default:
		http.Error(w, "Method not allowed",
			http.StatusMethodNotAllowed)
	}
}

func main() {
	flag.Parse()

	log.SetFlags(log.Lmicroseconds)

	go commandRunner()

	http.HandleFunc("/", handleReq)
	http.HandleFunc("/favicon.ico",
		func(w http.ResponseWriter, req *http.Request) {
			http.Error(w, "No favicon", http.StatusGone)
		})
	log.Fatal(http.ListenAndServe(*port, nil))
}
