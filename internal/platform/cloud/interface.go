package cloud

// StorageUtility interface
type StorageUtility interface {
	GetFileByFileName(fileName string, directoryCloud string) ([]byte, error)
	UploadFileToCloud(file string, fileName string, directoryCloud string) error
	DeleteFileCloud(fileName string, directoryCloud string) error
}
