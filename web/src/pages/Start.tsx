import { useState } from 'react';
import DecklistInput from '../components/DecklistInput';
import CardSetsInput from '../components/CardSetsInput';
import StartOddsDisplay from '../components/StartOddsDisplay';
import FeatureBadge from '../components/FeatureBadge';
import { calculateStartOdds } from '../api/client';
import type { StartOddsResponse } from '../types/api';

export default function Start() {
  const [decklist, setDecklist] = useState('');
  const [cardSetsJson, setCardSetsJson] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [response, setResponse] = useState<StartOddsResponse | null>(null);

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

      const result = await calculateStartOdds({
        decklist,
        cardSets,
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
        <h1 className="flex flex-wrap items-center gap-x-3 gap-y-2 text-3xl font-bold text-gray-900 mb-4">
          <span>Calculate Starting Hand Probabilities</span>
          <FeatureBadge stage="beta" />
        </h1>
        <p className="text-gray-600 mb-8">
          Calculate the odds of starting with particular basics, how likely you are to have cards in your starting hand,
          and the probability of having custom card combinations in your starting hand.
        </p>

        <form onSubmit={handleSubmit} className="space-y-6 mb-8">
          <div className="bg-white p-6 rounded-lg shadow">
            <DecklistInput value={decklist} onChange={setDecklist} />
          </div>

          <div className="bg-white p-6 rounded-lg shadow">
            <CardSetsInput value={cardSetsJson} onChange={setCardSetsJson} decklist={decklist} />
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
            <p className="mt-4 text-gray-600">Calculating start odds...</p>
          </div>
        )}

        {response && !loading && (
          <div className="bg-white p-6 rounded-lg shadow">
            <StartOddsDisplay response={response} />
          </div>
        )}
      </div>
    </div>
  );
}




