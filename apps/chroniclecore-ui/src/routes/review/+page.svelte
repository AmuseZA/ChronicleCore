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
        // ML Suggestion fields
        has_ml_suggestion?: boolean;
        ml_suggested_profile_id?: number;
        ml_confidence?: number;
    }

    interface GroupedBlock {
        group_key: string;
        primary_app_name: string;
        app_id: number;
        title_context: string;
        date: string; // YYYY-MM-DD
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
    let pagination: Pagination = {
        page: 1,
        per_page: 50,
        total: 0,
        total_pages: 0,
    };
    let loading = true;
    let processingId: number | null = null;
    let expandedGroups: Set<string> = new Set();
    let blacklisting = false;

    // ML Status
    let mlStatus: MLStatus | null = null;
    let isTraining = false;
    let trainingMessage = "";

    // Settings
    let skipDeleteConfirmation = false;

    // Load settings from localStorage
    function loadSettings() {
        skipDeleteConfirmation =
            localStorage.getItem("skipDeleteConfirmation") === "true";
    }

    // Separate groups by confidence
    $: unassignedGroups = groups.filter((g) =>
        g.blocks.every((b) => !b.profile_id),
    );
    $: lowConfidenceGroups = groups.filter((g) =>
        g.blocks.some((b) => b.profile_id && b.confidence === "LOW"),
    );
    $: mediumConfidenceGroups = groups.filter(
        (g) =>
            g.blocks.some((b) => b.profile_id && b.confidence === "MEDIUM") &&
            !g.blocks.some((b) => b.confidence === "LOW"),
    );

    async function loadData(page = 1) {
        loading = true;
        try {
            const response = await fetchApi(
                `/blocks/grouped?needs_review=true&page=${page}&per_page=50`,
            );
            groups = response?.data || [];
            pagination = response?.pagination || {
                page: 1,
                per_page: 50,
                total: 0,
                total_pages: 0,
            };
            await loadMLStatus();
        } catch (e) {
            console.error(e);
        } finally {
            loading = false;
        }
    }

    async function loadMLStatus() {
        try {
            mlStatus = await fetchApi("/ml/status");
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
            const result = await fetchApi("/ml/train", { method: "POST" });
            trainingMessage = `Training complete! Accuracy: ${(result.metrics?.accuracy * 100 || 0).toFixed(1)}%`;
            await loadMLStatus();
            // After training, generate predictions
            await fetchApi("/ml/predict", { method: "POST" });
            await loadData(pagination.page);
        } catch (e: any) {
            trainingMessage = `Training failed: ${e.message || "Unknown error"}`;
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
        if (
            !skipDeleteConfirmation &&
            !confirm("Are you sure you want to delete this block?")
        )
            return;
        const scrollY = window.scrollY;
        processingId = blockId;
        try {
            await fetchApi(`/blocks/${blockId}`, { method: "DELETE" });
            await loadData(pagination.page);
            // Restore scroll position after DOM update
            requestAnimationFrame(() => window.scrollTo(0, scrollY));
        } catch (e) {
            console.error(e);
            alert("Failed to delete block");
        } finally {
            processingId = null;
        }
    }

    async function deleteGroup(group: GroupedBlock) {
        if (
            !skipDeleteConfirmation &&
            !confirm(
                `Are you sure you want to delete all ${group.block_count} items in this group?`,
            )
        )
            return;
        const scrollY = window.scrollY;
        processingId = group.blocks[0]?.block_id || null;
        try {
            for (const block of group.blocks) {
                await fetchApi(`/blocks/${block.block_id}`, {
                    method: "DELETE",
                });
            }
            await loadData(pagination.page);
            requestAnimationFrame(() => window.scrollTo(0, scrollY));
        } catch (e) {
            console.error(e);
            alert("Failed to delete group");
        } finally {
            processingId = null;
        }
    }

    // Blacklist Modal State
    let showBlacklistModal = false;
    let blacklistModalData: {
        appName: string;
        appId: number;
        titleContext: string;
    } | null = null;
    let blacklistKeywordInput = "";

    function openBlacklistModal(
        appName: string,
        appId: number,
        titleContext: string,
    ) {
        blacklistModalData = { appName, appId, titleContext };
        blacklistKeywordInput = titleContext;
        showBlacklistModal = true;
    }

    function closeBlacklistModal() {
        showBlacklistModal = false;
        blacklistModalData = null;
        blacklistKeywordInput = "";
    }

    async function blacklistByApp() {
        if (!blacklistModalData) return;
        blacklisting = true;
        try {
            await fetchApi("/blacklist/with-delete", {
                method: "POST",
                body: JSON.stringify({
                    app_id: blacklistModalData.appId,
                    reason: "Blacklisted from review page",
                }),
            });
            closeBlacklistModal();
            await loadData(pagination.page);
        } catch (e) {
            console.error(e);
            alert("Failed to blacklist app");
        } finally {
            blacklisting = false;
        }
    }

    async function blacklistByKeyword() {
        if (!blacklistKeywordInput.trim()) {
            alert("Please enter a keyword");
            return;
        }
        blacklisting = true;
        try {
            await fetchApi("/blacklist/keywords", {
                method: "POST",
                body: JSON.stringify({
                    keyword: blacklistKeywordInput.trim(),
                    reason: "Blacklisted from review page",
                }),
            });
            closeBlacklistModal();
            await loadData(pagination.page);
        } catch (e) {
            console.error(e);
            alert("Failed to blacklist keyword");
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

    onMount(() => {
        loadSettings();
        loadData();
    });
</script>

<div class="max-w-7xl mx-auto space-y-6 pb-20">
    <!-- Header -->
    <header class="flex justify-between items-start">
        <div>
            <h1 class="text-2xl font-bold text-slate-900">Needs Review</h1>
            <p class="text-slate-500">
                Activities grouped by context - assign profiles or blacklist
                apps.
            </p>
        </div>
        <div class="flex items-center gap-3">
            <!-- Add Entry Link -->
            <a
                href="/history"
                class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
            >
                + Add Entry
            </a>
            <!-- ML Status & Training -->
            {#if mlStatus}
                <div
                    class="flex items-center gap-3 px-4 py-2 bg-slate-50 rounded-lg border border-slate-200"
                >
                    <div class="text-xs text-slate-500">
                        <span class="font-medium"
                            >{mlStatus.training_samples}</span
                        > data points
                    </div>
                    {#if mlStatus.has_trained_model}
                        <div
                            class="w-2 h-2 rounded-full bg-green-500"
                            title="Model trained"
                        ></div>
                    {/if}
                    <button
                        on:click={triggerTraining}
                        disabled={!mlStatus.ready_for_training || isTraining}
                        class="px-3 py-1.5 text-xs font-medium rounded-md transition-colors
                               {mlStatus.ready_for_training && !isTraining
                            ? 'bg-blue-600 text-white hover:bg-blue-700'
                            : 'bg-slate-200 text-slate-400 cursor-not-allowed'}"
                    >
                        {isTraining ? "Training..." : "Train ML"}
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
        <div
            class="px-4 py-3 rounded-lg text-sm {trainingMessage.includes(
                'failed',
            )
                ? 'bg-red-50 text-red-700 border border-red-200'
                : 'bg-green-50 text-green-700 border border-green-200'}"
        >
            {trainingMessage}
        </div>
    {/if}

    {#if loading}
        <div
            class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-12 text-center text-slate-500 dark:text-slate-400"
        >
            Loading...
        </div>
    {:else if groups.length === 0}
        <div
            class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-12 text-center"
        >
            <div
                class="w-12 h-12 bg-green-100 text-green-600 rounded-full flex items-center justify-center mx-auto mb-3"
            >
                <svg
                    class="w-6 h-6"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                >
                    <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M5 13l4 4L19 7"
                    />
                </svg>
            </div>
            <h3 class="text-lg font-medium text-slate-900">All caught up!</h3>
            <p class="text-slate-500">No activities need review right now.</p>
        </div>
    {:else}
        <div class="space-y-8">
            <!-- Unassigned Section -->
            {#if unassignedGroups.length > 0}
                <section>
                    <div class="flex items-center gap-3 mb-4">
                        <span
                            class="px-2.5 py-1 rounded-md text-xs font-semibold bg-amber-100 text-amber-800 uppercase tracking-wider"
                        >
                            Unassigned
                        </span>
                        <div class="h-px bg-slate-200 flex-1"></div>
                    </div>
                    <div class="grid grid-cols-1 gap-4">
                        {#each unassignedGroups as group (group.group_key)}
                            <div
                                class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 shadow-sm overflow-hidden hover:shadow-md transition-shadow"
                            >
                                <div class="p-5">
                                    <div
                                        class="flex justify-between items-start gap-4"
                                    >
                                        <div class="flex-1 min-w-0">
                                            <div
                                                class="flex items-center gap-2 mb-1"
                                            >
                                                {#if group.date}
                                                    <span
                                                        class="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold bg-blue-100 text-blue-700 border border-blue-200 whitespace-nowrap"
                                                    >
                                                        {format(parseISO(group.date), "MMM d")}
                                                    </span>
                                                {/if}
                                                <h3
                                                    class="font-semibold text-lg text-slate-900 leading-tight break-words"
                                                >
                                                    {group.title_context}
                                                </h3>
                                                <span
                                                    class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium bg-slate-100 text-slate-600 border border-slate-200 uppercase tracking-wide whitespace-nowrap"
                                                >
                                                    {group.primary_app_name}
                                                </span>
                                            </div>
                                            <div
                                                class="flex items-center flex-wrap gap-x-4 gap-y-1 text-sm text-slate-500 mt-2"
                                            >
                                                <div
                                                    class="flex items-center gap-1.5"
                                                >
                                                    <svg
                                                        class="w-4 h-4 text-slate-400"
                                                        fill="none"
                                                        viewBox="0 0 24 24"
                                                        stroke="currentColor"
                                                    >
                                                        <path
                                                            stroke-linecap="round"
                                                            stroke-linejoin="round"
                                                            stroke-width="2"
                                                            d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                                                        />
                                                    </svg>
                                                    <span
                                                        class="font-medium text-slate-700"
                                                        >{group.total_minutes.toFixed(
                                                            0,
                                                        )} min</span
                                                    >
                                                </div>
                                                <div
                                                    class="flex items-center gap-1.5"
                                                >
                                                    <svg
                                                        class="w-4 h-4 text-slate-400"
                                                        fill="none"
                                                        viewBox="0 0 24 24"
                                                        stroke="currentColor"
                                                    >
                                                        <path
                                                            stroke-linecap="round"
                                                            stroke-linejoin="round"
                                                            stroke-width="2"
                                                            d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
                                                        />
                                                    </svg>
                                                    <span
                                                        >{format(
                                                            parseISO(
                                                                group.first_ts,
                                                            ),
                                                            "MMM d",
                                                        )}</span
                                                    >
                                                    <span
                                                        class="text-slate-300 mx-1"
                                                        >•</span
                                                    >
                                                    <span
                                                        >{format(
                                                            parseISO(
                                                                group.first_ts,
                                                            ),
                                                            "HH:mm",
                                                        )} - {format(
                                                            parseISO(
                                                                group.last_ts,
                                                            ),
                                                            "HH:mm",
                                                        )}</span
                                                    >
                                                </div>
                                                <div
                                                    class="flex items-center gap-1.5"
                                                >
                                                    <svg
                                                        class="w-4 h-4 text-slate-400"
                                                        fill="none"
                                                        viewBox="0 0 24 24"
                                                        stroke="currentColor"
                                                    >
                                                        <path
                                                            stroke-linecap="round"
                                                            stroke-linejoin="round"
                                                            stroke-width="2"
                                                            d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
                                                        />
                                                    </svg>
                                                    <span
                                                        >{group.block_count} items</span
                                                    >
                                                </div>
                                            </div>
                                        </div>
                                        <div
                                            class="flex flex-col gap-2 min-w-[140px]"
                                        >
                                            <ProfileSelector
                                                placeholder="Assign Group..."
                                                on:change={(e) =>
                                                    assignAllInGroup(
                                                        group,
                                                        e.detail,
                                                    )}
                                            />
                                        </div>
                                    </div>
                                </div>
                                <div
                                    class="bg-slate-50 px-5 py-3 border-t border-slate-100 flex justify-between items-center"
                                >
                                    <button
                                        on:click={() =>
                                            toggleGroup(group.group_key)}
                                        class="text-xs font-medium text-slate-500 hover:text-slate-800 flex items-center gap-1"
                                    >
                                        {#if expandedGroups.has(group.group_key)}
                                            Hide Details
                                            <svg
                                                class="w-4 h-4"
                                                fill="none"
                                                stroke="currentColor"
                                                viewBox="0 0 24 24"
                                                ><path
                                                    stroke-linecap="round"
                                                    stroke-linejoin="round"
                                                    stroke-width="2"
                                                    d="M5 15l7-7 7 7"
                                                /></svg
                                            >
                                        {:else}
                                            Show {group.block_count} Items
                                            <svg
                                                class="w-4 h-4"
                                                fill="none"
                                                stroke="currentColor"
                                                viewBox="0 0 24 24"
                                                ><path
                                                    stroke-linecap="round"
                                                    stroke-linejoin="round"
                                                    stroke-width="2"
                                                    d="M19 9l-7 7-7-7"
                                                /></svg
                                            >
                                        {/if}
                                    </button>
                                    <div class="flex gap-3">
                                        <button
                                            on:click={() => deleteGroup(group)}
                                            class="text-xs text-slate-400 hover:text-red-600 hover:underline"
                                        >
                                            Delete Group
                                        </button>
                                        <button
                                            on:click={() =>
                                                openBlacklistModal(
                                                    group.primary_app_name,
                                                    group.app_id,
                                                    group.title_context,
                                                )}
                                            class="text-xs text-slate-400 hover:text-red-600 hover:underline"
                                        >
                                            Blacklist
                                        </button>
                                    </div>
                                </div>

                                {#if expandedGroups.has(group.group_key)}
                                    <div
                                        class="border-t border-slate-100 bg-white"
                                    >
                                        <div class="divide-y divide-slate-100">
                                            {#each group.blocks as block}
                                                <div
                                                    class="p-4 hover:bg-slate-50 transition-colors flex gap-4 items-start text-sm"
                                                >
                                                    <div
                                                        class="w-32 shrink-0 text-slate-500 text-xs mt-0.5"
                                                    >
                                                        <div>
                                                            {format(
                                                                parseISO(
                                                                    block.ts_start,
                                                                ),
                                                                "HH:mm",
                                                            )} - {format(
                                                                parseISO(
                                                                    block.ts_end,
                                                                ),
                                                                "HH:mm",
                                                            )}
                                                        </div>
                                                        <div
                                                            class="mt-1 font-medium text-slate-600"
                                                        >
                                                            {block.duration_minutes.toFixed(
                                                                0,
                                                            )} min
                                                        </div>
                                                    </div>
                                                    <div class="flex-1 min-w-0">
                                                        <p
                                                            class="text-slate-700 break-words leading-relaxed"
                                                        >
                                                            {block.title_summary ||
                                                                block.primary_app_name}
                                                        </p>
                                                        {#if block.activity_score !== undefined}
                                                            <div
                                                                class="flex items-center gap-2 mt-2"
                                                            >
                                                                <div
                                                                    class="w-16 h-1 bg-slate-100 rounded-full overflow-hidden"
                                                                >
                                                                    <div
                                                                        class="h-full bg-green-500 rounded-full"
                                                                        style="width: {block.activity_score *
                                                                            100}%"
                                                                    ></div>
                                                                </div>
                                                                <span
                                                                    class="text-[10px] text-slate-400"
                                                                    >Activity
                                                                    Level</span
                                                                >
                                                            </div>
                                                        {/if}
                                                        <!-- ML Suggestion Badge -->
                                                        {#if block.confidence?.startsWith("ML")}
                                                            <div
                                                                class="mt-2 inline-flex items-center gap-1.5 px-2 py-1 bg-purple-50 text-purple-700 text-xs rounded border border-purple-100"
                                                            >
                                                                <svg
                                                                    class="w-3 h-3"
                                                                    fill="none"
                                                                    stroke="currentColor"
                                                                    viewBox="0 0 24 24"
                                                                    ><path
                                                                        stroke-linecap="round"
                                                                        stroke-linejoin="round"
                                                                        stroke-width="2"
                                                                        d="M13 10V3L4 14h7v7l9-11h-7z"
                                                                    /></svg
                                                                >
                                                                <span
                                                                    class="font-medium"
                                                                    >AI
                                                                    Suggested</span
                                                                >
                                                            </div>
                                                        {/if}
                                                    </div>
                                                    <div
                                                        class="w-48 shrink-0 flex flex-col gap-2 items-end"
                                                    >
                                                        <ProfileSelector
                                                            placeholder={block.has_ml_suggestion
                                                                ? "Has AI suggestion..."
                                                                : "Assign..."}
                                                            value={block.profile_id}
                                                            disabled={processingId ===
                                                                block.block_id}
                                                            on:change={(e) =>
                                                                assignBlock(
                                                                    block.block_id,
                                                                    e.detail,
                                                                )}
                                                        />
                                                        <button
                                                            on:click={() =>
                                                                deleteBlock(
                                                                    block.block_id,
                                                                )}
                                                            class="text-xs text-slate-400 hover:text-red-600"
                                                        >
                                                            Delete
                                                        </button>
                                                    </div>
                                                </div>
                                            {/each}
                                        </div>
                                    </div>
                                {/if}
                            </div>
                        {/each}
                    </div>
                </section>
            {/if}

            <!-- Verified/Review Section (Medium/Low Confidence) -->
            {#if mediumConfidenceGroups.length > 0 || lowConfidenceGroups.length > 0}
                <section>
                    <div class="flex items-center gap-3 mb-4 mt-8">
                        <span
                            class="px-2.5 py-1 rounded-md text-xs font-semibold bg-blue-100 text-blue-800 uppercase tracking-wider"
                        >
                            Verify Assignments
                        </span>
                        <div class="h-px bg-slate-200 flex-1"></div>
                    </div>
                    <div class="grid grid-cols-1 gap-4">
                        {#each [...mediumConfidenceGroups, ...lowConfidenceGroups] as group}
                            <div
                                class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 shadow-sm overflow-hidden hover:shadow-md transition-shadow"
                            >
                                <div class="p-5">
                                    <div
                                        class="flex justify-between items-start gap-4"
                                    >
                                        <div class="flex-1 min-w-0">
                                            <div
                                                class="flex items-center gap-2 mb-1"
                                            >
                                                {#if group.date}
                                                    <span
                                                        class="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold bg-blue-100 text-blue-700 border border-blue-200 whitespace-nowrap"
                                                    >
                                                        {format(parseISO(group.date), "MMM d")}
                                                    </span>
                                                {/if}
                                                <h3
                                                    class="font-semibold text-lg text-slate-900 leading-tight break-words"
                                                >
                                                    {group.title_context}
                                                </h3>
                                                {#if group.blocks.some( (b) => b.confidence?.startsWith("ML"), )}
                                                    <span
                                                        class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-bold bg-purple-100 text-purple-700 border border-purple-200 uppercase tracking-wide"
                                                    >
                                                        AI MATCH
                                                    </span>
                                                {/if}
                                            </div>
                                            <div
                                                class="flex items-center flex-wrap gap-x-4 gap-y-1 text-sm text-slate-500 mt-2"
                                            >
                                                <span
                                                    class="font-medium text-slate-700"
                                                    >{group.primary_app_name}</span
                                                >
                                                <span class="text-slate-300"
                                                    >|</span
                                                >
                                                <span
                                                    >{group.total_minutes.toFixed(
                                                        0,
                                                    )} min</span
                                                >
                                                <span class="text-slate-300"
                                                    >|</span
                                                >
                                                <span
                                                    >{group.block_count} items</span
                                                >
                                            </div>
                                        </div>
                                        <div
                                            class="flex flex-col gap-2 min-w-[140px]"
                                        >
                                            <ProfileSelector
                                                placeholder="Reassign..."
                                                value={group.blocks.find(
                                                    (b) => b.profile_id,
                                                )?.profile_id}
                                                on:change={(e) =>
                                                    assignAllInGroup(
                                                        group,
                                                        e.detail,
                                                    )}
                                            />
                                        </div>
                                    </div>
                                </div>
                                <div
                                    class="bg-slate-50 px-5 py-3 border-t border-slate-100 flex justify-between items-center text-xs"
                                >
                                    <button
                                        on:click={() =>
                                            toggleGroup(group.group_key)}
                                        class="font-medium text-slate-500 hover:text-slate-800 flex items-center gap-1"
                                    >
                                        {#if expandedGroups.has(group.group_key)}
                                            Hide Details
                                            <svg
                                                class="w-4 h-4"
                                                fill="none"
                                                stroke="currentColor"
                                                viewBox="0 0 24 24"
                                                ><path
                                                    stroke-linecap="round"
                                                    stroke-linejoin="round"
                                                    stroke-width="2"
                                                    d="M5 15l7-7 7 7"
                                                /></svg
                                            >
                                        {:else}
                                            Show Details
                                            <svg
                                                class="w-4 h-4"
                                                fill="none"
                                                stroke="currentColor"
                                                viewBox="0 0 24 24"
                                                ><path
                                                    stroke-linecap="round"
                                                    stroke-linejoin="round"
                                                    stroke-width="2"
                                                    d="M19 9l-7 7-7-7"
                                                /></svg
                                            >
                                        {/if}
                                    </button>
                                </div>
                                {#if expandedGroups.has(group.group_key)}
                                    <div
                                        class="border-t border-slate-100 bg-white"
                                    >
                                        <div class="divide-y divide-slate-100">
                                            {#each group.blocks as block}
                                                <div
                                                    class="p-4 hover:bg-slate-50 transition-colors flex gap-4 items-start text-sm"
                                                >
                                                    <div
                                                        class="w-32 shrink-0 text-slate-500 text-xs mt-0.5"
                                                    >
                                                        <div>
                                                            {format(
                                                                parseISO(
                                                                    block.ts_start,
                                                                ),
                                                                "HH:mm",
                                                            )} - {format(
                                                                parseISO(
                                                                    block.ts_end,
                                                                ),
                                                                "HH:mm",
                                                            )}
                                                        </div>
                                                    </div>
                                                    <div class="flex-1 min-w-0">
                                                        <p
                                                            class="text-slate-700 break-words"
                                                        >
                                                            {block.title_summary ||
                                                                block.primary_app_name}
                                                        </p>
                                                        {#if block.confidence?.startsWith("ML")}
                                                            <div
                                                                class="mt-1 text-purple-600 text-xs font-medium flex items-center gap-1"
                                                            >
                                                                <svg
                                                                    class="w-3 h-3"
                                                                    fill="none"
                                                                    stroke="currentColor"
                                                                    viewBox="0 0 24 24"
                                                                    ><path
                                                                        stroke-linecap="round"
                                                                        stroke-linejoin="round"
                                                                        stroke-width="2"
                                                                        d="M13 10V3L4 14h7v7l9-11h-7z"
                                                                    /></svg
                                                                >
                                                                AI Suggested Match
                                                            </div>
                                                        {/if}
                                                    </div>
                                                    <div class="w-48 shrink-0">
                                                        <ProfileSelector
                                                            value={block.profile_id}
                                                            disabled={processingId ===
                                                                block.block_id}
                                                            on:change={(e) =>
                                                                assignBlock(
                                                                    block.block_id,
                                                                    e.detail,
                                                                )}
                                                        />
                                                    </div>
                                                </div>
                                            {/each}
                                        </div>
                                    </div>
                                {/if}
                            </div>
                        {/each}
                    </div>
                </section>
            {/if}

            <!-- Pagination -->
            {#if pagination.total_pages > 1}
                <div class="flex justify-center pt-8">
                    <div
                        class="inline-flex items-center gap-2 p-1 bg-white border border-slate-200 rounded-lg shadow-sm"
                    >
                        <button
                            on:click={() => goToPage(pagination.page - 1)}
                            disabled={pagination.page <= 1}
                            class="px-3 py-1.5 text-sm font-medium rounded-md hover:bg-slate-50 disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            Previous
                        </button>
                        {#each pageNumbers as pageNum}
                            <button
                                on:click={() => goToPage(pageNum)}
                                class="w-8 h-8 text-sm font-medium rounded-md {pageNum ===
                                pagination.page
                                    ? 'bg-slate-900 text-white shadow-sm'
                                    : 'text-slate-600 hover:bg-slate-50'}"
                            >
                                {pageNum}
                            </button>
                        {/each}
                        <button
                            on:click={() => goToPage(pagination.page + 1)}
                            disabled={pagination.page >= pagination.total_pages}
                            class="px-3 py-1.5 text-sm font-medium rounded-md hover:bg-slate-50 disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            Next
                        </button>
                    </div>
                </div>
            {/if}
        </div>
    {/if}
</div>

<!-- Blacklist Modal -->
{#if showBlacklistModal && blacklistModalData}
    <div
        class="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
        on:click={closeBlacklistModal}
    >
        <div
            class="bg-white dark:bg-slate-800 rounded-2xl shadow-xl max-w-md w-full mx-4 overflow-hidden"
            on:click|stopPropagation
        >
            <div class="px-6 py-4 border-b border-slate-100 bg-slate-50">
                <h3 class="text-lg font-semibold text-slate-900">
                    Blacklist Options
                </h3>
                <p class="text-sm text-slate-500 mt-1">
                    Choose how to block "{blacklistModalData.titleContext}"
                    activities
                </p>
            </div>

            <div class="p-6 space-y-4">
                <!-- Option 1: Block by App -->
                <div
                    class="border border-slate-200 rounded-xl p-4 hover:border-slate-300 transition-colors"
                >
                    <div class="flex items-start gap-3">
                        <div
                            class="w-10 h-10 bg-red-100 rounded-lg flex items-center justify-center flex-shrink-0"
                        >
                            <svg
                                class="w-5 h-5 text-red-600"
                                fill="none"
                                viewBox="0 0 24 24"
                                stroke="currentColor"
                            >
                                <path
                                    stroke-linecap="round"
                                    stroke-linejoin="round"
                                    stroke-width="2"
                                    d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636"
                                />
                            </svg>
                        </div>
                        <div class="flex-1">
                            <h4 class="font-medium text-slate-900">
                                Block Entire App
                            </h4>
                            <p class="text-sm text-slate-500 mt-0.5">
                                Hide all activities from <strong
                                    >{blacklistModalData.appName}</strong
                                >
                            </p>
                            <button
                                on:click={blacklistByApp}
                                disabled={blacklisting}
                                class="mt-3 px-4 py-2 bg-red-600 text-white text-sm font-medium rounded-lg hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                            >
                                {blacklisting
                                    ? "Blocking..."
                                    : `Block ${blacklistModalData.appName}`}
                            </button>
                        </div>
                    </div>
                </div>

                <!-- Option 2: Block by Keyword -->
                <div
                    class="border border-slate-200 rounded-xl p-4 hover:border-slate-300 transition-colors"
                >
                    <div class="flex items-start gap-3">
                        <div
                            class="w-10 h-10 bg-amber-100 rounded-lg flex items-center justify-center flex-shrink-0"
                        >
                            <svg
                                class="w-5 h-5 text-amber-600"
                                fill="none"
                                viewBox="0 0 24 24"
                                stroke="currentColor"
                            >
                                <path
                                    stroke-linecap="round"
                                    stroke-linejoin="round"
                                    stroke-width="2"
                                    d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z"
                                />
                            </svg>
                        </div>
                        <div class="flex-1">
                            <h4 class="font-medium text-slate-900">
                                Block by Keyword
                            </h4>
                            <p class="text-sm text-slate-500 mt-0.5">
                                Hide activities containing a specific word or
                                phrase
                            </p>
                            <input
                                type="text"
                                bind:value={blacklistKeywordInput}
                                placeholder="Enter keyword..."
                                class="mt-3 w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none"
                            />
                            <button
                                on:click={blacklistByKeyword}
                                disabled={blacklisting ||
                                    !blacklistKeywordInput.trim()}
                                class="mt-2 px-4 py-2 bg-amber-600 text-white text-sm font-medium rounded-lg hover:bg-amber-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                            >
                                {blacklisting ? "Blocking..." : "Block Keyword"}
                            </button>
                        </div>
                    </div>
                </div>
            </div>

            <div
                class="px-6 py-4 border-t border-slate-100 bg-slate-50 flex justify-end"
            >
                <button
                    on:click={closeBlacklistModal}
                    class="px-4 py-2 text-sm font-medium text-slate-600 hover:text-slate-800 transition-colors"
                >
                    Cancel
                </button>
            </div>
        </div>
    </div>
{/if}
