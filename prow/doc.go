// Package prow downloads Prometheus data for a given CI
// run, alongside with some metadata.
//
// The function Metrics returns the start time, the finish time, the status of
// the build and the metrics archive.
//
//	metrics, err := prow.Metrics(baseURL, jobName, "330")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer metrics.Close()
//
//	log.Printf("Job started  at %s", metrics.StartedAt)
//	log.Printf("Job finished at %s", metrics.FinishedAt)
//
//	f, err := os.Create("prom.tar")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer f.Close()
//
//	n, err := io.Copy(f, metrics)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Printf("Copied %d bytes.", n)
package prow
