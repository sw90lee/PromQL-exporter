package exporter

import (
	"crypto/tls"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"io"
	"kt.com/p5g/cnf-exporter/samsung-cpc/cfg"
	"kt.com/p5g/cnf-exporter/samsung-cpc/logger"
	"kt.com/p5g/cnf-exporter/samsung-cpc/pkg/k8sClient"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Collector metric groups
type Collector struct {
	Collects []Collect
}

// Collect collect structure
type Collect struct {
	Metrics Metrics
}

// Metrics metric map
type Metrics map[string]*Metric

// Metric metric map
type Metric struct {
	Name        string
	Prefix      string
	Type        string
	Description string
	Url         string
	Labels      []string
	MetricDesc  *prometheus.Desc
}

type DeviceCollector struct {
	Collects   []Collect
	StatusDesc *prometheus.Desc
}

// Describe prometheus describe
func (c *DeviceCollector) Describe(ch chan<- *prometheus.Desc) {
}

// Collect prometheus collect
func (c *DeviceCollector) Collect(ch chan<- prometheus.Metric) {
	for _, instance := range c.Collects {
		for _, value := range instance.Metrics {
			switch value.Prefix {
			case "p5g_mec":
				ymlConfig := cfg.InitConfig()
				// openshift mec client
				k8sclient, _ := k8sClient.CreateCustomClientSet(ymlConfig.File.MEC_CONFIG)

				// service access token 토큰만들기
				token, err := k8sClient.CreateToken(k8sclient, "openshift-monitoring", "prometheus-k8s")
				if err != nil {
					logger.LogErr("mecCPU k8s client Token error", err)
				}
				err = c.scrape(value, ch, token)
				if err != nil {
					logger.LogErr("Scrpe Error : ", err)
					continue
				}
			default:
				err := c.scrape(value, ch, "")
				if err != nil {
					logger.LogErr("Scrpe Error : ", err)
					continue
				}
			}
		}

	}
}

// scrape connnect to database and gather query result
func (c *DeviceCollector) scrape(m *Metric, ch chan<- prometheus.Metric, token string) error {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   5 * time.Second,
	}
	req, err := http.NewRequest("GET", m.Url, nil)
	if err != nil {
		logger.LogErr("request Error ", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		logger.LogErr("Client Request Error", err)
		return errors.Cause(err)
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.LogErr("Body Test Error", err)
		return errors.Cause(err)
	}

	var result CoreData
	_ = json.Unmarshal(bodyText, &result)

	for _, v := range result.Data.Result {
		var labelVals []string
		metric := v.Metric
		value := v.Value[1].(string)

		if strings.Contains(m.Description, "cpu") {
			labelVals = []string{metric.Container, metric.CPU, metric.Endpoint, metric.Instance, metric.Job, metric.Mode, metric.Namespace, metric.Pod, metric.Service}
		} else if strings.Contains(m.Description, "memory") {
			labelVals = []string{metric.Container, metric.Endpoint, metric.Instance, metric.Job, metric.Namespace, metric.Pod, metric.Service}
		} else if strings.Contains(m.Description, "pod") {
			labelVals = []string{metric.Container, metric.Namespace, metric.Node, metric.Pod}
		}

		// value 파싱
		data, err := strconv.ParseFloat(value, 64)
		if err != nil {
			logger.LogErr("Error parsing value", err)
			continue
		}

		switch strings.ToLower(m.Type) {
		case "counter":
			ch <- prometheus.MustNewConstMetric(m.MetricDesc, prometheus.CounterValue, data, labelVals...)
		case "gauge":
			ch <- prometheus.MustNewConstMetric(m.MetricDesc, prometheus.GaugeValue, data, labelVals...)
		default:
			logger.LogErr(m.Description+" Metric type support only counter|gauge, skip", err)
			continue
		}
	}
	logger.LogInfo("["+m.Description+"]"+" Metric Saved", zap.String("Type", m.Type), zap.String("Prefix", m.Prefix))
	return nil
}
