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

func batch(ctx context.Context, fileNames []string, w string, ext string, cmd string, fr file_transfer.Fser) (err error) {
	ts := getTaskList()
	sp := file_transfer.NewFileTransfer(fr)
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
