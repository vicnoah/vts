package sftp

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"time"

	"github.com/pkg/sftp"
	"github.com/vicnoah/progress"
	"golang.org/x/crypto/ssh"
)

// Connect 连接sftp
func Connect(user, password, host string, port int) (client *sftp.Client, err error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		sshClient    *ssh.Client
	)
	// get auth method
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(password))

	clientConfig = &ssh.ClientConfig{
		User:            user,
		Auth:            auth,
		Timeout:         30 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //ssh.FixedHostKey(hostKey),
	}

	// connet to ssh
	addr = fmt.Sprintf("%s:%d", host, port)
	sshClient, err = ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return
	}

	// create sftp client
	client, err = sftp.NewClient(sshClient)
	return
}

// ReadDir 读取目录
func ReadDir(sftpClient *sftp.Client, path string) (files []os.FileInfo, err error) {
	files, err = sftpClient.ReadDir(path)
	return
}

// Upload 上传文件
func Upload(sftpClient *sftp.Client, localPath string, remotePath string) (err error) {
	fileInfo, err := os.Stat(localPath)
	if err != nil {
		return
	}
	srcFile, err := os.Open(localPath)
	if err != nil {
		return
	}
	defer srcFile.Close()

	size := fileInfo.Size()
	r := progress.NewReader(srcFile)
	// Start a goroutine printing progress
	go func() {
		ctx := context.Background()
		progressChan := progress.NewTicker(ctx, r, size, 1*time.Second)
		for p := range progressChan {
			fmt.Printf("\ruploading: %s per: %d, remaining: %v", path.Base(localPath), int64(math.Floor(p.Percent())), p.Remaining().Round(time.Second))
		}
	}()

	dstFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return
	}
	defer dstFile.Close()

	_, err = io.CopyBuffer(dstFile, r, make([]byte, 204800))
	return
}

// Download 下载文件
func Download(sftpClient *sftp.Client, remotePath string, localPath string) (err error) {
	fileInfo, err := sftpClient.Stat(remotePath)
	if err != nil {
		return
	}
	srcFile, err := sftpClient.Open(remotePath)
	if err != nil {
		return
	}
	defer srcFile.Close()

	size := fileInfo.Size()
	r := progress.NewReader(srcFile)
	// Start a goroutine printing progress
	go func() {
		ctx := context.Background()
		progressChan := progress.NewTicker(ctx, r, size, 1*time.Second)
		for p := range progressChan {
			fmt.Printf("\rdownloading: %s per: %d, remaining: %v", path.Base(remotePath), int64(math.Floor(p.Percent())), p.Remaining().Round(time.Second))
		}
	}()

	dstFile, err := os.Create(localPath)
	if err != nil {
		return
	}
	defer dstFile.Close()
	_, err = io.CopyBuffer(dstFile, r, make([]byte, 204800))
	return
}

// WriteFile 写文件
func WriteFile(fileName string, sftpClient *sftp.Client, f int, b []byte) (err error) {
	file, err := sftpClient.OpenFile(fileName, f)
	if err != nil {
		return
	}
	defer file.Close()
	_, err = file.Write(b)
	return
}

// Remove 删除文件
func Remove(fileName string, sftpClient *sftp.Client) error {
	return sftpClient.Remove(fileName)
}

// Rename 重命名文件
func Rename(oldName string, newName string, sftpClient *sftp.Client) error {
	fmt.Println(oldName)
	fmt.Println(newName)
	return sftpClient.Rename(oldName, newName)
}

// Stat 读取文件状态
func Stat(fileName string, sftpClient *sftp.Client) (os.FileInfo, error) {
	return sftpClient.Stat(fileName)
}
