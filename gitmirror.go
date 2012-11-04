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
	"sync"
	"time"
)

var thePath = flag.String("dir", "/tmp", "working directory")
var git = flag.String("git", "/usr/bin/git", "path to git")

type CommandRequest struct {
	w       http.ResponseWriter
	abspath string
	bg      bool
	after   time.Time
	cmds    []*exec.Cmd
	ch      chan bool
}

var reqch = make(chan CommandRequest, 100)
var updates = map[string]time.Time{}
var updateLock sync.Mutex

func exists(path string) (rv bool) {
	rv = true
	if _, err := os.Stat(path); os.IsNotExist(err) {
		rv = false
	}
	return
}

func runCommands(w http.ResponseWriter, bg bool,
	abspath string, cmds []*exec.Cmd) {

	var stderr io.Writer = ioutil.Discard
	var stdout io.Writer = ioutil.Discard

	if !bg {
		stderr = &bytes.Buffer{}
		stdout = &bytes.Buffer{}
	}

	for _, cmd := range cmds {
		if exists(cmd.Path) {
			log.Printf("Running %v", cmd.Args)
			fmt.Fprintf(stdout, "# Running %v\n", cmd.Args)
			fmt.Fprintf(stderr, "# Running %v\n", cmd.Args)

			cmd.Stdout = stdout
			cmd.Stderr = stderr
			cmd.Dir = abspath
			err := cmd.Run()

			if err != nil {
				log.Printf("Error running command:  %v", err)
				if !bg {
					fmt.Fprintf(stderr,
						"\n[gitmirror internal error:  %v]\n", err)
				}
			}

		}
	}

	if !bg {
		fmt.Fprintf(w, "---- stdout ----\n")
		stdout.(*bytes.Buffer).WriteTo(w)
		fmt.Fprintf(w, "\n----\n\n\n---- stderr ----\n")
		stderr.(*bytes.Buffer).WriteTo(w)
		fmt.Fprintf(w, "\n----\n")
	}
}

func shouldRun(path string, after time.Time) bool {
	updateLock.Lock()
	defer updateLock.Unlock()

	lastRun := updates[path]
	if lastRun.Before(after) {
		return true
	}
	return false
}

func didRun(path string) {
	updateLock.Lock()
	defer updateLock.Unlock()
	updates[path] = time.Now()
}

func commandRunner() {
	for r := range reqch {
		if shouldRun(r.abspath, r.after) {
			runCommands(r.w, r.bg, r.abspath, r.cmds)
			didRun(r.abspath)
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

func queueCommand(w http.ResponseWriter, bg bool,
	abspath string, cmds []*exec.Cmd) chan bool {
	req := CommandRequest{w, abspath, bg, time.Now(),
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
		exec.Command(*git, "remote", "update", "-p"),
		exec.Command(*git, "gc", "--auto"),
		exec.Command(filepath.Join(abspath, "hooks/post-fetch")),
		exec.Command("bin/post-fetch"),
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
	runCommands(w, bg, "/tmp", cmds)
}

func doUpdate(w http.ResponseWriter, path string,
	bg bool, payload []byte) {
	if bg {
		go updateGit(w, path, bg, []byte{})
		w.WriteHeader(201)
	} else {
		updateGit(w, path, bg, []byte{})
	}
}

func handleGet(w http.ResponseWriter, req *http.Request, bg bool) {
	path := getPath(req)
	doUpdate(w, path, bg, []byte{})
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
	log.Fatal(http.ListenAndServe(":8124", nil))
}
