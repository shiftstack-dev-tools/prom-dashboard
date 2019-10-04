package prow

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// const baseURL = "https://gcsweb-ci.svc.ci.openshift.org/gcs/origin-ci-test/logs/"

func TestMetrics(t *testing.T) {
	baseURL := "/release-openshift-ocp-installer-e2e-openstack-4.2/16"
	tarball := []byte("This text represents the binary tarball.")
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		log.Println("path:", req.URL.Path)

		switch req.URL.Path {

		case baseURL + "/started.json":
			rw.Write([]byte(`{"timestamp":1569934191,"repos":{"/":""}}`))

		case baseURL + "/finished.json":
			rw.Write([]byte(`{"timestamp":1569939439,"passed":false,"metadata":{"infra-commit":"","job-version":"","pod":"4.2.0-0.nightly-2019-10-01-124419-openstack","repo":"/","repo-commit":"","repos":{"/":""},"work-namespace":"ci-op-i281gs2x"},"result":"FAILURE"}`))

		case baseURL + "/artifacts/e2e-openstack/metrics/prometheus.tar":
			rw.Write(tarball)

		default:
			rw.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	defer ts.Close()

	data, err := Metrics(ts.URL, "release-openshift-ocp-installer-e2e-openstack-4.2", "16")
	if err != nil {
		t.Fatalf("while fetching the data: %v", err)
	}

	if want := "FAILURE"; want != data.Result {
		t.Errorf("expected result to be %q, found %q", want, data.Result)
	}

	expectedTime, err := time.Parse(time.RFC3339, "2019-10-01T14:17:19Z")
	if err != nil {
		panic(err)
	}

	if have := data.FinishedAt; expectedTime != have {
		t.Errorf("expected finish time to be %q, found %q", expectedTime, have)
	}
}
