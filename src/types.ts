import { DataQuery, DataSourceJsonData } from '@grafana/data';

// ---------------------------------------------------------
// 1. Query Editor State (What the user configures in a panel)
// ---------------------------------------------------------
export interface MediahubQuery extends DataQuery {
  databaseId?: string;
  model: 'get metadata table' | 'get preview' | 'get entry' | 'get audit logs';
  
  // Pagination & Time
  tstart?: number; // Epoch ms
  tend?: number;   // Epoch ms
  limit?: number;

  // Metadata Table Options
  addPreviewLink?: boolean;
  addEntryLink?: boolean;

  // Preview & Entry Options
  targetSelection?: 'get ID' | 'get last' | 'get last in range';
  entryId?: string; // String to allow Grafana variables like ${entry_id}
  base64?: boolean;
  maxFileSize?: number; // In MB
}

// Set sensible defaults for a new panel
export const defaultQuery: Partial<MediahubQuery> = {
  model: 'get metadata table',
  limit: 50,
  addPreviewLink: true,
  addEntryLink: false,
  base64: false,
  maxFileSize: 4,
  targetSelection: 'get last',
};

// ---------------------------------------------------------
// 2. Datasource Configuration (What the admin sets up)
// ---------------------------------------------------------
export interface MediahubDataSourceOptions extends DataSourceJsonData {
  url?: string;
  username?: string;
}

export interface MediahubSecureJsonData {
  password?: string;
}

// ---------------------------------------------------------
// 3. Internal Types (Responses from our Go backend)
// ---------------------------------------------------------
export interface MediahubDatabase {
  id: string;
  name: string;
  content_type: string;
}

export interface MediahubConfigResponse {
  is_admin: boolean;
  databases: MediahubDatabase[];
}