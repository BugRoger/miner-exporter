package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	namespace = "miner"
)

type Metrics struct {
	Name    string
	Version string
	Uptime  float64
	Stats   []Stats
	Temps   []float64
	Fans    []float64
	Pools   []string
}

type Stats struct {
	Coin      string
	TotalRate float64
	Accepted  float64
	Rejected  float64
	GPURates  []float64
}

type Exporter struct {
	miner        Miner
	up           *prometheus.Desc
	uptime       *prometheus.Desc
	info         *prometheus.Desc
	rates        *prometheus.Desc
	ratesTotal   *prometheus.Desc
	shares       *prometheus.Desc
	temperatures *prometheus.Desc
	fans         *prometheus.Desc
}

type Miner interface {
	Collect() (*Metrics, error)
}

func NewExporter() *Exporter {
	return &Exporter{
		miner: NewClaymoreDualMinerClient("tcp", "localhost:3333"),
		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Could the miner be reached.",
			nil,
			nil,
		),
		uptime: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "uptime"),
			"Number of seconds since the miner started.",
			nil,
			nil,
		),
		info: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "info"),
			"Information about this miner",
			[]string{"name", "version"},
			nil,
		),
		temperatures: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "gpu", "temperatures"),
			"Temperatures for each GPU",
			[]string{"gpu"},
			nil,
		),
		fans: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "gpu", "fans"),
			"Fan Speed for each GPU",
			[]string{"gpu"},
			nil,
		),
		rates: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "rates"),
			"Mining rate by Coin and GPU",
			[]string{"coin", "gpu"},
			nil,
		),
		shares: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "shares"),
			"Shares by Coin and Status",
			[]string{"coin", "status"},
			nil,
		),
		ratesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "rates", "total"),
			"Mining rate by Coin and GPU",
			[]string{"coin"},
			nil,
		),
	}
}

func main() {
	var (
		listenAddress = flag.String("web.listen-address", ":9278", "Address to listen on for web interface and telemetry.")
		metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	)
	flag.Parse()

	prometheus.MustRegister(NewExporter())
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Miner Exporter</title></head>
             <body>
             <h1>Miner Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})
	fmt.Println("Starting HTTP server on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.up
	ch <- e.uptime
	ch <- e.info
	ch <- e.temperatures
	ch <- e.fans
	ch <- e.rates
	ch <- e.ratesTotal
	ch <- e.shares
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	data, err := e.miner.Collect()
	if err != nil {
		ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0)
		log.Printf("Failed to collect stats from miner: %s\n", err)
		return
	}

	fmt.Printf("%+v\n", data)

	ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 1)
	ch <- prometheus.MustNewConstMetric(e.info, prometheus.GaugeValue, 1, data.Name, data.Version)
	ch <- prometheus.MustNewConstMetric(e.uptime, prometheus.CounterValue, data.Uptime)

	for gpu, temp := range data.Temps {
		ch <- prometheus.MustNewConstMetric(e.temperatures, prometheus.GaugeValue, temp, fmt.Sprintf("%v", gpu))
	}

	for gpu, speed := range data.Fans {
		ch <- prometheus.MustNewConstMetric(e.fans, prometheus.GaugeValue, speed, fmt.Sprintf("%v", gpu))
	}

	for _, stat := range data.Stats {
		for gpu, r := range stat.GPURates {
			ch <- prometheus.MustNewConstMetric(e.rates, prometheus.GaugeValue, r, stat.Coin, fmt.Sprintf("%v", gpu))
		}

		ch <- prometheus.MustNewConstMetric(e.ratesTotal, prometheus.GaugeValue, stat.TotalRate, stat.Coin)
		ch <- prometheus.MustNewConstMetric(e.shares, prometheus.GaugeValue, stat.Accepted, stat.Coin, "accepted")
		ch <- prometheus.MustNewConstMetric(e.shares, prometheus.GaugeValue, stat.Rejected, stat.Coin, "rejected")
	}
}
