package jolokia

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"regexp"

	"crypto/tls"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

var (
	keyRegExp = regexp.MustCompile("[^a-zA-Z0-9:_]")
)

// Exporter exports jolokia metrics for prometheus.
type Exporter struct {
	namespace         string
	URI               string
	mutex             sync.Mutex
	basicAuthUser     string
	basicAuthPassword string

	client   *http.Client
	up       *prometheus.Desc
	duration *prometheus.Desc
}

// NewExporter returns an initialized Exporter.
func NewExporter(namespace string, insecure bool, uri, basicAuthUser, basicAuthPassword string) *Exporter {
	return &Exporter{
		URI:               uri,
		namespace:         namespace,
		basicAuthUser:     basicAuthUser,
		basicAuthPassword: basicAuthPassword,
		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Could jolokia endpoint be reached",
			nil,
			nil),
		duration: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "response_duration"),
			"How long the jolokia endpoint took to deliver the metrics",
			nil,
			nil),
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
			},
		},
	}
}

// Describe describes all the metrics ever exported by the jolokia endpoint exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.up
	ch <- e.duration
}

// json data structure for jolokia endpoint
type jsonData map[string]float64

// Collect fetches the stats from configured location and delivers them
// as Prometheus metrics.
// It implements prometheus.Collector.
func (e *Exporter) collect(ch chan<- prometheus.Metric) error {
	req, err := http.NewRequest(http.MethodGet, e.URI, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.SetBasicAuth(e.basicAuthUser, e.basicAuthPassword)
	startTime := time.Now()

	resp, err := e.client.Do(req)
	ch <- prometheus.MustNewConstMetric(e.duration, prometheus.GaugeValue, time.Since(startTime).Seconds())

	if err != nil {
		ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0)
		return fmt.Errorf("error scraping jolokia endpoint: %v", err)
	}
	ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 1)

	defer resp.Body.Close()
	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	if resp.StatusCode != 200 {
		log.Errorf("There was an error, response code is %d, expected 200.", resp.StatusCode)
		return fmt.Errorf("status %s (%d)", resp.Status, resp.StatusCode)
	}

	var data jsonData
	if err := json.Unmarshal(body, &data); err != nil {
		log.Fatal(err)
	}

	log.Infof("Result has %d rows", len(data))

	for key, value := range data {
		snakeKey := keyToSnake(key)

		log.Infof("Adding key %s (originally %s) with value %v", snakeKey, key, value)

		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(e.namespace, "", snakeKey),
				key,
				nil,
				nil),
			prometheus.UntypedValue,
			value)
	}

	return nil
}

// converts any given key string to a prometheus acceptable key string
func keyToSnake(key string) string {
	return keyRegExp.ReplaceAllString(key, "_")
}

// Collects metrics, implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()
	if err := e.collect(ch); err != nil {
		log.Errorf("Error scraping jolokia endpoint: %s", err)
	}
	return
}
