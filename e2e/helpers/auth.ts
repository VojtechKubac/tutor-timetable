import type { APIRequestContext, Page } from '@playwright/test';
import { apiURL, seedEmail, seedPassword } from './env';

export async function loginViaUI(page: Page, email = seedEmail, password = seedPassword) {
	await page.goto('/login');
	await page.getByTestId('login-email').fill(email);
	await page.getByTestId('login-password').fill(password);
	await page.getByTestId('login-submit').click();
	await page.waitForURL((url) => url.pathname === '/');
	await page.getByTestId('nav-logout').waitFor();
}

/** API login returning the Set-Cookie header value for auth_token. */
export async function loginViaAPI(request: APIRequestContext) {
	const res = await request.post(`${apiURL}/auth/login`, {
		data: { email: seedEmail, password: seedPassword }
	});
	if (!res.ok()) {
		throw new Error(`API login failed: ${res.status()} ${await res.text()}`);
	}
	return res;
}

export type AvailabilitySlot = {
	day_of_week: number;
	start_time: string;
	end_time: string;
};

export async function apiAuthed(
	request: APIRequestContext,
	method: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE',
	path: string,
	data?: unknown
) {
	const res = await request.fetch(`${apiURL}${path}`, {
		method,
		headers: data !== undefined ? { 'Content-Type': 'application/json' } : undefined,
		data
	});
	if (!res.ok() && res.status() !== 204) {
		throw new Error(`${method} ${path} failed: ${res.status()} ${await res.text()}`);
	}
	if (res.status() === 204) return undefined;
	const text = await res.text();
	return text ? JSON.parse(text) : undefined;
}

/** Ensure Mon–style morning window is painted (cells are 30 min from 07:00; 09:00–12:00 = 4..9). */
export async function paintMorningAvailability(page: Page, day = 0) {
	const editor = page.getByTestId('availability-editor');
	await editor.waitFor();
	for (let c = 4; c <= 9; c++) {
		const cell = editor.locator(
			`[data-testid="availability-cell"][data-day="${day}"][data-cell="${c}"]`
		);
		await cell.scrollIntoViewIfNeeded();
		if ((await cell.getAttribute('data-active')) !== 'true') {
			await cell.click({ force: true });
		}
	}
	await expectActive(editor, day, 4);
}

async function expectActive(
	editor: ReturnType<Page['getByTestId']>,
	day: number,
	cell: number
) {
	await editor
		.locator(
			`[data-testid="availability-cell"][data-day="${day}"][data-cell="${cell}"][data-active="true"]`
		)
		.waitFor();
}
