import type { CardJSON } from '../types/api';

/**
 * Parses a decklist string in Pokemon TCG Live format to extract cards.
 * Format: <count> <card name> <setCode> <number>
 * 
 * @param decklist - The decklist string from Pokemon TCG Live
 * @returns Array of unique CardJSON objects found in the decklist
 */
export function parseDecklist(decklist: string): CardJSON[] {
  const cards: CardJSON[] = [];
  const seen = new Set<string>();
  
  const lines = decklist.split('\n');
  
  for (const line of lines) {
    const trimmed = line.trim();
    
    // Skip empty lines and category headers (e.g., "Pokemon: 20", "Total Cards: 60")
    if (!trimmed || trimmed.includes(':') || trimmed.toLowerCase().includes('total')) {
      continue;
    }
    
    // Parse format: <count> <card name> <setCode> <number>
    const parts = trimmed.split(/\s+/);
    
    // Need at least 4 parts: count, name (may have spaces), setCode, number
    if (parts.length < 4) {
      continue;
    }
    
    // First part should be the count (number)
    const countStr = parts[0];
    if (!/^\d+$/.test(countStr)) {
      continue;
    }
    
    // Last two parts are setCode and number
    const setCode = parts[parts.length - 2];
    const number = parts[parts.length - 1];
    
    // Everything in between is the card name
    const name = parts.slice(1, parts.length - 2).join(' ');
    
    // Create a unique key for this card
    const key = `${name}|${setCode}|${number}`;
    
    // Only add if we haven't seen this card before
    if (!seen.has(key)) {
      seen.add(key);
      cards.push({
        name,
        setCode,
        number,
      });
    }
  }
  
  return cards;
}

/**
 * Formats a CardJSON as a string in the format "Name SetCode Number"
 */
export function formatCardString(card: CardJSON): string {
  return `${card.name} ${card.setCode} ${card.number}`;
}

/**
 * Parses a card string in the format "Name SetCode Number" to CardJSON
 * Returns null if the format is invalid
 */
export function parseCardString(cardString: string): CardJSON | null {
  const trimmed = cardString.trim();
  if (!trimmed) {
    return null;
  }
  
  const parts = trimmed.split(/\s+/);
  
  // Need at least 3 parts: name, setCode, number
  if (parts.length < 3) {
    return null;
  }
  
  // Last two parts are setCode and number
  const setCode = parts[parts.length - 2];
  const number = parts[parts.length - 1];
  
  // Everything before is the card name
  const name = parts.slice(0, parts.length - 2).join(' ');
  
  if (!name || !setCode || !number) {
    return null;
  }
  
  return {
    name,
    setCode,
    number,
  };
}





