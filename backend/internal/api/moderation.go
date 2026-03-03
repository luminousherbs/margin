package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"margin.at/internal/config"
	"margin.at/internal/db"
	"margin.at/internal/logger"
)

type ModerationHandler struct {
	db        *db.DB
	refresher *TokenRefresher
}

func NewModerationHandler(database *db.DB, refresher *TokenRefresher) *ModerationHandler {
	return &ModerationHandler{db: database, refresher: refresher}
}

func (m *ModerationHandler) BlockUser(w http.ResponseWriter, r *http.Request) {
	session, err := m.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		DID string `json:"did"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DID == "" {
		http.Error(w, "did is required", http.StatusBadRequest)
		return
	}

	if req.DID == session.DID {
		http.Error(w, "Cannot block yourself", http.StatusBadRequest)
		return
	}

	if err := m.db.CreateBlock(session.DID, req.DID); err != nil {
		logger.Error("Failed to create block: %v", err)
		http.Error(w, "Failed to block user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (m *ModerationHandler) UnblockUser(w http.ResponseWriter, r *http.Request) {
	session, err := m.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	did := r.URL.Query().Get("did")
	if did == "" {
		http.Error(w, "did query parameter required", http.StatusBadRequest)
		return
	}

	if err := m.db.DeleteBlock(session.DID, did); err != nil {
		logger.Error("Failed to delete block: %v", err)
		http.Error(w, "Failed to unblock user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (m *ModerationHandler) GetBlocks(w http.ResponseWriter, r *http.Request) {
	session, err := m.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	blocks, err := m.db.GetBlocks(session.DID)
	if err != nil {
		http.Error(w, "Failed to fetch blocks", http.StatusInternalServerError)
		return
	}

	dids := make([]string, len(blocks))
	for i, b := range blocks {
		dids[i] = b.SubjectDID
	}
	profiles := fetchProfilesForDIDs(m.db, dids)

	type BlockedUser struct {
		DID       string `json:"did"`
		Author    Author `json:"author"`
		CreatedAt string `json:"createdAt"`
	}

	items := make([]BlockedUser, len(blocks))
	for i, b := range blocks {
		items[i] = BlockedUser{
			DID:       b.SubjectDID,
			Author:    profiles[b.SubjectDID],
			CreatedAt: b.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"items": items})
}

func (m *ModerationHandler) MuteUser(w http.ResponseWriter, r *http.Request) {
	session, err := m.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		DID string `json:"did"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DID == "" {
		http.Error(w, "did is required", http.StatusBadRequest)
		return
	}

	if req.DID == session.DID {
		http.Error(w, "Cannot mute yourself", http.StatusBadRequest)
		return
	}

	if err := m.db.CreateMute(session.DID, req.DID); err != nil {
		logger.Error("Failed to create mute: %v", err)
		http.Error(w, "Failed to mute user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (m *ModerationHandler) UnmuteUser(w http.ResponseWriter, r *http.Request) {
	session, err := m.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	did := r.URL.Query().Get("did")
	if did == "" {
		http.Error(w, "did query parameter required", http.StatusBadRequest)
		return
	}

	if err := m.db.DeleteMute(session.DID, did); err != nil {
		logger.Error("Failed to delete mute: %v", err)
		http.Error(w, "Failed to unmute user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (m *ModerationHandler) GetMutes(w http.ResponseWriter, r *http.Request) {
	session, err := m.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	mutes, err := m.db.GetMutes(session.DID)
	if err != nil {
		http.Error(w, "Failed to fetch mutes", http.StatusInternalServerError)
		return
	}

	dids := make([]string, len(mutes))
	for i, mu := range mutes {
		dids[i] = mu.SubjectDID
	}
	profiles := fetchProfilesForDIDs(m.db, dids)

	type MutedUser struct {
		DID       string `json:"did"`
		Author    Author `json:"author"`
		CreatedAt string `json:"createdAt"`
	}

	items := make([]MutedUser, len(mutes))
	for i, mu := range mutes {
		items[i] = MutedUser{
			DID:       mu.SubjectDID,
			Author:    profiles[mu.SubjectDID],
			CreatedAt: mu.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"items": items})
}

func (m *ModerationHandler) GetRelationship(w http.ResponseWriter, r *http.Request) {
	viewerDID := m.getViewerDID(r)
	subjectDID := r.URL.Query().Get("did")

	if subjectDID == "" {
		http.Error(w, "did query parameter required", http.StatusBadRequest)
		return
	}

	blocked, muted, blockedBy, err := m.db.GetViewerRelationship(viewerDID, subjectDID)
	if err != nil {
		http.Error(w, "Failed to get relationship", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"blocking":  blocked,
		"muting":    muted,
		"blockedBy": blockedBy,
	})
}

func (m *ModerationHandler) CreateReport(w http.ResponseWriter, r *http.Request) {
	session, err := m.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		SubjectDID string  `json:"subjectDid"`
		SubjectURI *string `json:"subjectUri,omitempty"`
		ReasonType string  `json:"reasonType"`
		ReasonText *string `json:"reasonText,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SubjectDID == "" || req.ReasonType == "" {
		http.Error(w, "subjectDid and reasonType are required", http.StatusBadRequest)
		return
	}

	validReasons := map[string]bool{
		"spam":       true,
		"violation":  true,
		"misleading": true,
		"sexual":     true,
		"rude":       true,
		"other":      true,
	}

	if !validReasons[req.ReasonType] {
		http.Error(w, "Invalid reasonType", http.StatusBadRequest)
		return
	}

	id, err := m.db.CreateReport(session.DID, req.SubjectDID, req.SubjectURI, req.ReasonType, req.ReasonText)
	if err != nil {
		logger.Error("Failed to create report: %v", err)
		http.Error(w, "Failed to submit report", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "status": "ok"})
}

func (m *ModerationHandler) AdminGetReports(w http.ResponseWriter, r *http.Request) {
	session, err := m.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !config.Get().IsAdmin(session.DID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	status := r.URL.Query().Get("status")
	limit := parseIntParam(r, "limit", 50)
	offset := parseIntParam(r, "offset", 0)

	reports, err := m.db.GetReports(status, limit, offset)
	if err != nil {
		http.Error(w, "Failed to fetch reports", http.StatusInternalServerError)
		return
	}

	uniqueDIDs := make(map[string]bool)
	for _, rpt := range reports {
		uniqueDIDs[rpt.ReporterDID] = true
		uniqueDIDs[rpt.SubjectDID] = true
	}
	dids := make([]string, 0, len(uniqueDIDs))
	for did := range uniqueDIDs {
		dids = append(dids, did)
	}
	profiles := fetchProfilesForDIDs(m.db, dids)

	type HydratedReport struct {
		ID         int     `json:"id"`
		Reporter   Author  `json:"reporter"`
		Subject    Author  `json:"subject"`
		SubjectURI *string `json:"subjectUri,omitempty"`
		ReasonType string  `json:"reasonType"`
		ReasonText *string `json:"reasonText,omitempty"`
		Status     string  `json:"status"`
		CreatedAt  string  `json:"createdAt"`
		ResolvedAt *string `json:"resolvedAt,omitempty"`
		ResolvedBy *string `json:"resolvedBy,omitempty"`
	}

	items := make([]HydratedReport, len(reports))
	for i, rpt := range reports {
		items[i] = HydratedReport{
			ID:         rpt.ID,
			Reporter:   profiles[rpt.ReporterDID],
			Subject:    profiles[rpt.SubjectDID],
			SubjectURI: rpt.SubjectURI,
			ReasonType: rpt.ReasonType,
			ReasonText: rpt.ReasonText,
			Status:     rpt.Status,
			CreatedAt:  rpt.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if rpt.ResolvedAt != nil {
			resolved := rpt.ResolvedAt.Format("2006-01-02T15:04:05Z")
			items[i].ResolvedAt = &resolved
		}
		items[i].ResolvedBy = rpt.ResolvedBy
	}

	pendingCount, _ := m.db.GetReportCount("pending")
	totalCount, _ := m.db.GetReportCount("")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"items":        items,
		"totalItems":   totalCount,
		"pendingCount": pendingCount,
	})
}

func (m *ModerationHandler) AdminTakeAction(w http.ResponseWriter, r *http.Request) {
	session, err := m.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !config.Get().IsAdmin(session.DID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req struct {
		ReportID int     `json:"reportId"`
		Action   string  `json:"action"`
		Comment  *string `json:"comment,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	validActions := map[string]bool{
		"acknowledge": true,
		"escalate":    true,
		"takedown":    true,
		"dismiss":     true,
	}

	if !validActions[req.Action] {
		http.Error(w, "Invalid action", http.StatusBadRequest)
		return
	}

	report, err := m.db.GetReport(req.ReportID)
	if err != nil {
		http.Error(w, "Report not found", http.StatusNotFound)
		return
	}

	if err := m.db.CreateModerationAction(req.ReportID, session.DID, req.Action, req.Comment); err != nil {
		logger.Error("Failed to create moderation action: %v", err)
		http.Error(w, "Failed to take action", http.StatusInternalServerError)
		return
	}

	resolveStatus := "resolved"
	switch req.Action {
	case "dismiss":
		resolveStatus = "dismissed"
	case "escalate":
		resolveStatus = "escalated"
	case "takedown":
		resolveStatus = "resolved"
		if report.SubjectURI != nil && *report.SubjectURI != "" {
			m.deleteContent(*report.SubjectURI)
		}
	case "acknowledge":
		resolveStatus = "acknowledged"
	}

	if err := m.db.ResolveReport(req.ReportID, session.DID, resolveStatus); err != nil {
		logger.Error("Failed to resolve report: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (m *ModerationHandler) AdminGetReport(w http.ResponseWriter, r *http.Request) {
	session, err := m.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !config.Get().IsAdmin(session.DID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid report ID", http.StatusBadRequest)
		return
	}

	report, err := m.db.GetReport(id)
	if err != nil {
		http.Error(w, "Report not found", http.StatusNotFound)
		return
	}

	actions, _ := m.db.GetReportActions(id)

	profiles := fetchProfilesForDIDs(m.db, []string{report.ReporterDID, report.SubjectDID})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"report":   report,
		"reporter": profiles[report.ReporterDID],
		"subject":  profiles[report.SubjectDID],
		"actions":  actions,
	})
}

func (m *ModerationHandler) AdminCheckAccess(w http.ResponseWriter, r *http.Request) {
	session, err := m.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"isAdmin": false})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"isAdmin": config.Get().IsAdmin(session.DID)})
}

