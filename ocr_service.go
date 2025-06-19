package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// OCR API 호출 서비스
type OCRService struct {
	apiURL    string
	secretKey string
	timeout   time.Duration
}

// OCR 서비스 생성자
func NewOCRService() *OCRService {
	timeoutSeconds := getOCRTimeoutSeconds()
	return &OCRService{
		apiURL:    getOCRAPIURL(),
		secretKey: getOCRSecret(),
		timeout:   time.Duration(timeoutSeconds) * time.Second,
	}
}

// 카테고리와 함께 여러 이미지를 비동기로 개별 OCR API 호출하고 결과 반환
func (s *OCRService) ProcessMultipleImagesAsyncWithCategory(imageFiles []ImageFileWithCategory) ([]*SingleImageOCRResultWithCategory, error) {
	if len(imageFiles) == 0 {
		return nil, fmt.Errorf("처리할 이미지가 없습니다")
	}

	log.Printf("=== 비동기 OCR 처리 시작 (카테고리 포함) ===")
	log.Printf("처리할 이미지 수: %d개 (각각 개별 API 호출)", len(imageFiles))

	var wg sync.WaitGroup
	results := make([]*SingleImageOCRResultWithCategory, len(imageFiles))
	startTime := time.Now()

	// 각 이미지에 대해 비동기로 OCR 처리
	for i, imageFileWithCategory := range imageFiles {
		wg.Add(1)
		go func(index int, imgFileWithCategory ImageFileWithCategory) {
			defer wg.Done()

			log.Printf("이미지 %d/%d 처리 시작: %s (카테고리: %s, 비고: %s)",
				index+1, len(imageFiles), imgFileWithCategory.ImageFile.Filename,
				imgFileWithCategory.Category, imgFileWithCategory.Remarks)

			// 개별 이미지 OCR 처리
			result, err := s.processSingleImage(imgFileWithCategory.ImageFile, index)
			if err != nil {
				log.Printf("❌ 이미지 %d (%s) 처리 실패: %v", index+1, imgFileWithCategory.ImageFile.Filename, err)
				results[index] = &SingleImageOCRResultWithCategory{
					SingleImageOCRResult: SingleImageOCRResult{
						ImageIndex: index,
						ImageName:  imgFileWithCategory.ImageFile.Filename,
						Response:   nil,
						Error:      err,
					},
					Category:        imgFileWithCategory.Category,
					Remarks:         imgFileWithCategory.Remarks,
					BusinessContent: imgFileWithCategory.BusinessContent,
					Purpose:         imgFileWithCategory.Purpose,
				}
				return
			}

			log.Printf("✅ 이미지 %d (%s) 처리 완료", index+1, imgFileWithCategory.ImageFile.Filename)
			results[index] = &SingleImageOCRResultWithCategory{
				SingleImageOCRResult: SingleImageOCRResult{
					ImageIndex: index,
					ImageName:  imgFileWithCategory.ImageFile.Filename,
					Response:   result,
					Error:      nil,
				},
				Category:        imgFileWithCategory.Category,
				Remarks:         imgFileWithCategory.Remarks,
				BusinessContent: imgFileWithCategory.BusinessContent,
				Purpose:         imgFileWithCategory.Purpose,
			}
		}(i, imageFileWithCategory)
	}

	// 모든 고루틴 완료 대기
	wg.Wait()
	totalDuration := time.Since(startTime)

	// 결과 요약
	successCount := 0
	failCount := 0
	for _, result := range results {
		if result.Error == nil {
			successCount++
		} else {
			failCount++
		}
	}

	log.Printf("=== 비동기 OCR 처리 완료 ===")
	log.Printf("총 소요 시간: %v", totalDuration)
	log.Printf("성공: %d개, 실패: %d개", successCount, failCount)

	if successCount == 0 {
		return nil, fmt.Errorf("모든 이미지 OCR 처리에 실패했습니다")
	}

	return results, nil
}

