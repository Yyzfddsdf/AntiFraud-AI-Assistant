package tool_test

import (
	"strings"
	"testing"

	agenttool "antifraud/internal/modules/multi_agent/adapters/outbound/tool"
)

func TestCalculateRiskAssessment_HighRiskSignals(t *testing.T) {
	result, err := agenttool.CalculateRiskAssessment(agenttool.RiskAssessmentInput{
		Impersonation:              true,
		Urgency:                    true,
		ThreatPressure:             true,
		BenefitInducement:          false,
		ChannelSwitchRequest:       true,
		InviteCodeRequest:          true,
		VerificationProcessRequest: true,
		ScreenshotOrRecordRequest:  true,
		TrustBuildingPressure:      true,
		MoneyTransferRequest:       true,
		VerificationCodeRequest:    true,
		RemoteControlRequest:       false,
		LinkOrAppInstallRequest:    true,
		SensitiveInfoRequest:       true,
		PrivateAccountCollection:   true,
		FakeOfficialVisuals:        true,
		SimilarCaseStrength:        "strong",
		MultimodalEvidence:         "medium",
		VictimActionStage:          "transferred_money",
		KeyEvidence:                []string{"要求向私人账户转账", "索要短信验证码"},
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

func TestCalculateRiskAssessment_PreliminarySignals(t *testing.T) {
	result, err := agenttool.CalculateRiskAssessment(agenttool.RiskAssessmentInput{
		Impersonation:              true,
		ChannelSwitchRequest:       true,
		InviteCodeRequest:          true,
		VerificationProcessRequest: true,
		ScreenshotOrRecordRequest:  true,
		TrustBuildingPressure:      true,
		FakeOfficialVisuals:        true,
	})
	if err != nil {
		t.Fatalf("calculate risk assessment failed: %v", err)
	}
	if result.Score < 25 {
		t.Fatalf("expected preliminary signals to produce a non-trivial score, got %d", result.Score)
	}
	if result.Score >= 75 {
		t.Fatalf("expected preliminary signals to stay below terminal high-risk range, got %d", result.Score)
	}

	expectedRules := []string{
		"冒充身份",
		"引导切换线路/平台",
		"索要邀请码/口令",
		"前置认证/额度验证",
		"要求截图/录屏",
		"逐步建立信任",
		"出现仿冒官方视觉证据",
	}
	for _, rule := range expectedRules {
		if !strings.Contains(result.StructuredSummary, rule) {
			t.Fatalf("expected structured summary to contain %q, got %q", rule, result.StructuredSummary)
		}
	}
	for _, rule := range []string{"要求转账/充值", "要求验证码", "要求远程控制", "要求向私人账户收款"} {
		if strings.Contains(result.StructuredSummary, rule) {
			t.Fatalf("did not expect terminal rule %q in preliminary-only result: %q", rule, result.StructuredSummary)
		}
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
