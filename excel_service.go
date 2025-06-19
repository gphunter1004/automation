package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// Excel 서비스
type ExcelService struct{}

// Excel 서비스 생성자
func NewExcelService() *ExcelService {
	return &ExcelService{}
}

// 금액 텍스트 정리 (예: "32,300 원" -> "32300")
func (e *ExcelService) cleanAmountText(amountText string) string {
	// 숫자와 쉼표만 추출
	re := regexp.MustCompile("[0-9,]+")
	matches := re.FindAllString(amountText, -1)

	if len(matches) > 0 {
		// 쉼표 제거
		return strings.ReplaceAll(matches[0], ",", "")
	}

	return amountText
}

// 날짜 형식 변환 (공통 함수 사용)
func (e *ExcelService) convertDateFormat(dateText string) string {
	return convertDateToYYYYMMDD(dateText)
}

// 결제일 계산 (현재 날짜 기준 10일 이전이면 당월 15일, 이후면 다음달 15일)
func (e *ExcelService) calculatePaymentDate() string {
	now := time.Now()
	currentDay := now.Day()

	var paymentDate time.Time

	if currentDay <= 10 {
		// 현재 날짜가 10일 이전이면 당월 15일
		paymentDate = time.Date(now.Year(), now.Month(), 15, 0, 0, 0, 0, now.Location())
	} else {
		// 현재 날짜가 10일 이후면 다음달 15일
		nextMonth := now.AddDate(0, 1, 0)
		paymentDate = time.Date(nextMonth.Year(), nextMonth.Month(), 15, 0, 0, 0, 0, now.Location())
	}

	return paymentDate.Format("20060102")
}

// 기본 RMK_DC 생성 (MM/DD_이름_카테고리 형식)
func (e *ExcelService) generateDefaultRemark(issueDate, userName, category string) string {
	if userName == "" {
		return fmt.Sprintf("카테고리: %s", e.getCategoryLabel(category))
	}

	// YYYYMMDD 형식에서 MM/DD 추출
	monthDay := "MM/DD"
	if len(issueDate) >= 8 {
		month := issueDate[4:6]
		day := issueDate[6:8]
		monthDay = fmt.Sprintf("%s/%s", month, day)
	} else {
		// 공통 함수로 날짜 변환 후 MM/DD 추출
		convertedDate := convertDateToYYYYMMDD(issueDate)
		if len(convertedDate) >= 8 {
			month := convertedDate[4:6]
			day := convertedDate[6:8]
			monthDay = fmt.Sprintf("%s/%s", month, day)
		}
	}

	categoryLabel := e.getCategoryLabel(category)
	return fmt.Sprintf("%s_%s_%s", monthDay, userName, categoryLabel)
}

// ISS_DT 기준으로 오름차순 정렬
func (e *ExcelService) sortByISSDate(data []*ExcelData) {
	// 버블 정렬을 사용한 간단한 정렬 구현
	n := len(data)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			// ISS_DT 문자열 비교 (YYYYMMDD 형식이므로 문자열 비교로 충분)
			if data[j].ISSDT > data[j+1].ISSDT {
				// 스왑
				data[j], data[j+1] = data[j+1], data[j]
			}
		}
	}
}

// 카테고리 코드를 라벨로 변환
func (e *ExcelService) getCategoryLabel(category string) string {
	switch category {
	case "6110":
		return "조식"
	case "6120":
		return "중식"
	case "6130":
		return "석식"
	case "6310":
		return "교통정산"
	case "6320":
		return "국내출장"
	default:
		return category
	}
}

