import { useState } from 'react';
import { calculateDrawSupporterOdds } from '../api/client';
import type { DrawSupporterOddsResponse } from '../types/api';

function formatPercent(p: number): string {
  if (!Number.isFinite(p)) return '—';
  return `${(p * 100).toFixed(2)}%`;
}

function parseIntOrNull(value: string): number | null {
  if (!value.trim()) return null;
  const parsed = Number.parseInt(value, 10);
  if (Number.isNaN(parsed)) return null;
  return parsed;
}

function normalizeInputValue(value: string, min: number, max: number): string {
  const parsed = parseIntOrNull(value);
  if (parsed === null) return '';
  const clamped = Math.min(max, Math.max(min, parsed));
  return String(clamped);
}

export default function DrawSupporters() {
  const [deckSizeInput, setDeckSizeInput] = useState('46');
  const [knownBottomInput, setKnownBottomInput] = useState('0');
  const [handSizeInput, setHandSizeInput] = useState('7');
  const [prizeCardsInput, setPrizeCardsInput] = useState('6');

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [response, setResponse] = useState<DrawSupporterOddsResponse | null>(null);
  const [showPairOdds, setShowPairOdds] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setResponse(null);

    try {
      const deckSize = parseIntOrNull(deckSizeInput);
      const knownBottom = parseIntOrNull(knownBottomInput) ?? 0;
      const handSize = parseIntOrNull(handSizeInput);
      const prizeCards = parseIntOrNull(prizeCardsInput);

      if (deckSize === null || handSize === null || prizeCards === null) {
        throw new Error('Please enter deck, hand, and prize card counts.');
      }
      if (deckSize < 1 || deckSize > 60) {
        throw new Error('Deck size must be between 1 and 60.');
      }
      if (knownBottom < 0 || knownBottom > deckSize - 1) {
        throw new Error('Known bottom must be between 0 and deck size - 1.');
      }
      if (handSize < 1 || handSize > 60) {
        throw new Error('Hand size must be between 1 and 60.');
      }
      if (prizeCards < 1 || prizeCards > 6) {
        throw new Error('Prize cards must be between 1 and 6.');
      }
      if (deckSize + handSize + prizeCards > 59) {
        throw new Error(
          'Constraint violated: deck + hand + prizes must be ≤ 59 (at least 1 card must be in play).'
        );
      }

      const result = await calculateDrawSupporterOdds({
        deckSize,
        knownBottom,
        handSize,
        prizeCards,
      });
      setResponse(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An unknown error occurred');
    } finally {
      setLoading(false);
    }
  };

  const supporterOrder = ["Iono", "Professor's Research", "Lillie's Determination"] as const;
  const showBottomOdds = Boolean(response?.bottomOdds && Object.keys(response.bottomOdds).length > 0);

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8">
        <h1 className="text-3xl font-bold text-gray-900 mb-4">Draw Supporters</h1>
        <p className="text-gray-600 mb-8">
          Calculate abstract odds of drawing at least one copy of a card (given 1–4 copies remaining) off common draw supporters.
        </p>
        <p className="text-gray-600 mb-8">
          You can think about equivalent cards in a combined fashion, e.g. "draw an energy, given 1 Fire and 2 Psychic energy" as drawing from 3 copies.
        </p>

        <form onSubmit={handleSubmit} className="space-y-6 mb-8">
          <div className="bg-white p-6 rounded-lg shadow space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label htmlFor="deckSize" className="block text-sm font-medium text-gray-700 mb-1">
                  Number of cards in deck
                </label>
                <input
                  id="deckSize"
                  type="number"
                  min={1}
                  max={60}
                  value={deckSizeInput}
                  onChange={(e) => setDeckSizeInput(e.target.value)}
                  onBlur={() => setDeckSizeInput((value) => normalizeInputValue(value, 1, 60))}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 focus:ring-blue-500 focus:border-blue-500"
                />
              </div>

              <div>
                <label htmlFor="knownBottom" className="block text-sm font-medium text-gray-700 mb-1">
                  Known cards at bottom of deck
                </label>
                <input
                  id="knownBottom"
                  type="number"
                  min={0}
                  max={59}
                  value={knownBottomInput}
                  onChange={(e) => setKnownBottomInput(e.target.value)}
                  onBlur={() => setKnownBottomInput((value) => normalizeInputValue(value, 0, 59))}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 focus:ring-blue-500 focus:border-blue-500"
                />
              </div>

              <div>
                <label htmlFor="handSize" className="block text-sm font-medium text-gray-700 mb-1">
                  Number of cards in hand
                </label>
                <input
                  id="handSize"
                  type="number"
                  min={1}
                  max={60}
                  value={handSizeInput}
                  onChange={(e) => setHandSizeInput(e.target.value)}
                  onBlur={() => setHandSizeInput((value) => normalizeInputValue(value, 1, 60))}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 focus:ring-blue-500 focus:border-blue-500"
                />
              </div>

              <div>
                <label htmlFor="prizeCards" className="block text-sm font-medium text-gray-700 mb-1">
                  Prize cards left
                </label>
                <input
                  id="prizeCards"
                  type="number"
                  min={1}
                  max={6}
                  value={prizeCardsInput}
                  onChange={(e) => setPrizeCardsInput(e.target.value)}
                  onBlur={() => setPrizeCardsInput((value) => normalizeInputValue(value, 1, 6))}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 focus:ring-blue-500 focus:border-blue-500"
                />
              </div>
            </div>
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full md:w-auto bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed text-white font-bold py-3 px-8 rounded-lg text-lg transition-colors shadow-lg"
          >
            {loading ? 'Calculating...' : 'Calculate'}
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
            <p className="mt-4 text-gray-600">Calculating draw odds...</p>
          </div>
        )}

        {response && !loading && (
          <div className="bg-white p-6 rounded-lg shadow">
            <h2 className="text-xl font-bold text-gray-900 mb-4">Odds of drawing 1+ copies</h2>
            <p className="text-sm text-gray-600 mb-4">
              For Iono and Professor&apos;s Research, draws are modeled from the{' '}
              <span className="font-medium">top, randomized pool</span> of the deck (excluding known bottom cards).
              Lillie&apos;s Determination is modeled from a <span className="font-medium">shuffled pool</span> (deck + hand).
            </p>
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Supporter
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      1 copy
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      2 copies
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      3 copies
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      4 copies
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {supporterOrder.map((name) => {
                    const row = response.odds[name] ?? [];
                    return (
                      <tr key={name}>
                        <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                          <div>{name}</div>
                          {response?.drawCounts?.[name] !== undefined && (
                            <div className="text-xs text-gray-500">
                              {(() => {
                                const requested = response.drawCounts[name];
                                const effective = response.effectiveDrawCounts?.[name];
                                if (effective === undefined) return `Draw ${requested}`;
                                if (effective === requested) return `Draw ${effective}`;
                                return `Draw ${effective} (requested ${requested})`;
                              })()}
                            </div>
                          )}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-700">
                          {formatPercent(row[0])}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-700">
                          {formatPercent(row[1])}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-700">
                          {formatPercent(row[2])}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-700">
                          {formatPercent(row[3])}
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>

            <div className="mt-8 border-t border-gray-200 pt-6">
              <div className="flex items-start gap-3">
                <input
                  id="showPairOdds"
                  type="checkbox"
                  checked={showPairOdds}
                  onChange={(e) => setShowPairOdds(e.target.checked)}
                  className="mt-1 w-5 h-5 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                />
                <div>
                  <label htmlFor="showPairOdds" className="text-sm font-medium text-gray-900 cursor-pointer">
                    Show pair-odds tables
                  </label>
                  <p className="text-sm text-gray-600">
                    Pair odds means drawing <span className="font-medium">at least 1 of card A and at least 1 of card B</span> in the supporter&apos;s draw.
                    For Iono and Professor&apos;s Research, this is the top-of-deck pool only (known bottom cards excluded). For Lillie&apos;s Determination,
                    this uses the shuffled pool (deck + hand). Pair odds never model bottom cards.
                  </p>
                </div>
              </div>

              {showPairOdds && (
                <div className="mt-6 space-y-8">
                  {supporterOrder.map((name) => {
                    const pair = response.pairOdds?.[name] ?? [];
                    const drawNote = (() => {
                      const requested = response?.drawCounts?.[name];
                      if (requested === undefined) return undefined;
                      const effective = response?.effectiveDrawCounts?.[name];
                      if (effective === undefined || effective === requested) return `Draw ${requested}`;
                      return `Draw ${effective} (requested ${requested})`;
                    })();

                    return (
                      <div key={name}>
                        <div className="flex items-baseline justify-between">
                          <h3 className="text-lg font-bold text-gray-900">{name}</h3>
                          {drawNote && <span className="text-sm text-gray-500">{drawNote}</span>}
                        </div>
                        <div className="overflow-x-auto mt-3">
                          <table className="min-w-full divide-y divide-gray-200">
                            <thead className="bg-gray-50">
                              <tr>
                                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                  Card A copies \ Card B copies
                                </th>
                                {[1, 2, 3, 4].map((b) => (
                                  <th
                                    key={b}
                                    className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                                  >
                                    {b}
                                  </th>
                                ))}
                              </tr>
                            </thead>
                            <tbody className="bg-white divide-y divide-gray-200">
                              {[1, 2, 3, 4].map((a) => (
                                <tr key={a}>
                                  <td className="px-4 py-2 whitespace-nowrap text-sm font-medium text-gray-900">
                                    {a}
                                  </td>
                                  {[1, 2, 3, 4].map((b) => (
                                    <td key={b} className="px-4 py-2 whitespace-nowrap text-sm text-gray-700">
                                      {formatPercent(pair?.[a - 1]?.[b - 1])}
                                    </td>
                                  ))}
                                </tr>
                              ))}
                            </tbody>
                          </table>
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
            </div>

            {showBottomOdds && (
              <div className="mt-8 border-t border-gray-200 pt-6">
                <h3 className="text-lg font-bold text-gray-900">Bottom cards (drawn into)</h3>
                <p className="text-sm text-gray-600 mb-4">
                  These tables show odds for the <span className="font-medium">known bottom cards</span> when the draw goes past the top of deck.
                  Bottom cards are treated as unordered (no layering).
                </p>
                <div className="space-y-6">
                  {supporterOrder.map((name) => {
                    const bottom = response.bottomOdds?.[name];
                    if (!bottom) return null;
                    const bottomDraw = response.bottomDrawCounts?.[name];
                    return (
                      <div key={`${name}-bottom`}>
                        <div className="flex items-baseline justify-between">
                          <h4 className="text-base font-bold text-gray-900">{name}</h4>
                          {bottomDraw !== undefined && (
                            <span className="text-sm text-gray-500">Bottom draw {bottomDraw}</span>
                          )}
                        </div>
                        <div className="overflow-x-auto mt-3">
                          <table className="min-w-full divide-y divide-gray-200">
                            <thead className="bg-gray-50">
                              <tr>
                                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                  1 copy
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                  2 copies
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                  3 copies
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                  4 copies
                                </th>
                              </tr>
                            </thead>
                            <tbody className="bg-white divide-y divide-gray-200">
                              <tr>
                                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-700">
                                  {formatPercent(bottom[0])}
                                </td>
                                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-700">
                                  {formatPercent(bottom[1])}
                                </td>
                                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-700">
                                  {formatPercent(bottom[2])}
                                </td>
                                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-700">
                                  {formatPercent(bottom[3])}
                                </td>
                              </tr>
                            </tbody>
                          </table>
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

