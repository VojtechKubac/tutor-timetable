import { test, expect } from '@playwright/test';
import {
	apiAuthed,
	loginViaAPI,
	loginViaUI,
	paintMorningAvailability,
	type AvailabilitySlot
} from '../helpers/auth';

const weekMorning: AvailabilitySlot[] = [
	{ day_of_week: 0, start_time: '09:00', end_time: '16:00' },
	{ day_of_week: 1, start_time: '09:00', end_time: '16:00' },
	{ day_of_week: 2, start_time: '09:00', end_time: '16:00' }
];

test.describe('timetable generate and pin', () => {
	test('generate places lessons; pin survives regenerate', async ({ page, request }) => {
		await loginViaAPI(request);

		await apiAuthed(request, 'PUT', '/teacher/availability', weekMorning);

		const studentName = `Pinned Student ${Date.now()}`;
		const student = (await apiAuthed(request, 'POST', '/students', {
			name: studentName,
			email: '',
			notes: ''
		})) as { id: string };
		await apiAuthed(request, 'PUT', `/students/${student.id}/availability`, weekMorning);

		// Second student so regenerate still has work to do
		const other = (await apiAuthed(request, 'POST', '/students', {
			name: `Other Student ${Date.now()}`,
			email: '',
			notes: ''
		})) as { id: string };
		await apiAuthed(request, 'PUT', `/students/${other.id}/availability`, weekMorning);

		await loginViaUI(page);
		await page.getByTestId('timetable-generate').click();
		await page.getByTestId('timetable-grid').waitFor();

		const lesson = page.locator(`[data-testid="lesson"][data-student-name="${studentName}"]`);
		await expect(lesson.first()).toBeVisible();
		const startBefore = await lesson.first().getAttribute('data-start');
		const dayBefore = await lesson.first().getAttribute('data-day');

		await lesson.first().getByTestId('lesson-pin').click();
		await expect(lesson.first()).toHaveAttribute('data-pinned', 'true');

		await page.getByTestId('timetable-generate').click();
		await page.getByTestId('timetable-grid').waitFor();

		const pinned = page.locator(
			`[data-testid="lesson"][data-student-name="${studentName}"][data-pinned="true"]`
		);
		await expect(pinned.first()).toBeVisible();
		await expect(pinned.first()).toHaveAttribute('data-start', startBefore!);
		await expect(pinned.first()).toHaveAttribute('data-day', dayBefore!);

		// Unpin via UI for hygiene
		await pinned.first().getByTestId('lesson-pin').click();
		await expect(
			page.locator(`[data-testid="lesson"][data-student-name="${studentName}"]`).first()
		).toHaveAttribute('data-pinned', 'false');
	});
});

// Keep a light UI-only path for availability → generate without relying solely on API setup
test.describe('timetable UI generate smoke', () => {
	test('generate after painting availability via UI', async ({ page }) => {
		await loginViaUI(page);
		await page.getByTestId('nav-settings').click();
		// Use Thursday (day 3) — often empty after API fixtures that fill Mon–Wed
		await paintMorningAvailability(page, 3);
		await page.getByTestId('settings-save-availability').click();
		await expect(page.getByTestId('saved-indicator').first()).toBeVisible();

		const name = `Gen Student ${Date.now()}`;
		await page.getByTestId('nav-students').click();
		await page.getByTestId('students-add').click();
		await page.getByTestId('students-name').fill(name);
		await page.getByTestId('students-save').click();
		await page
			.locator('[data-testid="student-row"]')
			.filter({ hasText: name })
			.getByTestId('student-link')
			.click();
		await paintMorningAvailability(page, 3);
		await page.getByTestId('student-save-availability').click();
		await expect(page.getByTestId('saved-indicator').first()).toBeVisible();

		await page.getByTestId('nav-timetable').click();
		await page.getByTestId('timetable-generate').click();
		await expect(page.getByTestId('timetable-grid')).toBeVisible();
		await expect(
			page.locator(`[data-testid="lesson"][data-student-name="${name}"]`).first()
		).toBeVisible();
	});
});
