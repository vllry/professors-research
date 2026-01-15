import { useState } from 'react';
import DecklistInput from '../components/DecklistInput';
import CardSetsInput from '../components/CardSetsInput';
import OddsDisplay from '../components/OddsDisplay';
import { calculatePrizeOdds } from '../api/client';
import type { PrizeOddsResponse } from '../types/api';

export default function Prizes() {
  const [decklist, setDecklist] = useState('');
  const [cardSetsJson, setCardSetsJson] = useState('');
  const [prized, setPrized] = useState(true);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [response, setResponse] = useState<PrizeOddsResponse | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setResponse(null);

    try {
      let cardSets: Record<string, any[]> | undefined;
      if (cardSetsJson.trim()) {
        try {
          cardSets = JSON.parse(cardSetsJson);
        } catch (parseError) {
          throw new Error('Invalid JSON in Card Groups field');
        }
      }

      const result = await calculatePrizeOdds({
        decklist,
        cardSets,
        prized,
      });

      setResponse(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An unknown error occurred');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
        <h1 className="text-3xl font-bold text-gray-900 mb-4">
          Calculate Prize Probabilities
        </h1>
        <p className="text-gray-600 mb-8">
          Calculate the probability of prizing (or not-prizing) specific cards, or custom card combinations.
        </p>

        <form onSubmit={handleSubmit} className="space-y-6 mb-8">
          <div className="bg-white p-6 rounded-lg shadow">
            <DecklistInput value={decklist} onChange={setDecklist} />
          </div>

          <div className="bg-white p-6 rounded-lg shadow">
            <CardSetsInput value={cardSetsJson} onChange={setCardSetsJson} decklist={decklist} />
          </div>

          <div className="bg-white p-6 rounded-lg shadow">
            <label className="flex items-center space-x-3 cursor-pointer">
              <input
                type="checkbox"
                checked={prized}
                onChange={(e) => setPrized(e.target.checked)}
                className="w-5 h-5 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
              />
              <span className="text-sm font-medium text-gray-700">
                Calculate odds for prized cards (uncheck for not-prized cards)
              </span>
            </label>
          </div>

          <button
            type="submit"
            disabled={loading || !decklist.trim()}
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
            <p className="mt-4 text-gray-600">
              Calculating {prized ? 'prize' : 'not-prized'} odds...
            </p>
          </div>
        )}

        {response && !loading && (
          <div className="bg-white p-6 rounded-lg shadow">
            <OddsDisplay response={response} prized={prized} />
          </div>
        )}
      </div>
    </div>
  );
}

