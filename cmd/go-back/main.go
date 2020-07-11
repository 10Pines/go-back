package main

import (
	"flag"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	. "go-re/internal/repositories"
	. "go-re/internal/upload"
	"log"
)

const GitHubOrganizationName = "10Pines"
const GitLabOrganizationId = 1152254

func main() {
	bucket := *flag.String("bucket", "", "bucket to upload to")
	region := *flag.String("region", "", "region to be used when making requests")
	backupFolder := *flag.String("backupFolder", "repos", "Backup folder to clone repositories into")
	workerCount := *flag.Int("workers", 8, "Workers count")

	flag.Parse()

	auths := MakeAuthsFromEnv()
	cloneConfig := MakeCloneConfig(workerCount, backupFolder)

	ghRepositories := GetGithubRepos(auths, GitHubOrganizationName)
	log.Printf("Fetched %d repositories from GitHub", len(ghRepositories))

	glRepositories := GetGitlabRepos(auths, GitLabOrganizationId)
	log.Printf("Fetched %d repositories from GitLab", len(glRepositories))

	allRepositories := append(ghRepositories, glRepositories...)

	wg, cloneQueue := MakeCloneWorkerPool(cloneConfig)
	for _, repository := range allRepositories[:10] {
		cloneQueue <- repository
	}

	close(cloneQueue)
	wg.Wait()
	log.Printf("Cloned %d repositories", len(allRepositories))

	upload(bucket, region, backupFolder)
}

func upload(bucket string, region string, path string) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		log.Fatal(err)
	}
	uploader := s3manager.NewUploader(sess)

	iter := NewSyncFolderIterator(path, bucket)
	log.Printf("Uploading %d objects to s3", iter.Length())
	if err := uploader.UploadWithIterator(aws.BackgroundContext(), iter); err != nil {
		log.Fatal(err)
	}

	if err := iter.Err(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Finished uploading %s folder", path)
}
