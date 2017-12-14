package spring

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"net/http"
	"net/http/httptest"

	"path"

	"bytes"

	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)

func authTestHandler(w http.ResponseWriter, r *http.Request) {
	if u, p, ok := r.BasicAuth(); !ok {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Unauthorized")
		return
	} else if u != "admin" || p != "secret" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Wrong credentials")
		return
	}

	w.Header().Add("Content-Type", "application/json")
	http.ServeFile(w, r, path.Join("fixtures", "response.json"))
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	http.ServeFile(w, r, path.Join("fixtures", "response.json"))
}

func getPromResponse(t *testing.T) string {
	file, err := ioutil.ReadFile(filepath.Join("fixtures", "metrics.txt"))

	if err != nil {
		t.Fatalf("Unexpected exception reading file: %v", err)
	}

	return string(file)
}

func TestExporter_Describe(t *testing.T) {
	exp := NewExporter(log.Base(), Namespace, false, "http://test/test", "", "")
	c := make(chan *prometheus.Desc, 1024)

	exp.Describe(c)

	if len(c) != 2 {
		t.Fatalf("Expected channel to have 2 objects, got %d", len(c))
	}

	up := <-c
	if up.String() != "Desc{fqName: \"spring_up\", help: \"Could spring endpoint be reached\", constLabels: {}, variableLabels: []}" {
		t.Errorf("Unexpected up metric description: %s", up.String())
	}

	duration := <-c
	if duration.String() != "Desc{fqName: \"spring_response_duration\", help: \"How long the spring endpoint took to deliver the metrics\", constLabels: {}, variableLabels: []}" {
		t.Errorf("Unexpected duration metric description: %s", duration.String())
	}
}

func TestExporter_Collect_NoAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(testHandler))

	buf := bytes.NewBufferString("")
	logger := log.NewLogger(buf)
	logger.SetLevel("warn")
	exp := NewExporter(logger, Namespace, false, srv.URL, "", "")
	c := make(chan prometheus.Metric, 1024)

	exp.Collect(c)

	bufStr := buf.String()
	if len(bufStr) != 0 {
		t.Fatalf("unexpect collect output: %v", bufStr)
	}

	if len(c) != 40 {
		t.Fatalf("Expected channel to have 38 objects, got %d", len(c))
	}
}

func TestExporter_Collect_WithAuth(t *testing.T) {
	buf := bytes.NewBufferString("")
	logger := log.NewLogger(buf)
	logger.SetLevel("warn")

	srv := httptest.NewServer(http.HandlerFunc(authTestHandler))
	exp := NewExporter(logger, Namespace, false, srv.URL, "admin", "secret")
	c := make(chan prometheus.Metric, 1024)

	exp.Collect(c)

	bufStr := buf.String()
	if len(bufStr) != 0 {
		t.Fatalf("unexpect collect output: %v", bufStr)
	}

	if len(c) != 40 {
		t.Fatalf("Expected channel to have 38 objects, got %d", len(c))
	}
}

func TestExporter_Collect_WithAuthButNoneGiven(t *testing.T) {
	buf := bytes.NewBufferString("")
	logger := log.NewLogger(buf)
	logger.SetLevel("warn")

	srv := httptest.NewServer(http.HandlerFunc(authTestHandler))
	exp := NewExporter(logger, Namespace, false, srv.URL, "", "")
	c := make(chan prometheus.Metric, 1024)

	exp.Collect(c)

	bufStr := buf.String()
	if ! strings.Contains(bufStr, "Error scraping spring endpoint: there was an error, response code is 401, expected 200") {
		t.Fatalf("unexpect collect output: %v", bufStr)
	}
}

func TestExporter_Collect_WithAuthButWrongGiven(t *testing.T) {
	buf := bytes.NewBufferString("")
	logger := log.NewLogger(buf)
	logger.SetLevel("warn")

	srv := httptest.NewServer(http.HandlerFunc(authTestHandler))
	exp := NewExporter(logger, Namespace, false, srv.URL, "", "")
	c := make(chan prometheus.Metric, 1024)

	exp.Collect(c)

	bufStr := buf.String()
	if ! strings.Contains(bufStr, "Error scraping spring endpoint: there was an error, response code is 401, expected 200") {
		t.Fatalf("unexpect collect output: %v", bufStr)
	}
}

func TestExporter_Collect_WithPrometheus(t *testing.T) {
	buf := bytes.NewBufferString("")
	logger := log.NewLogger(buf)
	logger.SetLevel("warn")

	fixtureSrv := httptest.NewServer(http.HandlerFunc(authTestHandler))
	exp := NewExporter(logger, Namespace, false, fixtureSrv.URL, "admin", "secret")

	prometheus.MustRegister(exp)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rw := httptest.NewRecorder()
	promhttp.Handler().ServeHTTP(rw, req)

	if rw.Code != 200 {
		t.Errorf("expected status code to be %d, got %d", 200, rw.Code)
	}

	if rw.Body == nil {
		t.Fatal("Response does not have a body")
	}

	bufStr := buf.String()
	if len(bufStr) != 0 {
		t.Fatalf("unexpect collect output: %v", bufStr)
	}

	resBody := rw.Body.String()
	expectedBody := getPromResponse(t)

	if strings.Contains(resBody, expectedBody) {
		t.Errorf("expected body to contain metrics, but doesn't: %s.", resBody)
	}
}
