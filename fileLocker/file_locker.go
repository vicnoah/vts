package file_locker

import (
	"os"
	"path"
	"sync"

	sp "github.com/vicnoah/vts/sftp"

	"github.com/pkg/sftp"
)

const (
	lock = ".lock"
)

// New 获取文件锁实例
func New() *FileLock {
	return &FileLock{}
}

// FileLock 文件锁结构
type FileLock struct {
	mu sync.Mutex
}

// Lock 创建文件锁
func (l *FileLock) Lock(fileName string, client *sftp.Client) (err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	err = sp.WriteFile(lockFileName(fileName), client, os.O_WRONLY|os.O_CREATE, []byte("locker"))
	return
}

// Unlock 删除文件锁
func (l *FileLock) Unlock(fileName string, client *sftp.Client) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return sp.Remove(lockFileName(fileName), client)
}

// IsLock 判断文件锁是否存在
func (l *FileLock) IsLock(fileName string, client *sftp.Client) (isLock bool, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	_, err = sp.Stat(lockFileName(fileName), client)
	if err != nil {
		if os.IsNotExist(err) {
			isLock = false
			err = nil
			return
		}
		return
	}
	isLock = true
	return
}

func lockFileName(fileName string) string {
	return path.Join(path.Dir(fileName),
		path.Base(fileName)+lock)
}
