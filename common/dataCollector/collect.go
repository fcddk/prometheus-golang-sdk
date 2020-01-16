package dataCollector

import (
	"fmt"
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

//采集源数据解析、打标签、整理之后生成对象，为之后prometheus格式化做准备
type MonitorMetric struct {
	//Category        string            `json:"category"`
	Name            string            `json:"name"`
	Help            string            `json:"help"`
	Type            string            `json:"type"`
	Dimensions      []string          `json:"dimensions"` //给该资源监控指标打上的维度标签（不能和ConstLabel冲突）
	DimensionsValue map[string]string `json:"dimensions_value"`
	ConstLabels     prometheus.Labels `json:"const_labels"` //注册到收集器的标签，定位一个样本资源
	Value           float64           `json:"value"`
}

//组织prometheus metric格式数据, 注册和采集一体（基于）
type resourceMetric struct {
	Type            string   //指标类型：guage or counter
	Dimensions      []string `json:"dimensions"` //给该资源监控指标打上的维度标签（不能和ConstLables）
	DimensionsValue map[string]string
	Value           float64
	MetricDesc      *prometheus.Desc //Prometheus注册指标时使用的描述对象
}

//new a domainMetric object
func newResourceMetric(metricData *MonitorMetric) *resourceMetric {
	dimensions := make([]string, 0)
	if len(metricData.Dimensions) != 0 {
		for _, value := range metricData.Dimensions {
			if value == "" {
				continue
			}
			dimensions = append(dimensions, value)
		}
	}
	return &resourceMetric{
		Type:            metricData.Type,
		DimensionsValue: metricData.DimensionsValue,
		Dimensions:      dimensions,
		Value:           metricData.Value,
		MetricDesc: prometheus.NewDesc(
			prometheus.BuildFQName("", "", metricData.Name),
			metricData.Help,
			dimensions,
			metricData.ConstLabels,
		),
	}
}

// Describe simply sends the Desc in the struct to the channel.
//描述监控指标，将指标描述放入channel, 等待采集器消费
func (c *resourceMetric) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.MetricDesc
}

//collect the metric
//采集监控指标数据
func (c *resourceMetric) Collect(ch chan<- prometheus.Metric) {
	defer prometheus.Unregister(c)
	switch c.Type {
	case string(GAUGE):
		metricResultDimensions, err := processAndVerifyDimensionsValue(c.Dimensions, c.DimensionsValue)
		if err != nil {
			log.Printf("failed to verify dimensions, error: %s", err.Error())
			return
		}

		ch <- prometheus.MustNewConstMetric(
			c.MetricDesc,
			prometheus.GaugeValue,
			float64(c.Value),
			metricResultDimensions...,
		)
	case string(COUNTER):
		metricResultDimensions, err := processAndVerifyDimensionsValue(c.Dimensions, c.DimensionsValue)
		if err != nil {
			log.Printf("failed to verify dimensions, error: %s", err.Error())
			return
		}

		ch <- prometheus.MustNewConstMetric(
			c.MetricDesc,
			prometheus.CounterValue,
			float64(c.Value),
			metricResultDimensions...,
		)
	}
}

//avoid invalid value and assembly dimension's target value
func processAndVerifyDimensionsValue(dimensions []string, sourceValue map[string]string) (result []string, err error) {
	if len(dimensions) == 0 {
		err = fmt.Errorf("the length of dimensons is 0")
		log.Printf("processAndVerifyDimensionsValue error: %s", err.Error())
		return result, err
	}

	for _, value := range dimensions {
		if value != "" {
			if mapValue, ok := sourceValue[value]; ok {
				result = append(result, mapValue)
			} else {
				err = fmt.Errorf("dimension[%s] value is nil, sourceMap:%v", value, sourceValue)
				return
			}
		}
	}

	return
}

func CollectMetricData(metric *MonitorMetric) (err error) {
	metricObj := newResourceMetric(metric)
	err = prometheus.Register(metricObj)
	if err != nil {
		log.Printf("register a domainMetric[metric:%v] error: %s", metricObj, err.Error())
		return
	}
	return nil
}
