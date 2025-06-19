// utils.js - 공통 유틸리티 함수

// 날짜 관련 유틸리티
const DateUtils = {
    // 날짜를 YYYYMMDD 형식으로 변환 (서버 측과 동일한 로직)
    convertToYYYYMMDD(dateText) {
        if (!dateText) return '';
        
        console.log('날짜 변환 시도:', dateText);
        
        // 1. YYYY. M. D. HH: MM:SS 형식 (예: "2025. 4. 23. 18: 21:56")
        let match = dateText.match(DATE_PATTERNS.YYYY_M_D_TIME);
        if (match) {
            const year = match[1];
            const month = match[2].padStart(2, '0');
            const day = match[3].padStart(2, '0');
            const result = year + month + day;
            console.log('YYYY. M. D. HH: MM:SS 형식으로 변환:', result);
            return result;
        }
        
        // 2. YYYY. M. D. HH: MM 형식 (예: "2025. 4. 23. 18: 21")
        match = dateText.match(DATE_PATTERNS.YYYY_M_D_TIME_SHORT);
        if (match) {
            const year = match[1];
            const month = match[2].padStart(2, '0');
            const day = match[3].padStart(2, '0');
            const result = year + month + day;
            console.log('YYYY. M. D. HH: MM 형식으로 변환:', result);
            return result;
        }
        
        // 3. YYYY. M. D 형식 (예: "2025. 4. 23")
        match = dateText.match(DATE_PATTERNS.YYYY_M_D);
        if (match) {
            const year = match[1];
            const month = match[2].padStart(2, '0');
            const day = match[3].padStart(2, '0');
            const result = year + month + day;
            console.log('YYYY. M. D 형식으로 변환:', result);
            return result;
        }
        
        // 4. YYYY.MM.DD 형식 (기존)
        match = dateText.match(DATE_PATTERNS.YYYY_MM_DD);
        if (match) {
            const result = match[1] + match[2] + match[3];
            console.log('YYYY.MM.DD 형식으로 변환:', result);
            return result;
        }
        
        // 5. YY.MM.DD HH:MM 형식 (공백 있음)
        match = dateText.match(DATE_PATTERNS.YY_MM_DD_TIME);
        if (match) {
            const result = '20' + match[1] + match[2] + match[3];
            console.log('YY.MM.DD HH:MM 형식으로 변환:', result);
            return result;
        }
        
        // 6. YY.MM.DDHH:MM 형식 (공백 없음)
        match = dateText.match(DATE_PATTERNS.YY_MM_DD_NO_SPACE);
        if (match) {
            const result = '20' + match[1] + match[2] + match[3];
            console.log('YY.MM.DDHH:MM 형식으로 변환:', result);
            return result;
        }
        
        // 7. YY.MM.DD 형식 (시간 없음)
        match = dateText.match(DATE_PATTERNS.YY_MM_DD);
        if (match) {
            const result = '20' + match[1] + match[2] + match[3];
            console.log('YY.MM.DD 형식으로 변환:', result);
            return result;
        }
        
        // 8. YYYYMMDD 형식이 이미 있는 경우
        match = dateText.match(DATE_PATTERNS.YYYYMMDD);
        if (match) {
            console.log('YYYYMMDD 형식 그대로 사용:', match[0]);
            return match[0];
        }
        
        console.log('날짜 형식을 인식하지 못함, 원본 반환:', dateText);
        return dateText;
    },
    
    // YYYYMMDD에서 MM/DD 형식으로 변환
    formatToMMDD(dateText) {
        const convertedDate = this.convertToYYYYMMDD(dateText);
        if (convertedDate.length >= 8) {
            const month = convertedDate.substring(4, 6);
            const day = convertedDate.substring(6, 8);
            return `${month}/${day}`;
        }
        return 'MM/DD';
    },
    
    // 현재 날짜를 YYYYMMDD 형식으로 반환
    getCurrentYYYYMMDD() {
        const now = new Date();
        const year = now.getFullYear();
        const month = String(now.getMonth() + 1).padStart(2, '0');
        const day = String(now.getDate()).padStart(2, '0');
        return year + month + day;
    },
    
    // 결제일 계산 (현재 날짜 기준 10일 이전이면 당월 15일, 이후면 다음달 15일)
    calculatePaymentDate() {
        const now = new Date();
        const currentDay = now.getDate(); // getDay()가 아니라 getDate()
        
        let paymentDate;
        if (currentDay <= 10) {
            paymentDate = new Date(now.getFullYear(), now.getMonth(), 15);
        } else {
            paymentDate = new Date(now.getFullYear(), now.getMonth() + 1, 15);
        }
        
        return paymentDate.getFullYear() + 
               String(paymentDate.getMonth() + 1).padStart(2, '0') + 
               String(paymentDate.getDate()).padStart(2, '0');
    }
};

