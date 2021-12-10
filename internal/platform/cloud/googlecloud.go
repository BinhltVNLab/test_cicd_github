package cloud

import (
	"cloud.google.com/go/storage"
	"context"
	"encoding/base64"
	"github.com/labstack/echo/v4"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// GcsStorage struct
type GcsStorage struct {
	Logger echo.Logger
	Ctx    context.Context
	Bucket *storage.BucketHandle
}

// NewGcsStorage export
func NewGcsStorage(logger echo.Logger) (cloud *GcsStorage) {
	cloud = &GcsStorage{}
	cloud.Init(logger)
	return
}

// Init new connect
func (cloud *GcsStorage) Init(logger echo.Logger) {
	cloud.Logger = logger
	cloud.Ctx = context.Background()
	client, err := storage.NewClient(cloud.Ctx)
	if err != nil {
		cloud.Logger.Error(err)
	}

	cloud.Bucket = client.Bucket(os.Getenv("GOOGLE_STORAGE_BUCKET"))
}

// GetFileByFileName show image
func (cloud *GcsStorage) GetFileByFileName(fileName string, directoryCloud string) ([]byte, error) {
	linkCloud := directoryCloud + fileName
	reader, err := cloud.Bucket.Object(linkCloud).NewReader(cloud.Ctx)
	if err != nil {
		cloud.Logger.Error(err)
		return nil, err
	}
	defer reader.Close()

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		cloud.Logger.Error(err)
		return nil, err
	}

	return data, nil
}

// UploadFileToCloud : upload image to cloud
// Params      : directory file name
// Returns     : error or nil
func (cloud *GcsStorage) UploadFileToCloud(file string, fileName string, directoryCloud string) error {
	linkCloud := directoryCloud + fileName
	wc := cloud.Bucket.Object(linkCloud).NewWriter(cloud.Ctx)
	f := base64.NewDecoder(base64.StdEncoding, strings.NewReader(file))
	if _, err := io.Copy(wc, f); err != nil {
		cloud.Logger.Error(err)
		return err
	}

	if err := wc.Close(); err != nil {
		cloud.Logger.Error(err)
		return err
	}
	// [END upload_file]
	return nil
}

// DeleteFileCloud : delete image in cloud
// Params      : file name
// Returns     : error or nil
func (cloud *GcsStorage) DeleteFileCloud(fileName string, directoryCloud string) error {
	linkCloud := directoryCloud + fileName
	o := cloud.Bucket.Object(linkCloud)

	if err := o.Delete(cloud.Ctx); err != nil {
		cloud.Logger.Error(err)
		return err
	}

	return nil
}
