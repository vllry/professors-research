import { test, expect } from '@playwright/test';

test.describe('Draw Supporters Page', () => {
  test('should calculate draw supporter odds', async ({ page }) => {
    await page.goto('/draw-supporters', { waitUntil: 'domcontentloaded' });

    await expect(page.getByRole('heading', { name: 'Draw Supporters' })).toBeVisible({
      timeout: 10000,
    });

    await page.getByLabel('Number of cards in deck').fill('10');
    await page.getByLabel('Known cards at bottom of deck').fill('5');
    await page.getByLabel('Number of cards in hand').fill('7');
    await page.getByLabel('Prize cards left').fill('6');

    const calculateButton = page.getByRole('button', { name: 'Calculate' });
    await expect(calculateButton).toBeEnabled();
    await calculateButton.click();

    const loadingText = page.getByText('Calculating draw odds...');
    try {
      await expect(loadingText).toBeVisible({ timeout: 2000 });
    } catch {
      // Loading may not appear if calculation is very fast.
    }

    await expect(loadingText).not.toBeVisible({ timeout: 30000 });

    await expect(page.getByRole('heading', { name: 'Odds of drawing 1+ copies' })).toBeVisible();

    const table = page.locator('table').first();
    await expect(table).toBeVisible();

    // Sanity check: percentage formatting is present.
    const percentagePattern = /\d+\.\d+%/;
    const tableContent = await table.textContent();
    expect(tableContent).toMatch(percentagePattern);

    // Spot check one deterministic value: Iono with 1 copy, 6 draws from 5 (top of deck) => 100.00%
    await expect(table).toContainText('Iono');
    await expect(table).toContainText('100.00%');

    // Pair odds tables are behind a discrete selection.
    await page.getByLabel('Show pair-odds tables').check();
    await expect(page.getByRole('heading', { name: 'Iono', level: 3 })).toBeVisible();
    await expect(page.getByRole('heading', { name: "Lillie's Determination", level: 3 })).toBeVisible();
    await expect(
      page.getByRole('columnheader', { name: 'Card A copies \\ Card B copies' }).first()
    ).toBeVisible();

    await expect(page.getByRole('heading', { name: 'Bottom cards (drawn into)' })).toBeVisible();
    await expect(page.getByText('Bottom draw 1')).toBeVisible();
  });
});

