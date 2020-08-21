package s3

import (
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type syncFolderIterator struct {
	bucket    string
	fileInfos []fileInfo
	err       error
	fileCount int
}

type fileInfo struct {
	key      string
	fullPath string
}

type Uploader struct {
	bucket string
	region string
}

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

// Next will determine whether or not there is any remaining files to
// be uploaded.
func (iter *syncFolderIterator) Next() bool {
	return len(iter.fileInfos) > 0
}

// Err returns any error when os.Open is called.
func (iter *syncFolderIterator) Err() error {
	return iter.err
}

// UploadObject will prep the new s3 object by open that file and constructing a new
// s3manager.UploadInput.
func (iter *syncFolderIterator) UploadObject() s3manager.BatchUploadObject {
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

func (iter *syncFolderIterator) Length() int {
	return len(iter.fileInfos)
}

func New(bucket, region string) Uploader {
	return Uploader{
		bucket: bucket,
		region: region,
	}
}

func (u Uploader) Sync(path string) {
	log.Printf("Syncronizing %s folder into Bucket[%s]", path, u.bucket)
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(u.region),
	})
	if err != nil {
		log.Fatal(err)
	}
	uploader := s3manager.NewUploader(sess)

	iter := newSyncFolderIterator(path, u.bucket)
	log.Printf("Uploading %d objects to s3", iter.Length())
	if err := uploader.UploadWithIterator(aws.BackgroundContext(), iter); err != nil {
		log.Fatal(err)
	}

	log.Printf("Finished uploading %s folder", path)
}