func (m *ModerationHandler) deleteContent(uri string) {
	m.db.Exec("DELETE FROM annotations WHERE uri = $1", uri)
	m.db.Exec("DELETE FROM highlights WHERE uri = $1", uri)
	m.db.Exec("DELETE FROM bookmarks WHERE uri = $1", uri)
	m.db.Exec("DELETE FROM replies WHERE uri = $1", uri)
}

func (m *ModerationHandler) AdminCreateLabel(w http.ResponseWriter, r *http.Request) {
	session, err := m.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !config.Get().IsAdmin(session.DID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req struct {
		Src string `json:"src"`
		URI string `json:"uri"`
		Val string `json:"val"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Val == "" {
		http.Error(w, "val is required", http.StatusBadRequest)
		return
	}

	labelerDID := config.Get().ServiceDID
	if labelerDID == "" {
		http.Error(w, "SERVICE_DID not configured — cannot issue labels", http.StatusInternalServerError)
		return
	}

	targetURI := req.URI
	if targetURI == "" {
		targetURI = req.Src
	}
	if targetURI == "" {
		http.Error(w, "src or uri is required", http.StatusBadRequest)
		return
	}

	validLabels := map[string]bool{
		"sexual":     true,
		"nudity":     true,
		"violence":   true,
		"gore":       true,
		"spam":       true,
		"misleading": true,
	}

	if !validLabels[req.Val] {
		http.Error(w, "Invalid label value. Must be one of: sexual, nudity, violence, gore, spam, misleading", http.StatusBadRequest)
		return
	}

	if err := m.db.CreateContentLabel(labelerDID, targetURI, req.Val, session.DID); err != nil {
		logger.Error("Failed to create content label: %v", err)
		http.Error(w, "Failed to create label", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (m *ModerationHandler) AdminDeleteLabel(w http.ResponseWriter, r *http.Request) {
	session, err := m.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !config.Get().IsAdmin(session.DID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid label ID", http.StatusBadRequest)
		return
	}

	if err := m.db.DeleteContentLabel(id); err != nil {
		http.Error(w, "Failed to delete label", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (m *ModerationHandler) AdminGetLabels(w http.ResponseWriter, r *http.Request) {
	session, err := m.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !config.Get().IsAdmin(session.DID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	limit := parseIntParam(r, "limit", 50)
	offset := parseIntParam(r, "offset", 0)

	labels, err := m.db.GetAllContentLabels(limit, offset)
	if err != nil {
		http.Error(w, "Failed to fetch labels", http.StatusInternalServerError)
		return
	}

	uniqueDIDs := make(map[string]bool)
	for _, l := range labels {
		uniqueDIDs[l.CreatedBy] = true
		if len(l.Src) > 4 && l.Src[:4] == "did:" {
			uniqueDIDs[l.Src] = true
		}
	}
	dids := make([]string, 0, len(uniqueDIDs))
	for did := range uniqueDIDs {
		dids = append(dids, did)
	}
	profiles := fetchProfilesForDIDs(m.db, dids)

	type HydratedLabel struct {
		ID        int     `json:"id"`
		Src       string  `json:"src"`
		URI       string  `json:"uri"`
		Val       string  `json:"val"`
		CreatedBy Author  `json:"createdBy"`
		CreatedAt string  `json:"createdAt"`
		Subject   *Author `json:"subject,omitempty"`
	}

	items := make([]HydratedLabel, len(labels))
	for i, l := range labels {
		items[i] = HydratedLabel{
			ID:        l.ID,
			Src:       l.Src,
			URI:       l.URI,
			Val:       l.Val,
			CreatedBy: profiles[l.CreatedBy],
			CreatedAt: l.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if len(l.Src) > 4 && l.Src[:4] == "did:" {
			subj := profiles[l.Src]
			items[i].Subject = &subj
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"items": items})
}

func (m *ModerationHandler) getViewerDID(r *http.Request) string {
	cookie, err := r.Cookie("margin_session")
	if err != nil {
		return ""
	}
	did, _, _, _, _, err := m.db.GetSession(cookie.Value)
	if err != nil {
		return ""
	}
	return did
}

func (m *ModerationHandler) GetLabelerInfo(w http.ResponseWriter, r *http.Request) {
	serviceDID := config.Get().ServiceDID

	type LabelDefinition struct {
		Identifier  string `json:"identifier"`
		Severity    string `json:"severity"`
		Blurs       string `json:"blurs"`
		Description string `json:"description"`
	}

	labels := []LabelDefinition{
		{Identifier: "sexual", Severity: "inform", Blurs: "content", Description: "Sexual content"},
		{Identifier: "nudity", Severity: "inform", Blurs: "content", Description: "Nudity"},
		{Identifier: "violence", Severity: "inform", Blurs: "content", Description: "Violence"},
		{Identifier: "gore", Severity: "alert", Blurs: "content", Description: "Graphic/gory content"},
		{Identifier: "spam", Severity: "inform", Blurs: "content", Description: "Spam or unwanted content"},
		{Identifier: "misleading", Severity: "inform", Blurs: "content", Description: "Misleading information"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"did":    serviceDID,
		"name":   "Margin Moderation",
		"labels": labels,
	})
}
