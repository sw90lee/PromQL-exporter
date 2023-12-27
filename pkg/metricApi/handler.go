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
	"github.com/gin-gonic/gin"
	g "github.com/gosnmp/gosnmp"
	"kt.com/p5g/cnf-exporter/samsung-cpc/cfg"
	"log"
	"net/http"
	"os"
)

func CnfMetricHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ymlConfig := cfg.InitConfig()

		// RAN 기준 location 추출
		airLocations, err := FindCommonLocation("Air_MAC_Packet", ymlConfig)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"data":  "Ran.location is not find",
				"error": err,
			})
			return
		}
		cellLocations, err := FindCommonLocation("Air_MAC_Packet_(PCell)", ymlConfig)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"data":  "Ran.location is not find",
				"error": err,
			})
			return
		}

		data := &Metrics{}
		appsInfo := []RanAppDetailInfo{}
		phyInfo := []RanPhysicalDetailInfo{}
		coreInfo := []CoreAppDetailInfo{}

		// ran.application.detail의 json값 과 ran.application.sum 중 ueActiveDLAvg,ueActiveDLMax 값을 반환
		totAvgSum, totMaxSum, totAirMacUL, totAirMacDL, err := FincRanAppDetail(airLocations, ymlConfig, appsInfo, data)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"data":  "Ran.application is not find",
				"error": err,
			})
			return
		}
		data.Ran.Application.Sum.UeActiveDLAvg = totAvgSum
		data.Ran.Application.Sum.UeActiveDLMax = totMaxSum
		data.Ran.Application.Sum.AirMacULByte = totAirMacUL
		data.Ran.Application.Sum.AirMacDLByte = totAirMacDL

		// ran.physical.detail의 json값과 ran.physical.sum 의 AirMacULByte_PCELL, AirMacDLByte_PCELL, AirMacULByte_SCELL, AirMacDLByte_SCELL 값을 반환
		totPcellUL, totPcellDL, totScellUL, totScellDL, err := FincRanPhysicalDetail(cellLocations, ymlConfig, phyInfo, data)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"data":  "ran.physical is not find",
				"error": err,
			})
			return
		}
		data.Ran.Physical.Sum.AirMacULByte_PCELL = totPcellUL
		data.Ran.Physical.Sum.AirMacDLByte_PCELL = totPcellDL
		data.Ran.Physical.Sum.AirMacULByte_SCELL = totScellUL
		data.Ran.Physical.Sum.AirMacDLByte_SCELL = totScellDL

		// core.application.detail의 json값과 core.application.sum의 UeconAmfCRatio 값을 반환
		ueconAvgSum, err := FindUECONAppDetail(ymlConfig.File.CORE_NAME, ymlConfig, coreInfo, data)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"data":  "Core.UECON_AMF is not find",
				"error": err,
			})
			return
		}
		// core.application.sum의 ueidAvgSum 값을 반환
		ueidAvgSum, err := FindUEIDAppDetail(ymlConfig.File.CORE_NAME, ymlConfig)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"data":  "Core.UEID_AMF is not find",
				"error": err,
			})
			return
		}
		// core.application.sum의 amftpsSum,amfmsSum 값을 반환
		amftpsSum, amfmsSum, err := FindAMFTPSAppDetail(ymlConfig.File.CORE_NAME, ymlConfig)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"data":  "CORE.amftpsSum, amfmsSum is not find",
				"error": err,
			})
			return
		}

		data.Core.Application.Sum.UeconAmfCRatio = ueconAvgSum
		data.Core.Application.Sum.UeidAmfCRatio = ueidAvgSum
		data.Core.Application.Sum.AmftpsTotalMsg = amftpsSum
		data.Core.Application.Sum.AmfmsCurCmConn = amfmsSum

		c.JSON(http.StatusOK, data)

	}
}

func TrapService() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Default is a pointer to a GoSNMP struct that contains sensible defaults
		// eg port 161, community public, etc
		g.Default.Target = "116.89.189.122"
		//g.Default.Target = "172.30.10.122"
		g.Default.Port = 1162
		g.Default.Version = g.Version2c
		g.Default.Community = "public"
		g.Default.Logger = g.NewLogger(log.New(os.Stdout, "", 0))
		err := g.Default.Connect()
		if err != nil {
			log.Fatalf("Connect() err: %v", err)
		}

		defer g.Default.Conn.Close()

		location := g.SnmpPDU{
			Name:  "1.3.6.1.4.1.236.4.3.101.1.2.1.2.1.2.0",
			Type:  g.OctetString,
			Value: "/EMS-MFSM0 memory",
		}

		code := g.SnmpPDU{
			Name:  "1.3.6.1.4.1.236.4.3.101.1.2.1.4.0",
			Type:  g.Integer,
			Value: 1590,
		}

		msg := g.SnmpPDU{
			Name:  "1.3.6.1.4.1.236.4.3.101.1.2.1.2.1.5.0",
			Type:  g.OctetString,
			Value: "EMS Resource Alarm ( memory=100% )",
		}

		rating := g.SnmpPDU{
			Name:  "1.3.6.1.4.1.236.4.3.101.1.2.1.2.1.6.0",
			Type:  g.Integer,
			Value: 1,
		}

		device := g.SnmpPDU{
			Name:  "1.3.6.1.4.1.236.4.3.101.1.2.1.2.1.12.0",
			Type:  g.OctetString,
			Value: "ccpc",
		}

		//hex := g.SnmpPDU{
		//	Name:  "1.3.6.1.4.1.236.4.3.101.1.2.1.2.1.3.0",
		//	Type:  g.OctetString,
		//	Value: "0x07e701040b021d002b0900",
		//}

		trap := g.SnmpTrap{
			Variables: []g.SnmpPDU{location, rating, device, code, msg},
		}

		_, err = g.Default.SendTrap(trap)
		if err != nil {
			log.Fatalf("SendTrap() err: %v", err)
		}
	}
}
