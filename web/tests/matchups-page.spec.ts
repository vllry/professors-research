import { test, expect } from '@playwright/test';

test.describe('Matchups Page', () => {
  test('should load page and display tournament checkboxes', async ({ page }) => {
    await page.goto('/matchups', { waitUntil: 'domcontentloaded' });
    await page.waitForSelector('h1', { timeout: 15000 });

    await expect(
      page.getByRole('heading', { name: 'Matchup Stats' })
    ).toBeVisible({ timeout: 10000 });

    // Tournament checkboxes should load from the API
    await expect(page.getByText('Curitiba 2026')).toBeVisible({ timeout: 10000 });

    // Archetype input should be present
    await expect(page.getByLabel('Archetype')).toBeVisible();

    // Submit button should be disabled when archetype is empty
    const submitButton = page.getByRole('button', { name: 'Get Matchups' });
    await expect(submitButton).toBeDisabled();
  });

  test('should fetch and display matchup stats', async ({ page }) => {
    await page.goto('/matchups', { waitUntil: 'domcontentloaded' });
    await page.waitForSelector('h1', { timeout: 15000 });

    // Wait for tournaments to load
    await expect(page.getByText('Curitiba 2026')).toBeVisible({ timeout: 10000 });

    // Tournament should be checked by default (all selected on load)
    const checkbox = page.getByRole('checkbox').first();
    await expect(checkbox).toBeChecked();

    // Select an archetype
    const archetypeInput = page.getByLabel('Archetype');
    await archetypeInput.selectOption('Dragapult Dusknoir');

    // Submit
    const submitButton = page.getByRole('button', { name: 'Get Matchups' });
    await expect(submitButton).toBeEnabled();
    await submitButton.click();

    // Wait for loading to finish
    const loadingText = page.getByText('Fetching matchup data...');
    try {
      await expect(loadingText).toBeVisible({ timeout: 2000 });
    } catch {
      // May complete too fast to see
    }
    await expect(loadingText).not.toBeVisible({ timeout: 30000 });

    // Results should appear with a matchup table
    const resultsTable = page.locator('table').first();
    await expect(resultsTable).toBeVisible();

    // Table should have opponent column header
    await expect(page.getByRole('columnheader', { name: 'Opponent' })).toBeVisible();

    // Table should have data rows with win rate percentages
    const tableContent = await resultsTable.textContent();
    expect(tableContent).toMatch(/\d+\.\d+%/);

    // Metagame breakdown should appear
    await expect(
      page.getByRole('heading', { name: 'Metagame Breakdown' })
    ).toBeVisible();
  });

  test('should keep submit button disabled when no archetype is selected', async ({ page }) => {
    await page.goto('/matchups', { waitUntil: 'domcontentloaded' });
    await page.waitForSelector('h1', { timeout: 15000 });

    // Wait for tournaments to load
    await expect(page.getByText('Curitiba 2026')).toBeVisible({ timeout: 10000 });

    // Without selecting an archetype the submit button must remain disabled
    const submitButton = page.getByRole('button', { name: 'Get Matchups' });
    await expect(submitButton).toBeDisabled();
  });

  test('should support placement filters', async ({ page }) => {
    await page.goto('/matchups', { waitUntil: 'domcontentloaded' });
    await page.waitForSelector('h1', { timeout: 15000 });

    await expect(page.getByText('Curitiba 2026')).toBeVisible({ timeout: 10000 });

    // Select archetype
    await page.getByLabel('Archetype').selectOption('Dragapult Dusknoir');

    // Set placement filters
    await page.getByLabel('Player Placement (top %)').fill('50');
    await page.getByLabel('Opponent Placement (top %)').fill('50');

    // Submit
    const submitButton = page.getByRole('button', { name: 'Get Matchups' });
    await submitButton.click();

    // Wait for loading to finish
    const loadingText = page.getByText('Fetching matchup data...');
    await expect(loadingText).not.toBeVisible({ timeout: 30000 });

    // Results table should appear
    const resultsTable = page.locator('table').first();
    await expect(resultsTable).toBeVisible();
  });

  test('should toggle variant filters section', async ({ page }) => {
    await page.goto('/matchups', { waitUntil: 'domcontentloaded' });
    await page.waitForSelector('h1', { timeout: 15000 });

    // Variant filters should be collapsed by default
    const variantToggle = page.getByText('Variant Filters (optional)');
    await expect(variantToggle).toBeVisible();

    // "Add variant" button should not be visible
    await expect(page.getByText('+ Add variant')).not.toBeVisible();

    // Click to expand
    await variantToggle.click();

    // Now the add variant button should be visible
    await expect(page.getByText('+ Add variant')).toBeVisible();

    // Add a variant
    await page.getByText('+ Add variant').click();
    await expect(page.getByText('Variant 0')).toBeVisible();

    // Should have a card name input and count input
    await expect(page.getByPlaceholder('Card name')).toBeVisible();
  });

  test('should populate form fields from URL params on load', async ({ page }) => {
    await page.goto('/matchups?a=Dragapult+Dusknoir&pp=50&op=25', { waitUntil: 'domcontentloaded' });
    await page.waitForSelector('h1', { timeout: 15000 });

    // Wait for tournaments to load
    await expect(page.getByText('Curitiba 2026')).toBeVisible({ timeout: 10000 });

    // Archetype select should reflect the URL param
    await expect(page.getByLabel('Archetype')).toHaveValue('Dragapult Dusknoir');

    // Placement fields should be pre-filled from URL
    await expect(page.getByLabel('Player Placement (top %)')).toHaveValue('50');
    await expect(page.getByLabel('Opponent Placement (top %)')).toHaveValue('25');
  });

  test('should expand variant panel and populate cards from URL', async ({ page }) => {
    // sv=1 opens panel; v encodes "Fire Starter:2,Charizard:1"
    await page.goto('/matchups?sv=1&v=Fire%20Starter%3A2%2CCharizard%3A1', { waitUntil: 'domcontentloaded' });
    await page.waitForSelector('h1', { timeout: 15000 });

    // Variant panel should be expanded
    await expect(page.getByText('+ Add variant')).toBeVisible({ timeout: 10000 });

    // The first variant's cards should be populated
    const cardInputs = page.getByPlaceholder('Card name');
    await expect(cardInputs.first()).toHaveValue('Fire Starter');
    await expect(cardInputs.nth(1)).toHaveValue('Charizard');

    // Count inputs should also match
    const countInputs = page.locator('input[type="number"][min="1"]');
    await expect(countInputs.first()).toHaveValue('2');
    await expect(countInputs.nth(1)).toHaveValue('1');
  });

  test('should update URL when archetype is selected', async ({ page }) => {
    await page.goto('/matchups', { waitUntil: 'domcontentloaded' });
    await page.waitForSelector('h1', { timeout: 15000 });
    await expect(page.getByText('Curitiba 2026')).toBeVisible({ timeout: 10000 });

    await page.getByLabel('Archetype').selectOption('Dragapult Dusknoir');

    // URL should update to include the archetype param
    await expect(page).toHaveURL(/a=Dragapult/);
  });

  test('should update URL when placement filters are set', async ({ page }) => {
    await page.goto('/matchups', { waitUntil: 'domcontentloaded' });
    await page.waitForSelector('h1', { timeout: 15000 });
    await expect(page.getByText('Curitiba 2026')).toBeVisible({ timeout: 10000 });

    await page.getByLabel('Player Placement (top %)').fill('75');
    await page.getByLabel('Player Placement (top %)').press('Tab');

    await expect(page).toHaveURL(/pp=75/);
  });

  test('should update URL with tournament IDs when a tournament is deselected', async ({ page }) => {
    await page.goto('/matchups', { waitUntil: 'domcontentloaded' });
    await page.waitForSelector('h1', { timeout: 15000 });
    await expect(page.getByText('Curitiba 2026')).toBeVisible({ timeout: 10000 });

    // Initially all tournaments selected — URL should have no 't' params
    await expect(page).not.toHaveURL(/[?&]t=/);

    // Deselect the first tournament
    await page.getByRole('checkbox').first().uncheck();

    // URL should now include 't' params for the remaining selected tournaments
    await expect(page).toHaveURL(/[?&]t=/);
  });

  test('should restore tournament selection from URL params', async ({ page }) => {
    // Navigate with only one specific tournament selected
    await page.goto('/matchups?t=CU01wDygvn34WEPNJ3ou', { waitUntil: 'domcontentloaded' });
    await page.waitForSelector('h1', { timeout: 15000 });
    await expect(page.getByText('Curitiba 2026')).toBeVisible({ timeout: 10000 });

    // The Curitiba checkbox should be checked
    const checkboxes = page.getByRole('checkbox');
    const curitibaCheckbox = checkboxes.first();
    await expect(curitibaCheckbox).toBeChecked();
  });

  test('should update URL when variant panel is opened', async ({ page }) => {
    await page.goto('/matchups', { waitUntil: 'domcontentloaded' });
    await page.waitForSelector('h1', { timeout: 15000 });

    // Open variant panel
    await page.getByText('Variant Filters (optional)').click();

    // URL should include sv=1
    await expect(page).toHaveURL(/sv=1/);

    // Close it
    await page.getByText('Variant Filters (optional)').click();

    // URL should no longer include sv=1
    await expect(page).not.toHaveURL(/sv=1/);
  });

  test('should show W/L/T columns with fewer than 3 variants', async ({ page }) => {
    await page.goto('/matchups', { waitUntil: 'domcontentloaded' });
    await page.waitForSelector('h1', { timeout: 15000 });
    await expect(page.getByText('Curitiba 2026')).toBeVisible({ timeout: 10000 });

    await page.getByLabel('Archetype').selectOption('Dragapult Dusknoir');
    await page.getByRole('button', { name: 'Get Matchups' }).click();

    const loadingText = page.getByText('Fetching matchup data...');
    await expect(loadingText).not.toBeVisible({ timeout: 30000 });

    // With no variant filters there is only one variant key — W/L/T should be visible
    await expect(page.getByRole('columnheader', { name: 'W', exact: true })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: 'L', exact: true })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: 'T', exact: true })).toBeVisible();
  });

  test('should hide W/L/T columns when 3 or more variants are present', async ({ page }) => {
    // Mock the matchup-stats API to return 3 variant keys
    const mockResponse = {
      matchups: {
        '0': {
          cardCounts: { 'Dragapult ex': 1 },
          matchups: { 'Charizard ex': { wins: 10, losses: 5, ties: 1, winRate: 0.64 } },
        },
        '1': {
          cardCounts: { 'Dusknoir': 1 },
          matchups: { 'Charizard ex': { wins: 3, losses: 7, ties: 0, winRate: 0.3 } },
        },
        'other': {
          cardCounts: {},
          matchups: { 'Charizard ex': { wins: 2, losses: 2, ties: 0, winRate: 0.5 } },
        },
      },
      archetypeCounts: { 'Charizard ex': 30 },
      variantCounts: { '0': 20, '1': 10, 'other': 5 },
    };
    await page.route('**/matchup-stats', (route) => {
      route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(mockResponse) });
    });

    await page.goto('/matchups', { waitUntil: 'domcontentloaded' });
    await page.waitForSelector('h1', { timeout: 15000 });
    await expect(page.getByText('Curitiba 2026')).toBeVisible({ timeout: 10000 });

    await page.getByLabel('Archetype').selectOption('Dragapult Dusknoir');
    await page.getByRole('button', { name: 'Get Matchups' }).click();

    await expect(page.locator('table').first()).toBeVisible({ timeout: 10000 });

    // With 3 variant keys, W/L/T headers should not appear
    await expect(page.getByRole('columnheader', { name: 'W', exact: true })).not.toBeVisible();
    await expect(page.getByRole('columnheader', { name: 'L', exact: true })).not.toBeVisible();
    await expect(page.getByRole('columnheader', { name: 'T', exact: true })).not.toBeVisible();

    // Total and Win Rate should still be visible
    await expect(page.getByRole('columnheader', { name: 'Total' }).first()).toBeVisible();
    await expect(page.getByRole('columnheader', { name: 'Win Rate' }).first()).toBeVisible();
  });

  test('should navigate via navbar link', async ({ page }) => {
    await page.goto('/', { waitUntil: 'domcontentloaded' });

    // Wait for the navbar to render
    await expect(page.getByRole('link', { name: 'Matchups' }).first()).toBeVisible({ timeout: 15000 });

    // Click the Matchups nav link
    await page.getByRole('link', { name: 'Matchups' }).first().click();

    await expect(
      page.getByRole('heading', { name: 'Matchup Stats' })
    ).toBeVisible({ timeout: 10000 });
  });
});
