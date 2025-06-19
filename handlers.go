package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// OCR 처리 핸들러 (결과만 반환)
func handleOCRProcess(c *fiber.Ctx) error {
	// 폼 데이터 파싱
	var req UploadRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "폼 데이터 파싱 실패: " + err.Error(),
		})
	}

	// 다중 파일 업로드 처리
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "멀티파트 폼 파싱 실패: " + err.Error(),
		})
	}

	files := form.File["images"]
	if len(files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "이미지 파일을 찾을 수 없습니다",
		})
	}

	// 파일 개수 제한 체크
	maxFiles := getMaxFiles()
	if len(files) > maxFiles {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("한 번에 최대 %d개의 파일만 업로드할 수 있습니다", maxFiles),
		})
	}

	// 카테고리, 비고, 추가 이름, 국내출장 관련 정보 추출
	categories := make(map[int]string)
	remarks := make(map[int]string)
	additionalNames := make(map[int]string)
	businessContents := make(map[int]string)
	purposes := make(map[int]string)

	for i := 0; i < len(files); i++ {
		categoryKey := fmt.Sprintf("category_%d", i)
		if categoryValues, exists := form.Value[categoryKey]; exists && len(categoryValues) > 0 {
			categories[i] = categoryValues[0]
		} else {
			categories[i] = "6130" // 기본값: 석식
		}

		remarksKey := fmt.Sprintf("remarks_%d", i)
		if remarksValues, exists := form.Value[remarksKey]; exists && len(remarksValues) > 0 {
			remarks[i] = remarksValues[0]
		} else {
			remarks[i] = "" // 기본값: 빈 문자열
		}

		additionalNamesKey := fmt.Sprintf("additional_names_%d", i)
		if additionalNamesValues, exists := form.Value[additionalNamesKey]; exists && len(additionalNamesValues) > 0 {
			additionalNames[i] = additionalNamesValues[0]
		} else {
			additionalNames[i] = "" // 기본값: 빈 문자열
		}

		// 국내출장 전용 필드들
		businessContentKey := fmt.Sprintf("business_content_%d", i)
		if businessContentValues, exists := form.Value[businessContentKey]; exists && len(businessContentValues) > 0 {
			businessContents[i] = businessContentValues[0]
		} else {
			businessContents[i] = "" // 기본값: 빈 문자열
		}

		purposeKey := fmt.Sprintf("purpose_%d", i)
		if purposeValues, exists := form.Value[purposeKey]; exists && len(purposeValues) > 0 {
			purposes[i] = purposeValues[0]
		} else {
			purposes[i] = "" // 기본값: 빈 문자열
		}
	}

	// 모든 파일 읽기 및 준비
	var imageFiles []ImageFileWithCategory
	for i, file := range files {
		// 파일 형식 검증
		if !isSupportedImageFormat(file.Filename) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("파일 '%s'는 지원하지 않는 형식입니다. (jpg, jpeg, png, pdf, tif, tiff만 지원)", file.Filename),
			})
		}

		// 파일 읽기
		fileReader, err := file.Open()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("파일 '%s' 읽기 실패: %v", file.Filename, err),
			})
		}

		fileData, err := io.ReadAll(fileReader)
		fileReader.Close()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("파일 '%s' 데이터 읽기 실패: %v", file.Filename, err),
			})
		}

		imageFiles = append(imageFiles, ImageFileWithCategory{
			ImageFile: ImageFile{
				Data:     fileData,
				Filename: file.Filename,
			},
			Category: categories[i],
			Remarks:  remarks[i],
		})
	}

	// 비동기 OCR 처리
	ocrService := NewOCRService()
	ocrResults, err := ocrService.ProcessMultipleImagesAsyncWithCategory(imageFiles)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "OCR 처리 실패: " + err.Error(),
		})
	}

	// OCR 결과를 프론트엔드용 데이터로 변환
	excelService := NewExcelService()
	results := make([]OCRResult, 0)

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
		ocrServiceInstance := NewOCRService()

		// 필드에서 값 추출
		purpose := ocrServiceInstance.ExtractFieldValue(image.Fields, "사용처")
		amount := ocrServiceInstance.ExtractFieldValue(image.Fields, "사용액")
		amount = excelService.cleanAmountText(amount)

		issueDate := ocrServiceInstance.ExtractFieldValue(image.Fields, "사용일")
		issueDate = excelService.convertDateFormat(issueDate)

		payDate := excelService.calculatePaymentDate()

		// 비고 생성 로직
		var remark string

		if result.Category == "6320" { // 국내출장
			// 국내출장: 출장내용_이름,추가이름_용도 형식
			businessContent := businessContents[result.ImageIndex]
			purposeText := purposes[result.ImageIndex]

			finalUserName := req.UserName
			if additionalNames[result.ImageIndex] != "" {
				finalUserName = fmt.Sprintf("%s,%s", req.UserName, additionalNames[result.ImageIndex])
			}

			remark = fmt.Sprintf("%s_%s_%s", businessContent, finalUserName, purposeText)
			log.Printf("✅ 국내출장 이미지 %s: 자동 생성된 비고 = %s",
				result.ImageName, remark)
		} else {
			// 일반 카테고리 처리
			// 사용자가 직접 입력한 비고가 있고, 임시 비고가 아닌 경우만 사용
			if result.Remarks != "" &&
				!strings.Contains(result.Remarks, "임시_") &&
				!strings.Contains(result.Remarks, "MM/DD") {
				remark = result.Remarks
			} else {
				// 이름 조합 생성 (사용자이름,추가이름)
				finalUserName := req.UserName
				if additionalNames[result.ImageIndex] != "" {
					finalUserName = fmt.Sprintf("%s,%s", req.UserName, additionalNames[result.ImageIndex])
				}
				// 실제 OCR에서 추출된 사용일로 비고 생성
				remark = excelService.generateDefaultRemark(issueDate, finalUserName, result.Category)
				log.Printf("✅ 일반 이미지 %s: 자동 생성된 비고 = %s (사용일: %s)",
					result.ImageName, remark, issueDate)
			}
		}

		ocrResult := OCRResult{
			FileName:  result.ImageName,
			Category:  result.Category,
			Remark:    remark,
			Purpose:   purpose,
			Amount:    amount,
			IssueDate: issueDate,
			PayDate:   payDate,
		}

		results = append(results, ocrResult)
	}

	if len(results) == 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "모든 이미지의 OCR 처리에 실패했습니다",
		})
	}

	// ISS_DT (사용일) 기준으로 오름차순 정렬 (엑셀과 동일하게)
	log.Printf("=== OCR 결과 ISS_DT 기준 정렬 시작 ===")
	log.Printf("정렬 전 순서:")
	for i, result := range results {
		log.Printf("  %d. %s (ISS_DT: %s)", i+1, result.FileName, result.IssueDate)
	}

	sortOCRResultsByIssueDate(results)

	log.Printf("정렬 후 순서:")
	for i, result := range results {
		log.Printf("  %d. %s (ISS_DT: %s)", i+1, result.FileName, result.IssueDate)
	}
	log.Printf("=== OCR 결과 정렬 완료 ===")

	// JSON 결과 반환 (Excel 파일 생성하지 않음!)
	return c.JSON(OCRProcessResponse{
		Success: true,
		Message: fmt.Sprintf("%d개 파일의 OCR 처리가 완료되었습니다", len(results)),
		Data:    results,
	})
}

