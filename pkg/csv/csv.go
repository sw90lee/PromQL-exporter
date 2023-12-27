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
package csv

import (
	"encoding/csv"
	"errors"
	"os"
	"strings"
)

func LoadCsv(path string) ([][]string, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// 파일이 존재하지 않음, 다음 파일을 시도
			return nil, err
		} else {
			// 다른 오류가 발생한 경우
			return nil, err
		}
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read the CSV data
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // Allow variable number of fields
	data, err := reader.ReadAll()

	return data, err
}

// 라벨의 값을 붙일 컬럼 갯수 반환
// 라벨 넣을 컬럼이 공통이므로 뺌
func MetricColumnCount(data [][]string) (int, error) {
	var count int

	if len(data) == 0 {
		return 0, errors.New("파일이 존재하지 않습니다.")
	} else {
		for _, v := range data[2][:] {
			if strings.Contains(v, "(") {
				//if strings.Contains(v, "(count)") || strings.Contains(v, "(msec)") {
				count++
			}
		}
	}

	count = (len(data[2][:]) - count)

	return count, nil
}
