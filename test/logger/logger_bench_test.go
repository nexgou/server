package logger_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/bytedance/sonic"
)

func BenchmarkLoggerJSONMarshalComparisons(b *testing.B) {
	entry := map[string]any{
		"ts":      time.Unix(0, 0).UTC().Format(time.RFC3339),
		"level":   "info",
		"context": "HTTP",
		"msg":     "request",
		"path":    "/health",
		"status":  200,
	}

	b.Run("stdlib", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for index := 0; index < b.N; index++ {
			if _, err := json.Marshal(entry); err != nil {
				b.Fatalf("stdlib marshal returned error: %v", err)
			}
		}
	})

	b.Run("sonic_std", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for index := 0; index < b.N; index++ {
			if _, err := sonic.ConfigStd.Marshal(entry); err != nil {
				b.Fatalf("sonic std marshal returned error: %v", err)
			}
		}
	})

	b.Run("sonic_default", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for index := 0; index < b.N; index++ {
			if _, err := sonic.ConfigDefault.Marshal(entry); err != nil {
				b.Fatalf("sonic default marshal returned error: %v", err)
			}
		}
	})
}
