package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dev/prom-dashboard/data"
)

func main() {
	baseURL := "http://172.24.160.10:9090"
	queries := []*data.Query{}

	// Set up fsync query
	fsyncParams := make(map[string]string)
	fsyncParams["query"] = "histogram_quantile(0.99, rate(etcd_disk_wal_fsync_duration_seconds_bucket[5m]))"
	fsyncParams["start"] = "2019-09-23T18:18:00Z"
	fsyncParams["end"] = "2019-09-24T18:18:00Z"
	fsyncParams["step"] = "5m"

	fsync := new(data.Query)
	fsync.Name = "fsync"
	fsync.BaseURL = baseURL
	fsync.QueryType = "range"
	fsync.Params = fsyncParams

	queries = append(queries, fsync)

	// Execute the Queries
	for _, query := range queries {
		log.Printf("[%s]: Querying %s data\n", time.Now().Format(time.RFC3339), query.Name)
		result, err := data.GetData(query)
		if err != nil {
			log.Fatalf("[%s]: Failed to execute %s query: %v\n", time.Now().Format(time.RFC3339), query.Name, err)
		}
		fmt.Printf("%v", result)
	}
}
