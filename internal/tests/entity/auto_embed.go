package entity

// BaseStats — exported helper struct with Stats() and Slug(), embedded into AutoEmbedService and MixedEmbedService.
type BaseStats struct{}

func (s *BaseStats) Stats() int   { return 0 }
func (s *BaseStats) Slug() string { return "base" }

// AutoEmbedService embeds two interfaces (Pingable, Countable) plus BaseStats.
// BaseStats provides Stats() and Slug() via method promotion for stats.Repo.
type AutoEmbedService struct {
	Pingable  // embedded interface dep 1
	Countable // embedded interface dep 2
	BaseStats
}

// MixedEmbedService embeds Pingable AND has a named field of the same type.
// Tests deduplication: named field takes precedence, embedded is skipped.
// Also embeds *BaseStats for pointer embedding.
type MixedEmbedService struct {
	Pingable           // embedded — deduplicated by named Storage below
	Storage   Pingable // named field — takes precedence (exported, shadows embed)
	Countable          // embedded dep
	*BaseStats
}
