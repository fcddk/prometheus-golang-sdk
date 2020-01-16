package main

import (
	"flag"
	"log"
	"net/http"
	"prometheus-golang-sdk/common/dataCollector"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	var (
		listenAddress = flag.String("web.listen-address", ":9000", "Address to listen on for web interface and telemetry.")
		metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	)
	flag.Parse()
	go func() {
		prometheus.Unregister(prometheus.NewBuildInfoCollector())
		prometheus.Unregister(prometheus.NewGoCollector())
	}()

	demoDimesions := make([]string, 0)
	demoDimensionsValue := make(map[string]string)
	demoDimesions = append(demoDimesions, "resource_name")
	demoDimensionsValue["resource_name"] = "host001"
	demoMetricConstLabels := make(map[string]string)
	demoMetricConstLabels["resource_id"] = "uuuus-ssddas-11100-ssk"
	//cpu_busy 在用
	cpuBusy := &dataCollector.MonitorMetric{}
	cpuBusy.Name = "cpu_busy"
	cpuBusy.Help = ""
	cpuBusy.Type = "gauge"
	cpuBusy.Value = float64(10)
	cpuBusy.Dimensions = demoDimesions
	cpuBusy.DimensionsValue = demoDimensionsValue
	cpuBusy.ConstLabels = demoMetricConstLabels
	_ = dataCollector.CollectMetricData(cpuBusy)

	http.Handle(*metricsPath, promhttp.Handler())
	log.Fatal(http.ListenAndServe(*listenAddress, nil))

}
