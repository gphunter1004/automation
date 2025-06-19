// file-manager.js - 파일 선택, 목록 관리, 드래그앤드롭 (통합 정리)

const FileManager = {
    // 파일 선택 처리
    handleFileSelect(e) {
        const files = Array.from(e.target.files);
        
        console.log('파일 선택됨:', files.length, '개');
        
        // 파일 개수 제한 체크
        if (selectedFiles.length + files.length > CONFIG.MAX_FILES) {
            UIUtils.showError(`최대 ${CONFIG.MAX_FILES}개의 파일만 선택할 수 있습니다.`);
            return;
        }
        
        // 파일 추가 (중복 체크 및 유효성 검사)
        files.forEach(file => {
            this._addFileToSelection(file);
        });
        
        this.updateFileList();
    },
    
    // 파일을 선택 목록에 추가
    _addFileToSelection(file) {
        // 중복 체크
        const isDuplicate = selectedFiles.some(existingFile => 
            existingFile.file.name === file.name && existingFile.file.size === file.size
        );
        
        if (isDuplicate) {
            console.log('중복 파일 건너뜀:', file.name);
            return;
        }
        
        // 파일 유효성 검사
        const validation = ValidationUtils.validateFile(file);
        if (!validation.isValid) {
            UIUtils.showError(validation.message);
            return;
        }
        
        // 파일 메타데이터 생성
        const fileData = this._createFileData(file);
        selectedFiles.push(fileData);
        
        console.log('파일 추가됨:', file.name, '자동 카테고리:', fileData.category);
    },
    
    // 파일 데이터 객체 생성
    _createFileData(file) {
        const category = CategoryUtils.detectFromFilename(file.name);
        const extractedNames = FilenameUtils.extractNamesFromFilename(file.name);
        
        const fileData = {
            file: file,
            category: category,
            additionalNames: extractedNames,
            remarks: '',
            businessContent: '',
            purpose: ''
        };
        
        // 국내출장인 경우 파일명에서 용도 자동 추출
        if (category === '6320') {
            fileData.purpose = FilenameUtils.extractPurposeForBusinessTrip(file.name);
        }
        
        return fileData;
    },
    
    // 파일 목록 업데이트
    updateFileList() {
        const fileList = document.getElementById('fileList');
        if (!fileList) return;
        
        if (selectedFiles.length === 0) {
            fileList.style.display = 'none';
            return;
        }
        
        fileList.style.display = 'block';
        fileList.innerHTML = '';
        
        // 파일 아이템들 생성
        selectedFiles.forEach((fileData, index) => {
            const fileItem = this._createFileItem(fileData, index);
            fileList.appendChild(fileItem);
        });
        
        // 요약 정보 추가
        const summaryInfo = this._createSummaryInfo();
        fileList.appendChild(summaryInfo);
    },
    
    // 파일 아이템 DOM 생성
    _createFileItem(fileData, index) {
        const fileItem = document.createElement('div');
        fileItem.className = CONFIG.CSS_CLASSES.FILE_ITEM;
        
        // 파일 정보 섹션
        const fileInfoSection = this._createFileInfoSection(fileData);
        fileItem.appendChild(fileInfoSection);
        
        // 카테고리 선택 섹션
        const categorySection = this._createCategorySection(fileData, index);
        fileItem.appendChild(categorySection);
        
        // 카테고리별 추가 필드들
        if (fileData.category === '6320') {
            // 국내출장: 출장내용, 추가이름, 용도
            fileItem.appendChild(this._createBusinessContentSection(fileData, index));
            fileItem.appendChild(this._createAdditionalNamesSection(fileData, index));
            fileItem.appendChild(this._createPurposeSection(fileData, index));
        } else {
            // 일반 카테고리: 추가이름만
            fileItem.appendChild(this._createAdditionalNamesSection(fileData, index));
        }
        
        // 삭제 버튼
        const removeBtn = this._createRemoveButton(index);
        fileItem.appendChild(removeBtn);
        
        return fileItem;
    },
    
    // 파일 정보 섹션 생성
    _createFileInfoSection(fileData) {
        const section = document.createElement('div');
        section.className = 'file-info-section';
        
        const fileName = document.createElement('div');
        fileName.className = 'file-name';
        fileName.textContent = fileData.file.name;
        
        const fileSize = document.createElement('div');
        fileSize.className = 'file-size';
        fileSize.textContent = UTILS.formatFileSize(fileData.file.size);
        
        // 자동 감지 정보 표시 (기본값이 아닌 경우)
        if (fileData.category !== '6130') {
            const autoInfo = document.createElement('div');
            autoInfo.className = 'auto-detected-info';
            autoInfo.textContent = `🤖 자동 감지: ${CategoryUtils.getLabel(fileData.category)}`;
            section.appendChild(autoInfo);
        }
        
        section.appendChild(fileName);
        section.appendChild(fileSize);
        
        return section;
    },
    
    // 카테고리 선택 섹션 생성
    _createCategorySection(fileData, index) {
        const section = document.createElement('div');
        section.className = 'category-section';
        
        const label = document.createElement('label');
        label.className = 'category-label';
        label.textContent = '카테고리';
        
        const select = document.createElement('select');
        select.className = 'category-select';
        
        // 옵션 추가
        CATEGORY_OPTIONS.forEach(option => {
            const optionElement = document.createElement('option');
            optionElement.value = option.value;
            optionElement.textContent = option.label;
            if (option.value === fileData.category) {
                optionElement.selected = true;
            }
            select.appendChild(optionElement);
        });
        
        // 카테고리 변경 이벤트
        select.addEventListener('change', (e) => {
            this._handleCategoryChange(index, e.target.value);
        });
        
        section.appendChild(label);
        section.appendChild(select);
        
        return section;
    },
    
    // 추가 이름 섹션 생성
    _createAdditionalNamesSection(fileData, index) {
        const section = document.createElement('div');
        section.className = 'additional-names-section';
        
        const label = document.createElement('label');
        label.className = 'additional-names-label';
        label.textContent = '추가 이름 (선택사항)';
        
        const input = document.createElement('input');
        input.type = 'text';
        input.className = 'additional-names-input';
        input.placeholder = '다른 사람 이름을 입력하세요';
        input.value = fileData.additionalNames || '';
        
        input.addEventListener('input', (e) => {
            this._handleAdditionalNamesChange(index, e.target.value);
        });
        
        section.appendChild(label);
        section.appendChild(input);
        
        return section;
    },
    
    // 출장내용 섹션 생성 (국내출장 전용)
    _createBusinessContentSection(fileData, index) {
        const section = document.createElement('div');
        section.className = 'business-content-section';
        
        const label = document.createElement('label');
        label.className = 'business-content-label';
        label.textContent = '출장내용 *';
        
        const input = document.createElement('input');
        input.type = 'text';
        input.className = 'business-content-input';
        input.placeholder = '출장내용을 입력하세요';
        input.value = fileData.businessContent || '';
        input.required = true;
        
        input.addEventListener('input', (e) => {
            this._handleBusinessContentChange(index, e.target.value);
        });
        
        section.appendChild(label);
        section.appendChild(input);
        
        return section;
    },
    
    // 용도 섹션 생성 (국내출장 전용)
    _createPurposeSection(fileData, index) {
        const section = document.createElement('div');
        section.className = 'purpose-section';
        
        const label = document.createElement('label');
        label.className = 'purpose-label';
        label.textContent = '용도 *';
        
        const input = document.createElement('input');
        input.type = 'text';
        input.className = 'purpose-input';
        input.placeholder = '용도를 입력하세요';
        input.value = fileData.purpose || '';
        input.required = true;
        
        input.addEventListener('input', (e) => {
            this._handlePurposeChange(index, e.target.value);
        });
        
        section.appendChild(label);
        section.appendChild(input);
        
        return section;
    },
    
    // 삭제 버튼 생성
    _createRemoveButton(index) {
        const button = document.createElement('button');
        button.className = 'remove-file';
        button.textContent = '삭제';
        button.type = 'button';
        button.onclick = () => this.removeFile(index);
        
        return button;
    },
    
    // 요약 정보 생성
    _createSummaryInfo() {
        const summaryDiv = document.createElement('div');
        summaryDiv.className = 'file-summary';
        summaryDiv.id = 'fileSummary';
        
        const title = document.createElement('div');
        title.className = 'summary-title';
        title.textContent = '파일 카테고리별 요약';
        summaryDiv.appendChild(title);
        
        // 카테고리별 개수 계산
        const categoryCounts = {};
        selectedFiles.forEach(fileData => {
            const label = CategoryUtils.getLabel(fileData.category);
            categoryCounts[label] = (categoryCounts[label] || 0) + 1;
        });
        
        // 요약 아이템들 생성
        Object.entries(categoryCounts).forEach(([category, count]) => {
            const item = document.createElement('div');
            item.className = 'summary-item';
            
            const label = document.createElement('span');
            label.className = 'summary-label';
            label.textContent = category;
            
            const countSpan = document.createElement('span');
            countSpan.className = 'summary-count';
            countSpan.textContent = `${count}개`;
            
            item.appendChild(label);
            item.appendChild(countSpan);
            summaryDiv.appendChild(item);
        });
        
        // 총계
        const totalItem = document.createElement('div');
        totalItem.className = 'summary-item';
        totalItem.style.cssText = 'border-top: 1px solid #ddd; padding-top: 5px; margin-top: 5px; font-weight: bold;';
        
        const totalLabel = document.createElement('span');
        totalLabel.className = 'summary-label';
        totalLabel.textContent = '총 파일 수';
        
        const totalCount = document.createElement('span');
        totalCount.className = 'summary-count';
        totalCount.textContent = `${selectedFiles.length}개`;
        
        totalItem.appendChild(totalLabel);
        totalItem.appendChild(totalCount);
        summaryDiv.appendChild(totalItem);
        
        return summaryDiv;
    },
    
    // 이벤트 핸들러들
    _handleCategoryChange(index, newCategory) {
        selectedFiles[index].category = newCategory;
        
        // 국내출장으로 변경 시 파일명에서 용도 자동 추출
        if (newCategory === '6320') {
            const purpose = FilenameUtils.extractPurposeForBusinessTrip(selectedFiles[index].file.name);
            selectedFiles[index].purpose = purpose;
        }
        
        this.updateSummaryInfo();
        this._updateRemark(index);
        
        // 파일 목록 전체 재구성 (국내출장 필드 표시/숨김)
        this.updateFileList();
    },
    
    _handleAdditionalNamesChange(index, value) {
        selectedFiles[index].additionalNames = value;
        this._updateRemark(index);
    },
    
    _handleBusinessContentChange(index, value) {
        selectedFiles[index].businessContent = value;
        this._updateBusinessTripRemark(index);
    },
    
    _handlePurposeChange(index, value) {
        selectedFiles[index].purpose = value;
        this._updateBusinessTripRemark(index);
    },
    
    // 비고 업데이트 (일반 카테고리)
    _updateRemark(index) {
        const userName = UTILS.getFormValue('user_name');
        if (!userName) return;
        
        const fileData = selectedFiles[index];
        
        if (fileData.category === '6320') {
            this._updateBusinessTripRemark(index);
            return;
        }
        
        // 일반 카테고리 임시 비고 생성
        let finalName = userName;
        if (fileData.additionalNames) {
            finalName = `${userName},${fileData.additionalNames}`;
        }
        
        const categoryLabel = CategoryUtils.getLabel(fileData.category);
        const newRemark = `임시_${finalName}_${categoryLabel}`;
        
        selectedFiles[index].remarks = newRemark;
        console.log(`파일 ${index + 1} 임시 비고 설정: ${newRemark}`);
    },
    
    // 국내출장 비고 업데이트
    _updateBusinessTripRemark(index) {
        const userName = UTILS.getFormValue('user_name');
        const fileData = selectedFiles[index];
        
        const newRemark = RemarkUtils.generateBusinessTrip(
            fileData.businessContent,
            userName,
            fileData.additionalNames,
            fileData.purpose
        );
        
        selectedFiles[index].remarks = newRemark;
        console.log(`국내출장 파일 ${index + 1} 비고 업데이트: ${newRemark}`);
    },
    
    // 요약 정보만 업데이트
    updateSummaryInfo() {
        const existingSummary = document.getElementById('fileSummary');
        if (existingSummary) {
            const newSummary = this._createSummaryInfo();
            existingSummary.replaceWith(newSummary);
        }
    },
    
    // 파일 제거
    removeFile(index) {
        selectedFiles.splice(index, 1);
        this.updateFileList();
    },
    
    // 드래그 앤 드롭 설정
    setupDragAndDrop() {
        const container = document.querySelector('.container');
        if (!container) return;
        
        // 기본 동작 방지
        ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
            container.addEventListener(eventName, this._preventDefaults, false);
        });
        
        // 하이라이트 효과
        ['dragenter', 'dragover'].forEach(eventName => {
            container.addEventListener(eventName, this._highlight, false);
        });
        
        ['dragleave', 'drop'].forEach(eventName => {
            container.addEventListener(eventName, this._unhighlight, false);
        });
        
        container.addEventListener('drop', (e) => this._handleDrop(e), false);
    },
    
    // 드래그 이벤트 유틸리티
    _preventDefaults(e) {
        e.preventDefault();
        e.stopPropagation();
    },
    
    _highlight(e) {
        const container = document.querySelector('.container');
        if (container) {
            container.classList.add(CONFIG.CSS_CLASSES.DRAG_OVER);
        }
    },
    
    _unhighlight(e) {
        const container = document.querySelector('.container');
        if (container) {
            container.classList.remove(CONFIG.CSS_CLASSES.DRAG_OVER);
        }
    },
    
    // 드롭 처리
    _handleDrop(e) {
        const dt = e.dataTransfer;
        const files = Array.from(dt.files);
        
        // 이미지 파일만 필터링
        const validFiles = files.filter(file => UTILS.isSupportedFileType(file));
        
        if (selectedFiles.length + validFiles.length > CONFIG.MAX_FILES) {
            UIUtils.showError(`최대 ${CONFIG.MAX_FILES}개의 파일만 선택할 수 있습니다.`);
            return;
        }
        
        validFiles.forEach(file => {
            this._addFileToSelection(file);
        });
        
        this.updateFileList();
    },
    
    // 초기화
    reset() {
        selectedFiles = [];
        this.updateFileList();
        
        const imagesInput = document.getElementById('images');
        if (imagesInput) {
            imagesInput.value = '';
        }
    }
};

console.log('file-manager.js 로드됨');