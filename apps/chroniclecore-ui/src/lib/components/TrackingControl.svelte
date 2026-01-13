<script lang="ts">
    import { onMount, onDestroy } from "svelte";
    import { trackingStatus } from "$lib/stores";
    import { fetchApi } from "$lib/api";

    let refreshInterval: number;
    let lastError: string | null = null;
    let updating = false;

    async function refreshStatus() {
        try {
            const status = await fetchApi("/tracking/status");
            trackingStatus.set(status);
            lastError = null;
        } catch (err: any) {
            console.error("Failed to poll status", err);
            lastError = err.message || "Connection lost";
        }
    }

    async function control(action: "start" | "pause" | "resume" | "stop") {
        updating = true;
        try {
            await fetchApi(`/tracking/${action}`, { method: "POST" });
            await refreshStatus();
        } catch (e: any) {
            lastError = e.message;
        } finally {
            updating = false;
        }
    }

    onMount(() => {
        refreshStatus();
        refreshInterval = setInterval(refreshStatus, 2000);
    });

    onDestroy(() => {
        clearInterval(refreshInterval);
    });

    $: state = $trackingStatus?.state || "OFFLINE";
    $: current = $trackingStatus?.current_window;
    $: isOff = state === "STOPPED" || state === "OFFLINE";
</script>

<!-- Added Gradient Background here -->
<div
    class="bg-gradient-to-br from-white to-blue-50/50 rounded-2xl border border-slate-200 shadow-lg shadow-slate-200/50 p-8 mb-10 relative overflow-hidden"
>
    <!-- Background Decor -->
    <div
        class="absolute -top-12 -right-12 w-64 h-64 bg-blue-100 rounded-full blur-3xl opacity-30 pointer-events-none"
    ></div>
    <div
        class="absolute -bottom-12 -left-12 w-64 h-64 bg-slate-100 rounded-full blur-3xl opacity-50 pointer-events-none"
    ></div>

    {#if lastError}
        <div
            class="absolute top-0 left-0 right-0 bg-red-500 text-white text-xs px-4 py-1 text-center font-bold shadow-sm z-20"
        >
            ⚠️ Connection Lost: {lastError}
        </div>
    {/if}

    <div
        class="flex flex-col md:flex-row md:items-center justify-between gap-8 relative z-10"
    >
        <!-- Status -->
        <div class="flex items-center gap-6">
            <div class="relative group">
                <!-- Outer Pulse Ring -->
                {#if state === "ACTIVE"}
                    <div
                        class="absolute -inset-4 bg-emerald-400/20 rounded-full blur-xl animate-pulse"
                    ></div>
                {/if}

                <!-- Main Circle -->
                <div
                    class="w-16 h-16 rounded-2xl rotate-3 flex items-center justify-center shadow-xl transition-all duration-500
                    {state === 'ACTIVE'
                        ? 'bg-emerald-500 text-white shadow-emerald-500/30 ring-4 ring-white'
                        : state === 'PAUSED'
                          ? 'bg-amber-400 text-white shadow-amber-500/30'
                          : 'bg-white text-slate-300 shadow-slate-200/50'}"
                >
                    {#if state === "ACTIVE"}
                        <svg
                            class="w-8 h-8"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                        >
                            <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="2.5"
                                d="M13 10V3L4 14h7v7l9-11h-7z"
                            />
                        </svg>
                    {:else if state === "PAUSED"}
                        <svg
                            class="w-8 h-8"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                        >
                            <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="2.5"
                                d="M10 9v6m4-6v6"
                            />
                        </svg>
                    {:else}
                        <svg
                            class="w-8 h-8"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                        >
                            <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="2"
                                d="M5 12h.01M12 12h.01M19 12h.01M6 12a1 1 0 11-2 0 1 1 0 012 0zm7 0a1 1 0 11-2 0 1 1 0 012 0zm7 0a1 1 0 11-2 0 1 1 0 012 0z"
                            />
                        </svg>
                    {/if}
                </div>
            </div>

            <div>
                <div class="flex items-center gap-3 mb-1">
                    <h2
                        class="text-2xl font-bold text-slate-900 tracking-tight"
                    >
                        {#if state === "ACTIVE"}
                            Tracking Active
                        {:else if state === "PAUSED"}
                            Session Paused
                        {:else}
                            Ready to Work?
                        {/if}
                    </h2>
                    {#if state === "ACTIVE"}
                        <span
                            class="inline-flex px-2 py-0.5 rounded text-[10px] font-bold bg-emerald-100 text-emerald-700 tracking-wide uppercase"
                            >Live</span
                        >
                    {/if}
                </div>

                <div class="text-base text-slate-500 font-medium max-w-lg">
                    {#if state === "STOPPED"}
                        Launch an activity to begin your session.
                    {:else if current}
                        <span class="text-slate-800 font-semibold"
                            >{current.app_name}</span
                        >
                        <span class="text-slate-300 mx-1">/</span>
                        {current.title}
                    {:else}
                        Monitoring system events...
                    {/if}
                </div>
            </div>
        </div>

        <!-- Controls -->
        <div class="flex items-center gap-3 scale-100">
            {#if isOff}
                <button
                    on:click={() => control("start")}
                    disabled={updating}
                    class="group relative inline-flex items-center gap-3 bg-slate-900 text-white px-8 py-3.5 rounded-xl font-semibold shadow-xl shadow-slate-900/20 transition-all hover:scale-105 hover:shadow-2xl disabled:opacity-70 disabled:scale-100"
                >
                    <span
                        class="absolute inset-0 bg-white/10 rounded-xl opacity-0 group-hover:opacity-100 transition-opacity"
                    ></span>
                    <svg
                        class="w-5 h-5"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                    >
                        <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"
                        />
                        <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                        />
                    </svg>
                    Start Session
                </button>
            {:else}
                {#if state === "ACTIVE"}
                    <button
                        on:click={() => control("pause")}
                        disabled={updating}
                        class="flex items-center gap-2 bg-white border border-slate-200 hover:bg-slate-50 text-slate-700 px-6 py-3 rounded-xl font-semibold shadow-sm transition-all hover:-translate-y-0.5 disabled:opacity-50"
                    >
                        <svg
                            class="w-5 h-5 text-slate-400"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                        >
                            <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="2.5"
                                d="M10 9v6m4-6v6"
                            />
                        </svg>
                        Pause
                    </button>
                {:else}
                    <button
                        on:click={() => control("resume")}
                        disabled={updating}
                        class="flex items-center gap-2 bg-emerald-500 hover:bg-emerald-600 text-white px-6 py-3 rounded-xl font-semibold shadow-lg shadow-emerald-200 transition-all hover:-translate-y-0.5 disabled:opacity-50"
                    >
                        <svg
                            class="w-5 h-5"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                        >
                            <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="2.5"
                                d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"
                            />
                        </svg>
                        Resume
                    </button>
                {/if}

                <button
                    on:click={() => control("stop")}
                    disabled={updating}
                    class="flex items-center gap-2 bg-white border border-rose-100 hover:bg-rose-50 text-rose-600 px-6 py-3 rounded-xl font-semibold shadow-sm transition-all hover:border-rose-200 disabled:opacity-50"
                >
                    <svg
                        class="w-5 h-5"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                    >
                        <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                        />
                        <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2.5"
                            d="M9 9h6v6H9z"
                        />
                    </svg>
                    Finish
                </button>
            {/if}
        </div>
    </div>
</div>
