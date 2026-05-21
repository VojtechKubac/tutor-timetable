import { writable } from 'svelte/store';
import type { Teacher } from '../types';

export const teacher = writable<Teacher | null>(null);
export const authLoading = writable(true);
