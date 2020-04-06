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

func getJson(url string, target interface{}) error {
	client := &http.Client{Timeout: 2 * time.Second}
	r, err := client.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func recordMetrics(url string, sleep time.Duration) {
	go func(url string, sleep time.Duration) {
		for {
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

			time.Sleep(sleep * time.Second)
		}
	}(url, sleep)
}

var (
	solrUrl        string
	scrapeInterval int
	port           int
	solrMemoryFree = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "solr_memory_free",
	})
	solrMemoryTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "solr_memory_total",
	})
	solrMemoryMax = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "solr_memory_max",
	})
	solrMemoryUsed = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "solr_memory_used",
	})
	solrScrapeErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "solr_scrape_errors",
	})
)

func init() {
	flag.StringVar(&solrUrl, "solr-url", "http://localhost:8983/solr/admin/info/system?wt=json", "SORL info URL")
	flag.IntVar(&scrapeInterval, "scrape-interval", 2, "Scrape interval")
	flag.IntVar(&port, "port", 2112, "HTTP server port")
	flag.Parse()
}

func main() {
	recordMetrics(solrUrl, time.Duration(scrapeInterval))

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
