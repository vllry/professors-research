import { test, expect } from '@playwright/test';
import { readFileSync } from 'fs';
import { join } from 'path';

test.describe('Prize Page', () => {
  test('should calculate odds for a pasted decklist', async ({ page }) => {
    // Navigate directly to the prizes page
    await page.goto('/prizes', { waitUntil: 'domcontentloaded' });
    
    // Wait for React to hydrate and render
    await page.waitForSelector('h1', { timeout: 15000 });
    
    // Verify we're on the prize page
    await expect(page.getByRole('heading', { name: 'Calculate Prize Probabilities' })).toBeVisible({ timeout: 10000 });

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

    // Click the "Calculate" button
    const calculateButton = page.getByRole('button', { name: 'Calculate' });
    await expect(calculateButton).toBeEnabled();
    await calculateButton.click();

    // Wait for the loading state to appear (may appear quickly or not at all if calculation is fast)
    const loadingText = page.getByText('Calculating prize odds...');
    try {
      await expect(loadingText).toBeVisible({ timeout: 2000 });
    } catch {
      // Loading text may not appear if calculation is very fast, that's okay
    }

    // Wait for the results to appear (loading should disappear and results should show)
    await expect(loadingText).not.toBeVisible({ timeout: 30000 });

    // Verify the results are displayed
    await expect(page.getByRole('heading', { name: 'Individual Card Odds' })).toBeVisible();

    // Verify that at least one card odds row is displayed
    const oddsTable = page.locator('table').first();
    await expect(oddsTable).toBeVisible();

    // Verify the table has at least one data row (not just headers)
    const tableRows = oddsTable.locator('tbody tr');
    await expect(tableRows.first()).toBeVisible();

    // Verify that the table contains card names and probability values
    // The table should have cells with percentages
    const firstDataRow = tableRows.first();
    await expect(firstDataRow.locator('td').first()).toBeVisible();
    
    // Check that percentage values are displayed (format: XX.XX%)
    const percentagePattern = /\d+\.\d+%/;
    const tableContent = await oddsTable.textContent();
    expect(tableContent).toMatch(percentagePattern);
  });

  test('should calculate odds with CardSet using full text entry and autocomplete', async ({ page }) => {
    // Navigate directly to the prizes page
    await page.goto('/prizes', { waitUntil: 'domcontentloaded' });
    
    // Wait for React to hydrate and render
    await page.waitForSelector('h1', { timeout: 15000 });
    
    // Verify we're on the prize page
    await expect(page.getByRole('heading', { name: 'Calculate Prize Probabilities' })).toBeVisible({ timeout: 10000 });

    // Read the test decklist
    // In Docker, testdata is mounted at /testdata; locally, it's relative to project root
    const decklistPath = process.env.CI 
      ? '/testdata/dragapult_MEG_1.txt'
      : join(process.cwd(), '../testdata/dragapult_MEG_1.txt');
    const decklist = readFileSync(decklistPath, 'utf-8');

    // Find the decklist textarea and paste the decklist
    const decklistTextarea = page.getByLabel('Decklist (paste from Live)');
    await decklistTextarea.fill(decklist);

    // Expand the Card Sets section
    await page.getByRole('button', { name: '+ Add Card Groups (Advanced)' }).click();

    // Verify we're in builder mode (default)
    await expect(page.getByRole('button', { name: 'Builder' })).toHaveClass(/bg-blue-600/);

    // Add a new group
    const groupNameInput = page.getByPlaceholder('Group name (e.g., draw_support)');
    await groupNameInput.fill('draw_support');
    await page.getByRole('button', { name: 'Add Group' }).click();

    // Verify the group was added
    await expect(page.getByText('draw_support')).toBeVisible();

    // Add a card set to the group (click "Add Card Set" button)
    await page.getByRole('button', { name: '+ Add Card Set' }).click();

    // Verify a card set was added and wait for it to be visible
    await expect(page.getByText('Card Set 1')).toBeVisible();

    // Add an AnyOf pattern to the card set (required before adding cards)
    await page.getByRole('button', { name: '+ Add AnyOf' }).click();
    // Wait for the AnyOf pattern header to appear
    await expect(page.getByText('AnyOf Pattern')).toBeVisible({ timeout: 5000 });

    // Add a card using full text entry
    // The pattern should be expanded by default
    const addCardButtons = page.getByRole('button', { name: '+ Add Card' });
    await expect(addCardButtons.first()).toBeVisible();
    await addCardButtons.first().click();

    // Wait for the card input to appear
    // Find the card input field and enter a card using full text
    const cardInputs = page.locator('input[placeholder="Card Name SetCode Number"]');
    await expect(cardInputs.first()).toBeVisible({ timeout: 10000 });
    const firstCardInput = cardInputs.first();
    await firstCardInput.fill('Iono PAL 185');
    await firstCardInput.blur(); // Trigger validation

    // Verify the card was entered (input should be valid - no red border)
    await expect(firstCardInput).toHaveValue('Iono PAL 185');

    // Add another card using autocomplete
    await addCardButtons.first().click();

    // Find the second card input
    await expect(cardInputs.nth(1)).toBeVisible({ timeout: 10000 });
    const secondCardInput = cardInputs.nth(1);
    
    // Type part of a card name to trigger autocomplete
    await secondCardInput.fill('Lill');
    
    // Wait for autocomplete dropdown to appear and find the card
    // The autocomplete shows buttons with card name as the main text
    const lillieOption = page.locator('button').filter({ hasText: /Lillie.*Determination/ }).first();
    await expect(lillieOption).toBeVisible({ timeout: 5000 });
    
    // Click on the autocomplete option
    await lillieOption.click();
    await page.waitForTimeout(300);

    // Verify the card was selected (should show the full card string)
    await expect(secondCardInput).toHaveValue(/Lillie.*MEG.*119/);

    // Click the "Calculate" button
    const calculateButton = page.getByRole('button', { name: 'Calculate' });
    await expect(calculateButton).toBeEnabled();
    await calculateButton.click();

    // Wait for the loading state to appear (may appear quickly or not at all if calculation is fast)
    const loadingText2 = page.getByText('Calculating prize odds...');
    try {
      await expect(loadingText2).toBeVisible({ timeout: 2000 });
    } catch {
      // Loading text may not appear if calculation is very fast, that's okay
    }

    // Wait for the results to appear (loading should disappear and results should show)
    await expect(loadingText2).not.toBeVisible({ timeout: 30000 });

    // Verify the individual card odds are displayed
    await expect(page.getByRole('heading', { name: 'Individual Card Odds' })).toBeVisible();

    // Verify the Card Set odds are displayed
    const cardSetOddsHeading = page.getByRole('heading', { name: 'Card Set Odds' });
    await expect(cardSetOddsHeading).toBeVisible();

    // Verify the group name appears in the Card Set Odds section
    // The group name is rendered as an h4 heading within the Card Set Odds section
    const cardSetSection = cardSetOddsHeading.locator('..');
    await expect(cardSetSection.getByRole('heading', { name: 'draw_support', level: 4 })).toBeVisible();

    // Verify that percentage values are displayed for Card Set odds
    const cardSetContent = await cardSetSection.textContent();
    const percentagePattern = /\d+\.\d+%/;
    expect(cardSetContent).toMatch(percentagePattern);
  });

  test('should handle multiple AllOf/AnyOf clauses with text fields', async ({ page }) => {
    // Navigate directly to the prizes page
    await page.goto('/prizes', { waitUntil: 'domcontentloaded' });
    
    // Wait for React to hydrate and render
    await page.waitForSelector('h1', { timeout: 15000 });
    
    // Verify we're on the prize page
    await expect(page.getByRole('heading', { name: 'Calculate Prize Probabilities' })).toBeVisible({ timeout: 10000 });

    // Read the test decklist
    // In Docker, testdata is mounted at /testdata; locally, it's relative to project root
    const decklistPath = process.env.CI 
      ? '/testdata/dragapult_MEG_1.txt'
      : join(process.cwd(), '../testdata/dragapult_MEG_1.txt');
    const decklist = readFileSync(decklistPath, 'utf-8');

    // Find the decklist textarea and paste the decklist
    const decklistTextarea = page.getByLabel('Decklist (paste from Live)');
    await decklistTextarea.fill(decklist);

    // Expand the Card Sets section
    await page.getByRole('button', { name: '+ Add Card Groups (Advanced)' }).click();

    // Add a new group
    const groupNameInput = page.getByPlaceholder('Group name (e.g., draw_support)');
    await groupNameInput.fill('test_group');
    await page.getByRole('button', { name: 'Add Group' }).click();

    // Verify the group was added
    await expect(page.getByText('test_group')).toBeVisible();

    // Add a card set to the group
    await page.getByRole('button', { name: '+ Add Card Set' }).click();
    await expect(page.getByText('Card Set 1')).toBeVisible();

    // Add first AnyOf pattern
    await page.getByRole('button', { name: '+ Add AnyOf' }).click();
    // Wait for the first AnyOf pattern header to appear
    await expect(page.getByRole('button', { name: /AnyOf Pattern/ })).toBeVisible({ timeout: 5000 });

    // Add a card to the first AnyOf pattern
    const addCardButtons = page.getByRole('button', { name: '+ Add Card' });
    await expect(addCardButtons.first()).toBeVisible();
    await addCardButtons.first().click();

    // Enter a card in the first pattern
    const cardInputs = page.locator('input[placeholder="Card Name SetCode Number"]');
    await expect(cardInputs.first()).toBeVisible({ timeout: 10000 });
    const firstPatternFirstInput = cardInputs.first();
    await firstPatternFirstInput.fill('Iono PAL 185');
    await firstPatternFirstInput.blur();
    await expect(firstPatternFirstInput).toHaveValue('Iono PAL 185');

    // Add second AnyOf pattern
    await page.getByRole('button', { name: '+ Add AnyOf' }).click();
    // Wait until we have two AnyOf pattern headers
    const anyOfButtons = page.getByRole('button', { name: /AnyOf Pattern/ });
    await expect(anyOfButtons).toHaveCount(2);

    // Add a card to the second AnyOf pattern
    // Find the second "+ Add Card" button (should be in the second pattern)
    const allAddCardButtons = page.getByRole('button', { name: '+ Add Card' });
    await expect(allAddCardButtons.nth(1)).toBeVisible();
    await allAddCardButtons.nth(1).click();

    // Try to enter a card in the second pattern's text field
    // This is where the bug manifests - text fields after the 1st pattern don't work
    const allCardInputs = page.locator('input[placeholder="Card Name SetCode Number"]');
    // The second pattern's input should be the second or third input (depending on if first pattern has one card)
    // Let's find all inputs and try to use one that's not the first
    const inputCount = await allCardInputs.count();
    expect(inputCount).toBeGreaterThanOrEqual(2);

    // Try to fill the second input (which should be in the second pattern)
    const secondPatternInput = allCardInputs.nth(1);
    await expect(secondPatternInput).toBeVisible({ timeout: 10000 });
    
    // Verify the first input still has its value (to ensure they're independent)
    const firstInputValue = await allCardInputs.first().inputValue();
    expect(firstInputValue).toBe('Iono PAL 185');
    
    // Attempt to type in the second pattern's input field
    // The bug: this input may not accept text properly, or typing here may affect the first input
    await secondPatternInput.click();
    await secondPatternInput.fill('Lillie MEG 119');
    await page.waitForTimeout(300);
    
    // Verify the input value - this will fail if the bug exists
    // The bug prevents text from being entered in fields after the 1st pattern
    const secondInputValue = await secondPatternInput.inputValue();
    expect(secondInputValue).toBe('Lillie MEG 119');
    
    // Also verify the first input wasn't affected (bug might cause cross-contamination)
    const firstInputValueAfter = await allCardInputs.first().inputValue();
    expect(firstInputValueAfter).toBe('Iono PAL 185');
    
    // Also test with AllOf patterns
    // Switch to AllOf mode
    await page.getByRole('button', { name: 'Switch to AllOf' }).click();

    // Add another AllOf pattern (there may already be multiple patterns after the type switch)
    await page.getByRole('button', { name: '+ Add AllOf' }).click();
    const allOfButtons = page.getByRole('button', { name: /AllOf Pattern/ });
    const allOfCount = await allOfButtons.count();
    expect(allOfCount).toBeGreaterThanOrEqual(2);

    // Add a card to the second AllOf pattern
    const allOfAddCardButtons = page.getByRole('button', { name: '+ Add Card' });
    await expect(allOfAddCardButtons.nth(1)).toBeVisible();
    await allOfAddCardButtons.nth(1).click();

    // Try to enter a card in the second AllOf pattern's text field
    const allOfCardInputs = page.locator('input[placeholder="Card Name SetCode Number"]');
    const allOfInputCount = await allOfCardInputs.count();
    expect(allOfInputCount).toBeGreaterThanOrEqual(2);

    // Try to fill the second AllOf pattern's input
    const secondAllOfInput = allOfCardInputs.nth(1);
    await expect(secondAllOfInput).toBeVisible({ timeout: 10000 });
    
    // Verify the first AllOf input still has its value (to ensure they're independent)
    const firstAllOfInputValue = await allOfCardInputs.first().inputValue();
    expect(firstAllOfInputValue).toBeTruthy(); // Should have some value from previous test
    
    // Attempt to type in the second AllOf pattern's input field
    // The bug: this input may not accept text properly, or typing here may affect the first input
    await secondAllOfInput.click();
    await secondAllOfInput.fill('Professor Research PAL 240');
    await page.waitForTimeout(300);
    
    // Verify the input value - this will fail if the bug exists
    const secondAllOfInputValue = await secondAllOfInput.inputValue();
    expect(secondAllOfInputValue).toBe('Professor Research PAL 240');
    
    // Also verify the first AllOf input wasn't affected (bug might cause cross-contamination)
    const firstAllOfInputValueAfter = await allOfCardInputs.first().inputValue();
    expect(firstAllOfInputValueAfter).toBe(firstAllOfInputValue);
  });
});

