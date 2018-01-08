package main

import (
	"fmt"
	"github.com/baramiya/easyssh"
)

func main() {
	ssh := easyssh.New()
	ssh.Host = "localhost"
	ssh.User = "john"

	// Call ExecCommand method to run a command on the remote server
	output, err := ssh.ExecCommand("ps ax", 1000)

	// Handle errors
	if err != nil {
		panic(fmt.Sprintf("Can't run remote command[%s]: %s: %s", output.Command, output.Stderr, err.Error()))
	} else {
		fmt.Println("stdout is :", output.Stdout, ";   stderr is :", output.Stderr)
	}
}
