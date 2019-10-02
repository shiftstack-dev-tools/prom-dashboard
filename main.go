package main

import (
	"log"

	"github.com/dev/prom-dashboard/prometheus"
)

func main() {
	const (
		step              = "1m"
		fsyncName         = "fsync"
		backendCommitName = "backend_commit"
	)

	baseURL := "http://172.24.160.10:9090"
	queries := []*prometheus.Query{}
	startTime := "2019-09-23T18:18:00Z"
	stopTime := "2019-09-24T18:18:00Z"

	// Set up fsync query
	fsyncParams := make(map[string]string)
	fsyncParams["query"] = "histogram_quantile(0.99,rate(etcd_disk_wal_fsync_duration_seconds_bucket[5m]))"

	fsync := prometheus.Query{
		Name:      fsyncName,
		QueryType: prometheus.QueryTypeRange,
		Params:    fsyncParams,
	}

	queries = append(queries, &fsync)

	// Set up backend commit query
	bcParams := make(map[string]string)
	bcParams["query"] = "histogram_quantile(0.99,rate(etcd_disk_backend_commit_duration_seconds_bucket[5m]))"

	backendCommit := prometheus.Query{
		Name:      backendCommitName,
		QueryType: prometheus.QueryTypeRange,
		Params:    bcParams,
	}

	queries = append(queries, &backendCommit)

	// Execute the Queries
	for _, query := range queries {
		// TODO(egarcia): make calls asynchronous
		log.Printf("Querying %s data\n", query.Name)
		query.BaseURL = baseURL
		query.Params["start"] = startTime
		query.Params["end"] = stopTime
		query.Params["step"] = step
		result, err := query.GetData()
		if err != nil {
			log.Fatalf("Failed to execute %s query: %v\n", query.Name, err)
		}

		log.Println(result)
	}

}
