import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		port: 3001,
		strictPort: true,
		proxy: {
			// Proxy API calls to the Go backend in dev mode
			'/auth': 'http://localhost:8081',
			'/teacher': 'http://localhost:8081',
			'/students': 'http://localhost:8081',
			'/timetable': 'http://localhost:8081'
		}
	}
});
