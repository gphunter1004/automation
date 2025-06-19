package main

import "regexp"

// DateTimePattern 날짜/시간 패턴 정의 구조체
type DateTimePattern struct {
	Regex       *regexp.Regexp
	Description string
	YearIndex   int // 년도 그룹 인덱스
	MonthIndex  int // 월 그룹 인덱스
	DayIndex    int // 일 그룹 인덱스
	HourIndex   int // 시간 그룹 인덱스 (-1이면 없음)
	MinuteIndex int // 분 그룹 인덱스 (-1이면 없음)
	SecondIndex int // 초 그룹 인덱스 (-1이면 없음)
}

// SimpleTimePattern 간단한 시간 패턴 (시간만 있는 경우)
type SimpleTimePattern struct {
	Regex       *regexp.Regexp
	Description string
	HourIndex   int
}

func GetDateTimePatterns() []DateTimePattern {
	return []DateTimePattern{
		// 🆕 새로운 패턴 추가 - YYYY.MM.DDHH:MM:SS (우선순위 높음)
		{
			Regex:       regexp.MustCompile(`(\d{4})\.(\d{2})\.(\d{2})(\d{2}):(\d{2}):(\d{2})`),
			Description: "YYYY.MM.DDHH:MM:SS",
			YearIndex:   1,
			MonthIndex:  2,
			DayIndex:    3,
			HourIndex:   4,
			MinuteIndex: 5,
			SecondIndex: 6,
		},
		// 🆕 새로운 패턴 추가 - YYYY.MM.DDHH:MM (분까지만)
		{
			Regex:       regexp.MustCompile(`(\d{4})\.(\d{2})\.(\d{2})(\d{2}):(\d{2})`),
			Description: "YYYY.MM.DDHH:MM",
			YearIndex:   1,
			MonthIndex:  2,
			DayIndex:    3,
			HourIndex:   4,
			MinuteIndex: 5,
			SecondIndex: -1,
		},
		// 기존 패턴들 (우선순위 순서대로)
		{
			Regex:       regexp.MustCompile(`(\d{4})\.\s*(\d{1,2})\.\s*(\d{1,2})\.\s*(\d{1,2}):\s*(\d{1,2}):\s*(\d{1,2})`),
			Description: "YYYY. M. D. HH:MM:SS",
			YearIndex:   1,
			MonthIndex:  2,
			DayIndex:    3,
			HourIndex:   4,
			MinuteIndex: 5,
			SecondIndex: 6,
		},
		{
			Regex:       regexp.MustCompile(`(\d{4})\.\s*(\d{1,2})\.\s*(\d{1,2})\.\s*(\d{1,2}):\s*(\d{1,2})`),
			Description: "YYYY. M. D. HH:MM",
			YearIndex:   1,
			MonthIndex:  2,
			DayIndex:    3,
			HourIndex:   4,
			MinuteIndex: 5,
			SecondIndex: -1,
		},
		{
			Regex:       regexp.MustCompile(`(\d{4})\.\s*(\d{1,2})\.\s*(\d{1,2})`),
			Description: "YYYY. M. D",
			YearIndex:   1,
			MonthIndex:  2,
			DayIndex:    3,
			HourIndex:   -1,
			MinuteIndex: -1,
			SecondIndex: -1,
		},
		{
			Regex:       regexp.MustCompile(`(\d{4})\.(\d{2})\.(\d{2})`),
			Description: "YYYY.MM.DD",
			YearIndex:   1,
			MonthIndex:  2,
			DayIndex:    3,
			HourIndex:   -1,
			MinuteIndex: -1,
			SecondIndex: -1,
		},
		{
			Regex:       regexp.MustCompile(`(\d{2})\.(\d{2})\.(\d{2})\s+(\d{1,2}):\s*(\d{1,2}):\s*(\d{1,2})`),
			Description: "YY.MM.DD HH:MM:SS",
			YearIndex:   1,
			MonthIndex:  2,
			DayIndex:    3,
			HourIndex:   4,
			MinuteIndex: 5,
			SecondIndex: 6,
		},
		{
			Regex:       regexp.MustCompile(`(\d{2})\.(\d{2})\.(\d{2})\s+[I|]\s+(\d{1,2}):\s*(\d{1,2}):\s*(\d{1,2})`),
			Description: "YY.MM.DD I HH:MM:SS",
			YearIndex:   1,
			MonthIndex:  2,
			DayIndex:    3,
			HourIndex:   4,
			MinuteIndex: 5,
			SecondIndex: 6,
		},
		{
			Regex:       regexp.MustCompile(`(\d{2})\.(\d{2})\.(\d{2})\s+(\d{1,2}):\s*(\d{1,2})`),
			Description: "YY.MM.DD HH:MM",
			YearIndex:   1,
			MonthIndex:  2,
			DayIndex:    3,
			HourIndex:   4,
			MinuteIndex: 5,
			SecondIndex: -1,
		},
		{
			Regex:       regexp.MustCompile(`(\d{2})\.(\d{2})\.(\d{2})(\d{2}):(\d{2})`),
			Description: "YY.MM.DDHH:MM",
			YearIndex:   1,
			MonthIndex:  2,
			DayIndex:    3,
			HourIndex:   4,
			MinuteIndex: 5,
			SecondIndex: -1,
		},
		{
			Regex:       regexp.MustCompile(`^(\d{2})\.(\d{2})\.(\d{2})$`),
			Description: "YY.MM.DD",
			YearIndex:   1,
			MonthIndex:  2,
			DayIndex:    3,
			HourIndex:   -1,
			MinuteIndex: -1,
			SecondIndex: -1,
		},
		// 추가 패턴들 - 확장 가능
		{
			Regex:       regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2})\s+(\d{2}):(\d{2}):(\d{2})`),
			Description: "YYYY-MM-DD HH:MM:SS (ISO 8601)",
			YearIndex:   1,
			MonthIndex:  2,
			DayIndex:    3,
			HourIndex:   4,
			MinuteIndex: 5,
			SecondIndex: 6,
		},
		{
			Regex:       regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2})`),
			Description: "YYYY-MM-DD (ISO 8601)",
			YearIndex:   1,
			MonthIndex:  2,
			DayIndex:    3,
			HourIndex:   -1,
			MinuteIndex: -1,
			SecondIndex: -1,
		},
		{
			Regex:       regexp.MustCompile(`(\d{4})/(\d{2})/(\d{2})\s+(\d{2}):(\d{2}):(\d{2})`),
			Description: "YYYY/MM/DD HH:MM:SS",
			YearIndex:   1,
			MonthIndex:  2,
			DayIndex:    3,
			HourIndex:   4,
			MinuteIndex: 5,
			SecondIndex: 6,
		},
		{
			Regex:       regexp.MustCompile(`(\d{4})/(\d{2})/(\d{2})`),
			Description: "YYYY/MM/DD",
			YearIndex:   1,
			MonthIndex:  2,
			DayIndex:    3,
			HourIndex:   -1,
			MinuteIndex: -1,
			SecondIndex: -1,
		},
	}
}

