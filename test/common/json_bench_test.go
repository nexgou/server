package common_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bytedance/sonic"
	"github.com/nexgou/server/src/common"
)

type benchmarkUserPayload struct {
	ID        int                     `json:"id"`
	Name      string                  `json:"name"`
	Email     string                  `json:"email"`
	Age       int                     `json:"age"`
	Active    bool                    `json:"active"`
	Roles     []string                `json:"roles"`
	Metadata  map[string]any          `json:"metadata"`
	CreatedAt string                  `json:"createdAt"`
	Profile   benchmarkProfilePayload `json:"profile"`
}

type benchmarkProfilePayload struct {
	Company string   `json:"company"`
	Tags    []string `json:"tags"`
}

type discardResponseWriter struct {
	header http.Header
	status int
}

func newDiscardResponseWriter() *discardResponseWriter {
	return &discardResponseWriter{header: make(http.Header)}
}

func (writer *discardResponseWriter) Header() http.Header {
	return writer.header
}

func (writer *discardResponseWriter) Write(data []byte) (int, error) {
	return io.Discard.Write(data)
}

func (writer *discardResponseWriter) WriteHeader(status int) {
	writer.status = status
}

func BenchmarkContextJSONStdlib(b *testing.B) {
	ctx := common.NewContext(newDiscardResponseWriter(), httptest.NewRequest(http.MethodGet, "/users/1", nil), nil)
	payload := newBenchmarkUserPayload()

	b.ReportAllocs()
	b.ResetTimer()

	for index := 0; index < b.N; index++ {
		if err := ctx.JSON(http.StatusOK, payload); err != nil {
			b.Fatalf("Context.JSON returned error: %v", err)
		}
	}
}

func BenchmarkContextBodyStdlib(b *testing.B) {
	encoded := benchmarkUserPayloadJSON(b)
	request := httptest.NewRequest(http.MethodPost, "/users", nil)
	ctx := common.NewContext(httptest.NewRecorder(), request, nil)

	b.ReportAllocs()
	b.ResetTimer()

	for index := 0; index < b.N; index++ {
		request.Body = io.NopCloser(bytes.NewReader(encoded))
		var payload benchmarkUserPayload
		if err := ctx.Body(&payload); err != nil {
			b.Fatalf("Context.Body returned error: %v", err)
		}
	}
}

func BenchmarkJSONEncodeStdlib(b *testing.B) {
	benchmarkJSONEncode(b, "stdlib", func(payload benchmarkUserPayload) error {
		return json.NewEncoder(io.Discard).Encode(payload)
	})
}

func BenchmarkJSONEncodeSonicStd(b *testing.B) {
	benchmarkJSONEncode(b, "sonic_std", func(payload benchmarkUserPayload) error {
		return sonic.ConfigStd.NewEncoder(io.Discard).Encode(payload)
	})
}

func BenchmarkJSONEncodeSonicDefault(b *testing.B) {
	benchmarkJSONEncode(b, "sonic_default", func(payload benchmarkUserPayload) error {
		return sonic.ConfigDefault.NewEncoder(io.Discard).Encode(payload)
	})
}

func BenchmarkJSONDecodeStdlib(b *testing.B) {
	encoded := benchmarkUserPayloadJSON(b)
	benchmarkJSONDecode(b, "stdlib", encoded, func(reader io.Reader, payload *benchmarkUserPayload) error {
		return json.NewDecoder(reader).Decode(payload)
	})
}

func BenchmarkJSONDecodeSonicStd(b *testing.B) {
	encoded := benchmarkUserPayloadJSON(b)
	benchmarkJSONDecode(b, "sonic_std", encoded, func(reader io.Reader, payload *benchmarkUserPayload) error {
		return sonic.ConfigStd.NewDecoder(reader).Decode(payload)
	})
}

func BenchmarkJSONDecodeSonicDefault(b *testing.B) {
	encoded := benchmarkUserPayloadJSON(b)
	benchmarkJSONDecode(b, "sonic_default", encoded, func(reader io.Reader, payload *benchmarkUserPayload) error {
		return sonic.ConfigDefault.NewDecoder(reader).Decode(payload)
	})
}

func benchmarkJSONEncode(b *testing.B, name string, encode func(benchmarkUserPayload) error) {
	b.Helper()
	b.Run(name, func(b *testing.B) {
		payload := newBenchmarkUserPayload()
		b.ReportAllocs()
		b.ResetTimer()

		for index := 0; index < b.N; index++ {
			if err := encode(payload); err != nil {
				b.Fatalf("encode returned error: %v", err)
			}
		}
	})
}

func benchmarkJSONDecode(b *testing.B, name string, encoded []byte, decode func(io.Reader, *benchmarkUserPayload) error) {
	b.Helper()
	b.Run(name, func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for index := 0; index < b.N; index++ {
			var payload benchmarkUserPayload
			if err := decode(bytes.NewReader(encoded), &payload); err != nil {
				b.Fatalf("decode returned error: %v", err)
			}
		}
	})
}

func benchmarkUserPayloadJSON(b *testing.B) []byte {
	b.Helper()
	encoded, err := json.Marshal(newBenchmarkUserPayload())
	if err != nil {
		b.Fatalf("marshal benchmark payload: %v", err)
	}
	return encoded
}

func newBenchmarkUserPayload() benchmarkUserPayload {
	return benchmarkUserPayload{
		ID:     1001,
		Name:   "Sergio Gonzalez",
		Email:  "sergio@example.com",
		Age:    34,
		Active: true,
		Roles:  []string{"admin", "editor", "auditor"},
		Metadata: map[string]any{
			"source":      "benchmark",
			"loginCount":  128,
			"lastLoginIP": "127.0.0.1",
		},
		CreatedAt: time.Unix(1_774_695_600, 0).UTC().Format(time.RFC3339),
		Profile: benchmarkProfilePayload{
			Company: "NexGou",
			Tags:    []string{"go", "http", "json", "sqlite"},
		},
	}
}
