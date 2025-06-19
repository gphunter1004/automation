// main.js - 메인 초기화 및 이벤트 바인딩 (추가 OCR 기능 포함)

// 전역 변수
let selectedFiles = [];
let ocrResults = [];

// 애플리케이션 초기화
const App = {
    // 애플리케이션 시작
    init() {
        console.log('=== OCR to Excel 애플리케이션 시작 ===');
        
        // DOM 로드 완료 대기
        if (document.readyState === 'loading') {
            document.addEventListener('DOMContentLoaded', () => this._initialize());
        } else {
            this._initialize();
        }
    },
    
    // 초기화 실행
    _initialize() {
        console.log('애플리케이션 초기화 시작');
        
        try {
            // 전역 추가 파일 배열 초기화
            window.additionalSelectedFiles = [];
            
            // 1. 저장된 데이터 불러오기
            StorageManager.loadFormData();
            
            // 2. 이벤트 리스너 설정
            this._setupEventListeners();
            
            // 3. 자동 저장 설정
            StorageManager.setupAutoSave();
            
            // 4. 드래그 앤 드롭 설정
            FileManager.setupDragAndDrop();
            
            // 5. 개발자 도구 추가 (조건부)
            StorageManager.addDevTools();
            
            console.log('애플리케이션 초기화 완료');
            
        } catch (error) {
            console.error('애플리케이션 초기화 중 오류:', error);
            UIUtils.showError('애플리케이션 초기화 중 오류가 발생했습니다.');
        }
    },
    
    // 이벤트 리스너 설정
    _setupEventListeners() {
        console.log('이벤트 리스너 설정 시작');
        
        // 폼 제출 방지 (버튼 클릭으로만 처리)
        this._setupFormSubmitPrevention();
        
        // 기본 OCR 처리 관련 이벤트
        this._setupBasicOCREvents();
        
        // 추가 OCR 처리 관련 이벤트
        this._setupAdditionalOCREvents();
        
        // 결과 관리 관련 이벤트
        this._setupResultsEvents();
        
        // 사용자 입력 관련 이벤트
        this._setupUserInputEvents();
        
        console.log('이벤트 리스너 설정 완료');
    },
    
    // 폼 제출 방지 설정
    _setupFormSubmitPrevention() {
        const uploadForm = document.getElementById('uploadForm');
        if (uploadForm) {
            uploadForm.addEventListener('submit', (e) => {
                e.preventDefault();
                e.stopPropagation();
                console.log('Form submit 이벤트 차단됨');
                return false;
            });
        }
    },
    
    // 기본 OCR 처리 이벤트 설정
    _setupBasicOCREvents() {
        // OCR 처리 버튼
        const submitBtn = document.getElementById('submitBtn');
        if (submitBtn) {
            submitBtn.addEventListener('click', (e) => {
                e.preventDefault();
                e.stopPropagation();
                console.log('OCR 처리 버튼 클릭');
                
                // 폼 데이터 저장 후 OCR 처리
                StorageManager.saveFormData();
                OCRProcessor.process();
            });
        }
        
        // 기본 파일 선택 입력
        const imagesInput = document.getElementById('images');
        if (imagesInput) {
            imagesInput.addEventListener('change', (e) => {
                console.log('기본 파일 선택 이벤트 발생');
                FileManager.handleFileSelect(e);
            });
        }
    },
    
    // 추가 OCR 처리 이벤트 설정
    _setupAdditionalOCREvents() {
        // 추가 파일 선택 입력
        const additionalImagesInput = document.getElementById('additionalImages');
        if (additionalImagesInput) {
            additionalImagesInput.addEventListener('change', (e) => {
                console.log('추가 파일 선택 이벤트 발생');
                FileManager.handleAdditionalFileSelect(e);
            });
        }
        
        // 추가 OCR 처리 버튼
        const additionalOcrBtn = document.getElementById('additionalOcrBtn');
        if (additionalOcrBtn) {
            additionalOcrBtn.addEventListener('click', (e) => {
                e.preventDefault();
                e.stopPropagation();
                console.log('추가 OCR 처리 버튼 클릭');
                
                // 폼 데이터 저장 후 추가 OCR 처리
                StorageManager.saveFormData();
                OCRProcessor.processAdditional();
            });
        }
    },
    
    // 결과 관리 이벤트 설정
    _setupResultsEvents() {
        // Excel 다운로드 버튼
        const downloadBtn = document.getElementById('downloadBtn');
        if (downloadBtn) {
            downloadBtn.addEventListener('click', (e) => {
                e.preventDefault();
                console.log('Excel 다운로드 버튼 클릭');
                ResultsManager.downloadExcel();
            });
        }
        
        // 리셋 버튼
        const resetBtn = document.getElementById('resetBtn');
        if (resetBtn) {
            resetBtn.addEventListener('click', (e) => {
                e.preventDefault();
                console.log('리셋 버튼 클릭');
                this._handleReset();
            });
        }
    },
    
    // 사용자 입력 이벤트 설정
    _setupUserInputEvents() {
        // 사용자 이름 입력 변경 시 비고 업데이트
        const userNameInput = document.getElementById('user_name');
        if (userNameInput) {
            // 실시간 입력 변경 감지 (디바운스 적용)
            const debouncedUpdate = UTILS.debounce(() => {
                this._handleUserNameChange();
            }, 300);
            
            userNameInput.addEventListener('input', debouncedUpdate);
            
            // 포커스 아웃 시 즉시 업데이트
            userNameInput.addEventListener('blur', () => {
                this._handleUserNameChange();
            });
        }
        
        // 기타 폼 필드 변경 감지 (자동 저장용)
        FORM_FIELDS.forEach(fieldId => {
            const element = document.getElementById(fieldId);
            if (element && fieldId !== 'user_name') {
                element.addEventListener('change', () => {
                    console.log(`폼 필드 변경: ${fieldId}`);
                    StorageManager.saveFormData();
                });
            }
        });
    },
    
    // 사용자 이름 변경 처리
    _handleUserNameChange() {
        const userName = UTILS.getFormValue('user_name');
        if (!userName) {
            console.log('사용자 이름이 비어있음 - 비고 업데이트 건너뜀');
            return;
        }
        
        console.log('사용자 이름 변경됨:', userName);
        
        // 기본 선택된 파일들의 비고 업데이트
        this._updateFileRemarks(selectedFiles, '기본');
        
        // 추가 파일들의 비고도 업데이트
        if (window.additionalSelectedFiles && window.additionalSelectedFiles.length > 0) {
            this._updateFileRemarks(window.additionalSelectedFiles, '추가');
        }
        
        // 파일 목록 UI 업데이트는 하지 않음 (성능상 이유)
        // 필요시 사용자가 카테고리 변경 등으로 트리거
    },
    
    // 파일 배열의 비고 업데이트
    _updateFileRemarks(fileArray, arrayType) {
        let updatedCount = 0;
        
        fileArray.forEach((fileData, index) => {
            const userName = UTILS.getFormValue('user_name');
            
            if (fileData.category === '6320') {
                // 국내출장은 출장내용과 용도가 있을 때만 업데이트
                if (fileData.businessContent && fileData.purpose) {
                    const newRemark = RemarkUtils.generateBusinessTrip(
                        fileData.businessContent,
                        userName,
                        fileData.additionalNames,
                        fileData.purpose
                    );
                    fileArray[index].remarks = newRemark;
                    updatedCount++;
                }
            } else {
                // 일반 카테고리는 항상 업데이트
                let finalName = userName;
                if (fileData.additionalNames) {
                    finalName = `${userName},${fileData.additionalNames}`;
                }
                
                const categoryLabel = CategoryUtils.getLabel(fileData.category);
                const newRemark = `임시_${finalName}_${categoryLabel}`;
                fileArray[index].remarks = newRemark;
                updatedCount++;
            }
        });
        
        if (updatedCount > 0) {
            console.log(`${arrayType} 파일 ${updatedCount}개의 비고가 업데이트됨`);
        }
    },
    
    // 전체 리셋 처리
    _handleReset() {
        const confirmMessage = '모든 데이터가 초기화됩니다. 계속하시겠습니까?';
        
        if (!confirm(confirmMessage)) {
            console.log('리셋 취소됨');
            return;
        }
        
        console.log('애플리케이션 리셋 시작');
        
        try {
            // 각 모듈 리셋
            OCRProcessor.reset();
            FileManager.reset();
            ResultsManager.reset();
            
            // 전역 변수 초기화
            selectedFiles = [];
            ocrResults = [];
            window.additionalSelectedFiles = [];
            
            // 폼 초기화 (사용자 이름 제외)
            this._resetFormExceptUserName();
            
            // 메시지 숨기기
            UIUtils.hideMessages();
            
            // 화면 상태 초기화
            UTILS.toggleElement('resultsSection', false);
            UTILS.toggleElement('uploadForm', true);
            
            console.log('애플리케이션 리셋 완료');
            
            // 리셋 완료 메시지
            setTimeout(() => {
                UIUtils.showSuccess('✅ 애플리케이션이 초기화되었습니다.');
            }, 100);
            
        } catch (error) {
            console.error('리셋 중 오류 발생:', error);
            UIUtils.showError('리셋 중 오류가 발생했습니다.');
        }
    },
    
    // 폼 초기화 (사용자 이름 제외)
    _resetFormExceptUserName() {
        const preserveFields = ['user_name']; // 보존할 필드들
        
        FORM_FIELDS.forEach(fieldId => {
            if (!preserveFields.includes(fieldId)) {
                const element = document.getElementById(fieldId);
                if (element) {
                    if (element.type === 'file') {
                        element.value = '';
                    } else if (element.tagName === 'SELECT') {
                        element.selectedIndex = 0;
                    } else {
                        element.value = '';
                    }
                }
            }
        });
        
        console.log('폼이 초기화됨 (사용자 이름 제외)');
    },
    
    // 에러 핸들링
    _handleGlobalError(error) {
        console.error('전역 오류:', error);
        
        // 에러 정보 수집
        const errorInfo = {
            message: error.message || '알 수 없는 오류',
            stack: error.stack || 'Stack trace 없음',
            timestamp: new Date().toISOString(),
            userAgent: navigator.userAgent,
            url: window.location.href
        };
        
        console.error('에러 상세 정보:', errorInfo);
        
        // 사용자에게 친화적인 메시지 표시
        UIUtils.showError('예상치 못한 오류가 발생했습니다. 페이지를 새로고침해 주세요.');
        
        // 개발 모드에서는 상세 에러 정보도 표시
        if (StorageManager.isDevMode()) {
            console.error('개발 모드 - 상세 에러:', errorInfo);
        }
    },
    
    // 애플리케이션 상태 확인
    getAppStatus() {
        return {
            selectedFilesCount: selectedFiles.length,
            additionalFilesCount: window.additionalSelectedFiles ? window.additionalSelectedFiles.length : 0,
            ocrResultsCount: ocrResults.length,
            isProcessing: OCRProcessor.isProcessing || false,
            currentScreen: this._getCurrentScreen()
        };
    },
    
    // 현재 화면 상태 확인
    _getCurrentScreen() {
        const resultsSection = document.getElementById('resultsSection');
        const uploadForm = document.getElementById('uploadForm');
        
        if (resultsSection && resultsSection.style.display !== 'none') {
            return 'results';
        } else if (uploadForm && uploadForm.style.display !== 'none') {
            return 'upload';
        } else {
            return 'unknown';
        }
    }
};

