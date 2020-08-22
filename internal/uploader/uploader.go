// Package uploader contains logic related with the uploading process of the backup
package uploader

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// Uploader represents the uploading process
type Uploader struct {
	bucket string
	region string
}

// New instantiates an Uploader
func New(bucket, region string) Uploader {
	return Uploader{
		bucket: bucket,
		region: region,
	}
}

// Sync uploads the content of the given path S3
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
	log.Printf("Uploading %d objects", iter.Length())
	if err := uploader.UploadWithIterator(aws.BackgroundContext(), iter); err != nil {
		log.Fatal(err)
	}
}
