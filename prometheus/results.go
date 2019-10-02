package prometheus

import (
	"fmt"
	"strings"
)

// RangeResult holds the raw prometheus data from a range query
type RangeResult struct {
	Success string     `json:"success"`
	Data    resultList `json:"data"`
}

type resultList struct {
	ResultType string   `json:"resultType"`
	Result     []result `json:"result"`
}

// result represents the result of the query for each pod
// it was measuring
type result struct {
	Metric metric `json:"metric"`
	Values [][]interface{}
}

type metric struct {
	Endpoint  string `json:"endpoint"`
	Instance  string `json:"instance"`
	Job       string `json:"job"`
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
	Service   string `json:"service"`
}

// Flatten creates a csv like slice of slices to hold the essential
// data from a RangeResult struct
// TODO(egarcia): move this to the main control loop to optimize
func (rr *RangeResult) Flatten() ([][]string, error) {
	if rr == nil {
		return nil, fmt.Errorf("RangeResult Pointer is nil")
	}
	result := [][]string{}

	for _, pod := range rr.Data.Result {
		podData := []string{}
		node, err := getNode(pod.Metric.Pod)
		if err != nil {
			return nil, err
		}
		podData = append(podData, node)
		for _, value := range pod.Values {
			data := fmt.Sprintf("%v", []interface{}(value)[1])
			podData = append(podData, data)
		}
		result = append(result, podData)
	}
	return result, nil
}

// Helper function to clean up main body
// Gets node name "master-#" "worker-#"
func getNode(pod string) (string, error) {
	nodeIndex := strings.Index(pod, "master")
	if nodeIndex == -1 {
		nodeIndex = strings.Index(pod, "worker")
		if nodeIndex == -1 {
			return "", fmt.Errorf("could not figure which node pod ran on")
		}
	}
	return pod[nodeIndex:], nil
}
