import { DataSourcePlugin } from '@grafana/data';
import { DataSource } from './datasource';
import { ConfigEditor } from './components/ConfigEditor';
import { QueryEditor } from './components/QueryEditor';
import { MediahubQuery, MediahubDataSourceOptions } from './types';

// This binds everything together so Grafana can render it
export const plugin = new DataSourcePlugin<DataSource, MediahubQuery, MediahubDataSourceOptions>(DataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);