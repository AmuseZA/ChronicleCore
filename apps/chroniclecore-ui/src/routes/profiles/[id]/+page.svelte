<script lang="ts">
    import { onMount } from "svelte";
    import { page } from "$app/stores";
    import { goto } from "$app/navigation";
    import { fetchApi } from "$lib/api";
    import { settings } from "$lib/stores/settings";
    import {
        format,
        parseISO,
        startOfMonth,
        endOfMonth,
        subMonths,
        addMonths,
    } from "date-fns";
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
        billableMinutes: number;
    }

    interface SubGroup {
        key: string;
        displayName: string;
        blocks: Block[];
        totalMinutes: number;
        billableMinutes: number;
    }

    // Expanded state for sub-groups within dates
    let expandedSubGroups: Set<string> = new Set();

    let stats: ProfileStats | null = null;
    let loading = true;
    let error: string | null = null;
    let processingBlockId: number | null = null;
    let expandedDates: Set<string> = new Set();

    // View mode: 'lifetime' | 'monthly' | 'custom'
    let viewMode: "lifetime" | "monthly" | "custom" = "monthly";

    // Month selector
    let selectedMonth: Date = new Date();

    // Custom date range
    let customStartDate: string = format(new Date(), "yyyy-MM-dd");
    let customEndDate: string = format(new Date(), "yyyy-MM-dd");

    $: profileId = $page.params.id;

    // Group blocks by date
    $: groupedBlocks = groupBlocksByDate(stats?.recent_blocks || []);

    // Extract a grouping key from block title/app for sub-grouping
    function getSubGroupKey(block: Block): string {
        const title = block.title_summary || block.primary_app_name;

        // Extract meaningful patterns from title
        // e.g., "Outlook - accounts@company.com" -> "Outlook - accounts@"
        // e.g., "Chrome - admin@website.com" -> "Chrome - admin@"

        // Pattern: Look for email-like patterns (user@) or app - prefix patterns
        const emailMatch = title.match(/^(.+?\s*-\s*)([a-zA-Z0-9._-]+@)/);
        if (emailMatch) {
            return `${emailMatch[1]}${emailMatch[2]}`;
        }

        // Pattern: App Name - Category/Context
        const appContextMatch = title.match(/^([^-]+\s*-\s*[^-]+)/);
        if (appContextMatch) {
            return appContextMatch[1].trim();
        }

        // Default: group by app name
        return block.primary_app_name;
    }

    function groupBlocksIntoSubGroups(blocks: Block[]): SubGroup[] {
        const groups = new Map<string, SubGroup>();

        for (const block of blocks) {
            const key = getSubGroupKey(block);

            if (!groups.has(key)) {
                groups.set(key, {
                    key,
                    displayName: key,
                    blocks: [],
                    totalMinutes: 0,
                    billableMinutes: 0,
                });
            }

            const group = groups.get(key)!;
            group.blocks.push(block);
            group.totalMinutes += block.duration_minutes;
            if (block.billable) {
                group.billableMinutes += block.duration_minutes;
            }
        }

        // Sort sub-groups by billable time descending
        return Array.from(groups.values()).sort((a, b) => b.billableMinutes - a.billableMinutes);
    }

    function toggleSubGroup(dateKey: string, subGroupKey: string) {
        const fullKey = `${dateKey}:${subGroupKey}`;
        if (expandedSubGroups.has(fullKey)) {
            expandedSubGroups.delete(fullKey);
        } else {
            expandedSubGroups.add(fullKey);
        }
        expandedSubGroups = expandedSubGroups;
    }

    function groupBlocksByDate(blocks: Block[]): GroupedByDate[] {
        const groups = new Map<string, GroupedByDate>();

        for (const block of blocks) {
            const date = format(parseISO(block.ts_start), "yyyy-MM-dd");

            if (!groups.has(date)) {
                groups.set(date, {
                    date,
                    dateFormatted: format(
                        parseISO(block.ts_start),
                        "EEEE, MMMM d, yyyy",
                    ),
                    blocks: [],
                    totalMinutes: 0,
                    billableMinutes: 0,
                });
            }

            const group = groups.get(date)!;
            group.blocks.push(block);
            group.totalMinutes += block.duration_minutes;
            if (block.billable) {
                group.billableMinutes += block.duration_minutes;
            }
        }

        // Sort by date ascending for reports (chronological order)
        return Array.from(groups.values()).sort((a, b) =>
            a.date.localeCompare(b.date),
        );
    }

    async function loadStats() {
        loading = true;
        error = null;
        try {
            await settings.detectLocale();

            let url = `/profiles/${profileId}/stats?include_blocks=true`;

            if (viewMode === "monthly") {
                const start = format(startOfMonth(selectedMonth), "yyyy-MM-dd");
                const end = format(endOfMonth(selectedMonth), "yyyy-MM-dd");
                url += `&start_date=${start}&end_date=${end}`;
            } else if (viewMode === "custom") {
                url += `&start_date=${customStartDate}&end_date=${customEndDate}`;
            }

            stats = await fetchApi(url);

            // Default to collapsed date groups
            expandedDates = new Set();
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

    function setViewMode(mode: "lifetime" | "monthly" | "custom") {
        viewMode = mode;
        if (mode !== "custom") {
            loadStats();
        }
    }

    function applyCustomRange() {
        if (customStartDate && customEndDate) {
            loadStats();
        }
    }

    function copyReportToClipboard() {
        if (!stats) return;

        const ratePerHour = stats.rate_amount;
        const currency = stats.currency_code;
        let report = `TIMESHEET REPORT\n`;
        report += `================\n\n`;
        report += `Client: ${stats.client_name}\n`;
        if (stats.project_name) report += `Project: ${stats.project_name}\n`;
        report += `Service: ${stats.service_name}\n`;
        report += `Rate: ${formatCurrency(ratePerHour, currency)}/hr\n`;
        report += `Period: ${format(parseISO(customStartDate), "MMMM d, yyyy")} - ${format(parseISO(customEndDate), "MMMM d, yyyy")}\n\n`;
        report += `DAILY BREAKDOWN\n`;
        report += `---------------\n\n`;

        let grandTotalMinutes = 0;
        let grandTotalBillable = 0;

        for (const group of groupedBlocks) {
            const hours = group.billableMinutes / 60;
            const amount = hours * ratePerHour;
            grandTotalMinutes += group.billableMinutes;
            grandTotalBillable += amount;

            report += `${group.dateFormatted}\n`;
            report += `  Time: ${hours.toFixed(2)} hours\n`;
            report += `  Amount: ${formatCurrency(amount, currency)}\n`;

            // Group tasks by sub-group for easier invoicing
            const subGroups = groupBlocksIntoSubGroups(group.blocks);
            for (const subGroup of subGroups) {
                if (subGroup.billableMinutes > 0) {
                    const subHours = subGroup.billableMinutes / 60;
                    const subAmount = subHours * ratePerHour;
                    report += `    ${subGroup.displayName}\n`;
                    report += `      ${subHours.toFixed(2)}h - ${formatCurrency(subAmount, currency)}\n`;
                }
            }
            report += `\n`;
        }

        report += `TOTALS\n`;
        report += `------\n`;
        report += `Total Hours: ${(grandTotalMinutes / 60).toFixed(2)}\n`;
        report += `Total Amount: ${formatCurrency(grandTotalBillable, currency)}\n`;

        navigator.clipboard.writeText(report);
        alert("Report copied to clipboard!");
    }

    function downloadReportCSV() {
        if (!stats) return;

        const ratePerHour = stats.rate_amount;
        const currency = stats.currency_code;

        let csv = "Date,Category,Task,Duration (hours),Rate,Amount,Currency\n";

        for (const group of groupedBlocks) {
            for (const block of group.blocks) {
                if (block.billable) {
                    const hours = block.duration_minutes / 60;
                    const amount = hours * ratePerHour;
                    const category = getSubGroupKey(block).replace(/"/g, '""');
                    const task = (
                        block.title_summary || block.primary_app_name
                    ).replace(/"/g, '""');
                    csv += `"${group.date}","${category}","${task}",${hours.toFixed(4)},${ratePerHour},${amount.toFixed(2)},${currency}\n`;
                }
            }
        }

        const blob = new Blob([csv], { type: "text/csv" });
        const url = URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = url;
        a.download = `timesheet_${stats.client_name.replace(/\s+/g, "_")}_${customStartDate}_${customEndDate}.csv`;
        a.click();
        URL.revokeObjectURL(url);
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
        if (
            !confirm(
                "Are you sure you want to unassign this block? It will appear in the Review queue.",
            )
        ) {
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
            on:click={() => goto("/profiles")}
            class="flex items-center gap-2 text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-300 transition-colors"
        >
            <svg
                class="w-5 h-5"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
            >
                <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M15 19l-7-7 7-7"
                />
            </svg>
            Back to Profiles
        </button>
    </div>

    {#if loading}
        <div
            class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-12 text-center text-slate-500 dark:text-slate-400"
        >
            Loading profile details...
        </div>
    {:else if error}
        <div
            class="bg-red-50 text-red-700 p-4 rounded-xl border border-red-100"
        >
            {error}
        </div>
    {:else if stats}
        <!-- Profile Header Card -->
        <div
            class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 shadow-sm p-6"
        >
            <div class="flex items-start justify-between">
                <div class="flex items-center gap-4">
                    <div
                        class="w-14 h-14 rounded-full bg-blue-100 text-blue-700 flex items-center justify-center font-bold text-lg"
                    >
                        {stats.client_name.slice(0, 2).toUpperCase()}
                    </div>
                    <div>
                        <h1
                            class="text-2xl font-bold text-slate-900 dark:text-slate-100"
                        >
                            {stats.client_name}
                        </h1>
                        {#if stats.project_name}
                            <div class="text-slate-500 dark:text-slate-400">
                                {stats.project_name}
                            </div>
                        {/if}
                        <div class="flex items-center gap-3 mt-1">
                            <span
                                class="inline-flex px-2.5 py-1 rounded-md bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300 text-xs font-medium border border-slate-200 dark:border-slate-600"
                            >
                                {stats.service_name}
                            </span>
                            <span
                                class="text-sm text-slate-500 dark:text-slate-400"
                            >
                                {stats.rate_name} - {formatCurrency(
                                    stats.rate_amount,
                                    stats.currency_code,
                                )}/hr
                            </span>
                        </div>
                    </div>
                </div>

                <!-- View Mode Toggle -->
                <div
                    class="flex items-center gap-1 bg-slate-100 dark:bg-slate-700 rounded-lg p-1"
                >
                    <button
                        on:click={() => setViewMode("monthly")}
                        class="px-4 py-2 text-sm font-medium rounded-md transition-colors {viewMode ===
                        'monthly'
                            ? 'bg-white dark:bg-slate-600 text-slate-900 dark:text-slate-100 shadow-sm'
                            : 'text-slate-600 dark:text-slate-300 hover:text-slate-900 dark:hover:text-slate-100'}"
                    >
                        Monthly
                    </button>
                    <button
                        on:click={() => setViewMode("custom")}
                        class="px-4 py-2 text-sm font-medium rounded-md transition-colors {viewMode ===
                        'custom'
                            ? 'bg-white dark:bg-slate-600 text-slate-900 dark:text-slate-100 shadow-sm'
                            : 'text-slate-600 dark:text-slate-300 hover:text-slate-900 dark:hover:text-slate-100'}"
                    >
                        Custom Range
                    </button>
                    <button
                        on:click={() => setViewMode("lifetime")}
                        class="px-4 py-2 text-sm font-medium rounded-md transition-colors {viewMode ===
                        'lifetime'
                            ? 'bg-white dark:bg-slate-600 text-slate-900 dark:text-slate-100 shadow-sm'
                            : 'text-slate-600 dark:text-slate-300 hover:text-slate-900 dark:hover:text-slate-100'}"
                    >
                        Lifetime
                    </button>
                </div>
            </div>

            <!-- Month Selector (only for monthly view) -->
            {#if viewMode === "monthly"}
                <div
                    class="mt-6 pt-6 border-t border-slate-100 dark:border-slate-700 flex items-center justify-between"
                >
                    <div class="flex items-center gap-3">
                        <button
                            on:click={prevMonth}
                            class="w-9 h-9 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 hover:bg-slate-50 dark:hover:bg-slate-600 flex items-center justify-center transition-colors"
                        >
                            <svg
                                class="w-5 h-5 text-slate-500 dark:text-slate-400"
                                fill="none"
                                stroke="currentColor"
                                viewBox="0 0 24 24"
                            >
                                <path
                                    stroke-linecap="round"
                                    stroke-linejoin="round"
                                    stroke-width="2"
                                    d="M15 19l-7-7 7-7"
                                />
                            </svg>
                        </button>
                        <div
                            class="text-lg font-semibold text-slate-900 dark:text-slate-100 min-w-[180px] text-center"
                        >
                            {format(selectedMonth, "MMMM yyyy")}
                        </div>
                        <button
                            on:click={nextMonth}
                            class="w-9 h-9 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 hover:bg-slate-50 dark:hover:bg-slate-600 flex items-center justify-center transition-colors"
                        >
                            <svg
                                class="w-5 h-5 text-slate-500 dark:text-slate-400"
                                fill="none"
                                stroke="currentColor"
                                viewBox="0 0 24 24"
                            >
                                <path
                                    stroke-linecap="round"
                                    stroke-linejoin="round"
                                    stroke-width="2"
                                    d="M9 5l7 7-7 7"
                                />
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

            <!-- Custom Date Range Selector -->
            {#if viewMode === "custom"}
                <div
                    class="mt-6 pt-6 border-t border-slate-100 dark:border-slate-700"
                >
                    <div class="flex flex-wrap items-end gap-4">
                        <div class="flex-1 min-w-[200px]">
                            <label
                                for="custom-start-date"
                                class="block text-xs font-medium text-slate-500 dark:text-slate-400 mb-1"
                                >Start Date</label
                            >
                            <div
                                class="relative cursor-pointer"
                                on:click={(e) => {
                                    const input = e.currentTarget.querySelector('input');
                                    if (input) input.showPicker?.();
                                }}
                                on:keydown={(e) => {
                                    if (e.key === 'Enter' || e.key === ' ') {
                                        const input = e.currentTarget.querySelector('input');
                                        if (input) input.showPicker?.();
                                    }
                                }}
                                role="button"
                                tabindex="-1"
                            >
                                <input
                                    id="custom-start-date"
                                    type="date"
                                    bind:value={customStartDate}
                                    class="w-full px-3 py-2 border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent cursor-pointer"
                                />
                            </div>
                        </div>
                        <div class="flex-1 min-w-[200px]">
                            <label
                                for="custom-end-date"
                                class="block text-xs font-medium text-slate-500 dark:text-slate-400 mb-1"
                                >End Date</label
                            >
                            <div
                                class="relative cursor-pointer"
                                on:click={(e) => {
                                    const input = e.currentTarget.querySelector('input');
                                    if (input) input.showPicker?.();
                                }}
                                on:keydown={(e) => {
                                    if (e.key === 'Enter' || e.key === ' ') {
                                        const input = e.currentTarget.querySelector('input');
                                        if (input) input.showPicker?.();
                                    }
                                }}
                                role="button"
                                tabindex="-1"
                            >
                                <input
                                    id="custom-end-date"
                                    type="date"
                                    bind:value={customEndDate}
                                    class="w-full px-3 py-2 border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent cursor-pointer"
                                />
                            </div>
                        </div>
                        <button
                            on:click={applyCustomRange}
                            class="px-5 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
                        >
                            Apply Range
                        </button>
                    </div>

                    <!-- Report Actions (show after data is loaded) -->
                    {#if stats && groupedBlocks.length > 0}
                        <div
                            class="mt-4 pt-4 border-t border-slate-100 dark:border-slate-700 flex items-center gap-3"
                        >
                            <span
                                class="text-sm text-slate-500 dark:text-slate-400"
                                >Export Report:</span
                            >
                            <button
                                on:click={copyReportToClipboard}
                                class="inline-flex items-center gap-2 px-4 py-2 bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-300 text-sm font-medium rounded-lg hover:bg-slate-200 dark:hover:bg-slate-600 transition-colors"
                            >
                                <svg
                                    class="w-4 h-4"
                                    fill="none"
                                    stroke="currentColor"
                                    viewBox="0 0 24 24"
                                >
                                    <path
                                        stroke-linecap="round"
                                        stroke-linejoin="round"
                                        stroke-width="2"
                                        d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3"
                                    />
                                </svg>
                                Copy as Text
                            </button>
                            <button
                                on:click={downloadReportCSV}
                                class="inline-flex items-center gap-2 px-4 py-2 bg-green-600 text-white text-sm font-medium rounded-lg hover:bg-green-700 transition-colors"
                            >
                                <svg
                                    class="w-4 h-4"
                                    fill="none"
                                    stroke="currentColor"
                                    viewBox="0 0 24 24"
                                >
                                    <path
                                        stroke-linecap="round"
                                        stroke-linejoin="round"
                                        stroke-width="2"
                                        d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                                    />
                                </svg>
                                Download CSV
                            </button>
                        </div>
                    {/if}
                </div>
            {/if}
        </div>

        <!-- Stats Cards -->
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <!-- Total Time -->
            <div
                class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 shadow-sm p-5"
            >
                <div
                    class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide"
                >
                    Total Time
                </div>
                <div
                    class="mt-2 text-2xl font-bold text-slate-900 dark:text-slate-100 tabular-nums"
                >
                    {stats.total_hours.toFixed(1)}h
                </div>
                <div class="text-sm text-slate-500 dark:text-slate-400 mt-1">
                    {stats.total_blocks} blocks
                </div>
            </div>

            <!-- Billable Time -->
            <div
                class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 shadow-sm p-5"
            >
                <div
                    class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide"
                >
                    Billable Time
                </div>
                <div
                    class="mt-2 text-2xl font-bold text-slate-900 dark:text-slate-100 tabular-nums"
                >
                    {stats.billable_hours.toFixed(1)}h
                </div>
                <div class="text-sm text-slate-500 dark:text-slate-400 mt-1">
                    {(
                        (stats.billable_hours / (stats.total_hours || 1)) *
                        100
                    ).toFixed(0)}% of total
                </div>
            </div>

            <!-- Estimated Billable -->
            <div
                class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 shadow-sm p-5"
            >
                <div
                    class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide"
                >
                    Est. Billable
                </div>
                <div
                    class="mt-2 text-2xl font-bold text-green-600 tabular-nums"
                >
                    {formatCurrency(
                        stats.estimated_billable,
                        stats.currency_code,
                    )}
                </div>
                <div class="text-sm text-slate-500 dark:text-slate-400 mt-1">
                    @ {formatCurrency(
                        stats.rate_amount,
                        stats.currency_code,
                    )}/hr
                </div>
            </div>

            <!-- Locked / Invoiced -->
            <div
                class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 shadow-sm p-5"
            >
                <div
                    class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide"
                >
                    Locked (Ready)
                </div>
                <div
                    class="mt-2 text-2xl font-bold text-slate-900 dark:text-slate-100 tabular-nums"
                >
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
                <h2
                    class="text-lg font-semibold text-slate-900 dark:text-slate-100"
                >
                    {#if viewMode === "lifetime"}
                        Recent Activity
                    {:else if viewMode === "custom"}
                        Activity: {format(parseISO(customStartDate), "MMM d")} -
                        {format(parseISO(customEndDate), "MMM d, yyyy")}
                    {:else}
                        Activity for {format(selectedMonth, "MMMM yyyy")}
                    {/if}
                </h2>
                <span class="text-sm text-slate-500 dark:text-slate-400">
                    {groupedBlocks.length} day{groupedBlocks.length !== 1
                        ? "s"
                        : ""} with activity
                </span>
            </div>

            {#if groupedBlocks.length > 0}
                {#each groupedBlocks as dateGroup (dateGroup.date)}
                    <div
                        class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 shadow-sm overflow-hidden"
                    >
                        <!-- Date Header -->
                        <button
                            on:click={() => toggleDateGroup(dateGroup.date)}
                            class="w-full px-6 py-4 flex items-center justify-between hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors"
                        >
                            <div class="flex items-center gap-4">
                                <div class="flex items-center gap-2">
                                    <svg
                                        class="w-5 h-5 text-slate-400"
                                        fill="none"
                                        stroke="currentColor"
                                        viewBox="0 0 24 24"
                                    >
                                        <path
                                            stroke-linecap="round"
                                            stroke-linejoin="round"
                                            stroke-width="2"
                                            d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
                                        />
                                    </svg>
                                    <span class="font-semibold text-slate-900 dark:text-slate-100"
                                        >{dateGroup.dateFormatted}</span
                                    >
                                </div>
                                <span
                                    class="px-2 py-0.5 rounded-full bg-slate-100 dark:bg-slate-600 text-slate-600 dark:text-slate-300 text-xs font-medium"
                                >
                                    {dateGroup.blocks.length} block{dateGroup
                                        .blocks.length !== 1
                                        ? "s"
                                        : ""}
                                </span>
                            </div>
                            <div class="flex items-center gap-4">
                                <div class="text-right">
                                    <span
                                        class="font-medium text-slate-700 dark:text-slate-300 tabular-nums"
                                    >
                                        {(
                                            dateGroup.billableMinutes / 60
                                        ).toFixed(1)}h
                                    </span>
                                    {#if stats}
                                        <span
                                            class="ml-2 text-sm text-green-600 font-medium tabular-nums"
                                        >
                                            {formatCurrency(
                                                (dateGroup.billableMinutes /
                                                    60) *
                                                    stats.rate_amount,
                                                stats.currency_code,
                                            )}
                                        </span>
                                    {/if}
                                </div>
                                <svg
                                    class="w-5 h-5 text-slate-400 transition-transform {expandedDates.has(
                                        dateGroup.date,
                                    )
                                        ? 'rotate-180'
                                        : ''}"
                                    fill="none"
                                    stroke="currentColor"
                                    viewBox="0 0 24 24"
                                >
                                    <path
                                        stroke-linecap="round"
                                        stroke-linejoin="round"
                                        stroke-width="2"
                                        d="M19 9l-7 7-7-7"
                                    />
                                </svg>
                            </div>
                        </button>

                        <!-- Blocks List (Expanded) -->
                        {#if expandedDates.has(dateGroup.date)}
                            <div
                                class="border-t border-slate-100 dark:border-slate-700"
                            >
                                <!-- Sub-groups for custom view -->
                                {#if viewMode === "custom"}
                                    {@const subGroups = groupBlocksIntoSubGroups(dateGroup.blocks)}
                                    <div class="divide-y divide-slate-100 dark:divide-slate-700">
                                        {#each subGroups as subGroup (subGroup.key)}
                                            <!-- Sub-group header -->
                                            <div>
                                                <button
                                                    on:click={() => toggleSubGroup(dateGroup.date, subGroup.key)}
                                                    class="w-full px-6 py-3 flex items-center justify-between bg-slate-50 dark:bg-slate-750 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors"
                                                >
                                                    <div class="flex items-center gap-3">
                                                        <svg
                                                            class="w-4 h-4 text-slate-400 transition-transform {expandedSubGroups.has(`${dateGroup.date}:${subGroup.key}`) ? 'rotate-90' : ''}"
                                                            fill="none"
                                                            stroke="currentColor"
                                                            viewBox="0 0 24 24"
                                                        >
                                                            <path
                                                                stroke-linecap="round"
                                                                stroke-linejoin="round"
                                                                stroke-width="2"
                                                                d="M9 5l7 7-7 7"
                                                            />
                                                        </svg>
                                                        <span class="font-medium text-slate-700 dark:text-slate-300 text-sm">
                                                            {subGroup.displayName}
                                                        </span>
                                                        <span class="px-2 py-0.5 rounded-full bg-slate-200 dark:bg-slate-600 text-slate-600 dark:text-slate-300 text-xs font-medium">
                                                            {subGroup.blocks.length} block{subGroup.blocks.length !== 1 ? 's' : ''}
                                                        </span>
                                                    </div>
                                                    <div class="flex items-center gap-3">
                                                        <span class="text-sm font-medium text-slate-600 dark:text-slate-400 tabular-nums">
                                                            {(subGroup.billableMinutes / 60).toFixed(1)}h
                                                        </span>
                                                        {#if stats}
                                                            <span class="text-sm text-green-600 font-medium tabular-nums">
                                                                {formatCurrency(
                                                                    (subGroup.billableMinutes / 60) * stats.rate_amount,
                                                                    stats.currency_code
                                                                )}
                                                            </span>
                                                        {/if}
                                                    </div>
                                                </button>

                                                <!-- Sub-group blocks (expanded) -->
                                                {#if expandedSubGroups.has(`${dateGroup.date}:${subGroup.key}`)}
                                                    <div class="divide-y divide-slate-50 dark:divide-slate-700 bg-white dark:bg-slate-800">
                                                        {#each subGroup.blocks as block (block.block_id)}
                                                            <div
                                                                class="px-6 py-3 pl-12 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors flex items-start gap-4"
                                                            >
                                                                <!-- Time Column -->
                                                                <div class="w-20 shrink-0">
                                                                    <div class="text-sm font-medium text-slate-900 dark:text-slate-100">
                                                                        {format(parseISO(block.ts_start), "HH:mm")}
                                                                    </div>
                                                                    <div class="text-xs text-slate-500 dark:text-slate-400">
                                                                        to {format(parseISO(block.ts_end), "HH:mm")}
                                                                    </div>
                                                                </div>

                                                                <!-- Activity Info -->
                                                                <div class="flex-1 min-w-0">
                                                                    <div class="text-sm text-slate-900 dark:text-slate-100 break-words">
                                                                        {block.title_summary || block.primary_app_name}
                                                                    </div>
                                                                    <div class="flex items-center gap-2 mt-1">
                                                                        <span class="text-xs text-slate-500 dark:text-slate-400">{block.primary_app_name}</span>
                                                                        <span class="text-slate-300 dark:text-slate-600">•</span>
                                                                        <span class="text-xs font-medium text-slate-600 dark:text-slate-400 tabular-nums">{block.duration_minutes.toFixed(0)} min</span>
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
                                                                        <div class="w-32">
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
                                                {/if}
                                            </div>
                                        {/each}
                                    </div>
                                {:else}
                                    <!-- Standard flat list for monthly/lifetime views -->
                                    <div
                                        class="divide-y divide-slate-100 dark:divide-slate-700"
                                    >
                                        {#each dateGroup.blocks as block (block.block_id)}
                                            <div
                                                class="px-6 py-4 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors flex items-start gap-4"
                                            >
                                                <!-- Time Column -->
                                                <div class="w-24 shrink-0">
                                                    <div
                                                        class="text-sm font-medium text-slate-900 dark:text-slate-100"
                                                    >
                                                        {format(
                                                            parseISO(
                                                                block.ts_start,
                                                            ),
                                                            "HH:mm",
                                                        )}
                                                    </div>
                                                    <div
                                                        class="text-xs text-slate-500 dark:text-slate-400"
                                                    >
                                                        to {format(
                                                            parseISO(block.ts_end),
                                                            "HH:mm",
                                                        )}
                                                    </div>
                                                </div>

                                                <!-- Activity Info -->
                                                <div class="flex-1 min-w-0">
                                                    <div
                                                        class="text-sm text-slate-900 dark:text-slate-100 break-words"
                                                    >
                                                        {block.title_summary ||
                                                            block.primary_app_name}
                                                    </div>
                                                    <div
                                                        class="flex items-center gap-2 mt-1"
                                                    >
                                                        <span
                                                            class="text-xs text-slate-500 dark:text-slate-400"
                                                            >{block.primary_app_name}</span
                                                        >
                                                        <span
                                                            class="text-slate-300 dark:text-slate-600"
                                                            >•</span
                                                        >
                                                        <span
                                                            class="text-xs font-medium text-slate-600 dark:text-slate-400 tabular-nums"
                                                            >{block.duration_minutes.toFixed(
                                                                0,
                                                            )} min</span
                                                        >
                                                    </div>
                                                </div>

                                                <!-- Status Badges -->
                                                <div
                                                    class="flex items-center gap-2 shrink-0"
                                                >
                                                    <span
                                                        class="text-xs px-2 py-1 rounded-md border {getConfidenceStyle(
                                                            block.confidence,
                                                        )}"
                                                    >
                                                        {block.confidence}
                                                    </span>
                                                    {#if block.locked}
                                                        <span
                                                            class="text-xs px-2 py-1 rounded-md bg-green-50 text-green-700 border border-green-200"
                                                        >
                                                            Locked
                                                        </span>
                                                    {/if}
                                                    {#if !block.billable}
                                                        <span
                                                            class="text-xs px-2 py-1 rounded-md bg-slate-50 text-slate-600 border border-slate-200"
                                                        >
                                                            Non-billable
                                                        </span>
                                                    {/if}
                                                </div>

                                                <!-- Actions -->
                                                <div
                                                    class="flex items-center gap-2 shrink-0"
                                                >
                                                    {#if !block.locked}
                                                        <div class="w-36">
                                                            <ProfileSelector
                                                                placeholder="Change..."
                                                                value={stats?.profile_id}
                                                                disabled={processingBlockId ===
                                                                    block.block_id}
                                                                on:change={(e) =>
                                                                    reassignBlock(
                                                                        block.block_id,
                                                                        e.detail,
                                                                    )}
                                                            />
                                                        </div>
                                                        <button
                                                            on:click={() =>
                                                                unassignBlock(
                                                                    block.block_id,
                                                                )}
                                                            disabled={processingBlockId ===
                                                                block.block_id}
                                                            class="text-xs text-slate-400 hover:text-red-600 transition-colors disabled:opacity-50"
                                                            title="Remove from this profile"
                                                        >
                                                            <svg
                                                                class="w-4 h-4"
                                                                fill="none"
                                                                stroke="currentColor"
                                                                viewBox="0 0 24 24"
                                                            >
                                                                <path
                                                                    stroke-linecap="round"
                                                                    stroke-linejoin="round"
                                                                    stroke-width="2"
                                                                    d="M6 18L18 6M6 6l12 12"
                                                                />
                                                            </svg>
                                                        </button>
                                                    {:else}
                                                        <span
                                                            class="text-xs text-slate-400 italic"
                                                            >Locked</span
                                                        >
                                                    {/if}
                                                </div>
                                            </div>
                                        {/each}
                                    </div>
                                {/if}
                            </div>
                        {/if}
                    </div>
                {/each}
            {:else}
                <div
                    class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-12 text-center"
                >
                    <div
                        class="w-12 h-12 bg-slate-100 text-slate-400 rounded-full flex items-center justify-center mx-auto mb-3"
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
                                d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                            />
                        </svg>
                    </div>
                    <h3
                        class="text-lg font-medium text-slate-900 dark:text-slate-100"
                    >
                        No activity recorded
                    </h3>
                    <p class="text-slate-500 dark:text-slate-400">
                        {#if viewMode === "monthly"}
                            No time blocks for {format(
                                selectedMonth,
                                "MMMM yyyy",
                            )}.
                        {:else}
                            No time blocks have been assigned to this profile
                            yet.
                        {/if}
                    </p>
                </div>
            {/if}
        </div>
    {/if}
</div>
