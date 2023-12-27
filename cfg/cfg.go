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
package cfg

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"strconv"
)

type Config struct {
	Logging  Logging
	File     File
	Exporter Exporter
	//Prom    Prom
}

type Logging struct {
	Level  string `mapstructure:"LEVEL"`
	Encode string `mapstructure:"ENCODE"`
}

type File struct {
	MEC_CONFIG  string   `mapstructure:"MEC_CONFIG"`
	CSV_Path    string   `mapstructure:"CSV_PATH"`
	API_Path    string   `mapstructure:"API_PATH"`
	Family_Name []string `mapstructure:"FAMILY_NAME"`
	RAN_NAME    []string `mapstructure:"RAN_NAME"`
	CORE_NAME   []string `mapstructure:"CORE_NAME"`
}

type Exporter struct {
	Curl_Url     string `mapstrcuture:"CURL_URL"`
	Oss_Username string `mapstructure:"OSS_USERNAME"`
	Oss_Password string `mapstructure:"OSS_PASSWORD"`
}

//type Prom struct {
//	Url string `mapstructure:"URL"`
//}

func InitConfig() Config {
	env := getEnv("ENV", "local") // 환경 변수를 통해 현재 환경을 확인하거나 기본값으로 'local'을 사용합니다.

	var configFileName string
	switch env {
	case "local":
		configFileName = "config_local"
	case "prd":
		configFileName = "config"
	default:
		fmt.Printf("Unknown environment: %s. Using default config.\n", env)
		configFileName = "config"
	}

	viper.SetConfigName(configFileName)
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")

	// viper overide
	viper.AutomaticEnv()

	// viper defaultSet 설정
	viper.SetDefault("loging.level", getEnv("LEVEL", "INFO"))
	viper.SetDefault("loging.encode", getEnv("ENCODE", "JSON"))
	viper.SetDefault("file.mec_config", getEnv("MEC_CONFIG", "/mnt/data/config"))
	viper.SetDefault("file.csv_path", getEnv("CSV_PATH", ""))
	viper.SetDefault("file.api_path", getEnv("API_PATH", ""))
	viper.SetDefault("file.family_name", getEnv("FAMILY_NAME", ""))
	viper.SetDefault("file.ran_name", getEnv("RAN_NAME", ""))
	viper.SetDefault("file.core_name", getEnv("CORE_NAME", ""))
	viper.SetDefault("exporter.curl_url", getEnv("CURL_URL", ""))
	viper.SetDefault("exporter.oss_username", getEnv("OSS_USERNAME", ""))
	viper.SetDefault("exporter.oss_password", getEnv("OSS_PASSWORD", ""))

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Printf("Config 파일 로드 중 에러: %v\n", err)
		}
		fmt.Println("Config 파일이 존재하지 않아 기본값을 사용합니다.")
	}

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		fmt.Println("config 매핑 에러")
	}

	return config
}

// env String 반환
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// env String 반환
func getEnvAsBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	valueBool, _ := strconv.ParseBool(value)
	return valueBool
}

// env int 반환
func getEnvAsInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	var intValue int
	_, err := fmt.Scan(value, &intValue)
	if err != nil {
		return defaultValue
	}
	return intValue
}