// 카테고리와 함께 비동기 OCR 결과를 Excel 데이터로 변환 (ISS_DT 정렬 포함)
func (e *ExcelService) ConvertAsyncOCRToExcelDataWithCategory(ocrResults []*SingleImageOCRResultWithCategory, userName, depositorDC, deptCD, empCD, bankCD, baNB, attrCD string) ([]*ExcelData, error) {
	if len(ocrResults) == 0 {
		return nil, fmt.Errorf("OCR 결과가 없습니다")
	}

	var allExcelData []*ExcelData
	ocrService := NewOCRService()

	log.Printf("=== Excel 데이터 변환 시작 (카테고리 포함) ===")

	// 성공한 OCR 결과들만 처리
	for _, result := range ocrResults {
		if result.Error != nil {
			log.Printf("❌ 이미지 %s: %v (건너뜀)", result.ImageName, result.Error)
			continue
		}

		if result.Response == nil || result.Response.InferResult != "SUCCESS" {
			log.Printf("❌ 이미지 %s: OCR 처리 실패 (건너뜀)", result.ImageName)
			continue
		}

		image := result.Response
		log.Printf("✅ 이미지 %s: Excel 데이터 변환 중 (카테고리: %s, 비고: %s)",
			image.Name, result.Category, result.Remarks)

		// 필드에서 값 추출
		// TR_NM: 사용처
		trNM := ocrService.ExtractFieldValue(image.Fields, "사용처")

		// SUP_AM: 사용액
		supAM := ocrService.ExtractFieldValue(image.Fields, "사용액")
		supAM = e.cleanAmountText(supAM) // 금액 텍스트 정리

		// ISS_DT: 사용일 (YYYYMMDD 형식으로 변환)
		issDT := ocrService.ExtractFieldValue(image.Fields, "사용일")
		issDT = e.convertDateFormat(issDT)

		// PAY_DT: 결제일 (현재 날짜 기준 계산)
		payDT := e.calculatePaymentDate()

		// RMK_DC: 비고가 있으면 비고 사용, 없으면 기본값 생성
		var rmkDC string
		if result.Remarks != "" {
			rmkDC = result.Remarks
		} else {
			rmkDC = e.generateDefaultRemark(issDT, userName, result.Category)
		}

		// ATTR_CD: 사용자 입력값 우선, 없으면 환경변수 기본값
		finalAttrCD := attrCD
		if finalAttrCD == "" {
			finalAttrCD = getDefaultAttrCD()
		}

		excelData := &ExcelData{
			CASHCD:      result.Category, // 카테고리 값을 CASH_CD에 설정
			RMKDC:       rmkDC,           // 비고 또는 기본값
			TRNM:        trNM,
			SUPAM:       supAM,
			VATAM:       "0",
			ATTRCD:      finalAttrCD, // 사용자 입력값 또는 환경변수 기본값
			ISSDT:       issDT,
			PAYDT:       payDT,
			BANKCD:      bankCD,      // 사용자 입력값 또는 환경변수 기본값
			BANB:        baNB,        // 사용자 입력값 또는 환경변수 기본값
			DEPOSITORDC: depositorDC, // 사용자 입력값 또는 환경변수 기본값
			DEPTCD:      deptCD,      // 사용자 입력값 또는 환경변수 기본값
			EMPCD:       empCD,       // 사용자 입력값 또는 환경변수 기본값
		}

		log.Printf("변환된 데이터 - 사용처: %s, 사용액: %s, 사용일: %s, 카테고리: %s, 비고: %s, RMK_DC: %s, ATTR_CD: %s",
			trNM, supAM, issDT, result.Category, result.Remarks, rmkDC, finalAttrCD)
		allExcelData = append(allExcelData, excelData)
	}

	if len(allExcelData) == 0 {
		return nil, fmt.Errorf("모든 이미지의 OCR 처리에 실패했습니다")
	}

	// ISS_DT 기준으로 오름차순 정렬
	log.Printf("=== ISS_DT 기준 정렬 시작 ===")
	log.Printf("정렬 전 순서:")
	for i, data := range allExcelData {
		log.Printf("  %d. %s (ISS_DT: %s, CASH_CD: %s)", i+1, data.RMKDC, data.ISSDT, data.CASHCD)
	}

	e.sortByISSDate(allExcelData)

	log.Printf("정렬 후 순서:")
	for i, data := range allExcelData {
		log.Printf("  %d. %s (ISS_DT: %s, CASH_CD: %s)", i+1, data.RMKDC, data.ISSDT, data.CASHCD)
	}
	log.Printf("=== 정렬 완료 ===")

	return allExcelData, nil
}

