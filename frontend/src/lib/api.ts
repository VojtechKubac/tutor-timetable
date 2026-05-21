import type { Teacher, TeacherSettings, Student, AvailabilitySlot, Lesson } from './types';

// In dev the Vite proxy handles routing to localhost:8080.
// In production (Docker) PUBLIC_API_URL is set at runtime via SvelteKit's $env/dynamic/public.
// We keep this file simple and use a relative base so the proxy works in dev.
let BASE = '';

export function setApiBase(url: string) {
	BASE = url || '';
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
	const res = await fetch(`${BASE}${path}`, {
		method,
		headers: body ? { 'Content-Type': 'application/json' } : {},
		credentials: 'include',
		body: body !== undefined ? JSON.stringify(body) : undefined
	});

	if (res.status === 204) return undefined as T;

	const data = await res.json().catch(() => ({ error: 'PARSE_ERROR' }));

	if (!res.ok) {
		throw new Error(data.error ?? 'UNKNOWN_ERROR');
	}
	return data as T;
}

export const api = {
	auth: {
		login: (email: string, password: string) =>
			request<{ id: string; name: string }>('POST', '/auth/login', { email, password }),
		logout: () => request<void>('POST', '/auth/logout')
	},
	teacher: {
		me: () => request<Teacher>('GET', '/teacher/me'),
		getSettings: () => request<TeacherSettings>('GET', '/teacher/settings'),
		updateSettings: (s: Partial<TeacherSettings>) =>
			request<void>('PUT', '/teacher/settings', s),
		getAvailability: () => request<AvailabilitySlot[]>('GET', '/teacher/availability'),
		updateAvailability: (slots: AvailabilitySlot[]) =>
			request<void>('PUT', '/teacher/availability', slots)
	},
	students: {
		list: () => request<Student[]>('GET', '/students'),
		create: (s: Partial<Student>) => request<Student>('POST', '/students', s),
		get: (id: string) => request<Student>('GET', `/students/${id}`),
		update: (id: string, s: Partial<Student>) => request<void>('PUT', `/students/${id}`, s),
		remove: (id: string) => request<void>('DELETE', `/students/${id}`),
		getAvailability: (id: string) =>
			request<AvailabilitySlot[]>('GET', `/students/${id}/availability`),
		updateAvailability: (id: string, slots: AvailabilitySlot[]) =>
			request<void>('PUT', `/students/${id}/availability`, slots)
	},
	timetable: {
		get: () => request<Lesson[]>('GET', '/timetable'),
		generate: () => request<Lesson[]>('POST', '/timetable/generate'),
		moveLesson: (id: string, day: number, start: string, end: string) =>
			request<void>('PUT', `/timetable/lessons/${id}`, {
				day_of_week: day,
				start_time: start,
				end_time: end
			}),
		togglePin: (id: string) => request<void>('PATCH', `/timetable/lessons/${id}/pin`)
	}
};
