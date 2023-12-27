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
package exporter

type CoreData struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric CoreMetricData `json:"metric"`
			Value  []interface{}  `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

type CoreMetricData struct {
	Name      string `json:"__name__"`
	Container string `json:"container"`
	CPU       string `json:"cpu"`
	Endpoint  string `json:"endpoint"`
	Instance  string `json:"instance"`
	Job       string `json:"job"`
	Mode      string `json:"mode"`
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
	Service   string `json:"service"`
	Node      string `json:"node"`
}
