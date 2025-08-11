import React, { useState, useEffect } from 'react';
import { X, Plus, Trash2, Save, ArrowUp, ArrowDown } from 'lucide-react';

interface SchemaField {
  id?: string;
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

interface SchemaEditorProps {
  datasetId: string;
  existingSchema?: DatasetSchema;
  onSave: (schema: DatasetSchema) => void;
  onCancel: () => void;
}

const dataTypes = [
  { value: 'string', label: 'Text' },
  { value: 'number', label: 'Number' },
  { value: 'boolean', label: 'Yes/No' },
  { value: 'date', label: 'Date' },
  { value: 'email', label: 'Email' },
];

const SchemaEditor: React.FC<SchemaEditorProps> = ({
  datasetId,
  existingSchema,
  onSave,
  onCancel,
}) => {
  const [schema, setSchema] = useState<DatasetSchema>({
    dataset_id: datasetId,
    name: '',
    description: '',
    fields: [],
  });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string>('');

  useEffect(() => {
    if (existingSchema) {
      setSchema(existingSchema);
    }
  }, [existingSchema]);

  const addField = () => {
    const newField: SchemaField = {
      name: '',
      display_name: '',
      data_type: 'string',
      is_required: false,
      is_unique: false,
      position: schema.fields.length + 1,
    };
    setSchema(prev => ({
      ...prev,
      fields: [...prev.fields, newField],
    }));
  };

  const removeField = (index: number) => {
    setSchema(prev => ({
      ...prev,
      fields: prev.fields.filter((_, i) => i !== index),
    }));
  };

  const updateField = (index: number, updates: Partial<SchemaField>) => {
    setSchema(prev => ({
      ...prev,
      fields: prev.fields.map((field, i) =>
        i === index ? { ...field, ...updates } : field
      ),
    }));
  };

  const moveField = (index: number, direction: 'up' | 'down') => {
    const newFields = [...schema.fields];
    if (direction === 'up' && index > 0) {
      [newFields[index], newFields[index - 1]] = [newFields[index - 1], newFields[index]];
    } else if (direction === 'down' && index < newFields.length - 1) {
      [newFields[index], newFields[index + 1]] = [newFields[index + 1], newFields[index]];
    }
    
    // Update positions
    newFields.forEach((field, i) => {
      field.position = i + 1;
    });
    
    setSchema(prev => ({ ...prev, fields: newFields }));
  };

  const handleSave = async () => {
    // Validation
    if (!schema.name.trim()) {
      setError('Schema name is required');
      return;
    }

    if (schema.fields.length === 0) {
      setError('At least one field is required');
      return;
    }

    for (const field of schema.fields) {
      if (!field.name.trim()) {
        setError('All fields must have a name');
        return;
      }
    }

    // Check for duplicate field names
    const fieldNames = schema.fields.map(f => f.name.toLowerCase());
    if (new Set(fieldNames).size !== fieldNames.length) {
      setError('Field names must be unique');
      return;
    }

    setSaving(true);
    setError('');

    try {
      const token = localStorage.getItem('token');
      const url = existingSchema?.id 
        ? `/api/schemas/${existingSchema.id}`
        : '/api/schemas';
      
      const method = existingSchema?.id ? 'PUT' : 'POST';

      const response = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify(schema),
      });

      if (response.ok) {
        const result = await response.json();
        onSave(result.schema);
      } else {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to save schema');
      }
    } catch (err) {
      console.error('Error saving schema:', err);
      setError(err instanceof Error ? err.message : 'Failed to save schema');
    } finally {
      setSaving(false);
    }
  };

  const generateFieldName = (displayName: string) => {
    return displayName
      .toLowerCase()
      .replace(/[^a-z0-9]/g, '_')
      .replace(/_+/g, '_')
      .replace(/^_|_$/g, '');
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-lg shadow-xl max-w-4xl w-full max-h-[90vh] overflow-hidden">
        <div className="flex justify-between items-center p-6 border-b">
          <h2 className="text-xl font-semibold">
            {existingSchema ? 'Edit Schema' : 'Create Schema'}
          </h2>
          <button
            onClick={onCancel}
            className="text-gray-400 hover:text-gray-600"
          >
            <X className="w-6 h-6" />
          </button>
        </div>

        <div className="p-6 overflow-y-auto max-h-[calc(90vh-140px)]">
          {error && (
            <div className="bg-red-50 border border-red-200 rounded-md p-4 mb-6">
              <p className="text-red-600">{error}</p>
            </div>
          )}

          <div className="space-y-6">
            {/* Schema Details */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Schema Name *
                </label>
                <input
                  type="text"
                  value={schema.name}
                  onChange={(e) => setSchema(prev => ({ ...prev, name: e.target.value }))}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="Enter schema name"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Description
                </label>
                <input
                  type="text"
                  value={schema.description || ''}
                  onChange={(e) => setSchema(prev => ({ ...prev, description: e.target.value }))}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="Optional description"
                />
              </div>
            </div>

            {/* Fields */}
            <div>
              <div className="flex justify-between items-center mb-4">
                <h3 className="text-lg font-medium text-gray-900">Fields</h3>
                <button
                  onClick={addField}
                  className="flex items-center px-3 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
                >
                  <Plus className="w-4 h-4 mr-2" />
                  Add Field
                </button>
              </div>

              <div className="space-y-4">
                {schema.fields.map((field, index) => (
                  <div key={index} className="border border-gray-200 rounded-lg p-4">
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          Display Name *
                        </label>
                        <input
                          type="text"
                          value={field.display_name}
                          onChange={(e) => {
                            const displayName = e.target.value;
                            updateField(index, { 
                              display_name: displayName,
                              name: field.name || generateFieldName(displayName)
                            });
                          }}
                          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                          placeholder="Field display name"
                        />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          Field Name *
                        </label>
                        <input
                          type="text"
                          value={field.name}
                          onChange={(e) => updateField(index, { name: e.target.value })}
                          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                          placeholder="field_name"
                        />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          Data Type *
                        </label>
                        <select
                          value={field.data_type}
                          onChange={(e) => updateField(index, { data_type: e.target.value as any })}
                          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                        >
                          {dataTypes.map(type => (
                            <option key={type.value} value={type.value}>
                              {type.label}
                            </option>
                          ))}
                        </select>
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          Default Value
                        </label>
                        <input
                          type="text"
                          value={field.default_value || ''}
                          onChange={(e) => updateField(index, { default_value: e.target.value })}
                          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                          placeholder="Optional default value"
                        />
                      </div>
                      <div className="flex items-center space-x-4">
                        <label className="flex items-center">
                          <input
                            type="checkbox"
                            checked={field.is_required}
                            onChange={(e) => updateField(index, { is_required: e.target.checked })}
                            className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                          />
                          <span className="ml-2 text-sm text-gray-700">Required</span>
                        </label>
                        <label className="flex items-center">
                          <input
                            type="checkbox"
                            checked={field.is_unique}
                            onChange={(e) => updateField(index, { is_unique: e.target.checked })}
                            className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                          />
                          <span className="ml-2 text-sm text-gray-700">Unique</span>
                        </label>
                      </div>
                      <div className="flex items-center space-x-2">
                        <button
                          onClick={() => moveField(index, 'up')}
                          disabled={index === 0}
                          className="p-1 text-gray-400 hover:text-gray-600 disabled:opacity-50"
                        >
                          <ArrowUp className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => moveField(index, 'down')}
                          disabled={index === schema.fields.length - 1}
                          className="p-1 text-gray-400 hover:text-gray-600 disabled:opacity-50"
                        >
                          <ArrowDown className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => removeField(index)}
                          className="p-1 text-red-400 hover:text-red-600"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </div>
                    </div>
                  </div>
                ))}
              </div>

              {schema.fields.length === 0 && (
                <div className="text-center py-8 text-gray-500">
                  No fields defined. Click "Add Field" to get started.
                </div>
              )}
            </div>
          </div>
        </div>

        <div className="flex justify-end space-x-3 p-6 border-t">
          <button
            onClick={onCancel}
            className="px-4 py-2 border border-gray-300 rounded text-gray-700 hover:bg-gray-50"
          >
            Cancel
          </button>
          <button
            onClick={handleSave}
            disabled={saving}
            className="flex items-center px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
          >
            {saving ? (
              <>
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                Saving...
              </>
            ) : (
              <>
                <Save className="w-4 h-4 mr-2" />
                Save Schema
              </>
            )}
          </button>
        </div>
      </div>
    </div>
  );
};

export default SchemaEditor;
