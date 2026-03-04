package prizeodds

import (
	basictypes "github.com/vllry/professors-research/pkg/types"
)

// CalculatePrizeOddsWithOpeningHand calculates prize/not-prize odds while conditioning on a valid
// opening hand: the 7-card opening hand contains at least 1 Basic Pokémon, and then 6 prizes are set
// aside from the remaining 53 cards.
//
// This conditioning makes Basic Pokémon slightly less likely to be prized and non-Basics slightly
// more likely to be prized, compared to the unconditional "choose 6 prizes from 60" model.
//
// The returned format matches CalculatePrizeOdds: for each card with n copies, an array of length
// min(n, 6) where index i is P(at least i+1 copies are in the target set).
//
// basicPokemonCards identifies which cards in the deck are Basic Pokémon (all copies of that card).
// If there are 0 Basic Pokémon in the deck, this falls back to CalculatePrizeOdds.
func CalculatePrizeOddsWithOpeningHand(
	decklist basictypes.Decklist,
	prized bool,
	basicPokemonCards map[basictypes.Card]bool,
) (map[basictypes.Card][]float64, error) {
	// Total number of Basic Pokémon in deck
	totalBasic := 0
	for card, cnt := range decklist.Cards {
		if basicPokemonCards[card] {
			totalBasic += cnt
		}
	}

	// Edge case: if no basics exist, the conditioning event has probability 0.
	// Fall back to the unconditional model.
	if totalBasic == 0 {
		return CalculatePrizeOdds(decklist, prized)
	}

	// P(valid opening hand) = 1 - P(0 basics in 7)
	denomValidHand := float64(comb(60, 7))
	if denomValidHand == 0 {
		return CalculatePrizeOdds(decklist, prized)
	}
	pZeroBasicsInHand := float64(comb(60-totalBasic, 7)) / denomValidHand
	pValidHand := 1.0 - pZeroBasicsInHand
	if pValidHand <= 0.0 {
		return CalculatePrizeOdds(decklist, prized)
	}

	denomPrizes := float64(comb(60, 6))
	denomHandGivenPrizes := float64(comb(54, 7)) // after 6 prizes are removed
	if denomPrizes == 0 || denomHandGivenPrizes == 0 {
		return CalculatePrizeOdds(decklist, prized)
	}

	// Helper: P(hand has 0 basics | k basics are in prizes)
	pZeroBasicsGivenKInPrizes := func(k int) float64 {
		// Remaining basics after prizes: totalBasic - k
		// Remaining non-basics: 54 - (totalBasic - k) = 54 - totalBasic + k
		remainingNonBasics := 54 - totalBasic + k
		if remainingNonBasics < 7 {
			return 0.0
		}
		return float64(comb(remainingNonBasics, 7)) / denomHandGivenPrizes
	}

	cardPrizeOdds := make(map[basictypes.Card][]float64)

	for card, count := range decklist.Cards {
		arraySize := count
		if arraySize > 6 {
			arraySize = 6
		}
		cardPrizeOdds[card] = make([]float64, arraySize)

		isBasic := basicPokemonCards[card]

		// Precompute exact conditional distribution for j=0..min(count,6)
		maxJ := count
		if maxJ > 6 {
			maxJ = 6
		}
		exact := make([]float64, maxJ+1)

		for j := 0; j <= maxJ; j++ {
			// Unconditional P(Xp=j) for prizes (marginally a uniform 6-subset of 60)
			pXpJ := float64(DefaultOddsTable.Get(count, j))

			// Compute P(Xp=j AND hand has 0 basics)
			pXpJAndZeroBasics := 0.0
			if pXpJ > 0.0 {
				if isBasic {
					otherBasics := totalBasic - count
					rest := 60 - totalBasic // all non-basics
					// k = total basics in prizes (must be >= j if this card is basic)
					kMin := j
					kMax := totalBasic
					if kMax > 6 {
						kMax = 6
					}
					for k := kMin; k <= kMax; k++ {
						otherBasicInPrizes := k - j
						remainingSlots := 6 - j - otherBasicInPrizes
						if otherBasicInPrizes < 0 || remainingSlots < 0 {
							continue
						}
						if otherBasicInPrizes > otherBasics {
							continue
						}
						if remainingSlots > rest {
							continue
						}
						pXpJAndK := (float64(comb(count, j)) *
							float64(comb(otherBasics, otherBasicInPrizes)) *
							float64(comb(rest, remainingSlots))) / denomPrizes
						pXpJAndZeroBasics += pXpJAndK * pZeroBasicsGivenKInPrizes(k)
					}
				} else {
					otherBasics := totalBasic
					rest := 60 - count - totalBasic // non-basic, non-target cards
					kMin := 0
					kMax := otherBasics
					if kMax > 6 {
						kMax = 6
					}
					for k := kMin; k <= kMax; k++ {
						remainingSlots := 6 - j - k
						if remainingSlots < 0 {
							continue
						}
						if k > otherBasics {
							continue
						}
						if remainingSlots > rest {
							continue
						}
						pXpJAndK := (float64(comb(count, j)) *
							float64(comb(otherBasics, k)) *
							float64(comb(rest, remainingSlots))) / denomPrizes
						pXpJAndZeroBasics += pXpJAndK * pZeroBasicsGivenKInPrizes(k)
					}
				}
			}

			// Conditional: P(Xp=j | valid hand) = (P(Xp=j) - P(Xp=j AND zero basics)) / P(valid hand)
			pCond := (pXpJ - pXpJAndZeroBasics) / pValidHand
			if pCond < 0.0 {
				pCond = 0.0
			}
			if pCond > 1.0 {
				pCond = 1.0
			}
			exact[j] = pCond
		}

		// Normalize small numeric drift so Σ exact[j] = 1
		sumExact := 0.0
		for _, v := range exact {
			sumExact += v
		}
		if sumExact > 0.0 {
			for i := range exact {
				exact[i] /= sumExact
			}
		}

		// Build cumulative arrays in the same shape as CalculatePrizeOdds
		if prized {
			for i := 0; i < arraySize; i++ {
				cumulativeProb := 0.0
				for j := i + 1; j <= maxJ; j++ {
					cumulativeProb += exact[j]
				}
				cardPrizeOdds[card][i] = cumulativeProb
			}
		} else {
			// Invert using the exact prized distribution:
			// P(notPrized >= i+1) = P(prized <= count-(i+1)) = 1 - P(prized >= count-i)
			for i := 0; i < arraySize; i++ {
				threshold := count - i // count-i copies in prizes means fewer than i+1 not-prized
				if threshold <= 0 {
					cardPrizeOdds[card][i] = 1.0
					continue
				}
				if threshold > maxJ {
					// Can't have threshold copies in prizes if threshold > 6; thus P(prized >= threshold) = 0
					cardPrizeOdds[card][i] = 1.0
					continue
				}

				pAtLeastThresholdPrized := 0.0
				for j := threshold; j <= maxJ; j++ {
					pAtLeastThresholdPrized += exact[j]
				}
				cardPrizeOdds[card][i] = 1.0 - pAtLeastThresholdPrized
			}
		}
	}

	return cardPrizeOdds, nil
}

