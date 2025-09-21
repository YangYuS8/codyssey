import { test, expect } from '@playwright/test';

// 占位：真实提交流程需要后端可写接口及种子题目/鉴权，这里只验证页面元素存在

test.describe('Submission Flow Placeholder', () => {
  test('visit submissions list', async ({ page }) => {
    await page.goto('/submissions');
    await expect(page.getByRole('heading', { name: 'Submissions' })).toBeVisible();
  });
});