// 카테고리 관련 유틸리티
const CategoryUtils = {
    // 카테고리 값을 라벨로 변환
    getLabel(categoryValue) {
        const option = CATEGORY_OPTIONS.find(opt => opt.value === categoryValue);
        return option ? option.label : categoryValue;
    },
    
    // 파일명에서 카테고리 자동 감지
    detectFromFilename(filename) {
        const lowerFilename = filename.toLowerCase();
        
        // 키워드 매핑을 순회하며 매칭되는 카테고리 찾기
        for (const [categoryValue, keywords] of Object.entries(CATEGORY_KEYWORDS)) {
            const matchedKeyword = keywords.find(keyword => lowerFilename.includes(keyword));
            if (matchedKeyword) {
                console.log(`파일 "${filename}": ${matchedKeyword} 키워드 감지, 카테고리를 ${this.getLabel(categoryValue)}으로 설정`);
                return categoryValue;
            }
        }
        
        // 기본값: 석식
        console.log(`파일 "${filename}": 특별한 키워드 없음, 기본값 석식으로 설정`);
        return '6130';
    }
};

// 파일명 관련 유틸리티
const FilenameUtils = {
    // 파일명에서 괄호 안의 내용 추출
    extractNamesFromFilename(filename) {
        const patterns = [
            /\(([^)]+)\)/g,  // 소괄호 ()
            /\[([^\]]+)\]/g, // 대괄호 []
            /\{([^}]+)\}/g   // 중괄호 {}
        ];
        
        let extractedNames = [];
        
        patterns.forEach(pattern => {
            let match;
            while ((match = pattern.exec(filename)) !== null) {
                extractedNames.push(match[1].trim());
            }
        });
        
        // 중복 제거 및 콤마로 연결
        const uniqueNames = [...new Set(extractedNames)];
        const result = uniqueNames.join(',');
        
        console.log(`파일명 "${filename}"에서 추출된 이름:`, result);
        return result;
    },
    
    // 국내출장 파일명에서 용도 추출 (언더스코어 뒤 내용)
    extractPurposeForBusinessTrip(filename) {
        const underscoreIdx = filename.lastIndexOf('_');
        if (underscoreIdx > -1 && underscoreIdx < filename.length - 1) {
            const extIdx = filename.lastIndexOf('.');
            if (extIdx > underscoreIdx) {
                return filename.substring(underscoreIdx + 1, extIdx).trim();
            } else {
                return filename.substring(underscoreIdx + 1).trim();
            }
        }
        return '';
    }
};

// 비고 생성 유틸리티
const RemarkUtils = {
    // 기본 비고 생성 (MM/DD_이름_카테고리 형식)
    generateDefault(issueDate, userName, categoryValue) {
        if (!userName) {
            return `카테고리: ${CategoryUtils.getLabel(categoryValue)}`;
        }
        
        const monthDay = DateUtils.formatToMMDD(issueDate);
        const categoryLabel = CategoryUtils.getLabel(categoryValue);
        
        return `${monthDay}_${userName}_${categoryLabel}`;
    },
    
    // 국내출장 비고 생성 (출장내용_이름_용도 형식)
    generateBusinessTrip(businessContent, userName, additionalNames, purpose) {
        let finalName = userName;
        if (additionalNames) {
            finalName = `${userName},${additionalNames}`;
        }
        
        return `${businessContent}_${finalName}_${purpose}`;
    },
    
    // 비고가 자동 생성된 형태인지 확인
    isAutoGenerated(remark) {
        return remark && remark.includes('_') && 
               (remark.includes('/') || remark.includes('임시_'));
    }
};

// 금액 관련 유틸리티
const AmountUtils = {
    // 금액 텍스트 정리 (예: "32,300 원" -> "32300")
    clean(amountText) {
        if (!amountText) return '';
        
        const matches = amountText.match(/[0-9,]+/);
        if (matches && matches.length > 0) {
            return matches[0].replace(/,/g, '');
        }
        
        return amountText;
    }
};

