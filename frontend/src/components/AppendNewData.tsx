import React, { useState, useEffect } from 'react';
import { Upload, FileText, AlertCircle, CheckCircle, Eye, Edit3, Save, X } from 'lucide-react';

interface DataSubmission {
  id: string;
  dataset_id: string;
  submitted_by: string;
  file_name: string;
  file_size: number;
  row_count: number;
  status: 'pending' | 'under_review' | 'approved' | 'rejected' | 'applied';
  validation_results?: ValidationResult;
  admin_notes?: string;
  submitted_at: string;
  dataset_name: string;
  project_name: string;
}

interface ValidationResult {
  is_valid: boolean;
  total_rows: number;
  valid_rows: number;
  invalid_rows: number;
  warning_rows: number;
  schema_errors: ValidationError[];
  business_rule_errors: ValidationError[];
  field_stats: Record<string, FieldStats>;
}

interface ValidationError {
  row_index: number;
  field_name: string;
  error_type: string;
  message: string;
  actual_value: string;
  expected_value?: string;
}

interface FieldStats {
  total_values: number;
  unique_values: number;
  null_values: number;
  invalid_values: number;
}

interface StagingData {
  id: string;
  submission_id: string;
  row_index: number;
  data: Record<string, any>;
  validation_status: 'valid' | 'invalid' | 'warning';
  validation_errors?: ValidationError[];
}

interface AppendNewDataProps {
  datasetId: string;
  datasetName: string;
  onClose: () => void;
}

