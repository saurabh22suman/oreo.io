import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { Edit, Trash2, Save, X, Eye } from 'lucide-react';
import SQLQuery from './SQLQuery';

interface SchemaField {
  id: string;
  name: string;
  display_name: string;
  data_type: 'string' | 'number' | 'boolean' | 'date' | 'email';
  is_required: boolean;
  is_unique: boolean;
  default_value?: string;
  position: number;
  validation?: any;
}

interface DatasetSchema {
  id?: string;
  dataset_id: string;
  name: string;
  description?: string;
  fields: SchemaField[];
}

interface DataRow {
  [key: string]: any;
}

interface DataPageResult {
  data: DataRow[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

interface DataEditorProps {
  onOpenSchemaEditor?: () => void;
  schema?: DatasetSchema | null;
  onSchemaChange?: (schema: DatasetSchema | null) => void;
}

const DataEditor: React.FC<DataEditorProps> = ({ onOpenSchemaEditor, schema: propSchema, onSchemaChange }) => {
  const { datasetId } = useParams<{ datasetId: string }>();
  const [schema, setSchema] = useState<DatasetSchema | null>(propSchema || null);
  const [dataResult, setDataResult] = useState<DataPageResult | null>(null);
  const [queryResult, setQueryResult] = useState<DataPageResult | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(50);
  const [editingRow, setEditingRow] = useState<number | null>(null);
  const [editingData, setEditingData] = useState<DataRow>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>('');
  const [isQueryMode, setIsQueryMode] = useState(false);
  const [isPreviewMode, setIsPreviewMode] = useState(true); // New: Start in preview mode

  // Maximum rows allowed (backend enforces 1000 limit)
  const MAX_ROWS = 1000;
  const PREVIEW_ROWS = 100; // New: Preview mode shows only 100 rows

  useEffect(() => {
    if (datasetId) {
      loadSchema();
      loadData(); // Load data immediately regardless of schema
    }
  }, [datasetId]);

  useEffect(() => {
    if (propSchema !== undefined) {
      setSchema(propSchema);
    }
  }, [propSchema]);

  useEffect(() => {
    if (datasetId) {
      loadData(); // Load data regardless of schema
    }
  }, [datasetId, currentPage, pageSize, isPreviewMode]);

  const loadSchema = async () => {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`/api/schemas/dataset/${datasetId}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const result = await response.json();
        setSchema(result.schema);
        onSchemaChange?.(result.schema);
      } else if (response.status === 404) {
        // No schema exists yet, user can create one
        setSchema(null);
        onSchemaChange?.(null);
      } else {
        // Schema loading failed, but we can still show data without schema
        console.warn('Failed to load schema, continuing without schema');
        setSchema(null);
        onSchemaChange?.(null);
      }
    } catch (err) {
      console.error('Error loading schema:', err);
      // Don't set error state for schema loading failures
      // We can still show data without schema
      setSchema(null);
      onSchemaChange?.(null);
    }
  };

  const loadData = async () => {
    try {
      console.log('[DEBUG] DataEditor loadData: Starting for dataset ID:', datasetId, 'Preview mode:', isPreviewMode);
      setLoading(true);
      
      const token = localStorage.getItem('accessToken');
      console.log('[DEBUG] DataEditor loadData: Token exists:', !!token);
      
      // In preview mode, limit to first 100 rows (5 pages of 20 each)
      let actualPageSize = pageSize;
      let actualPage = currentPage;
      let maxRows = MAX_ROWS;
      
      if (isPreviewMode) {
        maxRows = PREVIEW_ROWS;
        actualPageSize = Math.min(pageSize, 20); // Smaller page size for preview
        const maxPreviewPage = Math.ceil(PREVIEW_ROWS / actualPageSize);
        actualPage = Math.min(currentPage, maxPreviewPage);
      } else {
        const maxPage = Math.ceil(MAX_ROWS / pageSize);
        actualPage = Math.min(currentPage, maxPage);
      }
      
      console.log('[DEBUG] DataEditor loadData: Mode calculations - page:', actualPage, 'pageSize:', actualPageSize, 'maxRows:', maxRows);
      
      const apiUrl = `${import.meta.env.VITE_API_URL || 'http://localhost:8080'}/api/v1/data/dataset/${datasetId}?page=${actualPage}&page_size=${actualPageSize}`;
      console.log('[DEBUG] DataEditor loadData: Calling API:', apiUrl);
      
      const response = await fetch(apiUrl, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      console.log('[DEBUG] DataEditor loadData: Response status:', response.status);

      if (response.ok) {
        const result = await response.json();
        console.log('[DEBUG] DataEditor loadData: Data loaded:', result);
        
        // Apply limits based on mode
        if (isPreviewMode) {
          // In preview mode, limit to PREVIEW_ROWS
          if (result.total > PREVIEW_ROWS) {
            result.total = PREVIEW_ROWS;
            result.total_pages = Math.ceil(PREVIEW_ROWS / actualPageSize);
          }
        } else {
          // In full mode, limit to MAX_ROWS
          if (result.total > MAX_ROWS) {
            result.total = MAX_ROWS;
            result.total_pages = Math.ceil(MAX_ROWS / pageSize);
          }
        }
        
        setDataResult(result);
        setIsQueryMode(false);
        setError('');
      } else if (response.status === 404) {
        console.log('[DEBUG] DataEditor loadData: No data exists yet (404)');
        // No data exists yet, show empty result
        setDataResult({ data: [], total: 0, page: 1, page_size: actualPageSize, total_pages: 0 });
        setIsQueryMode(false);
        setError('');
      } else {
        const errorText = await response.text();
        console.error('[ERROR] DataEditor loadData: Failed to load data:', response.status, errorText);
        setError('Failed to load data');
      }
    } catch (err) {
      console.error('[ERROR] DataEditor loadData: Exception:', err);
      setError('Failed to load data');
    } finally {
      setLoading(false);
    }
  };

  const handleQueryResult = (result: DataPageResult) => {
    setQueryResult(result);
    setIsQueryMode(true);
    setError(''); // Clear any previous errors
  };

  const clearQuery = () => {
    setQueryResult(null);
    setIsQueryMode(false);
    loadData();
  };

  const toggleToFullData = () => {
    setIsPreviewMode(false);
    setCurrentPage(1); // Reset to first page
    setPageSize(50); // Use normal page size
  };

  const toggleToPreview = () => {
    setIsPreviewMode(true);
    setCurrentPage(1); // Reset to first page
    setPageSize(20); // Use smaller page size for preview
  };

  const handleEdit = (rowIndex: number, rowData: DataRow) => {
    setEditingRow(rowIndex);
    setEditingData({ ...rowData });
  };

  const handleSave = async () => {
    if (editingRow === null) return;

    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`/api/data/dataset/${datasetId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          row_index: editingRow,
          data: editingData,
        }),
      });

