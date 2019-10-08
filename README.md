# prom-dashboard

This tool can be uses to fetch prometheus metrics from a list of specified CI runs in PROW, and outputs the data in a csv.

## Usage

Running it is simple, it takes a yaml file and a directory as input. The directory should be empty.
```sh
go run main.go -c <yaml config> -o <metadata and output dir>
```

The yaml supports the following customizations:

```yaml
// testIDs lists the ID of the CI Prow jobs you want to pull data from
testIDs:
    - 42
    - 41
    - 40

// promMetrics lists the prometheus metrics that you want to gather for every test in testIDs
// +optional, defaults to: [
//      "etcd_disk_wal_fsync_duration_seconds_bucket",
//      "etcd_disk_backend_commit_duration_seconds_bucket",
//      "etcd_network_peer_round_trip_time_seconds_bucket",
//  ]
promMetrics:
    - etcd_disk_backend_commit_duration_seconds_bucket

// Step allows you to set the step for ranged queries
// +optional: default: "1m"
step: 5m
```

## Output
The final output of a run will be written to `output-dir/results.csv`. This csv file has the following schema:

| TestID | Metric | Node | Time Series Data |
| ---    | ---    | ---  | ---              |

The Time series data is in time differentials based on the `step` you provided. So the first cell is 0, and the second is +`step`.