import { useState, useEffect, useCallback } from 'react';
import type { CardSetJSON } from '../types/api';
import { parseDecklist } from '../utils/decklistParser';
import CardSetBuilder from './CardSetBuilder';

interface CardSetsInputProps {
  value: string;
  onChange: (value: string) => void;
  decklist?: string;
}

const EXAMPLE_CARDSETS = `{
  "draw_support": [
    {
      "anyOfs": [
        {
          "cards": [
            {
              "card": {
                "name": "Lillies Determination",
                "setCode": "MEG",
                "number": "119"
              },
              "count": 1
            },
            {
              "card": {
                "name": "Iono",
                "setCode": "PAL",
                "number": "185"
              },
              "count": 1
            }
          ]
        }
      ]
    }
  ]
}`;

export default function CardSetsInput({ value, onChange, decklist = '' }: CardSetsInputProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const [showExamples, setShowExamples] = useState(false);
  const [mode, setMode] = useState<'builder' | 'json'>('builder');
  const [cardSets, setCardSets] = useState<Record<string, CardSetJSON[]>>({});
  const [jsonError, setJsonError] = useState<string | null>(null);

  // Parse decklist to get cards for autocomplete
  const decklistCards = decklist ? parseDecklist(decklist) : [];

  // Parse JSON string to CardSets structure
  const parseJsonToCardSets = useCallback((jsonString: string): Record<string, CardSetJSON[]> | null => {
    if (!jsonString.trim()) {
      return {};
    }
    try {
      const parsed = JSON.parse(jsonString);
      if (typeof parsed !== 'object' || parsed === null || Array.isArray(parsed)) {
        throw new Error('CardSets must be an object');
      }
      // Validate structure
      for (const [key, value] of Object.entries(parsed)) {
        if (!Array.isArray(value)) {
          throw new Error(`Group "${key}" must be an array`);
        }
        for (const cardSet of value) {
          if (typeof cardSet !== 'object' || cardSet === null) {
            throw new Error(`CardSet in group "${key}" must be an object`);
          }
          if (cardSet.anyOfs && !Array.isArray(cardSet.anyOfs)) {
            throw new Error(`anyOfs in group "${key}" must be an array`);
          }
          if (cardSet.allOfs && !Array.isArray(cardSet.allOfs)) {
            throw new Error(`allOfs in group "${key}" must be an array`);
          }
        }
      }
      return parsed as Record<string, CardSetJSON[]>;
    } catch (error) {
      setJsonError(error instanceof Error ? error.message : 'Invalid JSON');
      return null;
    }
  }, []);

  // Convert CardSets structure to JSON string
  const cardSetsToJson = useCallback((cardSets: Record<string, CardSetJSON[]>): string => {
    try {
      // Remove empty groups
      const cleaned: Record<string, CardSetJSON[]> = {};
      for (const [key, value] of Object.entries(cardSets)) {
        if (value && value.length > 0) {
          cleaned[key] = value;
        }
      }
      return JSON.stringify(cleaned, null, 2);
    } catch (error) {
      console.error('Error serializing card sets to JSON:', error);
      return '{}';
    }
  }, []);

  // Initial sync: parse existing JSON on mount
  useEffect(() => {
    if (value.trim()) {
      const parsed = parseJsonToCardSets(value);
      if (parsed !== null) {
        setCardSets(parsed);
        setJsonError(null);
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // Only run on mount - parseJsonToCardSets is stable

  // Sync from JSON to builder when JSON mode changes value
  useEffect(() => {
    if (mode === 'json') {
      const parsed = parseJsonToCardSets(value);
      if (parsed !== null) {
        setCardSets(parsed);
        setJsonError(null);
      }
    }
  }, [value, mode, parseJsonToCardSets]);

  // Sync from builder to JSON when builder changes
  const handleBuilderChange = useCallback((newCardSets: Record<string, CardSetJSON[]>) => {
    try {
      setCardSets(newCardSets);
      const jsonString = cardSetsToJson(newCardSets);
      onChange(jsonString);
      setJsonError(null);
    } catch (error) {
      console.error('Error in handleBuilderChange:', error);
      setJsonError(error instanceof Error ? error.message : 'Error updating card sets');
    }
  }, [onChange, cardSetsToJson]);

  // Handle JSON input change
  const handleJsonChange = (newValue: string) => {
    onChange(newValue);
    setJsonError(null);
  };

  // Copy JSON to clipboard
  const handleCopyJson = async () => {
    try {
      const jsonToCopy = value.trim() || '{}';
      await navigator.clipboard.writeText(jsonToCopy);
      // Could add a toast notification here
    } catch (error) {
      console.error('Failed to copy JSON:', error);
    }
  };

  // Paste JSON from clipboard
  const handlePasteJson = async () => {
    try {
      const text = await navigator.clipboard.readText();
      const parsed = parseJsonToCardSets(text);
      if (parsed !== null) {
        const jsonString = cardSetsToJson(parsed);
        onChange(jsonString);
        setJsonError(null);
      }
    } catch (error) {
      setJsonError('Failed to paste from clipboard');
    }
  };

  if (!isExpanded) {
    return (
      <div>
        <button
          type="button"
          onClick={() => setIsExpanded(true)}
          className="text-sm text-blue-600 hover:text-blue-800 font-medium"
        >
          + Add Card Groups (Advanced)
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-2">
      <div className="flex justify-between items-center">
        <label htmlFor="cardsets" className="block text-sm font-medium text-gray-700">
          Card Groups (Optional)
        </label>
        <button
          type="button"
          onClick={() => {
            setIsExpanded(false);
            onChange('');
          }}
          className="text-sm text-gray-500 hover:text-gray-700"
        >
          Hide
        </button>
      </div>
      <div className="bg-blue-50 border border-blue-200 rounded-md p-4 mb-2 space-y-3">
        <div>
          <h3 className="text-sm font-semibold text-gray-900 mb-2">How Card Groups Work</h3>
          <p className="text-sm text-gray-700 mb-3">
            <strong>Card Groups</strong> organize related card combinations for calculating prize odds. 
            Each group contains one or more <strong>card sets</strong>, and the system calculates the probability that 
            <strong className="text-blue-700"> ANY card set in the group</strong> is prized.
            You can use card groups to think about combinations of cards that would restrict higher level actions,
            such as having trouble setting up or recovering a key Pokemon.
          </p>
        </div>

        <div>
          <h4 className="text-xs font-semibold text-gray-800 mb-1">Card Sets</h4>
          <p className="text-sm text-gray-700 mb-2">
            Each card set defines an abstracted combination of cards, to avoid having to list all possible combinations.
            A card set can operate in two modes:
          </p>
          <ul className="text-sm text-gray-700 space-y-1 ml-4 list-disc">
            <li><strong>AnyOf mode:</strong> The card set matches if <strong>ANY</strong> of its patterns match</li>
            <li><strong>AllOf mode:</strong> The card set matches if <strong>ALL</strong> of its patterns match</li>
          </ul>
        </div>

        <div>
          <h4 className="text-xs font-semibold text-gray-800 mb-1">Patterns</h4>
          <p className="text-sm text-gray-700 mb-2">
            Each pattern specifies a combination of cards with counts. A pattern matches if you have 
            <strong> at least</strong> the specified count of each card in your 6 prize cards.
          </p>
          <p className="text-xs text-gray-600 italic">
            Example: A pattern with "1x Kirlia, 1x Rare Candy" matches if you prize at least 1 Kirlia AND at least 1 Rare Candy.
          </p>
        </div>

        <div className="bg-white border border-blue-300 rounded mt-3">
          <button
            type="button"
            onClick={() => setShowExamples(!showExamples)}
            className="w-full flex items-center justify-between p-3 text-left hover:bg-blue-50 transition-colors rounded"
          >
            <h4 className="text-xs font-semibold text-gray-800">See Examples</h4>
            <span className="text-gray-500 text-sm">{showExamples ? '▼' : '▶'}</span>
          </button>
          {showExamples && (
            <div className="p-3 pt-0 space-y-4 border-t border-blue-200">
              <div>
                <p className="text-xs font-semibold text-gray-800 mb-1">
                  Example 1: "I want to check odds for either Kirlia OR Rare Candy (any one)"
                </p>
                <p className="text-xs text-gray-700 mb-2">
                  Create one card set in <strong>AnyOf</strong> mode with one pattern containing both "1x Kirlia" and "1x Rare Candy". 
                  In AnyOf mode, a pattern with multiple cards means "choose one" (OR logic), so this matches if you have at least 1 Kirlia OR at least 1 Rare Candy.
                </p>
                <div className="bg-gray-50 rounded p-2 text-xs">
                  <p className="font-semibold text-gray-700 mb-1">Minimum matching combinations:</p>
                  <ul className="text-gray-600 space-y-0.5 ml-4 list-disc">
                    <li>[Kirlia]</li>
                    <li>[Rare Candy]</li>
                  </ul>
                </div>
              </div>

              <div>
                <p className="text-xs font-semibold text-gray-800 mb-1">
                  Example 2: "I want to check odds for both Kirlia AND Rare Candy (together)"
                </p>
                <p className="text-xs text-gray-700 mb-2">
                  Create one card set in <strong>AnyOf</strong> mode with one pattern containing "1x Kirlia, 1x Rare Candy". 
                  The pattern matches only if you have both cards in your prizes.
                </p>
                <div className="bg-gray-50 rounded p-2 text-xs">
                  <p className="font-semibold text-gray-700 mb-1">Minimum matching combination:</p>
                  <ul className="text-gray-600 space-y-0.5 ml-4 list-disc">
                    <li>[Kirlia, Rare Candy]</li>
                  </ul>
                </div>
              </div>

              <div>
                <p className="text-xs font-semibold text-gray-800 mb-1">
                  Example 3: "I want to check odds for 2 Gardevoir, or 1 Gardevoir + 1 of (Night Stretcher, Super Rod)"
                </p>
                <p className="text-xs text-gray-700 mb-2">
                  Create a group with two card sets: Card Set 1 (AnyOf) has one pattern "2x Gardevoir", 
                  Card Set 2 (AnyOf) has two patterns: Pattern 1 contains "1x Gardevoir", Pattern 2 contains "1x Night Stretcher, 1x Super Rod" 
                  (meaning choose one from each pattern: 1 Gardevoir AND 1 of either Night Stretcher OR Super Rod). The group matches if either card set matches.
                </p>
                <div className="bg-gray-50 rounded p-2 text-xs">
                  <p className="font-semibold text-gray-700 mb-1">Minimum matching combinations:</p>
                  <ul className="text-gray-600 space-y-0.5 ml-4 list-disc">
                    <li>[Gardevoir, Gardevoir]</li>
                    <li>[Gardevoir, Night Stretcher]</li>
                    <li>[Gardevoir, Super Rod]</li>
                  </ul>
                </div>
              </div>

              <div>
                <p className="text-xs font-semibold text-gray-800 mb-1">
                  Example 4: "I want to check odds for 2 in any combination of Munkidori + basic dark energy"
                </p>
                <p className="text-xs text-gray-700 mb-2">
                  This means you need exactly 2 cards total, which can be: 2 Munkidori, OR 2 Basic Dark Energy, OR 1 of each.
                  There are multiple ways to express this. 
                  Create a group with three card sets, each in <strong>AllOf</strong> mode: Card Set 1 has one pattern containing "2x Munkidori", 
                  Card Set 2 has one pattern containing "2x Basic Dark Energy", Card Set 3 has one pattern containing "1x Munkidori, 1x Basic Dark Energy". 
                  The group calculates the odds that <strong>ANY</strong> of these card sets match (union probability).
                </p>
                <div className="bg-gray-50 rounded p-2 text-xs">
                  <p className="font-semibold text-gray-700 mb-1">Minimum matching combinations:</p>
                  <ul className="text-gray-600 space-y-0.5 ml-4 list-disc">
                    <li>[Munkidori, Munkidori]</li>
                    <li>[Basic Dark Energy, Basic Dark Energy]</li>
                    <li>[Munkidori, Basic Dark Energy]</li>
                  </ul>
                </div>
              </div>

            </div>
          )}
        </div>
      </div>

      {/* Mode Toggle */}
      <div className="flex gap-2 items-center border-b border-gray-200 pb-2">
        <button
          type="button"
          onClick={() => {
            // Sync JSON to builder when switching to builder mode
            if (mode === 'json') {
              const parsed = parseJsonToCardSets(value);
              if (parsed !== null) {
                setCardSets(parsed);
                setJsonError(null);
              }
            }
            setMode('builder');
          }}
          className={`px-4 py-2 text-sm font-medium rounded-md transition-colors ${
            mode === 'builder'
              ? 'bg-blue-600 text-white'
              : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
          }`}
        >
          Builder
        </button>
        <button
          type="button"
          onClick={() => {
            // Sync builder to JSON when switching to JSON mode
            if (mode === 'builder') {
              const jsonString = cardSetsToJson(cardSets);
              onChange(jsonString);
            }
            setMode('json');
          }}
          className={`px-4 py-2 text-sm font-medium rounded-md transition-colors ${
            mode === 'json'
              ? 'bg-blue-600 text-white'
              : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
          }`}
        >
          JSON
        </button>
        {mode === 'json' && (
          <div className="flex gap-2 ml-auto">
            <button
              type="button"
              onClick={handleCopyJson}
              className="px-3 py-2 text-sm text-blue-600 hover:text-blue-800 hover:bg-blue-50 rounded border border-blue-300"
            >
              Copy JSON
            </button>
            <button
              type="button"
              onClick={handlePasteJson}
              className="px-3 py-2 text-sm text-blue-600 hover:text-blue-800 hover:bg-blue-50 rounded border border-blue-300"
            >
              Paste JSON
            </button>
          </div>
        )}
      </div>

      {/* Builder Mode */}
      {mode === 'builder' && (
        <div className="border border-gray-300 rounded-md p-4 bg-gray-50">
          <CardSetBuilder
            cardSets={cardSets}
            onChange={handleBuilderChange}
            decklistCards={decklistCards}
          />
        </div>
      )}

      {/* JSON Mode */}
      {mode === 'json' && (
        <div>
          <textarea
            id="cardsets"
            value={value}
            onChange={(e) => handleJsonChange(e.target.value)}
            placeholder={EXAMPLE_CARDSETS}
            className={`w-full h-48 px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-2 font-mono text-sm ${
              jsonError
                ? 'border-red-300 focus:ring-red-500 focus:border-red-500'
                : 'border-gray-300 focus:ring-blue-500 focus:border-blue-500'
            }`}
          />
          {jsonError && (
            <p className="text-xs text-red-600 mt-1">{jsonError}</p>
          )}
          <p className="text-xs text-gray-500 mt-1">
            Enter Card Groups as JSON. Leave empty to calculate only individual card odds.
          </p>
        </div>
      )}
    </div>
  );
}
