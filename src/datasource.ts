import {
  DataSourceInstanceSettings,
  ScopedVars,
  MetricFindValue,
  CustomVariableSupport,
  DataQueryRequest,
  DataQueryResponse, 
} from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';
import { VariableQueryEditor } from './components/VariableQueryEditor';
import { from, Observable } from 'rxjs';

import { 
  MediahubQuery, 
  MediahubDataSourceOptions, 
  MediahubConfigResponse,
  MediahubVariableQuery
} from './types';

// -------------------------------------------------------------
// 1. Extend the Abstract Variable Support Class
// -------------------------------------------------------------
export class MediahubVariableSupport extends CustomVariableSupport<DataSource, MediahubVariableQuery> {
  editor = VariableQueryEditor;

  constructor(private datasource: DataSource) {
    super();
  }

  query(request: DataQueryRequest<MediahubVariableQuery>): Observable<DataQueryResponse> {
    const query = request.targets[0];
    
    // Use RxJS 'from' to convert the Promise into an Observable
    const promise = this.datasource.metricFindQuery(query, request).then((data) => {
      return { data };
    });

    return from(promise);
  }
}

export class DataSource extends DataSourceWithBackend<MediahubQuery, MediahubDataSourceOptions> {

  annotations = {};
  
  constructor(instanceSettings: DataSourceInstanceSettings<MediahubDataSourceOptions>) {
    super(instanceSettings);
    // 2. Wire up the instantiated Variable Support subclass here
    this.variables = new MediahubVariableSupport(this);
  }

  applyTemplateVariables(query: MediahubQuery, scopedVars: ScopedVars): Record<string, any> {
    const templateSrv = getTemplateSrv();
    return {
      ...query,
      databaseId: query.databaseId ? templateSrv.replace(query.databaseId, scopedVars) : query.databaseId,
      entryId: query.entryId ? templateSrv.replace(query.entryId, scopedVars) : query.entryId,
    };
  }

  async getMediahubConfig(): Promise<MediahubConfigResponse> {
    return this.getResource('config');
  }

  filterQuery(query: MediahubQuery): boolean {
    if (!query.databaseId) { return false; }
    if ((query.model === 'get preview' || query.model === 'get entry') && 
        query.targetSelection === 'get ID' && !query.entryId) {
      return false;
    }
    return true;
  }

  // 3. Keep our data fetching logic intact
  async metricFindQuery(query: any, options?: any): Promise<MetricFindValue[]> {
    if (!query) {
      return [];
    }

    let command = '';
    let dbId = '';
    let useTimeFilter = false;

    if (typeof query === 'string') {
      const raw = query.trim().toLowerCase();
      if (raw.startsWith('entries ')) {
        command = 'entries';
        dbId = raw.substring(8).trim();
      } else {
        command = raw;
      }
    } else {
      command = query.command || 'databases';
      dbId = query.databaseId || '';
      useTimeFilter = query.useTimeFilter || false;
    }

    const templateSrv = getTemplateSrv();
    const interpolatedDbId = templateSrv.replace(dbId, options?.scopedVars);

    if (command === 'databases') {
      const config = await this.getMediahubConfig();
      return config.databases.map((db) => ({ text: db.name, value: db.id }));
    }

    if (command === 'database_names') {
      const config = await this.getMediahubConfig();
      return config.databases.map((db) => ({ text: db.name, value: db.name }));
    }

    if (command === 'database_ulids') {
      const config = await this.getMediahubConfig();
      return config.databases.map((db) => ({ text: db.id, value: db.id }));
    }

    if (command === 'entries') {
      if (!interpolatedDbId) {
        return [];
      }
      try {
        let url = `variables/entries/${interpolatedDbId}`;
        
        // Append the dashboard time bounds
        if (useTimeFilter && options?.range) {
          const from = options.range.from.valueOf();
          const to = options.range.to.valueOf();
          url += `?from=${from}&to=${to}`;
        }

        const entries = await this.getResource(url);
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