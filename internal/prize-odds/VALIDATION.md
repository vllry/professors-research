# Validation Approaches for Prize Odds Calculation

## Why TLA+ Doesn't Apply Here

**TLA+ (Temporal Logic of Actions)** is designed for:
- ✅ Concurrent/distributed systems
- ✅ State machines and protocols
- ✅ Temporal properties (liveness, safety)
- ✅ Finding race conditions, deadlocks
- ✅ Algorithm correctness with state transitions

**This code is:**
- ❌ Pure mathematical computation (no concurrency)
- ❌ No state machines or temporal logic
- ❌ Probabilistic calculations, not system behavior
- ❌ Deterministic functions, not protocols

**Conclusion**: TLA+ is not the right tool for validating probabilistic/mathematical correctness.

## Better Validation Approaches

### 1. Property-Based Testing (Recommended)

Property-based testing generates random inputs and verifies mathematical properties hold. This is ideal for probabilistic code.

**Properties to verify:**
- Probabilities are always in [0, 1]
- P(at least k) >= P(exactly k) for any k
- Inclusion-exclusion: P(A ∪ B) = P(A) + P(B) - P(A ∩ B)
- Monotonicity: More restrictive requirements have lower probabilities
- Sum of probabilities for disjoint events equals union probability
- Symmetry: Order of cards in combination doesn't matter

**Tools:**
- **gopter** (Go property-based testing)
- **rapid** (Go property-based testing, simpler API)

### 2. Mathematical Proof Verification

For formal mathematical correctness:
- **Coq/Isabelle/Lean**: Proof assistants that can verify mathematical theorems
- **SymPy/Mathematica**: Symbolic computation to verify formulas match known results
- **Paper proofs**: Manual mathematical proofs of correctness

### 3. Known Results Comparison

Compare against:
- Published probability tables
- Manual calculations for specific cases
- Alternative implementations
- Mathematical identities (e.g., hypergeometric distribution formulas)

### 4. Comprehensive Test Cases

Systematic testing with:
- Edge cases (0 cards, 6 cards, all cards)
- Boundary conditions (exactly 6 copies, exactly 1 copy)
- Known probability values from literature
- Symmetry tests (order independence)

## Recommended Next Steps

1. **Add property-based tests** using `gopter` or `rapid` to verify mathematical properties
2. **Add more known-value tests** comparing against manually calculated probabilities
3. **Add symmetry tests** to ensure order independence
4. **Document mathematical proofs** of key formulas (inclusion-exclusion, hypergeometric)

## Example: Property-Based Test Structure

```go
// Properties to verify:
// 1. All probabilities are in [0, 1]
// 2. P(at least k) >= P(exactly k)
// 3. More cards = lower probability (for same requirement)
// 4. Inclusion-exclusion holds for any two combinations
// 5. Union probability <= sum of individual probabilities
```






