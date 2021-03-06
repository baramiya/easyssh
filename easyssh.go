package easyssh

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"github.com/mitchellh/go-homedir"
	"bytes"
	"time"
)

type (
	EasySSH struct {
		Host     string
		Port     string
		User     string
		Password string
		KeyPath  string

		KeySigner  ssh.Signer
		SSHClient  *ssh.Client
		SFTPClient *sftp.Client
	}
	sshOutput struct {
		Command string
		Error   error
		Stdout  string
		Stderr  string
	}
)

func New() *EasySSH {
	return &EasySSH{
		Port:    "22",
		KeyPath: "~/.ssh/id_rsa",
	}
}

func (easySSH *EasySSH) setKeySigner() error {
	privateKeyPath, err := homedir.Expand(easySSH.KeyPath)
	if err != nil {
		return err
	}
	privateKeyBytes, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return err
	}
	privateKey, err := ssh.ParsePrivateKey(privateKeyBytes)
	if err != nil {
		return err
	}

	easySSH.KeySigner = privateKey
	return nil
}

func (easySSH *EasySSH) newClient() (*ssh.Client, error) {
	if easySSH.SSHClient == nil {
		authMethods := []ssh.AuthMethod{}
		if easySSH.Password != "" {
			authMethods = append(authMethods, ssh.Password(easySSH.Password))
		}
		if easySSH.KeyPath != "" {
			if err := easySSH.setKeySigner(); err == nil {
				authMethods = append(authMethods, ssh.PublicKeys(easySSH.KeySigner))
			}
		}

		clientConfig := &ssh.ClientConfig{
			User:            easySSH.User,
			Auth:            authMethods,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         500 * time.Millisecond,
		}

		client, err := ssh.Dial("tcp", easySSH.Host+":"+easySSH.Port, clientConfig)
		if err != nil {
			return nil, err
		}

		easySSH.SSHClient = client
	}

	return easySSH.SSHClient, nil
}

func (easySSH *EasySSH) ExecCommand(command string) (*sshOutput, error) {
	output := &sshOutput{
		Command: command,
	}

	client, err := easySSH.newClient()
	if err != nil {
		output.Stderr = "Could not establish ssh connection"
		return output, err
	}

	session, err := client.NewSession()
	if err != nil {
		output.Stderr = "Could not establish ssh session"
		easySSH.SSHClient.Close()
		easySSH.SSHClient = nil
		return output, err
	}

	var stdout, stderr bytes.Buffer
	session.Stdout, session.Stderr = &stdout, &stderr
	err = session.Run(command)
	session.Close()
	output.Stdout = stdout.String()
	output.Stderr = stderr.String()

	return output, err
}

func (easySSH *EasySSH) Scp(sourceFilePath, targetFilePath string) error {
	client, err := easySSH.newClient()
	if err != nil {
		return err
	}

	session, err := client.NewSession()
	if err != nil {
		easySSH.SSHClient.Close()
		easySSH.SSHClient = nil
		return err
	}
	defer session.Close()

	sourceFile, err := os.Open(sourceFilePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	sourceFileStat, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	targetFile := filepath.Base(targetFilePath)
	go func() {
		w, _ := session.StdinPipe()

		fmt.Fprintln(w, "C0644", sourceFileStat.Size(), targetFile)

		if sourceFileStat.Size() > 0 {
			io.Copy(w, sourceFile)
			fmt.Fprint(w, "\x00")
			w.Close()
		} else {
			fmt.Fprint(w, "\x00")
			w.Close()
		}
	}()

	if err := session.Run(fmt.Sprintf("scp -tr %s", targetFilePath)); err != nil {
		return err
	}

	return nil
}

func (easySSH *EasySSH) Sftp(sourceFilePath, targetFilePath string) error {
	client, err := easySSH.newClient()
	if err != nil {
		return err
	}

	session, err := client.NewSession()
	if err != nil {
		easySSH.SSHClient.Close()
		easySSH.SSHClient = nil
		return err
	}
	defer session.Close()

	easySSH.SFTPClient, err = sftp.NewClient(easySSH.SSHClient)
	if err != nil {
		return err
	}
	defer func() {
		easySSH.SFTPClient.Close()
		easySSH.SFTPClient = nil
	}()

	sourceFile, err := os.Open(sourceFilePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	targetFile, err := easySSH.SFTPClient.Create(targetFilePath)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	if _, err := io.Copy(targetFile, sourceFile); err != nil {
		return err
	}

	return nil
}
