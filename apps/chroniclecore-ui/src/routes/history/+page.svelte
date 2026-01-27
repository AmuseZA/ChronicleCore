<script lang="ts">
    import { onMount } from "svelte";
    import { fetchApi } from "$lib/api";
    import { format, subDays, parseISO } from "date-fns";
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

    // State
    let blocks: any[] = [];
    let activityGroups: ActivityGroupType[] = [];
    let loading = false;
    let showManualModal = false;
    let dateRange = {
        start: format(subDays(new Date(), 7), "yyyy-MM-dd"),
        end: format(new Date(), "yyyy-MM-dd"),
    };

    // Computed stats
    $: totalDuration = blocks.reduce(
        (acc, b) => acc + (b.duration_minutes || 0),
        0,
    );
    $: totalHours = (totalDuration / 60).toFixed(1);

    // Group blocks by profile and contiguous time (gaps < 30 minutes)
    function groupBlocks(blocks: any[]): ActivityGroupType[] {
        if (!blocks.length) return [];

        const groups: ActivityGroupType[] = [];
        let currentGroup: ActivityGroupType | null = null;

        // Sort blocks by start time (oldest first for grouping)
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

    async function loadHistory() {
        loading = true;
        try {
            // Construct query params
            const params = new URLSearchParams({
                start: new Date(dateRange.start).toISOString(),
                end: new Date(dateRange.end + "T23:59:59").toISOString(), // End of day
            });

            const res = await fetchApi(`/blocks?${params.toString()}`);
            blocks = res || [];
            activityGroups = groupBlocks(blocks);
        } catch (e) {
            console.error(e);
            alert("Failed to load history");
        } finally {
            loading = false;
        }
    }

    function handleManualEntryCreated() {
        loadHistory();
    }

    function exportCsv() {
        if (!blocks.length) return alert("No data to export");

        const headers = [
            "Date",
            "Start Time",
            "End Time",
            "Duration (min)",
            "Client",
            "Service",
            "Activity",
            "Notes",
        ];
        const rows = blocks.map((b) => [
            format(parseISO(b.ts_start), "yyyy-MM-dd"),
            format(parseISO(b.ts_start), "HH:mm"),
            format(parseISO(b.ts_end), "HH:mm"),
            b.duration_minutes,
            b.client_name || "Unassigned",
            b.service_name || "",
            b.title_summary || b.primary_app_name,
            `"${(b.notes || "").replace(/"/g, '""')}"`, // Escape quotes
        ]);

        const csvContent = [
            headers.join(","),
            ...rows.map((r) => r.join(",")),
        ].join("\n");
        const blob = new Blob([csvContent], {
            type: "text/csv;charset=utf-8;",
        });
        const url = URL.createObjectURL(blob);
        const link = document.createElement("a");
        link.setAttribute("href", url);
        link.setAttribute(
            "download",
            `chronicle_export_${dateRange.start}_${dateRange.end}.csv`,
        );
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
    }

    onMount(loadHistory);
</script>

<div class="max-w-7xl mx-auto space-y-6">
    <header
        class="flex flex-col md:flex-row md:items-center justify-between gap-4"
    >
        <div>
            <h1 class="text-2xl font-bold text-slate-900">
                History & Calendar
            </h1>
            <p class="text-slate-500">
                Review past activity and generate reports.
            </p>
        </div>

        <div class="flex items-center gap-3">
            <button
                on:click={() => (showManualModal = true)}
                class="flex items-center gap-2 bg-white hover:bg-slate-50 text-slate-700 px-4 py-2 rounded-lg font-medium text-sm transition-colors border border-slate-200 shadow-sm"
            >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                </svg>
                Add Entry
            </button>
            <div
                class="flex items-center gap-2 bg-white p-2 rounded-lg border border-slate-200 shadow-sm"
            >
                <input
                    type="date"
                    bind:value={dateRange.start}
                    class="border-none text-sm text-slate-700 focus:ring-0 p-1"
                />
                <span class="text-slate-400">to</span>
                <input
                    type="date"
                    bind:value={dateRange.end}
                    class="border-none text-sm text-slate-700 focus:ring-0 p-1"
                />
                <button
                    on:click={loadHistory}
                    class="ml-2 bg-slate-900 text-white px-4 py-1.5 rounded-md text-sm font-medium hover:bg-slate-800"
                >
                    Apply
                </button>
            </div>
        </div>
    </header>

    <!-- Summary Cards -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div class="bg-white dark:bg-slate-800 p-6 rounded-xl border border-slate-200 dark:border-slate-700 shadow-sm">
            <h3
                class="text-xs font-semibold text-slate-500 uppercase tracking-wide"
            >
                Total Hours
            </h3>
            <div class="mt-2 flex items-baseline gap-2">
                <span class="text-3xl font-bold text-slate-900"
                    >{totalHours}</span
                >
                <span class="text-sm text-slate-500">hrs</span>
            </div>
        </div>
        <div class="bg-white dark:bg-slate-800 p-6 rounded-xl border border-slate-200 dark:border-slate-700 shadow-sm">
            <h3
                class="text-xs font-semibold text-slate-500 uppercase tracking-wide"
            >
                Block Count
            </h3>
            <div class="mt-2 flex items-baseline gap-2">
                <span class="text-3xl font-bold text-slate-900"
                    >{blocks.length}</span
                >
                <span class="text-sm text-slate-500">blocks</span>
            </div>
        </div>
        <div
            class="bg-white dark:bg-slate-800 p-6 rounded-xl border border-slate-200 dark:border-slate-700 shadow-sm flex items-center justify-center"
        >
            <button
                on:click={exportCsv}
                class="text-blue-600 font-medium hover:underline flex items-center gap-2"
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
                        d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
                    /></svg
                >
                Export CSV
            </button>
        </div>
    </div>

    <!-- Content -->
    <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 shadow-sm overflow-hidden">
        <div class="px-6 py-4 border-b border-slate-100 flex justify-between items-center bg-slate-50/50">
            <h2 class="text-lg font-bold text-slate-900">Activity Groups</h2>
            <span class="text-sm text-slate-500">{activityGroups.length} groups</span>
        </div>
        <div class="p-4 space-y-3">
            {#if loading}
                <div class="text-center py-8 text-slate-500">
                    Loading history...
                </div>
            {:else if activityGroups.length === 0}
                <div class="text-center py-12">
                    <div class="mx-auto w-12 h-12 bg-slate-100 rounded-full flex items-center justify-center mb-3 text-slate-400">
                        <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                    </div>
                    <h3 class="text-sm font-medium text-slate-900">No activity found</h3>
                    <p class="text-sm text-slate-500 mt-1">Try adjusting the date range.</p>
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
