package prizeodds

// OddsTable is a read-only table of prizing odds.
// Table[x][y] represents the probability of prizing y copies when x copies are in deck.
type OddsTable struct {
	table map[int]map[int]float64
}

// Get returns the probability of prizing y copies when x copies are in deck.
// Returns 0.0 if x or y are out of valid range (x: 0-59, y: 0-6).
func (ot *OddsTable) Get(x, y int) float64 {
	if x < 0 || x > 59 || y < 0 || y > 6 {
		return 0.0
	}
	if row, ok := ot.table[x]; ok {
		if val, ok := row[y]; ok {
			return val
		}
	}
	return 0.0
}

// comb calculates n choose k (binomial coefficient)
func comb(n, k int) int64 {
	if k > n || k < 0 {
		return 0
	}
	if k > n-k {
		k = n - k // Take advantage of symmetry
	}
	result := int64(1)
	for i := 0; i < k; i++ {
		result = result * int64(n-i) / int64(i+1)
	}
	return result
}

// calculateOddsTable computes the odds table using hypergeometric distribution.
// Constraints: 60 card deck, max 59 copies of a card.
// P(X = j) = C(i, j) * C(60-i, 6-j) / C(60, 6)
func calculateOddsTable() *OddsTable {
	table := make(map[int]map[int]float64)
	for i := 0; i <= 59; i++ {
		table[i] = make(map[int]float64)
		for j := 0; j <= 6; j++ {
			// Hypergeometric distribution: probability of j successes in 6 draws
			// from a population of 60 with i successes
			if j > i || (6-j) > (60-i) {
				table[i][j] = 0.0
			} else {
				numerator := float64(comb(i, j)) * float64(comb(60-i, 6-j))
				denominator := float64(comb(60, 6))
				table[i][j] = numerator / denominator
			}
		}
	}
	return &OddsTable{table: table}
}

// StartOddsTable is a read-only table of starting hand odds.
// Table[x][y] represents the probability of having y copies in starting hand when x copies are in deck.
type StartOddsTable struct {
	table map[int]map[int]float64
}

// Get returns the probability of having y copies in starting hand when x copies are in deck.
// Returns 0.0 if x or y are out of valid range (x: 0-59, y: 0-8).
func (ot *StartOddsTable) Get(x, y int) float64 {
	if x < 0 || x > 59 || y < 0 || y > 8 {
		return 0.0
	}
	if row, ok := ot.table[x]; ok {
		if val, ok := row[y]; ok {
			return val
		}
	}
	return 0.0
}

// calculateStartOddsTable computes the odds table using hypergeometric distribution.
// Constraints: 60 card deck, max 59 copies of a card.
// P(X = j) = C(i, j) * C(60-i, 8-j) / C(60, 8)
func calculateStartOddsTable() *StartOddsTable {
	table := make(map[int]map[int]float64)
	for i := 0; i <= 59; i++ {
		table[i] = make(map[int]float64)
		for j := 0; j <= 8; j++ {
			// Hypergeometric distribution: probability of j successes in 8 draws
			// from a population of 60 with i successes
			if j > i || (8-j) > (60-i) {
				table[i][j] = 0.0
			} else {
				numerator := float64(comb(i, j)) * float64(comb(60-i, 8-j))
				denominator := float64(comb(60, 8))
				table[i][j] = numerator / denominator
			}
		}
	}
	return &StartOddsTable{table: table}
}

// DefaultOddsTable is the pre-calculated odds table, initialized at package startup.
var DefaultOddsTable *OddsTable

// DefaultStartOddsTable is the pre-calculated starting hand odds table, initialized at package startup.
var DefaultStartOddsTable *StartOddsTable

// SevenCardOddsTable is a read-only table of 7-card draw odds.
// Table[x][y] represents the probability of having y copies in 7-card draw when x copies are in deck.
type SevenCardOddsTable struct {
	table map[int]map[int]float64
}

// Get returns the probability of having y copies in 7-card draw when x copies are in deck.
// Returns 0.0 if x or y are out of valid range (x: 0-59, y: 0-7).
func (ot *SevenCardOddsTable) Get(x, y int) float64 {
	if x < 0 || x > 59 || y < 0 || y > 7 {
		return 0.0
	}
	if row, ok := ot.table[x]; ok {
		if val, ok := row[y]; ok {
			return val
		}
	}
	return 0.0
}

// calculateSevenCardOddsTable computes the odds table using hypergeometric distribution.
// Constraints: 60 card deck, max 59 copies of a card.
// P(X = j) = C(i, j) * C(60-i, 7-j) / C(60, 7)
func calculateSevenCardOddsTable() *SevenCardOddsTable {
	table := make(map[int]map[int]float64)
	for i := 0; i <= 59; i++ {
		table[i] = make(map[int]float64)
		for j := 0; j <= 7; j++ {
			// Hypergeometric distribution: probability of j successes in 7 draws
			// from a population of 60 with i successes
			if j > i || (7-j) > (60-i) {
				table[i][j] = 0.0
			} else {
				numerator := float64(comb(i, j)) * float64(comb(60-i, 7-j))
				denominator := float64(comb(60, 7))
				table[i][j] = numerator / denominator
			}
		}
	}
	return &SevenCardOddsTable{table: table}
}

// DefaultSevenCardOddsTable is the pre-calculated 7-card draw odds table, initialized at package startup.
var DefaultSevenCardOddsTable *SevenCardOddsTable

func init() {
	DefaultOddsTable = calculateOddsTable()
	DefaultStartOddsTable = calculateStartOddsTable()
	DefaultSevenCardOddsTable = calculateSevenCardOddsTable()
}
