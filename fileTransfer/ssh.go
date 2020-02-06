package file_transfer

import (
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
)

// NewSSH 新建ssh实例
func NewSSH() *SSH {
	return &SSH{}
}

// SSH 结构
type SSH struct {
	addr   string
	client *ssh.Client
	config *ssh.ClientConfig
}

// SetConfig 配置ssh
func (s *SSH) SetConfig(user, password, host string, port int) (err error) {
	var (
		auth []ssh.AuthMethod
	)
	// get auth method
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(password))

	s.config = &ssh.ClientConfig{
		User:            user,
		Auth:            auth,
		Timeout:         30 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //ssh.FixedHostKey(hostKey),
	}
	s.addr = fmt.Sprintf("%s:%d", host, port)
	return
}

// SetConfigWithKey 配置key认证服务器
func (s *SSH) SetConfigWithKey(user string, key []byte, host string, port int) (err error) {
	var (
		auth []ssh.AuthMethod
	)
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return
	}
	// get auth method
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.PublicKeys(signer))

	s.config = &ssh.ClientConfig{
		User:            user,
		Auth:            auth,
		Timeout:         30 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //ssh.FixedHostKey(hostKey),
	}

	s.addr = fmt.Sprintf("%s:%d", host, port)
	return
}

// Connect 连接服务器
func (s *SSH) Connect() (err error) {
	// connet to ssh
	s.client, err = ssh.Dial("tcp", s.addr, s.config)
	return
}

// Config 获取配置
func (s *SSH) Config() *ssh.ClientConfig {
	return s.config
}

// Addr 获取地址
func (s *SSH) Addr() string {
	return s.addr
}

// Client 返回客户端
func (s *SSH) Client() *ssh.Client {
	return s.client
}

// Close 断开连接
func (s *SSH) Close() error {
	return s.Close()
}
