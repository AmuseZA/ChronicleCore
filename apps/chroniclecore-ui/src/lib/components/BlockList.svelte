<script lang="ts">
    import { format, parseISO } from "date-fns";
    import { suggestions } from "$lib/stores/suggestions";
    export let blocks: any[] = [];

    function formatTime(iso: string) {
        return format(parseISO(iso), "HH:mm");
    }

    function getDuration(mins: number) {
        const h = Math.floor(mins / 60);
        const m = mins % 60;
        if (h > 0) return `${h}h ${m}m`;
        return `${m}m`;
    }
</script>

<div
    class="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden"
>
    <!-- Header -->
    <div
        class="px-6 py-4 border-b border-slate-200 flex justify-between items-center bg-slate-50/50"
    >
        <div class="flex items-center gap-2">
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
                    d="M4 6h16M4 10h16M4 14h16M4 18h16"
                />
            </svg>
            <h3 class="text-sm font-semibold text-slate-900">Activity Log</h3>
        </div>
        <span
            class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-slate-100 text-slate-600"
        >
            {blocks.length} items
        </span>
    </div>

    <table class="w-full text-left border-collapse">
        <thead>
            <tr>
                <th
                    class="px-6 py-3 bg-slate-50 text-xs font-semibold text-slate-500 uppercase tracking-wider border-b border-slate-200 w-32"
                    >Time</th
                >
                <th
                    class="px-6 py-3 bg-slate-50 text-xs font-semibold text-slate-500 uppercase tracking-wider border-b border-slate-200 w-24"
                    >Dur</th
                >
                <th
                    class="px-6 py-3 bg-slate-50 text-xs font-semibold text-slate-500 uppercase tracking-wider border-b border-slate-200"
                    >Activity</th
                >
                <th
                    class="px-6 py-3 bg-slate-50 text-xs font-semibold text-slate-500 uppercase tracking-wider border-b border-slate-200 w-40"
                    >Client / Task</th
                >
                <th
                    class="px-6 py-3 bg-slate-50 text-xs font-semibold text-slate-500 uppercase tracking-wider border-b border-slate-200 w-32 text-right"
                    >Status</th
                >
            </tr>
        </thead>
        <tbody class="divide-y divide-slate-100">
            {#each blocks as block}
                <tr class="hover:bg-blue-50/30 group transition-colors">
                    <!-- Time -->
                    <td
                        class="px-6 py-4 text-sm text-slate-500 tabular-nums whitespace-nowrap"
                    >
                        <div class="flex flex-col">
                            <span class="text-slate-900 font-medium"
                                >{formatTime(block.ts_start)}</span
                            >
                            <span class="text-xs text-slate-400"
                                >to {formatTime(block.ts_end)}</span
                            >
                        </div>
                    </td>

                    <!-- Duration -->
                    <td class="px-6 py-4">
                        <span
                            class="inline-flex items-center px-2 py-1 rounded-md text-xs font-medium bg-slate-100 text-slate-600 border border-slate-200 cursor-default"
                        >
                            {getDuration(block.duration_minutes)}
                        </span>
                    </td>

                    <!-- Activity -->
                    <td class="px-6 py-4">
                        <div class="flex items-start gap-3">
                            <div
                                class="mt-1 p-1.5 rounded-md bg-slate-100 text-slate-500 shrink-0"
                            >
                                <svg
                                    class="w-4 h-4"
                                    fill="none"
                                    viewBox="0 0 24 24"
                                    stroke="currentColor"
                                >
                                    {#if block.primary_app_name
                                        .toLowerCase()
                                        .includes("excel")}
                                        <path
                                            stroke-linecap="round"
                                            stroke-linejoin="round"
                                            stroke-width="2"
                                            d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                                        />
                                    {:else if block.primary_app_name
                                        .toLowerCase()
                                        .includes("chrome") || block.primary_app_name
                                            .toLowerCase()
                                            .includes("edge")}
                                        <path
                                            stroke-linecap="round"
                                            stroke-linejoin="round"
                                            stroke-width="2"
                                            d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9"
                                        />
                                    {:else}
                                        <path
                                            stroke-linecap="round"
                                            stroke-linejoin="round"
                                            stroke-width="2"
                                            d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
                                        />
                                    {/if}
                                </svg>
                            </div>
                            <div>
                                <div
                                    class="text-sm font-medium text-slate-900 line-clamp-1"
                                    title={block.title_summary}
                                >
                                    {block.title_summary ||
                                        block.primary_app_name}
                                </div>
                                {#if block.title_summary && block.title_summary !== block.primary_app_name}
                                    <div class="text-xs text-slate-500 mt-0.5">
                                        {block.primary_app_name}
                                    </div>
                                {/if}
                            </div>
                        </div>
                    </td>

                    <!-- Profile -->
                    <td class="px-6 py-4">
                        {#if block.client_name}
                            <div class="flex flex-col items-start gap-1">
                                <span
                                    class="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold bg-blue-50 text-blue-700 border border-blue-100"
                                >
                                    {block.client_name}
                                </span>
                                {#if block.service_name}
                                    <span
                                        class="text-[11px] text-slate-400 pl-1"
                                        >{block.service_name}</span
                                    >
                                {/if}
                            </div>
                        {:else}
                            <span class="text-sm text-slate-400 italic"
                                >Unassigned</span
                            >
                        {/if}
                    </td>

                    <!-- Status -->
                    <td class="px-6 py-4 text-right">
                        {#if block.confidence === "LOW" && !block.profile_id}
                            <div class="flex flex-col items-end gap-1">
                                <span
                                    class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-amber-50 text-amber-700 border border-amber-200 shadow-sm animate-pulse"
                                >
                                    Needs Review
                                </span>
                                {#if $suggestions[block.block_id]?.length > 0}
                                    <span
                                        class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-medium bg-emerald-50 text-emerald-700 border border-emerald-100"
                                    >
                                        <svg
                                            class="w-3 h-3"
                                            fill="none"
                                            viewBox="0 0 24 24"
                                            stroke="currentColor"
                                        >
                                            <path
                                                stroke-linecap="round"
                                                stroke-linejoin="round"
                                                stroke-width="2"
                                                d="M13 10V3L4 14h7v7l9-11h-7z"
                                            />
                                        </svg>
                                        Suggestion
                                    </span>
                                {/if}
                            </div>
                        {:else if block.locked}
                            <span
                                class="inline-flex items-center gap-1 text-xs text-slate-400"
                            >
                                <svg
                                    class="w-3 h-3"
                                    fill="none"
                                    viewBox="0 0 24 24"
                                    stroke="currentColor"
                                >
                                    <path
                                        stroke-linecap="round"
                                        stroke-linejoin="round"
                                        stroke-width="2"
                                        d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
                                    />
                                </svg>
                                Locked
                            </span>
                        {:else}
                            <span
                                class="inline-flex items-center gap-1 text-xs text-green-600 font-medium"
                            >
                                <svg
                                    class="w-3 h-3"
                                    fill="none"
                                    viewBox="0 0 24 24"
                                    stroke="currentColor"
                                >
                                    <path
                                        stroke-linecap="round"
                                        stroke-linejoin="round"
                                        stroke-width="2"
                                        d="M5 13l4 4L19 7"
                                    />
                                </svg>
                                Logged
                            </span>
                        {/if}
                    </td>
                </tr>
            {:else}
                <tr>
                    <td colspan="5" class="px-6 py-16 text-center">
                        <div
                            class="mx-auto w-12 h-12 bg-slate-100 rounded-full flex items-center justify-center mb-3 text-slate-400"
                        >
                            <svg
                                class="w-6 h-6"
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
                        </div>
                        <h3 class="text-sm font-medium text-slate-900">
                            No activity yet
                        </h3>
                        <p class="text-sm text-slate-500 mt-1">
                            Start tracking to see your work blocks here.
                        </p>
                    </td>
                </tr>
            {/each}
        </tbody>
    </table>
</div>
