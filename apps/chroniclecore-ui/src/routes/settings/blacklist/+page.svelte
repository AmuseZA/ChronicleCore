<script lang="ts">
    import { onMount } from "svelte";
    import { fetchApi } from "$lib/api";

    // App Blacklist
    let blacklist: any[] = [];
    let loading = true;
    let error: string | null = null;
    let newAppName = "";
    let adding = false;

    // Keyword Blacklist
    let keywordBlacklist: any[] = [];
    let keywordLoading = true;
    let keywordError: string | null = null;
    let newKeyword = "";
    let addingKeyword = false;

    async function loadBlacklist() {
        try {
            loading = true;
            error = null;
            const res = await fetchApi("/blacklist");
            blacklist = res || [];
        } catch (err: any) {
            console.error(err);
            error = err.message || "Failed to load blacklist";
        } finally {
            loading = false;
        }
    }

    async function loadKeywordBlacklist() {
        try {
            keywordLoading = true;
            keywordError = null;
            const res = await fetchApi("/blacklist/keywords");
            keywordBlacklist = res || [];
        } catch (err: any) {
            console.error(err);
            keywordError = err.message || "Failed to load keyword blacklist";
        } finally {
            keywordLoading = false;
        }
    }

    async function addToBlacklist() {
        if (!newAppName.trim()) return;

        try {
            adding = true;
            await fetchApi("/blacklist", {
                method: "POST",
                body: JSON.stringify({ app_name: newAppName }),
            });
            newAppName = "";
            await loadBlacklist();
        } catch (err: any) {
            alert("Failed to add to blacklist: " + err.message);
        } finally {
            adding = false;
        }
    }

    async function addKeywordToBlacklist() {
        if (!newKeyword.trim()) return;

        try {
            addingKeyword = true;
            await fetchApi("/blacklist/keywords", {
                method: "POST",
                body: JSON.stringify({ keyword: newKeyword }),
            });
            newKeyword = "";
            await loadKeywordBlacklist();
        } catch (err: any) {
            alert("Failed to add keyword to blacklist: " + err.message);
        } finally {
            addingKeyword = false;
        }
    }

    async function removeFromBlacklist(id: number) {
        if (
            !confirm(
                "Are you sure you want to remove this app from the blacklist?",
            )
        )
            return;

        try {
            await fetchApi(`/blacklist/${id}`, { method: "DELETE" });
            await loadBlacklist();
        } catch (err: any) {
            alert("Failed to remove from blacklist: " + err.message);
        }
    }

    async function removeKeywordFromBlacklist(id: number) {
        if (
            !confirm(
                "Are you sure you want to remove this keyword from the blacklist?",
            )
        )
            return;

        try {
            await fetchApi(`/blacklist/keywords/${id}`, { method: "DELETE" });
            await loadKeywordBlacklist();
        } catch (err: any) {
            alert("Failed to remove keyword from blacklist: " + err.message);
        }
    }

    onMount(() => {
        loadBlacklist();
        loadKeywordBlacklist();
    });
</script>

