import React, { useEffect, useState } from 'react';
import { InlineField, Select, Input, Switch, Spinner } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { MediahubQuery, MediahubDataSourceOptions, defaultQuery, MediahubConfigResponse } from '../types';

type Props = QueryEditorProps<DataSource, MediahubQuery, MediahubDataSourceOptions>;

export function QueryEditor(props: Props) {
  const { onChange, onRunQuery, datasource } = props;
  const query = { ...defaultQuery, ...props.query };
  
  const [config, setConfig] = useState<MediahubConfigResponse | null>(null);
  const [loading, setLoading] = useState(true);

  // Fetch dynamic configuration on component mount
  useEffect(() => {
    datasource.getMediahubConfig().then((res) => {
      setConfig(res);
      setLoading(false);
    }).catch((err) => {
      console.error("Failed to load MediaHub config", err);
      setLoading(false);
    });
  }, [datasource]);

  if (loading) {
    return <div><Spinner /> Loading MediaHub Configuration...</div>;
  }

  // Generate Database Options
  const dbOptions = config?.databases.map(db => ({
    label: `${db.name} (${db.content_type})`,
    value: db.id,
    contentType: db.content_type
  })) || [];

  const selectedDb = dbOptions.find(db => db.value === query.databaseId);

  // Generate Model Options based on Admin status and Content Type
  const modelOptions = [
    { label: 'Get Metadata Table', value: 'get metadata table' },
    { label: 'Get Entry', value: 'get entry' },
  ];

  // Hide preview if it's explicitly a generic file
  if (selectedDb?.contentType !== 'file') {
    modelOptions.splice(1, 0, { label: 'Get Preview', value: 'get preview' });
  }

  // Only show audit logs to admins
  if (config?.is_admin) {
    modelOptions.push({ label: 'Get Audit Logs', value: 'get audit logs' });
  }

  const targetOptions = [
    { label: 'Get ID', value: 'get ID' },
    { label: 'Get Last', value: 'get last' },
    { label: 'Get Last In Range', value: 'get last in range' },
  ];

  // Universal handler for updating query state
  const onQueryPropChange = (prop: keyof MediahubQuery, value: any, runQuery = false) => {
    onChange({ ...query, [prop]: value });
    if (runQuery) {
      onRunQuery();
    }
  };

  return (
    <div>
      <div className="gf-form">
        <InlineField label="Database" labelWidth={14} grow>
          <Select
            options={dbOptions}
            value={query.databaseId}
            onChange={(v) => onQueryPropChange('databaseId', v.value, true)}
            placeholder="Select a database"
          />
        </InlineField>
        
        <InlineField label="Model" labelWidth={14} grow>
          <Select
            options={modelOptions}
            value={query.model}
            onChange={(v) => onQueryPropChange('model', v.value, true)}
          />
        </InlineField>
      </div>

      {/* CONDITIONAL RENDER: Metadata Table & Audit Logs */}
      {(query.model === 'get metadata table' || query.model === 'get audit logs') && (
        <div className="gf-form">
          <InlineField label="Limit" labelWidth={14}>
            <Input
              type="number"
              value={query.limit}
              onChange={(e) => onQueryPropChange('limit', parseInt(e.currentTarget.value, 10) || 50)}
              onBlur={onRunQuery}
              width={10}
            />
          </InlineField>

          {query.model === 'get metadata table' && (
            <>
              {selectedDb?.contentType !== 'file' && (
                <InlineField label="Add Preview Link" labelWidth={20}>
                  <Switch
                    value={query.addPreviewLink}
                    onChange={(e) => onQueryPropChange('addPreviewLink', e.currentTarget.checked, true)}
                  />
                </InlineField>
              )}
              <InlineField label="Add Entry Link" labelWidth={20}>
                <Switch
                  value={query.addEntryLink}
                  onChange={(e) => onQueryPropChange('addEntryLink', e.currentTarget.checked, true)}
                />
              </InlineField>
            </>
          )}
        </div>
      )}

      {/* CONDITIONAL RENDER: Previews & Entries */}
      {(query.model === 'get preview' || query.model === 'get entry') && (
        <div className="gf-form">
          <InlineField label="Target" labelWidth={14}>
            <Select
              options={targetOptions}
              value={query.targetSelection}
              onChange={(v) => onQueryPropChange('targetSelection', v.value, true)}
              width={20}
            />
          </InlineField>

          {query.targetSelection === 'get ID' && (
            <InlineField label="Entry ID" labelWidth={12} tooltip="Integer ID or variable like ${entry_id}">
              <Input
                value={query.entryId || ''}
                onChange={(e) => onQueryPropChange('entryId', e.currentTarget.value)}
                onBlur={onRunQuery}
                placeholder="1234"
                width={15}
              />
            </InlineField>
          )}

          <InlineField label="Base64 Content" labelWidth={18} tooltip="Return raw base64 instead of a URL proxy link.">
            <Switch
              value={query.base64}
              onChange={(e) => onQueryPropChange('base64', e.currentTarget.checked, true)}
            />
          </InlineField>

          {query.model === 'get entry' && (
            <InlineField label="Max File Size (MB)" labelWidth={20} tooltip="Fails the query if the remote file exceeds this size.">
              <Input
                type="number"
                value={query.maxFileSize}
                onChange={(e) => onQueryPropChange('maxFileSize', parseFloat(e.currentTarget.value) || 4)}
                onBlur={onRunQuery}
                width={10}
              />
            </InlineField>
          )}
        </div>
      )}
    </div>
  );
}