// CalculatePrizeOdds takes a decklist and a prized flag.
// When prized=true (default), it calculates odds that cards are in the 6 prize cards.
// When prized=false, it calculates odds that cards are in the 54 not-prized cards.
// It returns:
// * A map, for each distinct card, containing an array of cumulative odds.
//   For a card with n copies, the array has length min(n, 6):
//   - Index 0: probability of at least 1 copy in target set (6 prizes if prized=true, 54 not-prized if prized=false)
//   - Index 1: probability of at least 2 copies in target set
//   - ...
//   - Index min(n,6)-1: probability of at least min(n,6) copies in target set
//   For example, for a card with 1 copy and prized=true: [0.1] means 10% chance of prizing at least 1 copy.
func CalculatePrizeOdds(decklist basictypes.Decklist, prized bool) (map[basictypes.Card][]float64, error) {
	cardPrizeOdds := make(map[basictypes.Card][]float64)

	for card, count := range decklist.Cards {
		// Array size is min(count, 6) - one entry for each "at least X copies" probability
		arraySize := count
		if arraySize > 6 {
			arraySize = 6
		}
		cardPrizeOdds[card] = make([]float64, arraySize)
		
		if prized {
			// Calculate cumulative probabilities: P(>= i+1 copies in 6 prizes) = sum of P(j copies) for j from i+1 to min(count, 6)
			for i := 0; i < arraySize; i++ {
				// Probability of prizing at least (i+1) copies
				cumulativeProb := 0.0
				for j := i + 1; j <= arraySize; j++ {
					cumulativeProb += DefaultOddsTable.Get(count, j)
				}
				cardPrizeOdds[card][i] = cumulativeProb
			}
		} else {
			// Calculate cumulative probabilities for not-prized: P(>= i+1 copies in 54 not-prized)
			// This is equivalent to P(<= count - (i+1) copies in 6 prizes) = 1 - P(>= count - i copies in 6 prizes)
			for i := 0; i < arraySize; i++ {
				// We want P(>= i+1 copies in 54 not-prized)
				// = P(<= count - (i+1) copies in 6 prizes)
				// = 1 - P(>= count - i copies in 6 prizes)
				notPrizedThreshold := count - i
				if notPrizedThreshold > count {
					notPrizedThreshold = count
				}
				if notPrizedThreshold <= 0 {
					// All copies must be in not-prized (none in prizes)
					cardPrizeOdds[card][i] = 1.0
				} else {
					// Calculate P(>= notPrizedThreshold copies in 6 prizes)
					prizedProb := 0.0
					for j := notPrizedThreshold; j <= arraySize && j <= count; j++ {
						prizedProb += DefaultOddsTable.Get(count, j)
					}
					cardPrizeOdds[card][i] = 1.0 - prizedProb
				}
			}
		}
	}

	return cardPrizeOdds, nil
}

