export interface CardJSON {
  name: string;
  setCode: string;
  number: string;
}

export interface AnyOfJSON {
  cards: Array<{
    card: CardJSON;
    count: number;
  }>;
}

export interface AllOfJSON {
  cards: Array<{
    card: CardJSON;
    count: number;
  }>;
}

export interface CardSetJSON {
  anyOfs?: AnyOfJSON[];
  allOfs?: AllOfJSON[];
}

export interface PrizeOddsRequest {
  decklist: string;
  cardSets?: Record<string, CardSetJSON[]>;
  prized?: boolean; // true = in 6 prize cards (default), false = in 54 not-prized cards
}

export interface APIError {
  type: string;
  info: string;
}

export interface PrizeOddsResponse {
  odds: Record<string, number[]>;
  cardSetOdds?: Record<string, Record<string, number>>;
  errors?: APIError[];
}

export interface StartOddsRequest {
  decklist: string;
  cardSets?: Record<string, CardSetJSON[]>;
}

export interface StartOddsResponse {
  odds: Record<string, number[]>;
  possibleStarters: Record<string, number>;
  forcedStarters: Record<string, number>;
  mulliganOdds: number;
  atLeastOneBasic: number;
  atLeastTwoBasic: number;
  cardSetOdds?: Record<string, Record<string, number>>;
  errors?: APIError[];
}

export interface DrawSupporterOddsRequest {
  deckSize: number;
  knownBottom: number;
  handSize: number;
  prizeCards: number;
}

export interface DrawSupporterOddsResponse {
  // odds maps supporter name -> odds for 1,2,3,4 copies of the target in the pool.
  odds: Record<string, number[]>;
  // pairOdds maps supporter name -> 4x4 table for odds of drawing at least one of BOTH cards,
  // indexed by [countA-1][countB-1] where countA,countB are 1..4.
  pairOdds: Record<string, number[][]>;
  // bottomOdds maps supporter name -> odds for 1,2,3,4 copies of the target
  // among known bottom cards when the draw goes past the top of deck.
  bottomOdds?: Record<string, number[]>;
  // drawCounts maps supporter name -> draw count for the top/shuffled pool.
  drawCounts: Record<string, number>;
  // effectiveDrawCounts maps supporter name -> actual draw count used in the model
  // after clamping to the available pool.
  effectiveDrawCounts: Record<string, number>;
  // bottomDrawCounts maps supporter name -> number of cards drawn into known bottom.
  bottomDrawCounts?: Record<string, number>;
}

export interface ErrorResponse {
  error: string;
}

export interface TournamentResponse {
  id: string;
  year: number;
  location: string;
}

export interface MatchupRecord {
  wins: number;
  losses: number;
  ties: number;
  winRate: number;
}

export interface VariantMatchupStats {
  cardCounts: Record<string, number>;
  matchups: Record<string, MatchupRecord>;
}

export interface MatchupStatsRequest {
  tournamentIds: string[];
  archetype: string;
  variants?: Array<Record<string, number>>;
  playerPlacement?: number;
  opponentPlacement?: number;
}

export interface MatchupStatsResponse {
  matchups: Record<string, VariantMatchupStats>;
  archetypeCounts: Record<string, number>;
  variantCounts: Record<string, number>;
}

