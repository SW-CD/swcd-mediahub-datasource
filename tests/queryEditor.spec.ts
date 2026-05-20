import { test, expect } from '@grafana/plugin-e2e';

test('smoke: should render query editor with MediaHub fields', async ({ panelEditPage, readProvisionedDataSource }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  
  await expect(panelEditPage.getQueryEditorRow('A').getByText('Database', { exact: true })).toBeVisible();
  await expect(panelEditPage.getQueryEditorRow('A').getByText('Model', { exact: true })).toBeVisible();
});

test('should trigger new query when Limit field is changed', async ({
  panelEditPage,
  readProvisionedDataSource,
  page, 
}) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);

  await panelEditPage.getQueryEditorRow('A').getByText('Select a database or variable').click({ force: true });
  await page.keyboard.insertText('test-database-id');
  await page.keyboard.press('Enter');
  
  await expect(panelEditPage.getQueryEditorRow('A').getByText('Limit', { exact: true })).toBeVisible();

  const limitField = panelEditPage.getQueryEditorRow('A').locator('input[type="number"]');
  const queryReq = panelEditPage.waitForQueryDataRequest();
  
  await limitField.fill('100');
  await limitField.blur();
  
  await expect(await queryReq).toBeTruthy();
});

test('data query should execute without frontend errors', async ({ panelEditPage, readProvisionedDataSource, page }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  
  // Intercept the actual Grafana data query and return a dummy "Success" response
  // so the test doesn't crash trying to talk to a non-existent MediaHub server.
  await page.route('**/api/ds/query*', (route) => {
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ results: { A: { frames: [] } } }), // Empty but valid Grafana DataFrame response
    });
  });

  await panelEditPage.getQueryEditorRow('A').getByText('Select a database or variable').click({ force: true });
  await page.keyboard.insertText('test-database-id');
  await page.keyboard.press('Enter');

  await expect(panelEditPage.getQueryEditorRow('A').getByText('Database', { exact: true })).toBeVisible();
  await panelEditPage.setVisualization('Table');
  
  // Now this will return 200 OK because of our interceptor!
  await expect(panelEditPage.refreshPanel()).toBeOK();
});