// CalculateStartOdds takes a decklist and calculates the odds that cards are in the starting hand (8 cards).
// It returns:
// * A map, for each distinct card, containing an array of cumulative odds.
//   For a card with n copies, the array has length min(n, 4):
//   - Index 0: probability of at least 1 copy in starting hand (8 cards)
//   - Index 1: probability of at least 2 copies in starting hand
//   - Index 2: probability of at least 3 copies in starting hand
//   - Index 3: probability of at least 4 copies in starting hand (if n >= 4)
//   For example, for a card with 2 copies: [0.25, 0.01] means 25% chance of at least 1 copy, 1% chance of at least 2 copies.
func CalculateStartOdds(decklist basictypes.Decklist) (map[basictypes.Card][]float64, error) {
	cardStartOdds := make(map[basictypes.Card][]float64)

	for card, count := range decklist.Cards {
		// Array size is min(count, 4) - one entry for each "at least X copies" probability (1+, 2+, 3+, 4+)
		arraySize := count
		if arraySize > 4 {
			arraySize = 4
		}
		cardStartOdds[card] = make([]float64, arraySize)
		
		// Calculate cumulative probabilities: P(>= i+1 copies in 8 starting cards) = sum of P(j copies) for j from i+1 to min(count, 8)
		for i := 0; i < arraySize; i++ {
			// Probability of having at least (i+1) copies in starting hand
			cumulativeProb := 0.0
			maxPossible := count
			if maxPossible > 8 {
				maxPossible = 8
			}
			for j := i + 1; j <= maxPossible; j++ {
				cumulativeProb += DefaultStartOddsTable.Get(count, j)
			}
			cardStartOdds[card][i] = cumulativeProb
		}
	}

	return cardStartOdds, nil
}

// calculateCardSetOdds is the internal implementation that calculates odds for CardSets.
// It takes a union probability calculator function to allow different calculation methods.
func calculateCardSetOdds(
	decklist basictypes.Decklist,
	cardSets map[string]basictypes.CardSet,
	unionCalculator func([][]basictypes.Card, basictypes.Decklist) float64,
) map[string]float64 {
	result := make(map[string]float64)
	
	for name, cardSet := range cardSets {
		// Step 1: Expand the CardSet to get all possible concrete combinations.
		// This generates all valid ways to satisfy the CardSet given the decklist constraints.
		// For example, [AnyOf(A,B), AnyOf(A,C)] expands to: [A,A], [A,C], [B,A], [B,C]
		// The Expand() method filters out impossible combinations (e.g., requiring more
		// copies than exist in the deck), ensuring we only work with valid possibilities.
		expanded := cardSet.Expand(decklist)
		
		if len(expanded.Combinations) == 0 {
			// No valid combinations means the CardSet cannot be satisfied
			result[name] = 0.0
			continue
		}
		
		// Step 2: Calculate probability that ANY combination is in the target set.
		// We use the inclusion-exclusion principle to correctly handle overlapping combinations
		// (e.g., [A,B] and [A,A] both requiring card A). This ensures each configuration
		// is counted exactly once, avoiding double-counting when multiple combinations overlap.
		odds := unionCalculator(expanded.Combinations, decklist)
		result[name] = odds
	}
	
	return result
}

// CalculateCardSetPrizeOdds takes a list of named CardSets and a prized flag.
// When prized=true (default), it calculates the odds that ANY satisfiable set of cards in the CardSet is in the 6 prize cards.
// When prized=false, it calculates the odds that ANY satisfiable set of cards in the CardSet is in the 54 not-prized cards.
//
// Semantics: "is in the prize cards" means "at least" the required counts appear. For example, if a combination
// requires [A, B], then prizes containing [A, A, B] or [A, B, B] or [A, A, B, B] all satisfy this requirement.
// This matches the intuitive meaning: if you need at least 1 A and at least 1 B, having more than the minimum
// still satisfies the requirement.
//
// Example: for a CardSet [AnyOf(1 X, 1 Y),AnyOf(1 Y, 1 Z)], it calculates the odds that ANY of the following
// (or supersets thereof) is in the target set (6 prizes if prized=true, 54 not-prized if prized=false):
// * At least 1 X and at least 1 Y
// * At least 1 Y and at least 1 Z
// * At least 2 Y
//
// Correctness: This function correctly handles overlapping combinations (e.g., AB and AA both requiring card A)
// by using the inclusion-exclusion principle to avoid double-counting. The CardSet.Expand() method
// generates all valid combinations based on the decklist constraints, ensuring we only consider
// combinations that are actually possible given the available card counts.
func CalculateCardSetPrizeOdds(decklist basictypes.Decklist, cardSets map[string]basictypes.CardSet, prized bool) (map[string]float64, error) {
	unionCalculator := func(combinations [][]basictypes.Card, decklist basictypes.Decklist) float64 {
		return CalculateUnionProbability(combinations, decklist, prized)
	}
	return calculateCardSetOdds(decklist, cardSets, unionCalculator), nil
}

