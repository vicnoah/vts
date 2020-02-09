package boot

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"syscall"

	"github.com/vicnoah/vts/worker"
)

var (
	help             bool   // 帮助
	workDir          string // 本地工作路径
	remoteDir        string // 远程工作路径
	sftpUser         string // sftp用户名
	sftpPass         string // sftp密码
	sftpAddr         string // sftp地址
	sftpPort         int    // sftp端口
	cmd              string // 转码命令
	mode             string // 工作模式
	ext              string // 文件扩展名
	sftpAuth         string // sftp认证方式
	sftpIdentityFile string // openssh密钥文件
	sftpIdentityPass string // openssh私钥密码
	formats          string // 需转码格式
	filters          string // 需过滤路径包含字符
)

func init() {
	flag.BoolVar(&help, "help", false, "help")

	flag.StringVar(&cmd, "cmd",
		`docker run -i --rm -v=%workdir%:%workdir% --device /dev/dri/renderD128 jrottenberg/ffmpeg:4.1-vaapi -hwaccel vaapi -hwaccel_output_format vaapi -hwaccel_device /dev/dri/renderD128 -i %input% -c:v vp9_vaapi -c:a libvorbis %output%`, "`command template`. Similar to %name% are variables. %workdir%: work directory, %input%: input video file, %output%: output vodeo file, %ext%: ext name")
	flag.StringVar(&ext, "ext", "webm", "`target file ext name`")
	flag.StringVar(&formats, "fmt", "mp4, mpeg4, wmv, mkv, avi", "You would like to transcoding `video's extension name`")
	flag.StringVar(&filters, "flt", "vr", "You would like to `filting's path`")

	flag.StringVar(&mode, "mode", "sftp", "`work mode.` eg: sftp, local, nfs")

	flag.StringVar(&workDir, "w", "/opt/vts", "local `workdir`: download, transcode, upload")
	flag.StringVar(&remoteDir, "r", "/emby/video", "set `remote directory` path")

	flag.StringVar(&sftpUser, "sftp_user", "root", "sftp connect's `user name`")
	flag.StringVar(&sftpPass, "sftp_pass", "123456", "sftp connect's `user password`")
	flag.StringVar(&sftpAddr, "sftp_addr", "192.168.0.1", "sftp server's `network address`")
	flag.IntVar(&sftpPort, "sftp_port", 22, "sftp's `port`")
	flag.StringVar(&sftpAuth, "sftp_auth", "password", "`SFTP authentication mode` supports password authentication and key authentication")
	flag.StringVar(&sftpIdentityFile, "sftp_identity_file", "~/.ssh/id_rsa", "`sftp private key file path`")
	flag.StringVar(&sftpIdentityPass, "sftp_indentity_pass", "", "`sftp private key's password`")

	// 改变默认的 Usage，flag包中的Usage 其实是一个函数类型。这里是覆盖默认函数实现，具体见后面Usage部分的分析
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `vts version: vts/1.0.0
Usage: vts [-u username] [-pass password] [-add network address] [-w workdir] [-w remotedir]

Options:
`)
		flag.PrintDefaults()
	}
}

// Start 启动程序
func Start() {
	flag.Parse()
	// 路径一致处理
	workDir = path.Clean(workDir)
	remoteDir = path.Clean(remoteDir)
	// 系统home环境变量处理
	if runtime.GOOS == "windows" {
		homeEnv := "USERPROFILE"
		envKey := "%userprofile%"
		if strings.Contains(sftpIdentityFile, envKey) {
			sftpIdentityFile = parseHome(homeEnv, envKey, sftpIdentityFile)
		}
	} else {
		homeEnv := "HOME"
		envKey := "~"
		if strings.Contains(sftpIdentityFile, envKey) {
			sftpIdentityFile = parseHome(homeEnv, envKey, sftpIdentityFile)
		}
	}

	if help {
		flag.Usage()
		return
	}
	//创建监听退出chan
	c := make(chan os.Signal)
	defer close(c)
	ctx, cancel := context.WithCancel(context.Background())
	//监听指定信号 ctrl+c kill
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		for s := range c {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				cancel()
				fmt.Printf("\nQuit: %s\n", s)
			default:
				fmt.Printf("\nother: %s\n", s)
			}
		}
	}()

	boot(ctx)
	return
}

// boot 引导主程序
func boot(ctx context.Context) {
	err := worker.Run(ctx, cmd, ext, formats, filters, workDir, remoteDir, mode, sftpUser, sftpPass, sftpAddr, sftpPort, sftpAuth, sftpIdentityFile, sftpIdentityPass)
	if err != nil {
		fmt.Printf("\n%v\n", err)
	}
	return
}

// parseHome 解析系统HOME环境变量值
func parseHome(homeEnv, envKey, iden string) (path string) {
	environ := os.Environ()
	for _, env := range environ {
		e := strings.Split(env, "=")
		if e[0] == homeEnv {
			path = strings.Replace(iden, envKey, e[1], 1)
		}
	}
	return
}
