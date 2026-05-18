import {
  DataSourceInstanceSettings,
  DataQueryRequest,
  DataQueryResponse,
  ScopedVars,
  MetricFindValue,
} from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';

import { 
  MediahubQuery, 
  MediahubDataSourceOptions, 
  MediahubConfigResponse 
} from './types';

export class DataSource extends DataSourceWithBackend<MediahubQuery, MediahubDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<MediahubDataSourceOptions>) {
    super(instanceSettings);
  }

// Grafana natively calls this method on every query target before sending it to the backend.
  applyTemplateVariables(query: MediahubQuery, scopedVars: ScopedVars): Record<string, any> {
    const templateSrv = getTemplateSrv();

    return {
      ...query,
      // Interpolate the Database ID
      databaseId: query.databaseId ? templateSrv.replace(query.databaseId, scopedVars) : query.databaseId,
      // Interpolate the Entry ID
      entryId: query.entryId ? templateSrv.replace(query.entryId, scopedVars) : query.entryId,
    };
  }

  // Fetches the dynamic configuration (Admin status & Databases) for the Query Editor.
  async getMediahubConfig(): Promise<MediahubConfigResponse> {
    return this.getResource('config');
  }

  // Grafana calls this before executing any query. 
  // If it returns false, the query is cleanly aborted without throwing an error.
  filterQuery(query: MediahubQuery): boolean {
    // If no database ID is selected (or if it's an empty string), block the query.
    if (!query.databaseId) {
      return false;
    }
    
    // Optionally, if you want to ensure the entry ID is present for "Get ID" targets:
    if ((query.model === 'get preview' || query.model === 'get entry') && 
        query.targetSelection === 'get ID' && !query.entryId) {
      return false;
    }

    return true;
  }

  // This method is called automatically when you click "Run Query" in the Dashboard Variables settings.
  async metricFindQuery(query: any, options?: any): Promise<MetricFindValue[]> {
    if (!query) {
      return [];
    }

    // Safely extract the query string whether Grafana passed a string or a query object
    const rawQuery = typeof query === 'string' ? query : (query.query || '');
    
    if (!rawQuery) {
      return [];
    }

    const templateSrv = getTemplateSrv();
    const interpolatedQuery = templateSrv.replace(rawQuery.trim(), options?.scopedVars);
    const command = interpolatedQuery.toLowerCase();

    if (command === 'databases') {
      const config = await this.getMediahubConfig();
      return config.databases.map((db) => ({
        text: db.name,
        value: db.id,
      }));
    }

    if (command === 'database_names') {
      const config = await this.getMediahubConfig();
      return config.databases.map((db) => ({
        text: db.name,
        value: db.name,
      }));
    }

    if (command === 'database_ulids') {
      const config = await this.getMediahubConfig();
      return config.databases.map((db) => ({
        text: db.id,
        value: db.id,
      }));
    }

    if (command.startsWith('entries ')) {
      const dbId = interpolatedQuery.substring(8).trim();
      try {
        const entries = await this.getResource(`variables/entries/${dbId}`);
        return entries.map((e: { id: number; filename: string }) => ({
          text: `${e.filename} (ID: ${e.id})`,
          value: String(e.id),
        }));
      } catch (err) {
        console.error("Failed to fetch variable entries", err);
        return [];
      }
    }

    return [];
  }
}