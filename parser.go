package gitprompt

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// GitStatus is the parsed status for the current state in git.
type GitStatus struct {
	Sha       string
	Branch    string
	Untracked int
	Modified  int
	Staged    int
	Conflicts int
	Ahead     int
	Behind    int
	Stashed   int
	Upstream  string
	Clean     bool
}

// Parse parses the status for the repository from git. Returns nil if the
// current directory is not part of a git repository.
func Parse() (*GitStatus, error) {

	stat, err := runGitCommand("git", "status", "--branch", "--porcelain=2")
	if err != nil {
		if strings.HasPrefix(err.Error(), "fatal:") {
			return nil, nil
		}
		return nil, err
	}

	status := &GitStatus{}

	lines := strings.Split(stat, "\n")
	for _, line := range lines {
		switch line[0] {
		case '#':
			parseHeader(line, status)
		case '?':
			status.Untracked++
		case 'u':
			status.Conflicts++
		case '1', '2':
			if line[2] != '.' {
				status.Staged++
			}
			if line[3] != '.' {
				status.Modified++
			}
		}
	}

	status.Clean = !(status.Untracked != 0 ||
		status.Modified != 0 ||
		status.Staged != 0 ||
		status.Conflicts != 0)

	if stashed, err := runGitCommand("git", "rev-list", "--walk-reflogs", "--count", "refs/stash"); err == nil {
		if s, err := strconv.Atoi(stashed); err == nil {
			status.Stashed = s
		}
	}

	return status, nil

}

func parseHeader(h string, s *GitStatus) {
	if strings.HasPrefix(h, "# branch.oid") {
		hash := h[13:]
		if hash != "(initial)" {
			s.Sha = hash
		}
		return
	}
	if strings.HasPrefix(h, "# branch.head") {
		branch := h[14:]
		if branch != "(detached)" {
			s.Branch = branch
		}
		return
	}
	if strings.HasPrefix(h, "# branch.upstream") {
		s.Upstream = h[18:]
	}
	if strings.HasPrefix(h, "# branch.ab") {
		parts := strings.Split(h, " ")
		s.Ahead, _ = strconv.Atoi(strings.TrimPrefix(parts[2], "+"))
		s.Behind, _ = strconv.Atoi(strings.TrimPrefix(parts[3], "-"))
		return
	}
}

func runGitCommand(cmd string, args ...string) (string, error) {

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	command := exec.Command(cmd, args...)
	command.Stdout = bufio.NewWriter(&stdout)
	command.Stderr = bufio.NewWriter(&stderr)
	command.Env = os.Environ()
	command.Env = append(command.Env, "LC_ALL=C")

	if err := command.Run(); err != nil {
		if stderr.Len() > 0 {
			return "", errors.New(stderr.String())
		}
		return "", err
	}

	return strings.TrimSpace(stdout.String()), nil

}
