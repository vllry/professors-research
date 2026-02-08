package prizeodds

// CalculateDrawOdds returns the probability of drawing 1+ copies of a target card
// when drawing drawCount cards from a poolSize-card population that contains
// targetCount copies of the target.
//
// This is the hypergeometric complement:
//   P(X >= 1) = 1 - P(X = 0)
//            = 1 - C(poolSize-targetCount, drawCount) / C(poolSize, drawCount)
func CalculateDrawOdds(poolSize, drawCount, targetCount int) float64 {
	if poolSize <= 0 || drawCount <= 0 || targetCount <= 0 {
		return 0.0
	}
	if drawCount > poolSize {
		return 0.0
	}
	if targetCount > poolSize {
		return 0.0
	}

	denominator := comb(poolSize, drawCount)
	if denominator == 0 {
		return 0.0
	}

	// If we can't draw drawCount cards that are all non-target, then P(0 targets) = 0.
	numerator := comb(poolSize-targetCount, drawCount)
	p0 := float64(numerator) / float64(denominator)
	return 1.0 - p0
}

// CalculateDrawPairOdds returns the probability of drawing at least one copy of BOTH
// card A and card B when drawing drawCount cards from a poolSize-card population.
// The population contains countA copies of A and countB copies of B.
//
// Using inclusion-exclusion:
//   P(A>=1 AND B>=1) = 1 - P(A=0) - P(B=0) + P(A=0 AND B=0)
func CalculateDrawPairOdds(poolSize, drawCount, countA, countB int) float64 {
	if poolSize <= 0 || drawCount <= 0 || countA <= 0 || countB <= 0 {
		return 0.0
	}
	if drawCount > poolSize {
		return 0.0
	}
	if countA > poolSize || countB > poolSize {
		return 0.0
	}
	if countA+countB > poolSize {
		return 0.0
	}

	denominator := comb(poolSize, drawCount)
	if denominator == 0 {
		return 0.0
	}

	pNoA := float64(comb(poolSize-countA, drawCount)) / float64(denominator)
	pNoB := float64(comb(poolSize-countB, drawCount)) / float64(denominator)
	pNoAB := float64(comb(poolSize-countA-countB, drawCount)) / float64(denominator)

	p := 1.0 - pNoA - pNoB + pNoAB
	if p < 0.0 {
		return 0.0
	}
	if p > 1.0 {
		return 1.0
	}
	return p
}

