/*
* Samsung-cpc version 1.0
*
*  Copyright ⓒ 2023 kt corp. All rights reserved.
*
*  This is a proprietary software of kt corp, and you may not use this file except in
*  compliance with license agreement with kt corp. Any redistribution or use of this
*  software, with or without modification shall be strictly prohibited without prior written
*  approval of kt corp, and the copyright notice above does not evidence any actual or
*  intended publication of such software.
 */
package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"io"
	"kt.com/p5g/cnf-exporter/samsung-cpc/cfg"
	"kt.com/p5g/cnf-exporter/samsung-cpc/logger"
	"kt.com/p5g/cnf-exporter/samsung-cpc/pkg/exporter"

	"kt.com/p5g/cnf-exporter/samsung-cpc/pkg/csv"
	"kt.com/p5g/cnf-exporter/samsung-cpc/pkg/curl"
	"kt.com/p5g/cnf-exporter/samsung-cpc/pkg/metricApi"
	"kt.com/p5g/cnf-exporter/samsung-cpc/pkg/utils"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	Metrics map[string]struct {
		Description    string
		Type           string
		Labels         []string
		Value          string
		Value_Sequence int
		MetricDesc     *prometheus.Desc
	}
}

var metricConfig Config
var collectors map[string]*exporter.Collector

func main() {
	var err error
	var configFile string
	var deviceConfig string
	// =====================
	// Get OS parameter
	// =====================

	flag.StringVar(&configFile, "metricConfig", "cnf_config.yml", "configuration file")
	flag.StringVar(&deviceConfig, "config-metrics", "app_config.yml", "configuration metrics")

	flag.Parse()

	var b []byte
	if b, err = os.ReadFile(configFile); err != nil {
		logger.LogErr("Failed to read metricConfig file: ", err)
		os.Exit(1)
	}

	// Load yaml
	if err := yaml.Unmarshal(b, &metricConfig); err != nil {
		logger.LogErr("Failed to load metricConfig: ", err)
		os.Exit(1)
	}

	if b, err = os.ReadFile(deviceConfig); err != nil {
		logger.LogErr("Failed to read deviceConfig file: ", err)
		os.Exit(1)
	}

	if err := yaml.Unmarshal(b, &collectors); err != nil {
		logger.LogErr("Failed to load deviceConfig: ", err)
		os.Exit(1)
	}

	statusDesc := prometheus.NewDesc(
		prometheus.BuildFQName("cnf_exporter", "", "status"),
		"cnf_exporter collect status",
		[]string{"instance"}, nil,
	)

	router := gin.Default()

	/*
		APP Exporter
	*/
	for path, collector := range collectors {
		logger.LogInfo("path : " + path)

		for i := range collector.Collects {
			collect := &collector.Collects[i]

			for metricKey, metric := range collect.Metrics {
				metric.MetricDesc = prometheus.NewDesc(
					prometheus.BuildFQName(metric.Prefix, "", metricKey),
					metric.Description,
					metric.Labels, nil,
				)
			}
		}

		registry := prometheus.NewRegistry()
		registry.Register(&exporter.DeviceCollector{Collects: collector.Collects, StatusDesc: statusDesc})

		router.GET("/"+path, gin.WrapH(
			promhttp.HandlerFor(prometheus.Gatherers{
				registry,
			},
				promhttp.HandlerOpts{})),
		)
	}

	/*
		Prometheus에 Metric Data 보내기
	*/
	cnf := prometheus.NewRegistry()
	cnf.Register(version.NewCollector("cnf_exporter"))
	cnf.Register(&CnfCollector{})

	router.GET("/api/metrics", metricApi.CnfMetricHandler())

	router.GET("/metrics", gin.WrapH(
		promhttp.HandlerFor(prometheus.Gatherers{cnf},
			promhttp.HandlerOpts{})),
	)

	router.Run(":8080")

}

type CnfCollector struct{}

// Describe prometheus describe
// 메트릭에 사용하는 스펙정의
func (c *CnfCollector) Describe(ch chan<- *prometheus.Desc) {
	for metricName, metric := range metricConfig.Metrics {
		metric.MetricDesc = prometheus.NewDesc(
			prometheus.BuildFQName("p5g_exporter", "", metricName),
			metric.Description,
			//라벨명 배열, 이 순서로 라벨값들이 추후 맵핑되어야 함
			metric.Labels,
			nil,
		)
		metricConfig.Metrics[metricName] = metric
		logger.LogInfo("Metric description register : "+metricName, zap.String("MetricName", metricName))

	}
}

