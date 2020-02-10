package worker

import (
	"context"
	"fmt"

	file_transfer "github.com/vicnoah/vts/fileTransfer"
)

const (
	_SFTP_AUTH_PASSWORD = "password"
	_SFTP_AUTH_KEY      = "key"
)

func runSFTP(ctx context.Context,
	cmd,
	ext,
	formats,
	filters,
	w,
	r string,
	bufferSize int,
	u,
	pass,
	addr string,
	port int,
	auth,
	identityFile,
	identityPass string) (err error) {
	var (
		ssh = file_transfer.NewSSH()
		sp  = file_transfer.NewSFTP()
	)
	if auth == _SFTP_AUTH_KEY {
		err = ssh.SetConfigWithKey(u,
			identityFile,
			identityPass,
			addr,
			port)
		if err != nil {
			err = fmt.Errorf("ssh config error: %v", err)
			return
		}
	} else if auth == _SFTP_AUTH_PASSWORD {
		err = ssh.SetConfig(u,
			pass,
			addr,
			port)
		if err != nil {
			err = fmt.Errorf("ssh config error: %v", err)
			return
		}
	} else {
		err = fmt.Errorf("ssh config err: %s %s", auth, "authentication mode does not exist")
		return
	}

	err = sp.Connect(ssh.Addr(), ssh.Config())
	if err != nil {
		err = fmt.Errorf("sftp connect error: %v", err)
		return
	}
	defer sp.Close()

	fd, err := sp.ReadDir(r)
	if err != nil {
		err = fmt.Errorf("reading sftp \"%s\" directory error: %v", r, err)
		return
	}
	pts, err := files(fd, r, formats, filters, sp)
	if err != nil {
		err = fmt.Errorf("reading for sftp remote directory error: %v", err)
		return
	}
	fmt.Printf("\n待转码列表:\n")
	for _, f := range pts {
		fmt.Println(f)
	}
	fmt.Printf("\n开始转码作业:\n")
	err = batch(ctx, pts, w, ext, cmd, bufferSize, sp)
	if err != nil {
		err = fmt.Errorf("batch transcode video error: %v", err)
		return
	}
	return
}
