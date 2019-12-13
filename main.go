package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/shiftstack-dev-tools/prom-dashboard/frontend"
	"github.com/shiftstack-dev-tools/prom-dashboard/prometheus"
	"github.com/shiftstack-dev-tools/prom-dashboard/prow"
)

func main() {
	const (
		fsyncName         = "fsync"
		backendCommitName = "backend_commit"
		baseURL           = "https://gcsweb-ci.svc.ci.openshift.org/gcs/origin-ci-test/logs"
		jobName           = "release-openshift-ocp-installer-e2e-openstack-4.3"
	)

	app := frontend.NewApp()
	err := app.ValidateInput()
	req, err := app.ReadInput()
	if err != nil {
		log.Fatalf("%v", err)
	}

	// make prom-data dir
	promDir := filepath.Join(app.DataDir, "/promData")
	os.Mkdir(promDir, os.ModePerm)

	// Collect Data
	flattenedData := [][]string{}
	for _, id := range req.TestIDs {
		log.Printf("Preparing test %s", id)

		idDir := filepath.Join(promDir, "/"+id)
		os.Mkdir(idDir, os.ModePerm)
		if err != nil {
			log.Fatalf("couldnt create file: %v", err)
		}

		// Download Metrics
		data, err := prow.Metrics(baseURL, jobName, id, idDir)
		if err != nil {
			log.Fatalf("Failed to get metrics: %v", err)
		}

		// Untar prom file
		promData := filepath.Join(idDir, "/prometheus")
		os.Mkdir(promData, os.ModePerm)
		if err != nil {
			log.Fatalf("couldnt create file: %v", err)
		}

		tarfile := filepath.Join(idDir, "/prometheus.tar")

		// If the file is less than 50 Kb, it is definately a dud --> skip data collection
		// for reference, they are usually upwards of 50 Mb tar'd
		f, err := os.Stat(tarfile)
		if err != nil {
			log.Fatalln(err)
		}
		if f.Size() < 50000 {
			log.Printf("Prometheus data from job ID %s is either emtpy or corrupted. Skipping data collection...", id)

			// If no prom data, then just record the start and end time of job and move to next job
			data := []string{
				id,
				"",
				data.StartedAt.String(),
				data.FinishedAt.String(),
				req.Step,
			}
			flattenedData = append(flattenedData, data)
			continue
		}

		err = Untar(promData, tarfile)
		if err != nil {
			log.Fatalf("couldnt untar file: %v", err)
		}

		// CHMOD all files in untar'd prom dir to 777
		cmd := exec.Command("chmod", []string{
			"-R",
			"777",
			promData,
		}...)

		msg, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatalf("couldnt chmod prom data: %s: %v", msg, err)
		}

		// Stand up docker container
		port := "9090"
		hostpath, err := filepath.Abs(promData)
		if err != nil {
			log.Fatalln(err)
		}
		container, err := prometheus.Up(port, hostpath)
		if err != nil {
			log.Fatalf("failed to create docker container: %v", err)
		}

		for _, metric := range req.TimeSeries {
			query := prometheus.Query{
				BaseURL:    fmt.Sprintf("http://localhost:%s", port),
				MetricName: metric,
				QueryType:  prometheus.QueryTypeRange,
				Params: map[string]string{
					"query": fmt.Sprintf("histogram_quantile(0.99,rate(%s[%s]))", metric, req.Step),
					"step":  req.Step,
					"start": data.StartedAt.Format(time.RFC3339),
					"end":   data.FinishedAt.Format(time.RFC3339),
				},
			}

			res, err := query.GetData()
			if err != nil {
				log.Fatal(err)
			}

			vals, err := res.Flatten()
			if err != nil {
				log.Fatalf("Failed to flatten %s data: %v\n", query.MetricName, err)
			}
			for _, val := range vals {
				data := []string{
					id,
					metric,
					data.StartedAt.String(),
					data.FinishedAt.String(),
					req.Step,
				}
				data = append(data, val...)
				flattenedData = append(flattenedData, data)
			}

			log.Printf("%s gathered for test %s", metric, id)
		}
		err = prometheus.Down(container)
		if err != nil {
			log.Fatalln(err)
		}
	}

	// Write the CSV File
	resFile := filepath.Join(app.DataDir, "/results.csv")
	file, err := os.Create(resFile)
	if err != nil {
		log.Fatalf("Could not write file %s: %v", resFile, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, row := range flattenedData {
		err := writer.Write(row)
		if err != nil {
			log.Fatalf("Could not write file %s: %v", resFile, err)
		}
	}
}

// Untar takes a destination path and a reader; a tar reader loops over the tarfile
// creating the file structure at 'dst' along the way, and writing any files
// Source https://medium.com/@skdomino/taring-untaring-files-in-go-6b07cf56bc07
func Untar(dst, src string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {

		// if no more files are found return
		case err == io.EOF:
			return nil

		// return any other error
		case err != nil:
			return err

		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		target := filepath.Join(dst, header.Name)

		// the following switch could also be done using fi.Mode(), not sure if there
		// a benefit of using one vs. the other.
		// fi := header.FileInfo()

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			f.Close()
		}
	}
}
