package file_locker

import (
	"os"
	"path"
	"sync"
)

const (
	lock = ".lock"
)

// New 获取文件锁实例
func New() *FileLock {
	return &FileLock{}
}

// Locker 锁接口
type Locker interface {
	WriteFile(string, int, []byte) error
	Remove(string) error
	Stat(string) (os.FileInfo, error)
}

// FileLock 文件锁结构
type FileLock struct {
	mu     sync.Mutex
	locker Locker
}

// Lock 创建文件锁
func (l *FileLock) Lock(fileName string, lk Locker) (err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	err = l.locker.WriteFile(lockFileName(fileName), os.O_WRONLY|os.O_CREATE, []byte("locker"))
	return
}

// Unlock 删除文件锁
func (l *FileLock) Unlock(fileName string, lk Locker) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return lk.Remove(lockFileName(fileName))
}

// IsLock 判断文件锁是否存在
func (l *FileLock) IsLock(fileName string, lk Locker) (isLock bool, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	_, err = lk.Stat(lockFileName(fileName))
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
