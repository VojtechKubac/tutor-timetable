<script lang="ts">
	import { onMount } from 'svelte';
	import { _ } from 'svelte-i18n';
	import { api } from '$lib/api';
	import type { TeacherSettings, AvailabilitySlot } from '$lib/types';
	import AvailabilityEditor from '$lib/components/AvailabilityEditor.svelte';

	let settings: TeacherSettings | null = null;
	let availability: AvailabilitySlot[] = [];
	let loading = true;
	let saving = false;
	let savedSettings = false;
	let savedAvail = false;
	let error = '';

	onMount(async () => {
		try {
			[settings, availability] = await Promise.all([
				api.teacher.getSettings(),
				api.teacher.getAvailability()
			]);
		} catch {
			error = $_('common.error');
		} finally {
			loading = false;
		}
	});

	async function saveSettings() {
		if (!settings) return;
		saving = true;
		try {
			await api.teacher.updateSettings(settings);
			savedSettings = true;
			setTimeout(() => (savedSettings = false), 2000);
		} catch {
			error = $_('common.error');
		} finally {
			saving = false;
		}
	}

	async function saveAvailability() {
		saving = true;
		try {
			await api.teacher.updateAvailability(availability);
			savedAvail = true;
			setTimeout(() => (savedAvail = false), 2000);
		} catch {
			error = $_('common.error');
		} finally {
			saving = false;
		}
	}
</script>

<h1 class="mb-6 text-2xl font-semibold text-gray-900">{$_('settings.title')}</h1>

{#if error}
	<p class="mb-4 text-sm text-red-600">{error}</p>
{/if}

{#if loading}
	<p class="text-gray-500">{$_('common.loading')}</p>
{:else if settings}
	<!-- Scheduling settings -->
	<section class="mb-8 rounded-xl border border-gray-200 bg-white p-6 shadow-sm">
		<h2 class="mb-4 text-sm font-semibold uppercase tracking-wide text-gray-500">
			{$_('settings.workingHours')}
		</h2>
		<form on:submit|preventDefault={saveSettings} class="space-y-4 max-w-lg">
			<div class="grid grid-cols-2 gap-4">
				<div>
					<label class="mb-1 block text-sm font-medium text-gray-700">{$_('settings.workingStart')}</label>
					<input
						type="time"
						bind:value={settings.working_start}
						step="300"
						class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
					/>
				</div>
				<div>
					<label class="mb-1 block text-sm font-medium text-gray-700">{$_('settings.workingEnd')}</label>
					<input
						type="time"
						bind:value={settings.working_end}
						step="300"
						class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
					/>
				</div>
			</div>

			<div class="grid grid-cols-2 gap-4">
				<div>
					<label class="mb-1 block text-sm font-medium text-gray-700">{$_('settings.lessonDuration')}</label>
					<input
						type="number"
						bind:value={settings.lesson_duration_minutes}
						min="5" max="180" step="5"
						class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
					/>
				</div>
				<div>
					<label class="mb-1 block text-sm font-medium text-gray-700">{$_('settings.maxGap')}</label>
					<input
						type="number"
						bind:value={settings.max_gap_minutes}
						min="0" max="120" step="5"
						class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
					/>
				</div>
			</div>

			<div class="grid grid-cols-3 gap-4">
				<div>
					<label class="mb-1 block text-sm font-medium text-gray-700">{$_('settings.maxConsecutive')}</label>
					<input
						type="number"
						bind:value={settings.max_consecutive_lessons}
						min="1" max="20"
						class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
					/>
				</div>
				<div>
					<label class="mb-1 block text-sm font-medium text-gray-700">{$_('settings.breakAfterN')}</label>
					<input
						type="number"
						bind:value={settings.break_after_n_lessons}
						min="0" max="20"
						class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
					/>
				</div>
				<div>
					<label class="mb-1 block text-sm font-medium text-gray-700">{$_('settings.breakDuration')}</label>
					<input
						type="number"
						bind:value={settings.break_duration_minutes}
						min="5" max="60" step="5"
						class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
					/>
				</div>
			</div>

			<div class="flex items-center gap-3">
				<button
					type="submit"
					disabled={saving}
					class="rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50"
				>
					{$_('settings.save')}
				</button>
				{#if savedSettings}
					<span class="text-sm text-green-600">{$_('settings.saved')}</span>
				{/if}
			</div>
		</form>
	</section>

	<!-- Teacher availability -->
	<section class="rounded-xl border border-gray-200 bg-white p-6 shadow-sm">
		<h2 class="mb-4 text-sm font-semibold uppercase tracking-wide text-gray-500">
			{$_('settings.availability')}
		</h2>
		<AvailabilityEditor bind:slots={availability} />
		<div class="mt-4 flex items-center gap-3">
			<button
				on:click={saveAvailability}
				disabled={saving}
				class="rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50"
			>
				{$_('availability.save')}
			</button>
			{#if savedAvail}
				<span class="text-sm text-green-600">{$_('settings.saved')}</span>
			{/if}
		</div>
	</section>
{/if}
