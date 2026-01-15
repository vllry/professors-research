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

export interface ErrorResponse {
  error: string;
}



