// main.js - 메인 초기화 및 이벤트 바인딩

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
        const uploadForm = document.getElementById('uploadForm');
        if (uploadForm) {
            uploadForm.addEventListener('submit', (e) => {
                e.preventDefault();
                e.stopPropagation();
                console.log('Form submit 이벤트 차단됨');
                return false;
            });
        }
        
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
        
        // 파일 선택 입력
        const imagesInput = document.getElementById('images');
        if (imagesInput) {
            imagesInput.addEventListener('change', (e) => {
                FileManager.handleFileSelect(e);
            });
        }
        
        // Excel 다운로드 버튼
        const downloadBtn = document.getElementById('downloadBtn');
        if (downloadBtn) {
            downloadBtn.addEventListener('click', () => {
                console.log('Excel 다운로드 버튼 클릭');
                ResultsManager.downloadExcel();
            });
        }
        
        // 리셋 버튼
        const resetBtn = document.getElementById('resetBtn');
        if (resetBtn) {
            resetBtn.addEventListener('click', () => {
                console.log('리셋 버튼 클릭');
                this._handleReset();
            });
        }
        
        // 사용자 이름 입력 변경 시 비고 업데이트
        const userNameInput = document.getElementById('user_name');
        if (userNameInput) {
            userNameInput.addEventListener('input', () => {
                this._handleUserNameChange();
            });
        }
        
        console.log('이벤트 리스너 설정 완료');
    },
    
    // 사용자 이름 변경 처리
    _handleUserNameChange() {
        const userName = UTILS.getFormValue('user_name');
        if (!userName) return;
        
        // 선택된 파일들의 비고 업데이트
        selectedFiles.forEach((fileData, index) => {
            if (fileData.category === '6320') {
                // 국내출장은 별도 처리
                if (fileData.businessContent && fileData.purpose) {
                    const newRemark = RemarkUtils.generateBusinessTrip(
                        fileData.businessContent,
                        userName,
                        fileData.additionalNames,
                        fileData.purpose
                    );
                    selectedFiles[index].remarks = newRemark;
                }
            } else {
                // 일반 카테고리는 임시 비고 생성
                let finalName = userName;
                if (fileData.additionalNames) {
                    finalName = `${userName},${fileData.additionalNames}`;
                }
                
                const categoryLabel = CategoryUtils.getLabel(fileData.category);
                const newRemark = `임시_${finalName}_${categoryLabel}`;
                selectedFiles[index].remarks = newRemark;
            }
        });
    },
    
    // 전체 리셋 처리
    _handleReset() {
        if (confirm('모든 데이터가 초기화됩니다. 계속하시겠습니까?')) {
            // 각 모듈 리셋
            OCRProcessor.reset();
            FileManager.reset();
            
            // 메시지 숨기기
            UIUtils.hideMessages();
            
            console.log('애플리케이션 리셋 완료');
        }
    },
    
    // 에러 핸들링
    _handleGlobalError(error) {
        console.error('전역 오류:', error);
        UIUtils.showError('예상치 못한 오류가 발생했습니다. 페이지를 새로고침해 주세요.');
    }
};

// 전역 에러 핸들러 설정
window.addEventListener('error', (event) => {
    App._handleGlobalError(event.error);
});

window.addEventListener('unhandledrejection', (event) => {
    App._handleGlobalError(event.reason);
});

// 전역에서 접근 가능한 함수들 (HTML onclick 등에서 사용)
window.handleFormSubmit = (e) => {
    e.preventDefault();
    OCRProcessor.process();
};

window.handleDownload = () => {
    ResultsManager.downloadExcel();
};

window.handleReset = () => {
    App._handleReset();
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
    console.log('디버그 모드: 전역 객체들이 window에 노출되었습니다.');
}

// 애플리케이션 시작
App.init();

console.log('main.js 로드 완료');