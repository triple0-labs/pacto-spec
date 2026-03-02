package claims

import (
	"testing"

	"pacto/internal/model"
	"pacto/internal/parser"
)

func TestExtractClaimsDedupesAndClassifies(t *testing.T) {
	p := parser.ParsedPlan{
		RawText: "`internal/app/status.go` `RunStatus` `GET /api/status` `go test ./...` `RunStatus`",
	}
	got := Extract(p, Options{Paths: true, Symbols: true, Endpoints: true, TestRefs: true})
	if len(got) == 0 {
		t.Fatalf("expected claims")
	}
	has := func(ct model.ClaimType) bool {
		for _, c := range got {
			if c.ClaimType == ct {
				return true
			}
		}
		return false
	}
	if !has(model.ClaimPath) || !has(model.ClaimSymbol) || !has(model.ClaimEndpoint) || !has(model.ClaimTestRef) {
		t.Fatalf("expected all claim categories, got %#v", got)
	}
}
