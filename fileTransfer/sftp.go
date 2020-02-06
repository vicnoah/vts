package file_transfer

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"time"

	"github.com/pkg/sftp"
	"github.com/vicnoah/progress"
	"golang.org/x/crypto/ssh"
)

// NewSFTP 实例
func NewSFTP() *SFTP {
	return &SFTP{}
}

// SFTP 结构
type SFTP struct {
	client *sftp.Client
}

// Connect 连接sftp
func (s *SFTP) Connect(addr string, sshConfig *ssh.ClientConfig) (err error) {
	sshClient, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return
	}
	// create sftp client
	s.client, err = sftp.NewClient(sshClient)
	return
}

// Close 关闭连接
func (s *SFTP) Close() error {
	return s.client.Close()
}

// Client 客户端
func (s *SFTP) Client() *sftp.Client {
	return s.client
}

// Upload 上传文件
func (s *SFTP) Upload(ctx context.Context, srcFile *os.File, dst io.Writer) (err error) {
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
				_, err = io.Copy(dst, src)
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

// Download 下载文件
func (s *SFTP) Download(ctx context.Context, srcFile *sftp.File, dst io.Writer) (err error) {
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
			fmt.Printf("\rdownloading per: %d, remaining: %v", int64(math.Floor(p.Percent())), p.Remaining().Round(time.Second))
		}
		fmt.Printf("\n")
	}()

	working := false
	done := false
	for {
		select {
		case <-ctx.Done():
			err = fmt.Errorf("close download manually")
			return
		default:
			if !working {
				working = true
				go func() {
					_, err = io.Copy(dst, src)
					done = true
					return
				}()
			}
			if done {
				return
			}
			time.Sleep(time.Second * 2)
		}
	}
}

// WriteFile 写文件
func (s *SFTP) WriteFile(fileName string, f int, b []byte) (err error) {
	file, err := s.client.OpenFile(fileName, f)
	if err != nil {
		return
	}
	defer file.Close()
	_, err = file.Write(b)
	return
}

// Open 打开文件
func (s *SFTP) Open(path string) (*sftp.File, error) {
	return s.client.Open(path)
}

// OpenFile 打开文件
func (s *SFTP) OpenFile(path string, f int) (*sftp.File, error) {
	return s.client.OpenFile(path, f)
}

// Remove 删除文件
func (s *SFTP) Remove(path string) error {
	return s.client.Remove(path)
}

// RemoveDirectory 删除目录
func (s *SFTP) RemoveDirectory(path string) error {
	return s.client.RemoveDirectory(path)
}

// Rename 重命名文件
func (s *SFTP) Rename(oldName string, newName string) error {
	return s.client.Rename(oldName, newName)
}

// Stat 读取文件状态
func (s *SFTP) Stat(p string) (os.FileInfo, error) {
	return s.client.Stat(p)
}

// LStat 读取文件列表状态
func (s *SFTP) LStat(p string) (os.FileInfo, error) {
	return s.client.Lstat(p)
}

// Chmod 改变权限
func (s *SFTP) Chmod(path string, mode os.FileMode) error {
	return s.client.Chmod(path, mode)
}

// Read 读取文件
func (s *SFTP) Read(p []byte) (n int, err error) {
	return s.client.Read(p)
}

// ReadDir 读取目录
func (s *SFTP) ReadDir(p string) ([]os.FileInfo, error) {
	return s.client.ReadDir(p)
}

// Mkdir 创建目录
func (s *SFTP) Mkdir(path string) error {
	return s.client.Mkdir(path)
}

// MkdirAll 递归创建目录
func (s *SFTP) MkdirAll(path string) error {
	return s.client.MkdirAll(path)
}
