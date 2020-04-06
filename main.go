package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type SolrInfo struct {
	Jvm struct {
		Memory struct {
			Raw struct {
				Free  int64 `json:"free"`
				Total int64 `json:"total"`
				Max   int64 `json:"max"`
				Used  int64 `json:"used"`
			} `json:"raw"`
		} `json:"memory"`
	}
}

var (
	solrUrl        string
	port           int
	solrMemoryFree = promauto.NewGauge(prometheus.GaugeOpts{
		Help: "SOLR free memory",
		Name: "solr_memory_free",
	})
	solrMemoryTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Help: "SOLR total memory",
		Name: "solr_memory_total",
	})
	solrMemoryMax = promauto.NewGauge(prometheus.GaugeOpts{
		Help: "SOLR max memory",
		Name: "solr_memory_max",
	})
	solrMemoryUsed = promauto.NewGauge(prometheus.GaugeOpts{
		Help: "SOLR used memory",
		Name: "solr_memory_used",
	})
	solrScrapeErrors = promauto.NewCounter(prometheus.CounterOpts{
		Help: "SOLR scraping errors count",
		Name: "solr_scrape_errors",
	})
)

func getJson(url string, target interface{}) error {
	client := &http.Client{Timeout: 2 * time.Second}
	r, err := client.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func wrapMetricsHandler(h http.Handler, url string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var info = SolrInfo{}
		err := getJson(url, &info)
		if err != nil {
			log.Println(err)
			solrScrapeErrors.Inc()
		}

		solrMemoryFree.Set(float64(info.Jvm.Memory.Raw.Free))
		solrMemoryTotal.Set(float64(info.Jvm.Memory.Raw.Total))
		solrMemoryMax.Set(float64(info.Jvm.Memory.Raw.Max))
		solrMemoryUsed.Set(float64(info.Jvm.Memory.Raw.Used))

		h.ServeHTTP(w, r)
	})
}

func init() {
	flag.StringVar(&solrUrl, "solr-url", "http://localhost:8983/solr/admin/info/system?wt=json", "SORL info URL")
	flag.IntVar(&port, "port", 2112, "HTTP server port")
	flag.Parse()
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`<html>
<head><title>SOLR Exporter</title></head>
<body>
<h1>SOLR Exporter</h1>
<p><a href="/metrics">Metrics</a></p>
</body>
</html>`))
		if err != nil {
			log.Println(err)
		}
	})
	http.Handle("/metrics", wrapMetricsHandler(promhttp.Handler(), solrUrl))
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
