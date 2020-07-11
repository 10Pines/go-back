package upload

import (
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

type SyncFolderIterator struct {
	bucket    string
	fileInfos []fileInfo
	err       error
}

type fileInfo struct {
	key      string
	fullPath string
}

func NewSyncFolderIterator(path, bucket string) *SyncFolderIterator {
	var metadata []fileInfo
	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			key := strings.TrimPrefix(p, path)
			metadata = append(metadata, fileInfo{key, p})
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	return &SyncFolderIterator{
		bucket,
		metadata,
		nil,
	}
}

// Next will determine whether or not there is any remaining files to
// be uploaded.
func (iter *SyncFolderIterator) Next() bool {
	return len(iter.fileInfos) > 0
}

// Err returns any error when os.Open is called.
func (iter *SyncFolderIterator) Err() error {
	return iter.err
}

// UploadObject will prep the new upload object by open that file and constructing a new
// s3manager.UploadInput.
func (iter *SyncFolderIterator) UploadObject() s3manager.BatchUploadObject {
	fi := iter.fileInfos[0]
	iter.fileInfos = iter.fileInfos[1:]
	body, err := os.Open(fi.fullPath)
	if err != nil {
		iter.err = err
	}

	extension := filepath.Ext(fi.key)
	mimeType := mime.TypeByExtension(extension)

	if mimeType == "" {
		mimeType = "binary/octet-stream"
	}

	input := s3manager.UploadInput{
		Bucket:      &iter.bucket,
		Key:         &fi.key,
		Body:        body,
		ContentType: &mimeType,
	}

	return s3manager.BatchUploadObject{
		Object: &input,
	}
}

func (iter *SyncFolderIterator) Length() int {
	return len(iter.fileInfos)
}