      if (response.ok) {
        setEditingRow(null);
        setEditingData({});
        loadData(); // Reload data
      } else {
        throw new Error('Failed to save data');
      }
    } catch (err) {
      console.error('Error saving data:', err);
      setError('Failed to save data');
    }
  };

  const handleCancel = () => {
    setEditingRow(null);
    setEditingData({});
  };

  const handleDelete = async (rowIndex: number) => {
    if (!window.confirm('Are you sure you want to delete this row?')) return;

    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`/api/data/dataset/${datasetId}/row/${rowIndex}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        loadData(); // Reload data
      } else {
        throw new Error('Failed to delete data');
      }
    } catch (err) {
      console.error('Error deleting data:', err);
      setError('Failed to delete data');
    }
  };

  const handleFieldChange = (fieldName: string, value: any) => {
    setEditingData(prev => ({
      ...prev,
      [fieldName]: value,
    }));
  };

  const renderFieldInput = (field: SchemaField, value: any) => {
    const commonProps = {
      value: value || '',
      onChange: (e: React.ChangeEvent<HTMLInputElement>) =>
        handleFieldChange(field.name, e.target.value),
      className: 'w-full px-2 py-1 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500',
    };

    switch (field.data_type) {
      case 'number':
        return <input type="number" {...commonProps} />;
      case 'date':
        return <input type="date" {...commonProps} />;
      case 'email':
        return <input type="email" {...commonProps} />;
      case 'boolean':
        return (
          <select
            value={value ? 'true' : 'false'}
            onChange={(e) => handleFieldChange(field.name, e.target.value === 'true')}
            className="w-full px-2 py-1 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="true">True</option>
            <option value="false">False</option>
          </select>
        );
      default:
        return <input type="text" {...commonProps} />;
    }
  };

  const renderCellValue = (field: SchemaField, value: any) => {
    if (value === null || value === undefined) return '-';
    
    switch (field.data_type) {
      case 'boolean':
        return value ? 'Yes' : 'No';
      case 'date':
        return new Date(value).toLocaleDateString();
      default:
        return String(value);
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-md p-4">
        <p className="text-red-600">{error}</p>
      </div>
    );
  }

  // Remove the early return for missing schema - we can show data without schema
  const displayResult = isQueryMode ? queryResult : dataResult;
  const { data, total, total_pages } = displayResult || { data: [], total: 0, total_pages: 1 };

  return (
    <div className="space-y-6">
      {/* SQL Query Interface */}
      {datasetId && (
        <SQLQuery 
          datasetId={datasetId} 
          onQueryResult={handleQueryResult}
        />
      )}

      {/* Query Results Header */}
      {isQueryMode && queryResult && (
        <div className="bg-blue-50 border border-blue-200 rounded-md p-4">
          <div className="flex justify-between items-center">
            <div>
              <h3 className="text-lg font-medium text-blue-900">Search Results</h3>
              <p className="text-blue-700">Found {queryResult.total} matching rows</p>
            </div>
            <button
              onClick={clearQuery}
              className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
            >
              Clear Search
            </button>
          </div>
        </div>
      )}

      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">{schema?.name || 'Dataset Viewer'}</h2>
          {schema?.description && (
            <p className="text-gray-600 mt-1">{schema.description}</p>
          )}
        </div>
        <div className="flex space-x-2">
          {/* View Mode Toggle Buttons */}
          {isPreviewMode ? (
            <button
              onClick={toggleToFullData}
              className="flex items-center px-3 py-2 bg-green-600 text-white rounded hover:bg-green-700"
            >
              <Eye className="w-4 h-4 mr-2" />
              Show Full Data
            </button>
          ) : (
            <button
              onClick={toggleToPreview}
              className="flex items-center px-3 py-2 bg-orange-600 text-white rounded hover:bg-orange-700"
            >
              <Eye className="w-4 h-4 mr-2" />
              Show Preview (100 rows)
            </button>
          )}
          
          <button
            onClick={() => onOpenSchemaEditor?.()}
            className={`flex items-center px-3 py-2 rounded ${
              schema 
                ? "bg-gray-100 text-gray-700 hover:bg-gray-200" 
                : "bg-purple-600 text-white hover:bg-purple-700"
            }`}
          >
            <Eye className="w-4 h-4 mr-2" />
            {schema ? "View Schema" : "Create Schema"}
          </button>
          <button
            onClick={loadData}
            className="flex items-center px-3 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
          >
            Refresh
          </button>
        </div>
      </div>

      {/* Data Table */}
      <div className="bg-white shadow overflow-hidden sm:rounded-md">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                {schema ? (
                  // When schema exists, use schema fields
                  schema.fields
                    .sort((a, b) => a.position - b.position)
                    .map((field) => (
                      <th
                        key={field.id}
                        className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                      >
                        {field.display_name}
                        {field.is_required && <span className="text-red-500 ml-1">*</span>}
                      </th>
                    ))
                ) : (
                  // When no schema, use column names from data
                  data.length > 0 && Object.keys(data[0])
                    .filter(key => key !== '_row_index') // Exclude internal row index
                    .map((columnName) => (
                      <th
                        key={columnName}
                        className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                      >
                        {columnName}
                      </th>
                    ))
                )}
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {data.map((row, rowIndex) => (
                <tr key={rowIndex} className="hover:bg-gray-50">
                  {schema ? (
                    // When schema exists, use schema fields
                    schema.fields
                      .sort((a, b) => a.position - b.position)
                      .map((field) => (
                        <td key={field.id} className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          {editingRow === rowIndex ? (
                            renderFieldInput(field, editingData[field.name])
                          ) : (
                            renderCellValue(field, row[field.name])
                          )}
                        </td>
                      ))
                  ) : (
                    // When no schema, display all columns
                    Object.keys(row)
                      .filter(key => key !== '_row_index') // Exclude internal row index
                      .map((columnName) => (
                        <td key={columnName} className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          {row[columnName]?.toString() || ''}
                        </td>
                      ))
                  )}
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    {editingRow === rowIndex ? (
                      <div className="flex justify-end space-x-2">
                        <button
                          onClick={handleSave}
                          className="text-green-600 hover:text-green-900"
                        >
                          <Save className="w-4 h-4" />
                        </button>
                        <button
                          onClick={handleCancel}
                          className="text-gray-600 hover:text-gray-900"
                        >
                          <X className="w-4 h-4" />
                        </button>
                      </div>
                    ) : (
                      <div className="flex justify-end space-x-2">
                        <button
                          onClick={() => handleEdit(rowIndex, row)}
                          className="text-indigo-600 hover:text-indigo-900"
                        >
                          <Edit className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => handleDelete(rowIndex)}
                          className="text-red-600 hover:text-red-900"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </div>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {/* Pagination */}
        {total_pages > 1 && (
          <div className="bg-white px-4 py-3 flex items-center justify-between border-t border-gray-200 sm:px-6">
            <div className="flex-1 flex justify-between sm:hidden">
              <button
                onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
                disabled={currentPage === 1}
                className="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50"
              >
                Previous
              </button>
              <button
                onClick={() => setCurrentPage(Math.min(total_pages, currentPage + 1))}
                disabled={currentPage === total_pages}
                className="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50"
              >
                Next
              </button>
            </div>
            <div className="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
              <div>
                <p className="text-sm text-gray-700">
                  Showing <span className="font-medium">{(currentPage - 1) * pageSize + 1}</span> to{' '}
                  <span className="font-medium">
                    {Math.min(currentPage * pageSize, total)}
                  </span>{' '}
                  of <span className="font-medium">{total}</span> results
                  {isPreviewMode && (
                    <span className="text-blue-600 ml-2">(Preview mode - showing first {PREVIEW_ROWS} rows)</span>
                  )}
                  {!isQueryMode && !isPreviewMode && total >= MAX_ROWS && (
                    <span className="text-orange-600 ml-2">(Limited to {MAX_ROWS} rows)</span>
                  )}
                </p>
              </div>
              <div>
                <nav className="relative z-0 inline-flex rounded-md shadow-sm -space-x-px">
                  <button
                    onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
                    disabled={currentPage === 1}
                    className="relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50"
                  >
                    Previous
                  </button>
                  <span className="relative inline-flex items-center px-4 py-2 border border-gray-300 bg-white text-sm font-medium text-gray-700">
                    {currentPage} / {total_pages}
                  </span>
                  <button
                    onClick={() => {
                      const maxPage = isQueryMode ? total_pages : Math.min(total_pages, Math.ceil(MAX_ROWS / pageSize));
                      setCurrentPage(Math.min(maxPage, currentPage + 1));
                    }}
                    disabled={currentPage === total_pages || (!isQueryMode && currentPage >= Math.ceil(MAX_ROWS / pageSize))}
                    className="relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50"
                  >
                    Next
                  </button>
                </nav>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default DataEditor;
