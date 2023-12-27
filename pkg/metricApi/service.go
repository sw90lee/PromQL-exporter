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
package metricApi

import (
	"fmt"
	"github.com/pkg/errors"
	"kt.com/p5g/cnf-exporter/samsung-cpc/cfg"
	"kt.com/p5g/cnf-exporter/samsung-cpc/logger"
	"kt.com/p5g/cnf-exporter/samsung-cpc/pkg/csv"
	"strconv"
	"strings"
)

// 로케이션 찾는 공통 함수
func FindCommonLocation(value string, ymlConfig cfg.Config) ([]string, error) {
	path := fmt.Sprintf(ymlConfig.File.API_Path + "/" + value + ".csv")
	csvData, err := csv.LoadCsv(path)
	if err != nil {
		logger.LogErr("FindCommonLocation Method error", err)
		return nil, errors.Cause(err)
	}

	var locations []string
	unique := make(map[string]struct{})

	for i := 3; i < len(csvData); i++ {
		neID := csvData[i][6]
		if _, exists := unique[neID]; !exists {
			locations = append(locations, csvData[i][6])
			unique[neID] = struct{}{}
		}
	}

	return locations, nil
}

// ran.application에 대한 데이터 가공
func FincRanAppDetail(locations []string, ymlConfig cfg.Config, detail []RanAppDetailInfo, ran *Metrics) (int, int, int, int, error) {

	var totAvgSum int
	var totMaxSum int
	var totAirMacUL int
	var totAirMacDL int

	for _, location := range locations {
		detailinfo := RanAppDetailInfo{}

		// Downlink_Active_UE_Number 불러오기 위한 csvData
		downActiveUePath := fmt.Sprintf(ymlConfig.File.API_Path + "/" + "Downlink_Active_UE_Number" + ".csv")
		UeData, err := csv.LoadCsv(downActiveUePath)
		if err != nil {
			logger.LogErr("Downlink_Active_UE_Number CSV File is not Open", err)
			return 0, 0, 0, 0, errors.Cause(err)
		}

		// Air_MAC_Packet 불러오기 위한 csvData
		airMacPath := fmt.Sprintf(ymlConfig.File.API_Path + "/" + "Air_MAC_Packet" + ".csv")
		airMacData, err := csv.LoadCsv(airMacPath)
		if err != nil {
			logger.LogErr("Air_MAC_Packet CSV File is not Open", err)
			return 0, 0, 0, 0, errors.Cause(err)
		}

		var count int64
		var sum int64
		var neid string
		var neName string
		var initTime string
		var airMacULByte int
		var airMacDLByte int

		// location에 최대값을 저장하기위한 map
		maxValues := make(map[string]int)

		// 로케이션에 해당하는 UEActiveDLAvg(count) 값을 찾기
		// 카운트 찾아 해당 값 더하기
		for i := 3; i < len(UeData); i++ {
			if strings.Contains(UeData[i][6], location) {
				count++
				// location의 UEActiveDLAvg
				data, _ := strconv.ParseInt(UeData[i][7], 10, 64)

				// location의 UEActiveDLAvg 평균값을 구하기위한 합
				sum += data
				// neID 정의
				neid = UeData[i][0]
				// neName 정의
				neName = UeData[i][2]
				// initTime 정의
				initTime = UeData[i][3]

				//UEActiveDLMaxCnt
				maxCountStr := UeData[i][10]
				// UEActiveDLMax(count) 값을 정수로 변환
				maxCount, _ := strconv.Atoi(maxCountStr)

				// 맵에 현재 LOCATION에 대한 최대값 업데이트
				if currentMax, ok := maxValues[location]; !ok || maxCount > currentMax {
					maxValues[location] = maxCount
				}

				// application.sum 데이터
				avgSum, _ := strconv.Atoi(UeData[i][7])
				if err != nil {
					logger.LogErr(location+"  UEActiveDLAvg(count) Sum failed ParseInt", err)
					return 0, 0, 0, 0, errors.Cause(err)
				}
				maxSum, _ := strconv.Atoi(UeData[i][10])
				totAvgSum += avgSum
				totMaxSum += maxSum
			}
		}

		//
		for i := 3; i < len(airMacData); i++ {
			if strings.Contains(airMacData[i][6], location) {
				airMacULByte, _ = strconv.Atoi(airMacData[i][7])
				airMacDLByte, _ = strconv.Atoi(airMacData[i][10])

				ul, _ := strconv.Atoi(airMacData[i][7])
				dl, _ := strconv.Atoi(airMacData[i][10])

				totAirMacUL += ul
				totAirMacDL += dl
			}
		}

		avg := sum / count
		detailinfo = RanAppDetailInfo{
			NeID:          neid,
			NeName:        neName,
			InitTime:      initTime,
			Location:      location,
			UeActiveDLAvg: int(avg),
			UeActiveDLMax: maxValues[location],
			AirMacULByte:  airMacULByte,
			AirMacDLByte:  airMacDLByte,
		}

		detail = append(detail, detailinfo)

		ran.Ran.Application.Detail = detail
	}
	return totAvgSum, totMaxSum, totAirMacUL, totAirMacDL, nil
}

