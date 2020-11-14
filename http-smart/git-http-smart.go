// Usage: ./git-http-smart -repo=/xxx/xxx/xxx/ -port=8882
package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"github.com/gorilla/mux"
)

var repoRoot *string

func main() {
	repoRoot = flag.String("repo", "/Users/zoker/Tmp/repositories/", "Specify a repositories dir.")
	port := flag.String("port", "8882", "Specify a port to start process.")
	flag.Parse()

	route := mux.NewRouter()
	route.Handle("/", http.FileServer(http.Dir(*repoRoot)))
	route.HandleFunc("/{repo}/info/refs", handleRefs)
	route.HandleFunc("/{repo}/{service}", processPack)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", *port), route); err != nil {
		fmt.Printf("Start failed, error: %s\n", err.Error())
	}
}

func processPack(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repoName := vars["repo"]
	// request repo not end with .git is supported with upload-pack
	repoPath := fmt.Sprintf("%s%s", *repoRoot, repoName)
	service := vars["service"]
	if operationIsForbidden(&w, service) {
		return
	}

	// a smart way handle stdin and stdout
	cmdPack := exec.Command("git", service[4:], "--stateless-rpc", repoPath)
	cmdStdin, err := cmdPack.StdinPipe()
	cmdStdout, err := cmdPack.StdoutPipe()
	err = cmdPack.Start()
	if err != nil {
		statusCodeWithMessage(&w, 500, "Server error")
		return
	}

	// transfer data
	go func() { _,_ = io.Copy(cmdStdin, r.Body) }()
	_, _ = io.Copy(w, cmdStdout)
}

func handleRefs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repoName := vars["repo"]
	repoPath := fmt.Sprintf("%s%s", *repoRoot, repoName)
	service := r.FormValue("service")
	if operationIsForbidden(&w, service) {
		return
	}

	// add git specifies header and get formatted body from git command
	cType := fmt.Sprintf("application/x-%s-advertisement", service)
	pFirst := fmt.Sprintf("# service=%s\n", service)
	w.Header().Add("Content-Type", cType)
	w.Header().Add("Cache-Control", "no-cache")

	cmdRefs := exec.Command("git", service[4:], "--stateless-rpc", "--advertise-refs", repoPath)
	refsBytes, _ := cmdRefs.Output()
	responseBody := fmt.Sprintf("%04x# service=%s\n0000%s", len(pFirst)+4, service, string(refsBytes))

	_, _ = w.Write([]byte(responseBody))
}

func operationIsForbidden(w *http.ResponseWriter, service string) (b bool) {
	if service != "git-upload-pack" && service != "git-receive-pack" {
		statusCodeWithMessage(w, 403, "Operation not permitted")
		return true
	}
	return false
}

func statusCodeWithMessage(w *http.ResponseWriter, code int, message string) {
	(*w).WriteHeader(code)
	_, _ = (*w).Write([]byte(message))
}