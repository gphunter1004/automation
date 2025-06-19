package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

// 라우트 설정
func setupRoutes(app *fiber.App) {
	// API 라우트 그룹
	api := app.Group("/api")

	// 새로운 2단계 플로우 엔드포인트
	api.Post("/process-ocr", func(c *fiber.Ctx) error {
		log.Printf("🎯 /api/process-ocr 엔드포인트 호출됨")
		return handleOCRProcess(c)
	})

	api.Post("/download-excel", func(c *fiber.Ctx) error {
		log.Printf("📄 /api/download-excel 엔드포인트 호출됨")
		return handleExcelDownload(c)
	})

	// 헬스 체크 엔드포인트
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "OCR to Excel API is running",
			"version": "2.0",
			"flow": map[string]interface{}{
				"two_step_process": map[string]string{
					"step1": "POST /api/process-ocr",
					"step2": "POST /api/download-excel",
				},
			},
		})
	})

	// 정적 파일 제공 (CSS, JS, 이미지 등)
	app.Static("/", "./static")

	// 기본 라우트 (index.html 제공)
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile("./static/index.html")
	})

	// API 문서 엔드포인트
	api.Get("/info", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"name":        "OCR to Excel API",
			"version":     "2.0.0",
			"description": "이미지 OCR 처리 및 Excel 변환 서비스",
			"endpoints": map[string]interface{}{
				"process_ocr": map[string]interface{}{
					"method":      "POST",
					"path":        "/api/process-ocr",
					"description": "이미지 OCR 처리 후 JSON 결과 반환",
					"params": []string{
						"user_name (required)",
						"images[] (required)",
						"category_N (optional)",
						"remarks_N (optional)",
						"depositor_dc, dept_cd, emp_cd, bank_cd, ba_nb (optional)",
					},
				},
				"download_excel": map[string]interface{}{
					"method":      "POST",
					"path":        "/api/download-excel",
					"description": "OCR 결과를 Excel 파일로 다운로드",
					"params": []string{
						"excel_data (required JSON string)",
						"user_name, depositor_dc, dept_cd, emp_cd, bank_cd, ba_nb (optional)",
					},
				},
				"health": map[string]interface{}{
					"method":      "GET",
					"path":        "/api/health",
					"description": "서버 상태 확인",
				},
			},
			"supported_formats": []string{"jpg", "jpeg", "png", "pdf", "tif", "tiff"},
			"categories": map[string]string{
				"6130": "석식",
				"6310": "교통정산",
				"6320": "국내출장",
			},
		})
	})
}
