package transcode

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

/*
	docker run -i --rm -v=%s:/trs/%s
	--device /dev/dri/renderD128
	jrottenberg/ffmpeg:4.1-vaapi
	-hwaccel vaapi
	-hwaccel_output_format vaapi
	-hwaccel_device /dev/dri/renderD128
	-i /trs%s
	-c:v vp9_vaapi
	-c:a libvorbis
	/trs%s
*/

// Run 转换视频
func Run(ctx context.Context, workDir string, srcPath string, dstPath string, command string) (err error) {
	command = strings.ReplaceAll(command, volume, workDir)
	command = strings.ReplaceAll(command, input, srcPath)
	command = strings.ReplaceAll(command, output, dstPath)
	fmt.Printf("转码命令->\n%s\n", command)
	ss := strings.Split(command, " ")
	name := ss[0]
	args := ss[1:]
	cmd := exec.CommandContext(ctx, name, args...)
	//指定输出位置
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Start()
	if err != nil {
		return
	}
	err = cmd.Wait()
	if err != nil {
		return
	}
	return
}
