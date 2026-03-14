package case_library

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"antifraud/database"
	model "antifraud/multi_agent/case_library/model"
)

type PendingReviewRecord = model.PendingReviewRecord
type PendingReviewPreview = model.PendingReviewPreview
type pendingReviewEntity = model.PendingReviewEntity

// CreatePendingReview 将案件写入待审核表。
func CreatePendingReview(ctx context.Context, userID string, input CreateHistoricalCaseInput) (PendingReviewRecord, error) {
	prepared, err := prepareHistoricalCaseInput(ctx, input)
	if err != nil {
		return PendingReviewRecord{}, err
	}
	if duplicateErr := detectDuplicateHistoricalCase(prepared.vector); duplicateErr != nil {
		return PendingReviewRecord{}, duplicateErr
	}

	entity := pendingReviewEntity{
		RecordID:           newPendingReviewID(),
		UserID:             normalizeUserID(userID),
		Title:              prepared.normalizedInput.Title,
		TargetGroup:        prepared.normalizedInput.TargetGroup,
		RiskLevel:          prepared.normalizedInput.RiskLevel,
		ScamType:           prepared.normalizedInput.ScamType,
		CaseDescription:    prepared.normalizedInput.CaseDescription,
		TypicalScripts:     encodeStringList(prepared.normalizedInput.TypicalScripts),
		Keywords:           encodeStringList(prepared.normalizedInput.Keywords),
		ViolatedLaw:        prepared.normalizedInput.ViolatedLaw,
		Suggestion:         prepared.normalizedInput.Suggestion,
		EmbeddingVector:    encodeFloatList(prepared.vector),
		EmbeddingModel:     strings.TrimSpace(prepared.modelName),
		EmbeddingDimension: len(prepared.vector),
	}

	db, err := database.GetHistoricalCaseDB()
	if err != nil {
		return PendingReviewRecord{}, err
	}
	if err := db.Create(&entity).Error; err != nil {
		return PendingReviewRecord{}, fmt.Errorf("insert pending review case failed: %w", err)
	}
	return pendingReviewRecordFromEntity(entity), nil
}

// APPEND_MARKER

// ListPendingReviewPreviews 返回所有待审核案件预览。
func ListPendingReviewPreviews() ([]PendingReviewPreview, error) {
	db, err := database.GetHistoricalCaseDB()
	if err != nil {
		return nil, err
	}

	var rows []pendingReviewEntity
	if err := db.Select("record_id", "title", "target_group", "risk_level", "scam_type", "violated_law", "created_at").
		Order("created_at desc").
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("query pending review previews failed: %w", err)
	}

	previews := make([]PendingReviewPreview, 0, len(rows))
	for _, row := range rows {
		normalizedRiskLevel := normalizeRiskLevel(row.RiskLevel)
		if normalizedRiskLevel == "" {
			normalizedRiskLevel = strings.TrimSpace(row.RiskLevel)
		}
		previews = append(previews, PendingReviewPreview{
			RecordID:    strings.TrimSpace(row.RecordID),
			Title:       strings.TrimSpace(row.Title),
			TargetGroup: strings.TrimSpace(row.TargetGroup),
			RiskLevel:   normalizedRiskLevel,
			ScamType:    strings.TrimSpace(row.ScamType),
			ViolatedLaw: strings.TrimSpace(row.ViolatedLaw),
			CreatedAt:   row.CreatedAt,
		})
	}
	return previews, nil
}

// GetPendingReviewByID 根据 record_id 返回完整待审核案件详情。
func GetPendingReviewByID(recordID string) (PendingReviewRecord, bool, error) {
	trimmed := strings.TrimSpace(recordID)
	if trimmed == "" {
		return PendingReviewRecord{}, false, nil
	}

	db, err := database.GetHistoricalCaseDB()
	if err != nil {
		return PendingReviewRecord{}, false, err
	}

	var entity pendingReviewEntity
	query := db.Where("record_id = ?", trimmed).Limit(1).Find(&entity)
	if query.Error != nil {
		return PendingReviewRecord{}, false, fmt.Errorf("query pending review detail failed: %w", query.Error)
	}
	if query.RowsAffected == 0 {
		return PendingReviewRecord{}, false, nil
	}
	return pendingReviewRecordFromEntity(entity), true, nil
}

