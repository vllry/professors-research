import { useState } from 'react';
import type { CardSetJSON, CardJSON } from '../types/api';
import PatternEditor from './PatternEditor';

interface CardSetGroupEditorProps {
  groupName: string;
  cardSets: CardSetJSON[];
  onChange: (cardSets: CardSetJSON[]) => void;
  onRemove: () => void;
  decklistCards: CardJSON[];
}

export default function CardSetGroupEditor({
  groupName,
  cardSets,
  onChange,
  onRemove,
  decklistCards,
}: CardSetGroupEditorProps) {
  const [isExpanded, setIsExpanded] = useState(true);

  const handleCardSetChange = (index: number, cardSet: CardSetJSON) => {
    const newCardSets = [...cardSets];
    newCardSets[index] = cardSet;
    onChange(newCardSets);
  };

  const handleCardSetRemove = (index: number) => {
    const newCardSets = cardSets.filter((_, i) => i !== index);
    onChange(newCardSets);
  };

  const handleAddCardSet = () => {
    try {
      const newCardSet: CardSetJSON = { anyOfs: [] };
      onChange([...cardSets, newCardSet]);
    } catch (error) {
      console.error('Error adding card set:', error);
    }
  };

  const handlePatternChange = (
    cardSetIndex: number,
    patternIndex: number,
    pattern: any,
    isAnyOf: boolean
  ) => {
    const cardSet = { ...cardSets[cardSetIndex] };
    if (isAnyOf) {
      const anyOfs = [...(cardSet.anyOfs || [])];
      anyOfs[patternIndex] = pattern;
      cardSet.anyOfs = anyOfs;
    } else {
      const allOfs = [...(cardSet.allOfs || [])];
      allOfs[patternIndex] = pattern;
      cardSet.allOfs = allOfs;
    }
    handleCardSetChange(cardSetIndex, cardSet);
  };

  const handlePatternRemove = (
    cardSetIndex: number,
    patternIndex: number,
    isAnyOf: boolean
  ) => {
    const cardSet = { ...cardSets[cardSetIndex] };
    if (isAnyOf) {
      const anyOfs = (cardSet.anyOfs || []).filter((_, i) => i !== patternIndex);
      cardSet.anyOfs = anyOfs.length > 0 ? anyOfs : undefined;
    } else {
      const allOfs = (cardSet.allOfs || []).filter((_, i) => i !== patternIndex);
      cardSet.allOfs = allOfs.length > 0 ? allOfs : undefined;
    }
    handleCardSetChange(cardSetIndex, cardSet);
  };

  const handleAddPattern = (cardSetIndex: number, isAnyOf: boolean) => {
    const cardSet = { ...cardSets[cardSetIndex] };
    if (isAnyOf) {
      cardSet.anyOfs = [...(cardSet.anyOfs || []), { cards: [] }];
    } else {
      cardSet.allOfs = [...(cardSet.allOfs || []), { cards: [] }];
    }
    handleCardSetChange(cardSetIndex, cardSet);
  };

  const handleToggleType = (cardSetIndex: number) => {
    const cardSet = { ...cardSets[cardSetIndex] };
    // Swap anyOfs and allOfs
    if (cardSet.anyOfs) {
      cardSet.allOfs = cardSet.anyOfs;
      cardSet.anyOfs = undefined;
    } else if (cardSet.allOfs) {
      cardSet.anyOfs = cardSet.allOfs;
      cardSet.allOfs = undefined;
    }
    handleCardSetChange(cardSetIndex, cardSet);
  };

  return (
    <div className="border border-gray-300 rounded-lg p-4 bg-white">
      <div className="flex justify-between items-center mb-3">
        <button
          type="button"
          onClick={() => setIsExpanded(!isExpanded)}
          className="flex items-center gap-2 text-base font-semibold text-gray-800 hover:text-gray-900"
        >
          <span>{isExpanded ? '▼' : '▶'}</span>
          <span>{groupName}</span>
          <span className="text-gray-500 text-sm font-normal">
            ({cardSets.length} set{cardSets.length !== 1 ? 's' : ''})
          </span>
        </button>
        <button
          type="button"
          onClick={onRemove}
          className="px-3 py-1 text-sm text-red-600 hover:text-red-800 hover:bg-red-50 rounded"
        >
          Remove Group
        </button>
      </div>

      {isExpanded && (
        <div className="space-y-4">
          {cardSets.length === 0 ? (
            <p className="text-sm text-gray-500 italic">No card sets in this group</p>
          ) : (
            cardSets.map((cardSet, cardSetIndex) => (
              <div
                key={`cardset-${groupName}-${cardSetIndex}`}
                className="border border-gray-200 rounded-lg p-3 bg-gray-50"
              >
                <div className="flex justify-between items-center mb-2">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium text-gray-700">Card Set {cardSetIndex + 1}</span>
                    <button
                      type="button"
                      onClick={() => handleToggleType(cardSetIndex)}
                      className="px-2 py-1 text-xs bg-blue-100 text-blue-700 hover:bg-blue-200 rounded"
                    >
                      Switch to {cardSet.anyOfs ? 'AllOf' : 'AnyOf'}
                    </button>
                  </div>
                  <div className="flex gap-2">
                    {cardSet.anyOfs && (
                      <button
                        type="button"
                        onClick={() => handleAddPattern(cardSetIndex, true)}
                        className="px-2 py-1 text-xs text-blue-600 hover:text-blue-800 hover:bg-blue-50 rounded"
                      >
                        + Add AnyOf
                      </button>
                    )}
                    {cardSet.allOfs && (
                      <button
                        type="button"
                        onClick={() => handleAddPattern(cardSetIndex, false)}
                        className="px-2 py-1 text-xs text-blue-600 hover:text-blue-800 hover:bg-blue-50 rounded"
                      >
                        + Add AllOf
                      </button>
                    )}
                    <button
                      type="button"
                      onClick={() => handleCardSetRemove(cardSetIndex)}
                      className="px-2 py-1 text-xs text-red-600 hover:text-red-800 hover:bg-red-50 rounded"
                    >
                      Remove Set
                    </button>
                  </div>
                </div>

                <div className="space-y-2">
                  {cardSet.anyOfs?.map((pattern, patternIndex) => (
                    <PatternEditor
                      key={`anyof-${patternIndex}`}
                      pattern={pattern}
                      onChange={(p) => handlePatternChange(cardSetIndex, patternIndex, p, true)}
                      onRemove={() => handlePatternRemove(cardSetIndex, patternIndex, true)}
                      decklistCards={decklistCards}
                      isAnyOf={true}
                    />
                  ))}
                  {cardSet.allOfs?.map((pattern, patternIndex) => (
                    <PatternEditor
                      key={`allof-${patternIndex}`}
                      pattern={pattern}
                      onChange={(p) => handlePatternChange(cardSetIndex, patternIndex, p, false)}
                      onRemove={() => handlePatternRemove(cardSetIndex, patternIndex, false)}
                      decklistCards={decklistCards}
                      isAnyOf={false}
                    />
                  ))}
                  {(!cardSet.anyOfs || cardSet.anyOfs.length === 0) &&
                    (!cardSet.allOfs || cardSet.allOfs.length === 0) && (
                      <p className="text-sm text-gray-500 italic">No patterns in this card set</p>
                    )}
                </div>
              </div>
            ))
          )}
          <button
            type="button"
            onClick={handleAddCardSet}
            className="w-full px-3 py-2 text-sm text-blue-600 hover:text-blue-800 hover:bg-blue-50 rounded border border-dashed border-blue-300"
          >
            + Add Card Set
          </button>
        </div>
      )}
    </div>
  );
}

