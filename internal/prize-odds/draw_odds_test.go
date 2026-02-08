package prizeodds

import (
	"math"
	"testing"
)

func TestCalculateDrawOdds_SingleCopy(t *testing.T) {
	// With 1 copy in pool, P(draw it) = draw/pool.
	got := CalculateDrawOdds(60, 7, 1)
	want := 7.0 / 60.0
	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("expected %.15f, got %.15f", want, got)
	}
}

func TestCalculateDrawOdds_MultipleCopies(t *testing.T) {
	got := CalculateDrawOdds(60, 7, 4)
	want := 1.0 - float64(comb(56, 7))/float64(comb(60, 7))
	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("expected %.15f, got %.15f", want, got)
	}
}

func TestCalculateDrawOdds_DrawAllCards(t *testing.T) {
	got := CalculateDrawOdds(10, 10, 1)
	if got != 1.0 {
		t.Fatalf("expected 1.0, got %.15f", got)
	}
}

func TestCalculateDrawOdds_InvalidInputs(t *testing.T) {
	cases := []struct {
		pool, draw, target int
	}{
		{0, 7, 1},
		{60, 0, 1},
		{60, 7, 0},
		{60, 7, -1},
		{60, 61, 1},
		{60, 7, 61},
	}

	for _, tc := range cases {
		got := CalculateDrawOdds(tc.pool, tc.draw, tc.target)
		if got != 0.0 {
			t.Fatalf("expected 0.0 for pool=%d draw=%d target=%d, got %.15f", tc.pool, tc.draw, tc.target, got)
		}
	}
}

func TestCalculateDrawPairOdds_SingleCopyEach(t *testing.T) {
	// With 1 copy of each card, probability of drawing both specific cards in drawCount draws is:
	// P = C(pool-2, draw-2) / C(pool, draw) = draw*(draw-1)/(pool*(pool-1))
	pool := 46
	draw := 6
	got := CalculateDrawPairOdds(pool, draw, 1, 1)
	want := float64(draw*(draw-1)) / float64(pool*(pool-1))
	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("expected %.15f, got %.15f", want, got)
	}
}

func TestCalculateDrawPairOdds_InvalidInputs(t *testing.T) {
	cases := []struct {
		pool, draw, a, b int
	}{
		{0, 6, 1, 1},
		{46, 0, 1, 1},
		{46, 6, 0, 1},
		{46, 6, 1, 0},
		{46, 6, 47, 1},
		{46, 6, 1, 47},
		{46, 6, 4, 43}, // a+b > pool
		{46, 60, 1, 1}, // draw > pool
	}

	for _, tc := range cases {
		got := CalculateDrawPairOdds(tc.pool, tc.draw, tc.a, tc.b)
		if got != 0.0 {
			t.Fatalf("expected 0.0 for pool=%d draw=%d a=%d b=%d, got %.15f", tc.pool, tc.draw, tc.a, tc.b, got)
		}
	}
}

func TestCalculateDrawPairOdds_GuaranteedWhenDrawingAllCards(t *testing.T) {
	// If we draw the entire pool and both cards exist, we must see both.
	got := CalculateDrawPairOdds(10, 10, 1, 1)
	if got != 1.0 {
		t.Fatalf("expected 1.0, got %.15f", got)
	}
}

func TestCalculateDrawPairOdds_ImpossibleWhenDrawingTooFew(t *testing.T) {
	// Can't draw at least 1 of A and 1 of B if drawing only 1 card.
	got := CalculateDrawPairOdds(10, 1, 1, 1)
	if got != 0.0 {
		t.Fatalf("expected 0.0, got %.15f", got)
	}
}

func TestCalculateDrawPairOdds_Symmetry(t *testing.T) {
	// Swapping counts should not change the probability.
	got1 := CalculateDrawPairOdds(60, 7, 2, 4)
	got2 := CalculateDrawPairOdds(60, 7, 4, 2)
	if math.Abs(got1-got2) > 1e-12 {
		t.Fatalf("expected symmetry, got %.15f vs %.15f", got1, got2)
	}
}

