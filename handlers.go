package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"sort"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"
)

// OCR 처리 핸들러
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
	if len(files) > getMaxFiles() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("한 번에 최대 %d개의 파일만 업로드할 수 있습니다", getMaxFiles()),
		})
	}

	// 메타데이터 추출 (통합 함수)
	metadata := extractFormMetadata(form, len(files))

	// 파일 처리 및 OCR 호출
	imageFiles, err := prepareImageFiles(files, metadata)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
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

	// 결과 변환
	results := convertOCRResults(ocrResults, req.UserName, metadata)
	if len(results) == 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "모든 이미지의 OCR 처리에 실패했습니다",
		})
	}

	// 날짜순 정렬
	sortResultsByIssueDate(results)

	return c.JSON(OCRProcessResponse{
		Success: true,
		Message: fmt.Sprintf("%d개 파일의 OCR 처리가 완료되었습니다", len(results)),
		Data:    results,
	})
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

	// 환경변수 기본값 설정
	setDefaultValues(&req.UploadRequest)

	// OCR 결과를 Excel 데이터로 변환
	allExcelData := convertResultsToExcelData(ocrResults, &req.UploadRequest)

	// Excel 파일 생성
	excelService := NewExcelService()
	excelFile, err := excelService.CreateExcelFileWithMultipleData(allExcelData)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Excel 파일 생성 실패: " + err.Error(),
		})
	}
	defer excelFile.Close()

	// 파일 다운로드 응답
	return sendExcelFile(c, excelFile, len(allExcelData))
}

// 헬퍼 함수들

// 폼 메타데이터 추출
func extractFormMetadata(form *multipart.Form, fileCount int) map[int]map[string]string {
	metadata := make(map[int]map[string]string)

	for i := 0; i < fileCount; i++ {
		metadata[i] = make(map[string]string)

		// 각 파일별 메타데이터 추출
		fields := []string{"category", "remarks", "additional_names", "business_content", "purpose"}
		for _, field := range fields {
			key := fmt.Sprintf("%s_%d", field, i)
			if values, exists := form.Value[key]; exists && len(values) > 0 {
				metadata[i][field] = values[0]
			} else {
				// 기본값 설정
				if field == "category" {
					metadata[i][field] = "6130" // 기본값: 석식
				} else {
					metadata[i][field] = ""
				}
			}
		}
	}

	return metadata
}

// 이미지 파일 준비
func prepareImageFiles(files []*multipart.FileHeader, metadata map[int]map[string]string) ([]ImageFileWithCategory, error) {
	var imageFiles []ImageFileWithCategory

	for i, file := range files {
		// 파일 형식 검증
		if !isSupportedImageFormat(file.Filename) {
			return nil, fmt.Errorf("파일 '%s'는 지원하지 않는 형식입니다", file.Filename)
		}

		// 파일 읽기
		fileReader, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("파일 '%s' 읽기 실패: %v", file.Filename, err)
		}

		fileData, err := io.ReadAll(fileReader)
		fileReader.Close()
		if err != nil {
			return nil, fmt.Errorf("파일 '%s' 데이터 읽기 실패: %v", file.Filename, err)
		}

		// 메타데이터 가져오기
		meta := metadata[i]
		imageFiles = append(imageFiles, ImageFileWithCategory{
			ImageFile: ImageFile{
				Data:     fileData,
				Filename: file.Filename,
			},
			Category:        meta["category"],
			Remarks:         meta["remarks"],
			BusinessContent: meta["business_content"],
			Purpose:         meta["purpose"],
		})
	}

	return imageFiles, nil
}

