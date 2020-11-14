// TODO: check user agent only contains *git* can access info/refs|git-upload-pack|git-receive-pack
// TODO: git protocol version 2 support
// TODO: gzip support
package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

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

func checkRepo(repoPath string) (err error) {
	if !strings.HasSuffix(repoPath, ".git") {
		repoPath = fmt.Sprintf("%s.git", repoPath)
	}

	fileInfo, fileErr := os.Stat(repoPath)

	if fileErr != nil {
		// TODO: fileErr may be a permission error or other
		// init repo if not exists
		initCmd := exec.Command("git", "init", "--bare", repoPath)
		samplePostHook := fmt.Sprintf("%s/hooks/post-update.sample", repoPath)
		PostHook := fmt.Sprintf("%s/hooks/post-update", repoPath)
		activePostHook := exec.Command("cp", samplePostHook, PostHook)
		if err = os.Mkdir(repoPath, 0755); err != nil {
			return err
		}
		if err = initCmd.Run(); err != nil {
			return err
		}
		err = activePostHook.Run()
	} else if fileInfo.IsDir() {
		// TODO: check if dir is a valid bare repo
		err = nil
	} else if fileInfo.Mode().IsRegular() {
		err = errors.New("Not a valid repository")
	}
	return err
}

func handleRefsHeader(w *http.ResponseWriter, service string) {
	cType := fmt.Sprintf("application/x-%s-advertisement", service)
	(*w).Header().Add("Content-Type", cType)
	(*w).Header().Set("Expires", "Fri, 01 Jan 1980 00:00:00 GMT")
	(*w).Header().Set("Pragma", "no-cache")
	(*w).Header().Set("Cache-Control", "no-cache, max-age=0, must-revalidate")
}

func handlePackHeader(w *http.ResponseWriter, service string) {
	(*w).Header().Set("Content-Type", fmt.Sprintf("application/x-%s-result", service))
	(*w).Header().Set("Connection", "Keep-Alive")
	(*w).Header().Set("Transfer-Encoding", "chunked")
	(*w).Header().Set("X-Content-Type-Options", "nosniff")
}
