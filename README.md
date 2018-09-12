# sshagent
Using Golang ssh-agent with unix socket

Run server
$go run cmd/server.go <username>

Run client 
$go run cmd/client.go <username>

Client will try to dial into unix socket created by server and write private key to it. Server reads the private key and adds it to the agent keyring. 
