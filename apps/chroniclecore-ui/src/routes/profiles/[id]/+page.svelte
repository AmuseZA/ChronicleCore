<script lang="ts">
    import { onMount } from "svelte";
    import { page } from "$app/stores";
    import { goto } from "$app/navigation";
    import { fetchApi } from "$lib/api";
    import { settings } from "$lib/stores/settings";
    import { format, parseISO, startOfMonth, endOfMonth, subMonths, addMonths } from "date-fns";
    import ProfileSelector from "$lib/components/ProfileSelector.svelte";

    interface ProfileStats {
        profile_id: number;
        client_name: string;
        project_name?: string;
        service_name: string;
        rate_name: string;
        rate_amount: number;
        currency_code: string;
        total_blocks: number;
        total_minutes: number;
        total_hours: number;
        billable_minutes: number;
        billable_hours: number;
        estimated_billable: number;
        locked_minutes: number;
        locked_hours: number;
        locked_billable: number;
        recent_blocks?: Block[];
    }

    interface Block {
        block_id: number;
        ts_start: string;
        ts_end: string;
        duration_minutes: number;
        primary_app_name: string;
        title_summary?: string;
        confidence: string;
        billable: boolean;
        locked: boolean;
    }

    interface GroupedByDate {
        date: string;
        dateFormatted: string;
        blocks: Block[];
        totalMinutes: number;
    }

    let stats: ProfileStats | null = null;
    let loading = true;
    let error: string | null = null;
    let processingBlockId: number | null = null;
    let expandedDates: Set<string> = new Set();

    // View mode: 'lifetime' | 'monthly'
    let viewMode: 'lifetime' | 'monthly' = 'monthly';

    // Month selector
    let selectedMonth: Date = new Date();

    $: profileId = $page.params.id;

    // Group blocks by date
    $: groupedBlocks = groupBlocksByDate(stats?.recent_blocks || []);

    function groupBlocksByDate(blocks: Block[]): GroupedByDate[] {
        const groups = new Map<string, GroupedByDate>();

        for (const block of blocks) {
            const date = format(parseISO(block.ts_start), "yyyy-MM-dd");

            if (!groups.has(date)) {
                groups.set(date, {
                    date,
                    dateFormatted: format(parseISO(block.ts_start), "EEEE, MMMM d, yyyy"),
                    blocks: [],
                    totalMinutes: 0
                });
            }

            const group = groups.get(date)!;
            group.blocks.push(block);
            group.totalMinutes += block.duration_minutes;
        }

        // Sort by date descending
        return Array.from(groups.values()).sort((a, b) => b.date.localeCompare(a.date));
    }

    async function loadStats() {
        loading = true;
        error = null;
        try {
            await settings.detectLocale();

            let url = `/profiles/${profileId}/stats?include_blocks=true`;

            if (viewMode === 'monthly') {
                const start = format(startOfMonth(selectedMonth), 'yyyy-MM-dd');
                const end = format(endOfMonth(selectedMonth), 'yyyy-MM-dd');
                url += `&start_date=${start}&end_date=${end}`;
            }

            stats = await fetchApi(url);

            // Auto-expand the first date group
            if (stats?.recent_blocks && stats.recent_blocks.length > 0) {
                const firstDate = format(parseISO(stats.recent_blocks[0].ts_start), "yyyy-MM-dd");
                expandedDates.add(firstDate);
                expandedDates = expandedDates;
            }
        } catch (e: any) {
            error = e.message || "Failed to load profile stats";
        } finally {
            loading = false;
        }
    }

    function formatCurrency(amount: number, code: string) {
        try {
            return new Intl.NumberFormat($settings.locale, {
                style: "currency",
                currency: code,
            }).format(amount);
        } catch (e) {
            return `${code} ${amount.toFixed(2)}`;
        }
    }

    function prevMonth() {
        selectedMonth = subMonths(selectedMonth, 1);
        loadStats();
    }

    function nextMonth() {
        selectedMonth = addMonths(selectedMonth, 1);
        loadStats();
    }

    function goToCurrentMonth() {
        selectedMonth = new Date();
        loadStats();
    }

    function setViewMode(mode: 'lifetime' | 'monthly') {
        viewMode = mode;
        loadStats();
    }

    function toggleDateGroup(date: string) {
        if (expandedDates.has(date)) {
            expandedDates.delete(date);
        } else {
            expandedDates.add(date);
        }
        expandedDates = expandedDates;
    }

    function getConfidenceStyle(confidence: string) {
        switch (confidence) {
            case "HIGH":
                return "bg-green-50 text-green-700 border-green-200";
            case "MEDIUM":
                return "bg-amber-50 text-amber-700 border-amber-200";
            case "LOW":
                return "bg-red-50 text-red-700 border-red-200";
            default:
                return "bg-slate-50 text-slate-700 border-slate-200";
        }
    }

    async function reassignBlock(blockId: number, newProfileId: number | null) {
        if (newProfileId === null) return;

        processingBlockId = blockId;
        try {
            await fetchApi(`/blocks/${blockId}/reassign`, {
                method: "POST",
                body: JSON.stringify({ profile_id: newProfileId }),
            });
            // Reload stats to reflect the change (block will be removed from this profile)
            await loadStats();
        } catch (e) {
            console.error(e);
            alert("Failed to reassign block");
        } finally {
            processingBlockId = null;
        }
    }

    async function unassignBlock(blockId: number) {
        if (!confirm("Are you sure you want to unassign this block? It will appear in the Review queue.")) {
            return;
        }

        processingBlockId = blockId;
        try {
            await fetchApi(`/blocks/${blockId}/reassign`, {
                method: "POST",
                body: JSON.stringify({ profile_id: null }),
            });
            await loadStats();
        } catch (e) {
            console.error(e);
            alert("Failed to unassign block");
        } finally {
            processingBlockId = null;
        }
    }

    onMount(loadStats);
