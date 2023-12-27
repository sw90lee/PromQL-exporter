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
package utils

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"io"
	"kt.com/p5g/cnf-exporter/samsung-cpc/logger"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Startime과 Endtime을 반환한다.
func IntervalTime() (start string, end string) {
	//15분전 ~ 현재시간 조회 설정
	startime := time.Now().Add(-16 * time.Minute)
	endtime := time.Now().Add(1 * time.Minute)
	// 초 부분을 00으로 설정
	startime = time.Date(startime.Year(), startime.Month(), startime.Day(), startime.Hour(), startime.Minute(), 0, 0, startime.Location())
	endtime = time.Date(endtime.Year(), endtime.Month(), endtime.Day(), endtime.Hour(), endtime.Minute(), 0, 0, endtime.Location())

	// 원하는 형식으로 문자열로 포맷팅
	startTimeStr := startime.Format("2006-01-02 15:04:05")
	endTimeStr := endtime.Format("2006-01-02 15:04:05")

	return startTimeStr, endTimeStr
}

func Mkdir(name, start, end, path string) error {
	if _, err := os.Stat(path + "/" + name); os.IsNotExist(err) {
		// 폴더가 존재 하지않으므로 생성
		err = os.MkdirAll(path+"/"+name, 0755)
		if err != nil {
			logger.LogErr("폴더생성 실패", err)
			return errors.Cause(err)
		}
	}
	return nil
}

func Backup(foldername, backupfolder, path string, familyName []string) error {
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