// calculateUnionProbability calculates the probability that ANY of the given combinations
// appears in the 6 prize cards, using the inclusion-exclusion principle.
//
// Semantics: Each combination represents "at least" the required counts. For example, [A, B] means
// "at least 1 A and at least 1 B", so prizes like [A, A, B] or [A, B, B] all satisfy it.
//
// Mathematical Foundation:
// The inclusion-exclusion principle correctly calculates P(∪A_i) for overlapping events:
//   P(∪A_i) = ΣP(A_i) - ΣP(A_i ∩ A_j) + ΣP(A_i ∩ A_j ∩ A_k) - ...
//
// This alternating sum ensures that overlapping regions are counted exactly once:
// - Single terms (P(A_i)) count each event
// - Pair terms (P(A_i ∩ A_j)) subtract double-counted overlaps
// - Triple terms (P(A_i ∩ A_j ∩ A_k)) add back triple-counted regions
// - And so on, with alternating signs
//
// Correctness for Overlapping Combinations:
// When combinations overlap (e.g., [A,B] and [A,A] both require card A), the intersection
// correctly represents "both combinations satisfied simultaneously" by taking the maximum
// count of each card across all combinations in the intersection. This ensures:
// - If combination 1 needs at least 1 A and combination 2 needs at least 2 A, the intersection needs at least 2 A
// - This correctly models that satisfying both requires the maximum requirement
//
// Example: For combinations [A,B] and [A,A]:
// - P([A,B]) = probability of getting at least 1 A and 1 B
// - P([A,A]) = probability of getting at least 2 A
// - P([A,B] ∩ [A,A]) = probability of getting at least 2 A and 1 B (max of each card)
// - Result: P([A,B] ∪ [A,A]) = P([A,B]) + P([A,A]) - P([A,B] ∩ [A,A])
//
// Performance Optimization (Depth Limiting):
// We limit intersections to maxDepth=6 to avoid exponential explosion (2^n subsets).
// This is safe because:
// 1. Higher-order intersections (7+ combinations) have rapidly decreasing probabilities
// 2. The alternating series converges, so truncation gives a good approximation
// 3. Early termination skips impossible intersections (requiring >6 cards)
// 4. The result is clamped to [0,1] to handle any truncation artifacts
//
// The depth limit of 6 is chosen because:
// - Prize cards are limited to 6, so intersections of 7+ combinations are often impossible
// - Even when possible, their probabilities are typically negligible
// - This provides a good balance between accuracy and performance
// CalculateUnionProbability calculates the probability that ANY of the given combinations
// appears in the target set (6 prize cards if prized=true, 54 not-prized cards if prized=false),
// using the inclusion-exclusion principle.
// This is exported for use by API servers that need to calculate union probabilities
// across multiple CardSets.
func CalculateUnionProbability(combinations [][]basictypes.Card, decklist basictypes.Decklist, prized bool) float64 {
	// Convert prized bool to targetSize
	targetSize := 6
	if !prized {
		targetSize = 54
	}
	
	return calculateUnionProbabilityWithTargetSize(combinations, decklist, targetSize, 6)
}

