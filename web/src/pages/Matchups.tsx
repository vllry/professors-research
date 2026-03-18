import { useState, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import { fetchTournaments, fetchArchetypes, getMatchupStats } from '../api/client';
import MatchupStatsDisplay from '../components/MatchupStatsDisplay';
import FeatureBadge from '../components/FeatureBadge';
import type {
  TournamentResponse,
  MatchupStatsResponse,
} from '../types/api';

export interface VariantEntry {
  cards: { name: string; count: number }[];
}

export function encodeVariant(v: VariantEntry): string {
  return v.cards
    .filter((c) => c.name.trim())
    .map((c) => `${encodeURIComponent(c.name.trim())}:${c.count}`)
    .join(',');
}

export function parseVariantsFromUrl(strings: string[]): VariantEntry[] {
  return strings
    .map((str) => ({
      cards: str
        .split(',')
        .map((pair) => {
          const colonIdx = pair.lastIndexOf(':');
          if (colonIdx === -1) return { name: decodeURIComponent(pair), count: 1 };
          return {
            name: decodeURIComponent(pair.slice(0, colonIdx)),
            count: parseInt(pair.slice(colonIdx + 1), 10) || 1,
          };
        })
        .filter((c) => c.name.trim()),
    }))
    .filter((v) => v.cards.length > 0);
}

export default function Matchups() {
  const [searchParams, setSearchParams] = useSearchParams();

  // Capture URL-specified tournament IDs at mount before tournaments load
  const [initialTournamentIds] = useState<string[]>(() => searchParams.getAll('t'));

  const [tournaments, setTournaments] = useState<TournamentResponse[]>([]);
  const [tournamentsLoading, setTournamentsLoading] = useState(true);
  const [tournamentsError, setTournamentsError] = useState<string | null>(null);

  const [archetypes, setArchetypes] = useState<string[]>([]);
  const [archetypesLoading, setArchetypesLoading] = useState(true);

  const [selectedTournamentIds, setSelectedTournamentIds] = useState<Set<string>>(new Set());
  const [archetype, setArchetype] = useState<string>(() => searchParams.get('a') ?? '');
  const [variants, setVariants] = useState<VariantEntry[]>(() =>
    parseVariantsFromUrl(searchParams.getAll('v'))
  );
  const [showVariants, setShowVariants] = useState<boolean>(
    () => searchParams.get('sv') === '1'
  );
  const [playerPlacement, setPlayerPlacement] = useState<number>(
    () => parseFloat(searchParams.get('pp') ?? '0') || 0
  );
  const [opponentPlacement, setOpponentPlacement] = useState<number>(
    () => parseFloat(searchParams.get('op') ?? '0') || 0
  );

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [response, setResponse] = useState<MatchupStatsResponse | null>(null);

  useEffect(() => {
    fetchTournaments()
      .then((data) => {
        setTournaments(data);
        if (initialTournamentIds.length > 0) {
          const validIds = new Set(data.map((t) => t.id));
          setSelectedTournamentIds(
            new Set(initialTournamentIds.filter((id) => validIds.has(id)))
          );
        } else {
          setSelectedTournamentIds(new Set(data.map((t) => t.id)));
        }
      })
      .catch((err) => {
        setTournamentsError(
          err instanceof Error ? err.message : 'Failed to load tournaments'
        );
      })
      .finally(() => setTournamentsLoading(false));
  }, []);

  useEffect(() => {
    fetchArchetypes()
      .then(setArchetypes)
      .catch(() => {})
      .finally(() => setArchetypesLoading(false));
  }, []);

  // Sync all form state to URL params so views are shareable
  useEffect(() => {
    if (tournamentsLoading) return;

    const params = new URLSearchParams();

    if (archetype.trim()) params.set('a', archetype.trim());

    // Omit 't' params when all tournaments are selected (keeps URL clean by default)
    if (selectedTournamentIds.size < tournaments.length) {
      for (const id of selectedTournamentIds) {
        params.append('t', id);
      }
    }

    if (showVariants) params.set('sv', '1');

    for (const v of variants) {
      const encoded = encodeVariant(v);
      if (encoded) params.append('v', encoded);
    }

    if (playerPlacement > 0) params.set('pp', String(playerPlacement));
    if (opponentPlacement > 0) params.set('op', String(opponentPlacement));

    setSearchParams(params, { replace: true });
  }, [
    tournamentsLoading,
    archetype,
    selectedTournamentIds,
    tournaments.length,
    showVariants,
    variants,
    playerPlacement,
    opponentPlacement,
  ]);

  const toggleTournament = (id: string) => {
    setSelectedTournamentIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  const toggleAllTournaments = () => {
    if (selectedTournamentIds.size === tournaments.length) {
      setSelectedTournamentIds(new Set());
    } else {
      setSelectedTournamentIds(new Set(tournaments.map((t) => t.id)));
    }
  };

  const addVariant = () => {
    setVariants((prev) => [...prev, { cards: [{ name: '', count: 1 }] }]);
  };

  const removeVariant = (index: number) => {
    setVariants((prev) => prev.filter((_, i) => i !== index));
  };

  const addCardToVariant = (variantIndex: number) => {
    setVariants((prev) =>
      prev.map((v, i) =>
        i === variantIndex
          ? { cards: [...v.cards, { name: '', count: 1 }] }
          : v
      )
    );
  };

  const removeCardFromVariant = (variantIndex: number, cardIndex: number) => {
    setVariants((prev) =>
      prev.map((v, i) =>
        i === variantIndex
          ? { cards: v.cards.filter((_, ci) => ci !== cardIndex) }
          : v
      )
    );
  };

  const updateVariantCard = (
    variantIndex: number,
    cardIndex: number,
    field: 'name' | 'count',
    value: string | number
  ) => {
    setVariants((prev) =>
      prev.map((v, i) =>
        i === variantIndex
          ? {
              cards: v.cards.map((c, ci) =>
                ci === cardIndex ? { ...c, [field]: value } : c
              ),
            }
          : v
      )
    );
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setResponse(null);

    try {
      const variantFilters = variants
        .map((v) => {
          const map: Record<string, number> = {};
          for (const card of v.cards) {
            if (card.name.trim()) {
              map[card.name.trim()] = card.count;
            }
          }
          return map;
        })
        .filter((m) => Object.keys(m).length > 0);

      const result = await getMatchupStats({
        tournamentIds: Array.from(selectedTournamentIds),
        archetype: archetype.trim(),
        variants: variantFilters.length > 0 ? variantFilters : undefined,
        playerPlacement: playerPlacement > 0 ? playerPlacement : undefined,
        opponentPlacement: opponentPlacement > 0 ? opponentPlacement : undefined,
      });

      setResponse(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An unknown error occurred');
    } finally {
      setLoading(false);
    }
  };

  const canSubmit =
    !loading &&
    archetype.trim() !== '' &&
    selectedTournamentIds.size > 0;

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
        <h1 className="flex flex-wrap items-center gap-x-3 gap-y-2 text-3xl font-bold text-gray-900 mb-4">
          <span>Matchup Stats</span>
          <FeatureBadge stage="alpha" />
        </h1>
        <p className="text-gray-600 mb-8">
          View win rates for a deck archetype against the field, across one or
          more tournaments, and compare variants with different card counts.
        </p>

        <form onSubmit={handleSubmit} className="space-y-6 mb-8">
          {/* Tournament selector */}
          <div className="bg-white p-6 rounded-lg shadow">
            <label className="block text-sm font-medium text-gray-700 mb-3">
              Tournaments
            </label>
            {tournamentsLoading && (
              <p className="text-sm text-gray-500">Loading tournaments...</p>
            )}
            {tournamentsError && (
              <p className="text-sm text-red-600">{tournamentsError}</p>
            )}
            {!tournamentsLoading && !tournamentsError && (
              <>
                <button
                  type="button"
                  onClick={toggleAllTournaments}
                  className="text-sm text-blue-600 hover:text-blue-800 mb-2"
                >
                  {selectedTournamentIds.size === tournaments.length
                    ? 'Deselect all'
                    : 'Select all'}
                </button>
                <div className="space-y-2">
                  {tournaments.map((t) => (
                    <label
                      key={t.id}
                      className="flex items-center space-x-3 cursor-pointer"
                    >
                      <input
                        type="checkbox"
                        checked={selectedTournamentIds.has(t.id)}
                        onChange={() => toggleTournament(t.id)}
                        className="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                      />
                      <span className="text-sm text-gray-700">
                        {t.location} {t.year}
                      </span>
                    </label>
                  ))}
                </div>
              </>
            )}
          </div>

          {/* Archetype picker */}
          <div className="bg-white p-6 rounded-lg shadow">
            <label
              htmlFor="archetype"
              className="block text-sm font-medium text-gray-700 mb-2"
            >
              Archetype
            </label>
            {archetypesLoading ? (
              <p className="text-sm text-gray-500">Loading archetypes...</p>
            ) : (
              <select
                id="archetype"
                value={archetype}
                onChange={(e) => setArchetype(e.target.value)}
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:ring-blue-500 focus:border-blue-500 bg-white"
              >
                <option value="">Select an archetype</option>
                {archetypes.map((name) => (
                  <option key={name} value={name}>
                    {name}
                  </option>
                ))}
              </select>
            )}
          </div>

          {/* Variant filters */}
          <div className="bg-white p-6 rounded-lg shadow">
            <button
              type="button"
              onClick={() => setShowVariants(!showVariants)}
              className="flex items-center text-sm font-medium text-gray-700"
            >
              <svg
                className={`w-4 h-4 mr-2 transition-transform ${showVariants ? 'rotate-90' : ''}`}
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9 5l7 7-7 7"
                />
              </svg>
              Variant Filters
            </button>
            {showVariants && (
              <div className="mt-4 space-y-4">
                <p className="text-xs text-gray-500">
                  Decks matching multiple variants are counted in each match.
                </p>
                {variants.map((variant, vi) => (
                  <div
                    key={vi}
                    className="border border-gray-200 rounded-md p-4"
                  >
                    <div className="flex items-center justify-between mb-3">
                      <span className="text-sm font-medium text-gray-600">
                        Variant {vi}
                      </span>
                      <button
                        type="button"
                        onClick={() => removeVariant(vi)}
                        className="text-sm text-red-600 hover:text-red-800"
                      >
                        Remove
                      </button>
                    </div>
                    {variant.cards.map((card, ci) => (
                      <div key={ci} className="flex items-center space-x-2 mb-2">
                        <input
                          type="text"
                          value={card.name}
                          onChange={(e) =>
                            updateVariantCard(vi, ci, 'name', e.target.value)
                          }
                          placeholder="Card name"
                          className="flex-1 border border-gray-300 rounded-md px-3 py-1.5 text-sm focus:ring-blue-500 focus:border-blue-500"
                        />
                        <input
                          type="number"
                          min={1}
                          value={card.count}
                          onChange={(e) =>
                            updateVariantCard(
                              vi,
                              ci,
                              'count',
                              parseInt(e.target.value, 10) || 1
                            )
                          }
                          className="w-16 border border-gray-300 rounded-md px-2 py-1.5 text-sm text-center focus:ring-blue-500 focus:border-blue-500"
                        />
                        <button
                          type="button"
                          onClick={() => removeCardFromVariant(vi, ci)}
                          className="text-gray-400 hover:text-red-600 text-sm"
                          aria-label="Remove card"
                        >
                          &times;
                        </button>
                      </div>
                    ))}
                    <button
                      type="button"
                      onClick={() => addCardToVariant(vi)}
                      className="text-sm text-blue-600 hover:text-blue-800"
                    >
                      + Add card
                    </button>
                  </div>
                ))}
                <button
                  type="button"
                  onClick={addVariant}
                  className="text-sm text-blue-600 hover:text-blue-800 font-medium"
                >
                  + Add variant
                </button>
              </div>
            )}
          </div>

          {/* Placement filters */}
          <div className="bg-white p-6 rounded-lg shadow">
            <label className="block text-sm font-medium text-gray-700 mb-3">
              Placement Filters
            </label>
            <p>
              Only count matches where the player and opponent finished in the top X% and Y% of players, respectively.
            </p>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <label
                  htmlFor="playerPlacement"
                  className="block text-xs text-gray-500 mb-1"
                >
                  Player Placement (top %)
                </label>
                <input
                  id="playerPlacement"
                  type="number"
                  min={0}
                  max={100}
                  value={playerPlacement}
                  onChange={(e) =>
                    setPlayerPlacement(parseFloat(e.target.value) || 0)
                  }
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:ring-blue-500 focus:border-blue-500"
                />
              </div>
              <div>
                <label
                  htmlFor="opponentPlacement"
                  className="block text-xs text-gray-500 mb-1"
                >
                  Opponent Placement (top %)
                </label>
                <input
                  id="opponentPlacement"
                  type="number"
                  min={0}
                  max={100}
                  value={opponentPlacement}
                  onChange={(e) =>
                    setOpponentPlacement(parseFloat(e.target.value) || 0)
                  }
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:ring-blue-500 focus:border-blue-500"
                />
              </div>
            </div>
          </div>

          <button
            type="submit"
            disabled={!canSubmit}
            className="w-full md:w-auto bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed text-white font-bold py-3 px-8 rounded-lg text-lg transition-colors shadow-lg"
          >
            {loading ? 'Loading...' : 'Get Matchups'}
          </button>
        </form>

        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg mb-6">
            <p className="font-medium">Error</p>
            <p className="text-sm">{error}</p>
          </div>
        )}

        {loading && (
          <div className="bg-white p-8 rounded-lg shadow text-center">
            <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
            <p className="mt-4 text-gray-600">Fetching matchup data...</p>
          </div>
        )}

        {response && !loading && (
          <div className="bg-white p-6 rounded-lg shadow">
            <MatchupStatsDisplay response={response} />
          </div>
        )}
      </div>
    </div>
  );
}
