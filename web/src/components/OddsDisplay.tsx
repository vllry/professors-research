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

  return (
    <div className="space-y-6">
      {/* Individual Card Odds */}
      <div>
        <h3 className="text-xl font-semibold text-gray-900 mb-4">
          Individual Card Odds
        </h3>
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Card
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    {statusLabel} at least 1
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    {statusLabel} at least 2
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    {statusLabel} at least 3
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    {statusLabel} at least 4
                  </th>
                  {hasMoreColumn && (
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      More
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

