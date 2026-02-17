export const overlayStyles = /* css */ `
:host { 
  all: initial; 
  --bg-primary: #18181b;
  --bg-secondary: #1e1e22;
  --bg-tertiary: #27272a;
  --bg-card: #1e1e22;
  --bg-elevated: #27272a;
  --bg-hover: #2e2e33;
  
  --text-primary: #fafafa;
  --text-secondary: #a1a1aa;
  --text-tertiary: #71717a;
  --border: rgba(255, 255, 255, 0.08);
  --border-strong: rgba(255, 255, 255, 0.14);
  
  --accent: #3b82f6;
  --accent-hover: #2563eb;
  --accent-subtle: rgba(59, 130, 246, 0.12);
  
  --highlight-yellow: #fbbf24;
  --highlight-green: #34d399;
  --highlight-blue: #60a5fa;
  --highlight-pink: #f472b6;
  --highlight-purple: #a78bfa;
}

:host(.light) {
  --bg-primary: #fafafa;
  --bg-secondary: #ffffff;
  --bg-tertiary: #f4f4f5;
  --bg-card: #ffffff;
  --bg-elevated: #ffffff;
  --bg-hover: #f4f4f5;
  
  --text-primary: #18181b;
  --text-secondary: #71717a;
  --text-tertiary: #a1a1aa;
  --border: rgba(0, 0, 0, 0.08);
  --border-strong: rgba(0, 0, 0, 0.14);
  
  --accent: #2563eb;
  --accent-hover: #1d4ed8;
  --accent-subtle: rgba(37, 99, 235, 0.08);
}

.margin-overlay {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
}

.color-picker {
  position: absolute;
  top: 100%;
  left: 50%;
  transform: translateX(-50%);
  margin-top: 6px;
  display: flex;
  gap: 6px;
  padding: 8px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 10px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.25);
  animation: toolbar-in 0.15s ease forwards;
}

.color-dot {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  border: 2px solid transparent;
  cursor: pointer;
  transition: all 0.15s ease;
}

.color-dot:hover {
  transform: scale(1.15);
  border-color: var(--text-primary);
}

.margin-popover {
  position: absolute;
  width: 320px;
  background: var(--bg-card);
  border: 1px solid var(--border-strong);
  border-radius: 16px;
  padding: 0;
  box-shadow: 0 8px 40px rgba(0, 0, 0, 0.35), 0 0 80px rgba(59, 130, 246, 0.04);
  display: flex;
  flex-direction: column;
  pointer-events: auto;
  z-index: 2147483647;
  font-family: "Inter", system-ui, -apple-system, BlinkMacSystemFont, sans-serif;
  color: var(--text-primary);
  opacity: 0;
  transform: translateY(-8px) scale(0.96);
  animation: popover-in 0.2s cubic-bezier(0.34, 1.56, 0.64, 1) forwards;
  max-height: 450px;
  overflow: hidden;
}

@keyframes popover-in { 
  to { opacity: 1; transform: translateY(0) scale(1); } 
}

.popover-header {
  padding: 14px 16px;
  border-bottom: 1px solid var(--border);
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: var(--bg-primary);
  border-radius: 16px 16px 0 0;
}

.popover-title {
  font-weight: 600;
  font-size: 13px;
  color: var(--text-primary);
  display: flex;
  align-items: center;
  gap: 6px;
}

.popover-count {
  font-size: 11px;
  color: var(--text-tertiary);
  background: var(--bg-tertiary);
  padding: 2px 8px;
  border-radius: 10px;
}

.popover-close { 
  background: none; 
  border: none; 
  color: var(--text-tertiary); 
  cursor: pointer; 
  padding: 4px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
}

.popover-close:hover { 
  background: var(--bg-hover);
  color: var(--text-primary); 
}

.popover-close svg {
  width: 16px;
  height: 16px;
}

.popover-scroll-area {
  overflow-y: auto;
  max-height: 350px;
  overscroll-behavior: contain;
  scrollbar-width: thin;
  scrollbar-color: var(--bg-tertiary) transparent;
}

.popover-scroll-area::-webkit-scrollbar {
  width: 6px;
}

.popover-scroll-area::-webkit-scrollbar-track {
  background: transparent;
  margin: 4px 0;
}

.popover-scroll-area::-webkit-scrollbar-thumb {
  background: var(--bg-tertiary);
  border-radius: 3px;
}

.popover-scroll-area::-webkit-scrollbar-thumb:hover {
  background: var(--text-tertiary);
}

.comment-item {
  padding: 14px 16px;
  border-bottom: 1px solid var(--border);
  transition: background 0.15s;
}

.comment-item:hover {
  background: var(--bg-hover);
}

.comment-item:last-child {
  border-bottom: none;
}

.comment-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.comment-avatar {
  width: 30px; 
  height: 30px; 
  border-radius: 50%; 
  background: linear-gradient(135deg, var(--accent), var(--accent-hover));
  display: flex; 
  align-items: center; 
  justify-content: center;
  font-size: 11px; 
  font-weight: 600;
  color: white;
  flex-shrink: 0;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.15);
}

.comment-meta {
  flex: 1;
  min-width: 0;
}

.comment-handle { 
  font-size: 13px; 
  font-weight: 600; 
  color: var(--text-primary); 
}

.comment-time {
  font-size: 11px;
  color: var(--text-tertiary);
}

.comment-text { 
  font-size: 13px; 
  line-height: 1.55; 
  color: var(--text-primary);
}

.highlight-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 11px;
  color: var(--text-tertiary);
  background: var(--bg-tertiary);
  padding: 4px 10px;
  border-radius: 12px;
  font-weight: 500;
}

.highlight-badge svg {
  width: 12px;
  height: 12px;
}

.comment-actions {
  display: flex;
  gap: 4px;
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid var(--border);
}

.comment-action-btn {
  background: none; 
  border: none;
  padding: 6px 10px;
  color: var(--text-tertiary); 
  font-size: 12px; 
  font-weight: 500;
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.15s;
  display: flex;
  align-items: center;
  gap: 5px;
}

.comment-action-btn svg {
  width: 14px;
  height: 14px;
}

.comment-action-btn:hover { 
  background: var(--bg-tertiary); 
  color: var(--text-primary); 
}

.btn-add-note {
  color: var(--text-tertiary);
}
.btn-add-note:hover {
  background: var(--bg-tertiary);
  color: var(--text-primary);
}

.add-note-form {
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.add-note-textarea {
  width: 100%;
  min-height: 60px;
  padding: 10px 12px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border);
  border-radius: 10px;
  color: var(--text-primary);
  font-family: "Inter", system-ui, -apple-system, sans-serif;
  font-size: 13px;
  line-height: 1.5;
  resize: vertical;
  outline: none;
  transition: border-color 0.15s;
  box-sizing: border-box;
}

.add-note-textarea:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 2px var(--accent-subtle);
}

.add-note-textarea::placeholder {
  color: var(--text-tertiary);
}

.add-note-actions {
  display: flex;
  justify-content: flex-end;
  gap: 6px;
}

.add-note-actions button {
  background: none;
  border: none;
  padding: 6px 8px;
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  transition: all 0.15s;
}

.add-note-actions button svg {
  width: 16px;
  height: 16px;
}

.add-note-cancel {
  color: var(--text-tertiary);
}
.add-note-cancel:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.add-note-submit {
  color: var(--text-tertiary);
}
.add-note-submit:hover {
  background: var(--bg-tertiary);
  color: var(--text-primary);
}
.add-note-submit:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.inline-compose-modal {
  position: fixed;
  width: 380px;
  max-width: calc(100vw - 32px);
  background: var(--bg-card);
  border: 1px solid var(--border-strong);
  border-radius: 16px;
  padding: 0;
  box-sizing: border-box;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.4), 0 0 80px rgba(59, 130, 246, 0.04);
  z-index: 2147483647;
  pointer-events: auto;
  font-family: "Inter", system-ui, -apple-system, BlinkMacSystemFont, sans-serif;
  color: var(--text-primary);
  animation: modal-in 0.25s cubic-bezier(0.34, 1.56, 0.64, 1) forwards;
  overflow: hidden;
}

@keyframes modal-in {
  from { opacity: 0; transform: scale(0.95) translateY(10px); }
  to { opacity: 1; transform: scale(1) translateY(0); }
}

.inline-compose-modal * {
  box-sizing: border-box;
}

.compose-header {
  padding: 14px 16px;
  border-bottom: 1px solid var(--border);
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: var(--bg-primary);
}

.compose-title {
  font-weight: 600;
  font-size: 14px;
  color: var(--text-primary);
}

.compose-close {
  background: none;
  border: none;
  color: var(--text-tertiary);
  cursor: pointer;
  padding: 4px;
  border-radius: 6px;
  display: flex;
  transition: all 0.15s;
}

.compose-close:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.compose-body {
  padding: 16px;
}

.inline-compose-quote {
  padding: 12px 14px;
  background: var(--accent-subtle);
  border-left: 3px solid var(--accent);
  border-radius: 0 8px 8px 0;
  font-size: 13px;
  color: var(--text-secondary);
  font-style: italic;
  margin-bottom: 14px;
  max-height: 80px;
  overflow: hidden;
  word-break: break-word;
  line-height: 1.5;
}

.inline-compose-textarea {
  width: 100%;
  min-height: 100px;
  padding: 12px 14px;
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: 10px;
  color: var(--text-primary);
  font-family: inherit;
  font-size: 14px;
  line-height: 1.5;
  resize: none;
  box-sizing: border-box;
  transition: border-color 0.15s;
}

.inline-compose-textarea::placeholder {
  color: var(--text-tertiary);
}

.inline-compose-textarea:focus {
  outline: none;
  border-color: var(--accent);
  box-shadow: 0 0 0 2px var(--accent-subtle);
}

.compose-footer {
  padding: 12px 16px;
  border-top: 1px solid var(--border);
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  background: var(--bg-primary);
}

.btn-cancel {
  padding: 9px 18px;
  background: transparent;
  border: 1px solid var(--border);
  border-radius: 8px;
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-cancel:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
  border-color: var(--border);
}

.btn-submit {
  padding: 9px 20px;
  background: var(--accent);
  border: none;
  border-radius: 8px;
  color: white;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-submit:hover {
  background: var(--accent-hover);
  transform: translateY(-1px);
}

.btn-submit:active {
  transform: translateY(0);
}

.btn-submit:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none;
}

.compose-tags-section {
  margin-top: 12px;
  position: relative;
}

.compose-tags-container {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  padding: 8px 10px;
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: 8px;
  min-height: 34px;
  cursor: text;
  transition: border-color 0.15s;
}

.compose-tags-container:focus-within {
  border-color: var(--accent);
  box-shadow: 0 0 0 2px var(--accent-subtle);
}

.compose-tag-pill {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  background: var(--accent-subtle);
  color: var(--accent);
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  white-space: nowrap;
}

.compose-tag-remove {
  background: none;
  border: none;
  color: inherit;
  cursor: pointer;
  padding: 0;
  display: flex;
  align-items: center;
  opacity: 0.7;
  transition: opacity 0.15s;
}

.compose-tag-remove:hover {
  opacity: 1;
}

.compose-tag-remove svg {
  width: 10px;
  height: 10px;
}

.compose-tag-input {
  flex: 1;
  min-width: 60px;
  background: transparent;
  border: none;
  outline: none;
  color: var(--text-primary);
  font-family: inherit;
  font-size: 12px;
  padding: 0;
}

.compose-tag-input::placeholder {
  color: var(--text-tertiary);
}

.compose-tag-suggestions {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  margin-top: 4px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 8px;
  box-shadow: 0 8px 24px rgba(0,0,0,0.3);
  overflow: hidden;
  z-index: 10;
  max-height: 140px;
  overflow-y: auto;
}

.compose-tag-suggestion-item {
  display: block;
  width: 100%;
  text-align: left;
  padding: 8px 12px;
  background: none;
  border: none;
  color: var(--text-secondary);
  font-size: 12px;
  font-family: inherit;
  cursor: pointer;
  transition: background 0.1s;
}

.compose-tag-suggestion-item:hover {
  background: var(--accent-subtle);
  color: var(--accent);
}

.margin-hover-indicator {
  position: fixed;
  display: flex;
  align-items: center;
  pointer-events: none;
  z-index: 2147483647;
  opacity: 0;
  transition: opacity 0.2s ease, transform 0.2s ease;
  transform: scale(0.8) translateX(4px);
}

.margin-hover-indicator.visible {
  opacity: 1;
  transform: scale(1) translateX(0);
}

.margin-toast {
  position: fixed;
  bottom: 24px;
  left: 50%;
  transform: translateX(-50%) translateY(20px);
  padding: 12px 20px;
  background: var(--bg-card);
  border: 1px solid var(--border-strong);
  border-radius: 12px;
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.3);
  font-family: "Inter", system-ui, -apple-system, BlinkMacSystemFont, sans-serif;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
  z-index: 2147483647;
  pointer-events: auto;
  display: flex;
  align-items: center;
  gap: 10px;
  opacity: 0;
  animation: toast-in 0.3s ease forwards;
}

@keyframes toast-in {
  to { opacity: 1; transform: translateX(-50%) translateY(0); }
}

.margin-toast.toast-out {
  animation: toast-out 0.2s ease forwards;
}

@keyframes toast-out {
  to { opacity: 0; transform: translateX(-50%) translateY(10px); }
}

.toast-icon {
  width: 18px;
  height: 18px;
  color: var(--accent);
}

.toast-success .toast-icon {
  color: #34d399;
}
`;
