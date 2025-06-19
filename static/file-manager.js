// file-manager.js - 통합된 파일 선택, 목록 관리, 드래그앤드롭

const FileManager = {
    // 파일 선택 처리 (기본 업로드)
    handleFileSelect(e) {
        const files = Array.from(e.target.files);
        console.log('기본 파일 선택됨:', files.length, '개');
        this._processFileSelection(files, 'basic');
    },

    // 추가 파일 선택 처리 (OCR 결과 화면에서)
    handleAdditionalFileSelect(e) {
        const files = Array.from(e.target.files);
        console.log('추가 파일 선택됨:', files.length, '개');
        this._processFileSelection(files, 'additional');
    },

    // 통합된 파일 선택 처리 로직
    _processFileSelection(files, type) {
        if (files.length === 0) {
            console.log('선택된 파일이 없습니다');
            return;
        }

        // 현재 파일 배열과 최대 허용 개수 확인
        const currentFiles = this._getCurrentFileArray(type);
        const maxFiles = this._getMaxFilesForType(type);
        
        // 파일 개수 제한 체크
        if (currentFiles.length + files.length > maxFiles) {
            const message = type === 'basic' 
                ? `최대 ${maxFiles}개의 파일만 선택할 수 있습니다.`
                : `총 파일 개수가 너무 많습니다. 한 번에 최대 ${maxFiles}개까지 처리 가능합니다.`;
            UIUtils.showError(message);
            return;
        }

        // 파일 추가 처리
        let addedCount = 0;
        files.forEach(file => {
            if (this._addFileToArray(file, type)) {
                addedCount++;
            }
        });

        // UI 업데이트
        this._updateFileListByType(type);

        // 결과 메시지
        if (addedCount > 0) {
            const typeLabel = type === 'basic' ? '' : '추가 ';
            UIUtils.showSuccess(`✅ ${addedCount}개 ${typeLabel}파일이 추가되었습니다.`);
        }
    },

    // 현재 파일 배열 반환
    _getCurrentFileArray(type) {
        if (type === 'additional') {
            if (!window.additionalSelectedFiles) {
                window.additionalSelectedFiles = [];
            }
            return window.additionalSelectedFiles;
        }
        return selectedFiles;
    },

    // 타입별 최대 파일 개수 반환
    _getMaxFilesForType(type) {
        if (type === 'additional') {
            const currentResults = ocrResults ? ocrResults.length : 0;
            const currentAdditional = window.additionalSelectedFiles ? window.additionalSelectedFiles.length : 0;
            return CONFIG.MAX_FILES * 2 - currentResults - currentAdditional;
        }
        return CONFIG.MAX_FILES;
    },

    // 파일을 해당 배열에 추가
    _addFileToArray(file, type) {
        // 파일 유효성 검사
        const validation = ValidationUtils.validateFile(file);
        if (!validation.isValid) {
            UIUtils.showError(validation.message);
            return false;
        }

        const fileArray = this._getCurrentFileArray(type);

        // 중복 체크
        if (this._isDuplicateFile(file, type)) {
            console.log('중복 파일 건너뜀:', file.name);
            if (type === 'additional') {
                UIUtils.showError(`파일 "${file.name}"은 이미 처리되었거나 선택된 파일입니다.`);
            }
            return false;
        }

        // 파일 메타데이터 생성 및 추가
        const fileData = this._createFileData(file);
        fileArray.push(fileData);

        console.log(`${type === 'additional' ? '추가 ' : ''}파일 추가됨:`, file.name, '자동 카테고리:', fileData.category);
        return true;
    },

    // 중복 파일 체크 (통합)
    _isDuplicateFile(file, type) {
        const currentArray = this._getCurrentFileArray(type);
        
        // 현재 배열 내 중복 체크
        const isDuplicateInCurrent = currentArray.some(existingFile => 
            existingFile.file.name === file.name && existingFile.file.size === file.size
        );

        if (isDuplicateInCurrent) {
            return true;
        }

        // 추가 파일의 경우 OCR 결과와도 비교
        if (type === 'additional') {
            const isDuplicateInResults = ocrResults && ocrResults.some(result => 
                result.fileName === file.name
            );
            
            // 기본 파일들과도 비교
            const isDuplicateInBasic = selectedFiles.some(existingFile => 
                existingFile.file.name === file.name && existingFile.file.size === file.size
            );

            return isDuplicateInResults || isDuplicateInBasic;
        }

        return false;
    },

    // 파일 데이터 객체 생성 (공통)
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

    // 타입별 파일 목록 업데이트
    _updateFileListByType(type) {
        if (type === 'additional') {
            this.updateAdditionalFileList();
        } else {
            this.updateFileList();
        }
    },
    
    // 기본 파일 목록 업데이트
    updateFileList() {
        this._updateFileListGeneric('fileList', selectedFiles);
    },

    // 추가 파일 목록 업데이트
    updateAdditionalFileList() {
        const additionalFiles = window.additionalSelectedFiles || [];
        this._updateFileListGeneric('additionalFileList', additionalFiles);
    },

    // 통합된 파일 목록 업데이트 로직
    _updateFileListGeneric(containerId, fileArray) {
        const fileList = document.getElementById(containerId);
        if (!fileList) return;
        
        if (fileArray.length === 0) {
            fileList.style.display = 'none';
            return;
        }
        
        fileList.style.display = 'block';
        fileList.innerHTML = '';
        
        const isAdditional = containerId === 'additionalFileList';
        
        // 파일 아이템들 생성
        fileArray.forEach((fileData, index) => {
            const fileItem = this._createFileItem(fileData, index, isAdditional);
            fileList.appendChild(fileItem);
        });
        
        // 요약 정보 추가
        const summaryInfo = this._createSummaryInfo(fileArray);
        fileList.appendChild(summaryInfo);
    },
    
    // 파일 아이템 DOM 생성
    _createFileItem(fileData, index, isAdditional = false) {
        const fileItem = document.createElement('div');
        fileItem.className = 'file-item';
        
        // 파일 정보 섹션
        const fileInfoSection = this._createFileInfoSection(fileData);
        fileItem.appendChild(fileInfoSection);
        
        // 카테고리 선택 섹션
        const categorySection = this._createCategorySection(fileData, index, isAdditional);
        fileItem.appendChild(categorySection);
        
        // 카테고리별 추가 필드들
        if (fileData.category === '6320') {
            // 국내출장: 출장내용, 추가이름, 용도
            fileItem.appendChild(this._createBusinessContentSection(fileData, index, isAdditional));
            fileItem.appendChild(this._createAdditionalNamesSection(fileData, index, isAdditional));
            fileItem.appendChild(this._createPurposeSection(fileData, index, isAdditional));
        } else {
            // 일반 카테고리: 추가이름만
            fileItem.appendChild(this._createAdditionalNamesSection(fileData, index, isAdditional));
        }
        
        // 삭제 버튼
        const removeBtn = this._createRemoveButton(index, isAdditional);
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
    _createCategorySection(fileData, index, isAdditional) {
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
            this._handleCategoryChange(index, e.target.value, isAdditional);
        });
        
        section.appendChild(label);
        section.appendChild(select);
        
        return section;
    },
    
    // 추가 이름 섹션 생성
    _createAdditionalNamesSection(fileData, index, isAdditional) {
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
            this._handleAdditionalNamesChange(index, e.target.value, isAdditional);
        });
        
        section.appendChild(label);
        section.appendChild(input);
        
        return section;
    },
    
    // 출장내용 섹션 생성 (국내출장 전용)
    _createBusinessContentSection(fileData, index, isAdditional) {
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
            this._handleBusinessContentChange(index, e.target.value, isAdditional);
        });
        
        section.appendChild(label);
        section.appendChild(input);
        
        return section;
    },
    
    // 용도 섹션 생성 (국내출장 전용)
    _createPurposeSection(fileData, index, isAdditional) {
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
            this._handlePurposeChange(index, e.target.value, isAdditional);
        });
        
        section.appendChild(label);
        section.appendChild(input);
        
        return section;
    },
    
    // 삭제 버튼 생성
    _createRemoveButton(index, isAdditional) {
        const button = document.createElement('button');
        button.className = 'remove-file';
        button.textContent = '삭제';
        button.type = 'button';
        button.onclick = () => this.removeFile(index, isAdditional);
        
        return button;
    },
    
    // 요약 정보 생성
    _createSummaryInfo(fileArray) {
        const summaryDiv = document.createElement('div');
        summaryDiv.className = 'file-summary';
        
        const title = document.createElement('div');
        title.className = 'summary-title';
        title.textContent = '파일 카테고리별 요약';
        summaryDiv.appendChild(title);
        
        // 카테고리별 개수 계산
        const categoryCounts = {};
        fileArray.forEach(fileData => {
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
        totalCount.textContent = `${fileArray.length}개`;
        
        totalItem.appendChild(totalLabel);
        totalItem.appendChild(totalCount);
        summaryDiv.appendChild(totalItem);
        
        return summaryDiv;
    },
    
    // 이벤트 핸들러들 (통합)
    _handleCategoryChange(index, newCategory, isAdditional) {
        const fileArray = this._getCurrentFileArray(isAdditional ? 'additional' : 'basic');
        fileArray[index].category = newCategory;
        
        // 국내출장으로 변경 시 파일명에서 용도 자동 추출
        if (newCategory === '6320') {
            const purpose = FilenameUtils.extractPurposeForBusinessTrip(fileArray[index].file.name);
            fileArray[index].purpose = purpose;
        }
        
        this._updateRemark(index, isAdditional);
        this._updateFileListByType(isAdditional ? 'additional' : 'basic');
    },
    
    _handleAdditionalNamesChange(index, value, isAdditional) {
        const fileArray = this._getCurrentFileArray(isAdditional ? 'additional' : 'basic');
        fileArray[index].additionalNames = value;
        this._updateRemark(index, isAdditional);
    },
    
    _handleBusinessContentChange(index, value, isAdditional) {
        const fileArray = this._getCurrentFileArray(isAdditional ? 'additional' : 'basic');
        fileArray[index].businessContent = value;
        this._updateBusinessTripRemark(index, isAdditional);
    },
    
    _handlePurposeChange(index, value, isAdditional) {
        const fileArray = this._getCurrentFileArray(isAdditional ? 'additional' : 'basic');
        fileArray[index].purpose = value;
        this._updateBusinessTripRemark(index, isAdditional);
    },
    
    // 비고 업데이트 (일반 카테고리)
    _updateRemark(index, isAdditional) {
        const userName = UTILS.getFormValue('user_name');
        if (!userName) return;
        
        const fileArray = this._getCurrentFileArray(isAdditional ? 'additional' : 'basic');
        const fileData = fileArray[index];
        
        if (fileData.category === '6320') {
            this._updateBusinessTripRemark(index, isAdditional);
            return;
        }
        
        // 일반 카테고리 임시 비고 생성
        let finalName = userName;
        if (fileData.additionalNames) {
            finalName = `${userName},${fileData.additionalNames}`;
        }
        
        const categoryLabel = CategoryUtils.getLabel(fileData.category);
        const newRemark = `임시_${finalName}_${categoryLabel}`;
        
        fileArray[index].remarks = newRemark;
        console.log(`${isAdditional ? '추가 ' : ''}파일 ${index + 1} 임시 비고 설정: ${newRemark}`);
    },
    
    // 국내출장 비고 업데이트
    _updateBusinessTripRemark(index, isAdditional) {
        const userName = UTILS.getFormValue('user_name');
        const fileArray = this._getCurrentFileArray(isAdditional ? 'additional' : 'basic');
        const fileData = fileArray[index];
        
        const newRemark = RemarkUtils.generateBusinessTrip(
            fileData.businessContent,
            userName,
            fileData.additionalNames,
            fileData.purpose
        );
        
        fileArray[index].remarks = newRemark;
        console.log(`${isAdditional ? '추가 ' : ''}국내출장 파일 ${index + 1} 비고 업데이트: ${newRemark}`);
    },
    
    // 파일 제거
    removeFile(index, isAdditional = false) {
        const fileArray = this._getCurrentFileArray(isAdditional ? 'additional' : 'basic');
        fileArray.splice(index, 1);
        this._updateFileListByType(isAdditional ? 'additional' : 'basic');
    },
    
    // 드래그 앤 드롭 설정
    setupDragAndDrop() {
        console.log('드래그앤드롭 설정 시작');
        
        const targets = [document.body, document.querySelector('.container')];
        
        targets.forEach(target => {
            if (!target) return;
            
            // 기본 동작 방지
            ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
                target.addEventListener(eventName, this._preventDefaults, false);
                document.addEventListener(eventName, this._preventDefaults, false);
            });
            
            // 하이라이트 효과
            ['dragenter', 'dragover'].forEach(eventName => {
                target.addEventListener(eventName, () => this._highlight(), false);
            });
            
            ['dragleave', 'drop'].forEach(eventName => {
                target.addEventListener(eventName, () => this._unhighlight(), false);
            });
            
            // 드롭 이벤트
            target.addEventListener('drop', (e) => this._handleDrop(e), false);
        });
        
        console.log('드래그앤드롭 설정 완료');
    },
    
    // 드래그 이벤트 유틸리티
    _preventDefaults(e) {
        e.preventDefault();
        e.stopPropagation();
        return false;
    },
    
    _highlight() {
        const container = document.querySelector('.container');
        if (container) {
            container.classList.add('drag-over');
        }
    },
    
    _unhighlight() {
        const container = document.querySelector('.container');
        if (container) {
            container.classList.remove('drag-over');
        }
    },
    
    // 통합된 드롭 처리
    _handleDrop(e) {
        console.log('드롭 이벤트 발생:', e);
        this._preventDefaults(e);
        
        const dt = e.dataTransfer;
        const files = Array.from(dt.files);
        
        if (files.length === 0) {
            console.log('드롭된 파일이 없습니다');
            return;
        }
        
        // 이미지 파일만 필터링
        const validFiles = files.filter(file => UTILS.isSupportedFileType(file));
        
        if (validFiles.length === 0) {
            UIUtils.showError('지원하지 않는 파일 형식입니다. JPG, PNG, PDF, TIF 파일만 업로드 가능합니다.');
            return;
        }
        
        // 현재 화면 상태에 따라 적절한 타입으로 처리
        const resultsSection = document.getElementById('resultsSection');
        const isResultsVisible = resultsSection && resultsSection.style.display !== 'none';
        
        const type = isResultsVisible ? 'additional' : 'basic';
        this._processFileSelection(validFiles, type);
    },
    
    // 초기화
    reset() {
        selectedFiles = [];
        window.additionalSelectedFiles = [];
        
        this.updateFileList();
        this.updateAdditionalFileList();
        
        // 파일 입력 초기화
        const imagesInput = document.getElementById('images');
        if (imagesInput) imagesInput.value = '';
        
        const additionalImagesInput = document.getElementById('additionalImages');
        if (additionalImagesInput) additionalImagesInput.value = '';
        
        console.log('파일 매니저 리셋 완료');
    }
};

console.log('file-manager.js 로드됨');