<script lang="ts">
    import { onMount } from "svelte";
    import { fetchApi } from "$lib/api";
    import { format, parseISO } from "date-fns";
    import ProfileSelector from "$lib/components/ProfileSelector.svelte";

    interface Block {
        block_id: number;
        app_id: number;
        ts_start: string;
        ts_end: string;
        duration_minutes: number;
        primary_app_name: string;
        title_summary?: string;
        profile_id?: number;
        client_name?: string;
        confidence: string;
        billable: boolean;
        activity_score?: number;
        notes?: string;
    }

    interface GroupedBlock {
        group_key: string;
        primary_app_name: string;
        app_id: number;
        title_context: string;
        total_minutes: number;
        total_hours: number;
        block_count: number;
        first_ts: string;
        last_ts: string;
        blocks: Block[];
    }

    interface Pagination {
        page: number;
        per_page: number;
        total: number;
        total_pages: number;
    }

    interface MLStatus {
        sidecar_running: boolean;
        training_samples: number;
        ready_for_training: boolean;
        has_trained_model: boolean;
        pending_suggestions: number;
    }

    let groups: GroupedBlock[] = [];
    let pagination: Pagination = { page: 1, per_page: 50, total: 0, total_pages: 0 };
    let loading = true;
    let processingId: number | null = null;
    let expandedGroups: Set<string> = new Set();
    let blacklisting = false;

    // ML Status
    let mlStatus: MLStatus | null = null;
    let isTraining = false;
    let trainingMessage = "";

    // Separate groups by confidence
    $: unassignedGroups = groups.filter(g => g.blocks.every(b => !b.profile_id));
    $: lowConfidenceGroups = groups.filter(g => g.blocks.some(b => b.profile_id && b.confidence === 'LOW'));
    $: mediumConfidenceGroups = groups.filter(g => g.blocks.some(b => b.profile_id && b.confidence === 'MEDIUM') && !g.blocks.some(b => b.confidence === 'LOW'));

    async function loadData(page = 1) {
        loading = true;
        try {
            const response = await fetchApi(`/blocks/grouped?needs_review=true&page=${page}&per_page=50`);
            groups = response?.data || [];
            pagination = response?.pagination || { page: 1, per_page: 50, total: 0, total_pages: 0 };
            await loadMLStatus();
        } catch (e) {
            console.error(e);
        } finally {
            loading = false;
        }
    }

    async function loadMLStatus() {
        try {
            mlStatus = await fetchApi('/ml/status');
        } catch (e) {
            console.warn("ML status unavailable:", e);
            mlStatus = null;
        }
    }

    async function triggerTraining() {
        if (!mlStatus?.ready_for_training) return;
        isTraining = true;
        trainingMessage = "";
        try {
            const result = await fetchApi('/ml/train', { method: 'POST' });
            trainingMessage = `Training complete! Accuracy: ${(result.metrics?.accuracy * 100 || 0).toFixed(1)}%`;
            await loadMLStatus();
            // After training, generate predictions
            await fetchApi('/ml/predict', { method: 'POST' });
            await loadData(pagination.page);
        } catch (e: any) {
            trainingMessage = `Training failed: ${e.message || 'Unknown error'}`;
        } finally {
            isTraining = false;
        }
    }

    function toggleGroup(groupKey: string) {
        if (expandedGroups.has(groupKey)) {
            expandedGroups.delete(groupKey);
        } else {
            expandedGroups.add(groupKey);
        }
        expandedGroups = expandedGroups;
    }

    async function assignBlock(blockId: number, profileId: number) {
        if (!profileId) return;
        processingId = blockId;
        try {
            await fetchApi(`/blocks/${blockId}/reassign`, {
                method: "POST",
                body: JSON.stringify({ profile_id: profileId }),
            });
            await loadData(pagination.page);
        } catch (e) {
            console.error(e);
            alert("Failed to reassign block");
        } finally {
            processingId = null;
        }
    }

    async function assignAllInGroup(group: GroupedBlock, profileId: number) {
        if (!profileId) return;
        processingId = group.blocks[0]?.block_id || null;
        try {
            for (const block of group.blocks) {
                await fetchApi(`/blocks/${block.block_id}/reassign`, {
                    method: "POST",
                    body: JSON.stringify({ profile_id: profileId }),
                });
            }
            await loadData(pagination.page);
        } catch (e) {
            console.error(e);
            alert("Failed to reassign blocks");
        } finally {
            processingId = null;
        }
    }

    async function deleteBlock(blockId: number) {
        if (!confirm("Are you sure you want to delete this block?")) return;
        processingId = blockId;
        try {
            await fetchApi(`/blocks/${blockId}`, { method: "DELETE" });
            await loadData(pagination.page);
        } catch (e) {
            console.error(e);
            alert("Failed to delete block");
        } finally {
            processingId = null;
        }
    }

    async function blacklistApp(appName: string, appId: number) {
        if (!confirm(`Blacklist "${appName}" and remove all related time entries? This cannot be undone.`)) return;
        blacklisting = true;
        try {
            await fetchApi("/blacklist/with-delete", {
                method: "POST",
                body: JSON.stringify({ app_id: appId, reason: "Blacklisted from review page" }),
            });
            await loadData(pagination.page);
        } catch (e) {
            console.error(e);
            alert("Failed to blacklist app");
        } finally {
            blacklisting = false;
        }
    }

    function goToPage(page: number) {
        if (page >= 1 && page <= pagination.total_pages) {
            loadData(page);
        }
    }

    // Compute pagination page numbers
    $: pageNumbers = (() => {
        const pages: number[] = [];
        const total = pagination.total_pages;
        const current = pagination.page;
        const maxVisible = 5;

        let start = Math.max(1, current - 2);
        let end = Math.min(total, start + maxVisible - 1);

        if (end - start + 1 < maxVisible) {
            start = Math.max(1, end - maxVisible + 1);
        }

        for (let i = start; i <= end; i++) {
            pages.push(i);
        }
        return pages;
    })();

    onMount(() => loadData());
