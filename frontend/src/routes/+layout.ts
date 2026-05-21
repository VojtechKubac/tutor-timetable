import { setupI18n } from '$lib/i18n';
import { browser } from '$app/environment';
import { waitLocale } from 'svelte-i18n';

export const ssr = false; // Pure SPA — avoids API base URL complexity in Phase 1

export async function load() {
	if (browser) {
		setupI18n();
		await waitLocale(); // wait for locale JSON to load before rendering
	}
}