// ran.application에 대한 데이터 가공
func FincRanPhysicalDetail(locations []string, ymlConfig cfg.Config, detail []RanPhysicalDetailInfo, ran *Metrics) (int, int, int, int, error) {

	var totULPCELLSum int
	var totDLPCELLSum int
	var totULSCELLSum int
	var totDLSCELLSum int

	for _, location := range locations {
		detailinfo := RanPhysicalDetailInfo{}

		// Downlink_Active_UE_Number 불러오기 위한 csvData
		pCellPath := fmt.Sprintf(ymlConfig.File.API_Path + "/" + "Air_MAC_Packet_(PCell)" + ".csv")
		pCellData, err := csv.LoadCsv(pCellPath)
		if err != nil {
			logger.LogErr("Air_MAC_Packet_(PCell) CSV File is not Open", err)
			return 0, 0, 0, 0, errors.Cause(err)
		}

		// Air_MAC_Packet 불러오기 위한 csvData
		sCellPath := fmt.Sprintf(ymlConfig.File.API_Path + "/" + "Air_MAC_Packet_(SCell)" + ".csv")
		sCellData, err := csv.LoadCsv(sCellPath)
		if err != nil {
			logger.LogErr("Air_MAC_Packet_(SCell) CSV File is not Open", err)
			return 0, 0, 0, 0, errors.Cause(err)
		}

		var pcellULsum int64
		var pcellDLsum int64
		var scellULsum int64
		var scellDLsum int64
		var neid string
		var neName string
		var initTime string

		// 로케이션에 해당하는 AirMacULByte_PCELL 값을 찾기
		// 카운트 찾아 해당 값 더하기
		for i := 3; i < len(pCellData); i++ {
			if strings.Contains(pCellData[i][6], location) {
				// location의 AirMacULByte_PCELL
				ULdata, _ := strconv.ParseInt(pCellData[i][7], 10, 64)
				// location의 Detail에 넣기 위한 sum
				pcellULsum += ULdata

				// location의 AirMacDLByte_PCELL
				DLdata, _ := strconv.ParseInt(pCellData[i][10], 10, 64)
				// location의 Detail에 넣기 위한 sum
				pcellDLsum += DLdata

				// neID 정의
				neid = pCellData[i][0]
				// neName 정의
				neName = pCellData[i][2]
				// initTime 정의
				initTime = pCellData[i][3]

				// application.sum 데이터
				ULpCellSum, _ := strconv.Atoi(pCellData[i][7])
				DLpCellSum, _ := strconv.Atoi(pCellData[i][10])
				totULPCELLSum += ULpCellSum
				totDLPCELLSum += DLpCellSum
			}
		}

		//
		for i := 3; i < len(sCellData); i++ {
			if strings.Contains(sCellData[i][6], location) {
				// location의 AirMacULByte_PCELL
				ULdata, _ := strconv.ParseInt(sCellData[i][7], 10, 64)
				// location의 Detail에 넣기 위한 sum
				scellULsum += ULdata

				// location의 AirMacDLByte_PCELL
				DLdata, _ := strconv.ParseInt(sCellData[i][10], 10, 64)
				// location의 Detail에 넣기 위한 sum
				scellDLsum += DLdata

				// application.sum 데이터
				ULpCellSum, _ := strconv.Atoi(sCellData[i][7])
				DLpCellSum, _ := strconv.Atoi(sCellData[i][10])
				totULSCELLSum += ULpCellSum
				totDLSCELLSum += DLpCellSum
			}
		}

		detailinfo = RanPhysicalDetailInfo{
			NeID:               neid,
			NeName:             neName,
			InitTime:           initTime,
			Location:           location,
			AirMacULByte_PCELL: int(pcellULsum),
			AirMacDLByte_PCELL: int(pcellDLsum),
			AirMacULByte_SCELL: int(scellULsum),
			AirMacDLByte_SCELL: int(scellDLsum),
		}

		detail = append(detail, detailinfo)

		ran.Ran.Physical.Detail = detail
	}
	return totULPCELLSum, totDLPCELLSum, totULSCELLSum, totDLSCELLSum, nil
}

