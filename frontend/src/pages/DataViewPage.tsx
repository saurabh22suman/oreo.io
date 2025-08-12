import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { ArrowLeft, Database, Settings, Sparkles } from 'lucide-react';
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
      console.log('[DEBUG] loadDataset: Starting for dataset ID:', datasetId);
      
      const token = localStorage.getItem('accessToken'); // Changed from 'token' to 'accessToken'
      console.log('[DEBUG] loadDataset: Token exists:', !!token);
      
      const apiUrl = `${import.meta.env.VITE_API_URL || 'http://localhost:8080'}/api/v1/datasets/${datasetId}`;
      console.log('[DEBUG] loadDataset: Calling API:', apiUrl);
      
      const response = await fetch(apiUrl, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      console.log('[DEBUG] loadDataset: Response status:', response.status);
      
      if (response.ok) {
        const dataset = await response.json();
        console.log('[DEBUG] loadDataset: Dataset loaded:', dataset);
        setDataset(dataset);
      } else {
        const errorText = await response.text();
        console.error('[ERROR] loadDataset: Failed to load dataset:', response.status, errorText);
        setError('Failed to load dataset information');
      }
    } catch (error) {
      console.error('[ERROR] loadDataset: Exception:', error);
      setError('Failed to load dataset information');
    } finally {
      setLoading(false);
    }
  };

  const loadSchema = async () => {
    try {
      console.log('[DEBUG] loadSchema: Starting for dataset ID:', datasetId);
      
      const token = localStorage.getItem('accessToken'); // Changed from 'token' to 'accessToken'
      console.log('[DEBUG] loadSchema: Token exists:', !!token);
      
      const apiUrl = `${import.meta.env.VITE_API_URL || 'http://localhost:8080'}/api/v1/schemas/dataset/${datasetId}`;
      console.log('[DEBUG] loadSchema: Calling API:', apiUrl);
      
      const response = await fetch(apiUrl, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      console.log('[DEBUG] loadSchema: Response status:', response.status);

      if (response.ok) {
        const result = await response.json();
        console.log('[DEBUG] loadSchema: Schema loaded:', result);
        setSchema(result.schema);
      } else if (response.status !== 404) {
        // 404 is expected if no schema exists yet
        const errorText = await response.text();
        console.error('[ERROR] loadSchema: Failed to load schema:', response.status, errorText);
      } else {
        console.log('[DEBUG] loadSchema: No schema found (404) - this is expected');
      }
    } catch (err) {
      console.error('[ERROR] loadSchema: Exception:', err);
      // Don't set error for schema loading failure
    }
  };

  const handleSchemaSaved = (savedSchema: DatasetSchema) => {
    setSchema(savedSchema);
    setShowSchemaEditor(false);
  };

  const inferSchema = async () => {
    try {
      console.log('[DEBUG] inferSchema: Starting for dataset ID:', datasetId);
      
      const token = localStorage.getItem('accessToken');
      if (!token) {
        console.error('[ERROR] inferSchema: No token found');
        return;
      }
      
      const apiUrl = `${import.meta.env.VITE_API_URL || 'http://localhost:8080'}/api/v1/schemas/infer/${datasetId}`;
      console.log('[DEBUG] inferSchema: Calling API:', apiUrl);
      
      setLoading(true);
      
      const response = await fetch(apiUrl, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      console.log('[DEBUG] inferSchema: Response status:', response.status);

      if (response.ok) {
        const result = await response.json();
        console.log('[DEBUG] inferSchema: Schema inferred:', result);
        
        // Transform the inferred schema to match our schema format
        const inferredSchema: DatasetSchema = {
          id: '', // Will be set when saved
          dataset_id: datasetId!,
          name: result.inferred_schema.name,
          description: result.inferred_schema.description,
          fields: result.inferred_schema.fields.map((field: any, index: number) => ({
            id: '',
            schema_id: '',
            name: field.name,
            display_name: field.display_name,
            data_type: field.data_type,
            is_required: field.is_required,
            is_unique: false,
            default_value: null,
            position: index + 1,
            validation: {
              min_length: field.constraints?.min_length,
              max_length: field.constraints?.max_length,
              min_value: field.constraints?.min,
              max_value: field.constraints?.max,
              pattern: field.pattern,
              format: field.constraints?.format,
            },
          })),
        };
        
        // Set the inferred schema and open the editor for review
        setSchema(inferredSchema);
        setShowSchemaEditor(true);
        
      } else {
        const errorText = await response.text();
        console.error('[ERROR] inferSchema: Failed to infer schema:', response.status, errorText);
        alert('Failed to infer schema. Please try again.');
      }
    } catch (err) {
      console.error('[ERROR] inferSchema: Exception:', err);
      alert('An error occurred while inferring the schema. Please try again.');
    } finally {
      setLoading(false);
    }
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
                  {!schema ? (
                    <>
                      <button
                        onClick={inferSchema}
                        disabled={loading}
                        className="flex items-center px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        <Sparkles className="w-4 h-4 mr-2" />
                        {loading ? 'Inferring...' : 'Infer Schema'}
                      </button>
                      <button
                        onClick={() => setShowSchemaEditor(true)}
                        className="flex items-center px-4 py-2 bg-gray-100 text-gray-700 rounded hover:bg-gray-200"
                      >
                        <Settings className="w-4 h-4 mr-2" />
                        Create Schema Manually
                      </button>
                    </>
                  ) : (
                    <button
                      onClick={() => setShowSchemaEditor(true)}
                      className="flex items-center px-4 py-2 bg-gray-100 text-gray-700 rounded hover:bg-gray-200"
                    >
                      <Settings className="w-4 h-4 mr-2" />
                      Edit Schema
                    </button>
                  )}
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Main Content */}
        <div className="bg-white shadow rounded-lg">
          <div className="p-6">
            {datasetId && (
              <DataEditor 
                onOpenSchemaEditor={() => setShowSchemaEditor(true)}
                schema={schema}
                onSchemaChange={setSchema}
              />
            )}
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
