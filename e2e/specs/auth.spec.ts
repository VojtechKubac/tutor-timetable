import { test, expect } from '@playwright/test';
import { loginViaUI } from '../helpers/auth';
import { seedEmail, seedPassword } from '../helpers/env';

test.describe('auth', () => {
	test('unauthenticated users are redirected to login', async ({ page }) => {
		await page.goto('/');
		await page.waitForURL('**/login');
		await expect(page.getByTestId('login-submit')).toBeVisible();
	});

	test('login and logout with seed credentials', async ({ page }) => {
		await loginViaUI(page, seedEmail, seedPassword);
		await expect(page.getByTestId('teacher-name')).toBeVisible();
		await expect(page.getByTestId('timetable-generate')).toBeVisible();

		await page.getByTestId('nav-logout').click();
		await page.waitForURL('**/login');
		await expect(page.getByTestId('login-submit')).toBeVisible();

		await page.goto('/');
		await page.waitForURL('**/login');
	});
});
