export interface Dataset {
  id: string;
  project_id: string;
  name: string;
  description: string;
  file_name: string;
  file_path: string;
  file_size: number;
  mime_type: string;
  row_count: number;
  column_count: number;
  status: 'processing' | 'ready' | 'error';
  uploaded_by: string;
  created_at: string;
  updated_at: string;
}

export interface DatasetWithProject extends Dataset {
  project_name: string;
}

export interface CreateDatasetRequest {
  project_id: string;
  name: string;
  description?: string;
}

export interface UploadDatasetRequest {
  project_id: string;
  name?: string;
  description?: string;
  file: File;
}
