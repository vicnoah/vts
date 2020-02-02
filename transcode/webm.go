package transcode

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// WebM 转换视频为webm
func WebM(workDir string, srcPath string, dstPath string) (err error) {
	command := fmt.Sprintf(vp9Vaapi, workDir, workDir, srcPath, dstPath)
	fmt.Printf("\n转码命令->\n%s\n", command)
	ss := strings.Split(command, " ")
	name := ss[0]
	args := ss[1:]
	cmd := exec.Command(name, args...)
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
