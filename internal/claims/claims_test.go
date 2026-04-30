package claims

import (
	"testing"

	"pacto/internal/domain/claim"
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
	has := func(ct claim.Type) bool {
		for _, c := range got {
			if c.ClaimType == ct {
				return true
			}
		}
		return false
	}
	if !has(claim.Path) || !has(claim.Symbol) || !has(claim.Endpoint) || !has(claim.TestRef) {
		t.Fatalf("expected all claim categories, got %#v", got)
	}
}

func TestExtractClaimsIgnoresMultilineInlineCode(t *testing.T) {
	p := parser.ParsedPlan{
		RawText: "`src/auth.go\nRunStatus`",
	}
	got := Extract(p, Options{Paths: true, Symbols: true, Endpoints: true, TestRefs: true})
	if len(got) != 0 {
		t.Fatalf("expected multiline inline code to be ignored, got %#v", got)
	}
}
