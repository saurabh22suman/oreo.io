import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from '@/contexts/AuthContext';
import ProtectedRoute from '@/components/ProtectedRoute';
import LoginPage from '@/pages/LoginPage';
import RegisterPage from '@/pages/RegisterPage';
import DashboardPage from '@/pages/DashboardPage';
import ProjectManagementPage from '@/pages/ProjectManagementPage';
import DataViewPage from '@/pages/DataViewPage';

const App: React.FC = () => {
  return (
    <AuthProvider>
      <Router>
        <div className="App">
          <Routes>
            {/* Public Routes */}
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<RegisterPage />} />
            
            {/* Protected Routes */}
            <Route 
              path="/dashboard" 
              element={
                <ProtectedRoute>
                  <DashboardPage />
                </ProtectedRoute>
              } 
            />
            
            <Route 
              path="/project/:projectId" 
              element={
                <ProtectedRoute>
                  <ProjectManagementPage />
                </ProtectedRoute>
              } 
            />
            
            <Route 
              path="/dataset/:datasetId/view" 
              element={
                <ProtectedRoute>
                  <DataViewPage />
                </ProtectedRoute>
              } 
            />
            
            {/* Default Redirect */}
            <Route path="/" element={<Navigate to="/dashboard" replace />} />
            
            {/* 404 - Redirect to dashboard */}
            <Route path="*" element={<Navigate to="/dashboard" replace />} />
          </Routes>
        </div>
      </Router>
    </AuthProvider>
  );
};

export default App;
