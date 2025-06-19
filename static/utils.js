// utils.js - 통합된 유틸리티 함수들 (중복 제거)

// 날짜 관련 유틸리티
const DateUtils = {
    // 날짜를 YYYYMMDD 형식으로 변환 (서버와 동일한 로직)
    convertToYYYYMMDD(dateText) {
        if (!dateText) return '';
        
        const patterns = [
            { regex: DATE_PATTERNS.YYYY_M_D_TIME, desc: "YYYY. M. D. HH:MM:SS" },
            { regex: DATE_PATTERNS.YYYY_M_D_TIME_SHORT, desc: "YYYY. M. D. HH:MM" },
            { regex: DATE_PATTERNS.YYYY_M_D, desc: "YYYY. M. D" },
            { regex: DATE_PATTERNS.YYYY_MM_DD, desc: "YYYY.MM.DD" },
            { regex: DATE_PATTERNS.YY_MM_DD_TIME, desc: "YY.MM.DD HH:MM" },
            { regex: DATE_PATTERNS.YY_MM_DD_NO_SPACE, desc: "YY.MM.DDHH:MM" },
            { regex: DATE_PATTERNS.YY_MM_DD, desc: "YY.MM.DD" }
        ];

        for (const pattern of patterns) {
            const match = dateText.match(pattern.regex);
            if (match && match.length >= 4) {
                let year = match[1];
                let month = match[2];
                let day = match[3];

                // YY를 YYYY로 변환
                if (year.length === 2) {
                    year = '20' + year;
                }

                // 월, 일을 2자리로 패딩
                month = month.padStart(2, '0');
                day = day.padStart(2, '0');

                return year + month + day;
            }
        }

        // YYYYMMDD 형식이 이미 있는 경우
        const yyyymmddMatch = dateText.match(DATE_PATTERNS.YYYYMMDD);
        if (yyyymmddMatch) {
            return yyyymmddMatch[0];
        }

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
    
    // 결제일 계산
    calculatePaymentDate() {
        const now = new Date();
        const currentDay = now.getDate();
        
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
    getLabel(categoryValue) {
        const option = CATEGORY_OPTIONS.find(opt => opt.value === categoryValue);
        return option ? option.label : categoryValue;
    },
    
    detectFromFilename(filename) {
        const lowerFilename = filename.toLowerCase();
        
        for (const [categoryValue, keywords] of Object.entries(CATEGORY_KEYWORDS)) {
            const matchedKeyword = keywords.find(keyword => lowerFilename.includes(keyword));
            if (matchedKeyword) {
                console.log(`파일 "${filename}": ${matchedKeyword} 키워드 감지, 카테고리를 ${this.getLabel(categoryValue)}으로 설정`);
                return categoryValue;
            }
        }
        
        console.log(`파일 "${filename}": 특별한 키워드 없음, 기본값 석식으로 설정`);
        return '6130';
    }
};

// 파일명 관련 유틸리티
const FilenameUtils = {
    extractNamesFromFilename(filename) {
        const patterns = [/\(([^)]+)\)/g, /\[([^\]]+)\]/g, /\{([^}]+)\}/g];
        let extractedNames = [];
        
        patterns.forEach(pattern => {
            let match;
            while ((match = pattern.exec(filename)) !== null) {
                extractedNames.push(match[1].trim());
            }
        });
        
        const uniqueNames = [...new Set(extractedNames)];
        const result = uniqueNames.join(',');
        
        console.log(`파일명 "${filename}"에서 추출된 이름:`, result);
        return result;
    },
    
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
    generateDefault(issueDate, userName, categoryValue) {
        if (!userName) {
            return `카테고리: ${CategoryUtils.getLabel(categoryValue)}`;
        }
        
        const monthDay = DateUtils.formatToMMDD(issueDate);
        const categoryLabel = CategoryUtils.getLabel(categoryValue);
        
        return `${monthDay}_${userName}_${categoryLabel}`;
    },
    
    generateBusinessTrip(businessContent, userName, additionalNames, purpose) {
        let finalName = userName;
        if (additionalNames) {
            finalName = `${userName},${additionalNames}`;
        }
        
        return `${businessContent}_${finalName}_${purpose}`;
    },
    
    isAutoGenerated(remark) {
        return remark && remark.includes('_') && 
               (remark.includes('/') || remark.includes('임시_'));
    }
};

// UI 관련 유틸리티
const UIUtils = {
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
    
    hideMessages() {
        const errorMsg = document.getElementById('errorMsg');
        const successMsg = document.getElementById('successMsg');
        
        if (errorMsg) errorMsg.style.display = 'none';
        if (successMsg) successMsg.style.display = 'none';
    },
    
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
    validateFile(file) {
        if (!UTILS.isSupportedFileType(file)) {
            return {
                isValid: false,
                message: `파일 '${file.name}'는 지원하지 않는 형식입니다. (${CONFIG.SUPPORTED_EXTENSIONS.join(', ')}만 지원)`
            };
        }
        
        if (!UTILS.isValidFileSize(file)) {
            return {
                isValid: false,
                message: `파일 '${file.name}'의 크기가 ${CONFIG.MAX_FILE_SIZE_MB}MB를 초과합니다.`
            };
        }
        
        return { isValid: true };
    },
    
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

console.log('utils.js 로드됨');