// 단일 이미지 OCR 처리
func (s *OCRService) processSingleImage(imageFile ImageFile, index int) (*OCRImageResult, error) {
	log.Printf("이미지 %d: %s (크기: %d bytes)", index+1, imageFile.Filename, len(imageFile.Data))

	// Base64 인코딩
	encodedImage := base64.StdEncoding.EncodeToString(imageFile.Data)
	log.Printf("Base64 인코딩 완료: %d characters", len(encodedImage))

	// OCR 요청 데이터 생성 (단일 이미지)
	requestId := uuid.New().String()
	timestamp := time.Now().UnixMilli()

	ocrRequest := OCRRequest{
		Version:   "V2",
		RequestID: requestId,
		Timestamp: timestamp,
		Lang:      "ko",
		Images: []OCRImage{
			{
				Format: getImageFormat(imageFile.Filename),
				Name:   imageFile.Filename,
				Data:   encodedImage,
			},
		},
	}

	log.Printf("이미지 %d OCR 요청 - Request ID: %s", index+1, requestId)

	// JSON 직렬화
	requestBody, err := json.Marshal(ocrRequest)
	if err != nil {
		return nil, fmt.Errorf("OCR 요청 JSON 생성 실패: %v", err)
	}

	log.Printf("이미지 %d 요청 본문 크기: %d bytes", index+1, len(requestBody))

	// OCR_VERBOSE_LOG가 설정된 경우에만 전체 요청 JSON 출력
	if isOCRVerboseLog() {
		log.Printf("=== 이미지 %d 전체 OCR 요청 JSON ===", index+1)
		requestJson, _ := json.MarshalIndent(ocrRequest, "", "  ")
		log.Printf("%s", string(requestJson))
	}

	// HTTP 요청 생성
	req, err := http.NewRequest("POST", s.apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("HTTP 요청 생성 실패: %v", err)
	}

	// 헤더 설정
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-OCR-SECRET", s.secretKey)

	// HTTP 클라이언트로 요청 전송
	client := &http.Client{
		Timeout: s.timeout,
	}

	log.Printf("이미지 %d OCR API 호출 시작", index+1)
	requestStartTime := time.Now()

	resp, err := client.Do(req)
	requestDuration := time.Since(requestStartTime)

	if err != nil {
		log.Printf("❌ 이미지 %d OCR API 요청 실패 (소요시간: %v): %v", index+1, requestDuration, err)
		return nil, fmt.Errorf("OCR API 요청 실패: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("이미지 %d OCR API 응답 수신 (소요시간: %v)", index+1, requestDuration)
	log.Printf("이미지 %d HTTP 상태: %d %s", index+1, resp.StatusCode, resp.Status)

	// 응답 상태 코드 확인
	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ 이미지 %d OCR API 응답 오류: %d", index+1, resp.StatusCode)

		// 에러 응답 body도 로깅
		errorBody := make([]byte, 1024)
		n, _ := resp.Body.Read(errorBody)
		if n > 0 {
			log.Printf("이미지 %d 에러 응답 내용: %s", index+1, string(errorBody[:n]))
		}

		return nil, fmt.Errorf("OCR API 응답 오류: %d", resp.StatusCode)
	}

	// 응답 body 전체를 먼저 읽어서 로깅
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ 이미지 %d 응답 body 읽기 실패: %v", index+1, err)
		return nil, fmt.Errorf("응답 body 읽기 실패: %v", err)
	}

	log.Printf("이미지 %d 응답 body 크기: %d bytes", index+1, len(responseBody))

	// 응답 내용 로그 (OCR_VERBOSE_LOG 설정에 따라)
	if isOCRVerboseLog() {
		log.Printf("=== 이미지 %d 전체 OCR 응답 JSON ===", index+1)
		log.Printf("%s", string(responseBody))
	} else {
		/*
			// JSON 포맷팅해서 보기 좋게 출력
			var tempResponse interface{}
			if err := json.Unmarshal(responseBody, &tempResponse); err == nil {
				formattedJson, _ := json.MarshalIndent(tempResponse, "", "  ")
				log.Printf("=== 이미지 %d OCR 응답 JSON ===\n%s", index+1, string(formattedJson))
			} else {
				log.Printf("=== 이미지 %d OCR 응답 Raw Data ===\n%s", index+1, string(responseBody))
			}
		*/
		// 포매팅 없이 원본 그대로 출력
		log.Printf("=== 이미지 %d OCR 응답 JSON (원본) ===\n%s", index+1, string(responseBody))
	}

	// 응답 JSON 파싱
	var ocrResponse OCRResponse
	if err := json.Unmarshal(responseBody, &ocrResponse); err != nil {
		log.Printf("❌ 이미지 %d OCR 응답 파싱 실패: %v", index+1, err)
		return nil, fmt.Errorf("OCR 응답 파싱 실패: %v", err)
	}

	// 첫 번째 이미지 결과 반환
	if len(ocrResponse.Images) > 0 {
		imageResult := &ocrResponse.Images[0]

		// 결과 로깅
		log.Printf("이미지 %d (%s) OCR 결과:", index+1, imageResult.Name)
		log.Printf("  - 처리 결과: %s", imageResult.InferResult)
		log.Printf("  - 메시지: %s", imageResult.Message)
		log.Printf("  - 추출된 필드 수: %d", len(imageResult.Fields))

		if imageResult.InferResult == "SUCCESS" {
			for j, field := range imageResult.Fields {
				log.Printf("    필드 %d: %s = '%s' (신뢰도: %.4f)",
					j+1, field.Name, field.InferText, field.InferConfidence)
			}
		}

		return imageResult, nil
	}

	return nil, fmt.Errorf("OCR 응답에 이미지 결과가 없습니다")
}

