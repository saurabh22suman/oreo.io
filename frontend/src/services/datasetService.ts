import { Dataset, DatasetWithProject, UploadDatasetRequest } from '@/types/dataset';

const API_BASE_URL = 'http://localhost:8080/api/v1';

// Get authentication token from localStorage
const getAuthToken = (): string | null => {
  return localStorage.getItem('accessToken');
};

// Create headers with authentication
const createAuthHeaders = () => {
  const token = getAuthToken();
  return {
    ...(token && { Authorization: `Bearer ${token}` }),
  };
};

export const datasetService = {
  // Upload a dataset file
  async uploadDataset(data: UploadDatasetRequest): Promise<{ dataset: Dataset; message: string }> {
    const formData = new FormData();
    formData.append('file', data.file);
    formData.append('project_id', data.project_id);
    
    if (data.name) {
      formData.append('name', data.name);
    }
    
    if (data.description) {
      formData.append('description', data.description);
    }

    const response = await fetch(`${API_BASE_URL}/datasets/upload`, {
      method: 'POST',
      headers: createAuthHeaders(),
      body: formData,
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Failed to upload dataset' }));
      throw new Error(error.error || 'Failed to upload dataset');
    }

    return response.json();
  },

  // Get datasets for a project
  async getDatasets(projectId: string): Promise<{ datasets: Dataset[]; count: number }> {
    const response = await fetch(`${API_BASE_URL}/datasets/project/${projectId}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        ...createAuthHeaders(),
      },
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Failed to fetch datasets' }));
      throw new Error(error.error || 'Failed to fetch datasets');
    }

    return response.json();
  },

  // Get all datasets uploaded by the authenticated user
  async getUserDatasets(): Promise<{ datasets: DatasetWithProject[]; count: number }> {
    const response = await fetch(`${API_BASE_URL}/datasets/user`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        ...createAuthHeaders(),
      },
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Failed to fetch user datasets' }));
      throw new Error(error.error || 'Failed to fetch user datasets');
    }

    return response.json();
  },

  // Delete a dataset
  async deleteDataset(id: string): Promise<{ message: string }> {
    const response = await fetch(`${API_BASE_URL}/datasets/${id}`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
        ...createAuthHeaders(),
      },
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Failed to delete dataset' }));
      throw new Error(error.error || 'Failed to delete dataset');
    }

    return response.json();
  },
};
