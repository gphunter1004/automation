package main

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// 환경변수 기본값 상수
const (
	DEFAULT_OCR_API_URL     = "https://xlqf1p9tsl.apigw.ntruss.com/custom/v1/43074/fd2e3084b0aedaa21b114b92ed8d5b021e1f2c88b9eb7caea4b7e4b602aa85ea/infer"
	DEFAULT_SERVER_PORT     = "3000"
	DEFAULT_UPLOAD_LIMIT_MB = 50
	DEFAULT_MAX_FILES       = 5
	DEFAULT_OCR_TIMEOUT     = 60
	DEFAULT_ATTR_CD         = "8A" // 신용카드매출전표(개인)
)

// 지원하는 이미지 형식
var SUPPORTED_IMAGE_FORMATS = []string{".jpg", ".jpeg", ".png", ".pdf", ".tif", ".tiff"}

// 패턴 캐시 (성능 최적화)
var (
	dateTimePatterns     []DateTimePattern
	simpleTimePatterns   []SimpleTimePattern
	existingDatePatterns []*regexp.Regexp
	patternsInitialized  bool
)

// initializePatterns 패턴들을 초기화하고 우선순위에 따라 정렬
func initializePatterns() {
	if patternsInitialized {
		return
	}

	dateTimePatterns = GetDateTimePatterns()
	simpleTimePatterns = GetSimpleTimePatterns()
	existingDatePatterns = GetExistingDatePatterns()

	// 우선순위에 따라 정렬 (성능 최적화)
	sort.Slice(dateTimePatterns, func(i, j int) bool {
		priorityI, existsI := PatternPriority[dateTimePatterns[i].Description]
		priorityJ, existsJ := PatternPriority[dateTimePatterns[j].Description]

		if !existsI {
			priorityI = 999 // 우선순위가 없으면 맨 뒤로
		}
		if !existsJ {
			priorityJ = 999
		}

		return priorityI < priorityJ
	})

	patternsInitialized = true
}

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

