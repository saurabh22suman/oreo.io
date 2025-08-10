import React, { useState, useRef } from 'react';
import { UploadDatasetRequest } from '@/types/dataset';

interface UploadDatasetModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: UploadDatasetRequest) => Promise<void>;
  projectId: string;
  projectName: string;
  isLoading?: boolean;
}

const UploadDatasetModal: React.FC<UploadDatasetModalProps> = ({
  isOpen,
  onClose,
  onSubmit,
  projectId,
  projectName,
  isLoading = false,
}) => {
  const [formData, setFormData] = useState({
    name: '',
    description: '',
  });
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [errors, setErrors] = useState<{ name?: string; file?: string }>({});
  const [dragActive, setDragActive] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
    // Clear error when user starts typing
    if (errors[name as keyof typeof errors]) {
      setErrors(prev => ({ ...prev, [name]: undefined }));
    }
  };

  const handleFileSelect = (file: File) => {
    if (!isValidFileType(file.name)) {
      setErrors(prev => ({ ...prev, file: 'Only CSV and Excel files are supported' }));
      return;
    }

    if (file.size > 50 * 1024 * 1024) { // 50MB limit
      setErrors(prev => ({ ...prev, file: 'File size must be less than 50MB' }));
      return;
    }

    setSelectedFile(file);
    if (!formData.name) {
      const nameWithoutExt = file.name.substring(0, file.name.lastIndexOf('.'));
      setFormData(prev => ({ ...prev, name: nameWithoutExt }));
    }
    setErrors(prev => ({ ...prev, file: undefined }));
  };

  const handleFileInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      handleFileSelect(file);
    }
  };

  const handleDrag = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.type === 'dragenter' || e.type === 'dragover') {
      setDragActive(true);
    } else if (e.type === 'dragleave') {
      setDragActive(false);
    }
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);

    const file = e.dataTransfer.files?.[0];
    if (file) {
      handleFileSelect(file);
    }
  };

  const validateForm = (): boolean => {
    const newErrors: { name?: string; file?: string } = {};

    if (!formData.name.trim()) {
      newErrors.name = 'Dataset name is required';
    }

    if (!selectedFile) {
      newErrors.file = 'Please select a file to upload';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!validateForm() || !selectedFile) {
      return;
    }

    try {
      await onSubmit({
        project_id: projectId,
        name: formData.name.trim(),
        description: formData.description.trim() || undefined,
        file: selectedFile,
      });
      
      // Reset form on success
      setFormData({ name: '', description: '' });
      setSelectedFile(null);
      setErrors({});
      onClose();
    } catch (error) {
      // Error handling is done by parent component
      console.error('Upload failed:', error);
    }
  };

  const isValidFileType = (filename: string): boolean => {
    const ext = filename.toLowerCase().split('.').pop();
    return ext === 'csv' || ext === 'xlsx' || ext === 'xls';
  };

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-lg shadow-xl max-w-md w-full max-h-[90vh] overflow-y-auto">
        <div className="flex justify-between items-center p-6 border-b">
          <h2 className="text-xl font-semibold text-gray-900">Upload Dataset</h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 transition-colors"
            disabled={isLoading}
          >
            <svg className="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        
        <form onSubmit={handleSubmit} className="p-6">
          <div className="mb-4">
            <p className="text-sm text-gray-600 mb-4">
              Upload a dataset to <span className="font-medium">{projectName}</span>
            </p>
          </div>

          {/* File Upload Area */}
          <div className="mb-6">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Select File
            </label>
            
            <div
              className={`border-2 border-dashed rounded-lg p-6 text-center transition-colors ${
                dragActive 
                  ? 'border-primary-500 bg-primary-50' 
                  : 'border-gray-300 hover:border-gray-400'
              }`}
              onDragEnter={handleDrag}
              onDragLeave={handleDrag}
              onDragOver={handleDrag}
              onDrop={handleDrop}
            >
              {selectedFile ? (
                <div className="space-y-2">
                  <div className="flex items-center justify-center">
                    <svg className="h-8 w-8 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                  </div>
                  <p className="text-sm font-medium text-gray-900">{selectedFile.name}</p>
                  <p className="text-xs text-gray-500">{formatFileSize(selectedFile.size)}</p>
                  <button
                    type="button"
                    onClick={() => setSelectedFile(null)}
                    className="text-xs text-primary-600 hover:text-primary-700"
                  >
                    Remove file
                  </button>
                </div>
              ) : (
                <div className="space-y-2">
                  <div className="flex items-center justify-center">
                    <svg className="h-8 w-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                    </svg>
                  </div>
                  <p className="text-sm text-gray-600">
                    <button
                      type="button"
                      onClick={() => fileInputRef.current?.click()}
                      className="text-primary-600 hover:text-primary-700 font-medium"
                    >
                      Click to upload
                    </button>
                    {' '}or drag and drop
                  </p>
                  <p className="text-xs text-gray-500">CSV, Excel files up to 50MB</p>
                </div>
              )}
            </div>
            
            <input
              ref={fileInputRef}
              type="file"
              accept=".csv,.xlsx,.xls"
              onChange={handleFileInputChange}
              className="hidden"
            />
            
            {errors.file && (
              <p className="mt-1 text-sm text-red-600">{errors.file}</p>
            )}
          </div>

          {/* Dataset Name */}
          <div className="mb-4">
            <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-1">
              Dataset Name *
            </label>
            <input
              type="text"
              id="name"
              name="name"
              value={formData.name}
              onChange={handleInputChange}
              className={`w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-1 focus:ring-primary-500 focus:border-primary-500 ${
                errors.name ? 'border-red-300' : 'border-gray-300'
              }`}
              placeholder="Enter dataset name"
              disabled={isLoading}
            />
            {errors.name && (
              <p className="mt-1 text-sm text-red-600">{errors.name}</p>
            )}
          </div>

          {/* Description */}
          <div className="mb-6">
            <label htmlFor="description" className="block text-sm font-medium text-gray-700 mb-1">
              Description
            </label>
            <textarea
              id="description"
              name="description"
              value={formData.description}
              onChange={handleInputChange}
              rows={3}
              className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-1 focus:ring-primary-500 focus:border-primary-500"
              placeholder="Describe your dataset (optional)"
              disabled={isLoading}
            />
          </div>

          {/* Submit Buttons */}
          <div className="flex justify-end space-x-3">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 border border-gray-300 rounded-md hover:bg-gray-200 focus:outline-none focus:ring-2 focus:ring-gray-500"
              disabled={isLoading}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="px-4 py-2 text-sm font-medium text-white bg-primary-600 border border-transparent rounded-md hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500 disabled:opacity-50 disabled:cursor-not-allowed"
              disabled={isLoading || !selectedFile}
            >
              {isLoading ? 'Uploading...' : 'Upload Dataset'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default UploadDatasetModal;
