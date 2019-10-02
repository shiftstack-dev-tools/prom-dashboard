package prometheus

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	httpRequestTimeout = 5 * time.Second

	//QueryTypeRange is a constant string used to identify ranged queries
	QueryTypeRange = "range"
)

// Query is a generic way to build a prometheus query
type Query struct {
	Name      string
	BaseURL   string
	QueryType string
	Params    map[string]string
}

// GetData makes a GET query against prometheus and returns data
func (query *Query) GetData() (*RangeResult, error) {
	if query == nil {
		log.Fatal("query parameter can not be nil")
	}
	// Execute Range Query
	if query.QueryType == QueryTypeRange {
		result, err := rangeQuery(query.BaseURL, &query.Params)
		if err != nil {
			return nil, fmt.Errorf("range query error: %v", err)
		}
		return result, nil
	}

	// TODO(egarcia): implement proper http error handling
	// TODO(egarcia): exponential backoff retry

	return nil, nil
}

func rangeQuery(baseURL string, params *map[string]string) (*RangeResult, error) {
	if params == nil || len(*params) <= 0 {
		return nil, fmt.Errorf("nil or empty rangeQuery params")
	}

	// Build Query
	query := baseURL + "/api/v1/query_range"
	count := 0
	for key, value := range *params {
		if count == 0 {
			query = query + "?" + key + "=" + value
		} else {
			query = query + "&" + key + "=" + value
		}
		count++
	}

	// Fetch time series Prometheus data
	client := http.Client{Timeout: httpRequestTimeout}
	res, err := client.Get(query)
	if err != nil {
		return nil, fmt.Errorf("http GET error: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error %d: %s", res.StatusCode, res.Status)
	}
	decoder := json.NewDecoder(res.Body)
	var result RangeResult
	err = decoder.Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
