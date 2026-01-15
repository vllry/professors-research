import { useState } from 'react';
import type { AnyOfJSON, AllOfJSON, CardJSON } from '../types/api';
import CardEntryEditor from './CardEntryEditor';

interface PatternEditorProps {
  pattern: AnyOfJSON | AllOfJSON;
  onChange: (pattern: AnyOfJSON | AllOfJSON) => void;
  onRemove: () => void;
  decklistCards: CardJSON[];
  isAnyOf: boolean;
}

export default function PatternEditor({
  pattern,
  onChange,
  onRemove,
  decklistCards,
  isAnyOf,
}: PatternEditorProps) {
  const [isExpanded, setIsExpanded] = useState(true);

  const handleCardChange = (index: number, card: { card: CardJSON; count: number }) => {
    const newCards = [...pattern.cards];
    newCards[index] = card;
    onChange({ ...pattern, cards: newCards });
  };

  const handleCardRemove = (index: number) => {
    const newCards = pattern.cards.filter((_, i) => i !== index);
    onChange({ ...pattern, cards: newCards });
  };

  const handleAddCard = () => {
    const newCard = {
      card: { name: '', setCode: '', number: '' },
      count: 1,
    };
    onChange({ ...pattern, cards: [...pattern.cards, newCard] });
  };

  return (
    <div className="border border-gray-200 rounded-lg p-4 bg-gray-50">
      <div className="flex justify-between items-center mb-3">
        <button
          type="button"
          onClick={() => setIsExpanded(!isExpanded)}
          className="flex items-center gap-2 text-sm font-medium text-gray-700 hover:text-gray-900"
        >
          <span>{isExpanded ? '▼' : '▶'}</span>
          <span>{isAnyOf ? 'AnyOf Pattern' : 'AllOf Pattern'}</span>
          <span className="text-gray-500 text-xs">
            ({pattern.cards.length} card{pattern.cards.length !== 1 ? 's' : ''})
          </span>
        </button>
        <div className="flex gap-2">
          <button
            type="button"
            onClick={handleAddCard}
            className="px-2 py-1 text-sm text-blue-600 hover:text-blue-800 hover:bg-blue-50 rounded"
          >
            + Add Card
          </button>
          <button
            type="button"
            onClick={onRemove}
            className="px-2 py-1 text-sm text-red-600 hover:text-red-800 hover:bg-red-50 rounded"
          >
            Remove Pattern
          </button>
        </div>
      </div>

      {isExpanded && (
        <div className="space-y-2">
          {pattern.cards.length === 0 ? (
            <p className="text-sm text-gray-500 italic">No cards in this pattern</p>
          ) : (
            pattern.cards.map((cardEntry, index) => (
              <CardEntryEditor
                key={index}
                card={cardEntry}
                onChange={(card) => handleCardChange(index, card)}
                onRemove={() => handleCardRemove(index)}
                decklistCards={decklistCards}
              />
            ))
          )}
        </div>
      )}
    </div>
  );
}


