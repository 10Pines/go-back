package main

import (
	"github.com/alexflint/go-arg"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	. "go-re/internal/compression"
	. "go-re/internal/repositories"
	. "go-re/internal/upload"
	"log"
	"os"
	"os/exec"
	"path"
)

const GitHubOrganizationName = "10Pines"
const GitLabOrganizationId = 1152254

var args struct {
	WorkerCount  int    `default:"8" help:"Number of cloning workers"`
	Bucket       string `arg:"required" help:"S3 bucket name to upload backups to"`
	Region       string `arg:"required" help:"S3 bucket region"`
	BackupFolder string `arg:"required" help:"Backup will be locally stored inside this folder"`
}

func main() {
	cmd := exec.Command("df", "-h")
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf(err.Error())
	}
	log.Print(string(output))

	arg.MustParse(&args)
	auths := MakeAuthsFromEnv()
	cloneConfig := MakeCloneConfig(args.WorkerCount, args.BackupFolder)

	ghRepositories := GetGithubRepos(auths, GitHubOrganizationName)
	log.Printf("Fetched %d repositories from GitHub", len(ghRepositories))

	glRepositories := GetGitlabRepos(auths, GitLabOrganizationId)
	log.Printf("Fetched %d repositories from GitLab", len(glRepositories))

	allRepositories := append(ghRepositories, glRepositories...)

	wg, cloneQueue := MakeCloneWorkerPool(cloneConfig)
	for _, repository := range allRepositories {
		cloneQueue <- repository
	}

	close(cloneQueue)
	wg.Wait()
	log.Printf("Cloned %d repositories", len(allRepositories))

	zipRepositories(allRepositories, cloneConfig)

	upload(args.Bucket, args.Region, path.Join(args.BackupFolder, "zip"))
}

func zipRepositories(allRepositories []*Repository, cloneConfig *CloneConfig) {
	for _, repository := range allRepositories {
		repositoryPath := cloneConfig.ClonePath(repository)
		output := cloneConfig.ZipPath(repository)
		err := os.MkdirAll(output, 0755)
		if err != nil {
			log.Fatalf(err.Error())
		}
		err = ZipFolder(repositoryPath, path.Join(output, repository.Name+".zip"))
		if err != nil {
			log.Fatalf(err.Error())
		}
	}
}

func upload(bucket string, region string, path string) {
	log.Printf("Syncronizing %s folder into Bucket[%s]", path, bucket)
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
