package mcp

import (
	"sort"

	"github.com/zzet/gortex/internal/graph"
)

// applyRerankBoosts reorders nodes by combining combo + frecency signals
// on top of the backend's original BM25-like order. Both inputs may be
// nil; if both produce unit multipliers, the stable sort is a no-op and
// the input order is preserved byte-for-byte. Kept in one pass so the
// comparator sees the final combined multiplier per node.
func applyRerankBoosts(nodes []*graph.Node, cm *comboManager, ft *frecencyTracker, query string) []*graph.Node {
	if len(nodes) < 2 {
		return nodes
	}
	var comboBoosts map[string]float64
	if cm != nil {
		comboBoosts = cm.BoostMap(query)
	}
	// Only sort when at least one signal has something to say.
	if len(comboBoosts) == 0 && (ft == nil || !ft.HasData()) {
		return nodes
	}

	multiplier := func(id string) float64 {
		m := 1.0
		if b, ok := comboBoosts[id]; ok {
			m *= b
		}
		if ft != nil {
			m *= ft.BoostFor(id)
		}
		return m
	}

	sort.SliceStable(nodes, func(i, j int) bool {
		return multiplier(nodes[i].ID) > multiplier(nodes[j].ID)
	})
	return nodes
}

// recordLastSearchFromNodes stores the query + top-limit IDs on the session
// so a subsequent get_symbol_source / get_editing_context can credit this
// search. Capped at limit to avoid crediting results the agent never saw.
func recordLastSearchFromNodes(sess *sessionState, query string, nodes []*graph.Node, limit int) {
	if sess == nil || len(nodes) == 0 {
		return
	}
	if limit <= 0 || limit > len(nodes) {
		limit = len(nodes)
	}
	ids := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		ids = append(ids, nodes[i].ID)
	}
	sess.recordLastSearch(query, ids)
}
