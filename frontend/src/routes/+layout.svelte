<script lang="ts">
	import '../app.css';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { _ } from 'svelte-i18n';
	import { api } from '$lib/api';
	import { teacher, authLoading } from '$lib/stores/auth';

	onMount(async () => {
		try {
			const me = await api.teacher.me();
			teacher.set(me);
		} catch {
			teacher.set(null);
		} finally {
			authLoading.set(false);
		}
	});

	$: isLoginPage = $page.url.pathname === '/login';
	$: if (!$authLoading && !$teacher && !isLoginPage) {
		goto('/login');
	}

	async function logout() {
		await api.auth.logout().catch(() => {});
		teacher.set(null);
		goto('/login');
	}
</script>

{#if $authLoading}
	<div class="flex h-screen items-center justify-center text-gray-500">
		{$_('common.loading')}
	</div>
{:else if isLoginPage}
	<slot />
{:else if $teacher}
	<div class="flex h-screen bg-gray-50">
		<!-- Sidebar -->
		<nav class="flex w-52 flex-col border-r border-gray-200 bg-white px-4 py-6">
			<span class="mb-8 text-lg font-semibold text-indigo-600">Tutor Timetable</span>

			<a
				href="/"
				data-testid="nav-timetable"
				class="mb-1 rounded-md px-3 py-2 text-sm font-medium transition-colors"
				class:bg-indigo-50={$page.url.pathname === '/'}
				class:text-indigo-700={$page.url.pathname === '/'}
				class:text-gray-700={$page.url.pathname !== '/'}
				class:hover:bg-gray-100={$page.url.pathname !== '/'}
			>
				{$_('nav.timetable')}
			</a>

			<a
				href="/students"
				data-testid="nav-students"
				class="mb-1 rounded-md px-3 py-2 text-sm font-medium transition-colors"
				class:bg-indigo-50={$page.url.pathname.startsWith('/students')}
				class:text-indigo-700={$page.url.pathname.startsWith('/students')}
				class:text-gray-700={!$page.url.pathname.startsWith('/students')}
				class:hover:bg-gray-100={!$page.url.pathname.startsWith('/students')}
			>
				{$_('nav.students')}
			</a>

			<a
				href="/settings"
				data-testid="nav-settings"
				class="mb-1 rounded-md px-3 py-2 text-sm font-medium transition-colors"
				class:bg-indigo-50={$page.url.pathname === '/settings'}
				class:text-indigo-700={$page.url.pathname === '/settings'}
				class:text-gray-700={$page.url.pathname !== '/settings'}
				class:hover:bg-gray-100={$page.url.pathname !== '/settings'}
			>
				{$_('nav.settings')}
			</a>

			<div class="mt-auto">
				<p class="mb-2 truncate px-3 text-xs text-gray-400" data-testid="teacher-name">{$teacher.name}</p>
				<button
					on:click={logout}
					data-testid="nav-logout"
					class="w-full rounded-md px-3 py-2 text-left text-sm text-gray-500 hover:bg-gray-100"
				>
					{$_('nav.logout')}
				</button>
			</div>
		</nav>

		<!-- Main content -->
		<main class="flex-1 overflow-auto p-8">
			<slot />
		</main>
	</div>
{/if}
