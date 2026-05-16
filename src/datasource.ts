import {
  DataSourceInstanceSettings,
} from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';

import { 
  MediahubQuery, 
  MediahubDataSourceOptions, 
  MediahubConfigResponse 
} from './types';

export class DataSource extends DataSourceWithBackend<MediahubQuery, MediahubDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<MediahubDataSourceOptions>) {
    super(instanceSettings);
  }

  // Fetches the dynamic configuration (Admin status & Databases) for the Query Editor.
  async getMediahubConfig(): Promise<MediahubConfigResponse> {
    // getResource is inherited from DataSourceWithBackend and automatically routes
    // to our Go httpadapter ServeMux
    return this.getResource('/config');
  }
}