// UI 관련 유틸리티
const UIUtils = {
    // 에러 메시지 표시
    showError(message) {
        const errorMsg = document.getElementById('errorMsg');
        const successMsg = document.getElementById('successMsg');
        
        if (errorMsg) {
            errorMsg.textContent = message;
            errorMsg.style.display = 'block';
        }
        
        if (successMsg) {
            successMsg.style.display = 'none';
        }
        
        console.error('Error:', message);
    },
    
    // 성공 메시지 표시
    showSuccess(message) {
        const successMsg = document.getElementById('successMsg');
        const errorMsg = document.getElementById('errorMsg');
        
        if (successMsg) {
            successMsg.textContent = message;
            successMsg.style.display = 'block';
        }
        
        if (errorMsg) {
            errorMsg.style.display = 'none';
        }
        
        console.log('Success:', message);
    },
    
    // 메시지 숨기기
    hideMessages() {
        const errorMsg = document.getElementById('errorMsg');
        const successMsg = document.getElementById('successMsg');
        
        if (errorMsg) errorMsg.style.display = 'none';
        if (successMsg) successMsg.style.display = 'none';
    },
    
    // 로딩 상태 설정
    setLoading(isLoading, options = {}) {
        const loading = document.getElementById('loading');
        const loadingText = document.getElementById('loadingText');
        const submitBtn = document.getElementById('submitBtn');
        
        if (loading) {
            loading.style.display = isLoading ? 'block' : 'none';
        }
        
        if (submitBtn) {
            submitBtn.disabled = isLoading;
            submitBtn.textContent = isLoading ? 
                (options.buttonText || '처리 중...') : 
                (options.originalButtonText || '📤 OCR 처리하기');
        }
        
        if (loadingText && options.loadingMessage) {
            loadingText.textContent = options.loadingMessage;
        }
    },
    
    // 버튼 상태 설정
    setButtonState(buttonId, disabled, text) {
        const button = document.getElementById(buttonId);
        if (button) {
            button.disabled = disabled;
            if (text) {
                button.textContent = text;
            }
        }
    }
};

// 유효성 검사 유틸리티
const ValidationUtils = {
    // 파일 유효성 검사
    validateFile(file) {
        // 파일 형식 검사
        if (!UTILS.isSupportedFileType(file)) {
            return {
                isValid: false,
                message: `파일 '${file.name}'는 지원하지 않는 형식입니다. (${CONFIG.SUPPORTED_EXTENSIONS.join(', ')}만 지원)`
            };
        }
        
        // 파일 크기 검사
        if (!UTILS.isValidFileSize(file)) {
            return {
                isValid: false,
                message: `파일 '${file.name}'의 크기가 ${CONFIG.MAX_FILE_SIZE_MB}MB를 초과합니다.`
            };
        }
        
        return { isValid: true };
    },
    
    // 국내출장 필드 유효성 검사
    validateBusinessTripFields(fileData) {
        if (fileData.category === '6320') {
            if (!fileData.businessContent || !fileData.businessContent.trim()) {
                return {
                    isValid: false,
                    message: `파일 "${fileData.file.name}": 국내출장 카테고리는 출장내용을 입력해야 합니다.`
                };
            }
            
            if (!fileData.purpose || !fileData.purpose.trim()) {
                return {
                    isValid: false,
                    message: `파일 "${fileData.file.name}": 국내출장 카테고리는 용도를 입력해야 합니다.`
                };
            }
        }
        
        return { isValid: true };
    },
    
    // 폼 필드 유효성 검사
    validateFormFields() {
        for (const fieldId of REQUIRED_FIELDS) {
            const value = UTILS.getFormValue(fieldId);
            if (!value) {
                return {
                    isValid: false,
                    message: `${fieldId === 'user_name' ? '이름' : fieldId}을(를) 입력해주세요.`
                };
            }
        }
        
        return { isValid: true };
    }
};

// 배열 정렬 유틸리티
const SortUtils = {
    // 날짜 기준 오름차순 정렬 (YYYYMMDD 형식)
    sortByDate(array, dateFieldName = 'issueDate') {
        return array.sort((a, b) => {
            const dateA = a[dateFieldName] || '';
            const dateB = b[dateFieldName] || '';
            return dateA.localeCompare(dateB);
        });
    }
};

console.log('utils.js 로드됨');