// OCR 결과 변환
func convertOCRResults(ocrResults []*SingleImageOCRResultWithCategory, userName string, metadata map[int]map[string]string) []OCRResult {
	var results []OCRResult
	excelService := NewExcelService()

	for _, result := range ocrResults {
		if result.SingleImageOCRResult.Error != nil || result.SingleImageOCRResult.Response == nil || result.SingleImageOCRResult.Response.InferResult != "SUCCESS" {
			log.Printf("❌ 이미지 %s: 처리 실패 (건너뜀)", result.SingleImageOCRResult.ImageName)
			continue
		}

		image := result.SingleImageOCRResult.Response
		ocrServiceInstance := NewOCRService()

		// 필드에서 값 추출
		purpose := ocrServiceInstance.ExtractFieldValue(image.Fields, "사용처")

		// 사용액 계산 (사용액이 없으면 공급가 + 부가세로 계산)
		amount := ocrServiceInstance.CalculateAmount(image.Fields)
		amount = excelService.cleanAmountText(amount)

		issueDate := ocrServiceInstance.ExtractFieldValue(image.Fields, "사용일")
		issueDate = excelService.convertDateFormat(issueDate)

		payDate := excelService.calculatePaymentDate()

		// 비고 생성
		remark := generateRemark(result, userName, metadata[result.SingleImageOCRResult.ImageIndex], issueDate, excelService)

		results = append(results, OCRResult{
			FileName:  result.SingleImageOCRResult.ImageName,
			Category:  result.Category,
			Remark:    remark,
			Purpose:   purpose,
			Amount:    amount,
			IssueDate: issueDate,
			PayDate:   payDate,
		})
	}

	return results
}

// 비고 생성 로직
func generateRemark(result *SingleImageOCRResultWithCategory, userName string, meta map[string]string, issueDate string, excelService *ExcelService) string {
	if result.Category == "6320" { // 국내출장
		businessContent := meta["business_content"]
		purposeText := meta["purpose"]

		finalUserName := userName
		if additionalNames := meta["additional_names"]; additionalNames != "" {
			finalUserName = fmt.Sprintf("%s,%s", userName, additionalNames)
		}

		return fmt.Sprintf("%s_%s_%s", businessContent, finalUserName, purposeText)
	}

	// 일반 카테고리 처리
	if result.Remarks != "" && !strings.Contains(result.Remarks, "임시_") && !strings.Contains(result.Remarks, "MM/DD") {
		return result.Remarks
	}

	// 자동 생성
	finalUserName := userName
	if additionalNames := meta["additional_names"]; additionalNames != "" {
		finalUserName = fmt.Sprintf("%s,%s", userName, additionalNames)
	}

	return excelService.generateDefaultRemark(issueDate, finalUserName, result.Category)
}

// 결과를 날짜순으로 정렬
func sortResultsByIssueDate(results []OCRResult) {
	sort.Slice(results, func(i, j int) bool {
		return results[i].IssueDate < results[j].IssueDate
	})
}

// 환경변수 기본값 설정
func setDefaultValues(req *UploadRequest) {
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
}

// OCR 결과를 Excel 데이터로 변환
func convertResultsToExcelData(ocrResults []OCRResult, req *UploadRequest) []*ExcelData {
	var allExcelData []*ExcelData

	for _, result := range ocrResults {
		excelData := &ExcelData{
			CASHCD:      result.Category,
			RMKDC:       result.Remark,
			TRNM:        result.Purpose,
			SUPAM:       result.Amount,
			VATAM:       "0",
			ATTRCD:      req.AttrCD,
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
	sort.Slice(allExcelData, func(i, j int) bool {
		return allExcelData[i].ISSDT < allExcelData[j].ISSDT
	})

	return allExcelData
}

// Excel 파일 전송
func sendExcelFile(c *fiber.Ctx, excelFile *excelize.File, dataCount int) error {
	buffer, err := excelFile.WriteToBuffer()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Excel 파일 버퍼 생성 실패: " + err.Error(),
		})
	}

	filename := fmt.Sprintf("ocr_results_%d_files_%s.xlsx", dataCount, time.Now().Format("20060102_150405"))

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Set("Content-Length", fmt.Sprintf("%d", buffer.Len()))

	return c.Send(buffer.Bytes())
}
