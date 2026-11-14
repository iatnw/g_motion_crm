package service

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type vinTransport func(*http.Request) (*http.Response, error)

func (f vinTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestVINDecoder(t *testing.T) {
	cases := []struct {
		name, body string
		status     int
		wantError  bool
	}{
		{"complete", `{"Results":[{"VIN":"1HGCM82633A004352","Make":"HONDA","Model":"Accord","Trim":"EX","ModelYear":"2003","ErrorCode":"0"}]}`, 200, false},
		{"missing trim", `{"Results":[{"VIN":"1HGCM82633A004352","Make":"HONDA","Model":"Accord","ErrorCode":"0"}]}`, 200, false},
		{"bad check digit", `{"Results":[{"VIN":"1HGCM82633A004352","Make":"HONDA","Model":"Accord","ErrorCode":"1"}]}`, 200, true},
		{"partial result", `{"Results":[{"VIN":"1HGCM82633A004352","Make":"HONDA","ErrorCode":"0"}]}`, 200, true},
		{"wrong VIN", `{"Results":[{"VIN":"1HGCM82633A004353","Make":"HONDA","Model":"Accord","ErrorCode":"0"}]}`, 200, true},
		{"empty", `{"Results":[]}`, 200, true},
		{"bad json", `not json`, 200, true},
		{"unavailable", `{}`, 503, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			decoder := NewVINDecoder(&http.Client{Transport: vinTransport(func(r *http.Request) (*http.Response, error) {
				if r.URL.Host != "vpic.nhtsa.dot.gov" || r.URL.Path != "/api/vehicles/DecodeVinValues/1HGCM82633A004352" || r.URL.Query().Get("format") != "json" {
					t.Fatalf("unexpected URL %s", r.URL)
				}
				if _, ok := r.Context().Deadline(); !ok {
					t.Fatal("missing timeout")
				}
				return &http.Response{StatusCode: tc.status, Body: io.NopCloser(strings.NewReader(tc.body)), Header: make(http.Header)}, nil
			})})
			result, err := decoder.Decode(context.Background(), " 1hgcm82633a004352 ")
			if (err != nil) != tc.wantError {
				t.Fatalf("result %+v, error %v", result, err)
			}
			if err == nil && (result.Make != "HONDA" || result.Model != "Accord") {
				t.Fatalf("incorrect vehicle %+v", result)
			}
		})
	}
}

func TestVINDecoderInvalidInputAndCancellation(t *testing.T) {
	decoder := NewVINDecoder(&http.Client{Transport: vinTransport(func(r *http.Request) (*http.Response, error) {
		if r.Context().Err() == nil {
			t.Fatal("invalid input should never reach upstream")
		}
		return nil, r.Context().Err()
	})})
	for _, vin := range []string{"", "short", "1HGCM82633A00435I", "../../etc/passwd!!"} {
		if _, err := decoder.Decode(context.Background(), vin); !errors.Is(err, ErrVINFormat) {
			t.Fatalf("unexpected input error %v", err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := decoder.Decode(ctx, "1HGCM82633A004352"); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
}
