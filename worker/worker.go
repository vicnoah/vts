package worker

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	file_locker "github.com/vicnoah/vts/fileLocker"
	sp "github.com/vicnoah/vts/sftp"
	"github.com/vicnoah/vts/transcode"

	"github.com/pkg/sftp"
)

// Run 开始转码作业
func Run(ctx context.Context, u string, pass string, addr string, port int, w string, r string, formats string, filters string) {
	var (
		client *sftp.Client
		err    error
	)
	client, err = sp.Connect(u,
		pass,
		addr,
		port)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer client.Close()
	files, err := sp.ReadDir(client, r)
	if err != nil {
		fmt.Println(err)
		return
	}
	pts, err := sftpFiles(files, r, client, formats, filters)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("\n待转码列表:\n")
	for _, f := range pts {
		fmt.Println(f)
	}
	fmt.Printf("\n开始转码作业:")
	err = batchTranscode(ctx, pts, client, w)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func sftpFiles(fs []os.FileInfo, basePath string, client *sftp.Client, formats string, filters string) (paths []string, err error) {
	for _, f := range fs {
		if f.IsDir() {
			nPath := path.Join(basePath, f.Name())
			nfs, er := sp.ReadDir(client, nPath)
			if er != nil {
				err = er
				return
			}
			pts, er := sftpFiles(nfs, nPath, client, formats, filters)
			if er != nil {
				err = er
				return
			}
			paths = append(paths, pts...)
		} else {
			isExist := false
			for _, format := range strings.Split(formats, ",") {
				if strings.Contains(strings.ToLower(path.Ext(f.Name())), strings.Trim(format, " ")) {
					isExist = true
				}
			}
			if isExist {
				isFilter := false
				for _, filter := range strings.Split(filters, ",") {
					if strings.Contains(strings.ToLower(f.Name()), strings.Trim(filter, " ")) {
						isFilter = true
					}
				}
				if !isFilter {
					fk := file_locker.New()
					isLock, er := fk.IsLock(path.Join(basePath, f.Name()), client)
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

func batchTranscode(ctx context.Context, fileNames []string, client *sftp.Client, w string) (err error) {
	for _, name := range fileNames {
		select {
		case <-ctx.Done():
			return
		default:
			tempFile := path.Join(w, path.Base(name)+".temp.webm")
			remoteTempFile := name + ".temp"
			baseName := path.Join(w, path.Base(name))
			remoteFile := path.Join(path.Dir(name), strings.Replace(path.Base(name), path.Ext(name), "", -1)+".webm")
			fmt.Printf("\n转码%s流程开始->\n", path.Base(name))
			time.Sleep(time.Second * 5)
			fmt.Printf("下载中:%s\n", name)
			er := sp.Download(client, name, baseName)
			if er != nil {
				err = er
				return
			}

			fmt.Printf("\n开始转码: %s\n", path.Base(name))

			// transcode
			er = transcode.WebM(w, baseName, tempFile)
			if er != nil {
				err = er
				return
			}

			// rename
			er = sp.Rename(name, remoteTempFile, client)
			if er != nil {
				err = er
				return
			}

			fmt.Printf("开始上传: %s\n", tempFile)
			// upload
			er = sp.Upload(client, tempFile, remoteFile)
			if er != nil {
				err = er
				return
			}

			// deleteRemoteTemp
			er = sp.Remove(remoteTempFile, client)
			if er != nil {
				err = er
				return
			}

			// locker
			fk := file_locker.New()
			er = fk.Lock(remoteFile, client)
			if er != nil {
				err = er
				return
			}

			// delete
			er = os.Remove(tempFile)
			if er != nil {
				err = er
				return
			}
			er = os.Remove(baseName)
			if er != nil {
				err = er
				return
			}
		}
	}
	return
}
