<script lang="ts">
    import { onMount } from "svelte";

    let skipDeleteConfirmation = false;

    onMount(() => {
        skipDeleteConfirmation =
            localStorage.getItem("skipDeleteConfirmation") === "true";
    });

    function toggleDeleteConfirmation() {
        skipDeleteConfirmation = !skipDeleteConfirmation;
        localStorage.setItem(
            "skipDeleteConfirmation",
            skipDeleteConfirmation.toString(),
        );
    }
</script>

<div class="max-w-4xl mx-auto space-y-6">
    <h1 class="text-2xl font-bold text-slate-900">Settings</h1>

    <div class="bg-white rounded-xl border border-slate-200 shadow-sm p-6">
        <h2 class="text-lg font-semibold text-slate-900 mb-4">
            Application Info
        </h2>
        <div class="space-y-4">
            <div>
                <div class="block text-sm font-medium text-slate-700">
                    Version
                </div>
                <div class="mt-1 text-sm text-slate-900">v1.8.9</div>
            </div>
            <div>
                <div class="block text-sm font-medium text-slate-700">
                    Data Location
                </div>
                <div
                    class="mt-1 text-sm text-slate-500 font-mono bg-slate-50 p-2 rounded"
                >
                    %LOCALAPPDATA%\ChronicleCore\chronicle.db
                </div>
            </div>
        </div>
    </div>

    <div
        class="bg-white rounded-xl border border-slate-200 shadow-sm p-6 relative overflow-hidden"
    >
        <div class="absolute top-0 right-0 p-4 opacity-5 pointer-events-none">
            <svg
                class="w-32 h-32 text-emerald-900"
                fill="currentColor"
                viewBox="0 0 24 24"
            >
                <path d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
        </div>

        <div class="flex items-start justify-between relative z-10">
            <div>
                <h2
                    class="text-lg font-semibold text-slate-900 flex items-center gap-2"
                >
                    Smart Suggestions
                    <span
                        class="px-2 py-0.5 rounded-full bg-emerald-100 text-emerald-700 text-[10px] font-bold uppercase tracking-wider"
                        >Beta</span
                    >
                </h2>
                <p class="text-sm text-slate-500 mt-1 max-w-lg">
                    Enable local machine learning to suggest profiles for your
                    activity blocks. Training happens entirely on your device.
                </p>

                <div
                    class="mt-4 flex items-center gap-2 text-xs text-slate-500 bg-slate-50 inline-flex px-3 py-1.5 rounded-lg border border-slate-200"
                >
                    <svg
                        class="w-3 h-3 text-emerald-600"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                    >
                        <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                        />
                    </svg>
                    Local-only • No cloud calls • Privacy preserved
                </div>
            </div>

            <div class="flex items-center gap-3">
                <div class="text-right mr-2">
                    <p class="text-sm font-medium text-slate-900">Enabled</p>
                    <p class="text-xs text-emerald-600">Model Active</p>
                </div>
                <button
                    class="relative inline-flex h-6 w-11 items-center rounded-full bg-emerald-600 transition-colors focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:ring-offset-2"
                >
                    <span
                        class="translate-x-6 inline-block h-4 w-4 transform rounded-full bg-white transition-transform"
                    ></span>
                </button>
            </div>
        </div>
    </div>

    <div class="bg-white rounded-xl border border-slate-200 shadow-sm p-6">
        <h2 class="text-lg font-semibold text-slate-900 mb-4">
            Tracking Preferences
        </h2>
        <div class="space-y-4">
            <div
                class="flex items-center justify-between p-4 bg-slate-50 rounded-lg border border-slate-100"
            >
                <div>
                    <h3 class="font-medium text-slate-900">
                        Blacklist Management
                    </h3>
                    <p class="text-sm text-slate-500">
                        Manage applications that are ignored by the usage
                        tracker.
                    </p>
                </div>
                <a
                    href="/settings/blacklist"
                    class="px-4 py-2 bg-white border border-slate-200 text-slate-700 rounded-lg hover:bg-slate-50 hover:text-slate-900 font-medium text-sm transition-colors shadow-sm"
                >
                    Manage Blacklist
                </a>
            </div>
        </div>
    </div>

    <div class="bg-white rounded-xl border border-slate-200 shadow-sm p-6">
        <h2 class="text-lg font-semibold text-slate-900 mb-4">
            User Preferences
        </h2>
        <div class="space-y-4">
            <div
                class="flex items-center justify-between p-4 bg-slate-50 rounded-lg border border-slate-100"
            >
                <div>
                    <h3 class="font-medium text-slate-900">
                        Skip Delete Confirmations
                    </h3>
                    <p class="text-sm text-slate-500">
                        Remove confirmation dialogs when deleting items or
                        groups.
                    </p>
                </div>
                <button
                    on:click={toggleDeleteConfirmation}
                    class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 {skipDeleteConfirmation
                        ? 'bg-blue-600'
                        : 'bg-slate-300'}"
                >
                    <span
                        class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform {skipDeleteConfirmation
                            ? 'translate-x-6'
                            : 'translate-x-1'}"
                    ></span>
                </button>
            </div>
        </div>
    </div>
</div>
