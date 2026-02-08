type Resource = {
  title: string;
  url: string;
  brief: string;
  description: string;
};

const RESOURCES: Resource[] = [
  {
    title: 'TrainerHill',
    url: 'https://www.trainerhill.com',
    brief: 'Matchup stats with deep filtering.',
    description:
      'View match up statistics for major and online tournaments, with filtering. This is one of the best ways you can predict popularity and aggregate performance of different decks. You can also view performance split by specific card counts.',
  },
  {
    title: 'Limitless',
    url: 'https://limitlesstcg.com',
    brief: 'Tournament results + top-performing decklists.',
    description:
      'Top performing deck lists, and tournament results in an easy browse format. Includes a playtesting tool good for setting up example board states with custom decklists.',
  },
  {
    title: 'Limitless Labs',
    url: 'https://labs.limitlesstcg.com',
    brief: 'High-level matchup stats from past events.',
    description:
      'Coarse matchup statistics from past events, such as win rates by deck archetype by phase.',
  },
  {
    title: 'Japan City League Results',
    url: 'https://limitlesstcg.com/tournaments/jp',
    brief: "Japan's City League metagame snapshot.",
    description:
      "Japan's city league tournaments are reported online, and give a macro snapshot of upcoming meta decks and trends. Note that metas are always regionalized, and the best of 1 format biases toward different builds than mixed or best of 3 events.",
  },
];

export default function OtherResources() {
  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Other Resources</h1>
          <p className="mt-2 text-gray-600">
            Other useful Pokemon TCG tools.
          </p>
        </div>
      </div>

      <div className="mt-6 grid grid-cols-1 sm:grid-cols-2 gap-4">
        {RESOURCES.map((resource) => (
          <div
            key={resource.title}
            className="rounded-lg border border-gray-200 bg-white p-5 shadow-sm hover:shadow transition-shadow flex flex-col"
          >
            <div>
              <h2 className="text-lg font-semibold text-gray-900">{resource.title}</h2>
              <p className="mt-1 text-sm font-medium text-gray-700">{resource.brief}</p>
            </div>

            <p className="mt-3 text-sm text-gray-600">{resource.description}</p>

            <div className="mt-4 pt-4 border-t border-gray-100">
              <a
                href={resource.url}
                target="_blank"
                rel="noreferrer"
                className="text-sm text-blue-700 hover:underline break-all"
              >
                {resource.url}
              </a>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

