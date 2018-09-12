package main

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
	sshagent "golang.org/x/crypto/ssh/agent"
)

func getSSHAgentForPrincipal(principal string) (sshagent.Agent, net.Conn, error) {
	tmpDir := os.TempDir()
	sock := fmt.Sprintf("%s/.ssh_socks/%s.sock", tmpDir, principal)
	if _, err := os.Stat(sock); os.IsNotExist(err) {
		return nil, nil, err
	}
	conn, err := net.Dial("unix", sock)
	if err != nil {
		return nil, nil, fmt.Errorf("GetSSHAgentForPrincipal: failed to get sock for principal %s, err: %v", principal, err)
	}
	agent := sshagent.NewClient(conn)
	return agent, conn, nil
}

func createPublicKey() (*rsa.PrivateKey, ssh.PublicKey, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	pub, err := ssh.NewPublicKey(priv.Public())
	if err != nil {
		return nil, nil, err
	}
	return priv, pub, nil
}

func main() {
	args := os.Args
	agent, _, err := getSSHAgentForPrincipal(args[1])
	if err != nil {
		panic(err)
	}
	// create dummy public-private key pair and load private key to the agent
	priv, _, err := createPublicKey()
	if err != nil {
		panic(err)
	}
	log.Println("created keys, loading to agent")
	if err := agent.Add(sshagent.AddedKey{PrivateKey: priv, Comment: "private key"}); err != nil {
		panic(err)
	}
	log.Println(agent.List())
}
