package ppsh

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

var (
	defaultCiphers = []string{
		"aes128-ctr",
		"aes192-ctr",
		"aes256-ctr",
		"aes128-gcm@openssh.com",
		"arcfour256",
		"arcfour128",
		"aes128-cbc",
		"3des-cbc",
		"aes192-cbc",
		"aes256-cbc"}

	defaultHost     = "localhost"
	defaultUser     = "root"
	defaultTimeout  = 60 * 5
	defaultPort     = 22
	defaultPlatform = "linux"
)

func validatedCiphers(ciphers []string) []string {
	if len(ciphers) == 0 {
		return defaultCiphers
	}
	return ciphers
}

func getSSHAuthMethod(password, certKey string) ([]ssh.AuthMethod, error) {
	// both password and cert key is null
	if password == "" && certKey == "" {
		return nil, fmt.Errorf("password and certKey cannot be null at the some time")
	}

	auth := make([]ssh.AuthMethod, 0)

	// only password
	if certKey == "" {
		auth = append(auth, ssh.Password(password))
		return auth, nil
	}

	// read the cert key data
	pemBytes, err := ioutil.ReadFile(certKey)
	if err != nil {
		return nil, err
	}

	var signer ssh.Signer
	if password != "" {
		// both password and cert key
		signer, err = ssh.ParsePrivateKeyWithPassphrase(pemBytes, []byte(password))
	} else {
		// only cert key
		signer, err = ssh.ParsePrivateKey(pemBytes)
	}
	if err != nil {
		return nil, err
	}

	auth = append(auth, ssh.PublicKeys(signer))

	return auth, nil
}

func connect(user, password, host, certKey string, port, timeout int, ciphers []string) (*ssh.Session, error) {
	// set ssh auth method
	auth, err := getSSHAuthMethod(password, certKey)
	if err != nil {
		return nil, err
	}

	// set ssh config
	clientConfig := &ssh.ClientConfig{
		User:    user,
		Auth:    auth,
		Timeout: time.Duration(timeout) * time.Second,
		Config: ssh.Config{
			Ciphers: validatedCiphers(ciphers),
		},
		HostKeyCallback: func(host string, remote net.Addr, certKey ssh.PublicKey) error {
			return nil
		},
	}

	// dial to ssh server and get the client
	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return nil, err
	}

	// create ssh session
	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}

	// simulate terminal login
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	if err := session.RequestPty("xterm", 100, 180, modes); err != nil {
		return nil, err
	}

	return session, nil
}

func doOther(user, password, host, certKey string, port, timeout int, ciphers, cmds []string, srch chan Result) {
	r := Result{Host: host}

	session, err := connect(user, password, host, certKey, port, timeout, ciphers)
	if err != nil {
		r.Error = err.Error()
		srch <- r
		return
	}
	defer session.Close()

	var outBuf, errBuf bytes.Buffer
	session.Stdout = &outBuf
	session.Stderr = &errBuf

	if err = session.Shell(); err != nil {
		r.Error = err.Error()
		srch <- r
		return
	}

	cmds = append(cmds, "exit") // remember to exit
	stdinBuf, _ := session.StdinPipe()
	for _, c := range cmds {
		c = c + "\n"
		r.Cmd += c
		stdinBuf.Write([]byte(c))
	}

	session.Wait()
	if errBuf.String() != "" {
		r.Error = errBuf.String()
	} else {
		r.Success = true
		r.Detail = outBuf.String()
	}
	srch <- r

	return
}

func doLinux(user, password, host, certKey string, port, timeout int, ciphers, cmds []string, srch chan Result) {
	cmds = append(cmds, "exit")
	cmd := strings.Join(cmds, " && ")

	r := Result{Host: host, Cmd: cmd}

	session, err := connect(user, password, host, certKey, port, timeout, ciphers)
	if err != nil {
		r.Error = err.Error()
		r.Code = -1
		srch <- r
		return
	}
	defer session.Close()

	var outBuf, errBuf bytes.Buffer

	{
		session.Stdout = &outBuf
		session.Stderr = &errBuf
	}

	if err := session.Run(cmd); err != nil {
		r.Error = err.Error()
		srch <- r
		return
	}

	if errBuf.String() != "" {
		r.Error = errBuf.String()
	} else {
		r.Success = true
		r.Detail = outBuf.String()
	}
	srch <- r

	return
}

func Do(user, password, host, key string, port int, timeout int, ciphers, cmds []string, flatform string, ch chan Result) {
	if host == "" {
		host = defaultHost
	}
	if user == "" {
		user = defaultUser
	}
	if timeout == 0 {
		timeout = defaultTimeout
	}
	if port == 0 {
		port = defaultPort
	}
	if flatform == "" {
		flatform = defaultPlatform
	}

	srch := make(chan Result)
	if flatform == "linux" {
		go doLinux(user, password, host, key, port, timeout, ciphers, cmds, srch)
	} else {
		go doOther(user, password, host, key, port, timeout, ciphers, cmds, srch)
	}

	r := Result{Host: host}
	select {
	case <-time.After(time.Duration(timeout) * time.Second):
		ch <- Result{
			Host:  host,
			Error: fmt.Sprintf("ssh run timeout in %d second", timeout),
		}
	case r = <-srch:
		ch <- r
	}
	return
}
