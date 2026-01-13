<script lang="ts">
    import { onMount } from "svelte";
    import { fetchApi } from "$lib/api";
    import { format, subDays, parseISO } from "date-fns";
    import BlockList from "$lib/components/BlockList.svelte";

    // State
    let blocks: any[] = [];
    let loading = false;
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
        } catch (e) {
            console.error(e);
            alert("Failed to load history");
        } finally {
            loading = false;
        }
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
    </header>

    <!-- Summary Cards -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div class="bg-white p-6 rounded-xl border border-slate-200 shadow-sm">
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
        <div class="bg-white p-6 rounded-xl border border-slate-200 shadow-sm">
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
            class="bg-white p-6 rounded-xl border border-slate-200 shadow-sm flex items-center justify-center"
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
    <div class="bg-white rounded-xl border border-slate-200 shadow-sm">
        {#if loading}
            <div class="p-12 text-center text-slate-500">
                Loading history...
            </div>
        {:else}
            <BlockList {blocks} />
        {/if}
    </div>
</div>
