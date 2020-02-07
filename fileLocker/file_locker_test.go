package file_locker

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLock(t *testing.T) {
	locker := New()
	lk := &localIO{}
	tests := []struct {
		fileName string // input
	}{
		{"???"},
		{"filename.go"},
		{"1.mp4"},
		{"1.mkv"},
		{"1.webm"},
		{"1"},
		{"xx.webm"},
		{"1.x.b.3.webm"},
	}

	for _, test := range tests {
		// Only pass t into top-level Convey calls
		Convey("Get locker file name", t, func() {
			test := test

			Convey("locker", func() {
				err := locker.Lock(test.fileName, lk)
				if err != nil {
					t.Error(err)
				}

				Convey("check locker", func() {
					isLock, err := locker.IsLock(test.fileName, lk)
					if err != nil {
						t.Error(err)
					}
					So(isLock, ShouldEqual, true)
					err = locker.Unlock(test.fileName, lk)
					if err != nil {
						t.Error(err)
					}
					isLock, err = locker.IsLock(test.fileName, lk)
					if err != nil {
						t.Error(err)
					}
					So(isLock, ShouldEqual, false)
				})
			})
		})
	}
}

// localIO 本地io接口
type localIO struct {
}

// WriteFile 写入文件
func (l *localIO) WriteFile(name string, f int, data []byte) (err error) {
	fi, err := os.OpenFile(name, f, os.ModePerm)
	if err != nil {
		return
	}
	defer fi.Close()
	_, err = fi.Write(data)
	return
}

// Remove 删除文件
func (l *localIO) Remove(path string) error {
	return os.Remove(path)
}

// Stat 获取文件状态
func (l *localIO) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}
