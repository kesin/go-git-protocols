// TODO: check user agent only contains *git* can access info/refs|git-upload-pack|git-receive-pack
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
