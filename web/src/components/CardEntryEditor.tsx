import { useState, useRef, useEffect } from 'react';
import type { CardJSON } from '../types/api';
import { parseCardString, formatCardString } from '../utils/decklistParser';

interface CardEntryEditorProps {
  card: {
    card: CardJSON;
    count: number;
  };
  onChange: (card: { card: CardJSON; count: number }) => void;
  onRemove: () => void;
  decklistCards: CardJSON[];
}

export default function CardEntryEditor({
  card,
  onChange,
  onRemove,
  decklistCards,
}: CardEntryEditorProps) {
  const [cardInput, setCardInput] = useState(formatCardString(card.card));
  const [showAutocomplete, setShowAutocomplete] = useState(false);
  const [filteredCards, setFilteredCards] = useState<CardJSON[]>([]);
  const justSelectedRef = useRef(false);
  const inputRef = useRef<HTMLInputElement>(null);
  const autocompleteRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    // Update input when card prop changes externally
    setCardInput(formatCardString(card.card));
  }, [card.card]);

  useEffect(() => {
    // Filter cards based on input
    const trimmedInput = cardInput.trim();
    if (trimmedInput) {
      // Use trimmed input for filtering to handle leading/trailing spaces correctly
      const query = trimmedInput.toLowerCase();
      const filtered = decklistCards.filter(
        (c) =>
          c.name.toLowerCase().includes(query) ||
          c.setCode.toLowerCase().includes(query) ||
          c.number.toLowerCase().includes(query) ||
          formatCardString(c).toLowerCase().includes(query)
      );
      setFilteredCards(filtered.slice(0, 10)); // Limit to 10 results
      // Don't show autocomplete if we just selected a card (user clicked an option)
      if (!justSelectedRef.current) {
        setShowAutocomplete(filtered.length > 0);
      } else {
        // Reset the flag after processing
        justSelectedRef.current = false;
      }
    } else {
      setFilteredCards([]);
      // Don't show autocomplete if we just selected a card
      if (!justSelectedRef.current) {
        setShowAutocomplete(false);
      } else {
        // Reset the flag after processing
        justSelectedRef.current = false;
      }
    }
  }, [cardInput, decklistCards]);

  useEffect(() => {
    // Close autocomplete when clicking outside
    const handleClickOutside = (event: MouseEvent) => {
      if (
        autocompleteRef.current &&
        !autocompleteRef.current.contains(event.target as Node) &&
        inputRef.current &&
        !inputRef.current.contains(event.target as Node)
      ) {
        setShowAutocomplete(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleCardInputChange = (value: string) => {
    setCardInput(value);
    // Try to parse and update if valid
    const parsed = parseCardString(value);
    if (parsed) {
      onChange({ card: parsed, count: card.count });
    }
  };

  const handleCardSelect = (selectedCard: CardJSON) => {
    justSelectedRef.current = true; // Prevent useEffect from showing autocomplete again
    setCardInput(formatCardString(selectedCard));
    onChange({ card: selectedCard, count: card.count });
    setShowAutocomplete(false);
    inputRef.current?.blur();
  };

  const handleCountChange = (count: number) => {
    const validCount = Math.max(1, Math.floor(count) || 1);
    onChange({ card: card.card, count: validCount });
  };

  const isValid = parseCardString(cardInput) !== null;

  return (
    <div className="flex gap-2 items-start">
      <div className="flex-1 relative">
        <input
          ref={inputRef}
          type="text"
          value={cardInput}
          onChange={(e) => handleCardInputChange(e.target.value)}
          onFocus={() => {
            if (cardInput.trim()) {
              // If there's input, show filtered results
              if (filteredCards.length > 0) {
                setShowAutocomplete(true);
              }
            } else {
              // If input is empty, show all cards
              if (decklistCards.length > 0) {
                setFilteredCards(decklistCards.slice(0, 10));
                setShowAutocomplete(true);
              }
            }
          }}
          onBlur={() => {
            // Close autocomplete when input loses focus
            // Use setTimeout to allow click events on autocomplete options to fire first
            setTimeout(() => {
              if (!autocompleteRef.current?.contains(document.activeElement)) {
                setShowAutocomplete(false);
              }
            }, 150);
          }}
          placeholder="Card Name SetCode Number"
          className={`w-full px-3 py-2 border rounded-md text-sm ${
            isValid
              ? 'border-gray-300 focus:ring-2 focus:ring-blue-500 focus:border-blue-500'
              : 'border-red-300 focus:ring-2 focus:ring-red-500 focus:border-red-500'
          }`}
        />
        {showAutocomplete && filteredCards.length > 0 && (
          <div
            ref={autocompleteRef}
            className="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-md shadow-lg max-h-60 overflow-y-auto"
          >
            {filteredCards.map((c, idx) => (
              <button
                key={`${c.name}-${c.setCode}-${c.number}-${idx}`}
                type="button"
                onClick={() => handleCardSelect(c)}
                className="w-full text-left px-3 py-2 hover:bg-blue-50 focus:bg-blue-50 focus:outline-none text-sm"
              >
                <div className="font-medium">{c.name}</div>
                <div className="text-xs text-gray-500">
                  {c.setCode} {c.number}
                </div>
              </button>
            ))}
          </div>
        )}
      </div>
      <input
        type="number"
        min="1"
        value={card.count}
        onChange={(e) => handleCountChange(parseInt(e.target.value, 10))}
        className="w-20 px-2 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
      />
      <button
        type="button"
        onClick={onRemove}
        className="px-3 py-2 text-red-600 hover:text-red-800 hover:bg-red-50 rounded-md text-sm font-medium"
      >
        Remove
      </button>
    </div>
  );
}