// APPEND_MARKER_2

// ApprovePendingReview 审核通过：读取待审核记录 → 直接写入历史案件库 → 删除待审核记录。
func ApprovePendingReview(ctx context.Context, recordID string) (HistoricalCaseRecord, error) {
	trimmed := strings.TrimSpace(recordID)
	if trimmed == "" {
		return HistoricalCaseRecord{}, fmt.Errorf("recordID is required")
	}

	db, err := database.GetHistoricalCaseDB()
	if err != nil {
		return HistoricalCaseRecord{}, err
	}

	var entity pendingReviewEntity
	query := db.Where("record_id = ?", trimmed).Limit(1).Find(&entity)
	if query.Error != nil {
		return HistoricalCaseRecord{}, fmt.Errorf("query pending review failed: %w", query.Error)
	}
	if query.RowsAffected == 0 {
		return HistoricalCaseRecord{}, fmt.Errorf("pending review case not found or already processed")
	}

	vector := decodeFloatList(entity.EmbeddingVector)
	if len(vector) == 0 {
		return HistoricalCaseRecord{}, fmt.Errorf("pending review record missing embedding vector")
	}
	if entity.EmbeddingDimension > 0 && len(vector) != entity.EmbeddingDimension {
		return HistoricalCaseRecord{}, fmt.Errorf("pending review embedding dimension mismatch")
	}

	record, createErr := insertHistoricalCasePrepared(entity.UserID, preparedHistoricalCaseInput{
		normalizedInput: CreateHistoricalCaseInput{
			Title:           strings.TrimSpace(entity.Title),
			TargetGroup:     strings.TrimSpace(entity.TargetGroup),
			RiskLevel:       strings.TrimSpace(entity.RiskLevel),
			ScamType:        strings.TrimSpace(entity.ScamType),
			CaseDescription: strings.TrimSpace(entity.CaseDescription),
			TypicalScripts:  decodeStringList(entity.TypicalScripts),
			Keywords:        decodeStringList(entity.Keywords),
			ViolatedLaw:     strings.TrimSpace(entity.ViolatedLaw),
			Suggestion:      strings.TrimSpace(entity.Suggestion),
		},
		vector:    vector,
		modelName: strings.TrimSpace(entity.EmbeddingModel),
	})
	if createErr != nil {
		return HistoricalCaseRecord{}, fmt.Errorf("approve and create historical case failed: %w", createErr)
	}

	if err := db.Where("record_id = ?", trimmed).Delete(&pendingReviewEntity{}).Error; err != nil {
		return record, fmt.Errorf("delete pending review record failed (case already created): %w", err)
	}

	return record, nil
}

func newPendingReviewID() string {
	buffer := make([]byte, 6)
	if _, err := rand.Read(buffer); err != nil {
		return fmt.Sprintf("PREV-%d", time.Now().UnixNano())
	}
	return "PREV-" + strings.ToUpper(hex.EncodeToString(buffer))
}

func pendingReviewRecordFromEntity(entity pendingReviewEntity) PendingReviewRecord {
	normalizedRiskLevel := normalizeRiskLevel(entity.RiskLevel)
	if normalizedRiskLevel == "" {
		normalizedRiskLevel = strings.TrimSpace(entity.RiskLevel)
	}

	return PendingReviewRecord{
		RecordID:        strings.TrimSpace(entity.RecordID),
		UserID:          strings.TrimSpace(entity.UserID),
		Title:           strings.TrimSpace(entity.Title),
		TargetGroup:     strings.TrimSpace(entity.TargetGroup),
		RiskLevel:       normalizedRiskLevel,
		ScamType:        strings.TrimSpace(entity.ScamType),
		CaseDescription: strings.TrimSpace(entity.CaseDescription),
		TypicalScripts:  decodeStringList(entity.TypicalScripts),
		Keywords:        decodeStringList(entity.Keywords),
		ViolatedLaw:     strings.TrimSpace(entity.ViolatedLaw),
		Suggestion:      strings.TrimSpace(entity.Suggestion),
		CreatedAt:       entity.CreatedAt,
		UpdatedAt:       entity.UpdatedAt,
	}
}
