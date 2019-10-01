package data

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	timeout = 15
)

// Query is a generic way to build a prometheus query
type Query struct {
	Name      string
	BaseURL   string
	QueryType string
	Params    map[string]string
}

// GetData makes a GET query against prometheus and returns data
func GetData(query *Query) (*RangeResult, error) {
	if query == nil {
		log.Fatal("query parameter can not be nil")
	}
	// Execute Range Query
	if query.QueryType == "range" {
		result, err := rangeQuery(query.BaseURL, &query.Params)
		if err != nil {
			return nil, fmt.Errorf("range query %s error: %v", query.Name, err)
		}
		return result, err
	}

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
	}

	// Execute Query
	res, err := http.Get(query)
	if err != nil {
		return nil, err
	}

	// Convert Results to Struct
	rawResult, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}

	var result RangeResult
	err = json.Unmarshal([]byte(rawResult), &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
