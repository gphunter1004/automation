package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// 환경변수에서 OCR Secret 키 가져오기
func getOCRSecret() string {
	secret := os.Getenv("X_OCR_SECRET")
	if secret == "" {
		panic("X_OCR_SECRET 환경변수가 설정되지 않았습니다")
	}
	return secret
}

// 환경변수에서 OCR API URL 가져오기
func getOCRAPIURL() string {
	url := os.Getenv("OCR_API_URL")
	if url == "" {
		// 기본값
		url = "https://xlqf1p9tsl.apigw.ntruss.com/custom/v1/43074/fd2e3084b0aedaa21b114b92ed8d5b021e1f2c88b9eb7caea4b7e4b602aa85ea/infer"
	}
	return url
}

// 디버그 로그 레벨 확인
func isDebugMode() bool {
	debug := strings.ToLower(os.Getenv("DEBUG_MODE"))
	return debug == "true" || debug == "1" || debug == "on"
}

// OCR 상세 로그 활성화 여부
func isOCRVerboseLog() bool {
	verbose := strings.ToLower(os.Getenv("OCR_VERBOSE_LOG"))
	return verbose == "true" || verbose == "1" || verbose == "on"
}

// 환경변수에서 기본 예금주명 가져오기
func getDefaultDepositorDC() string {
	return os.Getenv("DEFAULT_DEPOSITOR_DC")
}

// 환경변수에서 기본 부서코드 가져오기
func getDefaultDeptCD() string {
	return os.Getenv("DEFAULT_DEPT_CD")
}

// 환경변수에서 기본 사원코드 가져오기
func getDefaultEmpCD() string {
	return os.Getenv("DEFAULT_EMP_CD")
}

// 환경변수에서 기본 은행코드 가져오기
func getDefaultBankCD() string {
	return os.Getenv("DEFAULT_BANK_CD")
}

// 환경변수에서 기본 계좌번호 가져오기
func getDefaultBANB() string {
	return os.Getenv("DEFAULT_BA_NB")
}

// 환경변수에서 기본 ATTR_CD 가져오기 (사용자 입력이 없을 때만 사용)
func getDefaultAttrCD() string {
	attrCD := os.Getenv("DEFAULT_ATTR_CD")
	if attrCD == "" {
		return "8A" // 기본값: 신용카드매출전표(개인)
	}
	return attrCD
}

// 환경변수에서 서버 포트 가져오기
func getServerPort() string {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "3000" // 기본값
	}
	return ":" + port
}

// 환경변수에서 업로드 제한 크기 가져오기 (MB)
func getUploadLimitMB() int {
	limitStr := os.Getenv("UPLOAD_LIMIT_MB")
	if limitStr == "" {
		return 50 // 기본값 50MB
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return 50 // 기본값
	}
	return limit
}

// 환경변수에서 최대 파일 개수 가져오기
func getMaxFiles() int {
	maxStr := os.Getenv("MAX_FILES")
	if maxStr == "" {
		return 10 // 기본값
	}

	max, err := strconv.Atoi(maxStr)
	if err != nil {
		return 10 // 기본값
	}
	return max
}

// 환경변수에서 OCR 타임아웃 가져오기 (초)
func getOCRTimeoutSeconds() int {
	timeoutStr := os.Getenv("OCR_TIMEOUT_SECONDS")
	if timeoutStr == "" {
		return 60 // 기본값 60초
	}

	timeout, err := strconv.Atoi(timeoutStr)
	if err != nil {
		return 60 // 기본값
	}
	return timeout
}

// 지원하는 이미지 형식 확인
func isSupportedImageFormat(filename string) bool {
	supportedFormats := []string{".jpg", ".jpeg", ".png", ".pdf", ".tif", ".tiff"}

	for _, format := range supportedFormats {
		if len(filename) >= len(format) &&
			filename[len(filename)-len(format):] == format {
			return true
		}
	}
	return false
}

// 파일 확장자에서 형식 추출
func getImageFormat(filename string) string {
	if len(filename) >= 4 {
		ext := filename[len(filename)-4:]
		switch ext {
		case ".jpg":
			return "jpg"
		case "jpeg":
			return "jpeg"
		case ".png":
			return "png"
		case ".pdf":
			return "pdf"
		case ".tif":
			return "tif"
		case "tiff":
			return "tiff"
		}
	}
	if len(filename) >= 5 && filename[len(filename)-5:] == ".jpeg" {
		return "jpeg"
	}
	if len(filename) >= 5 && filename[len(filename)-5:] == ".tiff" {
		return "tiff"
	}
	return "jpg" // 기본값
}

