package uploader

import (
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type (
	syncFolderIterator struct {
		bucket    string
		fileInfos []fileInfo
		err       error
		fileCount int
	}

	fileInfo struct {
		key      string
		fullPath string
	}
)

func newSyncFolderIterator(path, bucket string) *syncFolderIterator {
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

	fileCount := len(metadata)

	return &syncFolderIterator{
		bucket,
		metadata,
		nil,
		fileCount,
	}
}

func (iter *syncFolderIterator) Next() bool {
	return len(iter.fileInfos) > 0
}

func (iter *syncFolderIterator) Err() error {
	return iter.err
}

func (iter *syncFolderIterator) UploadObject() s3manager.BatchUploadObject {
	fi := iter.fileInfos[0]
	iter.fileInfos = iter.fileInfos[1:]
	log.Printf("Uploading %s, %d to go", fi.key, iter.Length())
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

func (iter *syncFolderIterator) Length() int {
	return len(iter.fileInfos)
}
