/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 * Unless required by applicable law or agreed to in writing, software        *
 * distributed under the License is distributed on an "AS IS" BASIS,          *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.   *
 * See the License for the specific language governing permissions and        *
 * limitations under the License.                                             *
 *                                                                            *
 ******************************************************************************/

package sshclient

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/library/framework"
	"golang.org/x/crypto/ssh"
)

type SSHClientExecutor interface {
	RunCommandsInRemoteHost(host string, port int, sshType SSHType, user, passwd string, sudo bool, timeoutS int, commands []string) (result string, err error)
}

type SSHExecutor struct{}

func (client SSHExecutor) RunCommandsInRemoteHost(host string, port int, sshType SSHType, user, passwd string, sudo bool, timeoutS int, commands []string) (result string, err error) {
	c := new(SSHClient)
	c.InitSSHClient(host, port, sshType, user, passwd, timeoutS)
	if err = c.Connect(); err != nil {
		return "", err
	}
	defer c.Close()

	return c.RunCommandsInSession(sudo, commands)
}

type SSHType string

const (
	Passwd SSHType = "PassWord" // Auth by Password
	Key    SSHType = "Key"      // Auth by ssh key
)

type SSHClient struct {
	sshHost     string
	sshPort     int
	sshType     SSHType
	sshUser     string
	sshPassword string
	sshTimeout  time.Duration
	sshKeyPath  string //path of id_rsa

	client *ssh.Client
}

func (c *SSHClient) InitSSHClient(host string, port int, sshType SSHType, user, passwd string, timeoutS int) {
	c.sshHost = host
	c.sshPort = port
	c.sshType = sshType
	c.sshUser = user
	c.sshPassword = passwd

	c.sshTimeout = time.Duration(timeoutS) * time.Second
	c.sshKeyPath = filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa")
}

func (c *SSHClient) SetConnTimeOut(t time.Duration) {
	c.sshTimeout = t
}

func (c *SSHClient) SetKeyPath(path string) {
	c.sshKeyPath = path
}

func (c *SSHClient) Connect() (err error) {
	config := &ssh.ClientConfig{
		Timeout: c.sshTimeout,
		User:    c.sshUser,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			// Not Check SSH Server Key
			return nil
		},
	}
	if c.sshType == Passwd {
		config.Auth = []ssh.AuthMethod{ssh.Password(c.sshPassword)}
	} else {
		config.Auth = []ssh.AuthMethod{c.publicKeyAuthFunc()}
	}

	addr := fmt.Sprintf("%s:%d", c.sshHost, c.sshPort)
	c.client, err = ssh.Dial("tcp", addr, config)
	if err != nil {
		err = errors.NewErrorf(errors.TIEM_RESOURCE_CONNECT_TO_HOST_ERROR, "ssh client dial to addr %s@%s failed, %v", c.sshUser, addr, err)
		return
	}

	return nil
}

func (c *SSHClient) Close() {
	if c.client != nil {
		c.client.Close()
	}
}

func (c *SSHClient) RunCommandsInSession(sudo bool, commands []string) (result string, err error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", errors.NewErrorf(errors.TIEM_RESOURCE_NEW_SESSION_ERROR, "new ssh session failed for %s@%s:%d, %v", c.sshUser, c.sshHost, c.sshPort, err)
	}
	defer session.Close()

	command := strings.Join(commands, ";")
	if sudo {
		command = fmt.Sprintf("/usr/bin/sudo -H bash -c \"%s\"", command)
	}
	combo, err := session.CombinedOutput(command)
	if err != nil {
		return "", errors.NewErrorf(errors.TIEM_RESOURCE_RUN_COMMAND_ERROR, "exec command %s on %s@%s:%d failed, %v", command, c.sshUser, c.sshHost, c.sshPort, err)
	}
	result = string(combo)
	return
}

func (c *SSHClient) publicKeyAuthFunc() ssh.AuthMethod {
	log := framework.Log()
	key, err := ioutil.ReadFile(c.sshKeyPath)
	if err != nil {
		log.Fatalf("read ssh key file %s failed, %v", c.sshKeyPath, err)
		return nil
	}
	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("ssh key signer failed, %v", err)
		return nil
	}
	return ssh.PublicKeys(signer)
}
