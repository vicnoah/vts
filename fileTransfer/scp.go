package file_transfer

import (
	scp "github.com/bramvdbogaerde/go-scp"
	"golang.org/x/crypto/ssh"
)

// NewSCP scp实例
func NewSCP() *SCP {
	return &SCP{}
}

// SCP scp结构
type SCP struct {
	client scp.Client
}

// Connect 连接scp服务器
func (s *SCP) Connect(addr string, config *ssh.ClientConfig) {
	s.client = scp.NewClient(addr, config)
}
