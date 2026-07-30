import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { api, setApiBase } from './api';

function jsonResponse(body: unknown, status = 200): Response {
	return new Response(JSON.stringify(body), {
		status,
		headers: { 'Content-Type': 'application/json' }
	});
}

describe('api client', () => {
	const fetchMock = vi.fn();

	beforeEach(() => {
		setApiBase('');
		fetchMock.mockReset();
		vi.stubGlobal('fetch', fetchMock);
	});

	afterEach(() => {
		vi.unstubAllGlobals();
	});

	it('setApiBase prefixes request URLs', async () => {
		setApiBase('http://localhost:8081');
		fetchMock.mockResolvedValueOnce(jsonResponse({ id: 't1', name: 'Teacher' }));

		await api.auth.login('a@b.com', 'secret');

		expect(fetchMock).toHaveBeenCalledOnce();
		const [url, init] = fetchMock.mock.calls[0]!;
		expect(url).toBe('http://localhost:8081/auth/login');
		expect(init).toMatchObject({
			method: 'POST',
			credentials: 'include'
		});
		expect(JSON.parse(init.body as string)).toEqual({
			email: 'a@b.com',
			password: 'secret'
		});
	});

	it('uses relative paths when base is empty', async () => {
		fetchMock.mockResolvedValueOnce(jsonResponse([]));

		await api.students.list();

		const [url] = fetchMock.mock.calls[0]!;
		expect(url).toBe('/students');
	});

	it('throws backend error codes on non-OK responses', async () => {
		fetchMock.mockResolvedValueOnce(jsonResponse({ error: 'UNAUTHORIZED' }, 401));

		await expect(api.teacher.me()).rejects.toThrow('UNAUTHORIZED');
	});

	it('returns undefined for 204 responses', async () => {
		fetchMock.mockResolvedValueOnce(new Response(null, { status: 204 }));

		await expect(api.auth.logout()).resolves.toBeUndefined();
	});

	it('falls back to PARSE_ERROR when body is not JSON', async () => {
		fetchMock.mockResolvedValueOnce(
			new Response('not-json', {
				status: 500,
				headers: { 'Content-Type': 'text/plain' }
			})
		);

		await expect(api.timetable.get()).rejects.toThrow('PARSE_ERROR');
	});
});
