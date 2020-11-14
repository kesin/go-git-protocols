// Usage: ./http-smart -repo=/xxx/xxx/xxx/ -port=8882
package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"os/exec"
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
	fmt.Printf("Smart http server start at port %s on dir %s \n", *port, *repoRoot)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", *port), route); err != nil {
		fmt.Printf("Start failed, error: %s\n", err.Error())
	}
}

func handleRefs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repoName := vars["repo"]
	repoPath := fmt.Sprintf("%s%s", *repoRoot, repoName)
	if err := checkRepo(repoPath); err != nil {
		statusCodeWithMessage(&w, 500, err.Error())
		return
	}

	service := r.FormValue("service")
	if operationIsForbidden(&w, service) {
		return
	}

	// add git specifies header and get formatted body from git command
	pFirst := fmt.Sprintf("# service=%s\n", service) // protocol v1
	handleRefsHeader(&w, service)

	cmdRefs := exec.Command("git", service[4:], "--stateless-rpc", "--advertise-refs", repoPath)
	refsBytes, _ := cmdRefs.Output()
	responseBody := fmt.Sprintf("%04x# service=%s\n0000%s", len(pFirst)+4, service, string(refsBytes))

	_, _ = w.Write([]byte(responseBody))
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

	handlePackHeader(&w, service)

	// a smart way handle stdin and stdout
	cmdPack := exec.Command("git", service[4:], "--stateless-rpc", repoPath)
	cmdStdin, err := cmdPack.StdinPipe()
	cmdStdout, err := cmdPack.StdoutPipe()
	err = cmdPack.Start()
	if err != nil {
		statusCodeWithMessage(&w, 500, err.Error())
		return
	}

	// transfer data
	go func() {
		_, _ = io.Copy(cmdStdin, r.Body)
		_ = cmdStdin.Close()
	}()
	_, _ = io.Copy(w, cmdStdout)
	_ = cmdPack.Wait() // wait for std complete
}