// OCR 결과에서 필드 값 추출
func (s *OCRService) ExtractFieldValue(fields []Field, fieldName string) string {
	for _, field := range fields {
		if field.Name == fieldName {
			return field.InferText
		}
	}
	return ""
}

// 금액 텍스트를 숫자로 변환 (콤마 제거)
func (s *OCRService) parseAmount(amountText string) float64 {
	if amountText == "" {
		return 0
	}

	// 숫자와 콤마만 추출
	re := regexp.MustCompile(`[0-9,]+`)
	matches := re.FindString(amountText)
	if matches == "" {
		return 0
	}

	// 콤마 제거 후 숫자로 변환
	cleanAmount := strings.ReplaceAll(matches, ",", "")
	amount, err := strconv.ParseFloat(cleanAmount, 64)
	if err != nil {
		log.Printf("금액 파싱 실패: %s -> %v", amountText, err)
		return 0
	}

	return amount
}

// 사용액 계산 (사용액이 없으면 공급가 + 부가세로 계산)
func (s *OCRService) CalculateAmount(fields []Field) string {
	// 사용액 먼저 확인
	supAM := s.ExtractFieldValue(fields, "사용액")
	if supAM != "" && s.parseAmount(supAM) > 0 {
		log.Printf("사용액 필드 존재: %s", supAM)
		return supAM
	}

	// 사용액이 없거나 0이면 공급가 + 부가세로 계산
	gongAM := s.ExtractFieldValue(fields, "공급가")
	vatAM := s.ExtractFieldValue(fields, "부가세")

	log.Printf("사용액이 없거나 0원 -> 공급가: '%s', 부가세: '%s'", gongAM, vatAM)

	gongAmount := s.parseAmount(gongAM)
	vatAmount := s.parseAmount(vatAM)

	if gongAmount > 0 || vatAmount > 0 {
		totalAmount := gongAmount + vatAmount
		result := fmt.Sprintf("%.0f", totalAmount)
		log.Printf("계산된 사용액: 공급가(%.0f) + 부가세(%.0f) = %s", gongAmount, vatAmount, result)
		return result
	}

	log.Printf("공급가와 부가세 모두 없음 -> 사용액을 빈 값으로 반환")
	return ""
}
