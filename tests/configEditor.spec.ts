import { test, expect } from '@grafana/plugin-e2e';

test('smoke: should render config editor', async ({ createDataSourceConfigPage, readProvisionedDataSource, page }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await createDataSourceConfigPage({ type: ds.type });
  await expect(page.getByLabel('Base Url')).toBeVisible();
  await expect(page.getByLabel('Pipeline')).toBeVisible();
  await expect(page.getByLabel('API Key')).toBeVisible();
});
