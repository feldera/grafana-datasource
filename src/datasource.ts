import { DataSourceInstanceSettings, ScopedVars } from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';

import { FelderaQuery, FelderaOptions } from './types';

export class DataSource extends DataSourceWithBackend<FelderaQuery, FelderaOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<FelderaOptions>) {
    super(instanceSettings);
  }

  applyTemplateVariables(query: FelderaQuery, scopedVars: ScopedVars) {
    return {
      ...query,
      queryText: getTemplateSrv().replace(query.queryText, scopedVars),
    };
  }

  filterQuery(query: FelderaQuery): boolean {
    // if no query has been provided, prevent the query from being executed
    return !!query.queryText;
  }
}
