export default function BottomBar() {
  return (
    <footer className="border-t bg-white">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-3">
        <div className="flex flex-col gap-1 sm:flex-row sm:items-center sm:justify-between">
          <a
            href="https://github.com/vllry/professors-research"
            target="_blank"
            rel="noreferrer"
            className="text-sm font-medium text-gray-700 hover:text-gray-900 underline underline-offset-2"
          >
            GitHub
          </a>
          <a
            href="https://bsky.app/profile/isthelaststop.com"
            target="_blank"
            rel="noreferrer"
            className="text-xs text-gray-500 hover:text-gray-700 underline underline-offset-2"
          >
            Created by Vallery Lancey
          </a>
        </div>
      </div>
    </footer>
  );
}


