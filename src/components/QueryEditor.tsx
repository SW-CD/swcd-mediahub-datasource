import React, { useEffect, useState } from 'react';
import { InlineField, Select, Input, Switch, Spinner } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { getTemplateSrv } from '@grafana/runtime';
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

  // 1. Generate Database Options from the Backend
  const backendDbOptions = config?.databases.map(db => ({
    label: `${db.name} (${db.content_type})`,
    value: db.id,
    contentType: db.content_type
  })) || [];

  // 2. Fetch Dashboard Variables and format them as options
  const variableOptions = getTemplateSrv().getVariables().map(v => ({
    label: `$${v.name}`,
    value: `$${v.name}`,
    contentType: 'unknown' // We don't know the type until runtime
  }));

  // 3. Combine them so the user sees both in the dropdown
  const allDbOptions = [...backendDbOptions, ...variableOptions];

  const selectedDb = allDbOptions.find(db => db.value === query.databaseId);

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

  // Dedicated handler for database switching to enforce content-type rules
  const onDatabaseChange = (v: any) => {
    const newDbId = v.value;
    const dbInfo = allDbOptions.find(db => db.value === newDbId);
    const contentType = dbInfo?.contentType || 'unknown';

    // Start with a copy of the current query state
    let updatedQuery = { ...query, databaseId: newDbId };

    // Apply strict defaults based on the content type
    switch (contentType) {
      case 'file':
        // 1. Force disable preview generation
        updatedQuery.addPreviewLink = false;
        
        // 2. Safety Net: If the user had the "Get Preview" model selected for a previous DB,
        // kick them back to a safe model since generic files don't support previews.
        if (updatedQuery.model === 'get preview') {
          updatedQuery.model = 'get metadata table';
        }
        break;
        
      case 'image':
      case 'video':
      case 'audio':
        // Reset to true for rich media types when switching databases
        updatedQuery.addPreviewLink = true;
        break;
        
      case 'unknown':
        // If a variable is used (e.g., $my_db), we leave the current state alone
        // and let the user decide with the UI toggles.
        break;
    }

    // Push the sanitized query state to Grafana and run it
    onChange(updatedQuery);
    onRunQuery();
  };

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
            options={allDbOptions}
            value={query.databaseId}
            onChange={onDatabaseChange}
            placeholder="Select a database or variable"
            allowCustomValue={true}
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
        <>
          {/* Row 1: Target Selection */}
          <div className="gf-form">
            <InlineField label="Target" labelWidth={14}>
              <Select
                options={targetOptions}
                value={query.targetSelection || 'get last'}
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
          </div>

          {/* Row 2: Payload Modifiers */}
          <div className="gf-form">
            <InlineField label="Base64 Content" labelWidth={20} tooltip="Return raw base64 instead of a URL proxy link.">
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
        </>
      )}
    </div>
  );
}