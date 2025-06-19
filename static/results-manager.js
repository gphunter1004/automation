// results-manager.js - OCR 결과 관리 및 Excel 다운로드

const ResultsManager = {
    // OCR 결과 화면 표시
    showResults(results) {
        console.log('OCR 결과 표시:', results);
        
        if (!results || results.length === 0) {
            UIUtils.showError('표시할 OCR 결과가 없습니다.');
            return;
        }
        
        // 전역 변수에 결과 저장
        ocrResults = results;
        
        // 결과 테이블 생성
        this._displayResultsTable(results);
        
        // 화면 전환
        UTILS.toggleElement('uploadForm', false);
        UTILS.toggleElement('resultsSection', true);
    },
    
    // 결과 테이블 표시
    _displayResultsTable(results) {
        const tableBody = document.querySelector('#resultsTable tbody');
        if (!tableBody) {
            console.error('resultsTable tbody를 찾을 수 없습니다');
            return;
        }
        
        tableBody.innerHTML = '';
        
        results.forEach((result, index) => {
            const row = this._createResultRow(result, index);
            tableBody.appendChild(row);
        });
    },
    
    // 결과 행 생성
    _createResultRow(result, index) {
        const row = document.createElement('tr');
        
        // 각 컬럼 생성
        const columns = [
            this._createFileNameCell(result),
            this._createCategoryCell(result, index),
            this._createRemarkCell(result, index),
            this._createPurposeCell(result, index),
            this._createAmountCell(result, index),
            this._createIssueDateCell(result, index),
            this._createPayDateCell(result)
        ];
        
        columns.forEach(column => row.appendChild(column));
        
        return row;
    },
    
    // 파일명 셀
    _createFileNameCell(result) {
        const cell = document.createElement('td');
        cell.className = 'file-name-cell';
        cell.textContent = result.fileName;
        return cell;
    },
    
    // 카테고리 셀
    _createCategoryCell(result, index) {
        const cell = document.createElement('td');
        const select = document.createElement('select');
        
        CATEGORY_OPTIONS.forEach(option => {
            const optionElement = document.createElement('option');
            optionElement.value = option.value;
            optionElement.textContent = option.label;
            if (option.value === result.category) {
                optionElement.selected = true;
            }
            select.appendChild(optionElement);
        });
        
        select.addEventListener('change', (e) => {
            ocrResults[index].category = e.target.value;
            this._updateResultRemark(index);
        });
        
        cell.appendChild(select);
        return cell;
    },
    
    // 비고 셀
    _createRemarkCell(result, index) {
        const cell = document.createElement('td');
        const input = document.createElement('input');
        input.type = 'text';
        input.value = result.remark || '';
        
        input.addEventListener('input', (e) => {
            ocrResults[index].remark = e.target.value;
        });
        
        cell.appendChild(input);
        return cell;
    },
    
    // 사용처 셀
    _createPurposeCell(result, index) {
        const cell = document.createElement('td');
        const input = document.createElement('input');
        input.type = 'text';
        input.value = result.purpose || '';
        
        input.addEventListener('input', (e) => {
            ocrResults[index].purpose = e.target.value;
        });
        
        cell.appendChild(input);
        return cell;
    },
    
    // 사용액 셀
    _createAmountCell(result, index) {
        const cell = document.createElement('td');
        const input = document.createElement('input');
        input.type = 'text';
        input.value = result.amount || '';
        
        input.addEventListener('input', (e) => {
            ocrResults[index].amount = e.target.value;
        });
        
        cell.appendChild(input);
        return cell;
    },
    
    // 사용일 셀
    _createIssueDateCell(result, index) {
        const cell = document.createElement('td');
        const input = document.createElement('input');
        input.type = 'text';
        input.value = result.issueDate || '';
        input.placeholder = 'YYYYMMDD';
        
        input.addEventListener('input', (e) => {
            ocrResults[index].issueDate = e.target.value;
            this._updateResultRemark(index);
        });
        
        cell.appendChild(input);
        return cell;
    },
    
    // 결제일 셀 (읽기 전용)
    _createPayDateCell(result) {
        const cell = document.createElement('td');
        const input = document.createElement('input');
        input.type = 'text';
        input.value = result.payDate || '';
        input.readOnly = true;
        input.style.backgroundColor = '#f8f9fa';
        
        cell.appendChild(input);
        return cell;
    },
    
    // 결과 비고 업데이트
    _updateResultRemark(index) {
        const result = ocrResults[index];
        const userName = UTILS.getFormValue('user_name');
        
        if (!userName || !result.issueDate) return;
        
        // 현재 비고가 자동 생성된 형태인지 확인
        const currentRemark = result.remark || '';
        if (!RemarkUtils.isAutoGenerated(currentRemark) && currentRemark.trim()) {
            // 수동으로 입력된 비고는 변경하지 않음
            return;
        }
        
        // 국내출장인 경우 특별 처리 (출장내용과 용도 필요)
        if (result.category === '6320') {
            console.log(`국내출장 파일 ${index + 1}: 비고는 출장내용과 용도가 필요하므로 자동 업데이트하지 않음`);
            return;
        }
        
        // 추가 이름 찾기
        let finalUserName = userName;
        const matchingFile = selectedFiles.find(fileData => 
            fileData.file.name === result.fileName
        );
        
        if (matchingFile && matchingFile.additionalNames) {
            finalUserName = `${userName},${matchingFile.additionalNames}`;
        }
        
        // 새 비고 생성
        const newRemark = RemarkUtils.generateDefault(result.issueDate, finalUserName, result.category);
        result.remark = newRemark;
        
        // UI 업데이트
        const remarkInput = document.querySelector(`#resultsTable tbody tr:nth-child(${index + 1}) td:nth-child(3) input`);
        if (remarkInput) {
            remarkInput.value = newRemark;
        }
        
        console.log(`OCR 결과 ${index + 1} 비고 업데이트: ${newRemark}`);
    },
    
    // Excel 다운로드 처리
    async downloadExcel() {
        const downloadBtn = document.getElementById('downloadBtn');
        const originalText = downloadBtn.textContent;
        
        try {
            UIUtils.setButtonState('downloadBtn', true, '다운로드 중...');
            
            // 다운로드용 FormData 생성
            const formData = this._createDownloadFormData();
            
            // 다운로드 요청
            const response = await fetch(CONFIG.API.DOWNLOAD_EXCEL, {
                method: 'POST',
                body: formData
            });
            
            if (response.ok) {
                await this._handleDownloadResponse(response);
                UIUtils.showSuccess('✅ Excel 파일이 성공적으로 다운로드되었습니다!');
            } else {
                const errorData = await response.json();
                throw new Error(errorData.error || '다운로드 중 오류가 발생했습니다.');
            }
            
        } catch (error) {
            UIUtils.showError('❌ ' + error.message);
        } finally {
            UIUtils.setButtonState('downloadBtn', false, originalText);
        }
    },
    
    // 다운로드용 FormData 생성
    _createDownloadFormData() {
        const formData = new FormData();
        
        // 기본 정보 추가
        FORM_FIELDS.forEach(fieldId => {
            const value = UTILS.getFormValue(fieldId);
            if (value) {
                formData.append(fieldId, value);
            }
        });
        
        // OCR 결과 데이터 추가
        formData.append('excel_data', JSON.stringify(ocrResults));
        
        return formData;
    },
    
    // 다운로드 응답 처리
    async _handleDownloadResponse(response) {
        const blob = await response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        
        // Content-Disposition 헤더에서 파일명 추출
        const contentDisposition = response.headers.get('Content-Disposition');
        let filename = 'ocr_results.xlsx';
        
        if (contentDisposition) {
            const filenameMatch = contentDisposition.match(/filename=([^;]+)/);
            if (filenameMatch) {
                filename = filenameMatch[1].replace(/"/g, '');
            }
        }
        
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        
        // 정리
        window.URL.revokeObjectURL(url);
        document.body.removeChild(a);
    }
};

console.log('results-manager.js 로드됨');