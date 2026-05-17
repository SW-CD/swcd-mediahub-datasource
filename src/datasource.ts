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
    return this.getResource('/config');
  }

  // This method is called automatically when you click "Run Query" in the Dashboard Variables settings.
  async metricFindQuery(query: string, options?: any): Promise<MetricFindValue[]> {
    if (!query) {
      return [];
    }

    const templateSrv = getTemplateSrv();
    const interpolatedQuery = templateSrv.replace(query.trim(), options?.scopedVars);
    const command = interpolatedQuery.toLowerCase();

    // Command 1: "databases" (The Hybrid: Shows Name, Uses ULID)
    if (command === 'databases') {
      const config = await this.getMediahubConfig();
      return config.databases.map((db) => ({
        text: db.name,
        value: db.id,
      }));
    }

    // Command 2: "database_names" (Shows Name, Uses Name)
    if (command === 'database_names') {
      const config = await this.getMediahubConfig();
      return config.databases.map((db) => ({
        text: db.name,
        value: db.name,
      }));
    }

    // Command 3: "database_ulids" (Shows ULID, Uses ULID)
    if (command === 'database_ulids') {
      const config = await this.getMediahubConfig();
      return config.databases.map((db) => ({
        text: db.id,
        value: db.id,
      }));
    }

    // Command 4: "entries <database_id>"
    if (command.startsWith('entries ')) {
      const dbId = interpolatedQuery.substring(8).trim();
      try {
        const entries = await this.getResource(`/variables/entries/${dbId}`);
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