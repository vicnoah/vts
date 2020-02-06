// Package worker 工作
package worker

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	file_locker "github.com/vicnoah/vts/fileLocker"
	file_transfer "github.com/vicnoah/vts/fileTransfer"
	"github.com/vicnoah/vts/transcode"
)

// Run 开始转码作业
func Run(ctx context.Context, u string, pass string, addr string, port int, w string, r string, cmd string, ext string, formats string, filters string) {
	var (
		ssh = file_transfer.NewSSH()
		sp  = file_transfer.NewSFTP()
		err error
	)
	err = ssh.SetConfig(u,
		pass,
		addr,
		port)
	if err != nil {
		fmt.Printf("ssh config error: %v\n", err)
		return
	}

	err = sp.Connect(ssh.Addr(), ssh.Config())
	if err != nil {
		fmt.Printf("sftp connect error: %v\n", err)
		return
	}
	defer sp.Close()

	files, err := sp.ReadDir(r)
	if err != nil {
		fmt.Printf("reading sftp \"%s\" directory error: %v\n", r, err)
		return
	}
	pts, err := sftpFiles(files, r, formats, filters, sp)
	if err != nil {
		fmt.Printf("reading for sftp remote directory error: %v\n", err)
		return
	}
	fmt.Printf("\n待转码列表:\n")
	for _, f := range pts {
		fmt.Println(f)
	}
	fmt.Printf("\n开始转码作业:\n")
	err = batchTranscode(ctx, pts, w, ext, cmd, sp)
	if err != nil {
		fmt.Printf("\nbatch transcode video error: %v\n", err)
		return
	}
}

func sftpFiles(fs []os.FileInfo, basePath string, formats string, filters string, sp *file_transfer.SFTP) (paths []string, err error) {
	for _, f := range fs {
		if f.IsDir() {
			nPath := path.Join(basePath, f.Name())
			nfs, er := sp.ReadDir(nPath)
			if er != nil {
				err = er
				return
			}
			pts, er := sftpFiles(nfs, nPath, formats, filters, sp)
			if er != nil {
				err = er
				return
			}
			paths = append(paths, pts...)
		} else {
			isExist := false
			if formats != "" {
				for _, format := range strings.Split(formats, ",") {
					if strings.Contains(strings.ToLower(path.Ext(f.Name())), strings.Trim(format, " ")) {
						isExist = true
					}
				}
			}
			if isExist {
				isFilter := false
				if filters == "" {
					isFilter = false
				} else {
					for _, filter := range strings.Split(filters, ",") {
						if strings.Contains(strings.ToLower(path.Join(basePath, f.Name())), strings.Trim(filter, " ")) {
							isFilter = true
						}
					}
				}
				if !isFilter {
					fk := file_locker.New()
					isLock, er := fk.IsLock(path.Join(basePath, f.Name()), sp)
					if er != nil {
						err = er
						return
					}
					if !isLock {
						paths = append(paths, path.Join(basePath, f.Name()))
					}
				}
			}
		}
	}
	return
}

func batchTranscode(ctx context.Context, fileNames []string, w string, ext string, cmd string, sp *file_transfer.SFTP) (err error) {
	ts := getTaskList()
	defer func() {
		// 清理数据
		if err != nil {
			if er := ts.Run(); er != nil {
				err = er
			}
		}
	}()
	for _, name := range fileNames {
		ts.Clean()
		// 本地转码缓存文件(workdir + name + ".temp." + ext)
		localTempFile := path.Join(w, path.Base(name)+".temp."+ext)
		// 文件上传原始文件备份名称(remoteName + ".temp")
		remoteTempFile := name + ".temp"
		// 文件下载路径(workdir + name)
		localName := path.Join(w, path.Base(name))
		// 远程文件名称，转码后结果文件，替换原始文件。(remoteName - oldExt + newExt)
		remoteFile := path.Join(path.Dir(name), strings.Replace(path.Base(name), path.Ext(name), "", -1)+"."+ext)
		fmt.Printf("\n转码%s流程开始->\n", path.Base(name))
		time.Sleep(time.Second * 5)

		fmt.Printf("下载中:%s\n", name)

		dlSrcFile, er := sp.Open(name)
		if er != nil {
			err = er
			return
		}
		defer dlSrcFile.Close()

		dlDstFile, er := os.OpenFile(localName, os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if er != nil {
			err = er
			return
		}
		defer dlDstFile.Close()
		// 出错后已下载本地文件清理
		ts.Add(func() error {
			fmt.Println("remove", localName)
			return os.Remove(localName)
		})

		er = sp.Download(ctx, dlSrcFile, dlDstFile)
		if er != nil {
			err = er
			return
		}

		fmt.Printf("开始转码: %s\n", path.Base(name))
		// transcode
		er = transcode.Run(ctx, w, localName, localTempFile, ext, cmd)
		if er != nil {
			// 出错后本地转码缓存文件清理
			ts.Add(func() error {
				return os.Remove(localTempFile)
			})
			err = er
			return
		}

		// rename
		er = sp.Rename(name, remoteTempFile)
		if er != nil {
			err = er
			return
		}

		//  出错后远程备份文件还原
		ts.Add(func() error {
			return sp.Rename(remoteTempFile, name)
		})

		fmt.Printf("开始上传: %s\n", localTempFile)
		// upload
		ulSrcFile, er := os.Open(localTempFile)
		if er != nil {
			err = er
			return
		}
		defer ulSrcFile.Close()

		ulDstFile, er := sp.OpenFile(remoteFile, os.O_WRONLY|os.O_CREATE)
		if er != nil {
			err = er
			return
		}
		defer ulDstFile.Close()
		// 出错后已上传的文件清理
		ts.Add(func() error {
			return sp.Remove(remoteFile)
		})

		er = sp.Upload(ctx, ulSrcFile, ulDstFile)
		if er != nil {
			err = er
			return
		}

		// 前面主要任务无错误,无须执行错误清理任务
		ts.Not()

		// deleteRemoteTemp
		er = sp.Remove(remoteTempFile)
		if er != nil {
			err = er
			return
		}

		// locker
		fk := file_locker.New()
		er = fk.Lock(remoteFile, sp)
		if er != nil {
			err = er
			return
		}

		// delete
		er = os.Remove(localTempFile)
		if er != nil {
			err = er
			return
		}
		er = os.Remove(localName)
		if er != nil {
			err = er
			return
		}
	}
	return
}

// getTaskList 获取任务执行实例
func getTaskList() *taskList {
	return &taskList{}
}

type taskList struct {
	is   bool
	list []func() error
}

// Add 添加任务
func (t *taskList) Add(f func() error) {
	t.list = append(t.list, f)
}

// Clean 清除任务
func (t *taskList) Clean() {
	t.list = make([]func() error, 0)
}

// Is 需要执行
func (t *taskList) Is() {
	t.is = false
}

// Not 不需要执行
func (t *taskList) Not() {
	t.is = true
}

// Run 运行所有任务
func (t *taskList) Run() (err error) {
	if !t.is {
		t.Reverse()
		for _, f := range t.list {
			er := f()
			if er != nil {
				err = er
			}
		}
	}
	return
}

// Reverse 反转任务
func (t *taskList) Reverse() {
	for i, j := 0, len(t.list)-1; i < j; i, j = i+1, j-1 {
		t.list[i], t.list[j] = t.list[j], t.list[i]
	}
}
