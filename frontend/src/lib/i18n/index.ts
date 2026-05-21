import { register, init, getLocaleFromNavigator } from 'svelte-i18n';

register('en', () => import('./en.json'));

export function setupI18n(locale = 'en') {
	init({
		fallbackLocale: 'en',
		initialLocale: locale || getLocaleFromNavigator() || 'en'
	});
}