// 비동기 OCR 결과를 Excel 데이터로 변환 (ISS_DT 정렬 포함) - 기존 호환성
func (e *ExcelService) ConvertAsyncOCRToExcelData(ocrResults []*SingleImageOCRResult, depositorDC, deptCD, empCD, bankCD, baNB, attrCD string) ([]*ExcelData, error) {
	if len(ocrResults) == 0 {
		return nil, fmt.Errorf("OCR 결과가 없습니다")
	}

	var allExcelData []*ExcelData
	ocrService := NewOCRService()

	log.Printf("=== Excel 데이터 변환 시작 ===")

	// 성공한 OCR 결과들만 처리
	for _, result := range ocrResults {
		if result.Error != nil {
			log.Printf("❌ 이미지 %s: %v (건너뜀)", result.ImageName, result.Error)
			continue
		}

		if result.Response == nil || result.Response.InferResult != "SUCCESS" {
			log.Printf("❌ 이미지 %s: OCR 처리 실패 (건너뜀)", result.ImageName)
			continue
		}

		image := result.Response
		log.Printf("✅ 이미지 %s: Excel 데이터 변환 중", image.Name)

		// 필드에서 값 추출
		// TR_NM: 사용처
		trNM := ocrService.ExtractFieldValue(image.Fields, "사용처")

		// SUP_AM: 사용액
		supAM := ocrService.ExtractFieldValue(image.Fields, "사용액")
		supAM = e.cleanAmountText(supAM) // 금액 텍스트 정리

		// ISS_DT: 사용일 (YYYYMMDD 형식으로 변환)
		issDT := ocrService.ExtractFieldValue(image.Fields, "사용일")
		issDT = e.convertDateFormat(issDT)

		// PAY_DT: 결제일 (현재 날짜 기준 계산)
		payDT := e.calculatePaymentDate()

		// ATTR_CD: 사용자 입력값 우선, 없으면 환경변수 기본값
		finalAttrCD := attrCD
		if finalAttrCD == "" {
			finalAttrCD = getDefaultAttrCD()
		}

		excelData := &ExcelData{
			CASHCD:      "6130", // 기본값: 석식
			RMKDC:       fmt.Sprintf("파일%d: %s", result.ImageIndex+1, image.Name),
			TRNM:        trNM,
			SUPAM:       supAM,
			VATAM:       "0",
			ATTRCD:      finalAttrCD, // 사용자 입력값 또는 환경변수 기본값
			ISSDT:       issDT,
			PAYDT:       payDT,
			BANKCD:      bankCD,      // 사용자 입력값 또는 환경변수 기본값
			BANB:        baNB,        // 사용자 입력값 또는 환경변수 기본값
			DEPOSITORDC: depositorDC, // 사용자 입력값 또는 환경변수 기본값
			DEPTCD:      deptCD,      // 사용자 입력값 또는 환경변수 기본값
			EMPCD:       empCD,       // 사용자 입력값 또는 환경변수 기본값
		}

		log.Printf("변환된 데이터 - 사용처: %s, 사용액: %s, 사용일: %s", trNM, supAM, issDT)
		allExcelData = append(allExcelData, excelData)
	}

	if len(allExcelData) == 0 {
		return nil, fmt.Errorf("모든 이미지의 OCR 처리에 실패했습니다")
	}

	// ISS_DT 기준으로 오름차순 정렬
	log.Printf("=== ISS_DT 기준 정렬 시작 ===")
	log.Printf("정렬 전 순서:")
	for i, data := range allExcelData {
		log.Printf("  %d. %s (ISS_DT: %s)", i+1, data.RMKDC, data.ISSDT)
	}

	e.sortByISSDate(allExcelData)

	log.Printf("정렬 후 순서:")
	for i, data := range allExcelData {
		log.Printf("  %d. %s (ISS_DT: %s)", i+1, data.RMKDC, data.ISSDT)
	}
	log.Printf("=== 정렬 완료 ===")

	return allExcelData, nil
}

