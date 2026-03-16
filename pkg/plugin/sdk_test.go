package plugin

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRunWithInitRequest(t *testing.T) {
	input := `{"type":"init","config":{"key":"val"}}` + "\n"
	var out bytes.Buffer

	RunWith(strings.NewReader(input), &out, func(req Request) Response {
		if req.Type != "init" {
			t.Errorf("expected 'init', got %q", req.Type)
		}
		if req.Config["key"] != "val" {
			t.Errorf("expected config key=val, got %q", req.Config["key"])
		}
		return Response{Name: "TEST", Interval: 10, MinWidth: 30, MinHeight: 5}
	})

	var resp Response
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Name != "TEST" {
		t.Errorf("expected name 'TEST', got %q", resp.Name)
	}
	if resp.Interval != 10 {
		t.Errorf("expected interval 10, got %d", resp.Interval)
	}
}

func TestRunWithViewRequest(t *testing.T) {
	input := `{"type":"view","width":40,"height":10}` + "\n"
	var out bytes.Buffer

	RunWith(strings.NewReader(input), &out, func(req Request) Response {
		if req.Type != "view" {
			t.Errorf("expected 'view', got %q", req.Type)
		}
		if req.Width != 40 {
			t.Errorf("expected width 40, got %d", req.Width)
		}
		if req.Height != 10 {
			t.Errorf("expected height 10, got %d", req.Height)
		}
		return Response{Content: "hello"}
	})

	var resp Response
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Content != "hello" {
		t.Errorf("expected 'hello', got %q", resp.Content)
	}
}

func TestRunWithUpdateRequest(t *testing.T) {
	input := `{"type":"update"}` + "\n"
	var out bytes.Buffer

	RunWith(strings.NewReader(input), &out, func(req Request) Response {
		if req.Type != "update" {
			t.Errorf("expected 'update', got %q", req.Type)
		}
		return Response{Content: "updated data"}
	})

	var resp Response
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Content != "updated data" {
		t.Errorf("expected 'updated data', got %q", resp.Content)
	}
}

func TestRunWithMultipleRequests(t *testing.T) {
	input := `{"type":"init"}
{"type":"update"}
{"type":"view","width":20,"height":5}
`
	var out bytes.Buffer
	var types []string

	RunWith(strings.NewReader(input), &out, func(req Request) Response {
		types = append(types, req.Type)
		return Response{Content: req.Type}
	})

	if len(types) != 3 {
		t.Fatalf("expected 3 requests, got %d", len(types))
	}
	if types[0] != "init" || types[1] != "update" || types[2] != "view" {
		t.Errorf("unexpected order: %v", types)
	}

	// Verify 3 responses were written
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 response lines, got %d", len(lines))
	}
}

func TestRunWithBadJSON(t *testing.T) {
	input := "not valid json\n"
	var out bytes.Buffer

	RunWith(strings.NewReader(input), &out, func(req Request) Response {
		t.Error("handler should not be called for bad JSON")
		return Response{}
	})

	var resp Response
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Error == "" {
		t.Error("expected error response for bad JSON")
	}
}

func TestRunWithErrorResponse(t *testing.T) {
	input := `{"type":"init"}` + "\n"
	var out bytes.Buffer

	RunWith(strings.NewReader(input), &out, func(req Request) Response {
		return Response{Error: "something broke"}
	})

	var resp Response
	json.Unmarshal(out.Bytes(), &resp)
	if resp.Error != "something broke" {
		t.Errorf("expected error message, got %q", resp.Error)
	}
}

func TestRequestJSONRoundTrip(t *testing.T) {
	req := Request{
		Type:   "init",
		Config: map[string]string{"a": "b"},
		Width:  80,
		Height: 24,
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}

	var decoded Request
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Type != "init" || decoded.Config["a"] != "b" || decoded.Width != 80 {
		t.Errorf("roundtrip mismatch: %+v", decoded)
	}
}

func TestResponseJSONOmitsEmpty(t *testing.T) {
	resp := Response{Content: "hello"}
	data, _ := json.Marshal(resp)
	s := string(data)

	// Fields with zero values should be omitted
	if strings.Contains(s, "min_width") {
		t.Error("expected min_width to be omitted")
	}
	if strings.Contains(s, "interval") {
		t.Error("expected interval to be omitted")
	}
	if strings.Contains(s, "error") {
		t.Error("expected error to be omitted")
	}
}