// 시간 기반 카테고리 자동 재지정 활성화 여부 (기본값: true)
func isTimeCategoryEnabled() bool {
	value := os.Getenv("TIME_CATEGORY_ENABLED")
	if value == "" {
		return true // 기본값은 활성화
	}
	return getEnvBool("TIME_CATEGORY_ENABLED")
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
func convertDateToYYYYMMDD(dateText string) string {
	if dateText == "" {
		return dateText
	}

	// 패턴 초기화
	initializePatterns()

	// 기존 YYYYMMDD 형식 확인
	for _, pattern := range existingDatePatterns {
		if matches := pattern.FindStringSubmatch(dateText); len(matches) >= 1 {
			return matches[len(matches)-1] // 마지막 매치 반환
		}
	}

	// 공통 패턴을 사용하여 날짜 추출
	for _, pattern := range dateTimePatterns {
		if matches := pattern.Regex.FindStringSubmatch(dateText); len(matches) > pattern.DayIndex {
			year := matches[pattern.YearIndex]
			month := matches[pattern.MonthIndex]
			day := matches[pattern.DayIndex]

			// 숫자 변환 및 유효성 검사
			yearInt, _ := strconv.Atoi(year)
			monthInt, _ := strconv.Atoi(month)
			dayInt, _ := strconv.Atoi(day)

			// YY 형식을 YYYY로 변환
			if len(year) == 2 {
				if yearInt < 50 { // 2000년대로 가정
					yearInt = 2000 + yearInt
				} else { // 1900년대로 가정
					yearInt = 1900 + yearInt
				}
				year = strconv.Itoa(yearInt)
			}

			// 날짜 유효성 검증
			if !ValidateDateTime(yearInt, monthInt, dayInt, 0, 0, 0) {
				continue // 유효하지 않은 날짜면 다음 패턴 시도
			}

			// 월, 일을 2자리로 패딩
			month = fmt.Sprintf("%02s", month)
			day = fmt.Sprintf("%02s", day)

			return year + month + day
		}
	}

	return dateText
}

// 날짜/시간 문자열에서 시간(hour) 추출 (분리된 패턴 사용)
func extractHourFromDateTime(dateText string) int {
	if dateText == "" {
		return -1
	}

	// 패턴 초기화
	initializePatterns()

	// 주 패턴에서 시간 추출
	for _, pattern := range dateTimePatterns {
		if pattern.HourIndex == -1 {
			continue // 시간 정보가 없는 패턴은 건너뛰기
		}

		if matches := pattern.Regex.FindStringSubmatch(dateText); len(matches) > pattern.HourIndex {
			hourStr := matches[pattern.HourIndex]
			if hour, err := strconv.Atoi(hourStr); err == nil {

				// 오후 시간 처리 (12시간 형식)
				if strings.Contains(dateText, "오후") && hour < 12 {
					hour += 12
				}

				// 시간 유효성 검사
				if hour >= 0 && hour <= 23 {
					return hour
				}
			}
		}
	}

	// 간단한 시간 패턴 체크
	for _, pattern := range simpleTimePatterns {
		if matches := pattern.Regex.FindStringSubmatch(dateText); len(matches) > pattern.HourIndex {
			hourStr := matches[pattern.HourIndex]
			if hour, err := strconv.Atoi(hourStr); err == nil {

				// 오후 시간 처리
				if strings.Contains(pattern.Description, "오후") && hour < 12 {
					hour += 12
				}

				if hour >= 0 && hour <= 23 {
					return hour
				}
			}
		}
	}

	return -1
}

// 날짜/시간 문자열을 YYYY/MM/DD HH:MM 형식으로 변환
func formatDateTimeToStandard(dateText string) string {
	if dateText == "" {
		return ""
	}

	// 패턴 초기화
	initializePatterns()

	// 주 패턴에서 날짜/시간 추출
	for _, pattern := range dateTimePatterns {
		if matches := pattern.Regex.FindStringSubmatch(dateText); len(matches) > pattern.DayIndex {
			year := matches[pattern.YearIndex]
			month := matches[pattern.MonthIndex]
			day := matches[pattern.DayIndex]

			// 숫자 변환
			yearInt, _ := strconv.Atoi(year)
			monthInt, _ := strconv.Atoi(month)
			dayInt, _ := strconv.Atoi(day)

			// YY 형식을 YYYY로 변환
			if len(year) == 2 {
				if yearInt < 50 { // 2000년대로 가정
					yearInt = 2000 + yearInt
				} else { // 1900년대로 가정
					yearInt = 1900 + yearInt
				}
				year = strconv.Itoa(yearInt)
			}

			// 날짜 유효성 검증
			if !ValidateDateTime(yearInt, monthInt, dayInt, 0, 0, 0) {
				continue // 유효하지 않은 날짜면 다음 패턴 시도
			}

			// 월, 일을 2자리로 패딩
			month = fmt.Sprintf("%02d", monthInt)
			day = fmt.Sprintf("%02d", dayInt)

			result := fmt.Sprintf("%s/%s/%s", year, month, day)

			// 시간 정보가 있는 경우 추가
			if pattern.HourIndex != -1 && len(matches) > pattern.HourIndex {
				hourStr := matches[pattern.HourIndex]
				if hour, err := strconv.Atoi(hourStr); err == nil && hour >= 0 && hour <= 23 {
					minute := "00" // 기본값

					// 분 정보가 있는 경우
					if pattern.MinuteIndex != -1 && len(matches) > pattern.MinuteIndex {
						if minuteStr := matches[pattern.MinuteIndex]; minuteStr != "" {
							if min, err := strconv.Atoi(minuteStr); err == nil && min >= 0 && min <= 59 {
								minute = fmt.Sprintf("%02d", min)
							}
						}
					}

					// 오후 시간 처리
					if strings.Contains(dateText, "오후") && hour < 12 {
						hour += 12
					}

					result += fmt.Sprintf(" %02d:%s", hour, minute)
				}
			}

			return result
		}
	}

	// 간단한 시간만 있는 패턴 처리 (현재 날짜 기준)
	for _, pattern := range simpleTimePatterns {
		if matches := pattern.Regex.FindStringSubmatch(dateText); len(matches) > pattern.HourIndex {
			hourStr := matches[pattern.HourIndex]
			if hour, err := strconv.Atoi(hourStr); err == nil && hour >= 0 && hour <= 23 {
				// 현재 날짜 사용
				now := time.Now()
				result := fmt.Sprintf("%04d/%02d/%02d", now.Year(), now.Month(), now.Day())

				minute := "00"
				// 분 정보 추출 시도
				if minuteMatch := regexp.MustCompile(`(\d{1,2}):(\d{1,2})`).FindStringSubmatch(dateText); len(minuteMatch) >= 3 {
					if min, err := strconv.Atoi(minuteMatch[2]); err == nil && min >= 0 && min <= 59 {
						minute = fmt.Sprintf("%02d", min)
					}
				}

				// 오후 시간 처리
				if strings.Contains(pattern.Description, "오후") && hour < 12 {
					hour += 12
				}

				result += fmt.Sprintf(" %02d:%s", hour, minute)
				return result
			}
		}
	}

	// 변환할 수 없는 경우 원본 반환
	return dateText
}
