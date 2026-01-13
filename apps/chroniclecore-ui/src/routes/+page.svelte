<script lang="ts">
    import { onMount } from "svelte";
    import { fetchApi } from "$lib/api";
    import TrackingControl from "$lib/components/TrackingControl.svelte";
    import BlockList from "$lib/components/BlockList.svelte";

    let stats = {
        totalHours: 0,
        productivityScore: 0,
        billableHours: 0,
        reviewCount: 0,
    };

    let blocks: any[] = [];
    let loading = true;
    let error: string | null = null;

    async function loadData() {
        try {
            loading = true;
            error = null;
            const res = await fetchApi("/blocks");
            blocks = res || [];

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
            };
        } catch (err: any) {
            console.error(err);
            error = err.message || "Failed to load blocks";
        } finally {
            loading = false;
        }
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
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
        <!-- Total Time -->
        <div
            class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm relative overflow-hidden group"
        >
            <div
                class="absolute top-0 right-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity"
            >
                <svg
                    class="w-16 h-16 text-blue-600"
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
                Total Time
            </div>
            <div class="text-3xl font-bold text-slate-900">
                {stats.totalHours.toFixed(1)}
                <span class="text-lg text-slate-400 font-normal">hrs</span>
            </div>
            <div
                class="mt-4 h-1.5 w-full bg-slate-100 rounded-full overflow-hidden"
            >
                <div
                    class="h-full bg-blue-500 rounded-full"
                    style="width: 100%"
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

        <!-- Billable Ratio -->
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
        </div>
        <BlockList {blocks} />
    </div>
</div>
