package tool_test

import (
	"strings"
	"testing"

	agenttool "antifraud/multi_agent/tool"
)

func TestCalculateRiskAssessment_HighRiskSignals(t *testing.T) {
	result, err := agenttool.CalculateRiskAssessment(agenttool.RiskAssessmentInput{
		Impersonation:            true,
		Urgency:                  true,
		ThreatPressure:           true,
		BenefitInducement:        false,
		MoneyTransferRequest:     true,
		VerificationCodeRequest:  true,
		RemoteControlRequest:     false,
		LinkOrAppInstallRequest:  true,
		SensitiveInfoRequest:     true,
		PrivateAccountCollection: true,
		FakeOfficialVisuals:      true,
		SimilarCaseStrength:      "strong",
		MultimodalEvidence:       "medium",
		VictimActionStage:        "transferred_money",
		KeyEvidence:              []string{"要求向私人账户转账", "索要短信验证码"},
	})
	if err != nil {
		t.Fatalf("calculate risk assessment failed: %v", err)
	}
	if result.Score < 75 {
		t.Fatalf("expected high score, got %d", result.Score)
	}
	if !strings.Contains(result.StructuredSummary, `"score"`) {
		t.Fatalf("expected structured summary json, got %q", result.StructuredSummary)
	}
}

func TestCalculateRiskAssessment_RejectsInvalidEnums(t *testing.T) {
	_, err := agenttool.CalculateRiskAssessment(agenttool.RiskAssessmentInput{
		SimilarCaseStrength: "extreme",
	})
	if err == nil {
		t.Fatal("expected error for invalid strength")
	}
}
