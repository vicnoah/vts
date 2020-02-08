package file_transfer

// NewFileTransfer 获取file transfer实例
func NewFileTransfer(fs Fser) *FileTransfer {
	return &FileTransfer{Fser: fs}
}

// FileTransfer 文件操作结构
type FileTransfer struct {
	Fser
}
