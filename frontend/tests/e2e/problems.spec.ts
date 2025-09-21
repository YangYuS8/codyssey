import { test, expect } from '@playwright/test';

// 假设无需登录即可访问 problems（若需登录后续可补充登录流程）

test.describe('Problems List', () => {
  test('should load problems page and show table headers', async ({ page }) => {
    await page.goto('/problems');
    await expect(page.getByRole('heading', { name: '题目列表' })).toBeVisible();
    await expect(page.getByText('难度')).toBeVisible();
  });
});
