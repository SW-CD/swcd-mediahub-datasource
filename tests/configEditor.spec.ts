import { test, expect } from '@grafana/plugin-e2e';
import { MediahubDataSourceOptions, MediahubSecureJsonData } from '../src/types';

test('smoke: should render config editor', async ({ createDataSourceConfigPage, readProvisionedDataSource, page }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await createDataSourceConfigPage({ type: ds.type });
  await expect(page.getByPlaceholder('http://localhost:8080')).toBeVisible();
});

test('"Save & test" should be successful when configuration is valid', async ({
  createDataSourceConfigPage,
  readProvisionedDataSource,
  page,
}) => {
  const ds = await readProvisionedDataSource<MediahubDataSourceOptions, MediahubSecureJsonData>({ fileName: 'datasources.yml' });
  const configPage = await createDataSourceConfigPage({ type: ds.type });
  
  // Intercept the health check so it doesn't try to ping a non-existent server!
  await page.route('**/api/datasources/uid/*/health', (route) => {
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ status: 'success', message: 'Data source is working' }),
    });
  });

  await page.getByPlaceholder('http://localhost:8080').fill(ds.jsonData.url ?? 'http://localhost:8080');
  await page.getByPlaceholder('admin').fill(ds.jsonData.username ?? 'admin');
  await page.getByPlaceholder('Enter password').fill(ds.secureJsonData?.password ?? 'admin_password');
  
  await expect(configPage.saveAndTest()).toBeOK();
});

test('"Save & test" should fail when configuration is invalid', async ({
  createDataSourceConfigPage,
  readProvisionedDataSource,
  page,
}) => {
  const ds = await readProvisionedDataSource<MediahubDataSourceOptions, MediahubSecureJsonData>({ fileName: 'datasources.yml' });
  const configPage = await createDataSourceConfigPage({ type: ds.type });
  
  await page.getByPlaceholder('http://localhost:8080').fill(ds.jsonData.url ?? 'http://localhost:8080');
  await page.getByPlaceholder('admin').fill(ds.jsonData.username ?? 'admin');
  
  await expect(configPage.saveAndTest()).not.toBeOK();
  await expect(configPage).toHaveAlert('error'); 
});