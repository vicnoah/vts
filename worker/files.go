package worker

import (
	"os"
	"path"
	"strings"

	file_locker "github.com/vicnoah/vts/fileLocker"
	file_transfer "github.com/vicnoah/vts/fileTransfer"
)

// files 获取待转码文件列表
func files(fs []os.FileInfo, basePath string, formats string, filters string, fr file_transfer.Fser) (paths []string, err error) {
	for _, f := range fs {
		if f.IsDir() {
			nPath := path.Join(basePath, f.Name())
			nfs, er := fr.ReadDir(nPath)
			if er != nil {
				err = er
				return
			}
			pts, er := files(nfs, nPath, formats, filters, fr)
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
					isLock, er := fk.IsLock(path.Join(basePath, f.Name()), fr)
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
