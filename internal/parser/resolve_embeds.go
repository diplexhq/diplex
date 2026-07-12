package parser

import (
	"maps"
)

// resolveEmbeds resolves embedded interfaces by copying their methods into
// the parent interface. Uses the separate embeds map instead of InterfaceInfo.Embeds.
// If an embedded interface is not found in scope, the parent interface is
// removed from the result — without a complete method set, it cannot be
// reliably matched against implementations.
// Note: embs is a copy of the slice value from state.embeds[interfaceID];
// append/embs[1:] mutations operate on the copy, not the map value. This is intentional —
// embs acts as a BFS queue for embedded interfaces to flatten.
func (fp *Parser) resolveEmbeds(state *parseState) {
	for interfaceID, embs := range state.embeds {
		for len(embs) > 0 {
			embedInfo, ok := state.interfaces[embs[0]]
			if !ok {
				// Missing embed — interface cannot be fully resolved.
				fp.log.Debug("incomplete embed", "interface", interfaceID, "missing", embs[0])
				embs = embs[1:]

				continue
			}

			embed := embedInfo.Methods
			maps.Copy(state.interfaces[interfaceID].Methods, embed)

			embs = append(embs, state.embeds[embs[0]]...)
			embs = embs[1:]
		}

		fp.log.Debug("embeds flattened", "interface", interfaceID, "methods", len(state.interfaces[interfaceID].Methods))
	}

	state.embeds = nil
}
