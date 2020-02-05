package boot

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/vicnoah/vts/worker"
)

var (
	help      bool   // 帮助
	workDir   string // 本地工作路径
	remoteDir string // 远程工作路径
	user      string // 远程用户名
	pass      string // 远程密码
	addr      string // 地址
	port      int    // 远程端口
	cmd       string // 转码命令
	mode      string // 工作模式
	ext       string // 文件扩展名
	formats   string // 需转码格式
	filters   string // 需过滤路径包含字符
)

func init() {
	flag.BoolVar(&help, "help", false, "help")
	flag.StringVar(&user, "u", "root", "sftp connect's `user name`")
	flag.StringVar(&pass, "pass", "123456", "sftp connect's `user password`")
	flag.StringVar(&addr, "addr", "192.168.0.1", "sftp server's `network address`")
	flag.IntVar(&port, "p", 22, "sftp's `port`")
	flag.StringVar(&workDir, "w", "/opt/vts", "local `workdir`: download, transcode, upload")
	flag.StringVar(&remoteDir, "r", "/emby/video", "set `remote directory` path")
	flag.StringVar(&cmd, "cmd",
		`docker run -i --rm -v=%workdir%:%workdir% --device /dev/dri/renderD128 jrottenberg/ffmpeg:4.1-vaapi -hwaccel vaapi -hwaccel_output_format vaapi -hwaccel_device /dev/dri/renderD128 -i %input% -c:v vp9_vaapi -c:a libvorbis %output%`, "`command template`. Similar to %name% are variables. %workdir%: work directory, %input%: input video file, %output%: output vodeo file, %ext%: ext name")
	flag.StringVar(&ext, "ext", "webm", "`target file ext name`")
	flag.StringVar(&mode, "mode", "sftp", "`work mode.` eg: sftp, local, nfs")
	flag.StringVar(&formats, "fmt", "mp4, mpeg4, wmv, mkv, avi", "You would like to transcoding `video's extension name`")
	flag.StringVar(&filters, "flt", "vr", "You would like to `filting's path`")
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
	workDir = path.Clean(workDir)
	remoteDir = path.Clean(remoteDir)

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
	worker.Run(ctx, user, pass, addr, port, workDir, remoteDir, cmd, ext, formats, filters)
	return
}
