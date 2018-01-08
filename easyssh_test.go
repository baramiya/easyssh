package easyssh

import (
	"testing"
)

func TestEasySSH_New(t *testing.T) {
	ssh := New()
	if ssh == nil || ssh.Port != "22" || ssh.KeyPath != "~/.ssh/id_rsa" {
		t.FailNow()
	}
}

func TestEasySSH_setKeySigner(t *testing.T) {
	ssh := New()
	if err := ssh.setKeySigner(); err != nil {
		t.FailNow()
	}

	ssh.KeyPath = "testFiles/id_rsa"
	if err := ssh.setKeySigner(); err == nil {
		t.FailNow()
	}

	ssh.KeyPath = "testFiles/id_rsa.pub"
	if err := ssh.setKeySigner(); err == nil {
		t.FailNow()
	}
}

func TestEasySSH_newClient(t *testing.T) {
	ssh := New()
	client, err := ssh.newClient()
	if client != nil || err == nil {
		t.FailNow()
	}

	ssh.Host = "sdf.org"
	ssh.Port = "22"
	ssh.User = "new"

	client, err = ssh.newClient()
	if client == nil || err != nil {
		t.FailNow()
	}

	clientNew, err := ssh.newClient()
	if client != clientNew {
		t.FailNow()
	}
}

func TestEasySSH_ExecCommand(t *testing.T) {
	ssh := New()
	ssh.Host = "sdf.org"
	ssh.User = "new"

	output, _ := ssh.ExecCommand("echo")
	if output.Stderr == "Timeout exceeded while running command on the remote host" {
		t.FailNow()
	}
}