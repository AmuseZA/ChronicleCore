<script lang="ts">
    import { onMount } from "svelte";
    import { fetchApi } from "$lib/api";
    import TrackingControl from "$lib/components/TrackingControl.svelte";
    import ActivityGroup from "$lib/components/ActivityGroup.svelte";
    import ManualEntryModal from "$lib/components/ManualEntryModal.svelte";

    interface Activity {
        block_id: number;
        ts_start: string;
        ts_end: string;
        duration_minutes: number;
        primary_app_name: string;
        title_summary?: string;
        description?: string;
        confidence: string;
    }

    interface ActivityGroupType {
        profile_id?: number;
        profile_name: string;
        client_name?: string;
        start_time: string;
        end_time: string;
        total_minutes: number;
        summary: string;
        apps: string[];
        activities: Activity[];
        color?: string;
    }

    let stats = {
        totalHours: 0,
        productivityScore: 0,
        billableHours: 0,
        reviewCount: 0,
        totalActivities: 0,
    };

    let blocks: any[] = [];
    let activityGroups: ActivityGroupType[] = [];
    let loading = true;
    let error: string | null = null;
    let showManualModal = false;

    // Group blocks by profile and contiguous time (gaps < 30 minutes)
    function groupBlocks(blocks: any[]): ActivityGroupType[] {
        if (!blocks.length) return [];

        const groups: ActivityGroupType[] = [];
        let currentGroup: ActivityGroupType | null = null;

        // Sort blocks by start time (newest first for display, but we need oldest first for grouping)
        const sortedBlocks = [...blocks].sort((a, b) =>
            new Date(a.ts_start).getTime() - new Date(b.ts_start).getTime()
        );

        for (const block of sortedBlocks) {
            const blockStart = new Date(block.ts_start);
            const shouldStartNewGroup = !currentGroup ||
                block.profile_id !== currentGroup.profile_id ||
                (currentGroup.end_time &&
                    (blockStart.getTime() - new Date(currentGroup.end_time).getTime()) > 30 * 60 * 1000);

            if (shouldStartNewGroup) {
                if (currentGroup) {
                    currentGroup.summary = generateSummary(currentGroup);
                    groups.push(currentGroup);
                }
                currentGroup = {
                    profile_id: block.profile_id,
                    profile_name: block.client_name || "Unassigned",
                    client_name: block.client_name,
                    start_time: block.ts_start,
                    end_time: block.ts_end,
                    total_minutes: block.duration_minutes || 0,
                    summary: "",
                    apps: [block.primary_app_name],
                    activities: [{
                        block_id: block.block_id,
                        ts_start: block.ts_start,
                        ts_end: block.ts_end,
                        duration_minutes: block.duration_minutes || 0,
                        primary_app_name: block.primary_app_name,
                        title_summary: block.title_summary,
                        description: block.description,
                        confidence: block.confidence || "MEDIUM",
                    }],
                };
            } else {
                currentGroup!.end_time = block.ts_end;
                currentGroup!.total_minutes += block.duration_minutes || 0;
                if (!currentGroup!.apps.includes(block.primary_app_name)) {
                    currentGroup!.apps.push(block.primary_app_name);
                }
                currentGroup!.activities.push({
                    block_id: block.block_id,
                    ts_start: block.ts_start,
                    ts_end: block.ts_end,
                    duration_minutes: block.duration_minutes || 0,
                    primary_app_name: block.primary_app_name,
                    title_summary: block.title_summary,
                    description: block.description,
                    confidence: block.confidence || "MEDIUM",
                });
            }
        }

        if (currentGroup) {
            currentGroup.summary = generateSummary(currentGroup);
            groups.push(currentGroup);
        }

        // Return in reverse order (newest first)
        return groups.reverse();
    }

    function generateSummary(group: ActivityGroupType): string {
        const activities = group.activities;
        if (activities.length === 1) {
            return activities[0].title_summary || activities[0].primary_app_name;
        }

        // Count apps and create summary
        const appCounts: Record<string, number> = {};
        activities.forEach(a => {
            appCounts[a.primary_app_name] = (appCounts[a.primary_app_name] || 0) + 1;
        });

        const topApps = Object.entries(appCounts)
            .sort((a, b) => b[1] - a[1])
            .slice(0, 3)
            .map(([app]) => app);

        if (topApps.length === 1) {
            return `Working in ${topApps[0]}`;
        }
        return `Working across ${topApps.join(", ")}`;
    }

    async function loadData() {
        try {
            loading = true;
            error = null;
            const res = await fetchApi("/blocks");
            blocks = res || [];

            // Group blocks for Timely-style display
            activityGroups = groupBlocks(blocks);

            // Calculate Stats
            let totalDur = 0;
            let totalScore = 0;
            let scoreCount = 0;
            let billable = 0;
            let review = 0;

            blocks.forEach((b) => {
                totalDur += b.duration_hours || 0;

                if (b.activity_score !== undefined) {
                    totalScore += b.activity_score;
                    scoreCount++;
                }

                if (b.billable && b.profile_id) {
                    billable += b.duration_hours || 0;
                }

                if (b.confidence === "LOW" || !b.profile_id) {
                    review++;
                }
            });

            stats = {
                totalHours: totalDur,
                productivityScore:
                    scoreCount > 0 ? (totalScore / scoreCount) * 100 : 0,
                billableHours: billable,
                reviewCount: review,
                totalActivities: blocks.length,
            };
        } catch (err: any) {
            console.error(err);
            error = err.message || "Failed to load blocks";
        } finally {
            loading = false;
        }
    }

    function handleManualEntryCreated() {
        loadData();
    }

    onMount(() => {
        loadData();
        // Poll for block updates every minute
        const interval = setInterval(loadData, 60000);
        return () => clearInterval(interval);
    });