// 전역 에러 핸들러 설정
window.addEventListener('error', (event) => {
    App._handleGlobalError(event.error);
});

window.addEventListener('unhandledrejection', (event) => {
    App._handleGlobalError(event.reason);
    event.preventDefault(); // 콘솔 에러 출력 방지
});

// 전역에서 접근 가능한 함수들 (HTML onclick 등에서 사용)
window.handleFormSubmit = (e) => {
    if (e) e.preventDefault();
    console.log('전역 handleFormSubmit 호출');
    OCRProcessor.process();
};

window.handleAdditionalOCR = () => {
    console.log('전역 handleAdditionalOCR 호출');
    OCRProcessor.processAdditional();
};

window.handleDownload = () => {
    console.log('전역 handleDownload 호출');
    ResultsManager.downloadExcel();
};

window.handleReset = () => {
    console.log('전역 handleReset 호출');
    App._handleReset();
};

// 개발자 편의 함수들
window.getAppStatus = () => {
    return App.getAppStatus();
};

window.debugApp = () => {
    console.log('=== 애플리케이션 디버그 정보 ===');
    console.log('상태:', App.getAppStatus());
    console.log('선택된 파일들:', selectedFiles);
    console.log('추가 파일들:', window.additionalSelectedFiles);
    console.log('OCR 결과들:', ocrResults);
    console.log('현재 화면:', App._getCurrentScreen());
};

// 디버그용 전역 접근 (개발 모드에서만)
if (StorageManager.isDevMode()) {
    window.App = App;
    window.selectedFiles = selectedFiles;
    window.ocrResults = ocrResults;
    window.FileManager = FileManager;
    window.OCRProcessor = OCRProcessor;
    window.ResultsManager = ResultsManager;
    window.StorageManager = StorageManager;
    
    console.log('🛠️ 디버그 모드: 전역 객체들이 window에 노출되었습니다.');
    console.log('사용 가능한 디버그 함수: getAppStatus(), debugApp()');
}

// 애플리케이션 시작
App.init();

console.log('main.js 로드 완료 ✅');