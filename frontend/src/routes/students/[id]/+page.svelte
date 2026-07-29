<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { _ } from 'svelte-i18n';
	import { api } from '$lib/api';
	import type { Student, AvailabilitySlot } from '$lib/types';
	import AvailabilityEditor from '$lib/components/AvailabilityEditor.svelte';

	const studentId = $page.params.id;

	let student: Student | null = null;
	let availability: AvailabilitySlot[] = [];
	let loading = true;
	let saving = false;
	let savedInfo = false;
	let savedAvail = false;
	let error = '';

	onMount(async () => {
		try {
			[student, availability] = await Promise.all([
				api.students.get(studentId),
				api.students.getAvailability(studentId)
			]);
		} catch {
			error = $_('common.error');
		} finally {
			loading = false;
		}
	});

	async function saveInfo() {
		if (!student) return;
		saving = true;
		try {
			await api.students.update(studentId, {
				name: student.name,
				email: student.email,
				notes: student.notes
			});
			savedInfo = true;
			setTimeout(() => (savedInfo = false), 2000);
		} catch {
			error = $_('common.error');
		} finally {
			saving = false;
		}
	}

	async function saveAvailability() {
		saving = true;
		try {
			await api.students.updateAvailability(studentId, availability);
			savedAvail = true;
			setTimeout(() => (savedAvail = false), 2000);
		} catch {
			error = $_('common.error');
		} finally {
			saving = false;
		}
	}
</script>

<div class="mb-6 flex items-center gap-3">
	<a href="/students" class="text-sm text-gray-500 hover:text-gray-700">← {$_('common.back')}</a>
	<h1 class="text-2xl font-semibold text-gray-900">
		{loading ? $_('common.loading') : (student?.name ?? '')}
	</h1>
</div>

{#if error}
	<p class="mb-4 text-sm text-red-600">{error}</p>
{/if}

{#if !loading && student}
	<!-- Info form -->
	<section class="mb-8 rounded-xl border border-gray-200 bg-white p-6 shadow-sm">
		<h2 class="mb-4 text-sm font-semibold uppercase tracking-wide text-gray-500">
			{$_('students.title')}
		</h2>
		<form on:submit|preventDefault={saveInfo} class="space-y-3 max-w-lg">
			<div>
				<label class="mb-1 block text-sm font-medium text-gray-700">{$_('students.name')} *</label>
				<input
					type="text"
					bind:value={student.name}
					required
					data-testid="student-name"
					class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
				/>
			</div>
			<div>
				<label class="mb-1 block text-sm font-medium text-gray-700">{$_('students.email')}</label>
				<input
					type="email"
					bind:value={student.email}
					data-testid="student-email"
					class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
				/>
			</div>
			<div>
				<label class="mb-1 block text-sm font-medium text-gray-700">{$_('students.notes')}</label>
				<textarea
					bind:value={student.notes}
					rows="2"
					data-testid="student-notes"
					class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
				></textarea>
			</div>
			<div class="flex items-center gap-3">
				<button
					type="submit"
					disabled={saving}
					data-testid="student-save-info"
					class="rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50"
				>
					{$_('students.save')}
				</button>
				{#if savedInfo}
					<span class="text-sm text-green-600" data-testid="saved-indicator">{$_('settings.saved')}</span>
				{/if}
			</div>
		</form>
	</section>

	<!-- Availability editor -->
	<section class="rounded-xl border border-gray-200 bg-white p-6 shadow-sm">
		<h2 class="mb-4 text-sm font-semibold uppercase tracking-wide text-gray-500">
			{$_('availability.title')}
		</h2>
		<AvailabilityEditor bind:slots={availability} />
		<div class="mt-4 flex items-center gap-3">
			<button
				on:click={saveAvailability}
				disabled={saving}
				data-testid="student-save-availability"
				class="rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50"
			>
				{$_('availability.save')}
			</button>
			{#if savedAvail}
				<span class="text-sm text-green-600" data-testid="saved-indicator">{$_('settings.saved')}</span>
			{/if}
		</div>
	</section>
{/if}
