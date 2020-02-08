// Package worker 工作
package worker

import (
	"context"
	"fmt"
)

const (
	MODE_SFTP = "sftp"
)

// Run 开始转码作业
func Run(ctx context.Context, u string, pass string, addr string, port int, w string, r string, cmd string, ext string, formats string, filters string, mode string) (err error) {
	switch mode {
	case MODE_SFTP:
		er := runSFTP(ctx, u, pass, addr, port, w, r, cmd, ext, formats, filters)
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
