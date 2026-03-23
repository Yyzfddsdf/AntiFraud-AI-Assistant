package case_library

import "context"

// Service 作为案件库应用服务，对外暴露标准用例接口。
type Service struct{}

func NewService() *Service {
	return &Service{}
}

func DefaultService() *Service {
	return NewService()
}

func (s *Service) CreateHistoricalCase(ctx context.Context, userID string, input CreateHistoricalCaseInput) (HistoricalCaseRecord, error) {
	return CreateHistoricalCase(ctx, userID, input)
}

func (s *Service) ListHistoricalCasePreviews() ([]HistoricalCasePreview, error) {
	return ListHistoricalCasePreviews()
}

func (s *Service) GetHistoricalCaseByID(caseID string) (HistoricalCaseRecord, bool, error) {
	return GetHistoricalCaseByID(caseID)
}

func (s *Service) DeleteHistoricalCaseByID(caseID string) (bool, error) {
	return DeleteHistoricalCaseByID(caseID)
}

func (s *Service) ListScamTypes() []string {
	return ListScamTypes()
}

func (s *Service) ListTargetGroups() []string {
	return ListTargetGroups()
}

func (s *Service) ListPendingReviewPreviews() ([]PendingReviewPreview, error) {
	return ListPendingReviewPreviews()
}

func (s *Service) GetPendingReviewByID(recordID string) (PendingReviewRecord, bool, error) {
	return GetPendingReviewByID(recordID)
}

func (s *Service) ApprovePendingReview(ctx context.Context, recordID string) (HistoricalCaseRecord, error) {
	return ApprovePendingReview(ctx, recordID)
}

func (s *Service) RejectPendingReview(ctx context.Context, recordID string) error {
	return RejectPendingReview(ctx, recordID)
}
