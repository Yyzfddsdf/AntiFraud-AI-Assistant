package case_library

import (
	"fmt"
	"strings"

	"antifraud/database"
)

func insertHistoricalCasePrepared(userID string, prepared preparedHistoricalCaseInput) (HistoricalCaseRecord, error) {
	entity := historicalCaseEntity{
		CaseID:             newHistoricalCaseID(),
		CreatedBy:          normalizeUserID(userID),
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
		return HistoricalCaseRecord{}, err
	}
	if err := db.Create(&entity).Error; err != nil {
		return HistoricalCaseRecord{}, fmt.Errorf("insert historical case failed: %w", err)
	}

	record := recordFromEntity(entity)
	upsertHistoricalCaseVectorCache(record)
	touchHistoricalCaseGraphCacheVersion()
	return record, nil
}
