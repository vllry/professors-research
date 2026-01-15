import { test, expect } from '@playwright/test';
import { readFileSync } from 'fs';
import { join } from 'path';

test.describe('Start Page', () => {
  test('should calculate start odds for a pasted decklist', async ({ page }) => {
    // Navigate directly to the start page
    await page.goto('/start', { waitUntil: 'domcontentloaded' });

    // Wait for React to hydrate and render
    await page.waitForSelector('h1', { timeout: 15000 });

    // Verify we're on the start page
    await expect(
      page.getByRole('heading', { name: 'Calculate Starting Hand Probabilities' })
    ).toBeVisible({ timeout: 10000 });

    // Read the test decklist
    // In Docker, testdata is mounted at /testdata; locally, it's relative to project root
    const decklistPath = process.env.CI
      ? '/testdata/dragapult_MEG_1.txt'
      : join(process.cwd(), '../testdata/dragapult_MEG_1.txt');
    const decklist = readFileSync(decklistPath, 'utf-8');

    // Find the decklist textarea and paste the decklist
    const decklistTextarea = page.getByLabel('Decklist (paste from Live)');
    await decklistTextarea.fill(decklist);

    // Verify the decklist was pasted
    await expect(decklistTextarea).toHaveValue(decklist);

    // Click the "Calculate Start Odds" button
    const calculateButton = page.getByRole('button', { name: 'Calculate' });
    await expect(calculateButton).toBeEnabled();
    await calculateButton.click();

    // Wait for the loading state to appear (may appear quickly or not at all if calculation is fast)
    const loadingText = page.getByText('Calculating start odds...');
    try {
      await expect(loadingText).toBeVisible({ timeout: 2000 });
    } catch {
      // Loading text may not appear if calculation is very fast, that's okay
    }

    // Wait for the results to appear (loading should disappear and results should show)
    await expect(loadingText).not.toBeVisible({ timeout: 30000 });

    // Verify the results summary is displayed
    await expect(page.getByRole('heading', { name: 'Start Summary' })).toBeVisible();

    // Verify individual card start odds table is displayed
    await expect(
      page.getByRole('heading', { name: 'Individual Card Start Odds' })
    ).toBeVisible();

    const oddsTable = page.locator('table').first();
    await expect(oddsTable).toBeVisible();

    // Verify the table has at least one data row (not just headers)
    const tableRows = oddsTable.locator('tbody tr');
    await expect(tableRows.first()).toBeVisible();

    // Check that percentage values are displayed (format: XX.XX%)
    const percentagePattern = /\d+\.\d+%/;
    const tableContent = await oddsTable.textContent();
    expect(tableContent).toMatch(percentagePattern);
  });
});


