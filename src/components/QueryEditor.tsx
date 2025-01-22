import React  from 'react';
import { CodeEditor } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { FelderaOptions, FelderaQuery } from '../types';

type Props = QueryEditorProps<DataSource, FelderaQuery, FelderaOptions>;

export function QueryEditor({ query, onChange, onRunQuery }: Props) {
  const onQueryTextChange = (text: string) => {
    onChange({ ...query, queryText: text });
    onRunQuery();
  };

  return (
      <CodeEditor
        width=""
        height="100px"
        language="sql"
        value={query.queryText ?? ''}
        onSave={onQueryTextChange}
        onBlur={onQueryTextChange}
        showMiniMap={false}
        showLineNumbers={true} 
      />
  );
}
