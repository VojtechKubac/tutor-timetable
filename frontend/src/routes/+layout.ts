import { setupI18n } from '$lib/i18n';
import { browser } from '$app/environment';
import { waitLocale } from 'svelte-i18n';
import { env } from '$env/dynamic/public';
import { setApiBase } from '$lib/api';

export const ssr = false; // Pure SPA — avoids SSR API base URL complexity in Phase 1

export async function load() {
	if (browser) {
		// Empty in Vite dev (relative URLs + proxy); set in Docker via PUBLIC_API_URL
		setApiBase(env.PUBLIC_API_URL ?? '');
		setupI18n();
		await waitLocale(); // wait for locale JSON to load before rendering
	}
}