// core.UECON 대한 데이터 가공
func FindUECONAppDetail(coreNames []string, ymlConfig cfg.Config, detail []CoreAppDetailInfo, core *Metrics) (float64, error) {

	var ueconCratioAvg float64

	for _, name := range coreNames {
		detailinfo := CoreAppDetailInfo{}

		switch name {
		case "UECON_AMF":
			path := fmt.Sprintf(ymlConfig.File.API_Path + "/" + name + ".csv")
			data, err := csv.LoadCsv(path)
			if err != nil {
				logger.LogErr(name+" CSV File is not Open", err)
				return 0, errors.Cause(err)
			}
			var count int64
			var sum float64
			//var cRatio int64

			if len(data) != 3 {
				for i := 3; i < len(data); i++ {
					count++
					dataStr, _ := strconv.ParseFloat(data[i][18], 10)
					sum += dataStr

					if strings.Contains(data[0][0], "UECON_AMF") {
						attempt, _ := strconv.Atoi(data[i][7])
						success, _ := strconv.Atoi(data[i][8])
						cachehit, _ := strconv.ParseFloat(data[i][9], 10)
						cratio, _ := strconv.ParseFloat(data[i][18], 10)

						detailinfo = CoreAppDetailInfo{
							InitTime: data[i][3],
							Location: data[i][6],
							Attempt:  attempt,
							Success:  success,
							Cachehit: cachehit,
							CRatio:   cratio,
						}
						detail = append(detail, detailinfo)
					}
				}
				ueconCratioAvg = sum / float64(count)
			} else {
				detailinfo = CoreAppDetailInfo{
					InitTime: "",
					Location: "",
					Attempt:  0,
					Success:  0,
					Cachehit: 0,
					CRatio:   0,
				}
				detail = append(detail, detailinfo)
				core.Core.Application.Detail = detail
			}
			core.Core.Application.Detail = detail
		}
	}

	return ueconCratioAvg, nil
}

// core.UEID_AMF 대한 데이터 가공
func FindUEIDAppDetail(coreNames []string, ymlConfig cfg.Config) (float64, error) {

	var ueidCratioAvg float64

	for _, name := range coreNames {
		switch name {
		case "UEID_AMF":
			path := fmt.Sprintf(ymlConfig.File.API_Path + "/" + name + ".csv")
			data, err := csv.LoadCsv(path)
			if err != nil {
				logger.LogErr(name+" CSV File is not Open", err)
				return 0, errors.Cause(err)
			}
			var count int64
			var sum float64
			//var cRatio int64

			if len(data) != 3 {
				for i := 3; i < len(data); i++ {
					count++
					dataStr, _ := strconv.ParseFloat(data[i][18], 10)
					sum += dataStr
				}
				ueidCratioAvg = sum / float64(count)
			} else {
				return 0, nil
			}
		}
	}
	return ueidCratioAvg, nil
}

// core.application에 대한 데이터 가공
func FindAMFTPSAppDetail(coreNames []string, ymlConfig cfg.Config) (int, int, error) {

	var amftpsTotMsg int
	var amfmsCurReg int

	for _, name := range coreNames {
		path := fmt.Sprintf(ymlConfig.File.API_Path + "/" + name + ".csv")
		data, err := csv.LoadCsv(path)
		if err != nil {
			logger.LogErr(name+" CSV File is not Open", err)
			return 0, 0, errors.Cause(err)
		}

		var sum int64

		if len(data) == 3 {
			return 0, 0, nil
		} else {
			for i := 3; i < len(data); i++ {
				data, _ := strconv.ParseInt(data[i][9], 10, 64)
				sum += data
			}
			switch name {
			case "AMFMS":
				amftpsTotMsg = int(sum)
			case "AMFTPS":
				amfmsCurReg = int(sum)
			}

		}
	}
	return amftpsTotMsg, amfmsCurReg, nil
}
