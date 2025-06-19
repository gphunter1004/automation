// storage-manager.js - 쿠키/스토리지 관리

const StorageManager = {
    // 쿠키 설정 (Base64 인코딩으로 특수문자 문제 방지)
    setCookie(name, value, days = CONFIG.COOKIE.EXPIRY_DAYS) {
        const expires = new Date();
        expires.setTime(expires.getTime() + (days * 24 * 60 * 60 * 1000));
        const expiresString = expires.toUTCString();
        
        try {
            const encodedValue = btoa(encodeURIComponent(value));
            document.cookie = `${CONFIG.COOKIE.PREFIX}${name}=${encodedValue}; expires=${expiresString}; path=/; SameSite=Lax`;
            console.log(`쿠키 저장: ${name} = ${value}`);
            return true;
        } catch (error) {
            console.error(`쿠키 저장 실패: ${name}`, error);
            return false;
        }
    },
    
    // 쿠키 가져오기 (Base64 디코딩)
    getCookie(name) {
        const nameEQ = `${CONFIG.COOKIE.PREFIX}${name}=`;
        const cookies = document.cookie.split(';');
        
        for (let cookie of cookies) {
            let c = cookie.trim();
            if (c.indexOf(nameEQ) === 0) {
                try {
                    const encodedValue = c.substring(nameEQ.length);
                    const decodedValue = decodeURIComponent(atob(encodedValue));
                    console.log(`쿠키 불러오기: ${name} = ${decodedValue}`);
                    return decodedValue;
                } catch (error) {
                    console.error(`쿠키 디코딩 실패: ${name}`, error);
                    return '';
                }
            }
        }
        return '';
    },
    
    // 쿠키 삭제
    deleteCookie(name) {
        document.cookie = `${CONFIG.COOKIE.PREFIX}${name}=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;`;
        console.log(`쿠키 삭제: ${name}`);
    },
    
    // 폼 데이터를 쿠키에 저장
    saveFormData() {
        console.log('=== 폼 데이터 쿠키 저장 시작 ===');
        
        let savedCount = 0;
        
        FORM_FIELDS.forEach(fieldId => {
            const value = UTILS.getFormValue(fieldId);
            if (value) {
                if (this.setCookie(fieldId, value)) {
                    savedCount++;
                }
            } else {
                this.deleteCookie(fieldId);
            }
        });
        
        // 저장 시간 기록
        this.setCookie('last_saved', new Date().toISOString());
        
        console.log(`=== 폼 데이터 쿠키 저장 완료: ${savedCount}개 필드 ===`);
        return savedCount;
    },
    
    // 쿠키에서 폼 데이터 불러오기
    loadFormData() {
        console.log('=== 쿠키에서 폼 데이터 불러오기 시작 ===');
        
        let loadedCount = 0;
        
        FORM_FIELDS.forEach(fieldId => {
            const savedValue = this.getCookie(fieldId);
            if (savedValue) {
                UTILS.setFormValue(fieldId, savedValue);
                loadedCount++;
                
                // 시각적 피드백 (잠깐 하이라이트)
                this._highlightLoadedField(fieldId);
            }
        });
        
        // 마지막 저장 시간 확인 및 메시지 표시
        const lastSaved = this.getCookie('last_saved');
        if (lastSaved && loadedCount > 0) {
            this._showLoadMessage(loadedCount, lastSaved);
        }
        
        console.log(`=== 쿠키에서 폼 데이터 불러오기 완료: ${loadedCount}개 필드 ===`);
        return loadedCount;
    },
    
    // 필드 하이라이트 효과
    _highlightLoadedField(fieldId) {
        const element = document.getElementById(fieldId);
        if (element) {
            element.style.backgroundColor = '#e8f5e8';
            setTimeout(() => {
                element.style.backgroundColor = '';
            }, 1000);
        }
    },
    
    // 쿠키 로드 메시지 표시
    _showLoadMessage(loadedCount, lastSavedISO) {
        const savedDate = new Date(lastSavedISO);
        const now = new Date();
        const daysDiff = Math.floor((now - savedDate) / (1000 * 60 * 60 * 24));
        
        this._displayLoadMessage(loadedCount, daysDiff);
    },
    
    // 로드 메시지 UI 생성
    _displayLoadMessage(loadedCount, daysDiff) {
        // 기존 메시지가 있으면 제거
        const existingMessage = document.getElementById('cookieLoadMessage');
        if (existingMessage) {
            existingMessage.remove();
        }
        
        const message = document.createElement('div');
        message.id = 'cookieLoadMessage';
        message.className = 'cookie-load-message';
        
        let timeText = '';
        if (daysDiff === 0) {
            timeText = '오늘';
        } else if (daysDiff === 1) {
            timeText = '어제';
        } else {
            timeText = `${daysDiff}일 전`;
        }
        
        message.innerHTML = `
            <div class="message-content">
                <span class="message-icon">🍪</span>
                <span class="message-text">${loadedCount}개 필드가 ${timeText} 저장된 값으로 자동 입력되었습니다.</span>
                <button class="message-close" onclick="StorageManager.hideLoadMessage()">×</button>
            </div>
        `;
        
        // h1 요소 다음에 메시지 삽입
        const h1Element = document.querySelector('h1');
        if (h1Element) {
            h1Element.insertAdjacentElement('afterend', message);
        }
        
        // 5초 후 자동 숨김
        setTimeout(() => {
            this.hideLoadMessage();
        }, 5000);
    },
    
    // 쿠키 로드 메시지 숨기기
    hideLoadMessage() {
        const message = document.getElementById('cookieLoadMessage');
        if (message) {
            message.style.opacity = '0';
            setTimeout(() => {
                message.remove();
            }, 300);
        }
    },
    
    // 자동 저장 설정 (디바운스 적용)
    setupAutoSave() {
        const debouncedSave = UTILS.debounce(() => {
            this.saveFormData();
        }, 1000);
        
        FORM_FIELDS.forEach(fieldId => {
            const element = document.getElementById(fieldId);
            if (element) {
                // 입력 중 자동 저장 (1초 후)
                element.addEventListener('input', debouncedSave);
                
                // 포커스 아웃 시 즉시 저장
                element.addEventListener('blur', () => {
                    this.saveFormData();
                });
            }
        });
        
        console.log('자동 저장 이벤트 리스너 설정 완료');
    },
    
    // 모든 쿠키 삭제 (개발/테스트용)
    clearAllCookies() {
        FORM_FIELDS.forEach(fieldId => {
            this.deleteCookie(fieldId);
        });
        this.deleteCookie('last_saved');
        
        console.log('모든 폼 쿠키가 삭제되었습니다.');
        
        // 페이지 새로고침
        location.reload();
    },
    
    // 쿠키 정보 표시 (개발/디버그용)
    showCookieInfo() {
        console.log('=== 현재 저장된 쿠키 정보 ===');
        
        const cookieInfo = {};
        FORM_FIELDS.forEach(fieldId => {
            const value = this.getCookie(fieldId);
            cookieInfo[fieldId] = value || '(저장 안됨)';
        });
        
        const lastSaved = this.getCookie('last_saved');
        if (lastSaved) {
            cookieInfo['last_saved'] = new Date(lastSaved).toLocaleString();
        }
        
        console.table(cookieInfo);
        console.log('========================');
        
        return cookieInfo;
    },
    
    // 개발 모드 확인
    isDevMode() {
        return localStorage.getItem('dev_mode') === 'true';
    },
    
    // 개발자 도구 버튼 추가
    addDevTools() {
        if (!this.isDevMode()) return;
        
        const container = document.querySelector('.container');
        if (!container) return;
        
        const devSection = document.createElement('div');
        devSection.className = 'dev-section';
        devSection.innerHTML = `
            <h4 style="margin: 0 0 10px 0; color: #856404;">🛠️ 개발자 도구</h4>
            <div style="display: flex; gap: 10px; flex-wrap: wrap;">
                <button type="button" onclick="StorageManager.showCookieInfo()" 
                        style="background: #17a2b8; color: white; border: none; padding: 5px 10px; border-radius: 3px; cursor: pointer; font-size: 12px;">
                    쿠키 정보 보기
                </button>
                <button type="button" onclick="StorageManager.clearAllCookies()" 
                        style="background: #dc3545; color: white; border: none; padding: 5px 10px; border-radius: 3px; cursor: pointer; font-size: 12px;">
                    모든 쿠키 삭제
                </button>
                <button type="button" onclick="StorageManager.saveFormData()" 
                        style="background: #28a745; color: white; border: none; padding: 5px 10px; border-radius: 3px; cursor: pointer; font-size: 12px;">
                    수동 저장
                </button>
            </div>
            <p style="margin: 10px 0 0 0; font-size: 11px; color: #6c757d;">
                개발 모드: localStorage.setItem('dev_mode', 'false')로 숨길 수 있습니다.
            </p>
        `;
        
        container.appendChild(devSection);
    }
};

console.log('storage-manager.js 로드됨');