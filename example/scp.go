package main

import (
	"fmt"
	"github.com/baramiya/easyssh"
)

func main() {
	ssh := easyssh.New()
	ssh.Host = "localhost"
	ssh.User = "john"

	// Call Scp method with file you want to upload to remote server
	err := ssh.Scp("/tmp/source.csv", "/tmp/target.csv")

	// Handle errors
	if err != nil {
		panic("Can't run remote command: " + err.Error())
	} else {
		fmt.Println("success")

	}
}
