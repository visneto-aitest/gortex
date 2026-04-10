package contracts

// CrossLink represents a matched provider-consumer pair, possibly across repos.
type CrossLink struct {
	ContractID string   `json:"contract_id"`
	Provider   Contract `json:"provider"`
	Consumer   Contract `json:"consumer"`
	CrossRepo  bool     `json:"cross_repo"`
}

// MatchResult holds the output of a matching pass.
type MatchResult struct {
	Matched         []CrossLink `json:"matched"`
	OrphanProviders []Contract  `json:"orphan_providers"`
	OrphanConsumers []Contract  `json:"orphan_consumers"`
}

// Match analyses a registry and pairs providers with consumers by contract ID.
// A provider in repo A matched with a consumer in repo B is marked CrossRepo.
func Match(reg *Registry) MatchResult {
	var result MatchResult

	for _, id := range reg.AllIDs() {
		contracts := reg.ByID(id)

		var providers, consumers []Contract
		for _, c := range contracts {
			switch c.Role {
			case RoleProvider:
				providers = append(providers, c)
			case RoleConsumer:
				consumers = append(consumers, c)
			}
		}

		if len(consumers) == 0 {
			result.OrphanProviders = append(result.OrphanProviders, providers...)
			continue
		}
		if len(providers) == 0 {
			result.OrphanConsumers = append(result.OrphanConsumers, consumers...)
			continue
		}

		// Pair each consumer with each provider.
		for _, consumer := range consumers {
			for _, provider := range providers {
				crossRepo := provider.RepoPrefix != consumer.RepoPrefix
				result.Matched = append(result.Matched, CrossLink{
					ContractID: id,
					Provider:   provider,
					Consumer:   consumer,
					CrossRepo:  crossRepo,
				})
			}
		}
	}

	return result
}
