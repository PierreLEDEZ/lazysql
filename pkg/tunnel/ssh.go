package tunnel

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/lazysql/lazysql/pkg/config"
	"golang.org/x/crypto/ssh"
)

type Tunnel struct {
	Config     config.SSHConfig
	RemoteHost string
	RemotePort int
	LocalPort  int

	client   *ssh.Client
	listener net.Listener
	done     chan struct{}
}

func New(cfg config.SSHConfig, remoteHost string, remotePort int) *Tunnel {
	if cfg.Port == 0 {
		cfg.Port = 22
	}
	return &Tunnel{
		Config:     cfg,
		RemoteHost: remoteHost,
		RemotePort: remotePort,
		done:       make(chan struct{}),
	}
}

func (t *Tunnel) Start() (int, error) {
	authMethods := []ssh.AuthMethod{}

	if t.Config.KeyPath != "" {
		if method, err := loadPrivateKey(t.Config.KeyPath); err == nil {
			authMethods = append(authMethods, method)
		}
	}
	if t.Config.Password != "" {
		authMethods = append(authMethods, ssh.Password(t.Config.Password))
	}

	if len(authMethods) == 0 {
		return 0, fmt.Errorf("no SSH authentication method available")
	}

	sshConfig := &ssh.ClientConfig{
		User:            t.Config.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%d", t.Config.Host, t.Config.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return 0, fmt.Errorf("SSH dial failed: %w", err)
	}
	t.client = client

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		client.Close()
		return 0, fmt.Errorf("local listener failed: %w", err)
	}
	t.listener = listener
	t.LocalPort = listener.Addr().(*net.TCPAddr).Port

	go t.acceptLoop()

	return t.LocalPort, nil
}

func (t *Tunnel) Stop() {
	select {
	case <-t.done:
		return
	default:
		close(t.done)
	}
	if t.listener != nil {
		t.listener.Close()
	}
	if t.client != nil {
		t.client.Close()
	}
}

func (t *Tunnel) acceptLoop() {
	remoteAddr := fmt.Sprintf("%s:%d", t.RemoteHost, t.RemotePort)
	for {
		select {
		case <-t.done:
			return
		default:
		}

		localConn, err := t.listener.Accept()
		if err != nil {
			select {
			case <-t.done:
				return
			default:
				continue
			}
		}

		remoteConn, err := t.client.Dial("tcp", remoteAddr)
		if err != nil {
			localConn.Close()
			continue
		}

		go forward(localConn, remoteConn)
	}
}

func forward(local, remote net.Conn) {
	defer local.Close()
	defer remote.Close()

	done := make(chan struct{}, 2)
	go func() {
		io.Copy(local, remote)
		done <- struct{}{}
	}()
	go func() {
		io.Copy(remote, local)
		done <- struct{}{}
	}()
	<-done
	<-done
}

func loadPrivateKey(path string) (ssh.AuthMethod, error) {
	key, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(signer), nil
}
