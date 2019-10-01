package data

// RangeResult holds the prometheus data from a range query
type RangeResult struct {
	Success string     `json:success`
	Data    resultList `json:data`
}

type resultList struct {
	ResultType string   `json:resultType`
	Result     []result `json: result`
}

type result struct {
	Metric metric      `json:metric`
	Values [][]float64 `json:values`
}

type metric struct {
	Endpoint  string `json:endpoint`
	Instance  string `json:instance`
	Job       string `json:job`
	Namespace string `json:namespace`
	Pod       string `json:pod`
	Service   string `json:service`
}
