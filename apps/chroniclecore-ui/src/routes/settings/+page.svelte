<script lang="ts">
    import { onMount } from "svelte";

    let skipDeleteConfirmation = false;

    // Deep tracking settings
    interface TrackingSettings {
        full_tracking_mode: boolean;
        deep_tracking_enabled: boolean;
        track_browser_content: boolean;
        track_email_content: boolean;
        track_document_content: boolean;
        track_chat_content: boolean;
        privacy_mode: boolean;
        excluded_apps: string[];
        idle_threshold_seconds: number;
    }

    let trackingSettings: TrackingSettings = {
        full_tracking_mode: false,
        deep_tracking_enabled: false,
        track_browser_content: true,
        track_email_content: true,
        track_document_content: true,
        track_chat_content: true,
        privacy_mode: false,
        excluded_apps: [],
        idle_threshold_seconds: 300,
    };

    let settingsLoading = true;
    let settingsSaving = false;

    async function loadSettings() {
        try {
            const response = await fetch("/api/v1/settings");
            if (response.ok) {
                trackingSettings = await response.json();
            }
        } catch (error) {
            console.error("Failed to load settings:", error);
        } finally {
            settingsLoading = false;
        }
    }

    async function saveSettings() {
        settingsSaving = true;
        try {
            const response = await fetch("/api/v1/settings", {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(trackingSettings),
            });
            if (response.ok) {
                trackingSettings = await response.json();
            }
        } catch (error) {
            console.error("Failed to save settings:", error);
        } finally {
            settingsSaving = false;
        }
    }

    function toggleSetting(key: keyof TrackingSettings) {
        if (typeof trackingSettings[key] === "boolean") {
            (trackingSettings as any)[key] = !trackingSettings[key];
            saveSettings();
        }
    }

    onMount(() => {
        skipDeleteConfirmation =
            localStorage.getItem("skipDeleteConfirmation") === "true";
        loadSettings();
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
    <h1 class="text-2xl font-bold text-slate-900 dark:text-slate-100">
        Settings
    </h1>

    <div
        class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 shadow-sm p-6"
    >
        <h2
            class="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-4"
        >
            Application Info
        </h2>
        <div class="space-y-4">
            <div>
                <div
                    class="block text-sm font-medium text-slate-700 dark:text-slate-300"
                >
                    Version
                </div>
                <div class="mt-1 text-sm text-slate-900 dark:text-slate-100">
                    v2.1.0
                </div>
            </div>
            <div>
                <div
                    class="block text-sm font-medium text-slate-700 dark:text-slate-300"
                >
                    Data Location
                </div>
                <div
                    class="mt-1 text-sm text-slate-500 dark:text-slate-400 font-mono bg-slate-50 dark:bg-slate-900 p-2 rounded"
                >
                    %LOCALAPPDATA%\ChronicleCore\chronicle.db
                </div>
            </div>
        </div>
    </div>

    <div
        class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 shadow-sm p-6 relative overflow-hidden"
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
                    class="text-lg font-semibold text-slate-900 dark:text-slate-100 flex items-center gap-2"
                >
                    Smart Suggestions
                    <span
                        class="px-2 py-0.5 rounded-full bg-emerald-100 text-emerald-700 text-[10px] font-bold uppercase tracking-wider"
                        >Beta</span
                    >
                </h2>
                <p
                    class="text-sm text-slate-500 dark:text-slate-400 mt-1 max-w-lg"
                >
                    Enable local machine learning to suggest profiles for your
                    activity blocks. Training happens entirely on your device.
                </p>

                <div
                    class="mt-4 flex items-center gap-2 text-xs text-slate-500 dark:text-slate-400 bg-slate-50 dark:bg-slate-900 inline-flex px-3 py-1.5 rounded-lg border border-slate-200 dark:border-slate-700"
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
                    <p
                        class="text-sm font-medium text-slate-900 dark:text-slate-100"
                    >
                        Enabled
                    </p>
                    <p class="text-xs text-emerald-600 dark:text-emerald-400">
                        Model Active
                    </p>
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

    <div
        class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 shadow-sm p-6"
    >
        <h2
            class="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-4"
        >
            Tracking Preferences
        </h2>
        <div class="space-y-4">
            <div
                class="flex items-center justify-between p-4 bg-slate-50 dark:bg-slate-700 rounded-lg border border-slate-100 dark:border-slate-600"
            >
                <div>
                    <h3 class="font-medium text-slate-900 dark:text-slate-100">
                        Blacklist Management
                    </h3>
                    <p class="text-sm text-slate-500 dark:text-slate-400">
                        Manage applications that are ignored by the usage
                        tracker.
                    </p>
                </div>
                <a
                    href="/settings/blacklist"
                    class="px-4 py-2 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-300 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-700 hover:text-slate-900 dark:hover:text-slate-100 font-medium text-sm transition-colors shadow-sm"
                >
                    Manage Blacklist
                </a>
            </div>
        </div>
    </div>

    <!-- Deep Tracking Settings -->
    <div
        class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 shadow-sm p-6 relative overflow-hidden"
    >
        <div class="absolute top-0 right-0 p-4 opacity-5 pointer-events-none">
            <svg
                class="w-32 h-32 text-blue-900"
                fill="currentColor"
                viewBox="0 0 24 24"
            >
                <path
                    d="M12 4.5C7 4.5 2.73 7.61 1 12c1.73 4.39 6 7.5 11 7.5s9.27-3.11 11-7.5c-1.73-4.39-6-7.5-11-7.5zM12 17c-2.76 0-5-2.24-5-5s2.24-5 5-5 5 2.24 5 5-2.24 5-5 5zm0-8c-1.66 0-3 1.34-3 3s1.34 3 3 3 3-1.34 3-3-1.34-3-3-3z"
                />
            </svg>
        </div>

        <div class="flex items-start justify-between mb-6 relative z-10">
            <div>
                <h2
                    class="text-lg font-semibold text-slate-900 dark:text-slate-100 flex items-center gap-2"
                >
                    Deep Activity Tracking
                    <span
                        class="px-2 py-0.5 rounded-full bg-blue-100 text-blue-700 text-[10px] font-bold uppercase tracking-wider"
                        >New</span
                    >
                </h2>
                <p
                    class="text-sm text-slate-500 dark:text-slate-400 mt-1 max-w-lg"
                >
                    Extract detailed information like document names, email
                    subjects, and chat contacts for more accurate activity
                    descriptions.
                </p>
                <div
                    class="mt-3 flex items-center gap-2 text-xs text-slate-500 dark:text-slate-400 bg-slate-50 dark:bg-slate-900 inline-flex px-3 py-1.5 rounded-lg border border-slate-200 dark:border-slate-700"
                >
                    <svg
                        class="w-3 h-3 text-blue-600"
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
                    100% Local • No data leaves your device
                </div>
            </div>

            <div class="flex items-center gap-3">
                <div class="text-right mr-2">
                    <p
                        class="text-sm font-medium text-slate-900 dark:text-slate-100"
                    >
                        {trackingSettings.deep_tracking_enabled
                            ? "Enabled"
                            : "Disabled"}
                    </p>
                    <p
                        class="text-xs {trackingSettings.deep_tracking_enabled
                            ? 'text-blue-600'
                            : 'text-slate-400'}"
                    >
                        {trackingSettings.deep_tracking_enabled
                            ? "Tracking Details"
                            : "Basic Mode"}
                    </p>
                </div>
                <button
                    on:click={() => toggleSetting("deep_tracking_enabled")}
                    disabled={settingsLoading || settingsSaving}
                    class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 {trackingSettings.deep_tracking_enabled
                        ? 'bg-blue-600'
                        : 'bg-slate-300'}"
                >
                    <span
                        class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform {trackingSettings.deep_tracking_enabled
                            ? 'translate-x-6'
                            : 'translate-x-1'}"
                    ></span>
                </button>
            </div>
        </div>

        {#if trackingSettings.deep_tracking_enabled}
            <div
                class="space-y-3 pt-4 border-t border-slate-200 dark:border-slate-700"
            >
                <h3
                    class="text-sm font-medium text-slate-700 dark:text-slate-300 mb-3"
                >
                    Content Types to Track
                </h3>

                <!-- Browser Content -->
                <div
                    class="flex items-center justify-between p-3 bg-slate-50 dark:bg-slate-700 rounded-lg"
                >
                    <div class="flex items-center gap-3">
                        <div
                            class="w-8 h-8 rounded-lg bg-orange-100 dark:bg-orange-900/50 flex items-center justify-center"
                        >
                            <svg
                                class="w-4 h-4 text-orange-600"
                                fill="none"
                                viewBox="0 0 24 24"
                                stroke="currentColor"
                            >
                                <path
                                    stroke-linecap="round"
                                    stroke-linejoin="round"
                                    stroke-width="2"
                                    d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9"
                                />
                            </svg>
                        </div>
                        <div>
                            <p
                                class="text-sm font-medium text-slate-900 dark:text-slate-100"
                            >
                                Browser Activity
                            </p>
                            <p
                                class="text-xs text-slate-500 dark:text-slate-400"
                            >
                                URLs, page titles, domains
                            </p>
                        </div>
                    </div>
                    <button
                        on:click={() => toggleSetting("track_browser_content")}
                        class="relative inline-flex h-5 w-9 items-center rounded-full transition-colors {trackingSettings.track_browser_content
                            ? 'bg-blue-600'
                            : 'bg-slate-300'}"
                    >
                        <span
                            class="inline-block h-3 w-3 transform rounded-full bg-white transition-transform {trackingSettings.track_browser_content
                                ? 'translate-x-5'
                                : 'translate-x-1'}"
                        ></span>
                    </button>
                </div>

                <!-- Email Content -->
                <div
                    class="flex items-center justify-between p-3 bg-slate-50 dark:bg-slate-700 rounded-lg"
                >
                    <div class="flex items-center gap-3">
                        <div
                            class="w-8 h-8 rounded-lg bg-red-100 dark:bg-red-900/50 flex items-center justify-center"
                        >
                            <svg
                                class="w-4 h-4 text-red-600"
                                fill="none"
                                viewBox="0 0 24 24"
                                stroke="currentColor"
                            >
                                <path
                                    stroke-linecap="round"
                                    stroke-linejoin="round"
                                    stroke-width="2"
                                    d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
                                />
                            </svg>
                        </div>
                        <div>
                            <p
                                class="text-sm font-medium text-slate-900 dark:text-slate-100"
                            >
                                Email Activity
                            </p>
                            <p
                                class="text-xs text-slate-500 dark:text-slate-400"
                            >
                                Email subjects, sender info (Outlook, Gmail)
                            </p>
                        </div>
                    </div>
                    <button
                        on:click={() => toggleSetting("track_email_content")}
                        class="relative inline-flex h-5 w-9 items-center rounded-full transition-colors {trackingSettings.track_email_content
                            ? 'bg-blue-600'
                            : 'bg-slate-300'}"
                    >
                        <span
                            class="inline-block h-3 w-3 transform rounded-full bg-white transition-transform {trackingSettings.track_email_content
                                ? 'translate-x-5'
                                : 'translate-x-1'}"
                        ></span>
                    </button>
                </div>

                <!-- Document Content -->
                <div
                    class="flex items-center justify-between p-3 bg-slate-50 dark:bg-slate-700 rounded-lg"
                >
                    <div class="flex items-center gap-3">
                        <div
                            class="w-8 h-8 rounded-lg bg-blue-100 dark:bg-blue-900/50 flex items-center justify-center"
                        >
                            <svg
                                class="w-4 h-4 text-blue-600"
                                fill="none"
                                viewBox="0 0 24 24"
                                stroke="currentColor"
                            >
                                <path
                                    stroke-linecap="round"
                                    stroke-linejoin="round"
                                    stroke-width="2"
                                    d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                                />
                            </svg>
                        </div>
                        <div>
                            <p
                                class="text-sm font-medium text-slate-900 dark:text-slate-100"
                            >
                                Document Activity
                            </p>
                            <p
                                class="text-xs text-slate-500 dark:text-slate-400"
                            >
                                Document names, project files (Word, Excel, VS
                                Code)
                            </p>
                        </div>
                    </div>
                    <button
                        on:click={() => toggleSetting("track_document_content")}
                        class="relative inline-flex h-5 w-9 items-center rounded-full transition-colors {trackingSettings.track_document_content
                            ? 'bg-blue-600'
                            : 'bg-slate-300'}"
                    >
                        <span
                            class="inline-block h-3 w-3 transform rounded-full bg-white transition-transform {trackingSettings.track_document_content
                                ? 'translate-x-5'
                                : 'translate-x-1'}"
                        ></span>
                    </button>
                </div>

                <!-- Chat Content -->
                <div
                    class="flex items-center justify-between p-3 bg-slate-50 dark:bg-slate-700 rounded-lg"
                >
                    <div class="flex items-center gap-3">
                        <div
                            class="w-8 h-8 rounded-lg bg-green-100 dark:bg-green-900/50 flex items-center justify-center"
                        >
                            <svg
                                class="w-4 h-4 text-green-600"
                                fill="none"
                                viewBox="0 0 24 24"
                                stroke="currentColor"
                            >
                                <path
                                    stroke-linecap="round"
                                    stroke-linejoin="round"
                                    stroke-width="2"
                                    d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
                                />
                            </svg>
                        </div>
                        <div>
                            <p
                                class="text-sm font-medium text-slate-900 dark:text-slate-100"
                            >
                                Chat Activity
                            </p>
                            <p
                                class="text-xs text-slate-500 dark:text-slate-400"
                            >
                                Contact names, channels (Teams, Slack, WhatsApp)
                            </p>
                        </div>
                    </div>
                    <button
                        on:click={() => toggleSetting("track_chat_content")}
                        class="relative inline-flex h-5 w-9 items-center rounded-full transition-colors {trackingSettings.track_chat_content
                            ? 'bg-blue-600'
                            : 'bg-slate-300'}"
                    >
                        <span
                            class="inline-block h-3 w-3 transform rounded-full bg-white transition-transform {trackingSettings.track_chat_content
                                ? 'translate-x-5'
                                : 'translate-x-1'}"
                        ></span>
                    </button>
                </div>

                <!-- Privacy Mode -->
                <div
                    class="flex items-center justify-between p-3 bg-amber-50 dark:bg-amber-900/20 rounded-lg border border-amber-200 dark:border-amber-800 mt-4"
                >
                    <div class="flex items-center gap-3">
                        <div
                            class="w-8 h-8 rounded-lg bg-amber-100 dark:bg-amber-900/50 flex items-center justify-center"
                        >
                            <svg
                                class="w-4 h-4 text-amber-600"
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
                        </div>
                        <div>
                            <p
                                class="text-sm font-medium text-slate-900 dark:text-slate-100"
                            >
                                Privacy Mode
                            </p>
                            <p
                                class="text-xs text-slate-500 dark:text-slate-400"
                            >
                                Redact email addresses and truncate sensitive
                                content
                            </p>
                        </div>
                    </div>
                    <button
                        on:click={() => toggleSetting("privacy_mode")}
                        class="relative inline-flex h-5 w-9 items-center rounded-full transition-colors {trackingSettings.privacy_mode
                            ? 'bg-amber-600'
                            : 'bg-slate-300'}"
                    >
                        <span
                            class="inline-block h-3 w-3 transform rounded-full bg-white transition-transform {trackingSettings.privacy_mode
                                ? 'translate-x-5'
                                : 'translate-x-1'}"
                        ></span>
                    </button>
                </div>
            </div>
        {/if}
    </div>

    <div
        class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 shadow-sm p-6"
    >
        <h2
            class="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-4"
        >
            User Preferences
        </h2>
        <div class="space-y-4">
            <div
                class="flex items-center justify-between p-4 bg-slate-50 dark:bg-slate-700 rounded-lg border border-slate-100 dark:border-slate-600"
            >
                <div>
                    <h3 class="font-medium text-slate-900 dark:text-slate-100">
                        Skip Delete Confirmations
                    </h3>
                    <p class="text-sm text-slate-500 dark:text-slate-400">
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
