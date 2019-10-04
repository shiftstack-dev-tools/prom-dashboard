package main

import (
	"encoding/csv"
	"log"
	"os"
	"strconv"

	"github.com/shiftstack-dev-tools/prom-dashboard/frontend"
	"github.com/shiftstack-dev-tools/prom-dashboard/prometheus"
)

func main() {
	const (
		fsyncName         = "fsync"
		backendCommitName = "backend_commit"
	)

	app := frontend.NewApp()
	req, err := app.ReadInput()
	if err != nil {
		log.Fatalf("%v", err)
	}

	baseURL := "http://172.24.160.10:9090"
	queries := []*prometheus.Query{}
	startTime := "2019-09-23T18:00:00Z"
	stopTime := "2019-09-24T18:00:00Z"

	// Set up fsync query
	fsyncParams := make(map[string]string)
	fsyncParams["query"] = "histogram_quantile(0.99,rate(etcd_disk_wal_fsync_duration_seconds_bucket[5m]))"

	fsync := prometheus.Query{
		MetricName: fsyncName,
		QueryType:  prometheus.QueryTypeRange,
		Params:     fsyncParams,
	}

	queries = append(queries, &fsync)

	// Set up backend commit query
	bcParams := make(map[string]string)
	bcParams["query"] = "histogram_quantile(0.99,rate(etcd_disk_backend_commit_duration_seconds_bucket[5m]))"

	backendCommit := prometheus.Query{
		MetricName: backendCommitName,
		QueryType:  prometheus.QueryTypeRange,
		Params:     bcParams,
	}

	queries = append(queries, &backendCommit)

	// Set Up CSV
	flattenedData := [][]string{}

	// Execute the Queries
	for uuid, query := range queries {
		log.Printf("Querying %s data\n", query.MetricName)

		// TODO(egarcia): make queries to the same prom asynchronous
		query.BaseURL = baseURL
		query.Params["start"] = startTime
		query.Params["end"] = stopTime
		query.Params["step"] = req.Step
		result, err := query.GetData()
		if err != nil {
			log.Fatalf("Failed to execute %s query: %v\n", query.MetricName, err)
		}

		vals, err := result.Flatten()
		if err != nil {
			log.Fatalf("Failed to flatten %s data: %v\n", query.MetricName, err)
		}
		for _, val := range vals {
			data := []string{
				strconv.Itoa(uuid),
				query.TestID,
				query.QueryType,
			}
			data = append(data, val...)
			flattenedData = append(flattenedData, data)
		}
	}

	// Write the CSV File
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Could not write file %s: %v", filename, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, row := range flattenedData {
		err := writer.Write(row)
		if err != nil {
			log.Fatalf("Could not write file %s: %v", filename, err)
		}
	}
}
