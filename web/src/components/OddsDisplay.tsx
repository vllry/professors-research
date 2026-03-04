import type { PrizeOddsResponse } from '../types/api';

interface OddsDisplayProps {
  response: PrizeOddsResponse;
  prized: boolean;
}

function formatPercent(value: number): string {
  return `${(value * 100).toFixed(2)}%`;
}

function getColorForPercent(value: number): string {
  if (value >= 0.5) return 'bg-red-500';
  if (value >= 0.3) return 'bg-yellow-500';
  if (value >= 0.1) return 'bg-blue-500';
  return 'bg-green-500';
}

export default function OddsDisplay({ response, prized }: OddsDisplayProps) {
  const hasMoreColumn = Object.values(response.odds).some(odds => odds.length > 4);
  const statusLabel = prized ? 'Prize' : 'Not-prized';
  const targetLabel = prized ? 'the 6 prize cards' : 'the 54 non-prize cards';
  const locationLabel = prized ? 'prizes' : 'non-prize cards';
  const hasWarnings = Boolean(response.errors && response.errors.length > 0);

  return (
    <div className="space-y-6">
      {/* Errors/Warnings */}
      {hasWarnings && (
        <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
          <h3 className="text-lg font-semibold text-yellow-800 mb-2">Warnings</h3>
          <ul className="list-disc list-inside space-y-1">
            <li className="text-sm text-yellow-700">
              Some cards could not be identified. This matters slightly because the opening-hand adjustment depends on correctly identifying which cards are Basic Pokémon. Fixing these cards will make the results more accurate.
            </li>
            {response.errors!.map((error, index) => (
              <li key={index} className="text-sm text-yellow-700">
                {error.info}
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* Individual Card Odds */}
      <div>
        <h3 className="text-xl font-semibold text-gray-900 mb-4">
          Individual Card Odds
        </h3>
        <p className="text-sm text-gray-600 mb-4">
        The base odds come from choosing 6 cards out of 60 — a hypergeometric probability for each card based on how many copies are in the deck. We then adjust for the opening hand: since your 7-card hand must contain at least one Basic Pokémon, we filter out deals where all Basics ended up in prizes or the remaining deck, and renormalize. This makes Basics slightly less likely to appear in prizes and non-Basics slightly more likely, though the difference is marginal in most lists.
        </p>
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    <div>Card</div>
                    <div className="mt-1 text-[10px] font-normal normal-case text-gray-500">
                      Name / set / number
                    </div>
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    <div>{statusLabel} at least 1</div>
                    <div className="mt-1 text-[10px] font-normal normal-case text-gray-500">
                      One or more copies are in your {locationLabel}
                    </div>
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    <div>{statusLabel} at least 2</div>
                    <div className="mt-1 text-[10px] font-normal normal-case text-gray-500">
                      Two or more copies are in your {locationLabel}
                    </div>
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    <div>{statusLabel} at least 3</div>
                    <div className="mt-1 text-[10px] font-normal normal-case text-gray-500">
                      Three or more copies are in your {locationLabel}
                    </div>
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    <div>{statusLabel} at least 4</div>
                    <div className="mt-1 text-[10px] font-normal normal-case text-gray-500">
                      Four or more copies are in your {locationLabel}
                    </div>
                  </th>
                  {hasMoreColumn && (
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      <div>More</div>
                      <div className="mt-1 text-[10px] font-normal normal-case text-gray-500">
                        Extra thresholds not shown
                      </div>
                    </th>
                  )}
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {Object.entries(response.odds).map(([cardName, odds]) => (
                  <tr key={cardName} className="hover:bg-gray-50">
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                      {cardName}
                    </td>
                    {Array.from({ length: 4 }, (_, index) => {
                      if (index < odds.length) {
                        return (
                          <td key={index} className="px-6 py-4 whitespace-nowrap">
                            <div className="flex items-center space-x-2">
                              <span className="text-sm text-gray-900">
                                {formatPercent(odds[index])}
                              </span>
                              <div className="w-16 h-2 bg-gray-200 rounded-full overflow-hidden">
                                <div
                                  className={`h-full ${getColorForPercent(odds[index])}`}
                                  style={{ width: `${Math.min(odds[index] * 100, 100)}%` }}
                                />
                              </div>
                            </div>
                          </td>
                        );
                      } else {
                        return (
                          <td key={index} className="px-6 py-4 whitespace-nowrap text-sm text-gray-400">
                            -
                          </td>
                        );
                      }
                    })}
                    {hasMoreColumn && (
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {odds.length > 4 ? `+${odds.length - 4} more` : '-'}
                      </td>
                    )}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      {/* Card Set Odds */}
      {response.cardSetOdds && Object.keys(response.cardSetOdds).length > 0 && (
        <div>
          <h3 className="text-xl font-semibold text-gray-900 mb-4">
            Card Set Odds
          </h3>
          <p className="text-sm text-gray-600 mb-4">
            Each value is the chance that a satisfiable card combination appears in {targetLabel}. Card set odds use the simpler “choose from 60” model and do not condition on the opening hand rule.
          </p>
          <div className="bg-white rounded-lg shadow overflow-hidden">
            <div className="p-6 space-y-4">
              {Object.entries(response.cardSetOdds).map(([groupName, groupOdds]) => (
                <div key={groupName} className="border-b border-gray-200 last:border-b-0 pb-4 last:pb-0">
                  <h4 className="text-lg font-medium text-gray-900 mb-3">
                    {groupName}
                  </h4>
                  <div className="space-y-2">
                    {Object.entries(groupOdds).map(([setName, odds]) => (
                      <div key={setName} className="flex items-center justify-between">
                        <span className="text-sm text-gray-700">{setName}</span>
                        <div className="flex items-center space-x-3">
                          <span className="text-sm font-semibold text-gray-900">
                            {formatPercent(odds)}
                          </span>
                          <div className="w-32 h-3 bg-gray-200 rounded-full overflow-hidden">
                            <div
                              className={`h-full ${getColorForPercent(odds)}`}
                              style={{ width: `${Math.min(odds * 100, 100)}%` }}
                            />
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

