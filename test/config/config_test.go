package config_test

import (
	"testing"

	"github.com/nexgou/server/src/config"
)

func TestConfigService_Get(t *testing.T) {
	t.Setenv("NEXGOU_TEST_KEY", "hello")
	c := config.NewConfigService()
	if got := c.Get("NEXGOU_TEST_KEY"); got != "hello" {
		t.Errorf("Get: got %q, want %q", got, "hello")
	}
}

func TestConfigService_Get_Missing(t *testing.T) {
	c := config.NewConfigService()
	if got := c.Get("NEXGOU_DEFINITELY_NOT_SET_XYZ"); got != "" {
		t.Errorf("Get missing: got %q, want empty", got)
	}
}

func TestConfigService_GetOrDefault_Set(t *testing.T) {
	t.Setenv("NEXGOU_PORT", "9000")
	c := config.NewConfigService()
	if got := c.GetOrDefault("NEXGOU_PORT", "3000"); got != "9000" {
		t.Errorf("GetOrDefault set: got %q", got)
	}
}

func TestConfigService_GetOrDefault_Missing(t *testing.T) {
	c := config.NewConfigService()
	if got := c.GetOrDefault("NEXGOU_MISSING_VAR_ABC", "default"); got != "default" {
		t.Errorf("GetOrDefault missing: got %q", got)
	}
}

func TestConfigService_GetInt_Valid(t *testing.T) {
	t.Setenv("NEXGOU_INT", "42")
	c := config.NewConfigService()
	if got := c.GetInt("NEXGOU_INT", 0); got != 42 {
		t.Errorf("GetInt: got %d, want 42", got)
	}
}

func TestConfigService_GetInt_Invalid(t *testing.T) {
	t.Setenv("NEXGOU_INT_INVALID", "notanumber")
	c := config.NewConfigService()
	if got := c.GetInt("NEXGOU_INT_INVALID", 99); got != 99 {
		t.Errorf("GetInt invalid: got %d, want fallback 99", got)
	}
}

func TestConfigService_GetInt_Missing(t *testing.T) {
	c := config.NewConfigService()
	if got := c.GetInt("NEXGOU_INT_MISSING_XYZ", 7); got != 7 {
		t.Errorf("GetInt missing: got %d, want 7", got)
	}
}

func TestConfigService_GetBool_True(t *testing.T) {
	for _, val := range []string{"1", "true", "yes"} {
		t.Setenv("NEXGOU_BOOL", val)
		c := config.NewConfigService()
		if !c.GetBool("NEXGOU_BOOL", false) {
			t.Errorf("GetBool %q: expected true", val)
		}
	}
}

func TestConfigService_GetBool_False(t *testing.T) {
	for _, val := range []string{"0", "false", "no"} {
		t.Setenv("NEXGOU_BOOL", val)
		c := config.NewConfigService()
		if c.GetBool("NEXGOU_BOOL", true) {
			t.Errorf("GetBool %q: expected false", val)
		}
	}
}

func TestConfigService_GetBool_Fallback(t *testing.T) {
	c := config.NewConfigService()
	if !c.GetBool("NEXGOU_BOOL_FALLBACK_XYZ", true) {
		t.Error("GetBool fallback: expected true")
	}
}

func TestConfigService_MustGet_Set(t *testing.T) {
	t.Setenv("NEXGOU_REQUIRED", "present")
	c := config.NewConfigService()
	if got := c.MustGet("NEXGOU_REQUIRED"); got != "present" {
		t.Errorf("MustGet: got %q", got)
	}
}

func TestConfigService_MustGet_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustGet: expected panic for missing env var")
		}
	}()
	c := config.NewConfigService()
	c.MustGet("NEXGOU_DEFINITELY_MISSING_XYZ_999")
}
