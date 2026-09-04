import { matchesTagFilter, normalizeTagQuery } from "./markup.js";
import { state } from "./state.js";
import { findCardByIdentifier } from "./board_queries.js";

export function applyCardSearchHighlights() {
  const query = normalizeTagQuery(state.cardSearchQuery);
  document.querySelectorAll("#lists .card").forEach((cardElement) => {
    const currentCard = findCardByIdentifier(Number(cardElement.dataset.cardIdentifier));
    const cardTags = (currentCard && currentCard.tags) || [];
    const isMatch = Boolean(query) && cardTags.some((tag) => matchesTagFilter(tag, query));
    cardElement.classList.toggle("highlight", isMatch);
    cardElement.classList.toggle("dim", Boolean(query) && !isMatch);
  });
}
