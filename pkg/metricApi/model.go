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

type Metrics struct {
	Ran  RanStats  `json:"ran"`
	Core CoreStats `json:"core"`
}

type RanStats struct {
	Application RanApplicationStats `json:"application"`
	Physical    RanPhysicalStats    `json:"physical"`
}

type RanApplicationStats struct {
	Sum    RanAppSumStats     `json:"sum"`
	Detail []RanAppDetailInfo `json:"detail"`
}

type RanPhysicalStats struct {
	Sum    RanPhysicalSumStats     `json:"sum"`
	Detail []RanPhysicalDetailInfo `json:"detail"`
}

type RanAppSumStats struct {
	AirMacULByte  int `json:"airMacULByte"`
	AirMacDLByte  int `json:"airMacDLByte"`
	UeActiveDLAvg int `json:"ueActiveDLAvg"`
	UeActiveDLMax int `json:"ueActiveDLMax"`
}

type RanPhysicalSumStats struct {
	AirMacULByte_PCELL int `json:"airMacULByte_PCELL"`
	AirMacDLByte_PCELL int `json:"airMacDLByte_PCELL"`
	AirMacULByte_SCELL int `json:"airMacULByte_SCELL"`
	AirMacDLByte_SCELL int `json:"airMacDLByte_SCELL"`
}

type RanAppDetailInfo struct {
	NeID          string `json:"neId"`
	NeName        string `json:"neName"`
	InitTime      string `json:"initTile"`
	Location      string `json:"location"`
	AirMacULByte  int    `json:"airMacULByte"`
	AirMacDLByte  int    `json:"airMacDLByte"`
	UeActiveDLAvg int    `json:"ueActiveDLAvg"`
	UeActiveDLMax int    `json:"ueActiveDLMax"`
}

type RanPhysicalDetailInfo struct {
	NeID               string `json:"neId"`
	NeName             string `json:"neName"`
	InitTime           string `json:"initTile"`
	Location           string `json:"location"`
	AirMacULByte_PCELL int    `json:"airMacULByte_PCELL"`
	AirMacDLByte_PCELL int    `json:"airMacDLByte_PCELL"`
	AirMacULByte_SCELL int    `json:"airMacULByte_SCELL"`
	AirMacDLByte_SCELL int    `json:"airMacDLByte_SCELL"`
}

type CoreStats struct {
	Application CoreApplicationStats `json:"application"`
}
type CoreApplicationStats struct {
	Sum    CoreAppSumStats     `json:"sum"`
	Detail []CoreAppDetailInfo `json:"detail"`
}

type CoreAppSumStats struct {
	UeconAmfCRatio float64 `json:"ueconAmfCRatio"`
	UeidAmfCRatio  float64 `json:"ueidAmfCRatio"`
	AmftpsTotalMsg int     `json:"amftpsTotalMsg"`
	AmfmsCurCmConn int     `json:"amfmsCurCmConn"`
}

type CoreAppDetailInfo struct {
	InitTime string  `json:"initTile"`
	Location string  `json:"location"`
	Attempt  int     `json:"attempt"`
	Success  int     `json:"success"`
	Cachehit float64 `json:"cachehit"`
	CRatio   float64 `json:"cRatio"`
}
