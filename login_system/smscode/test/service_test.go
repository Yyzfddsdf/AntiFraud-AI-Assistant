package smscode_test

import (
	"context"
	"errors"
	"testing"

	"antifraud/login_system/smscode"
)

func TestNormalizePhone(t *testing.T) {
	phone, err := smscode.NormalizePhone("138-0013 8000")
	if err != nil {
		t.Fatalf("expected normalized phone, got error: %v", err)
	}
	if phone != "13800138000" {
		t.Fatalf("unexpected phone: %s", phone)
	}
}

func TestNormalizePhoneRejectsInvalidInput(t *testing.T) {
	_, err := smscode.NormalizePhone("abc")
	if !errors.Is(err, smscode.ErrInvalidPhoneFormat) {
		t.Fatalf("expected ErrInvalidPhoneFormat, got: %v", err)
	}
}

func TestDemoServiceSendCode(t *testing.T) {
	service := smscode.NewDemoService()
	if err := service.SendCode(context.Background(), "13800138000"); err != nil {
		t.Fatalf("expected send code success, got: %v", err)
	}
}

func TestDemoServiceVerifyCode(t *testing.T) {
	service := smscode.NewDemoService()
	if err := service.VerifyCode(context.Background(), "13800138000", smscode.DemoCode); err != nil {
		t.Fatalf("expected demo code success, got: %v", err)
	}
	if err := service.VerifyCode(context.Background(), "13800138000", "111111"); !errors.Is(err, smscode.ErrInvalidSMSCode) {
		t.Fatalf("expected ErrInvalidSMSCode, got: %v", err)
	}
}