// Collect prometheus collect
func (c *CnfCollector) Collect(ch chan<- prometheus.Metric) {
	ymlConfig := cfg.InitConfig()
	start, end := utils.IntervalTime()
	//YYYY-MM-DD 년-월-일로 폴더 생성
	foldername := start[:10]
	//hhmm-hhmm 시간분-시간분 으로 폴더생성
	backupTime := fmt.Sprintf("%s%s-%s%s", start[11:13], start[14:16], end[11:13], end[14:16])
	//수집전 curl 날려서 파일저장하기
	//curl 후 폴더만 생성진행 함
	err := curl.ExporterCurl(start, end, foldername, backupTime, ymlConfig)
	if err != nil {
		logger.LogErr("ExporterCurl Method Error", err)
	}

	var csvData [][]string
	for metricName, metricKey := range metricConfig.Metrics {
		//csv 파일 가져오기
		// metrics.descripon을 가져와 FamilyName.csv를 연다
		path := fmt.Sprintf(ymlConfig.File.CSV_Path + "/" + metricKey.Description + ".csv")
		csvData, err = csv.LoadCsv(path)
		if err != nil {
			logger.LogErr("CSV 파일 Open 실패", err)
			continue
		}
		// Metric Value 추출할 컬럼
		metricSequence := metricKey.Value_Sequence
		// Metric Type
		metricType := metricKey.Type
		// Metric Spec정의한 Describe 가져요기
		metricDesc := metricKey.MetricDesc

		// 데이터가 없을경우 넘어감 => 다음메트릭을 수집
		if len(csvData) == 3 {
			logger.LogWarn(metricName+" 에 해당 데이터가 없습니다.", zap.String("MetricName", metricName))
			continue
		} else {
			err = commonCollect(csvData, metricSequence, metricType, metricDesc, ch)
			if err != nil {
				logger.LogErr("Failed to run collector", err)
			}
		}
	}

	if _, err := os.Stat(ymlConfig.File.API_Path); err == nil {
		_ = os.Mkdir(ymlConfig.File.API_Path, 0755)
	} else {
		logger.LogInfo("폴더가 존재합니다.")
	}

	// API 파일경로가 없을경우 생성 ( 있으면 넘어감 )
	if _, err := os.Stat(ymlConfig.File.API_Path); err != nil {
		_ = os.MkdirAll(ymlConfig.File.API_Path, 0755)
	}

	// CORE API FILE BACKUP
	for _, value := range ymlConfig.File.CORE_NAME {
		err = copyFile(ymlConfig.File.CSV_Path, ymlConfig.File.API_Path, value)
		if err != nil {
			logger.LogErr("API metric backup failed", err)
		}
	}
	// CORE API FILE BACKUP
	for _, value := range ymlConfig.File.RAN_NAME {
		err = copyFile(ymlConfig.File.CSV_Path, ymlConfig.File.API_Path, value)
		if err != nil {
			logger.LogErr("API metric API backup failed", err)
		}
	}

	//파일 백업
	err = backup(foldername, backupTime, ymlConfig.File.CSV_Path, ymlConfig.File.Family_Name)
	if err != nil {
		logger.LogErr("exporter file backup failed", err)
	}
}

func commonCollect(csvData [][]string, metricSequnce int, metricType string, metricDesc *prometheus.Desc, ch chan<- prometheus.Metric) error {
	for i := 3; i < len(csvData); i++ {
		// 라벨 데이터
		labelVals := []string{}

		// label value 값 셋팅하기
		for j := 0; j < 7; j++ {
			labelVals = append(labelVals, csvData[i][j])
		}

		// 값을 넣기위해 float64 parser
		val, err := strconv.ParseFloat(csvData[i][metricSequnce], 64)
		if err != nil {
			logger.LogErr("csvData Parser Error", err)
			return errors.Cause(err)
		}
		// metricType 설정
		switch strings.ToLower(metricType) {
		case "counter":
			ch <- prometheus.MustNewConstMetric(metricDesc, prometheus.CounterValue, val, labelVals...)
		case "gauge":
			ch <- prometheus.MustNewConstMetric(metricDesc, prometheus.GaugeValue, val, labelVals...)
		default:
			logger.LogWarn("Fail to add metric for is not valid type")
			continue
		}
	}
	return nil
}

func backup(foldername, backupfolder, path string, familyName []string) error {
	//파일 이동
	for _, familyValue := range familyName {
		switch strings.Contains(familyValue, " ") {
		case true:
			NewFamilyValue := strings.ReplaceAll(familyValue, " ", "_")
			err := moveFile(foldername, backupfolder, path, NewFamilyValue)
			if err != nil {
				logger.LogErr("File backup failed", err)
				return errors.Cause(err)
			}
		case false:
			err := moveFile(foldername, backupfolder, path, familyValue)
			if err != nil {
				logger.LogErr("File backup failed", err)
				return errors.Cause(err)
			}
		}
	}
	return nil
}

func moveFile(folderName, backupFolderName, path, familyName string) error {
	err := os.Rename(path+"/"+familyName+".csv", path+"/"+folderName+"/"+backupFolderName+"/"+familyName+".csv")
	if err != nil {
		return err
	}
	logger.LogInfo("백업 파일이동 성공", zap.String("familyName", familyName))
	return nil
}

func copyFile(oldpath, newpath, familyName string) error {
	sourceFile, err := os.Open(oldpath + "/" + familyName + ".csv")
	if err != nil {
		logger.LogErr(familyName+": file open is failed", err)
		return errors.Cause(err)
	}
	defer sourceFile.Close()

	// 대상 파일 생성 또는 열기 (존재하면 덮어쓰기)
	destinationFilePath := filepath.Join(newpath, filepath.Base(oldpath+"/"+familyName+".csv"))
	destinationFile, err := os.Create(destinationFilePath)
	if err != nil {
		logger.LogErr(familyName+": file create is failed", err)
		return errors.Cause(err)
	}
	defer destinationFile.Close()

	// 파일 복사
	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		logger.LogErr(familyName+": file copy is failed", err)
		return errors.Cause(err)
	}
	return nil
}
