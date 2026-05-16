import React, { ChangeEvent } from 'react';
import { InlineField, Input, SecretInput } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { MediahubDataSourceOptions, MediahubSecureJsonData } from '../types';

interface Props extends DataSourcePluginOptionsEditorProps<MediahubDataSourceOptions, MediahubSecureJsonData> {}

export function ConfigEditor(props: Props) {
  const { onOptionsChange, options } = props;
  const jsonData = options.jsonData;
  const secureJsonFields = options.secureJsonFields || {};

  const onURLChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        url: event.target.value,
      },
    });
  };

  const onUsernameChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        username: event.target.value,
      },
    });
  };

  const onPasswordChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: {
        ...(options.secureJsonData || {}),
        password: event.target.value,
      },
    });
  };

  const onResetPassword = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        password: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        password: '',
      },
    });
  };

  return (
    <div className="gf-form-group">
      <h4>MediaHub API Settings</h4>
      
      <InlineField label="URL" labelWidth={14} tooltip="The base URL of your MediaHub instance (e.g., http://mediahub:8080)">
        <Input
          onChange={onURLChange}
          value={jsonData.url || ''}
          placeholder="http://localhost:8080"
          width={40}
        />
      </InlineField>

      <InlineField label="Username" labelWidth={14}>
        <Input
          onChange={onUsernameChange}
          value={jsonData.username || ''}
          placeholder="admin"
          width={40}
        />
      </InlineField>

      <InlineField label="Password" labelWidth={14}>
        <SecretInput
          isConfigured={secureJsonFields.password}
          value={options.secureJsonData?.password || ''}
          placeholder="Enter password"
          width={40}
          onReset={onResetPassword}
          onChange={onPasswordChange}
        />
      </InlineField>
    </div>
  );
}