// 기존 함수 호환성 유지
func (e *ExcelService) ConvertMultipleOCRToExcelData(ocrResponse *OCRResponse, depositorDC, deptCD, empCD, bankCD, baNB, attrCD string) ([]*ExcelData, error) {
	if len(ocrResponse.Images) == 0 {
		return nil, fmt.Errorf("OCR 결과에 이미지가 없습니다")
	}

	var allExcelData []*ExcelData
	ocrService := NewOCRService()

	// 각 이미지 결과에 대해 Excel 데이터 생성
	for i, image := range ocrResponse.Images {
		if image.InferResult != "SUCCESS" {
			// 실패한 이미지는 건너뛰고 로그 출력
			fmt.Printf("이미지 '%s' OCR 처리 실패: %s (건너뜀)\n", image.Name, image.Message)
			continue
		}

		// 필드에서 값 추출
		// TR_NM: 사용처
		trNM := ocrService.ExtractFieldValue(image.Fields, "사용처")

		// SUP_AM: 사용액
		supAM := ocrService.ExtractFieldValue(image.Fields, "사용액")
		supAM = e.cleanAmountText(supAM) // 금액 텍스트 정리

		// ISS_DT: 사용일 (YYYYMMDD 형식으로 변환)
		issDT := ocrService.ExtractFieldValue(image.Fields, "사용일")
		issDT = e.convertDateFormat(issDT)

		// PAY_DT: 결제일 (현재 날짜 기준 계산)
		payDT := e.calculatePaymentDate()

		// ATTR_CD: 사용자 입력값 우선, 없으면 환경변수 기본값
		finalAttrCD := attrCD
		if finalAttrCD == "" {
			finalAttrCD = getDefaultAttrCD()
		}

		excelData := &ExcelData{
			CASHCD:      "6130", // 기본값: 석식
			RMKDC:       fmt.Sprintf("파일%d: %s", i+1, image.Name),
			TRNM:        trNM,
			SUPAM:       supAM,
			VATAM:       "0",
			ATTRCD:      finalAttrCD, // 사용자 입력값 또는 환경변수 기본값
			ISSDT:       issDT,
			PAYDT:       payDT,
			BANKCD:      bankCD,      // 사용자 입력값 또는 환경변수 기본값
			BANB:        baNB,        // 사용자 입력값 또는 환경변수 기본값
			DEPOSITORDC: depositorDC, // 사용자 입력값 또는 환경변수 기본값
			DEPTCD:      deptCD,      // 사용자 입력값 또는 환경변수 기본값
			EMPCD:       empCD,       // 사용자 입력값 또는 환경변수 기본값
		}

		allExcelData = append(allExcelData, excelData)
	}

	if len(allExcelData) == 0 {
		return nil, fmt.Errorf("모든 이미지의 OCR 처리에 실패했습니다")
	}

	// ISS_DT 기준으로 정렬
	e.sortByISSDate(allExcelData)

	return allExcelData, nil
}

// 다중 데이터로 Excel 파일 생성
func (e *ExcelService) CreateExcelFileWithMultipleData(dataList []*ExcelData) (*excelize.File, error) {
	f := excelize.NewFile()

	// 헤더 설정 (2행에 추가)
	headers := []string{
		"CASH_CD", "RMK_DC", "TR_NM", "SUP_AM", "VAT_AM",
		"ATTR_CD", "ISS_DT", "PAY_DT", "BANK_CD", "BA_NB",
		"DEPOSITOR_DC", "DEPT_CD", "EMP_CD",
	}

	// 2행에 헤더 추가
	for i, header := range headers {
		cellName := fmt.Sprintf("%c2", 'A'+i)
		f.SetCellValue("Sheet1", cellName, header)
	}

	// 4행부터 데이터 추가 (각 이미지의 OCR 결과를 행별로)
	for idx, data := range dataList {
		row := 4 + idx // 4행부터 시작

		dataValues := []interface{}{
			data.CASHCD, data.RMKDC, data.TRNM, data.SUPAM, data.VATAM,
			data.ATTRCD, data.ISSDT, data.PAYDT, data.BANKCD, data.BANB,
			data.DEPOSITORDC, data.DEPTCD, data.EMPCD,
		}

		for i, value := range dataValues {
			cellName := fmt.Sprintf("%c%d", 'A'+i, row)
			f.SetCellValue("Sheet1", cellName, value)
		}
	}

	// 헤더 스타일 적용
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"E0E0E0"},
			Pattern: 1,
		},
	})
	if err != nil {
		return nil, err
	}

	// 헤더에 스타일 적용
	f.SetCellStyle("Sheet1", "A2", "M2", headerStyle)

	// 열 너비 자동 조정
	for i := 0; i < len(headers); i++ {
		columnName := fmt.Sprintf("%c", 'A'+i)
		f.SetColWidth("Sheet1", columnName, columnName, 15)
	}

	return f, nil
}
