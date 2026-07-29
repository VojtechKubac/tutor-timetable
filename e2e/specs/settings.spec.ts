import { test, expect } from '@playwright/test';
import { loginViaUI } from '../helpers/auth';

test.describe('settings', () => {
	test('settings load and persist after save', async ({ page }) => {
		await loginViaUI(page);
		await page.getByTestId('nav-settings').click();
		await page.waitForURL('**/settings');

		const maxGap = page.getByTestId('settings-max-gap');
		await maxGap.waitFor();
		await maxGap.fill('45');
		await page.getByTestId('settings-save').click();
		await expect(page.getByTestId('saved-indicator').first()).toBeVisible();

		await page.reload();
		await expect(page.getByTestId('settings-max-gap')).toHaveValue('45');

		// Restore a common default so later runs stay predictable
		await page.getByTestId('settings-max-gap').fill('30');
		await page.getByTestId('settings-save').click();
		await expect(page.getByTestId('saved-indicator').first()).toBeVisible();
	});
});
