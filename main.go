package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	// .env 파일 로드 (선택사항, 파일이 없어도 무시)
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env 파일을 찾을 수 없습니다. 시스템 환경변수를 사용합니다.")
	}

	// Fiber 앱 생성
	uploadLimitMB := getUploadLimitMB()
	app := fiber.New(fiber.Config{
		BodyLimit: uploadLimitMB * 1024 * 1024, // MB를 바이트로 변환
	})

	// 미들웨어 설정
	app.Use(logger.New())
	app.Use(cors.New())

	// 라우트 설정
	setupRoutes(app)

	// 서버 시작
	port := getServerPort()
	log.Printf("서버가 %s 포트에서 시작됩니다...", port)
	log.Printf("업로드 제한: %dMB, 최대 파일 수: %d개", uploadLimitMB, getMaxFiles())
	log.Printf("2단계 OCR 플로우:")
	log.Printf("  1단계: POST /api/process-ocr (OCR 처리)")
	log.Printf("  2단계: POST /api/download-excel (Excel 다운로드)")
	log.Fatal(app.Listen(port))
}
