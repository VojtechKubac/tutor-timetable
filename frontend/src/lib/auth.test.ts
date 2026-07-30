import { get } from 'svelte/store';
import { beforeEach, describe, expect, it } from 'vitest';
import { authLoading, teacher } from './stores/auth';
import type { Teacher } from './types';

const sampleTeacher: Teacher = {
	id: 't1',
	email: 'teacher@example.com',
	name: 'Music Teacher',
	created_at: '2026-01-01T00:00:00Z'
};

describe('auth stores', () => {
	beforeEach(() => {
		teacher.set(null);
		authLoading.set(true);
	});

	it('starts unauthenticated and loading', () => {
		expect(get(teacher)).toBeNull();
		expect(get(authLoading)).toBe(true);
	});

	it('stores the signed-in teacher and clears loading', () => {
		teacher.set(sampleTeacher);
		authLoading.set(false);

		expect(get(teacher)).toEqual(sampleTeacher);
		expect(get(authLoading)).toBe(false);
	});

	it('clears teacher on logout', () => {
		teacher.set(sampleTeacher);
		teacher.set(null);

		expect(get(teacher)).toBeNull();
	});
});
