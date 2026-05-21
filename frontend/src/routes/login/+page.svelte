<script lang="ts">
	import { goto } from '$app/navigation';
	import { _ } from 'svelte-i18n';
	import { api } from '$lib/api';
	import { teacher } from '$lib/stores/auth';

	let email = '';
	let password = '';
	let error = '';
	let loading = false;

	async function submit() {
		error = '';
		loading = true;
		try {
			const me = await api.auth.login(email, password);
			// Fetch full teacher profile after login
			const profile = await api.teacher.me();
			teacher.set(profile);
			goto('/');
		} catch (e: unknown) {
			const code = e instanceof Error ? e.message : 'default';
			const key = `login.error.${code}`;
			error = $_(`login.error.${code}`) !== key ? $_(`login.error.${code}`) : $_('login.error.default');
		} finally {
			loading = false;
		}
	}
</script>

<div class="flex min-h-screen items-center justify-center bg-gray-50">
	<div class="w-full max-w-sm rounded-xl border border-gray-200 bg-white p-8 shadow-sm">
		<h1 class="mb-6 text-2xl font-semibold text-gray-900">{$_('login.title')}</h1>

		<form on:submit|preventDefault={submit} class="space-y-4">
			<div>
				<label class="mb-1 block text-sm font-medium text-gray-700" for="email">
					{$_('login.email')}
				</label>
				<input
					id="email"
					type="email"
					bind:value={email}
					required
					class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
				/>
			</div>

			<div>
				<label class="mb-1 block text-sm font-medium text-gray-700" for="password">
					{$_('login.password')}
				</label>
				<input
					id="password"
					type="password"
					bind:value={password}
					required
					class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
				/>
			</div>

			{#if error}
				<p class="text-sm text-red-600">{error}</p>
			{/if}

			<button
				type="submit"
				disabled={loading}
				class="w-full rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50"
			>
				{loading ? $_('common.loading') : $_('login.submit')}
			</button>
		</form>
	</div>
</div>