// GetSimpleTimePatterns 시간만 있는 간단한 패턴들을 반환
func GetSimpleTimePatterns() []SimpleTimePattern {
	return []SimpleTimePattern{
		{
			Regex:       regexp.MustCompile(`\b(\d{1,2}):\d{1,2}:\d{1,2}\b`),
			Description: "HH:MM:SS",
			HourIndex:   1,
		},
		{
			Regex:       regexp.MustCompile(`\b(\d{1,2}):\d{1,2}\b`),
			Description: "HH:MM",
			HourIndex:   1,
		},
		{
			Regex:       regexp.MustCompile(`\b(\d{1,2})시\s*\d{1,2}분\s*\d{1,2}초`),
			Description: "HH시MM분SS초",
			HourIndex:   1,
		},
		{
			Regex:       regexp.MustCompile(`\b(\d{1,2})시\s*\d{1,2}분`),
			Description: "HH시MM분",
			HourIndex:   1,
		},
		{
			Regex:       regexp.MustCompile(`\b(\d{1,2})시\b`),
			Description: "HH시",
			HourIndex:   1,
		},
		{
			Regex:       regexp.MustCompile(`오전\s*(\d{1,2}):\d{1,2}`),
			Description: "오전 HH:MM",
			HourIndex:   1,
		},
		{
			Regex:       regexp.MustCompile(`오후\s*(\d{1,2}):\d{1,2}`),
			Description: "오후 HH:MM (12시간 형식)",
			HourIndex:   1, // 주의: 오후 시간은 +12 처리 필요
		},
	}
}

// GetExistingDatePatterns YYYYMMDD 형식이 이미 있는 경우를 확인하는 패턴
func GetExistingDatePatterns() []*regexp.Regexp {
	return []*regexp.Regexp{
		regexp.MustCompile(`(\d{8})`),               // YYYYMMDD
		regexp.MustCompile(`(\d{4})(\d{2})(\d{2})`), // YYYYMMDD (그룹으로)
	}
}

// PatternPriority 패턴 우선순위 정의 (자주 사용되는 패턴을 앞에)
var PatternPriority = map[string]int{
	"YYYY.MM.DDHH:MM:SS":   1, // 🆕 새로운 패턴 최우선
	"YYYY.MM.DDHH:MM":      2, // 🆕 새로운 패턴 (분까지)
	"YY.MM.DD HH:MM:SS":    3, // 기존 가장 일반적
	"YY.MM.DD HH:MM":       4,
	"YY.MM.DD":             5,
	"YYYY. M. D. HH:MM:SS": 6,
	"YYYY. M. D. HH:MM":    7,
	"YYYY. M. D":           8,
	"YYYY.MM.DD":           9,
	"YY.MM.DD I HH:MM:SS":  10,
	"YY.MM.DDHH:MM":        11,
	"YYYY-MM-DD HH:MM:SS":  12, // ISO 8601
	"YYYY-MM-DD":           13,
	"YYYY/MM/DD HH:MM:SS":  14,
	"YYYY/MM/DD":           15,
}

// ValidateDateTime 날짜/시간 유효성 검증 함수
func ValidateDateTime(year, month, day, hour, minute, second int) bool {
	// 기본 범위 검증
	if year < 1900 || year > 2100 {
		return false
	}
	if month < 1 || month > 12 {
		return false
	}
	if day < 1 || day > 31 {
		return false
	}
	if hour < 0 || hour > 23 {
		return false
	}
	if minute < 0 || minute > 59 {
		return false
	}
	if second < 0 || second > 59 {
		return false
	}

	// 월별 일수 검증 (간단 버전)
	daysInMonth := []int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

	// 윤년 계산
	if month == 2 && isLeapYear(year) {
		if day > 29 {
			return false
		}
	} else if day > daysInMonth[month-1] {
		return false
	}

	return true
}

// isLeapYear 윤년 계산
func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}
