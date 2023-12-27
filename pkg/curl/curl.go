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
package curl

import (
	"crypto/tls"
	"fmt"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"io"
	"kt.com/p5g/cnf-exporter/samsung-cpc/cfg"
	"kt.com/p5g/cnf-exporter/samsung-cpc/logger"
	"kt.com/p5g/cnf-exporter/samsung-cpc/pkg/k8sClient"
	"kt.com/p5g/cnf-exporter/samsung-cpc/pkg/utils"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// startime과 endtime은 15분단위로 설정 됨
// config.yml or k8s ENV에 설정시 사용하는 옵션
func ExporterCurl(startime, endtime, foldername, backupTime string, config cfg.Config) error {
	// k8sclient , k8sconfig 생성
	k8s_client, k8sconfig := k8sClient.CreateClientSet()
	podexec := k8sClient.NewK8sClient(*k8sconfig, k8s_client)

	// config.yml에  File에 적어둔 FamilyName 을 하나씩 가져와서 Curl을 날리고
	// k8s cp 를 통해 서버 로컬(config.file.path)에 저장 함
	for _, familyValue := range config.File.Family_Name {
		baseURL := config.Exporter.Curl_Url
		err := exporterCommon(baseURL, familyValue, startime, endtime, config, podexec)
		if err != nil {
			logger.LogErr("apicommon is error", err)
			return err
		}
	}
	/*
		파일 경로 설정 후 폴더만 생성
	*/
	// 첫번째 폴더 생성(년월일시)
	_ = utils.Mkdir(foldername, startime, endtime, config.File.CSV_Path)
	// 두번째 폴더 생성(시간분-시간분)
	_ = utils.Mkdir(backupTime, startime, endtime, config.File.CSV_Path+"/"+foldername)

	return nil
}

func curl(baseUrl, familyName, startTime, endTime, username, userpassword string) ([]byte, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// fullUrl : curl -XGET -k -g -u 'ossuser:osspasswd' 'https://116.89.173.40:7443/oss/performanceData?Family%20name=UECON_AMF&startTime=2023-11-08%2013:34:00&endTime=2023-11-08%2013:49:00'
	// baseUrl : https://116.89.173.40:7443/oss/performanceData
	// FamilyName : DU%20Power%20Consumption&startTime=2023-09-23%2001:15:00&endTime=2023-09-25%2017:15:00
	// Starttime : 2023-09-23 01:15:00
	// EndTime : 2023-09-23 01:30:00

	// Insecure Skip
	client := &http.Client{Transport: tr}
	// 파라미터 생성
	params := url.Values{}
	params.Add("Family name", familyName)
	params.Add("startTime", startTime)
	params.Add("endTime", endTime)

	fullUrl := baseUrl + "?" + params.Encode()

	logger.LogInfo(fullUrl)

	req, err := http.NewRequest("GET", fullUrl, nil)
	if err != nil {
		logger.LogErr("request Error ", err)
		return nil, errors.Cause(err)
	}
	req.SetBasicAuth(username, userpassword)

	resp, err := client.Do(req)
	if err != nil {
		logger.LogErr("response Error", err)
		return nil, errors.Cause(err)
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.LogErr("bodyText Error", err)
		return nil, errors.Cause(err)
	}

	return bodyText, nil
}

func exporterCommon(baseURL, familyValue, startime, endtime string, config cfg.Config, podexec *k8sClient.Client) error {
	// curl 날리는 명령어 확인하기
	logger.LogInfo("curl command", zap.String("baseURL", baseURL), zap.String("familyName", familyValue), zap.String("startime", startime), zap.String("endtime", endtime))
	output, err := curl(baseURL, familyValue, startime, endtime, config.Exporter.Oss_Username, config.Exporter.Oss_Password)
	if err != nil {
		logger.LogErr("Unable to execute the curl command.", err)
		return errors.Cause(err)
	}

	// curl을 날리고 떨어지는 값 (csv파일의 경로+파일명)
	// /home/vsm/aceman/web_oss/var/pm/performanceData_20231020_144317.csv
	logger.LogInfo(string(output))

	// output로 파일 떨구기
	if strings.Contains(familyValue, " ") {
		NewFamilyValue := strings.ReplaceAll(familyValue, " ", "_")
		// nameSpace : usm-compact
		// podName : mfsm-0
		err := podexec.CopyFromPod("mfsm-0", "usm-compact", "process", string(output), config.File.CSV_Path)
		oldFileName := strings.Split(string(output), "/")
		oldName := fmt.Sprint(config.File.CSV_Path + "/" + oldFileName[7])
		newName := fmt.Sprint(config.File.CSV_Path + "/" + NewFamilyValue + ".csv")
		os.Rename(oldName, newName)
		if err != nil {
			logger.LogErr("Copy Failed", err)
			return errors.Cause(err)
		}
		// ANSICODE 없는 name
	} else {
		// nameSpace : usm-compact
		// podName : mfsm-0
		err := podexec.CopyFromPod("mfsm-0", "usm-compact", "process", string(output), config.File.CSV_Path)
		oldFileName := strings.Split(string(output), "/")
		oldName := fmt.Sprint(config.File.CSV_Path + "/" + oldFileName[7])
		newName := fmt.Sprint(config.File.CSV_Path + "/" + familyValue + ".csv")
		os.Rename(oldName, newName)
		if err != nil {
			logger.LogErr("Copy Failed", err)
			return errors.Cause(err)
		}
	}
	// 1초 delay
	time.Sleep(1 * time.Second)

	return nil
}
