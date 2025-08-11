import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { Edit, Trash2, Save, X, Plus, Eye } from 'lucide-react';

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
  id: string;
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

const DataEditor: React.FC = () => {
  const { datasetId } = useParams<{ datasetId: string }>();
  const [schema, setSchema] = useState<DatasetSchema | null>(null);
  const [dataResult, setDataResult] = useState<DataPageResult | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(50);
  const [editingRow, setEditingRow] = useState<number | null>(null);
  const [editingData, setEditingData] = useState<DataRow>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>('');
  const [showSchemaEditor, setShowSchemaEditor] = useState(false);

  useEffect(() => {
    if (datasetId) {
      loadSchema();
      loadData();
    }
  }, [datasetId, currentPage, pageSize]);

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
      } else if (response.status === 404) {
        // No schema exists yet, user can create one
        setSchema(null);
      } else {
        throw new Error('Failed to load schema');
      }
    } catch (err) {
      console.error('Error loading schema:', err);
      setError('Failed to load schema');
    }
  };

  const loadData = async () => {
    try {
      setLoading(true);
      const token = localStorage.getItem('token');
      const response = await fetch(
        `/api/data/dataset/${datasetId}?page=${currentPage}&page_size=${pageSize}`,
        {
          headers: {
            'Authorization': `Bearer ${token}`,
          },
        }
      );

      if (response.ok) {
        const result = await response.json();
        setDataResult(result);
      } else {
        throw new Error('Failed to load data');
      }
    } catch (err) {
      console.error('Error loading data:', err);
      setError('Failed to load data');
    } finally {
      setLoading(false);
    }
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

  if (!schema) {
    return (
      <div className="text-center py-8">
        <h3 className="text-lg font-medium text-gray-900 mb-4">No Schema Defined</h3>
        <p className="text-gray-600 mb-6">
          This dataset doesn't have a schema yet. You need to define one to view and edit the data.
        </p>
        <button
          onClick={() => setShowSchemaEditor(true)}
          className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
        >
          Create Schema
        </button>
      </div>
    );
  }

  const { data, total, page, total_pages } = dataResult || { data: [], total: 0, page: 1, total_pages: 1 };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">{schema.name}</h2>
          {schema.description && (
            <p className="text-gray-600 mt-1">{schema.description}</p>
          )}
        </div>
        <div className="flex space-x-2">
          <button
            onClick={() => setShowSchemaEditor(true)}
            className="flex items-center px-3 py-2 bg-gray-100 text-gray-700 rounded hover:bg-gray-200"
          >
            <Eye className="w-4 h-4 mr-2" />
            View Schema
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
                {schema.fields
                  .sort((a, b) => a.position - b.position)
                  .map((field) => (
                    <th
                      key={field.id}
                      className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                    >
                      {field.display_name}
                      {field.is_required && <span className="text-red-500 ml-1">*</span>}
                    </th>
                  ))}
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {data.map((row, rowIndex) => (
                <tr key={rowIndex} className="hover:bg-gray-50">
                  {schema.fields
                    .sort((a, b) => a.position - b.position)
                    .map((field) => (
                      <td key={field.id} className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {editingRow === rowIndex ? (
                          renderFieldInput(field, editingData[field.name])
                        ) : (
                          renderCellValue(field, row[field.name])
                        )}
                      </td>
                    ))}
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
                    onClick={() => setCurrentPage(Math.min(total_pages, currentPage + 1))}
                    disabled={currentPage === total_pages}
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
