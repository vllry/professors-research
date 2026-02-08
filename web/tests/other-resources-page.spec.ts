import { test, expect } from '@playwright/test';

test.describe('Other Resources Page', () => {
  test('should list resources with external links', async ({ page }) => {
    await page.goto('/other-resources', { waitUntil: 'domcontentloaded' });

    await expect(page.getByRole('heading', { name: 'Other Resources' })).toBeVisible({
      timeout: 10000,
    });

    const resources = [
      { title: 'TrainerHill', url: 'https://www.trainerhill.com' },
      { title: 'Limitless', url: 'https://limitlesstcg.com' },
      { title: 'Limitless Labs', url: 'https://labs.limitlesstcg.com' },
      { title: 'Japan City League Results', url: 'https://limitlesstcg.com/tournaments/jp' },
    ];

    for (const resource of resources) {
      await expect(
        page.getByRole('heading', { name: resource.title, level: 2, exact: true })
      ).toBeVisible();
      const urlLink = page.getByRole('link', { name: resource.url, exact: true });
      await expect(urlLink).toHaveAttribute('href', resource.url);
      await expect(urlLink).toHaveAttribute('target', '_blank');
    }
  });
});

