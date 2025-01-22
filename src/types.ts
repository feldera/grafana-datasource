import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

export interface FelderaQuery extends DataQuery {
  queryText?: string;
}

/**
 * These are options configured for each DataSource instance
 */
export interface FelderaOptions extends DataSourceJsonData {
  baseUrl?: string;
  pipeline?: string;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface FelderaSecureJsonData {
  apiKey?: string;
}