<div class="max-w-4xl mx-auto space-y-6">
    <div class="flex items-center justify-between">
        <div>
            <h1 class="text-2xl font-bold text-slate-900 tracking-tight">
                Blacklist Management
            </h1>
            <p class="text-slate-500 text-sm">
                Manage applications that should be ignored by the tracker.
            </p>
        </div>
        <a
            href="/settings"
            class="text-indigo-600 hover:text-indigo-700 font-medium text-sm"
        >
            &larr; Back to Settings
        </a>
    </div>

    <!-- Add New -->
    <div class="bg-white dark:bg-slate-800 p-6 rounded-2xl border border-slate-200 dark:border-slate-700 shadow-sm">
        <h2 class="text-lg font-semibold text-slate-900 mb-4">
            Add Application
        </h2>
        <form on:submit|preventDefault={addToBlacklist} class="flex gap-4">
            <div class="flex-1">
                <input
                    type="text"
                    bind:value={newAppName}
                    placeholder="Enter application process name (e.g., spotify.exe, chrome.exe)"
                    class="w-full px-4 py-2 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none transition-all"
                />
                <p class="mt-1 text-xs text-slate-500">
                    Use the exact process name if known, or the application
                    name.
                </p>
            </div>
            <button
                type="submit"
                disabled={adding || !newAppName.trim()}
                class="bg-slate-900 text-white px-6 py-2 rounded-lg font-medium hover:bg-slate-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
                {adding ? "Adding..." : "Add to Blacklist"}
            </button>
        </form>
    </div>

    <!-- List -->
    <div
        class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 shadow-sm overflow-hidden"
    >
        <div class="px-6 py-4 border-b border-slate-100 bg-slate-50/50">
            <h2 class="text-lg font-semibold text-slate-900">
                Blacklisted Applications
            </h2>
        </div>

        {#if loading}
            <div class="p-8 text-center text-slate-500">
                Loading blacklist...
            </div>
        {:else if error}
            <div class="p-8 text-center text-red-600 bg-red-50">
                Error: {error}
            </div>
        {:else if blacklist.length === 0}
            <div class="p-12 text-center">
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
                            d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                        />
                    </svg>
                </div>
                <h3 class="text-lg font-medium text-slate-900">
                    No blacklisted apps
                </h3>
                <p class="text-slate-500 mt-1">
                    All applications are currently being tracked.
                </p>
            </div>
        {:else}
            <div class="divide-y divide-slate-100">
                {#each blacklist as item}
                    <div
                        class="p-4 px-6 flex items-center justify-between hover:bg-slate-50 transition-colors"
                    >
                        <div>
                            <div class="font-medium text-slate-900">
                                {item.app_name}
                            </div>
                            <div class="text-xs text-slate-500">
                                Added: {new Date(
                                    item.created_at,
                                ).toLocaleDateString()}
                            </div>
                        </div>
                        <button
                            on:click={() =>
                                removeFromBlacklist(item.blacklist_id)}
                            class="text-red-600 hover:text-red-700 text-sm font-medium px-3 py-1 rounded hover:bg-red-50 transition-colors"
                        >
                            Remove
                        </button>
                    </div>
                {/each}
            </div>
        {/if}
    </div>

    <!-- Keyword Blacklist Section -->
    <div
        class="bg-white dark:bg-slate-800 p-6 rounded-2xl border border-slate-200 dark:border-slate-700 shadow-sm mt-8"
    >
        <h2 class="text-lg font-semibold text-slate-900 mb-4">Add Keyword</h2>
        <p class="text-slate-500 text-sm mb-4">
            Block activities containing specific keywords in their title (e.g.,
            "Facebook", "YouTube").
        </p>
        <form
            on:submit|preventDefault={addKeywordToBlacklist}
            class="flex gap-4"
        >
            <div class="flex-1">
                <input
                    type="text"
                    bind:value={newKeyword}
                    placeholder="Enter keyword to block (e.g., Facebook, Netflix)"
                    class="w-full px-4 py-2 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none transition-all"
                />
            </div>
            <button
                type="submit"
                disabled={addingKeyword || !newKeyword.trim()}
                class="bg-slate-900 text-white px-6 py-2 rounded-lg font-medium hover:bg-slate-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
                {addingKeyword ? "Adding..." : "Add Keyword"}
            </button>
        </form>
    </div>

    <!-- Keyword List -->
    <div
        class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 shadow-sm overflow-hidden"
    >
        <div class="px-6 py-4 border-b border-slate-100 bg-slate-50/50">
            <h2 class="text-lg font-semibold text-slate-900">
                Blacklisted Keywords
            </h2>
        </div>

        {#if keywordLoading}
            <div class="p-8 text-center text-slate-500">
                Loading keywords...
            </div>
        {:else if keywordError}
            <div class="p-8 text-center text-red-600 bg-red-50">
                Error: {keywordError}
            </div>
        {:else if keywordBlacklist.length === 0}
            <div class="p-12 text-center">
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
                            d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z"
                        />
                    </svg>
                </div>
                <h3 class="text-lg font-medium text-slate-900">
                    No keyword filters
                </h3>
                <p class="text-slate-500 mt-1">
                    Add keywords to filter out unwanted activities.
                </p>
            </div>
        {:else}
            <div class="divide-y divide-slate-100">
                {#each keywordBlacklist as item}
                    <div
                        class="p-4 px-6 flex items-center justify-between hover:bg-slate-50 transition-colors"
                    >
                        <div>
                            <div class="font-medium text-slate-900">
                                "{item.keyword_text}"
                            </div>
                            <div class="text-xs text-slate-500">
                                Added: {new Date(
                                    item.created_at,
                                ).toLocaleDateString()}
                            </div>
                        </div>
                        <button
                            on:click={() =>
                                removeKeywordFromBlacklist(item.keyword_id)}
                            class="text-red-600 hover:text-red-700 text-sm font-medium px-3 py-1 rounded hover:bg-red-50 transition-colors"
                        >
                            Remove
                        </button>
                    </div>
                {/each}
            </div>
        {/if}
    </div>
</div>