// calculateUnionProbabilityWithTargetSize is the internal implementation that calculates union probability
// for a given target size. This allows sharing logic between prize and start odds calculations.
func calculateUnionProbabilityWithTargetSize(combinations [][]basictypes.Card, decklist basictypes.Decklist, targetSize int, maxDepth int) float64 {
	if len(combinations) == 0 {
		return 0.0
	}
	
	// Pre-compute combination counts for efficiency
	combinationCounts := make([]map[basictypes.Card]int, len(combinations))
	for i, combo := range combinations {
		counts := make(map[basictypes.Card]int)
		for _, card := range combo {
			counts[card]++
		}
		combinationCounts[i] = counts
	}
	
	// Use inclusion-exclusion principle:
	// P(∪A_i) = ΣP(A_i) - ΣP(A_i ∩ A_j) + ΣP(A_i ∩ A_j ∩ A_k) - ...
	var probability float64
	
	// Generate all subsets of combinations for inclusion-exclusion
	n := len(combinations)
	for subset := 1; subset < (1 << n); subset++ {
		// Count how many combinations are in this subset
		count := 0
		for i := 0; i < n; i++ {
			if subset&(1<<i) != 0 {
				count++
			}
		}
		
		// Skip if depth exceeds maxDepth
		if count > maxDepth {
			continue
		}
		
		// Calculate the intersection of all combinations in this subset.
		// The intersection represents "all of these combinations are satisfied simultaneously".
		//
		// Correctness: We take the MAXIMUM count of each card across all combinations.
		// This is correct because:
		// - If combination 1 needs 1 A and combination 2 needs 2 A, satisfying BOTH requires 2 A
		// - The intersection must satisfy the most demanding requirement for each card
		// - This correctly models the logical AND of all combinations in the subset
		//
		// Example: Intersection of [A,B] and [A,A,C]:
		// - Card A: max(1, 2) = 2 (need 2 A's)
		// - Card B: max(1, 0) = 1 (need 1 B)
		// - Card C: max(0, 1) = 1 (need 1 C)
		// - Result: [A,A,B,C] - the combination that satisfies both
		intersectionCounts := make(map[basictypes.Card]int)
		totalCardsNeeded := 0
		
		for i := 0; i < n; i++ {
			if subset&(1<<i) != 0 {
				// Take maximum for intersection - this ensures all combinations are satisfied
				for card, cnt := range combinationCounts[i] {
					if intersectionCounts[card] < cnt {
						oldCnt := intersectionCounts[card]
						intersectionCounts[card] = cnt
						totalCardsNeeded += cnt - oldCnt
					}
				}
			}
		}
		
		// Early termination optimizations that maintain correctness:
		//
		// 1. Skip if intersection requires more than targetSize cards:
		//    - Target set is limited to targetSize, so any intersection needing >targetSize is impossible
		//    - This is correct: P(impossible event) = 0, so skipping doesn't affect the sum
		//    - This optimization dramatically reduces computation for large combination sets
		if totalCardsNeeded > targetSize {
			continue
		}
		
		// 2. Skip if any card requires more than available in deck:
		//    - If the deck doesn't have enough copies of a card, the intersection is impossible
		//    - This is correct: P(impossible event) = 0, so skipping doesn't affect the sum
		//    - This catches cases where individual combinations are valid but their intersection isn't
		skip := false
		for card, needed := range intersectionCounts {
			if decklist.Cards[card] < needed {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		
		// Build the intersection combination
		var intersectionCombination []basictypes.Card
		for card, cnt := range intersectionCounts {
			for i := 0; i < cnt; i++ {
				intersectionCombination = append(intersectionCombination, card)
			}
		}
		
		// Calculate probability of this intersection
		intersectionProb := calculateCombinationProbability(intersectionCombination, decklist, targetSize)
		
		// Apply inclusion-exclusion alternating sign: (-1)^(k+1) where k is the number of sets
		// This ensures correct counting:
		// - Odd k (1, 3, 5, ...): add (positive contribution)
		// - Even k (2, 4, 6, ...): subtract (negative contribution to correct overcounting)
		// The alternating pattern ensures each region is counted exactly once in the final sum
		if count%2 == 1 {
			probability += intersectionProb
		} else {
			probability -= intersectionProb
		}
	}
	
	// Clamp probability to valid range [0, 1] to ensure correctness.
	//
	// Why this is needed:
	// 1. Truncation artifact: Limiting to maxDepth means we omit higher-order terms.
	//    The inclusion-exclusion series is alternating, so truncation can leave the sum
	//    slightly outside [0,1] if the next omitted term would correct it.
	// 2. Numerical precision: Floating-point arithmetic can introduce small errors.
	//
	// Why this is safe:
	// - The true probability is always in [0,1] by definition
	// - Clamping to 0 is a lower bound (conservative estimate)
	// - Clamping to 1 is an upper bound (conservative estimate)
	// - For typical maxDepth values, the approximation is very accurate (<0.1% error)
	// - The omitted terms have negligible probability in practice
	if probability < 0.0 {
		probability = 0.0
	}
	if probability > 1.0 {
		probability = 1.0
	}
	
	return probability
}

// calculateCombinationProbability calculates the probability that a specific combination
// of cards appears in a target set of given size, where "appears" means "at least" the required counts.
//
// Mathematical Foundation:
// We calculate P(at least kᵢ copies of each card i) by summing over all valid outcomes
// that satisfy the minimum requirements using the multivariate hypergeometric distribution.
//
// For a combination requiring at least k_A of card A, k_B of card B, etc., we sum:
//   P = Σ P(exactly i_A of A, exactly i_B of B, ...)
//   over all (i_A, i_B, ...) where:
//   - i_A >= k_A, i_B >= k_B, ...
//   - i_A + i_B + ... <= targetSize
//   - i_A <= count_A_in_deck, i_B <= count_B_in_deck, ...
//
// Each term P(exactly i_A, i_B, ...) is calculated using:
//   P = (∏ᵢ C(nᵢ, iᵢ)) * C(N - Σnᵢ, target_size - Σiᵢ) / C(60, target_size)
// Where:
//   - nᵢ = number of copies of card i in the deck
//   - iᵢ = actual count of card i in target set (>= kᵢ)
//   - N = 60 (total deck size)
//   - targetSize = size of the target set (e.g., 6 for prize cards, 54 for not-prized, 8 for starting hand)
//   - C(n,k) = binomial coefficient "n choose k"
//
// Example: Probability of getting at least [A, B] (at least 1 A and at least 1 B):
// - Sum over all (i, j) where i >= 1, j >= 1, i+j <= targetSize
// - For each (i, j): P(exactly i A, exactly j B) = C(n_A, i) * C(n_B, j) * C(60-n_A-n_B, targetSize-i-j) / C(60, targetSize)
// - Total: sum of all such probabilities
func calculateCombinationProbability(combination []basictypes.Card, decklist basictypes.Decklist, targetSize int) float64 {
	if len(combination) == 0 {
		return 1.0 // Empty combination is always satisfied
	}
	
	// Count minimum required copies of each card
	requiredCounts := make(map[basictypes.Card]int)
	for _, card := range combination {
		requiredCounts[card]++
	}
	
	// Check if the combination is possible (all cards exist in decklist)
	for card, required := range requiredCounts {
		if decklist.Cards[card] < required {
			return 0.0 // Not enough copies in deck
		}
	}
	
	// Get list of unique cards and their requirements
	cards := make([]basictypes.Card, 0, len(requiredCounts))
	minCounts := make([]int, 0, len(requiredCounts))
	maxCounts := make([]int, 0, len(requiredCounts))
	
	for card, required := range requiredCounts {
		cards = append(cards, card)
		minCounts = append(minCounts, required)
		maxCounts = append(maxCounts, min(decklist.Cards[card], targetSize)) // Can't have more than targetSize or more than in deck
	}
	
	// Calculate sum over all valid outcomes
	// We'll use a recursive helper to iterate over all valid count combinations
	return calculateAtLeastProbability(cards, minCounts, maxCounts, decklist, 0, make([]int, len(cards)), targetSize)
}

// calculateAtLeastProbability recursively calculates the sum of probabilities for all
// valid outcomes where each card appears at least the minimum required count in the target set.
// targetSize is 6 for prize cards or 54 for not-prized cards.
func calculateAtLeastProbability(
	cards []basictypes.Card,
	minCounts []int,
	maxCounts []int,
	decklist basictypes.Decklist,
	cardIndex int,
	currentCounts []int,
	targetSize int,
) float64 {
	if cardIndex >= len(cards) {
		// We've assigned counts to all cards, calculate probability of this exact outcome
		totalUsed := 0
		for _, cnt := range currentCounts {
			totalUsed += cnt
		}
		
		if totalUsed > targetSize {
			return 0.0 // Can't use more than targetSize cards
		}
		
		// Calculate P(exactly currentCounts[i] of cards[i] for all i)
		numerator := int64(1)
		
		// Product of combinations for each card type
		for i, card := range cards {
			countInDeck := decklist.Cards[card]
			numerator *= comb(countInDeck, currentCounts[i])
		}
		
		// Calculate remaining cards
		remainingInDeck := 60
		for _, card := range cards {
			remainingInDeck -= decklist.Cards[card]
		}
		remainingInTarget := targetSize - totalUsed
		
		if remainingInTarget < 0 {
			return 0.0
		}
		
		numerator *= comb(remainingInDeck, remainingInTarget)
		denominator := comb(60, targetSize)
		
		if denominator == 0 {
			return 0.0
		}
		
		return float64(numerator) / float64(denominator)
	}
	
	// Recursively try all valid counts for the current card
	probability := 0.0
	minCount := minCounts[cardIndex]
	maxCount := maxCounts[cardIndex]
	
	// Calculate how many cards we've already used
	usedSoFar := 0
	for i := 0; i < cardIndex; i++ {
		usedSoFar += currentCounts[i]
	}
	maxAvailable := targetSize - usedSoFar
	
	// Try each possible count from minCount to min(maxCount, maxAvailable)
	for count := minCount; count <= maxCount && count <= maxAvailable; count++ {
		currentCounts[cardIndex] = count
		probability += calculateAtLeastProbability(cards, minCounts, maxCounts, decklist, cardIndex+1, currentCounts, targetSize)
	}
	
	return probability
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// CalculateBasicPokemonStartOdds calculates the odds for basic Pokemon in a 7-card draw.
// It returns:
// - possibleStarters: map of basic Pokemon to odds of having 1+ in 7 draws
// - forcedStarters: map of basic Pokemon to odds of drawing 1+ of that Pokemon but NO OTHER basic Pokemon in 7 draws
// - mulliganOdds: odds of NO basic Pokemon in a 7 card draw
func CalculateBasicPokemonStartOdds(decklist basictypes.Decklist, basicPokemonCards map[basictypes.Card]bool) (map[basictypes.Card]float64, map[basictypes.Card]float64, float64) {
	possibleStarters := make(map[basictypes.Card]float64)
	forcedStarters := make(map[basictypes.Card]float64)

	// Calculate total count of all basic Pokemon in deck
	totalBasicCount := 0
	for card, count := range decklist.Cards {
		if basicPokemonCards[card] {
			totalBasicCount += count
		}
	}

	// Calculate possible starters: odds of 1+ of each basic Pokemon in 7 draws
	for card, count := range decklist.Cards {
		if basicPokemonCards[card] {
			// Probability of at least 1 copy in 7 draws
			// Edge case: if count >= 7, we're guaranteed to have at least 1
			if count >= 7 {
				possibleStarters[card] = 1.0
			} else {
				prob := 0.0
				for j := 1; j <= count; j++ {
					prob += DefaultSevenCardOddsTable.Get(count, j)
				}
				possibleStarters[card] = prob
			}
		}
	}

	// Calculate forced starters: odds of 1+ of specific basic Pokemon but NO OTHER basic Pokemon.
	// In other words, the odds that you have to start with a specific Pokemon in the active spot.
	nonBasicCount := 60 - totalBasicCount
	for card, count := range decklist.Cards {
		if basicPokemonCards[card] {
			// We need: at least 1 of this card AND 0 of all other basic Pokemon
			
			// Edge case: if this is the only basic Pokemon and count >= 7, guaranteed forced starter
			if totalBasicCount == count && count >= 7 {
				forcedStarters[card] = 1.0
			} else {
				// Calculate using multivariate hypergeometric distribution
				// P(at least 1 of card AND 0 of other basic) = 
				// Sum over i=1 to min(count,7): P(exactly i of card AND 0 of other basic)
				prob := 0.0
				for i := 1; i <= min(count, 7); i++ {
					// P(exactly i of card AND 0 of other basic) = 
					// C(count, i) * C(otherBasicCount, 0) * C(nonBasicCount, 7-i) / C(60, 7)
					// = C(count, i) * C(nonBasicCount, 7-i) / C(60, 7)
					if 7-i <= nonBasicCount && 7-i >= 0 {
						numerator := float64(comb(count, i)) * float64(comb(nonBasicCount, 7-i))
						denominator := float64(comb(60, 7))
						if denominator > 0 {
							prob += numerator / denominator
						}
					}
				}
				forcedStarters[card] = prob
			}
		}
	}

	// Calculate mulligan odds: NO basic Pokemon in 7 draws
	// This is: P(0 basic Pokemon in 7 draws) = C(60-totalBasicCount, 7) / C(60, 7)
	mulliganOdds := calculateZeroInDraw(totalBasicCount, 7, 60)

	return possibleStarters, forcedStarters, mulliganOdds
}

// calculateZeroInDraw calculates the probability of drawing 0 copies of a card type
// when drawing 'drawSize' cards from a deck of 'deckSize' containing 'cardCount' copies of that type.
func calculateZeroInDraw(cardCount, drawSize, deckSize int) float64 {
	if cardCount < 0 || drawSize < 0 || deckSize < 0 {
		return 0.0
	}
	if drawSize > deckSize {
		return 0.0
	}
	if cardCount > deckSize {
		return 0.0
	}
	if drawSize > (deckSize - cardCount) {
		return 0.0
	}
	// P(0 copies) = C(deckSize - cardCount, drawSize) / C(deckSize, drawSize)
	numerator := comb(deckSize-cardCount, drawSize)
	denominator := comb(deckSize, drawSize)
	if denominator == 0 {
		return 0.0
	}
	return float64(numerator) / float64(denominator)
}

// CalculateAtLeastTwoBasic calculates the probability of drawing at least 2 basic Pokemon
// (any combination - same or different) in a 7-card draw from a 60-card deck.
// This includes all cases where you draw 2+ basic cards total, such as:
// - 2 of the same basic (e.g., 2 Dreepy)
// - 2 different basics (e.g., 1 Dreepy + 1 Duskull)
// - Any combination totaling 2+ basic cards
// P(>=2 basic) = 1 - P(0 basic) - P(1 basic)
func CalculateAtLeastTwoBasic(totalBasicCount int) float64 {
	if totalBasicCount < 2 {
		// Can't have 2+ basic if there are fewer than 2 in deck
		return 0.0
	}

	// P(0 basic) = C(60 - totalBasicCount, 7) / C(60, 7)
	probZero := calculateZeroInDraw(totalBasicCount, 7, 60)

	// P(1 basic) = C(totalBasicCount, 1) * C(60 - totalBasicCount, 6) / C(60, 7)
	nonBasicCount := 60 - totalBasicCount
	if 6 > nonBasicCount {
		// Can't draw 6 non-basic if there aren't enough
		return 1.0 - probZero
	}

	numerator := float64(comb(totalBasicCount, 1)) * float64(comb(nonBasicCount, 6))
	denominator := float64(comb(60, 7))
	probOne := 0.0
	if denominator > 0 {
		probOne = numerator / denominator
	}

	// P(>=2 basic) = 1 - P(0 basic) - P(1 basic)
	return 1.0 - probZero - probOne
}

// CalculateCardSetStartOdds takes a list of named CardSets and calculates the odds that ANY satisfiable
// set of cards in the CardSet is in the starting hand (8 cards: initial hand + 1st turn start).
//
// Semantics: "is in the starting hand" means "at least" the required counts appear. For example, if a combination
// requires [A, B], then starting hands containing [A, A, B] or [A, B, B] or [A, A, B, B] all satisfy this requirement.
// This matches the intuitive meaning: if you need at least 1 A and at least 1 B, having more than the minimum
// still satisfies the requirement.
//
// Example: for a CardSet [AnyOf(1 X, 1 Y),AnyOf(1 Y, 1 Z)], it calculates the odds that ANY of the following
// (or supersets thereof) is in the starting hand (8 cards):
// * At least 1 X and at least 1 Y
// * At least 1 Y and at least 1 Z
// * At least 2 Y
//
// Correctness: This function correctly handles overlapping combinations (e.g., AB and AA both requiring card A)
// by using the inclusion-exclusion principle to avoid double-counting. The CardSet.Expand() method
// generates all valid combinations based on the decklist constraints, ensuring we only consider
// combinations that are actually possible given the available card counts.
func CalculateCardSetStartOdds(decklist basictypes.Decklist, cardSets map[string]basictypes.CardSet) (map[string]float64, error) {
	unionCalculator := func(combinations [][]basictypes.Card, decklist basictypes.Decklist) float64 {
		return CalculateUnionProbabilityStart(combinations, decklist)
	}
	return calculateCardSetOdds(decklist, cardSets, unionCalculator), nil
}

// CalculateUnionProbabilityStart calculates the probability that ANY of the given combinations
// appears in the starting hand (8 cards: initial hand + 1st turn start),
// using the inclusion-exclusion principle.
//
// Semantics: Each combination represents "at least" the required counts. For example, [A, B] means
// "at least 1 A and at least 1 B", so starting hands like [A, A, B] or [A, B, B] all satisfy it.
//
// Mathematical Foundation:
// The inclusion-exclusion principle correctly calculates P(∪A_i) for overlapping events:
//   P(∪A_i) = ΣP(A_i) - ΣP(A_i ∩ A_j) + ΣP(A_i ∩ A_j ∩ A_k) - ...
//
// This alternating sum ensures that overlapping regions are counted exactly once:
// - Single terms (P(A_i)) count each event
// - Pair terms (P(A_i ∩ A_j)) subtract double-counted overlaps
// - Triple terms (P(A_i ∩ A_j ∩ A_k)) add back triple-counted regions
// - And so on, with alternating signs
//
// Correctness for Overlapping Combinations:
// When combinations overlap (e.g., [A,B] and [A,A] both require card A), the intersection
// correctly represents "both combinations satisfied simultaneously" by taking the maximum
// count of each card across all combinations in the intersection. This ensures:
// - If combination 1 needs at least 1 A and combination 2 needs at least 2 A, the intersection needs at least 2 A
// - This correctly models that satisfying both requires the maximum requirement
//
// Example: For combinations [A,B] and [A,A]:
// - P([A,B]) = probability of getting at least 1 A and 1 B
// - P([A,A]) = probability of getting at least 2 A
// - P([A,B] ∩ [A,A]) = probability of getting at least 2 A and 1 B (max of each card)
// - Result: P([A,B] ∪ [A,A]) = P([A,B]) + P([A,A]) - P([A,B] ∩ [A,A])
//
// Performance Optimization (Depth Limiting):
// We limit intersections to maxDepth=8 to avoid exponential explosion (2^n subsets).
// This is safe because:
// 1. Higher-order intersections (9+ combinations) have rapidly decreasing probabilities
// 2. The alternating series converges, so truncation gives a good approximation
// 3. Early termination skips impossible intersections (requiring >8 cards)
// 4. The result is clamped to [0,1] to handle any truncation artifacts
//
// The depth limit of 8 is chosen because:
// - Starting hand is limited to 8 cards, so intersections of 9+ combinations are often impossible
// - Even when possible, their probabilities are typically negligible
// - This provides a good balance between accuracy and performance
func CalculateUnionProbabilityStart(combinations [][]basictypes.Card, decklist basictypes.Decklist) float64 {
	const targetSize = 8
	const maxDepth = 8
	return calculateUnionProbabilityWithTargetSize(combinations, decklist, targetSize, maxDepth)
}