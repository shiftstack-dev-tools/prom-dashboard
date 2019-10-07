package frontend

import (
	"fmt"
	"regexp"
)

// DataRequest stores user data requests from CI prom
type DataRequest struct {
	// Step allows you to set the step for ranged queries
	// +optional: default: "1m"
	Step string `yaml:"step,omitempty"`

	// TimeSeries allows you to specify which time series metrics you want to gather
	// These will get translated to range queries

	// +optional: default: [
	// "etcd_disk_wal_fsync_duration_seconds_bucket",
	// "etcd_disk_backend_commit_duration_seconds_bucket",
	// "etcd_network_peer_round_trip_time_seconds_bucket"]
	TimeSeries []string `yaml:"promMetrics,omitempty"`

	// TestIDs holds the UUID of the CI tests you want to pull data from
	TestIDs []string `yaml:"testIDs"`
}

// NewDataRequest object and set default values
func NewDataRequest() *DataRequest {
	// Set Defaults
	req := DataRequest{
		Step: "1m",
		TimeSeries: []string{
			"etcd_disk_wal_fsync_duration_seconds_bucket",
			"etcd_disk_backend_commit_duration_seconds_bucket",
			"etcd_network_peer_round_trip_time_seconds_bucket",
		},
	}

	return &req
}

// Validate DataRequest Objects
func (req *DataRequest) Validate() error {
	errors := []string{}
	if req == nil {
		return fmt.Errorf("nil DataRequest object")
	}
	if req.TestIDs == nil || len(req.TestIDs) == 0 {
		errors = append(errors, "You must specify at least 1 Test ID to gather data from")
	}
	if req.Step != "" {
		ok, err := regexp.MatchString("^\\d+\\w$", req.Step)
		if err != nil {
			return fmt.Errorf("Regex Error: %v", err)
		}
		if !ok {
			errors = append(errors, "Invalid Step: Some examples of valid steps are `1m`, `30s`")
		}
	}

	if len(errors) > 0 {
		allErrs := "Your config had the following errors:\n"
		for _, err := range errors {
			allErrs += fmt.Sprintf("\t%s\n", err)
		}
		return fmt.Errorf(allErrs)
	}

	return nil
}
