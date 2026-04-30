package i18n

import (
	"os"
	"testing"
)

func TestI18n(t *testing.T) {
	// Test Default (zh)
	os.Setenv("LANG", "zh_CN.UTF-8")
	Init()
	if T("test_key") != "测试" {
		t.Errorf("Expected '测试', got '%s'", T("test_key"))
	}

	// Test English
	os.Setenv("LANG", "en_US.UTF-8")
	Init()
	if T("test_key") != "Test" {
		t.Errorf("Expected 'Test', got '%s'", T("test_key"))
	}
}

func TestTForTranslatesWithoutChangingCurrentLanguage(t *testing.T) {
	os.Setenv("LANG", "zh_CN.UTF-8")
	Init()

	if got := TFor("en", "test_key"); got != "Test" {
		t.Fatalf("expected explicit English translation, got %q", got)
	}
	if got := T("test_key"); got != "测试" {
		t.Fatalf("expected current language to remain Chinese, got %q", got)
	}
}
