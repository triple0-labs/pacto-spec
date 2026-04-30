package claim

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestResultJSONShape pins the wire format of claim.Result. The pacto
// status --format=json public contract depends on these field names and
// omitempty semantics; changing them is a breaking change.
func TestResultJSONShape(t *testing.T) {
	r := Result{
		ClaimType:  Path,
		SourceText: "src/main.go",
		Evidence:   "found",
		Result:     "ok",
		References: []string{"a", "b"},
	}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(b)
	for _, want := range []string{
		`"claim_type":"path"`,
		`"source_text":"src/main.go"`,
		`"evidence":"found"`,
		`"result":"ok"`,
		`"references":["a","b"]`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %s", want, got)
		}
	}
}

func TestResultReferencesOmitEmpty(t *testing.T) {
	b, err := json.Marshal(Result{ClaimType: Symbol})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(b), "references") {
		t.Errorf("expected references to be omitted, got %s", b)
	}
}

func TestTypeConstants(t *testing.T) {
	cases := map[Type]string{
		Path:     "path",
		Symbol:   "symbol",
		Endpoint: "endpoint",
		TestRef:  "test_ref",
		Delta:    "delta",
	}
	for typ, want := range cases {
		if string(typ) != want {
			t.Errorf("Type %v = %q, want %q", typ, string(typ), want)
		}
	}
}

func TestResultRoundTrip(t *testing.T) {
	in := Result{
		ClaimType:  Endpoint,
		SourceText: "GET /v1/foo",
		Evidence:   "handler:HandleFoo",
		Result:     "ok",
		References: []string{"router.go:42"},
	}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out Result
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.ClaimType != in.ClaimType ||
		out.SourceText != in.SourceText ||
		out.Evidence != in.Evidence ||
		out.Result != in.Result ||
		len(out.References) != len(in.References) ||
		out.References[0] != in.References[0] {
		t.Errorf("round-trip mismatch: got %+v want %+v", out, in)
	}
}
