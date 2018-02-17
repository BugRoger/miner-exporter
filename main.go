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
	Version    string
	Uptime     float64
	Algorithms []Algorithm
}

type Algorithm struct {
	Name   string
	Shares Shares
	Rates  Rates
}

type Shares struct {
	Accepted float64
	Rejected float64
	Stale    float64
}

type Rates struct {
	Total float64
	ByGPU []float64
}

type Exporter struct {
	miner      Miner
	up         *prometheus.Desc
	uptime     *prometheus.Desc
	info       *prometheus.Desc
	rates      *prometheus.Desc
	ratesTotal *prometheus.Desc
	shares     *prometheus.Desc
}

type Miner interface {
	Name() string
	Collect() (*Metrics, error)
}

func NewExporter(miner Miner) *Exporter {
	return &Exporter{
		miner: miner,
		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Could the miner be reached.",
			nil,
			prometheus.Labels{"name": miner.Name()},
		),
		uptime: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "uptime"),
			"Number of seconds since the miner started.",
			nil,
			prometheus.Labels{"name": miner.Name()},
		),
		info: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "info"),
			"Information about this miner",
			[]string{"version"},
			prometheus.Labels{"name": miner.Name()},
		),
		rates: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "rates"),
			"Mining rate by Algorithm and GPU",
			[]string{"algorithm", "gpu"},
			prometheus.Labels{"name": miner.Name()},
		),
		shares: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "shares"),
			"Shares by Algorithm and Status",
			[]string{"algorithm", "status"},
			prometheus.Labels{"name": miner.Name()},
		),
		ratesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "rates", "total"),
			"Mining rate total by algorithm",
			[]string{"algorithm"},
			prometheus.Labels{"name": miner.Name()},
		),
	}
}

func main() {
	var (
		listenAddress = flag.String("web.listen-address", ":9278", "Address to listen on for web interface and telemetry.")
		metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
		ccminerFlag   = flag.String("ccminer", "", "Enable and read CCMiner metrics from this address")
		cdmFlag       = flag.String("claymoredualminer", "", "Enable and read Claymore Dual Miner metrics from this address")
		dstmFlag      = flag.String("dstm", "", "Enable and read DSTM metrics from this address")
	)
	flag.Parse()

	if *ccminerFlag != "" {
		prometheus.MustRegister(NewExporter(NewCCMinerClient(*ccminerFlag)))
	}

	if *cdmFlag != "" {
		prometheus.MustRegister(NewExporter(NewClaymoreDualMinerClient("tcp", *cdmFlag)))
	}

	if *dstmFlag != "" {
		prometheus.MustRegister(NewExporter(NewDSTMClient(*dstmFlag)))
	}

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
	ch <- prometheus.MustNewConstMetric(e.info, prometheus.GaugeValue, 1, data.Version)
	ch <- prometheus.MustNewConstMetric(e.uptime, prometheus.CounterValue, data.Uptime)

	for _, algo := range data.Algorithms {
		for gpu, r := range algo.Rates.ByGPU {
			ch <- prometheus.MustNewConstMetric(e.rates, prometheus.GaugeValue, r, algo.Name, fmt.Sprintf("%v", gpu))
		}

		ch <- prometheus.MustNewConstMetric(e.ratesTotal, prometheus.GaugeValue, algo.Rates.Total, algo.Name)
		ch <- prometheus.MustNewConstMetric(e.shares, prometheus.GaugeValue, algo.Shares.Accepted, algo.Name, "accepted")
		ch <- prometheus.MustNewConstMetric(e.shares, prometheus.GaugeValue, algo.Shares.Rejected, algo.Name, "rejected")
		ch <- prometheus.MustNewConstMetric(e.shares, prometheus.GaugeValue, algo.Shares.Stale, algo.Name, "stale")
	}
}
