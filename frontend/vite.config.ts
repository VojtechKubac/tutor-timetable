import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		proxy: {
			// Proxy API calls to the Go backend in dev mode
			'/auth': 'http://localhost:8080',
			'/teacher': 'http://localhost:8080',
			'/students': 'http://localhost:8080',
			'/timetable': 'http://localhost:8080'
		}
	}
});