const AppendNewData: React.FC<AppendNewDataProps> = ({ datasetId, datasetName, onClose }) => {
  const [currentStep, setCurrentStep] = useState<'upload' | 'validate' | 'edit' | 'submit'>('upload');
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [submission, setSubmission] = useState<DataSubmission | null>(null);
  const [stagingData, setStagingData] = useState<StagingData[]>([]);
  const [editingRow, setEditingRow] = useState<number | null>(null);
  const [editingData, setEditingData] = useState<Record<string, any>>({});
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize] = useState(20);

  // Upload file and validate
  const handleFileUpload = async () => {
    if (!selectedFile) return;

    setIsUploading(true);
    try {
      const formData = new FormData();
      formData.append('file', selectedFile);

      const token = localStorage.getItem('accessToken');
      const response = await fetch(`/api/v1/datasets/${datasetId}/append`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
        body: formData,
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || 'Failed to upload file');
      }

      const result = await response.json();
      setSubmission(result.submission);
      setCurrentStep('validate');
      
      // Load staging data
      await loadStagingData(result.submission.id);
    } catch (error) {
      console.error('Upload error:', error);
      alert(`Upload failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
    } finally {
      setIsUploading(false);
    }
  };

  // Load staging data for editing
  const loadStagingData = async (submissionId: string) => {
    try {
      const token = localStorage.getItem('accessToken');
      const response = await fetch(
        `/api/v1/submissions/${submissionId}/details?page=${currentPage}&page_size=${pageSize}`,
        {
          headers: {
            'Authorization': `Bearer ${token}`,
          },
        }
      );

      if (!response.ok) {
        throw new Error('Failed to load staging data');
      }

      const result = await response.json();
      setStagingData(result.staging_data || []);
    } catch (error) {
      console.error('Error loading staging data:', error);
    }
  };

  // Update staging data row (live editing)
  const updateStagingRow = async (stagingId: string, data: Record<string, any>) => {
    try {
      const token = localStorage.getItem('accessToken');
      const response = await fetch(`/api/v1/staging/${stagingId}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ data }),
      });

      if (!response.ok) {
        throw new Error('Failed to update row');
      }

      // Reload staging data to get updated validation status
      if (submission) {
        await loadStagingData(submission.id);
      }
    } catch (error) {
      console.error('Error updating row:', error);
      alert(`Update failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  };

  // Submit for admin review
  const submitForReview = async () => {
    if (!submission) return;

    try {
      const token = localStorage.getItem('accessToken');
      const response = await fetch(`/api/v1/admin/submissions/${submission.id}/review`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          status: 'under_review',
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to submit for review');
      }

      alert('Data submitted for admin review successfully!');
      onClose();
    } catch (error) {
      console.error('Error submitting for review:', error);
      alert(`Submission failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  };

  const handleEditRow = (rowIndex: number, rowData: Record<string, any>) => {
    setEditingRow(rowIndex);
    setEditingData({ ...rowData });
  };

  const handleSaveEdit = async () => {
    if (editingRow === null) return;

    const stagingRow = stagingData.find(row => row.row_index === editingRow);
    if (!stagingRow) return;

    await updateStagingRow(stagingRow.id, editingData);
    setEditingRow(null);
    setEditingData({});
  };

  const handleCancelEdit = () => {
    setEditingRow(null);
    setEditingData({});
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'valid': return 'text-green-600';
      case 'invalid': return 'text-red-600';
      case 'warning': return 'text-yellow-600';
      default: return 'text-gray-600';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'valid': return <CheckCircle className="w-4 h-4 text-green-600" />;
      case 'invalid': return <AlertCircle className="w-4 h-4 text-red-600" />;
      case 'warning': return <AlertCircle className="w-4 h-4 text-yellow-600" />;
      default: return null;
    }
  };

  useEffect(() => {
    if (submission && currentStep === 'edit') {
      loadStagingData(submission.id);
    }
  }, [currentPage, submission, currentStep]);

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl max-w-6xl w-full max-h-[90vh] overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b">
          <div>
            <h2 className="text-xl font-semibold text-gray-900">Append New Data</h2>
            <p className="text-sm text-gray-600">Dataset: {datasetName}</p>
          </div>
          <button
            onClick={onClose}
            className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Progress Steps */}
        <div className="px-6 py-4 border-b bg-gray-50">
          <div className="flex items-center space-x-4">
            <div className={`flex items-center space-x-2 ${currentStep === 'upload' ? 'text-blue-600' : currentStep === 'validate' || currentStep === 'edit' || currentStep === 'submit' ? 'text-green-600' : 'text-gray-400'}`}>
              <div className={`w-8 h-8 rounded-full flex items-center justify-center ${currentStep === 'upload' ? 'bg-blue-100' : currentStep === 'validate' || currentStep === 'edit' || currentStep === 'submit' ? 'bg-green-100' : 'bg-gray-100'}`}>
                <Upload className="w-4 h-4" />
              </div>
              <span className="text-sm font-medium">Upload</span>
            </div>
            <div className={`h-px flex-1 ${currentStep === 'validate' || currentStep === 'edit' || currentStep === 'submit' ? 'bg-green-600' : 'bg-gray-300'}`} />
            <div className={`flex items-center space-x-2 ${currentStep === 'validate' ? 'text-blue-600' : currentStep === 'edit' || currentStep === 'submit' ? 'text-green-600' : 'text-gray-400'}`}>
              <div className={`w-8 h-8 rounded-full flex items-center justify-center ${currentStep === 'validate' ? 'bg-blue-100' : currentStep === 'edit' || currentStep === 'submit' ? 'bg-green-100' : 'bg-gray-100'}`}>
                <Eye className="w-4 h-4" />
              </div>
              <span className="text-sm font-medium">Validate</span>
            </div>
            <div className={`h-px flex-1 ${currentStep === 'edit' || currentStep === 'submit' ? 'bg-green-600' : 'bg-gray-300'}`} />
            <div className={`flex items-center space-x-2 ${currentStep === 'edit' ? 'text-blue-600' : currentStep === 'submit' ? 'text-green-600' : 'text-gray-400'}`}>
              <div className={`w-8 h-8 rounded-full flex items-center justify-center ${currentStep === 'edit' ? 'bg-blue-100' : currentStep === 'submit' ? 'bg-green-100' : 'bg-gray-100'}`}>
                <Edit3 className="w-4 h-4" />
              </div>
              <span className="text-sm font-medium">Edit</span>
            </div>
            <div className={`h-px flex-1 ${currentStep === 'submit' ? 'bg-green-600' : 'bg-gray-300'}`} />
            <div className={`flex items-center space-x-2 ${currentStep === 'submit' ? 'text-blue-600' : 'text-gray-400'}`}>
              <div className={`w-8 h-8 rounded-full flex items-center justify-center ${currentStep === 'submit' ? 'bg-blue-100' : 'bg-gray-100'}`}>
                <CheckCircle className="w-4 h-4" />
              </div>
              <span className="text-sm font-medium">Submit</span>
            </div>
          </div>
        </div>

        {/* Content */}
        <div className="p-6 overflow-y-auto" style={{ maxHeight: 'calc(90vh - 200px)' }}>
          {currentStep === 'upload' && (
            <div className="space-y-6">
              <div className="text-center">
                <div className="border-2 border-dashed border-gray-300 rounded-lg p-8">
                  <Upload className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                  <div className="space-y-2">
                    <p className="text-lg font-medium text-gray-900">Upload CSV file to append</p>
                    <p className="text-sm text-gray-600">
                      File will be validated against the existing dataset schema and business rules
                    </p>
                  </div>
                  <div className="mt-4">
                    <input
                      type="file"
                      accept=".csv"
                      onChange={(e) => setSelectedFile(e.target.files?.[0] || null)}
                      className="hidden"
                      id="file-upload"
                    />
                    <label
                      htmlFor="file-upload"
                      className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 cursor-pointer"
                    >
                      <FileText className="w-4 h-4 mr-2" />
                      Choose CSV File
                    </label>
                  </div>
                </div>
              </div>

              {selectedFile && (
                <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="font-medium text-blue-900">{selectedFile.name}</p>
                      <p className="text-sm text-blue-700">
                        Size: {(selectedFile.size / 1024 / 1024).toFixed(2)} MB
                      </p>
                    </div>
                    <button
                      onClick={handleFileUpload}
                      disabled={isUploading}
                      className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
                    >
                      {isUploading ? 'Uploading...' : 'Upload & Validate'}
                    </button>
                  </div>
                </div>
              )}
            </div>
          )}

          {currentStep === 'validate' && submission && (
            <div className="space-y-6">
              <div className="bg-white border border-gray-200 rounded-lg p-6">
                <h3 className="text-lg font-medium text-gray-900 mb-4">Validation Results</h3>
                
                {submission.validation_results && (
                  <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
                    <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                      <div className="text-2xl font-bold text-blue-900">
                        {submission.validation_results.total_rows}
                      </div>
                      <div className="text-sm text-blue-700">Total Rows</div>
                    </div>
                    <div className="bg-green-50 border border-green-200 rounded-lg p-4">
                      <div className="text-2xl font-bold text-green-900">
                        {submission.validation_results.valid_rows}
                      </div>
                      <div className="text-sm text-green-700">Valid Rows</div>
                    </div>
                    <div className="bg-red-50 border border-red-200 rounded-lg p-4">
                      <div className="text-2xl font-bold text-red-900">
                        {submission.validation_results.invalid_rows}
                      </div>
                      <div className="text-sm text-red-700">Invalid Rows</div>
                    </div>
                    <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
                      <div className="text-2xl font-bold text-yellow-900">
                        {submission.validation_results.warning_rows}
                      </div>
                      <div className="text-sm text-yellow-700">Warning Rows</div>
                    </div>
                  </div>
                )}

                {submission.validation_results?.schema_errors && submission.validation_results.schema_errors.length > 0 && (
                  <div className="mb-4">
                    <h4 className="font-medium text-red-900 mb-2">Schema Errors</h4>
                    <div className="space-y-2">
                      {submission.validation_results.schema_errors.slice(0, 5).map((error, index) => (
                        <div key={index} className="bg-red-50 border border-red-200 rounded p-3">
                          <div className="text-sm">
                            <span className="font-medium">Row {error.row_index + 1}, Field: {error.field_name}</span>
                            <p className="text-red-700">{error.message}</p>
                          </div>
                        </div>
                      ))}
                      {submission.validation_results.schema_errors.length > 5 && (
                        <p className="text-sm text-gray-600">
                          ...and {submission.validation_results.schema_errors.length - 5} more errors
                        </p>
                      )}
                    </div>
                  </div>
                )}

                <div className="flex justify-between">
                  <button
                    onClick={() => setCurrentStep('upload')}
                    className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50"
                  >
                    Back
                  </button>
                  <button
                    onClick={() => setCurrentStep('edit')}
                    className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
                  >
                    Edit Data
                  </button>
                </div>
              </div>
            </div>
          )}

          {currentStep === 'edit' && submission && (
            <div className="space-y-6">
              <div className="bg-white border border-gray-200 rounded-lg">
                <div className="px-6 py-4 border-b">
                  <h3 className="text-lg font-medium text-gray-900">Edit Data</h3>
                  <p className="text-sm text-gray-600">Click on any cell to edit. Invalid rows are highlighted.</p>
                </div>
                
                <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                      <tr>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Row
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Status
                        </th>
                        {stagingData.length > 0 && Object.keys(stagingData[0].data).map((field) => (
                          <th key={field} className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                            {field}
                          </th>
                        ))}
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Actions
                        </th>
                      </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-gray-200">
                      {stagingData.map((row) => (
                        <tr
                          key={row.id}
                          className={`${
                            row.validation_status === 'invalid' ? 'bg-red-50' :
                            row.validation_status === 'warning' ? 'bg-yellow-50' :
                            'bg-white'
                          }`}
                        >
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            {row.row_index + 1}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap">
                            <div className="flex items-center space-x-1">
                              {getStatusIcon(row.validation_status)}
                              <span className={`text-sm ${getStatusColor(row.validation_status)}`}>
                                {row.validation_status}
                              </span>
                            </div>
                          </td>
                          {Object.entries(row.data).map(([field, value]) => (
                            <td key={field} className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                              {editingRow === row.row_index ? (
                                <input
                                  type="text"
                                  value={editingData[field] || ''}
                                  onChange={(e) => setEditingData({ ...editingData, [field]: e.target.value })}
                                  className="w-full px-2 py-1 border border-gray-300 rounded focus:ring-blue-500 focus:border-blue-500"
                                />
                              ) : (
                                <span
                                  onClick={() => handleEditRow(row.row_index, row.data)}
                                  className="cursor-pointer hover:bg-gray-100 px-2 py-1 rounded"
                                >
                                  {String(value)}
                                </span>
                              )}
                            </td>
                          ))}
                          <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                            {editingRow === row.row_index ? (
                              <div className="flex space-x-2">
                                <button
                                  onClick={handleSaveEdit}
                                  className="text-green-600 hover:text-green-900"
                                >
                                  <Save className="w-4 h-4" />
                                </button>
                                <button
                                  onClick={handleCancelEdit}
                                  className="text-gray-600 hover:text-gray-900"
                                >
                                  <X className="w-4 h-4" />
                                </button>
                              </div>
                            ) : (
                              <button
                                onClick={() => handleEditRow(row.row_index, row.data)}
                                className="text-blue-600 hover:text-blue-900"
                              >
                                <Edit3 className="w-4 h-4" />
                              </button>
                            )}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>

                <div className="px-6 py-4 border-t flex justify-between items-center">
                  <div className="flex space-x-2">
                    <button
                      onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
                      disabled={currentPage === 1}
                      className="px-3 py-1 border border-gray-300 rounded disabled:opacity-50"
                    >
                      Previous
                    </button>
                    <span className="px-3 py-1 text-sm text-gray-600">
                      Page {currentPage}
                    </span>
                    <button
                      onClick={() => setCurrentPage(currentPage + 1)}
                      disabled={stagingData.length < pageSize}
                      className="px-3 py-1 border border-gray-300 rounded disabled:opacity-50"
                    >
                      Next
                    </button>
                  </div>
                  <div className="flex space-x-2">
                    <button
                      onClick={() => setCurrentStep('validate')}
                      className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50"
                    >
                      Back
                    </button>
                    <button
                      onClick={() => setCurrentStep('submit')}
                      className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
                    >
                      Review & Submit
                    </button>
                  </div>
                </div>
              </div>
            </div>
          )}

          {currentStep === 'submit' && submission && (
            <div className="space-y-6">
              <div className="bg-white border border-gray-200 rounded-lg p-6">
                <h3 className="text-lg font-medium text-gray-900 mb-4">Review & Submit</h3>
                
                <div className="space-y-4">
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="text-sm font-medium text-gray-700">File Name</label>
                      <p className="text-sm text-gray-900">{submission.file_name}</p>
                    </div>
                    <div>
                      <label className="text-sm font-medium text-gray-700">Total Rows</label>
                      <p className="text-sm text-gray-900">{submission.row_count}</p>
                    </div>
                  </div>

                  {submission.validation_results && (
                    <div className="grid grid-cols-3 gap-4">
                      <div>
                        <label className="text-sm font-medium text-gray-700">Valid Rows</label>
                        <p className="text-sm text-green-900">{submission.validation_results.valid_rows}</p>
                      </div>
                      <div>
                        <label className="text-sm font-medium text-gray-700">Invalid Rows</label>
                        <p className="text-sm text-red-900">{submission.validation_results.invalid_rows}</p>
                      </div>
                      <div>
                        <label className="text-sm font-medium text-gray-700">Warning Rows</label>
                        <p className="text-sm text-yellow-900">{submission.validation_results.warning_rows}</p>
                      </div>
                    </div>
                  )}

                  <div className="border-t pt-4">
                    <p className="text-sm text-gray-600 mb-4">
                      By submitting this data, you acknowledge that:
                    </p>
                    <ul className="text-sm text-gray-600 space-y-1 list-disc list-inside">
                      <li>The data has been reviewed and edited as necessary</li>
                      <li>Only valid rows will be appended to the target dataset</li>
                      <li>An admin will review this submission before final approval</li>
                      <li>You will be notified of the review decision</li>
                    </ul>
                  </div>
                </div>

                <div className="flex justify-between mt-6">
                  <button
                    onClick={() => setCurrentStep('edit')}
                    className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50"
                  >
                    Back to Edit
                  </button>
                  <button
                    onClick={submitForReview}
                    className="px-6 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700"
                  >
                    Submit for Admin Review
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default AppendNewData;
