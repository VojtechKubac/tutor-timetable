<script lang="ts">
	import { onMount } from 'svelte';
	import { _ } from 'svelte-i18n';
	import { api } from '$lib/api';
	import type { Lesson } from '$lib/types';

	const DAYS = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday'];
	const START_HOUR = 8;
	const END_HOUR = 20;
	// Pixels per 5-minute slot
	const SLOT_PX = 12;
	const SLOT_MINS = 5;
	const TOTAL_SLOTS = ((END_HOUR - START_HOUR) * 60) / SLOT_MINS;

	let lessons: Lesson[] = [];
	let loading = false;
	let generating = false;
	let error = '';

	onMount(load);

	async function load() {
		loading = true;
		try {
			lessons = await api.timetable.get();
		} catch {
			error = $_('common.error');
		} finally {
			loading = false;
		}
	}

	async function generate() {
		generating = true;
		error = '';
		try {
			lessons = await api.timetable.generate();
		} catch {
			error = $_('common.error');
		} finally {
			generating = false;
		}
	}

	async function togglePin(lesson: Lesson) {
		try {
			await api.timetable.togglePin(lesson.id);
			lessons = lessons.map((l) =>
				l.id === lesson.id ? { ...l, is_pinned: !l.is_pinned } : l
			);
		} catch {
			error = $_('common.error');
		}
	}

	function lessonsForDay(day: number) {
		return lessons.filter((l) => l.day_of_week === day);
	}

	function topPx(startTime: string): number {
		const [h, m] = startTime.split(':').map(Number);
		return (((h - START_HOUR) * 60 + m) / SLOT_MINS) * SLOT_PX;
	}

	function heightPx(startTime: string, endTime: string): number {
		const [sh, sm] = startTime.split(':').map(Number);
		const [eh, em] = endTime.split(':').map(Number);
		return (((eh - sh) * 60 + (em - sm)) / SLOT_MINS) * SLOT_PX;
	}

	const gridHeight = TOTAL_SLOTS * SLOT_PX;
</script>

<div class="flex items-center justify-between mb-6">
	<h1 class="text-2xl font-semibold text-gray-900">{$_('timetable.title')}</h1>
	<button
		on:click={generate}
		disabled={generating}
		class="rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50"
	>
		{generating ? $_('timetable.generating') : $_('timetable.generate')}
	</button>
</div>

{#if error}
	<p class="mb-4 text-sm text-red-600">{error}</p>
{/if}

{#if loading}
	<p class="text-gray-500">{$_('common.loading')}</p>
{:else if lessons.length === 0}
	<p class="text-gray-400">{$_('timetable.empty')}</p>
{:else}
	<!-- Timetable grid -->
	<div class="overflow-x-auto rounded-xl border border-gray-200 bg-white">
		<div class="flex min-w-[700px]">
			<!-- Time axis -->
			<div class="w-14 shrink-0 border-r border-gray-100">
				<div class="h-8 border-b border-gray-100"></div>
				<div style="height:{gridHeight}px; position:relative;">
					{#each Array(END_HOUR - START_HOUR) as _, i}
						<div
							class="absolute left-0 right-0 border-t border-gray-100 text-right pr-2 text-xs text-gray-400"
							style="top:{(i * 60 / SLOT_MINS) * SLOT_PX}px; line-height:1;"
						>
							{String(START_HOUR + i).padStart(2, '0')}:00
						</div>
					{/each}
				</div>
			</div>

			<!-- Day columns -->
			{#each DAYS as day, d}
				<div class="flex-1 min-w-24 border-r border-gray-100 last:border-r-0">
					<div class="h-8 border-b border-gray-100 flex items-center justify-center text-xs font-medium text-gray-500">
						{day.slice(0, 3)}
					</div>
					<div style="height:{gridHeight}px; position:relative;">
						<!-- Hour grid lines -->
						{#each Array(END_HOUR - START_HOUR) as _, i}
							<div
								class="absolute left-0 right-0 border-t border-gray-100"
								style="top:{(i * 60 / SLOT_MINS) * SLOT_PX}px;"
							></div>
						{/each}

						<!-- Lessons -->
						{#each lessonsForDay(d) as lesson (lesson.id)}
							<div
								class="absolute left-1 right-1 rounded px-1.5 py-0.5 text-xs cursor-pointer select-none transition-shadow hover:shadow-md"
								class:bg-indigo-100={!lesson.is_pinned}
								class:text-indigo-800={!lesson.is_pinned}
								class:border={!lesson.is_pinned}
								class:border-indigo-200={!lesson.is_pinned}
								class:bg-amber-100={lesson.is_pinned}
								class:text-amber-800={lesson.is_pinned}
								class:border-amber-200={lesson.is_pinned}
								style="top:{topPx(lesson.start_time)}px; height:{heightPx(lesson.start_time, lesson.end_time)}px;"
								title="{lesson.student_name} {lesson.start_time}–{lesson.end_time}{lesson.is_pinned ? ' (pinned)' : ''}"
							>
								<div class="font-medium truncate">{lesson.student_name}</div>
								<div class="text-[10px] opacity-70">{lesson.start_time}–{lesson.end_time}</div>
								<button
									on:click|stopPropagation={() => togglePin(lesson)}
									class="absolute top-0.5 right-0.5 text-[10px] opacity-60 hover:opacity-100"
									title={lesson.is_pinned ? $_('timetable.unpin') : $_('timetable.pin')}
								>
									{lesson.is_pinned ? '📌' : '·'}
								</button>
							</div>
						{/each}
					</div>
				</div>
			{/each}
		</div>
	</div>
{/if}
