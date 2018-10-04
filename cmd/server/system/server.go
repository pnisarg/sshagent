// uses system's ssh-agent (/usr/bin/ssh-agent)
package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
)

// getAgentProcessFromLog returns pointer to agent process created by executing
// running exec command.
func getAgentProcessFromLog(reader io.Reader) (*os.Process, error) {
	r, err := regexp.Compile(`SSH_AGENT_PID=(\d+);`)
	if err != nil {
		return nil, err
	}
	br := bufio.NewScanner(reader)
	for br.Scan() {
		match := r.FindStringSubmatch(br.Text())
		if len(match) == 2 {
			pid, err := strconv.Atoi(match[1])
			if err != nil {
				return nil, err
			}
			return os.FindProcess(pid)
		}
	}
	return nil, errors.New("couldn't find agent's pid in output")
}

func createSSHAgent(ctx context.Context, sock string) (*os.Process, error) {
	agentPath, err := exec.LookPath("ssh-agent")
	if err != nil {
		return nil, fmt.Errorf("couldn't find ssh-agent")
	}
	args := []string{
		"-a", sock,
	}
	// remove stale socket
	os.Remove(sock)

	cmd := exec.CommandContext(ctx, agentPath, args...)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	proc, err := getAgentProcessFromLog(out)
	if err != nil {
		return nil, err
	}
	log.Printf("started agent pid: %v", proc.Pid)
	return proc, nil
}

// uses system provided ssh-agent.
func main() {
	args := os.Args
	tmpDir := os.TempDir()
	// socket will be created under /$TEMPDIR/.ssh_socks/<username>.sock
	sock := fmt.Sprintf("%s/.ssh_socks/%s.sock", tmpDir, args[1])
	createSSHAgent(context.Background(), sock)
}
