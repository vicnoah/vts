// Package worker 工作
package worker

import (
	"context"
	"fmt"
)

const (
	_MODE_LOCAL = "local"
	_MODE_SFTP  = "sftp"
)

// Run 开始转码作业
func Run(ctx context.Context, cmd, ext, formats, filters, workDir, remoteDir, mode, sftpUser, sftpPass, sftpAddr string, sftpPort int, sftpAuth, sftpIdentityFile, sftpIdentityPass string) (err error) {
	switch mode {
	case _MODE_SFTP:
		er := runSFTP(ctx, cmd, ext, formats, filters, workDir, remoteDir, sftpUser, sftpPass, sftpAddr, sftpPort, sftpAuth, sftpIdentityFile, sftpIdentityPass)
		if er != nil {
			err = er
			return
		}
	case _MODE_LOCAL:
		er := runLocal(ctx, cmd, ext, formats, filters, workDir, remoteDir)
		if er != nil {
			err = er
			return
		}
	default:
		err = fmt.Errorf("transcode mode %s is not supportd", mode)
		return
	}
	return
}
