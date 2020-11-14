package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
)

var repoPath *string

func main() {
	repoPath = flag.String("repo", "/Users/zoker/Tmp/xxxx333/testrepo/mirrors.git", "Specify a repository dir.")
	flag.Parse()
	http.HandleFunc("/mirrors.git/info/refs", handleRefs)
	http.HandleFunc("/mirrors.git/", processPack)
	if err := http.ListenAndServe(":8882", nil); err != nil {
		fmt.Printf("Start failed, error: %s\n", err.Error())
	}
}

func processPack(w http.ResponseWriter, r *http.Request) {
	params := strings.Split(r.URL.Path, "/")
	service := params[2] // []/[repo.git]/[git-upload-pack]/[]
	if operationIsForbidden(&w, service) {
		return
	}

	cmdPack := exec.Command("git", service[4:], "--stateless-rpc", *repoPath)
	// a basic way
	//cmdPack.Stdin = r.Body
	//cmdPack.Stdout = w
	//err := cmdPack.Run()

	// a smart way
	cmdStdin, err := cmdPack.StdinPipe()
	cmdStdout, err := cmdPack.StdoutPipe()
	err = cmdPack.Start()

	if err != nil {
		statusCode(&w, 500)
		return
	}

	go func() { _,_ = io.Copy(cmdStdin, r.Body) }()
	_, _ = io.Copy(w, cmdStdout)
}

func handleRefs(w http.ResponseWriter, r *http.Request) {
	service := r.FormValue("service")
	if operationIsForbidden(&w, service) {
		return
	}

	cType := fmt.Sprintf("application/x-%s-advertisement", service)
	pFirst := fmt.Sprintf("# service=%s\n", service)
	w.Header().Add("Content-Type", cType)
	w.Header().Add("Cache-Control", "no-cache")
	cmdRefs := exec.Command("git", service[4:], "--stateless-rpc", "--advertise-refs", *repoPath)
	refsBytes, _ := cmdRefs.Output()
	responseBody := fmt.Sprintf("%04x# service=%s\n0000%s", len(pFirst)+4, service, string(refsBytes))

	_, _ = w.Write([]byte(responseBody))
}

func operationIsForbidden(w *http.ResponseWriter, service string) (b bool) {
	if service != "git-upload-pack" && service != "git-receive-pack" {
		statusCode(w, 403)
		return true
	}
	return false
}

func statusCode(w *http.ResponseWriter, code int) {
	(*w).WriteHeader(code)
}