// Package easyssh provides a simple implementation of some SSH protocol features in Go.
// You can simply run command on remote server or get a file even simple than native console SSH client.
// Do not need to think about Dials, sessions, defers and public keys...Let easyssh will be think about it!
package easyssh

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

// Contains main authority information.
// User field should be a name of user on remote server (ex. john in ssh john@example.com).
// Server field should be a remote machine address (ex. example.com in ssh john@example.com)
// Key is a path to private key on your local machine.
// Port is SSH server port on remote machine.
// Note: easyssh looking for private key in user's home directory (ex. /home/john + Key).
// Then ensure your Key begins from '/' (ex. /.ssh/id_rsa)
type MakeConfig struct {
	User   string
	Server string
	Key    string
	Port   string

	Client *ssh.Client
}

// Contains command run result.
type Response struct {
	Stdout string
	Stderr string
	Error  error
}

// returns ssh.Signer from user you running app home path + cutted key path.
// (ex. pubkey,err := getKeyFile("/.ssh/id_rsa") )
func getKeyFile(keypath string) (ssh.Signer, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	file := usr.HomeDir + keypath
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	pubkey, err := ssh.ParsePrivateKey(buf)
	if err != nil {
		return nil, err
	}

	return pubkey, nil
}

// disconnects from remote server
func (ssh_conf *MakeConfig) close() {
	if ssh_conf.Client != nil {
		ssh_conf.Client.Close()
		ssh_conf.Client = nil
	}
}

// connects to remote server using MakeConfig struct and returns *ssh.Session
func (ssh_conf *MakeConfig) connect() (*ssh.Session, error) {
	if ssh_conf.Client == nil {
		pubkey, err := getKeyFile(ssh_conf.Key)
		if err != nil {
			return nil, err
		}

		config := &ssh.ClientConfig{
			User: ssh_conf.User,
			Auth: []ssh.AuthMethod{ssh.PublicKeys(pubkey)},
		}

		client, err := ssh.Dial("tcp", ssh_conf.Server+":"+ssh_conf.Port, config)
		if err != nil {
			return nil, err
		}

		ssh_conf.Client = client
	}

	session, err := ssh_conf.Client.NewSession()
	if err != nil {
		ssh_conf.close()
		return nil, err
	}

	return session, nil
}

// Runs command on remote machine and returns STDOUT
func (ssh_conf *MakeConfig) Run(command string) (response Response) {
	session, err := ssh_conf.connect()

	if err != nil {
		response.Error = err
		return
	}
	defer func() {
		session.Close()
	}()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr
	err = session.Run(command)
	response.Stdout = stdout.String()
	response.Stderr = stderr.String()
	response.Error = err

	return response
}

// Scp uploads sourceFile to remote machine like native scp console app.
func (ssh_conf *MakeConfig) Scp(sourceFile string) error {
	session, err := ssh_conf.connect()

	if err != nil {
		return err
	}
	defer func() {
		session.Close()
	}()

	targetFile := filepath.Base(sourceFile)

	src, srcErr := os.Open(sourceFile)

	if srcErr != nil {
		return srcErr
	}

	srcStat, statErr := src.Stat()

	if statErr != nil {
		return statErr
	}

	go func() {
		w, _ := session.StdinPipe()

		fmt.Fprintln(w, "C0644", srcStat.Size(), targetFile)

		if srcStat.Size() > 0 {
			io.Copy(w, src)
			fmt.Fprint(w, "\x00")
			w.Close()
		} else {
			fmt.Fprint(w, "\x00")
			w.Close()
		}
	}()

	if err := session.Run(fmt.Sprintf("scp -t %s", targetFile)); err != nil {
		return err
	}

	return nil
}
