import React from 'react';
import { InlineField, Select, Input, Switch } from '@grafana/ui';
import { MediahubVariableQuery } from '../types';

interface Props {
  query: MediahubVariableQuery;
  onChange: (query: MediahubVariableQuery, definition: string) => void;
}

const commandOptions = [
  { label: 'Databases (Name -> ID)', value: 'databases' },
  { label: 'Database Names (Name -> Name)', value: 'database_names' },
  { label: 'Database ULIDs (ID -> ID)', value: 'database_ulids' },
  { label: 'Entries', value: 'entries' },
];

export function VariableQueryEditor({ query, onChange }: Props) {
  // Fallback to 'databases' if creating a brand new variable
  const currentCommand = query.command || 'databases';

  const onCommandChange = (v: string) => {
    // The second parameter is the summary definition displayed in Grafana's variable table
    onChange({ ...query, command: v as any }, `Command: ${v}`);
  };

  const onDatabaseIdChange = (v: string) => {
    onChange({ ...query, databaseId: v }, `Entries for: ${v}`);
  };

  const onFilterChange = (v: boolean) => {
    onChange({ ...query, useTimeFilter: v }, `Entries for: ${query.databaseId || '?'}`);
  };

  return (
    <div className="gf-form-group">
      <div className="gf-form">
        <InlineField label="Command" labelWidth={14} grow>
          <Select
            options={commandOptions}
            value={currentCommand}
            onChange={(v) => onCommandChange(v.value!)}
          />
        </InlineField>
      </div>

      {currentCommand === 'entries' && (
        <>
          <div className="gf-form">
            <InlineField label="Database ID" labelWidth={14} grow tooltip="Enter a ULID or a variable like $database">
              <Input
                value={query.databaseId || ''}
                onChange={(e) => onDatabaseIdChange(e.currentTarget.value)}
                placeholder="01HGFB..."
              />
            </InlineField>
          </div>
          <div className="gf-form">
            <InlineField label="Filter by Time" labelWidth={20} tooltip="Only fetch entries created within the dashboard time range.">
              <Switch
                value={query.useTimeFilter || false}
                onChange={(e) => onFilterChange(e.currentTarget.checked)}
              />
            </InlineField>
          </div>
        </>
      )}
    </div>
  );
}