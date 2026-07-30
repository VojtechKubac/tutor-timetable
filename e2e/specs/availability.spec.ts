import { test, expect } from '@playwright/test';
import { loginViaUI, paintMorningAvailability } from '../helpers/auth';

test.describe('availability editors', () => {
	test.beforeEach(async ({ page }) => {
		await loginViaUI(page);
	});

	test('teacher availability save and reload', async ({ page }) => {
		await page.getByTestId('nav-settings').click();
		await page.waitForURL('**/settings');
		await paintMorningAvailability(page, 0);
		await page.getByTestId('settings-save-availability').click();
		await expect(page.getByTestId('saved-indicator').first()).toBeVisible();

		await page.reload();
		await page.getByTestId('availability-editor').waitFor();
		await expect(
			page.locator('[data-testid="availability-cell"][data-day="0"][data-cell="4"][data-active="true"]')
		).toBeVisible();
	});

	test('student availability save and reload', async ({ page }) => {
		const name = `Avail Student ${Date.now()}`;
		await page.getByTestId('nav-students').click();
		await page.getByTestId('students-add').click();
		await page.getByTestId('students-name').fill(name);
		await page.getByTestId('students-save').click();

		const row = page.locator('[data-testid="student-row"]').filter({ hasText: name });
		await row.getByTestId('student-link').click();
		await paintMorningAvailability(page, 0);
		await page.getByTestId('student-save-availability').click();
		await expect(page.getByTestId('saved-indicator').first()).toBeVisible();

		await page.reload();
		await page.getByTestId('availability-editor').waitFor();
		await expect(
			page.locator('[data-testid="availability-cell"][data-day="0"][data-cell="4"][data-active="true"]')
		).toBeVisible();
	});
});