</script>

<div class="max-w-7xl mx-auto space-y-8">
    <div class="flex items-center justify-between">
        <div>
            <h1 class="text-2xl font-bold text-slate-900 tracking-tight">
                Today's Overview
            </h1>
            <p class="text-slate-500 text-sm">
                {new Date().toLocaleDateString(undefined, {
                    weekday: "long",
                    year: "numeric",
                    month: "long",
                    day: "numeric",
                })}
            </p>
        </div>
        <div class="flex gap-2">
            <button
                on:click={() => (showManualModal = true)}
                class="flex items-center gap-2 bg-white hover:bg-slate-50 text-slate-700 px-4 py-2 rounded-lg font-medium text-sm transition-colors border border-slate-200 shadow-sm"
            >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                </svg>
                Add Entry
            </button>
            <a href="/review" class="relative group">
                <button
                    class="bg-indigo-600 hover:bg-indigo-700 text-white px-4 py-2 rounded-lg font-medium text-sm transition-colors shadow-sm"
                >
                    Review Queue
                </button>
                {#if stats.reviewCount > 0}
                    <span
                        class="absolute -top-1 -right-1 bg-red-500 text-white text-[10px] font-bold px-1.5 py-0.5 rounded-full border-2 border-white shadow-sm"
                    >
                        {stats.reviewCount}
                    </span>
                {/if}
            </a>
        </div>
    </div>

    <!-- Summary Cards -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
        <!-- Total Time Worked (NEW - First Card) -->
        <div
            class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm relative overflow-hidden group"
        >
            <div
                class="absolute top-0 right-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity"
            >
                <svg
                    class="w-16 h-16 text-indigo-600"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                    ><path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="1.5"
                        d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                    ></path></svg
                >
            </div>
            <div
                class="text-sm font-medium text-slate-500 uppercase tracking-wider mb-1"
            >
                Time Worked
            </div>
            <div class="text-3xl font-bold text-slate-900">
                {stats.totalHours.toFixed(1)}
                <span class="text-lg text-slate-400 font-normal">hrs</span>
            </div>
            <div class="mt-2 text-sm text-slate-500">
                {stats.totalActivities} activities
            </div>
        </div>

        <!-- Billable -->
        <div
            class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm relative overflow-hidden group"
        >
            <div
                class="absolute top-0 right-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity"
            >
                <svg
                    class="w-16 h-16 text-amber-600"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                    ><path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="1.5"
                        d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                    ></path></svg
                >
            </div>
            <div
                class="text-sm font-medium text-slate-500 uppercase tracking-wider mb-1"
            >
                Billable
            </div>
            <div class="text-3xl font-bold text-slate-900">
                {stats.billableHours.toFixed(1)}
                <span class="text-lg text-slate-400 font-normal">hrs</span>
            </div>
            <div
                class="mt-4 h-1.5 w-full bg-slate-100 rounded-full overflow-hidden"
            >
                <div
                    class="h-full bg-amber-500 rounded-full"
                    style="width: {stats.totalHours > 0
                        ? (stats.billableHours / stats.totalHours) * 100
                        : 0}%"
                ></div>
            </div>
        </div>

        <!-- Productivity Score -->
        <div
            class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm relative overflow-hidden group"
        >
            <div
                class="absolute top-0 right-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity"
            >
                <svg
                    class="w-16 h-16 text-green-600"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                    ><path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="1.5"
                        d="M13 10V3L4 14h7v7l9-11h-7z"
                    ></path></svg
                >
            </div>
            <div
                class="text-sm font-medium text-slate-500 uppercase tracking-wider mb-1"
            >
                Avg Focus
            </div>
            <div class="text-3xl font-bold text-slate-900">
                {stats.productivityScore.toFixed(0)}<span
                    class="text-lg text-slate-400 font-normal">%</span
                >
            </div>
            <div
                class="mt-4 h-1.5 w-full bg-slate-100 rounded-full overflow-hidden"
            >
                <div
                    class="h-full bg-green-500 rounded-full"
                    style="width: {stats.productivityScore}%"
                ></div>
            </div>
        </div>

        <!-- Needs Review -->
        <div
            class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm relative overflow-hidden group"
        >
            <div
                class="absolute top-0 right-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity"
            >
                <svg
                    class="w-16 h-16 text-rose-600"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                    ><path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="1.5"
                        d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"
                    ></path></svg
                >
            </div>
            <div
                class="text-sm font-medium text-slate-500 uppercase tracking-wider mb-1"
            >
                Needs Review
            </div>
            <div class="text-3xl font-bold text-slate-900">
                {stats.reviewCount}
                <span class="text-lg text-slate-400 font-normal">items</span>
            </div>
            <div class="mt-2">
                <a
                    href="/review"
                    class="text-sm text-indigo-600 hover:text-indigo-700 font-medium"
                >
                    Review now &rarr;
                </a>
            </div>
        </div>
    </div>

    {#if error}
        <div
            class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg flex items-center gap-2"
        >
            <svg
                class="w-5 h-5"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                ><path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                /></svg
            >
            <span
                >Error: {error} - Check backend connection at 127.0.0.1:8080</span
            >
        </div>
    {/if}

    <div
        class="bg-white rounded-2xl border border-slate-200 shadow-sm p-6 space-y-4"
    >
        <h2 class="text-lg font-bold text-slate-900">Live Activity</h2>
        <TrackingControl />
    </div>

    <div
        class="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden"
    >
        <div
            class="px-6 py-4 border-b border-slate-100 flex justify-between items-center bg-slate-50/50"
        >
            <h2 class="text-lg font-bold text-slate-900">Timeline</h2>
            <span class="text-sm text-slate-500">
                {activityGroups.length} groups
            </span>
        </div>
        <div class="p-4 space-y-3">
            {#if loading}
                <div class="text-center py-8 text-slate-500">
                    Loading activities...
                </div>
            {:else if activityGroups.length === 0}
                <div class="text-center py-12">
                    <div class="mx-auto w-12 h-12 bg-slate-100 rounded-full flex items-center justify-center mb-3 text-slate-400">
                        <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                    </div>
                    <h3 class="text-sm font-medium text-slate-900">No activity yet</h3>
                    <p class="text-sm text-slate-500 mt-1">Start tracking to see your work blocks here.</p>
                </div>
            {:else}
                {#each activityGroups as group}
                    <ActivityGroup {group} showActions={false} />
                {/each}
            {/if}
        </div>
    </div>
</div>

<!-- Manual Entry Modal -->
<ManualEntryModal
    bind:isOpen={showManualModal}
    on:created={handleManualEntryCreated}
    on:close={() => (showManualModal = false)}
/>