</script>

<div class="max-w-7xl mx-auto space-y-6">
    <!-- Header -->
    <header class="flex justify-between items-start">
        <div>
            <h1 class="text-2xl font-bold text-slate-900">Needs Review</h1>
            <p class="text-slate-500">
                Activities grouped by context - assign profiles or blacklist apps.
            </p>
        </div>
        <div class="flex items-center gap-3">
            <!-- ML Status & Training -->
            {#if mlStatus}
                <div class="flex items-center gap-3 px-4 py-2 bg-slate-50 rounded-lg border border-slate-200">
                    <div class="text-xs text-slate-500">
                        <span class="font-medium">{mlStatus.training_samples}</span> training samples
                    </div>
                    {#if mlStatus.has_trained_model}
                        <div class="w-2 h-2 rounded-full bg-green-500" title="Model trained"></div>
                    {/if}
                    <button
                        on:click={triggerTraining}
                        disabled={!mlStatus.ready_for_training || isTraining}
                        class="px-3 py-1.5 text-xs font-medium rounded-md transition-colors
                               {mlStatus.ready_for_training && !isTraining
                                   ? 'bg-blue-600 text-white hover:bg-blue-700'
                                   : 'bg-slate-200 text-slate-400 cursor-not-allowed'}"
                    >
                        {isTraining ? 'Training...' : 'Train ML'}
                    </button>
                </div>
            {/if}
            <button
                on:click={() => loadData(pagination.page)}
                class="text-sm text-blue-600 hover:text-blue-800 font-medium"
            >
                Refresh
            </button>
        </div>
    </header>

    {#if trainingMessage}
        <div class="px-4 py-3 rounded-lg text-sm {trainingMessage.includes('failed') ? 'bg-red-50 text-red-700 border border-red-200' : 'bg-green-50 text-green-700 border border-green-200'}">
            {trainingMessage}
        </div>
    {/if}

    {#if loading}
        <div class="bg-white rounded-xl border border-slate-200 p-12 text-center text-slate-500">
            Loading...
        </div>
    {:else if groups.length === 0}
        <div class="bg-white rounded-xl border border-slate-200 p-12 text-center">
            <div class="w-12 h-12 bg-green-100 text-green-600 rounded-full flex items-center justify-center mx-auto mb-3">
                <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                </svg>
            </div>
            <h3 class="text-lg font-medium text-slate-900">All caught up!</h3>
            <p class="text-slate-500">No activities need review right now.</p>
        </div>
    {:else}
        <!-- Unassigned Section -->
        {#if unassignedGroups.length > 0}
            <section class="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
                <div class="px-6 py-4 border-b border-slate-100 flex items-center justify-between bg-amber-50/30">
                    <div class="flex items-center gap-3">
                        <span class="px-2.5 py-1 rounded-md text-xs font-semibold bg-amber-100 text-amber-800">
                            Unassigned
                        </span>
                        <span class="text-sm text-slate-600">
                            {unassignedGroups.length} groups need profile assignment
                        </span>
                    </div>
                </div>
                <div class="divide-y divide-slate-100">
                    {#each unassignedGroups as group (group.group_key)}
                        <div class="border-l-4 border-l-amber-400">
                            <div class="px-6 py-4 flex items-center justify-between hover:bg-slate-50/50 transition-colors">
                                <button
                                    class="flex items-center gap-3 text-left flex-1"
                                    on:click={() => toggleGroup(group.group_key)}
                                >
                                    <div class="w-8 h-8 rounded-lg bg-slate-100 text-slate-700 flex items-center justify-center font-bold text-xs">
                                        {group.block_count}
                                    </div>
                                    <div class="flex-1 min-w-0">
                                        <div class="font-semibold text-slate-900 truncate">
                                            {group.title_context}
                                        </div>
                                        <div class="text-xs text-slate-500 flex items-center gap-2 flex-wrap">
                                            <span class="inline-flex items-center px-1.5 py-0.5 rounded bg-slate-100 text-slate-600">
                                                {group.primary_app_name}
                                            </span>
                                            <span>{group.total_minutes.toFixed(0)} min</span>
                                            <span class="text-slate-300">|</span>
                                            <span>{format(parseISO(group.first_ts), "MMM d")} - {format(parseISO(group.last_ts), "MMM d, HH:mm")}</span>
                                        </div>
                                    </div>
                                    <svg
                                        class="w-5 h-5 text-slate-400 transition-transform {expandedGroups.has(group.group_key) ? 'rotate-180' : ''}"
                                        fill="none" stroke="currentColor" viewBox="0 0 24 24"
                                    >
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                                    </svg>
                                </button>

                                <div class="flex items-center gap-3 ml-4">
                                    <div class="w-48">
                                        <ProfileSelector
                                            placeholder="Assign all..."
                                            on:change={(e) => assignAllInGroup(group, e.detail)}
                                        />
                                    </div>
                                    <button
                                        on:click={() => blacklistApp(group.primary_app_name, group.app_id)}
                                        disabled={blacklisting}
                                        class="text-xs text-slate-400 hover:text-red-600 px-2 py-1 rounded hover:bg-red-50 transition-colors"
                                        title="Blacklist this app"
                                    >
                                        Blacklist
                                    </button>
                                </div>
                            </div>

                            {#if expandedGroups.has(group.group_key)}
                                <div class="border-t border-slate-100 bg-slate-50/50">
                                    <table class="w-full text-left">
                                        <thead class="border-b border-slate-200">
                                            <tr>
                                                <th class="px-6 py-2 text-xs font-medium text-slate-500 uppercase">Time</th>
                                                <th class="px-6 py-2 text-xs font-medium text-slate-500 uppercase">Details</th>
                                                <th class="px-6 py-2 text-xs font-medium text-slate-500 uppercase">Status</th>
                                                <th class="px-6 py-2 text-xs font-medium text-slate-500 uppercase w-48">Assign</th>
                                                <th class="px-6 py-2 text-xs font-medium text-slate-500 uppercase w-20"></th>
                                            </tr>
                                        </thead>
                                        <tbody class="divide-y divide-slate-100 bg-white">
                                            {#each group.blocks as block (block.block_id)}
                                                <tr class="hover:bg-slate-50 transition-colors">
                                                    <td class="px-6 py-3 text-sm text-slate-600 whitespace-nowrap">
                                                        <div>{format(parseISO(block.ts_start), "HH:mm")} - {format(parseISO(block.ts_end), "HH:mm")}</div>
                                                        <div class="text-xs text-slate-400">{block.duration_minutes.toFixed(0)} min</div>
                                                    </td>
                                                    <td class="px-6 py-3">
                                                        <div class="text-sm text-slate-700 truncate max-w-md">
                                                            {block.title_summary || block.primary_app_name}
                                                        </div>
                                                        {#if block.activity_score !== undefined}
                                                            <div class="flex items-center gap-2 mt-1">
                                                                <div class="w-16 h-1.5 bg-slate-200 rounded-full overflow-hidden" title="Activity: {(block.activity_score * 100).toFixed(0)}%">
                                                                    <div class="h-full bg-green-500 rounded-full" style="width: {block.activity_score * 100}%"></div>
                                                                </div>
                                                            </div>
                                                        {/if}
                                                    </td>
                                                    <td class="px-6 py-3">
                                                        <span class="text-xs px-2 py-1 rounded-md bg-amber-50 text-amber-700 border border-amber-200">
                                                            Unassigned
                                                        </span>
                                                    </td>
                                                    <td class="px-6 py-3">
                                                        <ProfileSelector
                                                            disabled={processingId === block.block_id}
                                                            on:change={(e) => assignBlock(block.block_id, e.detail)}
                                                        />
                                                    </td>
                                                    <td class="px-6 py-3 text-right">
                                                        {#if processingId === block.block_id}
                                                            <span class="text-xs text-slate-400">...</span>
                                                        {:else}
                                                            <button
                                                                on:click={() => deleteBlock(block.block_id)}
                                                                class="text-xs text-slate-400 hover:text-red-600"
                                                            >
                                                                Delete
                                                            </button>
                                                        {/if}
                                                    </td>
                                                </tr>
                                            {/each}
                                        </tbody>
                                    </table>
                                </div>
                            {/if}
                        </div>
                    {/each}
                </div>
            </section>
        {/if}

        <!-- Low Confidence Section -->
        {#if lowConfidenceGroups.length > 0}
            <section class="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
                <div class="px-6 py-4 border-b border-slate-100 flex items-center justify-between bg-red-50/30">
                    <div class="flex items-center gap-3">
                        <span class="px-2.5 py-1 rounded-md text-xs font-semibold bg-red-100 text-red-800">
                            Needs Verification
                        </span>
                        <span class="text-sm text-slate-600">
                            {lowConfidenceGroups.length} groups with low confidence assignments
                        </span>
                    </div>
                </div>
                <div class="divide-y divide-slate-100">
                    {#each lowConfidenceGroups as group (group.group_key)}
                        <div class="border-l-4 border-l-red-400">
                            <div class="px-6 py-4 flex items-center justify-between hover:bg-slate-50/50 transition-colors">
                                <button
                                    class="flex items-center gap-3 text-left flex-1"
                                    on:click={() => toggleGroup(group.group_key)}
                                >
                                    <div class="w-8 h-8 rounded-lg bg-slate-100 text-slate-700 flex items-center justify-center font-bold text-xs">
                                        {group.block_count}
                                    </div>
                                    <div class="flex-1 min-w-0">
                                        <div class="font-semibold text-slate-900 truncate">
                                            {group.title_context}
                                        </div>
                                        <div class="text-xs text-slate-500 flex items-center gap-2 flex-wrap">
                                            <span class="inline-flex items-center px-1.5 py-0.5 rounded bg-slate-100 text-slate-600">
                                                {group.primary_app_name}
                                            </span>
                                            <span>{group.total_minutes.toFixed(0)} min</span>
                                            <span class="text-slate-300">|</span>
                                            <span>{format(parseISO(group.first_ts), "MMM d")} - {format(parseISO(group.last_ts), "MMM d, HH:mm")}</span>
                                        </div>
                                    </div>
                                    <svg
                                        class="w-5 h-5 text-slate-400 transition-transform {expandedGroups.has(group.group_key) ? 'rotate-180' : ''}"
                                        fill="none" stroke="currentColor" viewBox="0 0 24 24"
                                    >
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                                    </svg>
                                </button>

                                <div class="flex items-center gap-3 ml-4">
                                    <div class="w-48">
                                        <ProfileSelector
                                            placeholder="Assign all..."
                                            on:change={(e) => assignAllInGroup(group, e.detail)}
                                        />
                                    </div>
                                    <button
                                        on:click={() => blacklistApp(group.primary_app_name, group.app_id)}
                                        disabled={blacklisting}
                                        class="text-xs text-slate-400 hover:text-red-600 px-2 py-1 rounded hover:bg-red-50 transition-colors"
                                        title="Blacklist this app"
                                    >
                                        Blacklist
                                    </button>
                                </div>
                            </div>

                            {#if expandedGroups.has(group.group_key)}
                                <div class="border-t border-slate-100 bg-slate-50/50">
                                    <table class="w-full text-left">
                                        <thead class="border-b border-slate-200">
                                            <tr>
                                                <th class="px-6 py-2 text-xs font-medium text-slate-500 uppercase">Time</th>
                                                <th class="px-6 py-2 text-xs font-medium text-slate-500 uppercase">Details</th>
                                                <th class="px-6 py-2 text-xs font-medium text-slate-500 uppercase">Status</th>
                                                <th class="px-6 py-2 text-xs font-medium text-slate-500 uppercase w-48">Assign</th>
                                                <th class="px-6 py-2 text-xs font-medium text-slate-500 uppercase w-20"></th>
                                            </tr>
                                        </thead>
                                        <tbody class="divide-y divide-slate-100 bg-white">
                                            {#each group.blocks as block (block.block_id)}
                                                <tr class="hover:bg-slate-50 transition-colors">
                                                    <td class="px-6 py-3 text-sm text-slate-600 whitespace-nowrap">
                                                        <div>{format(parseISO(block.ts_start), "HH:mm")} - {format(parseISO(block.ts_end), "HH:mm")}</div>
                                                        <div class="text-xs text-slate-400">{block.duration_minutes.toFixed(0)} min</div>
                                                    </td>
                                                    <td class="px-6 py-3">
                                                        <div class="text-sm text-slate-700 truncate max-w-md">
                                                            {block.title_summary || block.primary_app_name}
                                                        </div>
                                                        {#if block.activity_score !== undefined}
                                                            <div class="flex items-center gap-2 mt-1">
                                                                <div class="w-16 h-1.5 bg-slate-200 rounded-full overflow-hidden" title="Activity: {(block.activity_score * 100).toFixed(0)}%">
                                                                    <div class="h-full bg-green-500 rounded-full" style="width: {block.activity_score * 100}%"></div>
                                                                </div>
                                                            </div>
                                                        {/if}
                                                    </td>
                                                    <td class="px-6 py-3">
                                                        {#if block.profile_id}
                                                            <span class="text-xs px-2 py-1 rounded-md border bg-red-50 text-red-700 border-red-200">
                                                                LOW
                                                            </span>
                                                            {#if block.client_name}
                                                                <span class="ml-2 text-xs text-slate-500">{block.client_name}</span>
                                                            {/if}
                                                        {:else}
                                                            <span class="text-xs px-2 py-1 rounded-md bg-amber-50 text-amber-700 border border-amber-200">
                                                                Unassigned
                                                            </span>
                                                        {/if}
                                                    </td>
                                                    <td class="px-6 py-3">
                                                        <ProfileSelector
                                                            disabled={processingId === block.block_id}
                                                            on:change={(e) => assignBlock(block.block_id, e.detail)}
                                                        />
                                                    </td>
                                                    <td class="px-6 py-3 text-right">
                                                        {#if processingId === block.block_id}
                                                            <span class="text-xs text-slate-400">...</span>
                                                        {:else}
                                                            <button
                                                                on:click={() => deleteBlock(block.block_id)}
                                                                class="text-xs text-slate-400 hover:text-red-600"
                                                            >
                                                                Delete
                                                            </button>
                                                        {/if}
                                                    </td>
                                                </tr>
                                            {/each}
                                        </tbody>
                                    </table>
                                </div>
                            {/if}
                        </div>
                    {/each}
                </div>
            </section>
        {/if}

        <!-- Medium Confidence Section -->
        {#if mediumConfidenceGroups.length > 0}
            <section class="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
                <div class="px-6 py-4 border-b border-slate-100 flex items-center justify-between bg-blue-50/30">
                    <div class="flex items-center gap-3">
                        <span class="px-2.5 py-1 rounded-md text-xs font-semibold bg-blue-100 text-blue-800">
                            Review Suggested
                        </span>
                        <span class="text-sm text-slate-600">
                            {mediumConfidenceGroups.length} groups with medium confidence - verify if needed
                        </span>
                    </div>
                </div>
                <div class="divide-y divide-slate-100">
                    {#each mediumConfidenceGroups as group (group.group_key)}
                        <div class="border-l-4 border-l-blue-400">
                            <div class="px-6 py-4 flex items-center justify-between hover:bg-slate-50/50 transition-colors">
                                <button
                                    class="flex items-center gap-3 text-left flex-1"
                                    on:click={() => toggleGroup(group.group_key)}
                                >
                                    <div class="w-8 h-8 rounded-lg bg-slate-100 text-slate-700 flex items-center justify-center font-bold text-xs">
                                        {group.block_count}
                                    </div>
                                    <div class="flex-1 min-w-0">
                                        <div class="font-semibold text-slate-900 truncate">
                                            {group.title_context}
                                        </div>
                                        <div class="text-xs text-slate-500 flex items-center gap-2 flex-wrap">
                                            <span class="inline-flex items-center px-1.5 py-0.5 rounded bg-slate-100 text-slate-600">
                                                {group.primary_app_name}
                                            </span>
                                            <span>{group.total_minutes.toFixed(0)} min</span>
                                            <span class="text-slate-300">|</span>
                                            <span>{format(parseISO(group.first_ts), "MMM d")} - {format(parseISO(group.last_ts), "MMM d, HH:mm")}</span>
                                        </div>
                                    </div>
                                    <svg
                                        class="w-5 h-5 text-slate-400 transition-transform {expandedGroups.has(group.group_key) ? 'rotate-180' : ''}"
                                        fill="none" stroke="currentColor" viewBox="0 0 24 24"
                                    >
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                                    </svg>
                                </button>

                                <div class="flex items-center gap-3 ml-4">
                                    <div class="w-48">
                                        <ProfileSelector
                                            placeholder="Assign all..."
                                            on:change={(e) => assignAllInGroup(group, e.detail)}
                                        />
                                    </div>
                                    <button
                                        on:click={() => blacklistApp(group.primary_app_name, group.app_id)}
                                        disabled={blacklisting}
                                        class="text-xs text-slate-400 hover:text-red-600 px-2 py-1 rounded hover:bg-red-50 transition-colors"
                                        title="Blacklist this app"
                                    >
                                        Blacklist
                                    </button>
                                </div>
                            </div>

                            {#if expandedGroups.has(group.group_key)}
                                <div class="border-t border-slate-100 bg-slate-50/50">
                                    <table class="w-full text-left">
                                        <thead class="border-b border-slate-200">
                                            <tr>
                                                <th class="px-6 py-2 text-xs font-medium text-slate-500 uppercase">Time</th>
                                                <th class="px-6 py-2 text-xs font-medium text-slate-500 uppercase">Details</th>
                                                <th class="px-6 py-2 text-xs font-medium text-slate-500 uppercase">Status</th>
                                                <th class="px-6 py-2 text-xs font-medium text-slate-500 uppercase w-48">Assign</th>
                                                <th class="px-6 py-2 text-xs font-medium text-slate-500 uppercase w-20"></th>
                                            </tr>
                                        </thead>
                                        <tbody class="divide-y divide-slate-100 bg-white">
                                            {#each group.blocks as block (block.block_id)}
                                                <tr class="hover:bg-slate-50 transition-colors">
                                                    <td class="px-6 py-3 text-sm text-slate-600 whitespace-nowrap">
                                                        <div>{format(parseISO(block.ts_start), "HH:mm")} - {format(parseISO(block.ts_end), "HH:mm")}</div>
                                                        <div class="text-xs text-slate-400">{block.duration_minutes.toFixed(0)} min</div>
                                                    </td>
                                                    <td class="px-6 py-3">
                                                        <div class="text-sm text-slate-700 truncate max-w-md">
                                                            {block.title_summary || block.primary_app_name}
                                                        </div>
                                                        {#if block.activity_score !== undefined}
                                                            <div class="flex items-center gap-2 mt-1">
                                                                <div class="w-16 h-1.5 bg-slate-200 rounded-full overflow-hidden" title="Activity: {(block.activity_score * 100).toFixed(0)}%">
                                                                    <div class="h-full bg-green-500 rounded-full" style="width: {block.activity_score * 100}%"></div>
                                                                </div>
                                                            </div>
                                                        {/if}
                                                    </td>
                                                    <td class="px-6 py-3">
                                                        {#if block.profile_id}
                                                            <span class="text-xs px-2 py-1 rounded-md border bg-blue-50 text-blue-700 border-blue-200">
                                                                MEDIUM
                                                            </span>
                                                            {#if block.client_name}
                                                                <span class="ml-2 text-xs text-slate-500">{block.client_name}</span>
                                                            {/if}
                                                        {:else}
                                                            <span class="text-xs px-2 py-1 rounded-md bg-amber-50 text-amber-700 border border-amber-200">
                                                                Unassigned
                                                            </span>
                                                        {/if}
                                                    </td>
                                                    <td class="px-6 py-3">
                                                        <ProfileSelector
                                                            disabled={processingId === block.block_id}
                                                            on:change={(e) => assignBlock(block.block_id, e.detail)}
                                                        />
                                                    </td>
                                                    <td class="px-6 py-3 text-right">
                                                        {#if processingId === block.block_id}
                                                            <span class="text-xs text-slate-400">...</span>
                                                        {:else}
                                                            <button
                                                                on:click={() => deleteBlock(block.block_id)}
                                                                class="text-xs text-slate-400 hover:text-red-600"
                                                            >
                                                                Delete
                                                            </button>
                                                        {/if}
                                                    </td>
                                                </tr>
                                            {/each}
                                        </tbody>
                                    </table>
                                </div>
                            {/if}
                        </div>
                    {/each}
                </div>
            </section>
        {/if}

        <!-- Pagination -->
        {#if pagination.total_pages > 1}
            <div class="px-6 py-4 bg-white rounded-xl border border-slate-200 flex items-center justify-between">
                <div class="text-sm text-slate-500">
                    Page {pagination.page} of {pagination.total_pages} ({pagination.total} groups)
                </div>
                <div class="flex items-center gap-2">
                    <button
                        on:click={() => goToPage(pagination.page - 1)}
                        disabled={pagination.page <= 1}
                        class="px-3 py-1.5 text-sm font-medium rounded-lg border border-slate-200 bg-white hover:bg-slate-50 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        Previous
                    </button>
                    {#each pageNumbers as pageNum}
                        <button
                            on:click={() => goToPage(pageNum)}
                            class="w-8 h-8 text-sm font-medium rounded-lg {pageNum === pagination.page
                                ? 'bg-blue-600 text-white'
                                : 'border border-slate-200 bg-white hover:bg-slate-50 text-slate-700'}"
                        >
                            {pageNum}
                        </button>
                    {/each}
                    <button
                        on:click={() => goToPage(pagination.page + 1)}
                        disabled={pagination.page >= pagination.total_pages}
                        class="px-3 py-1.5 text-sm font-medium rounded-lg border border-slate-200 bg-white hover:bg-slate-50 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        Next
                    </button>
                </div>
            </div>
        {/if}
    {/if}
</div>
