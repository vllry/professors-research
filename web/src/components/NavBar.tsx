import { useState } from 'react';
import { Link, useLocation } from 'react-router-dom';

export default function NavBar() {
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const location = useLocation();
  const isPrizesPage = location.pathname === '/prizes';
  const isDrawSupportersPage = location.pathname === '/draw-supporters';
  const isStartPage = location.pathname === '/start';
  const isMatchupsPage = location.pathname === '/matchups';
  const isOtherResourcesPage = location.pathname === '/other-resources';

  return (
    <nav className="text-white shadow-lg" style={{ backgroundColor: '#330625' }}>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center h-16">
          {/* Home button */}
          <Link
            to="/"
            className="text-xl font-bold hover:text-gray-300 transition-colors"
          >
            Professor's Research
          </Link>

          {/* Desktop menu */}
          <div className="hidden md:flex flex-1 items-center ml-6">
            <div className="flex items-center space-x-4">
              <Link
                to="/prizes"
                className={`px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                  isPrizesPage
                    ? 'bg-gray-700 text-white'
                    : 'text-gray-300 hover:bg-gray-700 hover:text-white'
                }`}
              >
                Prizes
              </Link>
              <Link
                to="/draw-supporters"
                className={`px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                  isDrawSupportersPage
                    ? 'bg-gray-700 text-white'
                    : 'text-gray-300 hover:bg-gray-700 hover:text-white'
                }`}
              >
                Draw Supporters
              </Link>
              <Link
                to="/start"
                className={`px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                  isStartPage
                    ? 'bg-gray-700 text-white'
                    : 'text-gray-300 hover:bg-gray-700 hover:text-white'
                }`}
              >
                Starting Hand
              </Link>
              <Link
                to="/matchups"
                className={`px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                  isMatchupsPage
                    ? 'bg-gray-700 text-white'
                    : 'text-gray-300 hover:bg-gray-700 hover:text-white'
                }`}
              >
                Matchups
              </Link>
            </div>

            <div className="ml-auto">
              <Link
                to="/other-resources"
                className={`px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                  isOtherResourcesPage
                    ? 'bg-gray-700 text-white'
                    : 'text-gray-300 hover:bg-gray-700 hover:text-white'
                }`}
              >
                Other Resources
              </Link>
            </div>
          </div>

          {/* Mobile menu button */}
          <div className="md:hidden">
            <button
              onClick={() => setIsMenuOpen(!isMenuOpen)}
              className="inline-flex items-center justify-center p-2 rounded-md text-gray-300 hover:text-white hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-white"
              aria-expanded="false"
            >
              <span className="sr-only">Open main menu</span>
              {!isMenuOpen ? (
                <svg
                  className="block h-6 w-6"
                  xmlns="http://www.w3.org/2000/svg"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M4 6h16M4 12h16M4 18h16"
                  />
                </svg>
              ) : (
                <svg
                  className="block h-6 w-6"
                  xmlns="http://www.w3.org/2000/svg"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M6 18L18 6M6 6l12 12"
                  />
                </svg>
              )}
            </button>
          </div>
        </div>
      </div>

      {/* Mobile menu */}
      {isMenuOpen && (
        <div className="md:hidden">
          <div className="px-2 pt-2 pb-3 space-y-1 sm:px-3 bg-gray-800">
            <Link
              to="/prizes"
              onClick={() => setIsMenuOpen(false)}
              className={`block w-full text-left px-3 py-2 rounded-md text-base font-medium transition-colors ${
                isPrizesPage
                  ? 'bg-gray-700 text-white'
                  : 'text-gray-300 hover:bg-gray-700 hover:text-white'
              }`}
            >
              Prizes
            </Link>
            <Link
              to="/draw-supporters"
              onClick={() => setIsMenuOpen(false)}
              className={`block w-full text-left px-3 py-2 rounded-md text-base font-medium transition-colors ${
                isDrawSupportersPage
                  ? 'bg-gray-700 text-white'
                  : 'text-gray-300 hover:bg-gray-700 hover:text-white'
              }`}
            >
              Draw Supporters
            </Link>
            <Link
              to="/start"
              onClick={() => setIsMenuOpen(false)}
              className={`block w-full text-left px-3 py-2 rounded-md text-base font-medium transition-colors ${
                isStartPage
                  ? 'bg-gray-700 text-white'
                  : 'text-gray-300 hover:bg-gray-700 hover:text-white'
              }`}
            >
              Starting Hand
            </Link>
            <Link
              to="/matchups"
              onClick={() => setIsMenuOpen(false)}
              className={`block w-full text-left px-3 py-2 rounded-md text-base font-medium transition-colors ${
                isMatchupsPage
                  ? 'bg-gray-700 text-white'
                  : 'text-gray-300 hover:bg-gray-700 hover:text-white'
              }`}
            >
              Matchups
            </Link>
            <div className="pt-2 mt-2 border-t border-gray-700">
              <Link
                to="/other-resources"
                onClick={() => setIsMenuOpen(false)}
                className={`block w-full text-left px-3 py-2 rounded-md text-base font-medium transition-colors ${
                  isOtherResourcesPage
                    ? 'bg-gray-700 text-white'
                    : 'text-gray-300 hover:bg-gray-700 hover:text-white'
                }`}
              >
                Other Resources
              </Link>
            </div>
          </div>
        </div>
      )}
    </nav>
  );
}

