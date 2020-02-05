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
func Run(ctx context.Context, u string, pass string, addr string, port int, w string, r string, cmd string, ext string, formats string, filters string) {
	var (
		client *sftp.Client
		err    error
	)
	client, err = sp.Connect(u,
		pass,
		addr,
		port)
	if err != nil {
		fmt.Printf("sftp connect error: %v\n", err)
		return
	}
	defer client.Close()
	files, err := sp.ReadDir(r, client)
	if err != nil {
		fmt.Printf("reading sftp \"%s\" directory error: %v\n", r, err)
		return
	}
	pts, err := sftpFiles(files, r, formats, filters, client)
	if err != nil {
		fmt.Printf("reading for sftp remote directory error: %v\n", err)
		return
	}
	fmt.Printf("\n待转码列表:\n")
	for _, f := range pts {
		fmt.Println(f)
	}
	fmt.Printf("\n开始转码作业:\n")
	err = batchTranscode(ctx, pts, w, ext, cmd, client)
	if err != nil {
		fmt.Printf("\nbatch transcode video error: %v\n", err)
		return
	}
}

func sftpFiles(fs []os.FileInfo, basePath string, formats string, filters string, client *sftp.Client) (paths []string, err error) {
	for _, f := range fs {
		if f.IsDir() {
			nPath := path.Join(basePath, f.Name())
			nfs, er := sp.ReadDir(nPath, client)
			if er != nil {
				err = er
				return
			}
			pts, er := sftpFiles(nfs, nPath, formats, filters, client)
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
					if strings.Contains(strings.ToLower(path.Join(basePath, f.Name())), strings.Trim(filter, " ")) {
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

func batchTranscode(ctx context.Context, fileNames []string, w string, ext string, cmd string, client *sftp.Client) (err error) {
	for _, name := range fileNames {
		tempFile := path.Join(w, path.Base(name)+".temp."+ext)
		remoteTempFile := name + ".temp"
		baseName := path.Join(w, path.Base(name))
		remoteFile := path.Join(path.Dir(name), strings.Replace(path.Base(name), path.Ext(name), "", -1)+"."+ext)
		fmt.Printf("\n转码%s流程开始->\n", path.Base(name))
		time.Sleep(time.Second * 5)

		fmt.Printf("下载中:%s\n", name)
		er := sp.Download(ctx, name, baseName, client)
		if er != nil {
			select {
			case <-ctx.Done():
				if e := os.Remove(baseName); e != nil {
					err = e
					return
				}
				err = er
				return
			default:
				err = er
				return
			}
		}

		fmt.Printf("开始转码: %s\n", path.Base(name))
		// transcode
		er = transcode.Run(ctx, w, baseName, tempFile, ext, cmd)
		if er != nil {
			select {
			case <-ctx.Done():
				if e := os.Remove(tempFile); e != nil {
					err = e
					return
				}
				if e := os.Remove(baseName); e != nil {
					err = e
					return
				}
				err = er
				return
			default:
				err = er
				return
			}
		}

		// rename
		er = sp.Rename(name, remoteTempFile, client)
		if er != nil {
			err = er
			return
		}

		fmt.Printf("开始上传: %s\n", tempFile)
		// upload
		er = sp.Upload(ctx, tempFile, remoteFile, client)
		if er != nil {
			select {
			case <-ctx.Done():
				if e := sp.Rename(remoteTempFile, name, client); e != nil {
					err = e
					return
				}
				if e := os.Remove(tempFile); e != nil {
					err = e
					return
				}
				if e := os.Remove(baseName); e != nil {
					err = e
					return
				}
				err = er
				return
			default:
				err = er
				return
			}
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
	return
}
