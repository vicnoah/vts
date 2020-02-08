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

func runSFTP(ctx context.Context, u string, pass string, addr string, port int, w string, r string, cmd string, ext string, formats string, filters string) (err error) {
	var (
		ssh = file_transfer.NewSSH()
		sp  = file_transfer.NewSFTP()
	)
	err = ssh.SetConfig(u,
		pass,
		addr,
		port)
	if err != nil {
		err = fmt.Errorf("ssh config error: %v", err)
		return
	}

	err = sp.Connect(ssh.Addr(), ssh.Config())
	if err != nil {
		err = fmt.Errorf("sftp connect error: %v", err)
		return
	}
	defer sp.Close()

	files, err := sp.ReadDir(r)
	if err != nil {
		err = fmt.Errorf("reading sftp \"%s\" directory error: %v", r, err)
		return
	}
	pts, err := sftpFiles(files, r, formats, filters, sp)
	if err != nil {
		err = fmt.Errorf("reading for sftp remote directory error: %v", err)
		return
	}
	fmt.Printf("\n待转码列表:\n")
	for _, f := range pts {
		fmt.Println(f)
	}
	fmt.Printf("\n开始转码作业:\n")
	err = batchTranscode(ctx, pts, w, ext, cmd, sp)
	if err != nil {
		err = fmt.Errorf("batch transcode video error: %v", err)
		return
	}
	return
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

func batchTranscode(ctx context.Context, fileNames []string, w string, ext string, cmd string, sfp *file_transfer.SFTP) (err error) {
	ts := getTaskList()
	sp := file_transfer.NewFileTransfer(sfp)
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
		ts.Is()
		// 文件下载路径(workdir + name)
		cacheName := path.Join(w, path.Base(name))
		// 本地转码缓存文件(workdir + name + ".temp." + ext)
		cacheTempFile := path.Join(w, path.Base(name)+".temp."+ext)
		// 远程文件名称，转码后结果文件，替换原始文件。(remoteName - oldExt + newExt)
		remoteFile := path.Join(path.Dir(name), strings.Replace(path.Base(name), path.Ext(name), "", -1)+"."+ext)
		// 文件上传原始文件备份名称(remoteName + ".temp")
		remoteTempFile := name + ".temp"

		fmt.Printf("\n转码%s流程开始->\n", path.Base(name))
		time.Sleep(time.Second * 5)

		fmt.Printf("下载中:%s\n", name)

		dlSrcFile, er := sp.Open(name)
		if er != nil {
			err = er
			return
		}
		defer dlSrcFile.Close()

		dlDstFile, er := os.OpenFile(cacheName, os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if er != nil {
			err = er
			return
		}
		defer dlDstFile.Close()
		// 出错后已下载本地文件清理
		ts.Add(func() error {
			return os.Remove(cacheName)
		})

		er = sp.Recv(ctx, dlSrcFile, dlDstFile)
		if er != nil {
			err = er
			return
		}

		fmt.Printf("开始转码: %s\n", path.Base(name))
		// transcode
		er = transcode.Run(ctx, w, cacheName, cacheTempFile, ext, cmd)
		if er != nil {
			// 出错后本地转码缓存文件清理
			ts.Add(func() error {
				return os.Remove(cacheTempFile)
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

		fmt.Printf("开始上传: %s\n", cacheTempFile)
		// upload
		ulSrcFile, er := os.Open(cacheTempFile)
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

		er = sp.Send(ctx, ulSrcFile, ulDstFile)
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
		er = os.Remove(cacheTempFile)
		if er != nil {
			err = er
			return
		}
		er = os.Remove(cacheName)
		if er != nil {
			err = er
			return
		}
	}
	return
}
