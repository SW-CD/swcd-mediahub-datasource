import { DataSourcePlugin } from '@grafana/data';
import { DataSource } from './datasource';
import { ConfigEditor } from './components/ConfigEditor';
import { QueryEditor } from './components/QueryEditor';
import { MediahubQuery, MediahubDataSourceOptions } from './types';

// We bind everything EXCEPT the variable editor here
export const plugin = new DataSourcePlugin<DataSource, MediahubQuery, MediahubDataSourceOptions>(DataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);