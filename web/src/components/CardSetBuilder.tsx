import { useState } from 'react';
import type { CardSetJSON, CardJSON } from '../types/api';
import CardSetGroupEditor from './CardSetGroupEditor';

interface CardSetBuilderProps {
  cardSets: Record<string, CardSetJSON[]>;
  onChange: (cardSets: Record<string, CardSetJSON[]>) => void;
  decklistCards: CardJSON[];
}

export default function CardSetBuilder({
  cardSets,
  onChange,
  decklistCards,
}: CardSetBuilderProps) {
  const [newGroupName, setNewGroupName] = useState('');

  const handleGroupChange = (groupName: string, groupCardSets: CardSetJSON[]) => {
    const newCardSets = { ...cardSets };
    if (groupCardSets.length === 0) {
      delete newCardSets[groupName];
    } else {
      newCardSets[groupName] = groupCardSets;
    }
    onChange(newCardSets);
  };

  const handleGroupRemove = (groupName: string) => {
    const newCardSets = { ...cardSets };
    delete newCardSets[groupName];
    onChange(newCardSets);
  };

  const handleAddGroup = () => {
    if (!newGroupName.trim()) {
      return;
    }
    const groupName = newGroupName.trim();
    if (cardSets[groupName]) {
      // Group already exists
      return;
    }
    onChange({ ...cardSets, [groupName]: [] });
    setNewGroupName('');
  };

  const groupNames = Object.keys(cardSets);

  return (
    <div className="space-y-4">
      <div className="flex gap-2 items-center">
        <input
          type="text"
          value={newGroupName}
          onChange={(e) => setNewGroupName(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter') {
              e.preventDefault();
              handleAddGroup();
            }
          }}
          placeholder="Group name (e.g., draw_support)"
          className="flex-1 px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
        />
        <button
          type="button"
          onClick={handleAddGroup}
          disabled={!newGroupName.trim()}
          className="px-4 py-2 bg-blue-600 text-white rounded-md text-sm font-medium hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed"
        >
          Add Group
        </button>
      </div>

      {groupNames.length === 0 ? (
        <p className="text-sm text-gray-500 italic text-center py-8">
          No groups yet. Add a group to get started.
        </p>
      ) : (
        <div className="space-y-3">
          {groupNames.map((groupName) => {
            const groupCardSets = cardSets[groupName] || [];
            return (
              <CardSetGroupEditor
                key={groupName}
                groupName={groupName}
                cardSets={groupCardSets}
                onChange={(cardSets) => handleGroupChange(groupName, cardSets)}
                onRemove={() => handleGroupRemove(groupName)}
                decklistCards={decklistCards}
              />
            );
          })}
        </div>
      )}
    </div>
  );
}

