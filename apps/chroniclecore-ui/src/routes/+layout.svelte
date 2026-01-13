<script lang="ts">
    import "../app.css";
    import Sidebar from "$lib/components/Sidebar.svelte";
    import { ui } from "$lib/stores/ui";
    import { onMount } from "svelte";
    import { fetchApi } from "$lib/api";

    interface UpdateInfo {
        update_available: boolean;
        current_version: string;
        latest_version?: string;
        release_notes?: string;
        download_url?: string;
        release_url?: string;
    }

    let updateInfo: UpdateInfo | null = null;
    let dismissed = false;

    onMount(async () => {
        // Check for updates on app load (with a slight delay to not block UI)
        setTimeout(async () => {
            try {
                const info = await fetchApi('/system/check-update');
                if (info?.update_available) {
                    updateInfo = info;
                }
            } catch (e) {
                // Silently fail - update check is not critical
                console.debug('Update check failed:', e);
            }
        }, 2000);
    });

    function dismiss() {
        dismissed = true;
    }

    function openDownload() {
        if (updateInfo?.download_url) {
            window.open(updateInfo.download_url, '_blank');
        } else if (updateInfo?.release_url) {
            window.open(updateInfo.release_url, '_blank');
        }
    }
</script>

<div class="flex min-h-screen bg-slate-50 font-sans">
    <Sidebar />
    <main
        class="flex-1 p-8 transition-all duration-300 {$ui.isSidebarCollapsed
            ? 'ml-20'
            : 'ml-72'}"
    >
        <!-- Update Banner -->
        {#if updateInfo && !dismissed}
            <div class="mb-6 bg-gradient-to-r from-blue-600 to-indigo-600 rounded-xl p-4 shadow-lg text-white">
                <div class="flex items-center justify-between">
                    <div class="flex items-center gap-3">
                        <div class="w-10 h-10 bg-white/20 rounded-lg flex items-center justify-center">
                            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                            </svg>
                        </div>
                        <div>
                            <h3 class="font-semibold">Update Available!</h3>
                            <p class="text-sm text-blue-100">
                                Version {updateInfo.latest_version} is ready to download
                                <span class="text-blue-200">(you have {updateInfo.current_version})</span>
                            </p>
                        </div>
                    </div>
                    <div class="flex items-center gap-2">
                        <button
                            on:click={openDownload}
                            class="px-4 py-2 bg-white text-blue-600 font-semibold rounded-lg hover:bg-blue-50 transition-colors text-sm"
                        >
                            Download Update
                        </button>
                        <button
                            on:click={dismiss}
                            class="p-2 hover:bg-white/10 rounded-lg transition-colors"
                            title="Dismiss"
                        >
                            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                            </svg>
                        </button>
                    </div>
                </div>
            </div>
        {/if}

        <slot />
    </main>
</div>
