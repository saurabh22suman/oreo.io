import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { projectService } from '@/services/projectService';
import { datasetService } from '@/services/datasetService';
import { Project, CreateProjectRequest } from '@/types/project';
import { Dataset } from '@/types/dataset';
import UploadDatasetModal from '@/components/UploadDatasetModal';
import CreateProjectModal from '@/components/CreateProjectModal';
import { Eye, Trash2 } from 'lucide-react';

const ProjectManagementPage: React.FC = () => {
  const { projectId } = useParams<{ projectId: string }>();
  const navigate = useNavigate();
  
  const [project, setProject] = useState<Project | null>(null);
  const [datasets, setDatasets] = useState<Dataset[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'overview' | 'datasets' | 'members' | 'settings'>('overview');
  const [showUploadModal, setShowUploadModal] = useState(false);
  const [inviteEmail, setInviteEmail] = useState('');
  const [inviteRole, setInviteRole] = useState<'editor' | 'reviewer' | 'viewer'>('viewer');
  const [isInviting, setIsInviting] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [isUpdating, setIsUpdating] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);

  useEffect(() => {
    if (!projectId) {
      navigate('/dashboard');
      return;
    }
    loadProjectData();
  }, [projectId, navigate]);

  const loadProjectData = async () => {
    try {
      setLoading(true);
      setError(null);
      
      // Load project details
      const projectResponse = await projectService.getProject(projectId!);
      setProject(projectResponse.project);
      
      // Load datasets
      try {
        const datasetsResponse = await datasetService.getDatasets(projectId!);
        setDatasets(datasetsResponse.datasets || []);
      } catch (err) {
        console.warn('Failed to load datasets:', err);
        setDatasets([]);
      }
      
      // TODO: Load project members when API is ready
      // const membersResponse = await projectService.getProjectMembers(projectId!);
      // setMembers(membersResponse.members || []);
      
    } catch (err) {
      console.error('Error loading project data:', err);
      setError(err instanceof Error ? err.message : 'Failed to load project data');
    } finally {
      setLoading(false);
    }
  };

  const handleUploadDataset = async (uploadData: any) => {
    try {
      await datasetService.uploadDataset(uploadData);
      setShowUploadModal(false);
      await loadProjectData(); // Refresh datasets
    } catch (err) {
      console.error('Error uploading dataset:', err);
      throw err;
    }
  };

  const handleInviteUser = async () => {
    if (!inviteEmail.trim()) return;
    
    try {
      setIsInviting(true);
      // TODO: Implement invite user API
      console.log('Inviting user:', { email: inviteEmail, role: inviteRole });
      setInviteEmail('');
      setInviteRole('viewer');
      // await loadProjectData(); // Refresh members list
    } catch (err) {
      console.error('Error inviting user:', err);
      setError(err instanceof Error ? err.message : 'Failed to invite user');
    } finally {
      setIsInviting(false);
    }
  };

  const handleDeleteDataset = async (datasetId: string) => {
    if (!confirm('Are you sure you want to delete this dataset? This action cannot be undone.')) {
      return;
    }
    
    try {
      await datasetService.deleteDataset(datasetId);
      await loadProjectData(); // Refresh datasets
    } catch (err) {
      console.error('Error deleting dataset:', err);
      setError(err instanceof Error ? err.message : 'Failed to delete dataset');
    }
  };

  const handleEditProject = async (projectData: CreateProjectRequest) => {
    try {
      setIsUpdating(true);
      const response = await projectService.updateProject(projectId!, projectData);
      setProject(response.project);
      setShowEditModal(false);
    } catch (err) {
      console.error('Error updating project:', err);
      setError(err instanceof Error ? err.message : 'Failed to update project');
      throw err;
    } finally {
      setIsUpdating(false);
    }
  };

  const handleDeleteProject = async () => {
    if (!confirm('Are you sure you want to delete this project? This action cannot be undone and will delete all associated datasets.')) {
      return;
    }
    
    try {
      setIsDeleting(true);
      await projectService.deleteProject(projectId!);
      navigate('/dashboard');
    } catch (err) {
      console.error('Error deleting project:', err);
      setError(err instanceof Error ? err.message : 'Failed to delete project');
    } finally {
      setIsDeleting(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading project...</p>
        </div>
      </div>
    );
  }

  if (error || !project) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <p className="text-red-600">{error || 'Project not found'}</p>
          <button 
            onClick={() => navigate('/dashboard')}
            className="mt-4 btn-primary"
          >
            Back to Dashboard
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-6">
            <div className="flex items-center space-x-4">
              <button 
                onClick={() => navigate('/dashboard')}
                className="text-gray-500 hover:text-gray-700"
              >
                <svg className="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                </svg>
              </button>
              <div>
                <h1 className="text-2xl font-bold text-gray-900">{project.name}</h1>
                {project.description && (
                  <p className="text-gray-600">{project.description}</p>
                )}
              </div>
            </div>
            <div className="flex items-center space-x-4">
              <span className="text-sm text-gray-500">
                Created {new Date(project.created_at).toLocaleDateString()}
              </span>
            </div>
          </div>
        </div>
      </header>

      {/* Navigation Tabs */}
      <div className="bg-white border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <nav className="flex space-x-8">
            {['overview', 'datasets', 'members', 'settings'].map((tab) => (
              <button
                key={tab}
                onClick={() => setActiveTab(tab as any)}
                className={`py-4 px-1 border-b-2 font-medium text-sm capitalize ${
                  activeTab === tab
                    ? 'border-primary-500 text-primary-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                }`}
              >
                {tab}
              </button>
            ))}
          </nav>
        </div>
      </div>

      {/* Content */}
      <main className="max-w-7xl mx-auto py-6 px-4 sm:px-6 lg:px-8">
        {error && (
          <div className="bg-red-50 border border-red-200 rounded-md p-4 mb-6">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" />
                </svg>
              </div>
              <div className="ml-3">
                <p className="text-sm text-red-700">{error}</p>
                <button 
                  onClick={() => setError(null)}
                  className="text-red-600 hover:text-red-500 text-sm underline mt-1"
                >
                  Dismiss
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Overview Tab */}
        {activeTab === 'overview' && (
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <div className="lg:col-span-2">
              <div className="card p-6">
                <h2 className="text-lg font-medium text-gray-900 mb-4">Project Statistics</h2>
                <div className="grid grid-cols-2 gap-4">
                  <div className="bg-gray-50 p-4 rounded-lg">
                    <h3 className="text-sm font-medium text-gray-500">Datasets</h3>
                    <p className="text-2xl font-bold text-gray-900">{datasets.length}</p>
                  </div>
                  <div className="bg-gray-50 p-4 rounded-lg">
                    <h3 className="text-sm font-medium text-gray-500">Members</h3>
                    <p className="text-2xl font-bold text-gray-900">0</p>
                  </div>
                </div>
              </div>
            </div>
            <div>
              <div className="card p-6">
                <h2 className="text-lg font-medium text-gray-900 mb-4">Quick Actions</h2>
                <div className="space-y-3">
                  <button 
                    onClick={() => setShowUploadModal(true)}
                    className="w-full btn-primary"
                  >
                    <svg className="h-4 w-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                    </svg>
                    Upload Dataset
                  </button>
                  <button 
                    onClick={() => setActiveTab('members')}
                    className="w-full btn-outline"
                  >
                    <svg className="h-4 w-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z" />
                    </svg>
                    Manage Members
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Datasets Tab */}
        {activeTab === 'datasets' && (
          <div>
            <div className="flex justify-between items-center mb-6">
              <h2 className="text-xl font-semibold text-gray-900">Datasets</h2>
              <button 
                onClick={() => setShowUploadModal(true)}
                className="btn-primary"
              >
                <svg className="h-4 w-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                </svg>
                Upload Dataset
              </button>
            </div>
            
            {datasets.length > 0 ? (
              <div className="card p-6">
                <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                      <tr>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Size</th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Uploaded</th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
                      </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-gray-200">
                      {datasets.map((dataset) => (
                        <tr key={dataset.id}>
                          <td className="px-6 py-4 whitespace-nowrap">
                            <div>
                              <div className="text-sm font-medium text-gray-900">{dataset.name}</div>
                              {dataset.description && (
                                <div className="text-sm text-gray-500">{dataset.description}</div>
                              )}
                            </div>
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap">
                            <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                              dataset.status === 'ready' ? 'bg-green-100 text-green-800' :
                              dataset.status === 'processing' ? 'bg-yellow-100 text-yellow-800' :
                              dataset.status === 'error' ? 'bg-red-100 text-red-800' :
                              'bg-gray-100 text-gray-800'
                            }`}>
                              {dataset.status}
                            </span>
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                            {dataset.file_size ? `${(dataset.file_size / 1024 / 1024).toFixed(2)} MB` : 'N/A'}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                            {new Date(dataset.created_at).toLocaleDateString()}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                            <div className="flex space-x-2">
                              <button 
                                onClick={() => navigate(`/dataset/${dataset.id}/view`)}
                                className="text-blue-600 hover:text-blue-900 flex items-center"
                                title="View and edit data"
                              >
                                <Eye className="w-4 h-4 mr-1" />
                                View Data
                              </button>
                              <button 
                                onClick={() => handleDeleteDataset(dataset.id)}
                                className="text-red-600 hover:text-red-900 flex items-center"
                                title="Delete dataset"
                              >
                                <Trash2 className="w-4 h-4 mr-1" />
                                Delete
                              </button>
                            </div>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            ) : (
              <div className="card p-12 text-center">
                <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                <h3 className="mt-2 text-sm font-medium text-gray-900">No datasets</h3>
                <p className="mt-1 text-sm text-gray-500">Get started by uploading your first dataset.</p>
                <div className="mt-6">
                  <button 
                    onClick={() => setShowUploadModal(true)}
                    className="btn-primary"
                  >
                    Upload Dataset
                  </button>
                </div>
              </div>
            )}
          </div>
        )}

        {/* Members Tab */}
        {activeTab === 'members' && (
          <div>
            <div className="flex justify-between items-center mb-6">
              <h2 className="text-xl font-semibold text-gray-900">Project Members</h2>
            </div>
            
            {/* Invite User Form */}
            <div className="card p-6 mb-6">
              <h3 className="text-lg font-medium text-gray-900 mb-4">Invite User</h3>
              <div className="flex space-x-4">
                <div className="flex-1">
                  <input
                    type="email"
                    placeholder="Enter email address"
                    value={inviteEmail}
                    onChange={(e) => setInviteEmail(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                  />
                </div>
                <div>
                  <select
                    value={inviteRole}
                    onChange={(e) => setInviteRole(e.target.value as any)}
                    className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                  >
                    <option value="viewer">Viewer</option>
                    <option value="reviewer">Reviewer</option>
                    <option value="editor">Editor</option>
                  </select>
                </div>
                <button
                  onClick={handleInviteUser}
                  disabled={isInviting || !inviteEmail.trim()}
                  className="btn-primary"
                >
                  {isInviting ? 'Inviting...' : 'Invite'}
                </button>
              </div>
              <p className="mt-2 text-sm text-gray-500">
                Users will receive an email invitation to join this project.
              </p>
            </div>

            {/* Members List */}
            <div className="card p-6">
              <div className="text-center py-12">
                <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z" />
                </svg>
                <h3 className="mt-2 text-sm font-medium text-gray-900">Member management coming soon</h3>
                <p className="mt-1 text-sm text-gray-500">
                  Project collaboration features are being developed. You'll be able to invite users and manage permissions here.
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Settings Tab */}
        {activeTab === 'settings' && (
          <div>
            <div className="flex justify-between items-center mb-6">
              <h2 className="text-xl font-semibold text-gray-900">Project Settings</h2>
            </div>
            
            <div className="space-y-6">
              {/* Edit Project */}
              <div className="card p-6">
                <h3 className="text-lg font-medium text-gray-900 mb-4">Project Information</h3>
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Project Name</label>
                    <p className="text-gray-900">{project.name}</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
                    <p className="text-gray-600">{project.description || 'No description provided'}</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Created</label>
                    <p className="text-gray-600">{new Date(project.created_at).toLocaleString()}</p>
                  </div>
                  <div className="pt-4">
                    <button 
                      onClick={() => setShowEditModal(true)}
                      className="btn-primary"
                      disabled={isUpdating}
                    >
                      <svg className="h-4 w-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                      </svg>
                      {isUpdating ? 'Updating...' : 'Edit Project'}
                    </button>
                  </div>
                </div>
              </div>

              {/* Danger Zone */}
              <div className="card p-6 border-red-200">
                <h3 className="text-lg font-medium text-red-900 mb-4">Danger Zone</h3>
                <div className="bg-red-50 border border-red-200 rounded-md p-4">
                  <div className="flex">
                    <div className="flex-shrink-0">
                      <svg className="h-5 w-5 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" />
                      </svg>
                    </div>
                    <div className="ml-3 flex-1">
                      <h4 className="text-sm font-medium text-red-800">Delete Project</h4>
                      <p className="mt-1 text-sm text-red-700">
                        Once you delete a project, there is no going back. This will permanently delete the project and all associated datasets.
                      </p>
                      <div className="mt-4">
                        <button 
                          onClick={handleDeleteProject}
                          disabled={isDeleting}
                          className="bg-red-600 hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed text-white px-4 py-2 rounded-md text-sm font-medium transition-colors"
                        >
                          {isDeleting ? 'Deleting...' : 'Delete Project'}
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}
      </main>

      {/* Edit Project Modal */}
      <CreateProjectModal
        isOpen={showEditModal}
        onClose={() => setShowEditModal(false)}
        onSubmit={handleEditProject}
        isLoading={isUpdating}
        initialData={{
          name: project?.name || '',
          description: project?.description || '',
        }}
        title="Edit Project"
      />

      {/* Upload Modal */}
      <UploadDatasetModal
        isOpen={showUploadModal}
        onClose={() => setShowUploadModal(false)}
        onSubmit={handleUploadDataset}
        projectId={projectId!}
        projectName={project.name}
      />
    </div>
  );
};

export default ProjectManagementPage;
