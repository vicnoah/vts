package file_transfer

import (
	"context"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/vicnoah/progress"

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

// Close 断开连接
func (s *SCP) Close() {
	s.client.Close()
}

// Upload 上传文件
func (s *SCP) Upload(ctx context.Context, srcFile *os.File, dst string) (err error) {
	fileInfo, err := srcFile.Stat()
	if err != nil {
		return
	}

	size := fileInfo.Size()
	src := progress.NewReader(srcFile)
	// Start a goroutine printing progress
	go func() {
		ctx := context.Background()
		progressChan := progress.NewTicker(ctx, src, size, 1*time.Second)
		for p := range progressChan {
			fmt.Printf("\ruploading per: %d, remaining: %v", int64(math.Floor(p.Percent())), p.Remaining().Round(time.Second))
		}
		fmt.Printf("\n")
	}()

	working := false
	done := false
	for {
		select {
		case <-ctx.Done():
			err = fmt.Errorf("close upload manually")
			return
		default:
			if !working {
				err = s.client.Copy(src, dst, "0655", size)
				done = true
				return
			}
			if done {
				return
			}
			time.Sleep(time.Second * 2)
		}
	}
}
