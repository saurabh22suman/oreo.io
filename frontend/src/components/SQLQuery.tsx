import React, { useState } from 'react';
import { Search, Play, AlertCircle } from 'lucide-react';

interface SQLQueryProps {
  datasetId: string;
  onQueryResult: (result: any) => void;
}

const SQLQuery: React.FC<SQLQueryProps> = ({ datasetId, onQueryResult }) => {
  const [query, setQuery] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [lastQuery, setLastQuery] = useState<string>('');

  const executeQuery = async () => {
    if (!query.trim()) {
      setError('Please enter a search query');
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`/api/v1/data/dataset/${datasetId}/query`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          query: query.trim(),
          page_size: 100
        }),
      });

      if (response.ok) {
        const result = await response.json();
        setLastQuery(query);
        onQueryResult(result);
      } else {
        const errorData = await response.json();
        setError(errorData.error || 'Query execution failed');
      }
    } catch (err) {
      console.error('Error executing query:', err);
      setError('Failed to execute query');
    } finally {
      setIsLoading(false);
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
      executeQuery();
    }
  };

  const clearQuery = () => {
    setQuery('');
    setError(null);
  };

  return (
    <div className="bg-white rounded-lg shadow-sm border p-6 mb-6">
      <div className="flex items-center gap-2 mb-4">
        <Search className="h-5 w-5 text-gray-600" />
        <h3 className="text-lg font-semibold text-gray-900">Search Dataset</h3>
      </div>

      <div className="space-y-4">
        <div>
          <label htmlFor="sql-query" className="block text-sm font-medium text-gray-700 mb-2">
            Search Query
          </label>
          <div className="relative">
            <textarea
              id="sql-query"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              onKeyDown={handleKeyPress}
              placeholder="Enter search terms to filter data... (e.g., 'Delta Air Lines', 'LAX', 'Flight 123')"
              className="w-full h-24 px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 font-mono text-sm resize-none"
              disabled={isLoading}
            />
            <div className="absolute bottom-2 right-2 text-xs text-gray-400">
              Ctrl+Enter to execute
            </div>
          </div>
        </div>

        {error && (
          <div className="flex items-center gap-2 p-3 bg-red-50 border border-red-200 rounded-md">
            <AlertCircle className="h-4 w-4 text-red-500 flex-shrink-0" />
            <span className="text-sm text-red-700">{error}</span>
          </div>
        )}

        <div className="flex items-center gap-3">
          <button
            onClick={executeQuery}
            disabled={isLoading || !query.trim()}
            className="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Play className="h-4 w-4" />
            {isLoading ? 'Searching...' : 'Search'}
          </button>

          <button
            onClick={clearQuery}
            disabled={isLoading}
            className="inline-flex items-center gap-2 px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded-md hover:bg-gray-50 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50"
          >
            Clear
          </button>

          {lastQuery && (
            <span className="text-sm text-gray-500">
              Last search: "{lastQuery}"
            </span>
          )}
        </div>

        <div className="text-xs text-gray-500 bg-gray-50 p-3 rounded-md">
          <strong>Search Tips:</strong>
          <ul className="mt-1 space-y-1">
            <li>• Enter any text to search across all data fields</li>
            <li>• Search is case-insensitive</li>
            <li>• Results are limited to 100 rows max</li>
            <li>• Use Ctrl+Enter to execute your search</li>
          </ul>
        </div>
      </div>
    </div>
  );
};

export default SQLQuery;
