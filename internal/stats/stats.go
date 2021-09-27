// Package stats contains backup related metrics logic
package stats

import (
	"log"
	"sync"
	"time"

	"github.com/10Pines/tracker/v2/pkg/tracker"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

type (
	stats struct {
		Time      int64
		RepoStats []RepoStats
	}

	// RepoStats contains information related with the clone and compression steps
	RepoStats struct {
		Name  string
		Host  string
		Clone int64
		Zip   int64
	}

	// Reporter is in charge of building and publishing metrics
	Reporter struct {
		cw        *cloudwatch.CloudWatch
		start     time.Time
		s         stats
		mu        sync.Mutex
		totalTime int64
		namespace string
		tracker   *tracker.Tracker
		taskName  string
	}
)

// NewReporter returns a new instance
func NewReporter(start time.Time, namespace string, cloudwatchClient *cloudwatch.CloudWatch, t *tracker.Tracker, taskName string) *Reporter {
	return &Reporter{
		cw:        cloudwatchClient,
		start:     start,
		s:         stats{},
		totalTime: 0,
		namespace: namespace,
		tracker:   t,
		taskName:  taskName,
	}
}

// Finished compute and publish metrics
func (r *Reporter) Finished() {
	if r.totalTime != 0 {
		log.Fatal("finish already called")
	}
	r.computeElapsedTime()
	r.report()
	r.reportBackup()
}

func (r *Reporter) computeElapsedTime() {
	r.totalTime = time.Since(r.start).Milliseconds()
}

// TrackRepository ingest a single repository stats. It's concurrent-safe
func (r *Reporter) TrackRepository(repoStat RepoStats) {
	r.mu.Lock()
	r.s.RepoStats = append(r.s.RepoStats, repoStat)
	r.mu.Unlock()
}

func (r *Reporter) report() {
	for _, repoStat := range r.s.RepoStats {
		r.putRepoMetric(repoStat)
	}
	r.putCountAndTimeMetrics()
}

func (r *Reporter) putRepoMetric(stat RepoStats) {
	metricDimensions := []*cloudwatch.Dimension{
		{
			Name:  aws.String("Repository"),
			Value: aws.String(stat.Name),
		},
		{
			Name:  aws.String("Host"),
			Value: aws.String(stat.Host),
		},
	}
	metricDataInput := &cloudwatch.PutMetricDataInput{
		MetricData: []*cloudwatch.MetricDatum{
			{
				MetricName: aws.String("Clone time"),
				Timestamp:  aws.Time(r.start),
				Value:      aws.Float64(float64(stat.Clone)),
				Dimensions: metricDimensions,
				Unit:       aws.String("Milliseconds"),
			},
			{
				MetricName: aws.String("Compression time"),
				Timestamp:  aws.Time(r.start),
				Value:      aws.Float64(float64(stat.Zip)),
				Dimensions: metricDimensions,
				Unit:       aws.String("Milliseconds"),
			},
		},
		Namespace: aws.String(r.namespace),
	}
	_, err := r.cw.PutMetricData(metricDataInput)
	if err != nil {
		log.Fatal(err)
	}
}

func (r *Reporter) putCountAndTimeMetrics() {
	repositoryCount := float64(len(r.s.RepoStats))
	totalTime := float64(r.totalTime)
	metricDataInput := &cloudwatch.PutMetricDataInput{
		MetricData: []*cloudwatch.MetricDatum{
			{
				MetricName: aws.String("Repository count"),
				Timestamp:  aws.Time(r.start),
				Value:      aws.Float64(repositoryCount),
				Unit:       aws.String("Count"),
			}, {
				MetricName: aws.String("Total time"),
				Timestamp:  aws.Time(r.start),
				Value:      aws.Float64(totalTime),
				Unit:       aws.String("Milliseconds"),
			},
		},
		Namespace: aws.String(r.namespace),
	}
	_, err := r.cw.PutMetricData(metricDataInput)
	if err != nil {
		log.Fatal(err)
	}
}

func (r *Reporter) reportBackup() {
	err := r.tracker.CreateBackup(r.taskName)
	if err != nil {
		log.Fatal(err)
	}
}
