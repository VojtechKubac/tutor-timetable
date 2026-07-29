import { test, expect } from '@playwright/test';
import { loginViaUI } from '../helpers/auth';

test.describe('students CRUD', () => {
	test.beforeEach(async ({ page }) => {
		await loginViaUI(page);
		await page.getByTestId('nav-students').click();
		await page.waitForURL('**/students');
	});

	test('create, edit, and delete a student', async ({ page }) => {
		const name = `E2E Student ${Date.now()}`;
		const edited = `${name} Edited`;

		await page.getByTestId('students-add').click();
		await page.getByTestId('students-name').fill(name);
		await page.getByTestId('students-email').fill('e2e-student@example.com');
		await page.getByTestId('students-notes').fill('playwright');
		await page.getByTestId('students-save').click();

		const row = page.locator('[data-testid="student-row"]').filter({ hasText: name });
		await expect(row).toBeVisible();

		await row.getByTestId('student-link').click();
		await page.getByTestId('student-name').waitFor();
		await page.getByTestId('student-name').fill(edited);
		await page.getByTestId('student-save-info').click();
		await expect(page.getByTestId('saved-indicator').first()).toBeVisible();

		await page.getByTestId('nav-students').click();
		await page.waitForURL('**/students');
		const editedRow = page.locator('[data-testid="student-row"]').filter({ hasText: edited });
		await expect(editedRow).toBeVisible();

		page.once('dialog', (dialog) => dialog.accept());
		await editedRow.getByTestId('student-delete').click();
		await expect(editedRow).toHaveCount(0);
	});
});
