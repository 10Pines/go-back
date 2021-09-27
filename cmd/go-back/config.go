package main

import (
	"log"
	"os"
	"time"

	"go-re/internal/backup"
	"go-re/internal/repository"
	"go-re/internal/stats"

	"github.com/10Pines/tracker/v2/pkg/tracker"
	"github.com/alexflint/go-arg"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

type env struct {
	GitLabToken   string
	GitHubToken   string
	TrackerAPIKey string
}

type appArgs struct {
	WorkerCount      int    `default:"8" help:"Number of cloning workers"`
	Bucket           string `arg:"required" help:"S3 bucket name where the backup is stored"`
	Region           string `arg:"required" help:"S3 bucket region"`
	BackupFolder     string `arg:"required" help:"Backup will be locally stored inside this folder"`
	MetricsNamespace string `arg:"required" help:"Cloudwatch namespace where metrics will be published"`
	MetricsRegion    string `arg:"required" help:"Cloudwatch region where metrics will be published"`
	TaskName         string `arg:"required" help:"Tracker task name"`
}

func mustParseArgs() appArgs {
	args := appArgs{}
	arg.MustParse(&args)
	return args
}

func buildProviders(environment env) (repository.GitHub, repository.GitLab) {
	gh := repository.NewGitHub(environment.GitHubToken)
	gl := repository.NewGitLab(environment.GitLabToken)
	return gh, gl
}

func buildBackups(args appArgs, t *tracker.Tracker, timestamp time.Time) backup.Backup {
	config := backup.Config{
		BasePath:     args.BackupFolder,
		WorkerCount:  args.WorkerCount,
		Bucket:       args.Bucket,
		BucketRegion: args.Region,
		Tracker:      t,
	}
	cw := buildCloudwatchClient(args.MetricsRegion)
	reporter := stats.NewReporter(timestamp, args.MetricsNamespace, cw, t, args.TaskName)
	return backup.New(config, reporter)
}

func buildCloudwatchClient(region string) *cloudwatch.CloudWatch {
	config := &aws.Config{
		Region: aws.String(region),
	}
	mySession := session.Must(session.NewSession(config))
	return cloudwatch.New(mySession)
}

func mustParseEnv() env {
	gitHubToken, found := os.LookupEnv("GH_TOKEN")
	if !found {
		log.Fatal("GH_TOKEN is missing")
	}
	gitLabToken, found := os.LookupEnv("GL_TOKEN")
	if !found {
		log.Fatal("GL_TOKEN is missing")
	}
	trackerAPIKey, found := os.LookupEnv("TRACKER_API_KEY")
	if !found {
		log.Fatal("TRACKER_API_KEY is missing")
	}

	return env{
		GitHubToken:   gitHubToken,
		GitLabToken:   gitLabToken,
		TrackerAPIKey: trackerAPIKey,
	}
}
