package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"

	sshagent "golang.org/x/crypto/ssh/agent"
)

func listenAndServe(ln net.Listener) {
	agent := sshagent.NewKeyring()
	for {
		log.Print("waiting for connection ")
		conn, err := ln.Accept()
		log.Printf("connected to %v", conn.LocalAddr())
		if err != nil {
			log.Printf("failed to accept connection, err: %v", err)
		}
		sshagent.ServeAgent(agent, conn)
	}
}

func createSSHAgent(sock string, wg *sync.WaitGroup) error {
	log.Printf("creating sock %v", sock)
	// remove stale socket
	os.Remove(sock)

	if err := os.MkdirAll(path.Dir(sock), 0700); err != nil {
		return fmt.Errorf("failed to create the parent directories for socket %s: %v", sock, err)
	}
	ln, err := net.Listen("unix", sock)
	if err != nil {
		return err
	}
	wg.Add(1)
	// create os.Signal channel and listen for any Interrupts
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func(ln net.Listener, wg *sync.WaitGroup, c chan os.Signal) {
		<-c
		log.Printf("closing socket %v\n", ln.Addr())
		ln.Close()
		wg.Done()
		os.Exit(0)
	}(ln, wg, sigc)

	go listenAndServe(ln)
	return nil
}

func main() {
	// $go run server.go <username>
	args := os.Args
	var wg sync.WaitGroup
	tmpDir := os.TempDir()
	// socket will be created under /$TEMPDIR/.ssh_socks/<username>.sock
	sock := fmt.Sprintf("%s/.ssh_socks/%s.sock", tmpDir, args[1])
	createSSHAgent(sock, &wg)
	wg.Wait()
}
