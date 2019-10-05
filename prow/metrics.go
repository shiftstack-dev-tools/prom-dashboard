package prow

import (
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type MetricsData struct {
	StartedAt  time.Time
	FinishedAt time.Time
	Result     string

	io.ReadCloser
}

// Metrics returns an io.ReadCloser that streams the tarball containing the
// Prometheus data. It is the caller responsibility to call Close on the
// returned MetricsData.
func Metrics(baseURL, jobName, jobID string) (MetricsData, error) {
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
		m.Result = finished.result
	}

	// Prepare the Prometheus data
	//TODO: can we generalise "e2e-openstack"?
	req, err := http.NewRequest("GET", baseURL+"/"+jobName+"/"+jobID+"/artifacts/e2e-openstack/metrics/prometheus.tar", nil)
	if err != nil {
		return m, err
	}

	res, err := client.Do(req)

	m.ReadCloser = res.Body

	return m, err
}
