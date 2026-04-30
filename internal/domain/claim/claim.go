// Package claim defines plan-claim domain types.
//
// A Claim is something a plan asserts about the repository (a path, symbol,
// endpoint, test reference, or delta). A Result captures verification outcome.
//
// This package is pure: it depends only on the standard library.
package claim

// Type enumerates supported claim kinds.
type Type string

const (
	Path     Type = "path"
	Symbol   Type = "symbol"
	Endpoint Type = "endpoint"
	TestRef  Type = "test_ref"
	Delta    Type = "delta"
)

// Result is the outcome of verifying (or merely extracting) a single claim.
type Result struct {
	ClaimType  Type     `json:"claim_type"`
	SourceText string   `json:"source_text"`
	Evidence   string   `json:"evidence"`
	Result     string   `json:"result"`
	References []string `json:"references,omitempty"`
}
