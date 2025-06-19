package main

// OCR 요청 구조체
type OCRRequest struct {
	Version   string     `json:"version"`
	RequestID string     `json:"requestId"`
	Timestamp int64      `json:"timestamp"`
	Lang      string     `json:"lang"`
	Images    []OCRImage `json:"images"`
}

type OCRImage struct {
	Format      string `json:"format"`
	Name        string `json:"name"`
	Data        string `json:"data"`
	TemplateIds int    `json:"templateIds,omitempty"`
}

// OCR 응답 구조체
type OCRResponse struct {
	Version   string           `json:"version"`
	RequestID string           `json:"requestId"`
	Timestamp int64            `json:"timestamp"`
	Images    []OCRImageResult `json:"images"`
}

type OCRImageResult struct {
	UID              string           `json:"uid"`
	Name             string           `json:"name"`
	InferResult      string           `json:"inferResult"`
	Message          string           `json:"message"`
	MatchedTemplate  MatchedTemplate  `json:"matchedTemplate"`
	ValidationResult ValidationResult `json:"validationResult"`
	Fields           []Field          `json:"fields"`
	Title            Title            `json:"title"`
}

type MatchedTemplate struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ValidationResult struct {
	Result string `json:"result"`
}

type Field struct {
	Name            string   `json:"name"`
	Bounding        Bounding `json:"bounding"`
	ValueType       string   `json:"valueType"`
	InferText       string   `json:"inferText"`
	InferConfidence float64  `json:"inferConfidence"`
}

type Bounding struct {
	Top    float64 `json:"top"`
	Left   float64 `json:"left"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type Title struct {
	Name            string   `json:"name"`
	Bounding        Bounding `json:"bounding"`
	InferText       string   `json:"inferText"`
	InferConfidence float64  `json:"inferConfidence"`
}

// 업로드 요청 구조체
type UploadRequest struct {
	UserName    string `form:"user_name"`
	AttrCD      string `form:"attr_cd"`
	DepositorDC string `form:"depositor_dc"`
	DeptCD      string `form:"dept_cd"`
	EmpCD       string `form:"emp_cd"`
	BankCD      string `form:"bank_cd"`
	BANB        string `form:"ba_nb"`
}

// Excel 데이터 구조체
type ExcelData struct {
	CASHCD      string
	RMKDC       string
	TRNM        string
	SUPAM       string
	VATAM       string
	ATTRCD      string
	ISSDT       string
	PAYDT       string
	BANKCD      string
	BANB        string
	DEPOSITORDC string
	DEPTCD      string
	EMPCD       string
}

// 이미지 파일 정보 구조체
type ImageFile struct {
	Data     []byte
	Filename string
}

// 카테고리와 비고가 포함된 이미지 파일 정보 구조체
type ImageFileWithCategory struct {
	ImageFile       ImageFile
	Category        string
	Remarks         string
	BusinessContent string // 국내출장 전용: 출장내용
	Purpose         string // 국내출장 전용: 용도
}

// 개별 이미지 OCR 결과
type SingleImageOCRResult struct {
	ImageIndex int
	ImageName  string
	Response   *OCRImageResult
	Error      error
}

// 카테고리와 비고가 포함된 개별 이미지 OCR 결과
type SingleImageOCRResultWithCategory struct {
	ImageIndex      int
	ImageName       string
	Category        string
	Remarks         string
	BusinessContent string // 국내출장 전용: 출장내용
	Purpose         string // 국내출장 전용: 용도
	Response        *OCRImageResult
	Error           error
}

// OCR 처리 응답 구조체 (프론트엔드용)
type OCRProcessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    []OCRResult `json:"data"`
}

// OCR 결과 데이터 구조체 (프론트엔드용)
type OCRResult struct {
	FileName        string `json:"fileName"`
	Category        string `json:"category"`
	Remark          string `json:"remark"`
	Purpose         string `json:"purpose"`
	Amount          string `json:"amount"`
	IssueDate       string `json:"issueDate"`
	PayDate         string `json:"payDate"`
	BusinessContent string `json:"businessContent,omitempty"` // 국내출장 전용
	BusinessPurpose string `json:"businessPurpose,omitempty"` // 국내출장 전용
}

// Excel 다운로드 요청 구조체
type ExcelDownloadRequest struct {
	UserName    string `form:"user_name"`
	AttrCD      string `form:"attr_cd"`
	DepositorDC string `form:"depositor_dc"`
	DeptCD      string `form:"dept_cd"`
	EmpCD       string `form:"emp_cd"`
	BankCD      string `form:"bank_cd"`
	BANB        string `form:"ba_nb"`
	ExcelData   string `form:"excel_data"`
}
