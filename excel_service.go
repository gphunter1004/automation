package main

import (
	"fmt"
	"log"
	"regexp"
	"sort"
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
	re := regexp.MustCompile("[0-9,]+")
	matches := re.FindAllString(amountText, -1)
	if len(matches) > 0 {
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
	var paymentDate time.Time

	if now.Day() <= 10 {
		paymentDate = time.Date(now.Year(), now.Month(), 15, 0, 0, 0, 0, now.Location())
	} else {
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

	monthDay := "MM/DD"
	convertedDate := convertDateToYYYYMMDD(issueDate)
	if len(convertedDate) >= 8 {
		month := convertedDate[4:6]
		day := convertedDate[6:8]
		monthDay = fmt.Sprintf("%s/%s", month, day)
	}

	categoryLabel := e.getCategoryLabel(category)
	return fmt.Sprintf("%s_%s_%s", monthDay, userName, categoryLabel)
}

// 카테고리 코드를 라벨로 변환
func (e *ExcelService) getCategoryLabel(category string) string {
	categoryLabels := map[string]string{
		"6110": "조식",
		"6120": "중식",
		"6130": "석식",
		"6310": "교통정산",
		"6320": "국내출장",
	}
	if label, exists := categoryLabels[category]; exists {
		return label
	}
	return category
}

// Excel 데이터 생성 (통합 함수)
func (e *ExcelService) createExcelData(
	category, remarks, trNM, supAM, attrCD, issDT, payDT,
	bankCD, baNB, depositorDC, deptCD, empCD string) *ExcelData {

	return &ExcelData{
		CASHCD:      category,
		RMKDC:       remarks,
		TRNM:        trNM,
		SUPAM:       e.cleanAmountText(supAM),
		VATAM:       "0",
		ATTRCD:      attrCD,
		ISSDT:       e.convertDateFormat(issDT),
		PAYDT:       payDT,
		BANKCD:      bankCD,
		BANB:        baNB,
		DEPOSITORDC: depositorDC,
		DEPTCD:      deptCD,
		EMPCD:       empCD,
	}
}

// 카테고리와 함께 비동기 OCR 결과를 Excel 데이터로 변환
func (e *ExcelService) ConvertAsyncOCRToExcelDataWithCategory(ocrResults []*SingleImageOCRResultWithCategory, userName, depositorDC, deptCD, empCD, bankCD, baNB, attrCD string) ([]*ExcelData, error) {
	if len(ocrResults) == 0 {
		return nil, fmt.Errorf("OCR 결과가 없습니다")
	}

	var allExcelData []*ExcelData
	ocrService := NewOCRService()
	payDT := e.calculatePaymentDate()

	log.Printf("=== Excel 데이터 변환 시작 (카테고리 포함) ===")

	// 성공한 OCR 결과들만 처리
	for _, result := range ocrResults {
		if result.SingleImageOCRResult.Error != nil || result.SingleImageOCRResult.Response == nil || result.SingleImageOCRResult.Response.InferResult != "SUCCESS" {
			log.Printf("❌ 이미지 %s: 처리 실패 (건너뜀)", result.SingleImageOCRResult.ImageName)
			continue
		}

		image := result.SingleImageOCRResult.Response
		log.Printf("✅ 이미지 %s: Excel 데이터 변환 중 (카테고리: %s)", image.Name, result.Category)

		// 필드에서 값 추출
		trNM := ocrService.ExtractFieldValue(image.Fields, "사용처")
		supAM := ocrService.ExtractFieldValue(image.Fields, "사용액")
		issDT := ocrService.ExtractFieldValue(image.Fields, "사용일")

		// 비고 생성
		var rmkDC string
		if result.Remarks != "" {
			rmkDC = result.Remarks
		} else {
			rmkDC = e.generateDefaultRemark(issDT, userName, result.Category)
		}

		// ATTR_CD 설정
		finalAttrCD := attrCD
		if finalAttrCD == "" {
			finalAttrCD = getDefaultAttrCD()
		}

		excelData := e.createExcelData(
			result.Category, rmkDC, trNM, supAM, finalAttrCD,
			issDT, payDT, bankCD, baNB, depositorDC, deptCD, empCD,
		)

		log.Printf("변환된 데이터 - 사용처: %s, 사용액: %s, 사용일: %s, 카테고리: %s",
			trNM, excelData.SUPAM, excelData.ISSDT, result.Category)
		allExcelData = append(allExcelData, excelData)
	}

	if len(allExcelData) == 0 {
		return nil, fmt.Errorf("모든 이미지의 OCR 처리에 실패했습니다")
	}

	// ISS_DT 기준으로 오름차순 정렬
	e.sortByISSDate(allExcelData)
	log.Printf("=== 정렬 완료: %d개 데이터 ===", len(allExcelData))

	return allExcelData, nil
}

// ISS_DT 기준으로 오름차순 정렬 (Go 표준 sort 사용)
func (e *ExcelService) sortByISSDate(data []*ExcelData) {
	sort.Slice(data, func(i, j int) bool {
		return data[i].ISSDT < data[j].ISSDT
	})
}

// 다중 데이터로 Excel 파일 생성
func (e *ExcelService) CreateExcelFileWithMultipleData(dataList []*ExcelData) (*excelize.File, error) {
	f := excelize.NewFile()

	// 헤더 설정
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

	// 4행부터 데이터 추가
	for idx, data := range dataList {
		row := 4 + idx
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
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"E0E0E0"}, Pattern: 1},
	})
	if err != nil {
		return nil, err
	}

	f.SetCellStyle("Sheet1", "A2", "M2", headerStyle)

	// 열 너비 조정
	for i := 0; i < len(headers); i++ {
		columnName := fmt.Sprintf("%c", 'A'+i)
		f.SetColWidth("Sheet1", columnName, columnName, 15)
	}

	return f, nil
}
