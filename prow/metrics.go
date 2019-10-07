package prow

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type MetricsData struct {
	StartedAt  time.Time
	FinishedAt time.Time
	PromFile   string
}

// Metrics returns an io.ReadCloser that streams the tarball containing the
// Prometheus data. It is the caller responsibility to call Close on the
// returned MetricsData.
func Metrics(baseURL, jobName, jobID, tarpath string) (MetricsData, error) {
	var (
		m      MetricsData
		client http.Client
	)

	// Get start metadata
	{
		req, err := http.NewRequest("GET", baseURL+"/"+jobName+"/"+jobID+"/started.json", nil)
		if err != nil {
			return m, err
		}

		res, err := client.Do(req)
		if err != nil {
			return m, err
		}

		var started metadata
		if err := json.NewDecoder(res.Body).Decode(&started); err != nil {
			return m, err
		}

		m.StartedAt = started.time
	}

	// Get finish metadata
	{
		req, err := http.NewRequest("GET", baseURL+"/"+jobName+"/"+jobID+"/finished.json", nil)
		if err != nil {
			return m, err
		}

		res, err := client.Do(req)
		if err != nil {
			return m, err
		}

		var finished metadata
		if err := json.NewDecoder(res.Body).Decode(&finished); err != nil {
			return m, err
		}

		m.FinishedAt = finished.time
		// how to get success/fail
		//	m.Result = finished.result
	}

	// Get Tarball
	{
		m.PromFile = filepath.Join(tarpath, "/prometheus.tar")
		err := downloadFile(m.PromFile, baseURL+"/"+jobName+"/"+jobID+"/artifacts/e2e-openstack/metrics/prometheus.tar")
		if err != nil {
			return m, fmt.Errorf("Failed to downlad tarball: %v", err)
		}
	}

	return m, nil

}

func downloadFile(filepath string, url string) error {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
