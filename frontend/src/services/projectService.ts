import { CreateProjectRequest, Project, UpdateProjectRequest } from '@/types/project';

const API_BASE_URL = 'http://localhost:8080/api/v1';

// Get authentication token from localStorage
const getAuthToken = (): string | null => {
  return localStorage.getItem('accessToken');
};

// Create headers with authentication
const createAuthHeaders = () => {
  const token = getAuthToken();
  return {
    'Content-Type': 'application/json',
    ...(token && { Authorization: `Bearer ${token}` }),
  };
};

export const projectService = {
  // Get all projects for the authenticated user
  async getProjects(): Promise<{ projects: Project[]; count: number }> {
    const response = await fetch(`${API_BASE_URL}/projects`, {
      method: 'GET',
      headers: createAuthHeaders(),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Failed to fetch projects' }));
      throw new Error(error.error || 'Failed to fetch projects');
    }

    return response.json();
  },

  // Create a new project
  async createProject(projectData: CreateProjectRequest): Promise<{ project: Project; message: string }> {
    const response = await fetch(`${API_BASE_URL}/projects`, {
      method: 'POST',
      headers: createAuthHeaders(),
      body: JSON.stringify(projectData),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Failed to create project' }));
      throw new Error(error.error || 'Failed to create project');
    }

    return response.json();
  },

  // Get a specific project
  async getProject(id: string): Promise<{ project: Project }> {
    const response = await fetch(`${API_BASE_URL}/projects/${id}`, {
      method: 'GET',
      headers: createAuthHeaders(),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Failed to fetch project' }));
      throw new Error(error.error || 'Failed to fetch project');
    }

    return response.json();
  },

  // Update a project
  async updateProject(id: string, updates: UpdateProjectRequest): Promise<{ project: Project; message: string }> {
    const response = await fetch(`${API_BASE_URL}/projects/${id}`, {
      method: 'PUT',
      headers: createAuthHeaders(),
      body: JSON.stringify(updates),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Failed to update project' }));
      throw new Error(error.error || 'Failed to update project');
    }

    return response.json();
  },

  // Delete a project
  async deleteProject(id: string): Promise<{ message: string }> {
    const response = await fetch(`${API_BASE_URL}/projects/${id}`, {
      method: 'DELETE',
      headers: createAuthHeaders(),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Failed to delete project' }));
      throw new Error(error.error || 'Failed to delete project');
    }

    return response.json();
  },
};
