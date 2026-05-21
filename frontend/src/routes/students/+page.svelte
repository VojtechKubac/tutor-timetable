<script lang="ts">
	import { onMount } from 'svelte';
	import { _ } from 'svelte-i18n';
	import { api } from '$lib/api';
	import type { Student } from '$lib/types';

	let students: Student[] = [];
	let loading = false;
	let error = '';

	// Add form state
	let showAdd = false;
	let newName = '';
	let newEmail = '';
	let newNotes = '';
	let saving = false;

	onMount(load);

	async function load() {
		loading = true;
		try {
			students = await api.students.list();
		} catch {
			error = $_('common.error');
		} finally {
			loading = false;
		}
	}

	async function addStudent() {
		if (!newName.trim()) return;
		saving = true;
		try {
			const s = await api.students.create({ name: newName.trim(), email: newEmail.trim(), notes: newNotes.trim() });
			students = [...students, s].sort((a, b) => a.name.localeCompare(b.name));
			newName = '';
			newEmail = '';
			newNotes = '';
			showAdd = false;
		} catch {
			error = $_('common.error');
		} finally {
			saving = false;
		}
	}

	async function removeStudent(student: Student) {
		const msg = $_('students.confirmDelete', { values: { name: student.name } });
		if (!confirm(msg)) return;
		try {
			await api.students.remove(student.id);
			students = students.filter((s) => s.id !== student.id);
		} catch {
			error = $_('common.error');
		}
	}
</script>

<div class="flex items-center justify-between mb-6">
	<h1 class="text-2xl font-semibold text-gray-900">{$_('students.title')}</h1>
	<button
		on:click={() => (showAdd = !showAdd)}
		class="rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700"
	>
		{$_('students.add')}
	</button>
</div>

{#if error}
	<p class="mb-4 text-sm text-red-600">{error}</p>
{/if}

<!-- Add student form -->
{#if showAdd}
	<form
		on:submit|preventDefault={addStudent}
		class="mb-6 rounded-xl border border-gray-200 bg-white p-5 shadow-sm space-y-3"
	>
		<div class="grid grid-cols-2 gap-3">
			<div>
				<label class="mb-1 block text-sm font-medium text-gray-700">{$_('students.name')} *</label>
				<input
					type="text"
					bind:value={newName}
					required
					class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
				/>
			</div>
			<div>
				<label class="mb-1 block text-sm font-medium text-gray-700">{$_('students.email')}</label>
				<input
					type="email"
					bind:value={newEmail}
					class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
				/>
			</div>
		</div>
		<div>
			<label class="mb-1 block text-sm font-medium text-gray-700">{$_('students.notes')}</label>
			<textarea
				bind:value={newNotes}
				rows="2"
				class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
			></textarea>
		</div>
		<div class="flex gap-2">
			<button
				type="submit"
				disabled={saving}
				class="rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50"
			>
				{saving ? $_('common.loading') : $_('students.save')}
			</button>
			<button
				type="button"
				on:click={() => (showAdd = false)}
				class="rounded-md border border-gray-300 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50"
			>
				{$_('students.cancel')}
			</button>
		</div>
	</form>
{/if}

{#if loading}
	<p class="text-gray-500">{$_('common.loading')}</p>
{:else if students.length === 0}
	<p class="text-gray-400">{$_('students.empty')}</p>
{:else}
	<div class="rounded-xl border border-gray-200 bg-white shadow-sm overflow-hidden">
		<table class="w-full text-sm">
			<tbody class="divide-y divide-gray-100">
				{#each students as student (student.id)}
					<tr class="hover:bg-gray-50">
						<td class="px-4 py-3 font-medium text-gray-900">
							<a href="/students/{student.id}" class="hover:text-indigo-600">
								{student.name}
							</a>
						</td>
						<td class="px-4 py-3 text-gray-500">{student.email || '—'}</td>
						<td class="px-4 py-3 text-gray-400 text-xs">{student.notes || ''}</td>
						<td class="px-4 py-3 text-right">
							<a
								href="/students/{student.id}"
								class="mr-3 text-indigo-600 hover:underline"
							>
								Edit
							</a>
							<button
								on:click={() => removeStudent(student)}
								class="text-red-500 hover:underline"
							>
								{$_('students.delete')}
							</button>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{/if}
