// config.js - 애플리케이션 설정 및 상수 (간소화)

// 전역 상수
const CONFIG = {
    // 파일 제한
    MAX_FILES: 5,
    MAX_FILE_SIZE_MB: 50,
    
    // 지원 파일 형식
    SUPPORTED_FILE_TYPES: ['image/jpeg', 'image/jpg', 'image/png', 'image/tiff', 'image/tif', 'application/pdf'],
    SUPPORTED_EXTENSIONS: ['.jpg', '.jpeg', '.png', '.pdf', '.tif', '.tiff'],
    
    // API 엔드포인트
    API: {
        PROCESS_OCR: '/api/process-ocr',
        DOWNLOAD_EXCEL: '/api/download-excel'
    },
    
    // 쿠키 설정
    COOKIE: {
        EXPIRY_DAYS: 30,
        PREFIX: 'ocr_form_'
    }
};

// 카테고리 옵션
const CATEGORY_OPTIONS = [
    { label: '조식', value: '6110' },
    { label: '중식', value: '6120' },
    { label: '석식', value: '6130' },
    { label: '교통정산', value: '6310' },
    { label: '국내출장', value: '6320' }
];

// 폼 필드 정의
const FORM_FIELDS = ['user_name', 'attr_cd', 'depositor_dc', 'dept_cd', 'emp_cd', 'bank_cd', 'ba_nb'];
const REQUIRED_FIELDS = ['user_name'];

// 키워드 기반 카테고리 매핑
const CATEGORY_KEYWORDS = {
    '6320': ['국내출장', '출장'],
    //'6310': ['교통', '택시', '지하철', '버스'],
    '6310': ['교통', '택시', '지하철'],
    '6110': ['조식', '아침'],
    '6120': ['중식', '점심'],
    '6130': ['석식', '저녁']
};

// 날짜 포맷 정규식 (통합)
const DATE_PATTERNS = {
    YYYY_M_D_TIME: /(\d{4})\.\s*(\d{1,2})\.\s*(\d{1,2})\.\s*\d{1,2}:\s*\d{1,2}:\s*\d{1,2}/,
    YYYY_M_D_TIME_SHORT: /(\d{4})\.\s*(\d{1,2})\.\s*(\d{1,2})\.\s*\d{1,2}:\s*\d{1,2}/,
    YYYY_M_D: /(\d{4})\.\s*(\d{1,2})\.\s*(\d{1,2})/,
    YYYY_MM_DD: /(\d{4})\.(\d{2})\.(\d{2})/,
    YY_MM_DD_TIME_FULL: /(\d{2})\.(\d{2})\.(\d{2})\s+\d{2}:\s*\d{2}:\s*\d{2}/,
    YY_MM_DD_TIME_WITH_I: /(\d{2})\.(\d{2})\.(\d{2})\s+[I|]\s+\d{2}:\s*\d{2}:\s*\d{2}/,
    YY_MM_DD_TIME: /(\d{2})\.(\d{2})\.(\d{2})\s+\d{2}:\s*\d{2}/,
    YY_MM_DD_NO_SPACE: /(\d{2})\.(\d{2})\.(\d{2})\d{2}:\d{2}/,
    YY_MM_DD: /^(\d{2})\.(\d{2})\.(\d{2})$/,
    YYYYMMDD: /(\d{8})/
};

// 유틸리티 함수 (통합)
const UTILS = {
    // 파일 관련
    getFileExtension(filename) {
        return filename.toLowerCase().substring(filename.lastIndexOf('.'));
    },
    
    isSupportedFileType(file) {
        return CONFIG.SUPPORTED_FILE_TYPES.includes(file.type) || 
               CONFIG.SUPPORTED_EXTENSIONS.includes(this.getFileExtension(file.name));
    },
    
    isValidFileSize(file) {
        return file.size <= CONFIG.MAX_FILE_SIZE_MB * 1024 * 1024;
    },
    
    formatFileSize(bytes) {
        return (bytes / 1024 / 1024).toFixed(2) + ' MB';
    },
    
    // DOM 관련
    toggleElement(elementId, show) {
        const element = document.getElementById(elementId);
        if (element) {
            element.style.display = show ? 'block' : 'none';
        }
    },
    
    getFormValue(elementId) {
        const element = document.getElementById(elementId);
        return element ? element.value.trim() : '';
    },
    
    setFormValue(elementId, value) {
        const element = document.getElementById(elementId);
        if (element) {
            element.value = value;
        }
    },
    
    // 기타
    debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }
};

console.log('config.js 로드됨');