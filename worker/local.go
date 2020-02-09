package worker

import (
	"fmt"
	"context"
	file_transfer "github.com/vicnoah/vts/fileTransfer"
)

func runLocal(ctx context.Context, cmd, ext, formats, filters, w, r string) (err error) {
	var (
		sp  = file_transfer.NewLocal()
	)

	fd, err := sp.ReadDir(r)
	if err != nil {
		err = fmt.Errorf("reading local \"%s\" directory error: %v", r, err)
		return
	}
	pts, err := files(fd, r, formats, filters, sp)
	if err != nil {
		err = fmt.Errorf("reading for remote directory error: %v", err)
		return
	}
	fmt.Printf("\n待转码列表:\n")
	for _, f := range pts {
		fmt.Println(f)
	}
	fmt.Printf("\n开始转码作业:\n")
	err = batch(ctx, pts, w, ext, cmd, sp)
	if err != nil {
		err = fmt.Errorf("batch transcode video error: %v", err)
		return
	}
	return
}