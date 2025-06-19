package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// 환경변수 기본값 상수
const (
	DEFAULT_OCR_API_URL     = "https://xlqf1p9tsl.apigw.ntruss.com/custom/v1/43074/fd2e3084b0aedaa21b114b92ed8d5b021e1f2c88b9eb7caea4b7e4b602aa85ea/infer"
	DEFAULT_SERVER_PORT     = "3000"
	DEFAULT_UPLOAD_LIMIT_MB = 50
	DEFAULT_MAX_FILES       = 10
	DEFAULT_OCR_TIMEOUT     = 60
	DEFAULT_ATTR_CD         = "8A" // 신용카드매출전표(개인)
)

// 지원하는 이미지 형식
var SUPPORTED_IMAGE_FORMATS = []string{".jpg", ".jpeg", ".png", ".pdf", ".tif", ".tiff"}

// 환경변수 getter 함수들 (통합)
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}

func getEnvBool(key string) bool {
	value := strings.ToLower(os.Getenv(key))
	return value == "true" || value == "1" || value == "on"
}

// OCR 관련 설정
func getOCRSecret() string {
	secret := os.Getenv("X_OCR_SECRET")
	if secret == "" {
		panic("X_OCR_SECRET 환경변수가 설정되지 않았습니다")
	}
	return secret
}

func getOCRAPIURL() string {
	return getEnvString("OCR_API_URL", DEFAULT_OCR_API_URL)
}

func getOCRTimeoutSeconds() int {
	return getEnvInt("OCR_TIMEOUT_SECONDS", DEFAULT_OCR_TIMEOUT)
}

func isOCRVerboseLog() bool {
	return getEnvBool("OCR_VERBOSE_LOG")
}

// 서버 설정
func getServerPort() string {
	port := getEnvString("SERVER_PORT", DEFAULT_SERVER_PORT)
	return ":" + port
}

func getUploadLimitMB() int {
	return getEnvInt("UPLOAD_LIMIT_MB", DEFAULT_UPLOAD_LIMIT_MB)
}

func getMaxFiles() int {
	return getEnvInt("MAX_FILES", DEFAULT_MAX_FILES)
}

// 기본값 설정들
func getDefaultDepositorDC() string {
	return getEnvString("DEFAULT_DEPOSITOR_DC", "")
}

func getDefaultDeptCD() string {
	return getEnvString("DEFAULT_DEPT_CD", "")
}

func getDefaultEmpCD() string {
	return getEnvString("DEFAULT_EMP_CD", "")
}

func getDefaultBankCD() string {
	return getEnvString("DEFAULT_BANK_CD", "")
}

func getDefaultBANB() string {
	return getEnvString("DEFAULT_BA_NB", "")
}

func getDefaultAttrCD() string {
	return getEnvString("DEFAULT_ATTR_CD", DEFAULT_ATTR_CD)
}

// 파일 형식 검증
func isSupportedImageFormat(filename string) bool {
	filename = strings.ToLower(filename)
	for _, format := range SUPPORTED_IMAGE_FORMATS {
		if strings.HasSuffix(filename, format) {
			return true
		}
	}
	return false
}

// 파일 확장자에서 형식 추출
func getImageFormat(filename string) string {
	filename = strings.ToLower(filename)

	if strings.HasSuffix(filename, ".jpeg") {
		return "jpeg"
	}
	if strings.HasSuffix(filename, ".jpg") {
		return "jpg"
	}
	if strings.HasSuffix(filename, ".png") {
		return "png"
	}
	if strings.HasSuffix(filename, ".pdf") {
		return "pdf"
	}
	if strings.HasSuffix(filename, ".tiff") {
		return "tiff"
	}
	if strings.HasSuffix(filename, ".tif") {
		return "tif"
	}

	return "jpg" // 기본값
}

// 통합된 날짜 변환 함수
// 통합된 날짜 변환 함수
func convertDateToYYYYMMDD(dateText string) string {
	if dateText == "" {
		return dateText
	}

	patterns := []struct {
		regex *regexp.Regexp
		desc  string
	}{
		{regexp.MustCompile(`(\d{4})\.\s*(\d{1,2})\.\s*(\d{1,2})\.\s*\d{1,2}:\s*\d{1,2}:\s*\d{1,2}`), "YYYY. M. D. HH:MM:SS"},
		{regexp.MustCompile(`(\d{4})\.\s*(\d{1,2})\.\s*(\d{1,2})\.\s*\d{1,2}:\s*\d{1,2}`), "YYYY. M. D. HH:MM"},
		{regexp.MustCompile(`(\d{4})\.\s*(\d{1,2})\.\s*(\d{1,2})`), "YYYY. M. D"},
		{regexp.MustCompile(`(\d{4})\.(\d{2})\.(\d{2})`), "YYYY.MM.DD"},
		{regexp.MustCompile(`(\d{2})\.(\d{2})\.(\d{2})\s+\d{2}:\s*\d{2}:\s*\d{2}`), "YY.MM.DD HH:MM:SS"},
		{regexp.MustCompile(`(\d{2})\.(\d{2})\.(\d{2})\s+[I|]\s+\d{2}:\s*\d{2}:\s*\d{2}`), "YY.MM.DD I HH:MM:SS"},
		{regexp.MustCompile(`(\d{2})\.(\d{2})\.(\d{2})\s+\d{2}:\s*\d{2}`), "YY.MM.DD HH:MM"},
		{regexp.MustCompile(`(\d{2})\.(\d{2})\.(\d{2})\d{2}:\d{2}`), "YY.MM.DDHH:MM"},
		{regexp.MustCompile(`^(\d{2})\.(\d{2})\.(\d{2})$`), "YY.MM.DD"},
	}

	for _, pattern := range patterns {
		if matches := pattern.regex.FindStringSubmatch(dateText); len(matches) >= 4 {
			year := matches[1]
			month := matches[2]
			day := matches[3]

			// YY 형식을 YYYY로 변환
			if len(year) == 2 {
				year = "20" + year
			}

			// 월, 일을 2자리로 패딩
			month = fmt.Sprintf("%02s", month)
			day = fmt.Sprintf("%02s", day)

			return year + month + day
		}
	}

	// YYYYMMDD 형식이 이미 있는 경우
	if matches := regexp.MustCompile(`(\d{8})`).FindStringSubmatch(dateText); len(matches) >= 1 {
		return matches[0]
	}

	return dateText
}