// OCR 결과를 IssueDate 기준으로 오름차순 정렬
func sortOCRResultsByIssueDate(results []OCRResult) {
	// 버블 정렬을 사용한 간단한 정렬 구현
	n := len(results)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			// IssueDate 문자열 비교 (YYYYMMDD 형식이므로 문자열 비교로 충분)
			if results[j].IssueDate > results[j+1].IssueDate {
				// 스왑
				results[j], results[j+1] = results[j+1], results[j]
			}
		}
	}
}

// Excel 다운로드 핸들러
func handleExcelDownload(c *fiber.Ctx) error {
	// 폼 데이터 파싱
	var req ExcelDownloadRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "폼 데이터 파싱 실패: " + err.Error(),
		})
	}

	// JSON 데이터 파싱
	var ocrResults []OCRResult
	if err := json.Unmarshal([]byte(req.ExcelData), &ocrResults); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "OCR 결과 데이터 파싱 실패: " + err.Error(),
		})
	}

	if len(ocrResults) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "변환할 데이터가 없습니다",
		})
	}

	// 환경변수에서 기본값 설정
	if req.DepositorDC == "" {
		req.DepositorDC = getDefaultDepositorDC()
	}
	if req.DeptCD == "" {
		req.DeptCD = getDefaultDeptCD()
	}
	if req.EmpCD == "" {
		req.EmpCD = getDefaultEmpCD()
	}
	if req.BankCD == "" {
		req.BankCD = getDefaultBankCD()
	}
	if req.BANB == "" {
		req.BANB = getDefaultBANB()
	}
	if req.AttrCD == "" {
		req.AttrCD = getDefaultAttrCD()
	}

	// OCR 결과를 Excel 데이터로 변환
	var allExcelData []*ExcelData
	for _, result := range ocrResults {
		excelData := &ExcelData{
			CASHCD:      result.Category,
			RMKDC:       result.Remark,
			TRNM:        result.Purpose,
			SUPAM:       result.Amount,
			VATAM:       "0",
			ATTRCD:      req.AttrCD, // 사용자 선택 또는 환경변수 기본값
			ISSDT:       result.IssueDate,
			PAYDT:       result.PayDate,
			BANKCD:      req.BankCD,
			BANB:        req.BANB,
			DEPOSITORDC: req.DepositorDC,
			DEPTCD:      req.DeptCD,
			EMPCD:       req.EmpCD,
		}
		allExcelData = append(allExcelData, excelData)
	}

	// ISS_DT 기준으로 정렬
	excelService := NewExcelService()
	excelService.sortByISSDate(allExcelData)

	// Excel 파일 생성
	excelFile, err := excelService.CreateExcelFileWithMultipleData(allExcelData)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Excel 파일 생성 실패: " + err.Error(),
		})
	}
	defer excelFile.Close()

	// Excel 파일을 바이트 배열로 변환
	buffer, err := excelFile.WriteToBuffer()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Excel 파일 버퍼 생성 실패: " + err.Error(),
		})
	}

	// 파일 다운로드 응답
	filename := fmt.Sprintf("ocr_results_%d_files_%s.xlsx", len(allExcelData), time.Now().Format("20060102_150405"))

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Set("Content-Length", fmt.Sprintf("%d", buffer.Len()))

	return c.Send(buffer.Bytes())
}
