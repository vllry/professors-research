import React from 'react';
import type { MatchupStatsResponse } from '../types/api';

interface Props {
  response: MatchupStatsResponse;
}

function winRateColor(rate: number): string {
  if (rate >= 0.55) return 'text-green-700 bg-green-50';
  if (rate <= 0.45) return 'text-red-700 bg-red-50';
  return 'text-gray-700';
}

function formatPercent(rate: number): string {
  return `${(rate * 100).toFixed(1)}%`;
}

function variantLabel(key: string, cardCounts: Record<string, number>): string {
  if (key === 'other') return 'Other';
  const cards = Object.entries(cardCounts);
  if (cards.length === 0) return `Variant ${key}`;
  return cards.map(([name, count]) => `${name}: ${count}`).join(', ');
}

export default function MatchupStatsDisplay({ response }: Props) {
  const variantKeys = Object.keys(response.matchups).sort((a, b) => {
    if (a === 'other') return 1;
    if (b === 'other') return -1;
    return Number(a) - Number(b);
  });

  const showCounts = variantKeys.length < 3;

  const allOpponents = Array.from(
    variantKeys.reduce((set, key) => {
      for (const opp of Object.keys(response.matchups[key].matchups)) {
        set.add(opp);
      }
      return set;
    }, new Set<string>())
  ).sort((a, b) => a.localeCompare(b));

  const totalArchetypeDecks = Object.values(response.archetypeCounts).reduce(
    (sum, c) => sum + c,
    0
  );

  const archetypeEntries = Object.entries(response.archetypeCounts).sort(
    ([nameA], [nameB]) => nameA.localeCompare(nameB)
  );

  return (
    <div className="space-y-8">
      {allOpponents.length > 0 && (
        <div className="overflow-x-auto">
          <table className="min-w-full text-sm">
            <thead>
              <tr className="border-b border-gray-200">
                <th
                  rowSpan={2}
                  className="text-left py-2 pr-4 font-medium text-gray-600 align-bottom"
                >
                  Opponent
                </th>
                {variantKeys.map((key) => {
                  const variant = response.matchups[key];
                  const label = variantLabel(key, variant.cardCounts);
                  const deckCount = response.variantCounts[key] ?? 0;
                  return (
                    <th
                      key={key}
                      colSpan={showCounts ? 5 : 2}
                      className="text-center py-2 px-2 font-semibold text-gray-900 border-l border-gray-200"
                    >
                      <div>{label}</div>
                      <div className="text-xs font-normal text-gray-500">
                        {deckCount} deck{deckCount !== 1 ? 's' : ''}
                      </div>
                    </th>
                  );
                })}
              </tr>
              <tr className="border-b border-gray-200">
                {variantKeys.map((key) => (
                  <React.Fragment key={key}>
                    {showCounts && (
                      <>
                        <th className="text-right py-1 px-2 font-medium text-gray-600 border-l border-gray-200">W</th>
                        <th className="text-right py-1 px-2 font-medium text-gray-600">L</th>
                        <th className="text-right py-1 px-2 font-medium text-gray-600">T</th>
                      </>
                    )}
                    <th className={`text-right py-1 px-2 font-medium text-gray-600${!showCounts ? ' border-l border-gray-200' : ''}`}>Total</th>
                    <th className="text-right py-1 px-2 font-medium text-gray-600">Win Rate</th>
                  </React.Fragment>
                ))}
              </tr>
            </thead>
            <tbody>
              {allOpponents.map((opponent) => (
                <tr
                  key={opponent}
                  className="border-b border-gray-100 hover:bg-gray-50"
                >
                  <td className="py-2 pr-4 font-medium text-gray-900">
                    {opponent}
                  </td>
                  {variantKeys.map((key) => {
                    const record = response.matchups[key]?.matchups?.[opponent];
                    if (!record) {
                      return (
                        <React.Fragment key={key}>
                          <td className="border-l border-gray-200" colSpan={showCounts ? 5 : 2} />
                        </React.Fragment>
                      );
                    }
                    const total = record.wins + record.losses + record.ties;
                    return (
                      <React.Fragment key={key}>
                        {showCounts && (
                          <>
                            <td className="text-right py-2 px-2 text-gray-700 border-l border-gray-200">
                              {record.wins}
                            </td>
                            <td className="text-right py-2 px-2 text-gray-700">
                              {record.losses}
                            </td>
                            <td className="text-right py-2 px-2 text-gray-700">
                              {record.ties}
                            </td>
                          </>
                        )}
                        <td className={`text-right py-2 px-2 text-gray-700${!showCounts ? ' border-l border-gray-200' : ''}`}>
                          {total}
                        </td>
                        <td
                          className={`text-right py-2 px-2 font-semibold rounded ${winRateColor(record.winRate)}`}
                        >
                          {formatPercent(record.winRate)}
                        </td>
                      </React.Fragment>
                    );
                  })}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Metagame breakdown */}
      <div>
        <h3 className="text-lg font-semibold text-gray-900 mb-3">
          Metagame Breakdown
        </h3>
        <div className="overflow-x-auto">
          <table className="min-w-full text-sm">
            <thead>
              <tr className="border-b border-gray-200">
                <th className="text-left py-2 pr-4 font-medium text-gray-600">
                  Archetype
                </th>
                <th className="text-right py-2 px-3 font-medium text-gray-600">
                  Decks
                </th>
                <th className="text-right py-2 pl-3 font-medium text-gray-600">
                  Share
                </th>
              </tr>
            </thead>
            <tbody>
              {archetypeEntries.map(([archetype, count]) => (
                <tr
                  key={archetype}
                  className="border-b border-gray-100 hover:bg-gray-50"
                >
                  <td className="py-2 pr-4 font-medium text-gray-900">
                    {archetype}
                  </td>
                  <td className="text-right py-2 px-3 text-gray-700">
                    {count}
                  </td>
                  <td className="text-right py-2 pl-3 text-gray-700">
                    {totalArchetypeDecks > 0
                      ? formatPercent(count / totalArchetypeDecks)
                      : '—'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