// 공통 날짜 변환 함수 (모든 곳에서 사용)
// 입력: "2025. 4. 23. 18: 21:56", "2025.04.02", "25.04.09 17:59", "25.04.2510:08" 등
// 출력: "20250423", "20250402", "20250409", "20250425" 등 (YYYYMMDD 형식)
func convertDateToYYYYMMDD(dateText string) string {
	if dateText == "" {
		return dateText
	}

	// 1. YYYY. M. D. HH: MM:SS 형식 (예: "2025. 4. 23. 18: 21:56")
	re1 := regexp.MustCompile(`(\d{4})\.\s*(\d{1,2})\.\s*(\d{1,2})\.\s*\d{1,2}:\s*\d{1,2}:\s*\d{1,2}`)
	matches1 := re1.FindStringSubmatch(dateText)
	if len(matches1) >= 4 {
		year := matches1[1]
		month := fmt.Sprintf("%02s", matches1[2]) // 2자리로 패딩
		day := fmt.Sprintf("%02s", matches1[3])   // 2자리로 패딩
		return year + month + day
	}

	// 2. YYYY. M. D. HH: MM 형식 (예: "2025. 4. 23. 18: 21")
	re2 := regexp.MustCompile(`(\d{4})\.\s*(\d{1,2})\.\s*(\d{1,2})\.\s*\d{1,2}:\s*\d{1,2}`)
	matches2 := re2.FindStringSubmatch(dateText)
	if len(matches2) >= 4 {
		year := matches2[1]
		month := fmt.Sprintf("%02s", matches2[2]) // 2자리로 패딩
		day := fmt.Sprintf("%02s", matches2[3])   // 2자리로 패딩
		return year + month + day
	}

	// 3. YYYY. M. D 형식 (예: "2025. 4. 23")
	re3 := regexp.MustCompile(`(\d{4})\.\s*(\d{1,2})\.\s*(\d{1,2})`)
	matches3 := re3.FindStringSubmatch(dateText)
	if len(matches3) >= 4 {
		year := matches3[1]
		month := fmt.Sprintf("%02s", matches3[2]) // 2자리로 패딩
		day := fmt.Sprintf("%02s", matches3[3])   // 2자리로 패딩
		return year + month + day
	}

	// 4. YYYY.MM.DD 형식 (기존)
	re4 := regexp.MustCompile(`(\d{4})\.(\d{2})\.(\d{2})`)
	matches4 := re4.FindStringSubmatch(dateText)
	if len(matches4) >= 4 {
		year := matches4[1]
		month := matches4[2]
		day := matches4[3]
		return year + month + day
	}

	// 5. YY.MM.DD HH:MM 형식 (공백 있음, 예: "25.04.09 17:59")
	re5 := regexp.MustCompile(`(\d{2})\.(\d{2})\.(\d{2})\s+\d{2}:\d{2}`)
	matches5 := re5.FindStringSubmatch(dateText)
	if len(matches5) >= 4 {
		yearShort := matches5[1]
		month := matches5[2]
		day := matches5[3]

		// YY를 YYYY로 변환 (20xx년으로 가정)
		yearFull := "20" + yearShort
		return yearFull + month + day
	}

	// 6. YY.MM.DDHH:MM 형식 (공백 없음, 예: "25.04.2510:08")
	re6 := regexp.MustCompile(`(\d{2})\.(\d{2})\.(\d{2})\d{2}:\d{2}`)
	matches6 := re6.FindStringSubmatch(dateText)
	if len(matches6) >= 4 {
		yearShort := matches6[1]
		month := matches6[2]
		day := matches6[3]

		// YY를 YYYY로 변환 (20xx년으로 가정)
		yearFull := "20" + yearShort
		return yearFull + month + day
	}

	// 7. YY.MM.DD 형식 (시간 없음, 예: "25.04.09")
	re7 := regexp.MustCompile(`^(\d{2})\.(\d{2})\.(\d{2})$`)
	matches7 := re7.FindStringSubmatch(dateText)
	if len(matches7) >= 4 {
		yearShort := matches7[1]
		month := matches7[2]
		day := matches7[3]

		// YY를 YYYY로 변환 (20xx년으로 가정)
		yearFull := "20" + yearShort
		return yearFull + month + day
	}

	// 8. YYYYMMDD 형식이 이미 있는 경우
	re8 := regexp.MustCompile(`(\d{8})`)
	matches8 := re8.FindStringSubmatch(dateText)
	if len(matches8) >= 1 {
		return matches8[0]
	}

	return dateText
}
