// ocr-processor.js - OCR 처리 및 서버 통신

const OCRProcessor = {
    // 처리 상태 플래그
    isProcessing: false,
    
    //OCR 처리 시작
    async process() {
        // 중복 호출 방지
        if (this.isProcessing) {
            console.log('이미 OCR 처리 중입니다. 요청 무시됨');
            return;
        }
        
        console.log('=== OCR 처리 시작 ===');
        this.isProcessing = true;
        
        try {
            // 유효성 검사
            const validation = this._validateInput();
            if (!validation.isValid) {
                UIUtils.showError(validation.message);
                return;
            }
            
            // UI 로딩 상태로 변경
            this._setLoadingState(true);
            
            // FormData 생성 및 전송
            const formData = this._createFormData();
            const result = await this._sendOCRRequest(formData);
            
            // 결과 처리
            if (result.success) {
                this._handleSuccess(result);
            } else {
                throw new Error(result.message || '알 수 없는 오류가 발생했습니다.');
            }
            
        } catch (error) {
            UIUtils.showError('❌ ' + error.message);
        } finally {
            this._setLoadingState(false);
            this.isProcessing = false;
        }
    },

    // 추가 OCR 처리 시작
    async processAdditional() {
        // 중복 호출 방지
        if (this.isProcessing) {
            console.log('이미 OCR 처리 중입니다. 요청 무시됨');
            return;
        }

        console.log('=== 추가 OCR 처리 시작 ===');
        this.isProcessing = true;

        try {
            // 추가 파일 유효성 검사
            const validation = this._validateAdditionalInput();
            if (!validation.isValid) {
                UIUtils.showError(validation.message);
                return;
            }

            // UI 로딩 상태로 변경
            this._setAdditionalLoadingState(true);

            // FormData 생성 및 전송
            const formData = this._createAdditionalFormData();
            const result = await this._sendOCRRequest(formData);

            // 결과 처리
            if (result.success) {
                this._handleAdditionalSuccess(result);
            } else {
                throw new Error(result.message || '알 수 없는 오류가 발생했습니다.');
            }

        } catch (error) {
            UIUtils.showError('❌ ' + error.message);
        } finally {
            this._setAdditionalLoadingState(false);
            this.isProcessing = false;
        }
    },
    
    // 입력 유효성 검사 (기본)
    _validateInput() {
        // 폼 필드 검사
        const formValidation = ValidationUtils.validateFormFields();
        if (!formValidation.isValid) {
            return formValidation;
        }
        
        // 파일 선택 검사
        const hasFiles = selectedFiles.length > 0 || 
                         (document.getElementById('images')?.files?.length > 0);
        
        if (!hasFiles) {
            return {
                isValid: false,
                message: '최소 1개의 이미지 파일을 선택해주세요.'
            };
        }
        
        // 업로드할 파일들 준비
        const filesToUpload = this._prepareFilesForUpload();
        
        // 국내출장 필드 유효성 검사
        for (const fileData of filesToUpload) {
            const validation = ValidationUtils.validateBusinessTripFields(fileData);
            if (!validation.isValid) {
                return validation;
            }
        }
        
        return { isValid: true, filesToUpload };
    },

    // 추가 입력 유효성 검사
    _validateAdditionalInput() {
        // 폼 필드 검사
        const formValidation = ValidationUtils.validateFormFields();
        if (!formValidation.isValid) {
            return formValidation;
        }

        // 추가 파일 선택 검사
        if (!window.additionalSelectedFiles || window.additionalSelectedFiles.length === 0) {
            return {
                isValid: false,
                message: '추가 처리할 파일을 선택해주세요.'
            };
        }

        // 국내출장 필드 유효성 검사
        for (const fileData of window.additionalSelectedFiles) {
            const validation = ValidationUtils.validateBusinessTripFields(fileData);
            if (!validation.isValid) {
                return validation;
            }
        }

        return { isValid: true };
    },
    
    // 업로드용 파일 준비 (기본)
    _prepareFilesForUpload() {
        const imagesInput = document.getElementById('images');
        const hasFilesInInput = imagesInput?.files?.length > 0;
        const hasFilesInArray = selectedFiles.length > 0;
        
        let filesToUpload = [];
        
        if (hasFilesInArray) {
            filesToUpload = selectedFiles;
        } else if (hasFilesInInput) {
            // input에서 파일 가져와서 기본 구조로 변환
            filesToUpload = Array.from(imagesInput.files).map(file => ({
                file: file,
                category: '6130', // 기본값: 석식
                additionalNames: '',
                remarks: '',
                businessContent: '',
                purpose: ''
            }));
        }
        
        return filesToUpload;
    },
    
    // FormData 생성 (기본)
    _createFormData() {
        const formData = new FormData();
        
        // 텍스트 필드 추가
        FORM_FIELDS.forEach(fieldId => {
            const value = UTILS.getFormValue(fieldId);
            if (value) {
                formData.append(fieldId, value);
            }
        });
        
        // 파일들과 메타데이터 추가
        const filesToUpload = this._prepareFilesForUpload();
        filesToUpload.forEach((fileData, index) => {
            formData.append('images', fileData.file);
            formData.append(`category_${index}`, fileData.category);
            formData.append(`remarks_${index}`, fileData.remarks || '');
            formData.append(`additional_names_${index}`, fileData.additionalNames || '');
            
            // 국내출장 전용 필드들
            if (fileData.category === '6320') {
                formData.append(`business_content_${index}`, fileData.businessContent || '');
                formData.append(`purpose_${index}`, fileData.purpose || '');
            }
        });
        
        return formData;
    },

    // FormData 생성 (추가)
    _createAdditionalFormData() {
        const formData = new FormData();
        
        // 텍스트 필드 추가
        FORM_FIELDS.forEach(fieldId => {
            const value = UTILS.getFormValue(fieldId);
            if (value) {
                formData.append(fieldId, value);
            }
        });
        
        // 추가 파일들과 메타데이터 추가
        window.additionalSelectedFiles.forEach((fileData, index) => {
            formData.append('images', fileData.file);
            formData.append(`category_${index}`, fileData.category);
            formData.append(`remarks_${index}`, fileData.remarks || '');
            formData.append(`additional_names_${index}`, fileData.additionalNames || '');
            
            // 국내출장 전용 필드들
            if (fileData.category === '6320') {
                formData.append(`business_content_${index}`, fileData.businessContent || '');
                formData.append(`purpose_${index}`, fileData.purpose || '');
            }
        });
        
        return formData;
    },
    
    // OCR API 요청 전송
    async _sendOCRRequest(formData) {
        console.log('OCR API 요청 전송:', CONFIG.API.PROCESS_OCR);
        
        const response = await fetch(CONFIG.API.PROCESS_OCR, {
            method: 'POST',
            body: formData
        });
        
        console.log('응답 상태:', response.status);
        
        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.error || `HTTP ${response.status} 오류`);
        }
        
        return await response.json();
    },
    
    // 성공 처리 (기본)
    _handleSuccess(result) {
        console.log('OCR 처리 성공:', result);
        
        // 전역 변수에 결과 저장
        ocrResults = result.data;
        
        // 성공 메시지 표시
        UIUtils.showSuccess(`✅ ${result.data.length}개 파일의 OCR 처리가 완료되었습니다!`);
        
        // 결과 화면으로 전환
        ResultsManager.showResults(result.data);
        
        // 폼 데이터 저장
        StorageManager.saveFormData();
    },

    // 추가 성공 처리
    _handleAdditionalSuccess(result) {
        console.log('추가 OCR 처리 성공:', result);
        
        // 기존 결과에 새 결과 추가
        ocrResults = ocrResults.concat(result.data);
        
        // 추가 파일 목록 초기화
        window.additionalSelectedFiles = [];
        FileManager.updateAdditionalFileList();
        
        // 추가 파일 입력 초기화
        const additionalImagesInput = document.getElementById('additionalImages');
        if (additionalImagesInput) {
            additionalImagesInput.value = '';
        }
        
        // 성공 메시지 표시
        UIUtils.showSuccess(`✅ ${result.data.length}개 추가 파일의 OCR 처리가 완료되었습니다! (총 ${ocrResults.length}개)`);
        
        // 결과 테이블 업데이트
        ResultsManager.showResults(ocrResults);
        
        // 폼 데이터 저장
        StorageManager.saveFormData();
    },
    
    // 로딩 상태 설정 (기본)
    _setLoadingState(isLoading) {
        const filesToUpload = this._prepareFilesForUpload();
        const fileCount = filesToUpload.length;
        
        UIUtils.setLoading(isLoading, {
            buttonText: '처리 중...',
            originalButtonText: '📤 OCR 처리하기',
            loadingMessage: `${fileCount}개의 이미지를 처리 중입니다... 잠시만 기다려주세요.`
        });
    },

    // 로딩 상태 설정 (추가)
    _setAdditionalLoadingState(isLoading) {
        const fileCount = window.additionalSelectedFiles ? window.additionalSelectedFiles.length : 0;
        
        // 추가 OCR 버튼 상태 변경
        UIUtils.setButtonState('additionalOcrBtn', isLoading, isLoading ? '처리 중...' : '➕ 추가 OCR 처리하기');
        
        // 로딩 메시지 업데이트
        const loadingText = document.getElementById('loadingText');
        if (loadingText && isLoading) {
            loadingText.textContent = `${fileCount}개의 추가 이미지를 처리 중입니다... 잠시만 기다려주세요.`;
        }
        
        // 로딩 스피너 표시
        const loading = document.getElementById('loading');
        if (loading) {
            loading.style.display = isLoading ? 'block' : 'none';
        }
    },
    
    // 리셋 처리
    reset() {
        // 결과 섹션 숨기기
        UTILS.toggleElement('resultsSection', false);
        UTILS.toggleElement('uploadForm', true);
        
        // 데이터 초기화
        ocrResults = [];
        window.additionalSelectedFiles = [];
        FileManager.reset();
        
        // 메시지 숨기기
        UIUtils.hideMessages();
        
        this.isProcessing = false;
        
        console.log('OCR 프로세서 리셋 완료');
    }
};

console.log('ocr-processor.js 로드됨');