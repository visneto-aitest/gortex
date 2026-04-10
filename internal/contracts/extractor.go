package contracts

import (
	"github.com/zzet/gortex/internal/graph"
)

// Extractor analyses source files and produces Contract values.
type Extractor interface {
	// Extract scans the source of a single file and returns any contracts found.
	// nodes and edges provide graph context so the extractor can resolve the
	// nearest enclosing symbol for each match.
	Extract(filePath string, src []byte, nodes []*graph.Node, edges []*graph.Edge) []Contract

	// SupportedLanguages returns the set of languages this extractor handles.
	SupportedLanguages() []string
}
