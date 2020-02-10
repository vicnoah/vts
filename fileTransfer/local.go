package file_transfer

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"time"

	"github.com/vicnoah/progress"
)

// NewLocal 本地文件系统操作实例
func NewLocal() *Local {
	return &Local{}
}

// Local 本地操作结构
type Local struct {
}

// Send 发送文件
func (l *Local) Send(ctx context.Context, srcFile ReadStater, dst io.Writer) (err error) {
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
				_, err = io.CopyBuffer(dst, src, make([]byte, 10*1024*1024))
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

// Recv 接收文件
func (l *Local) Recv(ctx context.Context, srcFile ReadStater, dst io.Writer) (err error) {
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
				_, err = io.CopyBuffer(dst, src, make([]byte, 10*1024*1024))
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

// ReadDir 读取目录
func (l *Local) ReadDir(dirname string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(dirname)
}

// Open 打开文件
func (l *Local) Open(name string) (ReadWriteStatCloser, error) {
	return os.Open(name)
}

// OpenFile 打开文件
func (l *Local) OpenFile(fileName string, f int) (ReadWriteStatCloser, error) {
	return os.OpenFile(fileName, f, os.ModePerm)
}

// Remove 删除文件
func (l *Local) Remove(name string) error {
	return os.Remove(name)
}

// Stat 获取文件信息
func (l *Local) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

// Rename 重命名文件
func (l *Local) Rename(oldName, newName string) error {
	return os.Rename(oldName, newName)
}

// WriteFile 写入文件
func (l *Local) WriteFile(fileName string, flag int, b []byte) (err error) {
	f, err := os.OpenFile(fileName, flag, os.ModePerm)
	if err != nil {
		return
	}
	defer f.Close()
	_, err = f.Write(b)
	return
}