</script>

<div class="max-w-7xl mx-auto space-y-6">
    <!-- Header with Back Button -->
    <div class="flex items-center gap-4">
        <button
            on:click={() => goto('/profiles')}
            class="flex items-center gap-2 text-slate-500 hover:text-slate-700 transition-colors"
        >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
            </svg>
            Back to Profiles
        </button>
    </div>

    {#if loading}
        <div class="bg-white rounded-xl border border-slate-200 p-12 text-center text-slate-500">
            Loading profile details...
        </div>
    {:else if error}
        <div class="bg-red-50 text-red-700 p-4 rounded-xl border border-red-100">
            {error}
        </div>
    {:else if stats}
        <!-- Profile Header Card -->
        <div class="bg-white rounded-xl border border-slate-200 shadow-sm p-6">
            <div class="flex items-start justify-between">
                <div class="flex items-center gap-4">
                    <div class="w-14 h-14 rounded-full bg-blue-100 text-blue-700 flex items-center justify-center font-bold text-lg">
                        {stats.client_name.slice(0, 2).toUpperCase()}
                    </div>
                    <div>
                        <h1 class="text-2xl font-bold text-slate-900">{stats.client_name}</h1>
                        {#if stats.project_name}
                            <div class="text-slate-500">{stats.project_name}</div>
                        {/if}
                        <div class="flex items-center gap-3 mt-1">
                            <span class="inline-flex px-2.5 py-1 rounded-md bg-slate-100 text-slate-600 text-xs font-medium border border-slate-200">
                                {stats.service_name}
                            </span>
                            <span class="text-sm text-slate-500">
                                {stats.rate_name} - {formatCurrency(stats.rate_amount, stats.currency_code)}/hr
                            </span>
                        </div>
                    </div>
                </div>

                <!-- View Mode Toggle -->
                <div class="flex items-center gap-1 bg-slate-100 rounded-lg p-1">
                    <button
                        on:click={() => setViewMode('monthly')}
                        class="px-4 py-2 text-sm font-medium rounded-md transition-colors {viewMode === 'monthly'
                            ? 'bg-white text-slate-900 shadow-sm'
                            : 'text-slate-600 hover:text-slate-900'}"
                    >
                        Monthly
                    </button>
                    <button
                        on:click={() => setViewMode('lifetime')}
                        class="px-4 py-2 text-sm font-medium rounded-md transition-colors {viewMode === 'lifetime'
                            ? 'bg-white text-slate-900 shadow-sm'
                            : 'text-slate-600 hover:text-slate-900'}"
                    >
                        Lifetime
                    </button>
                </div>
            </div>

            <!-- Month Selector (only for monthly view) -->
            {#if viewMode === 'monthly'}
                <div class="mt-6 pt-6 border-t border-slate-100 flex items-center justify-between">
                    <div class="flex items-center gap-3">
                        <button
                            on:click={prevMonth}
                            class="w-9 h-9 rounded-lg border border-slate-200 bg-white hover:bg-slate-50 flex items-center justify-center transition-colors"
                        >
                            <svg class="w-5 h-5 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
                            </svg>
                        </button>
                        <div class="text-lg font-semibold text-slate-900 min-w-[180px] text-center">
                            {format(selectedMonth, 'MMMM yyyy')}
                        </div>
                        <button
                            on:click={nextMonth}
                            class="w-9 h-9 rounded-lg border border-slate-200 bg-white hover:bg-slate-50 flex items-center justify-center transition-colors"
                        >
                            <svg class="w-5 h-5 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                            </svg>
                        </button>
                    </div>
                    <button
                        on:click={goToCurrentMonth}
                        class="text-sm text-blue-600 hover:text-blue-700 font-medium"
                    >
                        Go to Current Month
                    </button>
                </div>
            {/if}
        </div>

        <!-- Stats Cards -->
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <!-- Total Time -->
            <div class="bg-white rounded-xl border border-slate-200 shadow-sm p-5">
                <div class="text-xs font-medium text-slate-500 uppercase tracking-wide">Total Time</div>
                <div class="mt-2 text-2xl font-bold text-slate-900 tabular-nums">
                    {stats.total_hours.toFixed(1)}h
                </div>
                <div class="text-sm text-slate-500 mt-1">
                    {stats.total_blocks} blocks
                </div>
            </div>

            <!-- Billable Time -->
            <div class="bg-white rounded-xl border border-slate-200 shadow-sm p-5">
                <div class="text-xs font-medium text-slate-500 uppercase tracking-wide">Billable Time</div>
                <div class="mt-2 text-2xl font-bold text-slate-900 tabular-nums">
                    {stats.billable_hours.toFixed(1)}h
                </div>
                <div class="text-sm text-slate-500 mt-1">
                    {((stats.billable_hours / (stats.total_hours || 1)) * 100).toFixed(0)}% of total
                </div>
            </div>

            <!-- Estimated Billable -->
            <div class="bg-white rounded-xl border border-slate-200 shadow-sm p-5">
                <div class="text-xs font-medium text-slate-500 uppercase tracking-wide">Est. Billable</div>
                <div class="mt-2 text-2xl font-bold text-green-600 tabular-nums">
                    {formatCurrency(stats.estimated_billable, stats.currency_code)}
                </div>
                <div class="text-sm text-slate-500 mt-1">
                    @ {formatCurrency(stats.rate_amount, stats.currency_code)}/hr
                </div>
            </div>

            <!-- Locked / Invoiced -->
            <div class="bg-white rounded-xl border border-slate-200 shadow-sm p-5">
                <div class="text-xs font-medium text-slate-500 uppercase tracking-wide">Locked (Ready)</div>
                <div class="mt-2 text-2xl font-bold text-slate-900 tabular-nums">
                    {stats.locked_hours.toFixed(1)}h
                </div>
                <div class="text-sm text-green-600 mt-1">
                    {formatCurrency(stats.locked_billable, stats.currency_code)}
                </div>
            </div>
        </div>

        <!-- Activity Grouped by Date -->
        <div class="space-y-4">
            <div class="flex items-center justify-between">
                <h2 class="text-lg font-semibold text-slate-900">
                    {viewMode === 'lifetime' ? 'Recent Activity' : `Activity for ${format(selectedMonth, 'MMMM yyyy')}`}
                </h2>
                <span class="text-sm text-slate-500">
                    {groupedBlocks.length} day{groupedBlocks.length !== 1 ? 's' : ''} with activity
                </span>
            </div>

            {#if groupedBlocks.length > 0}
                {#each groupedBlocks as dateGroup (dateGroup.date)}
                    <div class="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
                        <!-- Date Header -->
                        <button
                            on:click={() => toggleDateGroup(dateGroup.date)}
                            class="w-full px-6 py-4 flex items-center justify-between hover:bg-slate-50 transition-colors"
                        >
                            <div class="flex items-center gap-4">
                                <div class="flex items-center gap-2">
                                    <svg class="w-5 h-5 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                                    </svg>
                                    <span class="font-semibold text-slate-900">{dateGroup.dateFormatted}</span>
                                </div>
                                <span class="px-2 py-0.5 rounded-full bg-slate-100 text-slate-600 text-xs font-medium">
                                    {dateGroup.blocks.length} block{dateGroup.blocks.length !== 1 ? 's' : ''}
                                </span>
                            </div>
                            <div class="flex items-center gap-4">
                                <span class="font-medium text-slate-700 tabular-nums">
                                    {(dateGroup.totalMinutes / 60).toFixed(1)}h
                                </span>
                                <svg
                                    class="w-5 h-5 text-slate-400 transition-transform {expandedDates.has(dateGroup.date) ? 'rotate-180' : ''}"
                                    fill="none" stroke="currentColor" viewBox="0 0 24 24"
                                >
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                                </svg>
                            </div>
                        </button>

                        <!-- Blocks List (Expanded) -->
                        {#if expandedDates.has(dateGroup.date)}
                            <div class="border-t border-slate-100">
                                <div class="divide-y divide-slate-100">
                                    {#each dateGroup.blocks as block (block.block_id)}
                                        <div class="px-6 py-4 hover:bg-slate-50 transition-colors flex items-start gap-4">
                                            <!-- Time Column -->
                                            <div class="w-24 shrink-0">
                                                <div class="text-sm font-medium text-slate-900">
                                                    {format(parseISO(block.ts_start), "HH:mm")}
                                                </div>
                                                <div class="text-xs text-slate-500">
                                                    to {format(parseISO(block.ts_end), "HH:mm")}
                                                </div>
                                            </div>

                                            <!-- Activity Info -->
                                            <div class="flex-1 min-w-0">
                                                <div class="text-sm text-slate-900 break-words">
                                                    {block.title_summary || block.primary_app_name}
                                                </div>
                                                <div class="flex items-center gap-2 mt-1">
                                                    <span class="text-xs text-slate-500">{block.primary_app_name}</span>
                                                    <span class="text-slate-300">•</span>
                                                    <span class="text-xs font-medium text-slate-600 tabular-nums">{block.duration_minutes.toFixed(0)} min</span>
                                                </div>
                                            </div>

                                            <!-- Status Badges -->
                                            <div class="flex items-center gap-2 shrink-0">
                                                <span class="text-xs px-2 py-1 rounded-md border {getConfidenceStyle(block.confidence)}">
                                                    {block.confidence}
                                                </span>
                                                {#if block.locked}
                                                    <span class="text-xs px-2 py-1 rounded-md bg-green-50 text-green-700 border border-green-200">
                                                        Locked
                                                    </span>
                                                {/if}
                                                {#if !block.billable}
                                                    <span class="text-xs px-2 py-1 rounded-md bg-slate-50 text-slate-600 border border-slate-200">
                                                        Non-billable
                                                    </span>
                                                {/if}
                                            </div>

                                            <!-- Actions -->
                                            <div class="flex items-center gap-2 shrink-0">
                                                {#if !block.locked}
                                                    <div class="w-36">
                                                        <ProfileSelector
                                                            placeholder="Change..."
                                                            value={stats?.profile_id}
                                                            disabled={processingBlockId === block.block_id}
                                                            on:change={(e) => reassignBlock(block.block_id, e.detail)}
                                                        />
                                                    </div>
                                                    <button
                                                        on:click={() => unassignBlock(block.block_id)}
                                                        disabled={processingBlockId === block.block_id}
                                                        class="text-xs text-slate-400 hover:text-red-600 transition-colors disabled:opacity-50"
                                                        title="Remove from this profile"
                                                    >
                                                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                                                        </svg>
                                                    </button>
                                                {:else}
                                                    <span class="text-xs text-slate-400 italic">Locked</span>
                                                {/if}
                                            </div>
                                        </div>
                                    {/each}
                                </div>
                            </div>
                        {/if}
                    </div>
                {/each}
            {:else}
                <div class="bg-white rounded-xl border border-slate-200 p-12 text-center">
                    <div class="w-12 h-12 bg-slate-100 text-slate-400 rounded-full flex items-center justify-center mx-auto mb-3">
                        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                    </div>
                    <h3 class="text-lg font-medium text-slate-900">No activity recorded</h3>
                    <p class="text-slate-500">
                        {#if viewMode === 'monthly'}
                            No time blocks for {format(selectedMonth, 'MMMM yyyy')}.
                        {:else}
                            No time blocks have been assigned to this profile yet.
                        {/if}
                    </p>
                </div>
            {/if}
        </div>
    {/if}
</div>
