<script lang="ts">
	import { _ } from 'svelte-i18n';
	import type { AvailabilitySlot } from '$lib/types';

	export let slots: AvailabilitySlot[] = [];

	const DAYS = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
	const START_HOUR = 7;
	const END_HOUR = 22;
	// One cell = 30 minutes for readability in the editor
	const CELL_MINS = 30;
	const CELLS_PER_HOUR = 60 / CELL_MINS;
	const TOTAL_CELLS = (END_HOUR - START_HOUR) * CELLS_PER_HOUR;

	// Build a 2D boolean grid [day][cellIndex] from slots
	let grid: boolean[][] = Array.from({ length: 7 }, () => new Array(TOTAL_CELLS).fill(false));

	$: {
		grid = Array.from({ length: 7 }, () => new Array(TOTAL_CELLS).fill(false));
		for (const slot of slots) {
			const startCell = timeToCell(slot.start_time);
			const endCell = timeToCell(slot.end_time);
			for (let c = startCell; c < endCell; c++) {
				if (c >= 0 && c < TOTAL_CELLS) {
					grid[slot.day_of_week][c] = true;
				}
			}
		}
	}

	function timeToCell(t: string): number {
		const [h, m] = t.split(':').map(Number);
		return ((h - START_HOUR) * 60 + m) / CELL_MINS;
	}

	function cellToTime(cell: number): string {
		const totalMins = START_HOUR * 60 + cell * CELL_MINS;
		return `${String(Math.floor(totalMins / 60)).padStart(2, '0')}:${String(totalMins % 60).padStart(2, '0')}`;
	}

	function gridToSlots(): AvailabilitySlot[] {
		const result: AvailabilitySlot[] = [];
		for (let day = 0; day < 7; day++) {
			let inRun = false;
			let runStart = 0;
			for (let c = 0; c <= TOTAL_CELLS; c++) {
				const active = c < TOTAL_CELLS && grid[day][c];
				if (active && !inRun) {
					inRun = true;
					runStart = c;
				} else if (!active && inRun) {
					result.push({
						day_of_week: day,
						start_time: cellToTime(runStart),
						end_time: cellToTime(c)
					});
					inRun = false;
				}
			}
		}
		return result;
	}

	// Drag-to-paint support
	let painting = false;
	let paintValue = false; // true = marking available, false = clearing

	function cellKey(day: number, cell: number) {
		return `${day}-${cell}`;
	}

	function startPaint(day: number, cell: number) {
		painting = true;
		paintValue = !grid[day][cell];
		applyPaint(day, cell);
	}

	function applyPaint(day: number, cell: number) {
		if (!painting) return;
		grid[day][cell] = paintValue;
		grid = grid; // trigger reactivity
		slots = gridToSlots();
	}

	function stopPaint() {
		painting = false;
	}

	function cellLabel(cell: number): string {
		if (cell % CELLS_PER_HOUR === 0) {
			return cellToTime(cell).slice(0, 5);
		}
		return '';
	}
</script>

<!-- svelte-ignore a11y-no-static-element-interactions -->
<div
	class="select-none overflow-x-auto"
	on:mouseup={stopPaint}
	on:mouseleave={stopPaint}
>
	<p class="mb-2 text-xs text-gray-400">{$_('availability.hint')}</p>

	<div class="flex min-w-[560px]">
		<!-- Time labels -->
		<div class="w-10 shrink-0 pt-6">
			{#each Array(TOTAL_CELLS) as _, c}
				<div class="h-5 flex items-center justify-end pr-1">
					<span class="text-[10px] text-gray-400">{cellLabel(c)}</span>
				</div>
			{/each}
		</div>

		<!-- Day columns -->
		{#each DAYS as day, d}
			<div class="flex-1 min-w-14">
				<div class="h-6 flex items-center justify-center text-xs font-medium text-gray-500 mb-0">
					{day}
				</div>
				{#each Array(TOTAL_CELLS) as _, c}
					<!-- svelte-ignore a11y-no-static-element-interactions -->
					<div
						class="h-5 mx-0.5 cursor-pointer rounded-sm border border-transparent transition-colors"
						class:bg-indigo-400={grid[d][c]}
						class:hover:bg-indigo-300={grid[d][c]}
						class:bg-gray-100={!grid[d][c]}
						class:hover:bg-indigo-100={!grid[d][c]}
						class:border-t-gray-200={c % CELLS_PER_HOUR === 0 && !grid[d][c]}
						on:mousedown={() => startPaint(d, c)}
						on:mouseenter={() => applyPaint(d, c)}
					></div>
				{/each}
			</div>
		{/each}
	</div>
</div>
