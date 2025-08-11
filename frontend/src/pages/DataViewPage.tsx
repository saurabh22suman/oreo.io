import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { ArrowLeft, Database, Settings } from 'lucide-react';
import DataEditor from '../components/DataEditor';
import SchemaEditor from '../components/SchemaEditor';

interface Dataset {
  id: string;
  name: string;
  filename: string;
  file_size: number;
  project_id: string;
  created_at: string;
}

interface DatasetSchema {
  id?: string;
  dataset_id: string;
  name: string;
  description?: string;
  fields: any[];
}

const DataViewPage: React.FC = () => {
  const { datasetId } = useParams<{ datasetId: string }>();
  const navigate = useNavigate();
  const [dataset, setDataset] = useState<Dataset | null>(null);
  const [schema, setSchema] = useState<DatasetSchema | null>(null);
  const [showSchemaEditor, setShowSchemaEditor] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>('');

  useEffect(() => {
    if (datasetId) {
      loadDataset();
      loadSchema();
    }
  }, [datasetId]);

  const loadDataset = async () => {
    try {
      const token = localStorage.getItem('token');
      // We need to get the dataset info - for now we'll create a mock response
      // In a real app, you'd have an endpoint to get dataset by ID
      const response = await fetch('/api/datasets/user', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const datasets = await response.json();
        const foundDataset = datasets.find((d: Dataset) => d.id === datasetId);
        if (foundDataset) {
          setDataset(foundDataset);
        } else {
          setError('Dataset not found');
        }
      } else {
        throw new Error('Failed to load dataset');
      }
    } catch (err) {
      console.error('Error loading dataset:', err);
      setError('Failed to load dataset information');
    } finally {
      setLoading(false);
    }
  };

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
      } else if (response.status !== 404) {
        // 404 is expected if no schema exists yet
        throw new Error('Failed to load schema');
      }
    } catch (err) {
      console.error('Error loading schema:', err);
      // Don't set error for schema loading failure
    }
  };

  const handleSchemaSaved = (savedSchema: DatasetSchema) => {
    setSchema(savedSchema);
    setShowSchemaEditor(false);
  };

  const formatFileSize = (bytes: number) => {
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    if (bytes === 0) return '0 Bytes';
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50">
        <div className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
          <div className="flex justify-center items-center h-64">
            <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-blue-500"></div>
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50">
        <div className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
          <div className="bg-red-50 border border-red-200 rounded-md p-4">
            <p className="text-red-600">{error}</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        {/* Header */}
        <div className="mb-6">
          <nav className="flex" aria-label="Breadcrumb">
            <ol className="flex items-center space-x-4">
              <li>
                <button
                  onClick={() => navigate('/dashboard')}
                  className="flex items-center text-gray-500 hover:text-gray-700"
                >
                  <ArrowLeft className="w-4 h-4 mr-2" />
                  Back to Dashboard
                </button>
              </li>
            </ol>
          </nav>
          
          {dataset && (
            <div className="mt-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center">
                  <Database className="w-8 h-8 text-blue-600 mr-3" />
                  <div>
                    <h1 className="text-2xl font-bold text-gray-900">{dataset.name}</h1>
                    <p className="text-sm text-gray-500">
                      {dataset.filename} • {formatFileSize(dataset.file_size)} • 
                      Created {new Date(dataset.created_at).toLocaleDateString()}
                    </p>
                  </div>
                </div>
                <div className="flex space-x-2">
                  <button
                    onClick={() => setShowSchemaEditor(true)}
                    className="flex items-center px-4 py-2 bg-gray-100 text-gray-700 rounded hover:bg-gray-200"
                  >
                    <Settings className="w-4 h-4 mr-2" />
                    {schema ? 'Edit Schema' : 'Create Schema'}
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Main Content */}
        <div className="bg-white shadow rounded-lg">
          <div className="p-6">
            {datasetId && <DataEditor />}
          </div>
        </div>

        {/* Schema Editor Modal */}
        {showSchemaEditor && datasetId && (
          <SchemaEditor
            datasetId={datasetId}
            existingSchema={schema || undefined}
            onSave={handleSchemaSaved}
            onCancel={() => setShowSchemaEditor(false)}
          />
        )}
      </div>
    </div>
  );
};

export default DataViewPage;
