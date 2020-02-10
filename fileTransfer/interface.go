package file_transfer

import (
	"context"
	"io"
	"os"
)

// Fser 文件操作接口
type Fser interface {
	Send(context.Context, ReadStater, io.Writer, int) error
	Recv(context.Context, ReadStater, io.Writer, int) error
	ReadDir(string) ([]os.FileInfo, error)
	Open(string) (ReadWriteStatCloser, error)
	OpenFile(string, int) (ReadWriteStatCloser, error)
	Remove(string) error
	Stat(string) (os.FileInfo, error)
	Rename(string, string) error
	WriteFile(string, int, []byte) error
}

// ReadWriteStatCloser 读取,写入,读取状态,关闭文件
type ReadWriteStatCloser interface {
	io.ReadWriteCloser
	Stater
}

// ReadStater 写入并读取文件状态
type ReadStater interface {
	io.Reader
	Stater
}

// Stater 读取文件状态
type Stater interface {
	Stat() (os.FileInfo, error)
}
