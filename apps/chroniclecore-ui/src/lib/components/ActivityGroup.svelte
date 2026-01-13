<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { format, parseISO } from "date-fns";

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

    interface ActivityGroup {
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

    export let group: ActivityGroup;
    export let showActions = true;

    const dispatch = createEventDispatcher();

    let expanded = false;

    function formatTime(ts: string): string {
        try {
            return format(parseISO(ts), "HH:mm");
        } catch {
            return ts.slice(11, 16);
        }
    }

    function formatDuration(minutes: number): string {
        const hours = Math.floor(minutes / 60);
        const mins = Math.round(minutes % 60);
        if (hours > 0) {
            return `${hours}h ${mins}m`;
        }
        return `${mins}m`;
    }

    function getAppIcon(appName: string): string {
        const lower = appName.toLowerCase();
        if (lower.includes("outlook") || lower.includes("mail")) return "mail";
        if (lower.includes("excel")) return "spreadsheet";
        if (lower.includes("word")) return "document";
        if (lower.includes("chrome") || lower.includes("edge") || lower.includes("firefox") || lower.includes("opera")) return "browser";
        if (lower.includes("slack") || lower.includes("teams") || lower.includes("discord")) return "chat";
        if (lower.includes("xero")) return "finance";
        if (lower.includes("manual")) return "clock";
        return "app";
    }

    function getConfidenceBadge(confidence: string): { bg: string; text: string } {
        switch (confidence) {
            case "HIGH":
                return { bg: "bg-emerald-100", text: "text-emerald-700" };
            case "MEDIUM":
                return { bg: "bg-amber-100", text: "text-amber-700" };
            default:
                return { bg: "bg-slate-100", text: "text-slate-600" };
        }
    }

    // Generate a color based on profile name (consistent)
    function getProfileColor(name: string): string {
        const colors = [
            "bg-teal-500",
            "bg-indigo-500",
            "bg-rose-500",
            "bg-amber-500",
            "bg-emerald-500",
            "bg-violet-500",
            "bg-cyan-500",
            "bg-orange-500",
        ];
        let hash = 0;
        for (let i = 0; i < name.length; i++) {
            hash = name.charCodeAt(i) + ((hash << 5) - hash);
        }
        return colors[Math.abs(hash) % colors.length];
    }

    $: profileColor = group.color || getProfileColor(group.profile_name || "Unassigned");
    $: uniqueApps = [...new Set(group.apps || group.activities.map(a => a.primary_app_name))];
</script>

<div class="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
    <!-- Group Header -->
    <button
        class="w-full px-4 py-3 flex items-center gap-4 hover:bg-slate-50 transition-colors text-left"
        on:click={() => (expanded = !expanded)}
    >
        <!-- Time Range -->
        <div class="w-16 flex-shrink-0 text-center">
            <div class="text-sm font-medium text-slate-900">
                {formatTime(group.start_time)}
            </div>
            <div class="text-xs text-slate-400">
                to {formatTime(group.end_time)}
            </div>
        </div>

        <!-- Profile Badge -->
        <div class="flex items-center gap-2 w-48 flex-shrink-0">
            <span class="w-3 h-3 rounded {profileColor}"></span>
            <div class="min-w-0">
                <div class="text-sm font-medium text-slate-900 truncate">
                    {group.profile_name || "Unassigned"}
                </div>
                {#if group.client_name && group.client_name !== group.profile_name}
                    <div class="text-xs text-slate-500 truncate">
                        {group.client_name}
                    </div>
                {/if}
            </div>
        </div>

        <!-- Duration -->
        <div class="w-20 flex-shrink-0 text-center">
            <div class="text-sm font-semibold text-slate-900 font-mono">
                {formatDuration(group.total_minutes)}
            </div>
        </div>

        <!-- Summary -->
        <div class="flex-1 min-w-0">
            <p class="text-sm text-slate-600 truncate">
                {group.summary}
            </p>
        </div>

        <!-- App Icons -->
        <div class="flex items-center gap-1 flex-shrink-0">
            {#each uniqueApps.slice(0, 4) as app}
                <div
                    class="w-6 h-6 rounded bg-slate-100 flex items-center justify-center"
                    title={app}
                >
                    {#if getAppIcon(app) === "mail"}
                        <svg class="w-3.5 h-3.5 text-slate-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                        </svg>
                    {:else if getAppIcon(app) === "spreadsheet"}
                        <svg class="w-3.5 h-3.5 text-emerald-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7m0 10a2 2 0 002 2h2a2 2 0 002-2V7a2 2 0 00-2-2h-2a2 2 0 00-2 2" />
                        </svg>
                    {:else if getAppIcon(app) === "document"}
                        <svg class="w-3.5 h-3.5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                        </svg>
                    {:else if getAppIcon(app) === "browser"}
                        <svg class="w-3.5 h-3.5 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
                        </svg>
                    {:else if getAppIcon(app) === "chat"}
                        <svg class="w-3.5 h-3.5 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                        </svg>
                    {:else if getAppIcon(app) === "finance"}
                        <svg class="w-3.5 h-3.5 text-teal-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                    {:else if getAppIcon(app) === "clock"}
                        <svg class="w-3.5 h-3.5 text-orange-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                    {:else}
                        <svg class="w-3.5 h-3.5 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
                        </svg>
                    {/if}
                </div>
            {/each}
            {#if uniqueApps.length > 4}
                <span class="text-xs text-slate-400">+{uniqueApps.length - 4}</span>
            {/if}
        </div>

        <!-- Expand Icon -->
        <div class="flex-shrink-0">
            <svg
                class="w-5 h-5 text-slate-400 transition-transform {expanded ? 'rotate-180' : ''}"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
            >
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
            </svg>
        </div>
    </button>

    <!-- Expanded Activities -->
    {#if expanded}
        <div class="border-t border-slate-100 divide-y divide-slate-100">
            {#each group.activities as activity}
                <div class="px-4 py-2 pl-20 flex items-center gap-4 hover:bg-slate-50">
                    <!-- Time -->
                    <div class="w-24 text-xs text-slate-500 font-mono">
                        {formatTime(activity.ts_start)} - {formatTime(activity.ts_end)}
                    </div>

                    <!-- App Icon -->
                    <div
                        class="w-6 h-6 rounded bg-slate-100 flex items-center justify-center flex-shrink-0"
                        title={activity.primary_app_name}
                    >
                        {#if getAppIcon(activity.primary_app_name) === "mail"}
                            <svg class="w-3.5 h-3.5 text-slate-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                            </svg>
                        {:else}
                            <svg class="w-3.5 h-3.5 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
                            </svg>
                        {/if}
                    </div>

                    <!-- Description -->
                    <div class="flex-1 min-w-0">
                        <p class="text-sm text-slate-700 truncate">
                            {activity.description || activity.title_summary || activity.primary_app_name}
                        </p>
                    </div>

                    <!-- Duration -->
                    <div class="text-xs text-slate-500 font-mono">
                        {formatDuration(activity.duration_minutes)}
                    </div>

                    <!-- Confidence -->
                    <span class="text-xs px-2 py-0.5 rounded {getConfidenceBadge(activity.confidence).bg} {getConfidenceBadge(activity.confidence).text}">
                        {activity.confidence}
                    </span>
                </div>
            {/each}
        </div>
    {/if}
</div>
