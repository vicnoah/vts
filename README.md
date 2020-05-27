# vts

1.首先需要系统安装有docker，并使用linux系统，要求docker能直接调用GPU，如果使用虚拟机需要能够将GPU直通，否则只能使用CPU转码，速度会相当慢

2.转码命令模板方式执行转码，每一个视频文件都会调用转码模板进行转码。

```shell
docker run -i --rm -v=%workdir%:%workdir% --device /dev/dri/renderD128 jrottenberg/ffmpeg:4.1-vaapi -hwaccel vaapi -hwaccel_output_format vaapi -hwaccel_device /dev/dri/renderD128 -i %input% -c:v vp9_vaapi -c:a libvorbis %output%
```

- 具体命令参考ffmpeg使用方法，可参考[docker ffmpeg](https://hub.docker.com/r/jrottenberg/ffmpeg/ "docker ffmpeg")的使用方法，例子是基于英特尔集显vaapi的，也可参考docker ffmpeg调用nvdia独显的nvenc。

- 旧CPU及旧显卡并不支持最新的视频编码具体支持程度请参考intel vaapi，nvida nvenc官网的描述

- 模版支持变量语法采用%name%方式，除%workdir%外另外两个参数由程序自动生成，无需关系具体内容

  %workdir%是软件工作时的缓存目录

  %input%是转码源文件路径

  %output%为转码目标文件路径

3.可使用vts --help查看具体用法

```shell
vts version: vts/1.0.0
Usage: vts [-u username] [-pass password] [-add network address] [-w workdir] [-w remotedir]

Options:
  -bs file copy cache size
    	file copy cache size in KB (default 10240)
  -cmd command template
    	command template. Similar to %name% are variables. %workdir%: work directory, %input%: input video file, %output%: output vodeo file, %ext%: ext name (default "docker run -i --rm -v=%workdir%:%workdir% --device /dev/dri/renderD128 jrottenberg/ffmpeg:4.1-vaapi -hwaccel vaapi -hwaccel_output_format vaapi -hwaccel_device /dev/dri/renderD128 -i %input% -c:v vp9_vaapi -c:a libvorbis %output%")
  -ext target file ext name
    	target file ext name (default "webm")
  -f Use the CMD file file instead of the CMD command
    	Use the CMD file file instead of the CMD command. Use UTF-8 encoding for the file.
  -flt filting's path
    	You would like to filting's path (default "vr")
  -fmt video's extension name
    	You would like to transcoding video's extension name (default "mp4, mpeg4, wmv, mkv, avi")
  -help
    	help
  -mode work mode.
    	work mode. eg: sftp, local, nfs (default "sftp")
  -r remote directory
    	set remote directory path (default "/emby/video")
  -s:addr network address
    	sftp server's network address (default "192.168.0.1")
  -s:auth SFTP authentication mode
    	SFTP authentication mode supports password authentication and key authentication (default "password")
  -s:iden:file sftp private key file path
    	sftp private key file path (default "~/.ssh/id_rsa")
  -s:iden:pass sftp private key's password
    	sftp private key's password
  -s:pass user password
    	sftp connect's user password (default "123456")
  -s:port port
    	sftp's port (default 22)
  -s:user user name
    	sftp connect's user name (default "root")
  -w workdir
    	local workdir: download, transcode, upload (default "~/vts")
```



3.开始转码，转码成功后会替换源文件，并且会将路径中所有视频文件都转码。建议将需转码视频复制到另一目录进行转码。（基于模板语法可实现灵活的转码逻辑，其它转码方式参考ffmpeg，docker-ffmpeg，及vts模板参数）

- 转码sftp文件为webm格式

  -f 参数用于指定转码模板文件

  -ext 参数用于指定转码的目标格式

  -w 参数用于指定转码缓存目录

  -r 待转码视频的目录

  其它参数用于指定sftp协议的信息，可使用sftp密码访问，例子中使用的是本地密钥访问

vts -f ~/vaapi_vp9.sh -ext webm -flt "" -w ~/vts -r /data/video -s:addr 192.168.5.4 -s:auth key -s:user root

- 转码本地文件为webm格式

  参数参考上一个例子

  转码本地文件需指定-mode参数

  -mode local 代表转码本地文件

vts -f ~/vaapi_vp9.sh -ext webm -flt "" -w ~/vts -r /data/video -mode local

