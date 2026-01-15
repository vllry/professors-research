interface DecklistInputProps {
  value: string;
  onChange: (value: string) => void;
}

export default function DecklistInput({ value, onChange }: DecklistInputProps) {
  return (
    <div className="space-y-2">
      <label htmlFor="decklist" className="block text-sm font-medium text-gray-700">
        Decklist (paste from Live)
      </label>
      <textarea
        id="decklist"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="w-full h-64 px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 font-mono text-sm"
      />
    </div>
  );
}

