import React, { useState, useEffect } from 'react';
import { CheckCircle, XCircle, Clock, Eye, FileText } from 'lucide-react';

interface DataSubmission {
  id: string;
  dataset_id: string;
  dataset_name: string;
  project_name: string;
  submitter_name: string;
  submitter_email: string;
  file_name: string;
  file_size: number;
  row_count: number;
  status: 'pending' | 'under_review' | 'approved' | 'rejected' | 'applied';
  validation_results?: ValidationResult;
  admin_notes?: string;
  submitted_at: string;
  reviewed_at?: string;
  reviewer_name?: string;
}

interface ValidationResult {
  is_valid: boolean;
  total_rows: number;
  valid_rows: number;
  invalid_rows: number;
  warning_rows: number;
  schema_errors: ValidationError[];
  business_rule_errors: ValidationError[];
}

interface ValidationError {
  row_index: number;
  field_name: string;
  error_type: string;
  message: string;
  actual_value: string;
  expected_value?: string;
}

interface StagingData {
  id: string;
  submission_id: string;
  row_index: number;
  data: Record<string, any>;
  validation_status: 'valid' | 'invalid' | 'warning';
  validation_errors?: ValidationError[];
}

const AdminDataSubmissionReview: React.FC = () => {
  const [submissions, setSubmissions] = useState<DataSubmission[]>([]);
  const [selectedSubmission, setSelectedSubmission] = useState<DataSubmission | null>(null);
  const [stagingData, setStagingData] = useState<StagingData[]>([]);
  const [loading, setLoading] = useState(true);
  const [reviewAction, setReviewAction] = useState<'approve' | 'reject' | null>(null);
  const [adminNotes, setAdminNotes] = useState('');
  const [currentPage] = useState(1);
  const [pageSize] = useState(20);

  useEffect(() => {
    loadPendingSubmissions();
  }, []);

  const loadPendingSubmissions = async () => {
    try {
      setLoading(true);
      const token = localStorage.getItem('accessToken');
      const response = await fetch('/api/v1/admin/submissions/pending', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to load pending submissions');
      }

      const result = await response.json();
      setSubmissions(result.submissions || []);
    } catch (error) {
      console.error('Error loading submissions:', error);
      alert('Failed to load submissions');
    } finally {
      setLoading(false);
    }
  };

  const loadSubmissionDetails = async (submissionId: string) => {
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
        throw new Error('Failed to load submission details');
      }

      const result = await response.json();
      setStagingData(result.staging_data || []);
    } catch (error) {
      console.error('Error loading submission details:', error);
    }
  };

  const handleReviewSubmission = async () => {
    if (!selectedSubmission || !reviewAction) return;

    try {
      const token = localStorage.getItem('accessToken');
      const response = await fetch(`/api/v1/admin/submissions/${selectedSubmission.id}/review`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          status: reviewAction,
          admin_notes: adminNotes || null,
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to review submission');
      }

      alert(`Submission ${reviewAction} successfully!`);
      setSelectedSubmission(null);
      setStagingData([]);
      setReviewAction(null);
      setAdminNotes('');
      await loadPendingSubmissions();
    } catch (error) {
      console.error('Error reviewing submission:', error);
      alert(`Failed to ${reviewAction} submission`);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'pending': return 'text-yellow-600 bg-yellow-100';
      case 'under_review': return 'text-blue-600 bg-blue-100';
      case 'approved': return 'text-green-600 bg-green-100';
      case 'rejected': return 'text-red-600 bg-red-100';
      case 'applied': return 'text-purple-600 bg-purple-100';
      default: return 'text-gray-600 bg-gray-100';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'pending': return <Clock className="w-4 h-4" />;
      case 'under_review': return <Eye className="w-4 h-4" />;
      case 'approved': return <CheckCircle className="w-4 h-4" />;
      case 'rejected': return <XCircle className="w-4 h-4" />;
      case 'applied': return <CheckCircle className="w-4 h-4" />;
      default: return null;
    }
  };

  const getValidationStatusColor = (status: string) => {
    switch (status) {
      case 'valid': return 'text-green-600';
      case 'invalid': return 'text-red-600';
      case 'warning': return 'text-yellow-600';
      default: return 'text-gray-600';
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-gray-600">Loading submissions...</div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">Data Submission Reviews</h1>
        <div className="flex items-center space-x-2 text-sm text-gray-600">
          <Clock className="w-4 h-4" />
          <span>{submissions.length} pending reviews</span>
        </div>
      </div>

      {submissions.length === 0 ? (
        <div className="text-center py-12">
          <CheckCircle className="w-12 h-12 text-green-400 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">All caught up!</h3>
          <p className="text-gray-600">No submissions pending review at this time.</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Submissions List */}
          <div className="space-y-4">
            <h2 className="text-lg font-medium text-gray-900">Pending Submissions</h2>
            <div className="space-y-4">
              {submissions.map((submission) => (
                <div
                  key={submission.id}
                  className={`border rounded-lg p-4 cursor-pointer transition-colors ${
                    selectedSubmission?.id === submission.id
                      ? 'border-blue-500 bg-blue-50'
                      : 'border-gray-200 hover:border-gray-300'
                  }`}
                  onClick={() => {
                    setSelectedSubmission(submission);
                    loadSubmissionDetails(submission.id);
                  }}
                >
                  <div className="flex items-start justify-between mb-3">
                    <div>
                      <h3 className="font-medium text-gray-900">{submission.dataset_name}</h3>
                      <p className="text-sm text-gray-600">{submission.project_name}</p>
                    </div>
                    <div className={`flex items-center space-x-1 px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(submission.status)}`}>
                      {getStatusIcon(submission.status)}
                      <span>{submission.status}</span>
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-4 text-sm">
                    <div>
                      <span className="text-gray-500">Submitted by:</span>
                      <p className="font-medium">{submission.submitter_name}</p>
                    </div>
                    <div>
                      <span className="text-gray-500">File:</span>
                      <p className="font-medium">{submission.file_name}</p>
                    </div>
                    <div>
                      <span className="text-gray-500">Rows:</span>
                      <p className="font-medium">{submission.row_count}</p>
                    </div>
                    <div>
                      <span className="text-gray-500">Size:</span>
                      <p className="font-medium">{(submission.file_size / 1024 / 1024).toFixed(2)} MB</p>
                    </div>
                  </div>

                  {submission.validation_results && (
                    <div className="mt-3 grid grid-cols-3 gap-2 text-xs">
                      <div className="text-center">
                        <div className="font-bold text-green-600">{submission.validation_results.valid_rows}</div>
                        <div className="text-gray-500">Valid</div>
                      </div>
                      <div className="text-center">
                        <div className="font-bold text-red-600">{submission.validation_results.invalid_rows}</div>
                        <div className="text-gray-500">Invalid</div>
                      </div>
                      <div className="text-center">
                        <div className="font-bold text-yellow-600">{submission.validation_results.warning_rows}</div>
                        <div className="text-gray-500">Warning</div>
                      </div>
                    </div>
                  )}

                  <div className="mt-3 text-xs text-gray-500">
                    Submitted {new Date(submission.submitted_at).toLocaleDateString()}
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Submission Details */}
          <div className="space-y-4">
            {selectedSubmission ? (
              <>
                <div className="flex items-center justify-between">
                  <h2 className="text-lg font-medium text-gray-900">Review Submission</h2>
                  <div className={`flex items-center space-x-1 px-3 py-1 rounded-full text-sm font-medium ${getStatusColor(selectedSubmission.status)}`}>
                    {getStatusIcon(selectedSubmission.status)}
                    <span>{selectedSubmission.status}</span>
                  </div>
                </div>

                {/* Submission Info */}
                <div className="bg-white border border-gray-200 rounded-lg p-6">
                  <h3 className="font-medium text-gray-900 mb-4">Submission Details</h3>
                  <div className="grid grid-cols-2 gap-4 text-sm">
                    <div>
                      <label className="text-gray-500">Dataset:</label>
                      <p className="font-medium">{selectedSubmission.dataset_name}</p>
                    </div>
                    <div>
                      <label className="text-gray-500">Project:</label>
                      <p className="font-medium">{selectedSubmission.project_name}</p>
                    </div>
                    <div>
                      <label className="text-gray-500">Submitter:</label>
                      <p className="font-medium">{selectedSubmission.submitter_name}</p>
                    </div>
                    <div>
                      <label className="text-gray-500">Email:</label>
                      <p className="font-medium">{selectedSubmission.submitter_email}</p>
                    </div>
                    <div>
                      <label className="text-gray-500">File:</label>
                      <p className="font-medium">{selectedSubmission.file_name}</p>
                    </div>
                    <div>
                      <label className="text-gray-500">Submitted:</label>
                      <p className="font-medium">{new Date(selectedSubmission.submitted_at).toLocaleString()}</p>
                    </div>
                  </div>
                </div>

                {/* Validation Results */}
                {selectedSubmission.validation_results && (
                  <div className="bg-white border border-gray-200 rounded-lg p-6">
                    <h3 className="font-medium text-gray-900 mb-4">Validation Results</h3>
                    
                    <div className="grid grid-cols-4 gap-4 mb-6">
                      <div className="text-center">
                        <div className="text-2xl font-bold text-blue-900">{selectedSubmission.validation_results.total_rows}</div>
                        <div className="text-sm text-blue-700">Total Rows</div>
                      </div>
                      <div className="text-center">
                        <div className="text-2xl font-bold text-green-900">{selectedSubmission.validation_results.valid_rows}</div>
                        <div className="text-sm text-green-700">Valid</div>
                      </div>
                      <div className="text-center">
                        <div className="text-2xl font-bold text-red-900">{selectedSubmission.validation_results.invalid_rows}</div>
                        <div className="text-sm text-red-700">Invalid</div>
                      </div>
                      <div className="text-center">
                        <div className="text-2xl font-bold text-yellow-900">{selectedSubmission.validation_results.warning_rows}</div>
                        <div className="text-sm text-yellow-700">Warning</div>
                      </div>
                    </div>

                    {selectedSubmission.validation_results.schema_errors?.length > 0 && (
                      <div className="mb-4">
                        <h4 className="font-medium text-red-900 mb-2">Schema Errors ({selectedSubmission.validation_results.schema_errors.length})</h4>
                        <div className="space-y-2 max-h-32 overflow-y-auto">
                          {selectedSubmission.validation_results.schema_errors.slice(0, 3).map((error, index) => (
                            <div key={index} className="bg-red-50 border border-red-200 rounded p-2">
                              <div className="text-sm">
                                <span className="font-medium">Row {error.row_index + 1}, {error.field_name}:</span>
                                <span className="text-red-700 ml-1">{error.message}</span>
                              </div>
                            </div>
                          ))}
                          {selectedSubmission.validation_results.schema_errors.length > 3 && (
                            <p className="text-sm text-gray-600">...and {selectedSubmission.validation_results.schema_errors.length - 3} more</p>
                          )}
                        </div>
                      </div>
                    )}

                    {selectedSubmission.validation_results.business_rule_errors?.length > 0 && (
                      <div>
                        <h4 className="font-medium text-orange-900 mb-2">Business Rule Errors ({selectedSubmission.validation_results.business_rule_errors.length})</h4>
                        <div className="space-y-2 max-h-32 overflow-y-auto">
                          {selectedSubmission.validation_results.business_rule_errors.slice(0, 3).map((error, index) => (
                            <div key={index} className="bg-orange-50 border border-orange-200 rounded p-2">
                              <div className="text-sm">
                                <span className="font-medium">Row {error.row_index + 1}, {error.field_name}:</span>
                                <span className="text-orange-700 ml-1">{error.message}</span>
                              </div>
                            </div>
                          ))}
                          {selectedSubmission.validation_results.business_rule_errors.length > 3 && (
                            <p className="text-sm text-gray-600">...and {selectedSubmission.validation_results.business_rule_errors.length - 3} more</p>
                          )}
                        </div>
                      </div>
                    )}
                  </div>
                )}

                {/* Data Preview */}
                <div className="bg-white border border-gray-200 rounded-lg">
                  <div className="px-6 py-4 border-b">
                    <h3 className="font-medium text-gray-900">Data Preview</h3>
                  </div>
                  <div className="overflow-x-auto">
                    <table className="min-w-full divide-y divide-gray-200">
                      <thead className="bg-gray-50">
                        <tr>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Row</th>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                          {stagingData.length > 0 && Object.keys(stagingData[0].data).map((field) => (
                            <th key={field} className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                              {field}
                            </th>
                          ))}
                        </tr>
                      </thead>
                      <tbody className="bg-white divide-y divide-gray-200">
                        {stagingData.slice(0, 10).map((row) => (
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
                              <span className={`text-sm ${getValidationStatusColor(row.validation_status)}`}>
                                {row.validation_status}
                              </span>
                            </td>
                            {Object.values(row.data).map((value, index) => (
                              <td key={index} className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                                {String(value)}
                              </td>
                            ))}
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                  {stagingData.length > 10 && (
                    <div className="px-6 py-3 text-sm text-gray-600 border-t">
                      Showing first 10 rows of {selectedSubmission.row_count} total
                    </div>
                  )}
                </div>

                {/* Review Actions */}
                <div className="bg-white border border-gray-200 rounded-lg p-6">
                  <h3 className="font-medium text-gray-900 mb-4">Admin Review</h3>
                  
                  <div className="space-y-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Review Decision
                      </label>
                      <div className="flex space-x-4">
                        <label className="flex items-center">
                          <input
                            type="radio"
                            name="reviewAction"
                            value="approve"
                            checked={reviewAction === 'approve'}
                            onChange={(e) => setReviewAction(e.target.value as 'approve')}
                            className="mr-2"
                          />
                          <CheckCircle className="w-4 h-4 text-green-600 mr-1" />
                          <span className="text-green-700">Approve</span>
                        </label>
                        <label className="flex items-center">
                          <input
                            type="radio"
                            name="reviewAction"
                            value="reject"
                            checked={reviewAction === 'reject'}
                            onChange={(e) => setReviewAction(e.target.value as 'reject')}
                            className="mr-2"
                          />
                          <XCircle className="w-4 h-4 text-red-600 mr-1" />
                          <span className="text-red-700">Reject</span>
                        </label>
                      </div>
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Admin Notes
                      </label>
                      <textarea
                        value={adminNotes}
                        onChange={(e) => setAdminNotes(e.target.value)}
                        rows={3}
                        className="w-full border border-gray-300 rounded-lg px-3 py-2 focus:ring-blue-500 focus:border-blue-500"
                        placeholder="Add notes for the submitter (optional)..."
                      />
                    </div>

                    <div className="flex justify-end space-x-3">
                      <button
                        onClick={() => {
                          setSelectedSubmission(null);
                          setStagingData([]);
                          setReviewAction(null);
                          setAdminNotes('');
                        }}
                        className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50"
                      >
                        Cancel
                      </button>
                      <button
                        onClick={handleReviewSubmission}
                        disabled={!reviewAction}
                        className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
                      >
                        Submit Review
                      </button>
                    </div>
                  </div>
                </div>
              </>
            ) : (
              <div className="text-center py-12 text-gray-500">
                <FileText className="w-12 h-12 mx-auto mb-4" />
                <p>Select a submission to review</p>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default AdminDataSubmissionReview;
