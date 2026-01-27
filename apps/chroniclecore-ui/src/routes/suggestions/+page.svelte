<script lang="ts">
    import { onMount } from "svelte";
    import { fetchApi } from "$lib/api";
    import { format, parseISO } from "date-fns";
    import ProfileSelector from "$lib/components/ProfileSelector.svelte";

    interface Suggestion {
        suggestion_id: number;
        entity_type: string;
        entity_id: number;
        suggestion_type: string;
        payload: {
            predicted_profile_id: number;
            confidence_level: string;
        };
        confidence: number;
        status: string;
        created_at: string;
        // Block details (fetched separately)
        block_details?: {
            app_name: string;
            title_summary: string;
            ts_start: string;
            ts_end: string;
            duration_minutes: number;
        };
        profile_name?: string;
    }

    interface Profile {
        profile_id: number;
        name: string;
        client_name?: string;
    }

    let suggestions: Suggestion[] = [];
    let profiles: Profile[] = [];
    let loading = true;
    let processing: number | null = null;

    // Separate suggestions by type
    $: profileSuggestions = suggestions.filter(
        (s) => s.suggestion_type === "PROFILE_ASSIGN",
    );
    $: deleteSuggestions = suggestions.filter(
        (s) => s.suggestion_type === "DELETE_SUGGEST",
    );

    // Group profile suggestions by confidence level
    $: highConfidence = profileSuggestions.filter((s) => s.confidence >= 0.85);
    $: mediumConfidence = profileSuggestions.filter(
        (s) => s.confidence >= 0.6 && s.confidence < 0.85,
    );
    $: lowConfidence = profileSuggestions.filter((s) => s.confidence < 0.6);

    async function loadSuggestions() {
        loading = true;
        try {
            // Fetch suggestions (now includes block_details and profile_name from API)
            const suggestionsData = await fetchApi("/ml/suggestions");
            suggestions = suggestionsData || [];

            // Fetch profiles for fallback name lookup
            const profilesData = await fetchApi("/profiles");
            profiles = profilesData || [];

            // Fill in profile_name if not provided by API (fallback)
            for (const s of suggestions) {
                if (!s.profile_name && s.payload?.predicted_profile_id) {
                    const profile = profiles.find(
                        (p) => p.profile_id === s.payload.predicted_profile_id,
                    );
                    if (profile) {
                        // Use client_name if profile.name is null/undefined
                        const displayName =
                            profile.name ||
                            profile.client_name ||
                            `Profile #${profile.profile_id}`;
                        s.profile_name =
                            profile.client_name && profile.name
                                ? `${profile.name} (${profile.client_name})`
                                : displayName;
                    }
                }
            }

            // Trigger reactivity
            suggestions = [...suggestions];
        } catch (e) {
            console.error("Failed to load suggestions:", e);
        } finally {
            loading = false;
        }
    }

    async function acceptSuggestion(suggestionId: number) {
        processing = suggestionId;
        try {
            await fetchApi("/ml/suggestions/accept", {
                method: "POST",
                body: JSON.stringify({ suggestion_id: suggestionId }),
            });
            // Remove from list
            suggestions = suggestions.filter(
                (s) => s.suggestion_id !== suggestionId,
            );
        } catch (e) {
            console.error("Failed to accept suggestion:", e);
            alert("Failed to accept suggestion");
        } finally {
            processing = null;
        }
    }

    async function rejectSuggestion(suggestionId: number) {
        processing = suggestionId;
        try {
            await fetchApi("/ml/suggestions/reject", {
                method: "POST",
                body: JSON.stringify({ suggestion_id: suggestionId }),
            });
            // Remove from list
            suggestions = suggestions.filter(
                (s) => s.suggestion_id !== suggestionId,
            );
        } catch (e) {
            console.error("Failed to reject suggestion:", e);
            alert("Failed to reject suggestion");
        } finally {
            processing = null;
        }
    }

    async function acceptAll(confidenceLevel: string) {
        let toAccept: Suggestion[] = [];
        if (confidenceLevel === "HIGH") toAccept = highConfidence;
        else if (confidenceLevel === "MEDIUM") toAccept = mediumConfidence;
        else toAccept = lowConfidence;

        for (const s of toAccept) {
            await acceptSuggestion(s.suggestion_id);
        }
    }

    async function acceptDeleteSuggestion(suggestion: Suggestion) {
        processing = suggestion.suggestion_id;
        try {
            // Delete the actual block
            await fetchApi(`/blocks/${suggestion.entity_id}`, {
                method: "DELETE",
            });
            // Then mark suggestion as accepted
            await fetchApi("/ml/suggestions/accept", {
                method: "POST",
                body: JSON.stringify({
                    suggestion_id: suggestion.suggestion_id,
                }),
            });
            suggestions = suggestions.filter(
                (s) => s.suggestion_id !== suggestion.suggestion_id,
            );
        } catch (e) {
            console.error("Failed to delete block:", e);
            alert("Failed to delete block");
        } finally {
            processing = null;
        }
    }

    function getConfidenceColor(confidence: number): string {
        if (confidence >= 0.85)
            return "bg-green-100 text-green-700 border-green-200";
        if (confidence >= 0.6)
            return "bg-amber-100 text-amber-700 border-amber-200";
        return "bg-red-100 text-red-700 border-red-200";
    }

    function getConfidenceLabel(confidence: number): string {
        if (confidence >= 0.85) return "HIGH";
        if (confidence >= 0.6) return "MEDIUM";
        return "LOW";
    }

    onMount(() => loadSuggestions());
</script>

<div class="max-w-7xl mx-auto space-y-6 pb-20">
    <!-- Header -->
    <header class="flex justify-between items-start">
        <div>
            <h1
                class="text-2xl font-bold text-slate-900 dark:text-slate-100 flex items-center gap-3"
            >
                ML Suggestions
                <span
                    class="px-2 py-0.5 rounded-full bg-purple-100 text-purple-700 text-xs font-bold uppercase"
                    >Beta</span
                >
            </h1>
            <p class="text-slate-500 dark:text-slate-400">
                Review AI-generated profile suggestions. Accept to apply, reject
                to dismiss.
            </p>
        </div>
        <button
            on:click={loadSuggestions}
            class="text-sm text-blue-600 hover:text-blue-800 font-medium"
        >
            Refresh
        </button>
    </header>

    {#if loading}
        <div
            class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-12 text-center text-slate-500 dark:text-slate-400"
        >
            Loading suggestions...
        </div>
    {:else if suggestions.length === 0}
        <div
            class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-12 text-center"
        >
            <div
                class="w-12 h-12 bg-purple-100 text-purple-600 rounded-full flex items-center justify-center mx-auto mb-3"
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
                        d="M13 10V3L4 14h7v7l9-11h-7z"
                    />
                </svg>
            </div>
            <h3 class="text-lg font-medium text-slate-900 dark:text-slate-100">
                No suggestions
            </h3>
            <p class="text-slate-500 dark:text-slate-400">
                Train the ML model from the Review page to generate suggestions.
            </p>
        </div>
    {:else}
        <div class="space-y-8">
            <!-- High Confidence -->
            {#if highConfidence.length > 0}
                <section>
                    <div class="flex items-center justify-between mb-4">
                        <div class="flex items-center gap-3">
                            <span
                                class="px-2.5 py-1 rounded-md text-xs font-semibold bg-green-100 text-green-800 uppercase tracking-wider"
                            >
                                High Confidence ({highConfidence.length})
                            </span>
                            <span
                                class="text-sm text-slate-500 dark:text-slate-400"
                                >≥85% match probability</span
                            >
                        </div>
                        <button
                            on:click={() => acceptAll("HIGH")}
                            class="px-3 py-1.5 text-xs font-medium bg-green-600 text-white rounded-md hover:bg-green-700"
                        >
                            Accept All
                        </button>
                    </div>
                    <div class="grid grid-cols-1 gap-4">
                        {#each highConfidence as suggestion (suggestion.suggestion_id)}
                            <div
                                class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 shadow-sm p-5 hover:shadow-md transition-shadow"
                            >
                                <div
                                    class="flex justify-between items-start gap-4"
                                >
                                    <div class="flex-1 min-w-0">
                                        <div
                                            class="flex items-center gap-2 mb-2"
                                        >
                                            <span
                                                class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium {getConfidenceColor(
                                                    suggestion.confidence,
                                                )} border"
                                            >
                                                {(
                                                    suggestion.confidence * 100
                                                ).toFixed(0)}%
                                            </span>
                                            {#if suggestion.block_details}
                                                <span
                                                    class="text-xs text-slate-400"
                                                >
                                                    {suggestion.block_details
                                                        .app_name}
                                                </span>
                                            {/if}
                                        </div>
                                        <h3
                                            class="font-medium text-slate-900 dark:text-slate-100 mb-1"
                                        >
                                            {suggestion.block_details
                                                ?.title_summary ||
                                                `Block #${suggestion.entity_id}`}
                                        </h3>
                                        <div
                                            class="flex items-center gap-4 text-sm text-slate-500"
                                        >
                                            {#if suggestion.block_details}
                                                <span
                                                    >{suggestion.block_details.duration_minutes.toFixed(
                                                        0,
                                                    )}
                                                    min</span
                                                >
                                                <span
                                                    >{format(
                                                        parseISO(
                                                            suggestion
                                                                .block_details
                                                                .ts_start,
                                                        ),
                                                        "MMM d, HH:mm",
                                                    )}</span
                                                >
                                            {/if}
                                        </div>
                                        <div
                                            class="mt-3 flex items-center gap-2"
                                        >
                                            <span
                                                class="text-sm text-slate-500 dark:text-slate-400"
                                                >Suggested:</span
                                            >
                                            <span
                                                class="px-2 py-1 bg-purple-50 text-purple-700 rounded text-sm font-medium"
                                            >
                                                {suggestion.profile_name ||
                                                    `Profile #${suggestion.payload.predicted_profile_id}`}
                                            </span>
                                        </div>
                                    </div>
                                    <div class="flex gap-2">
                                        <button
                                            on:click={() =>
                                                rejectSuggestion(
                                                    suggestion.suggestion_id,
                                                )}
                                            disabled={processing ===
                                                suggestion.suggestion_id}
                                            class="px-3 py-2 text-sm font-medium text-slate-600 bg-slate-100 rounded-lg hover:bg-slate-200 disabled:opacity-50"
                                        >
                                            Reject
                                        </button>
                                        <button
                                            on:click={() =>
                                                acceptSuggestion(
                                                    suggestion.suggestion_id,
                                                )}
                                            disabled={processing ===
                                                suggestion.suggestion_id}
                                            class="px-3 py-2 text-sm font-medium text-white bg-green-600 rounded-lg hover:bg-green-700 disabled:opacity-50"
                                        >
                                            Accept
                                        </button>
                                    </div>
                                </div>
                            </div>
                        {/each}
                    </div>
                </section>
            {/if}

            <!-- Medium Confidence -->
            {#if mediumConfidence.length > 0}
                <section>
                    <div class="flex items-center justify-between mb-4">
                        <div class="flex items-center gap-3">
                            <span
                                class="px-2.5 py-1 rounded-md text-xs font-semibold bg-amber-100 text-amber-800 uppercase tracking-wider"
                            >
                                Medium Confidence ({mediumConfidence.length})
                            </span>
                            <span
                                class="text-sm text-slate-500 dark:text-slate-400"
                                >60-85% match probability</span
                            >
                        </div>
                    </div>
                    <div class="grid grid-cols-1 gap-4">
                        {#each mediumConfidence as suggestion (suggestion.suggestion_id)}
                            <div
                                class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 shadow-sm p-5 hover:shadow-md transition-shadow"
                            >
                                <div
                                    class="flex justify-between items-start gap-4"
                                >
                                    <div class="flex-1 min-w-0">
                                        <div
                                            class="flex items-center gap-2 mb-2"
                                        >
                                            <span
                                                class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium {getConfidenceColor(
                                                    suggestion.confidence,
                                                )} border"
                                            >
                                                {(
                                                    suggestion.confidence * 100
                                                ).toFixed(0)}%
                                            </span>
                                            {#if suggestion.block_details}
                                                <span
                                                    class="text-xs text-slate-400"
                                                >
                                                    {suggestion.block_details
                                                        .app_name}
                                                </span>
                                            {/if}
                                        </div>
                                        <h3
                                            class="font-medium text-slate-900 mb-1"
                                        >
                                            {suggestion.block_details
                                                ?.title_summary ||
                                                `Block #${suggestion.entity_id}`}
                                        </h3>
                                        <div
                                            class="flex items-center gap-4 text-sm text-slate-500"
                                        >
                                            {#if suggestion.block_details}
                                                <span
                                                    >{suggestion.block_details.duration_minutes.toFixed(
                                                        0,
                                                    )}
                                                    min</span
                                                >
                                                <span
                                                    >{format(
                                                        parseISO(
                                                            suggestion
                                                                .block_details
                                                                .ts_start,
                                                        ),
                                                        "MMM d, HH:mm",
                                                    )}</span
                                                >
                                            {/if}
                                        </div>
                                        <div
                                            class="mt-3 flex items-center gap-2"
                                        >
                                            <span
                                                class="text-sm text-slate-500 dark:text-slate-400"
                                                >Suggested:</span
                                            >
                                            <span
                                                class="px-2 py-1 bg-purple-50 text-purple-700 rounded text-sm font-medium"
                                            >
                                                {suggestion.profile_name ||
                                                    `Profile #${suggestion.payload.predicted_profile_id}`}
                                            </span>
                                        </div>
                                    </div>
                                    <div class="flex gap-2">
                                        <button
                                            on:click={() =>
                                                rejectSuggestion(
                                                    suggestion.suggestion_id,
                                                )}
                                            disabled={processing ===
                                                suggestion.suggestion_id}
                                            class="px-3 py-2 text-sm font-medium text-slate-600 bg-slate-100 rounded-lg hover:bg-slate-200 disabled:opacity-50"
                                        >
                                            Reject
                                        </button>
                                        <button
                                            on:click={() =>
                                                acceptSuggestion(
                                                    suggestion.suggestion_id,
                                                )}
                                            disabled={processing ===
                                                suggestion.suggestion_id}
                                            class="px-3 py-2 text-sm font-medium text-white bg-amber-600 rounded-lg hover:bg-amber-700 disabled:opacity-50"
                                        >
                                            Accept
                                        </button>
                                    </div>
                                </div>
                            </div>
                        {/each}
                    </div>
                </section>
            {/if}

            <!-- Low Confidence -->
            {#if lowConfidence.length > 0}
                <section>
                    <div class="flex items-center justify-between mb-4">
                        <div class="flex items-center gap-3">
                            <span
                                class="px-2.5 py-1 rounded-md text-xs font-semibold bg-red-100 text-red-800 uppercase tracking-wider"
                            >
                                Low Confidence ({lowConfidence.length})
                            </span>
                            <span
                                class="text-sm text-slate-500 dark:text-slate-400"
                                >&lt;60% match probability</span
                            >
                        </div>
                    </div>
                    <div class="grid grid-cols-1 gap-4">
                        {#each lowConfidence as suggestion (suggestion.suggestion_id)}
                            <div
                                class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 shadow-sm p-5 hover:shadow-md transition-shadow"
                            >
                                <div
                                    class="flex justify-between items-start gap-4"
                                >
                                    <div class="flex-1 min-w-0">
                                        <div
                                            class="flex items-center gap-2 mb-2"
                                        >
                                            <span
                                                class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium {getConfidenceColor(
                                                    suggestion.confidence,
                                                )} border"
                                            >
                                                {(
                                                    suggestion.confidence * 100
                                                ).toFixed(0)}%
                                            </span>
                                            {#if suggestion.block_details}
                                                <span
                                                    class="text-xs text-slate-400"
                                                >
                                                    {suggestion.block_details
                                                        .app_name}
                                                </span>
                                            {/if}
                                        </div>
                                        <h3
                                            class="font-medium text-slate-900 mb-1"
                                        >
                                            {suggestion.block_details
                                                ?.title_summary ||
                                                `Block #${suggestion.entity_id}`}
                                        </h3>
                                        <div
                                            class="flex items-center gap-4 text-sm text-slate-500"
                                        >
                                            {#if suggestion.block_details}
                                                <span
                                                    >{suggestion.block_details.duration_minutes.toFixed(
                                                        0,
                                                    )}
                                                    min</span
                                                >
                                                <span
                                                    >{format(
                                                        parseISO(
                                                            suggestion
                                                                .block_details
                                                                .ts_start,
                                                        ),
                                                        "MMM d, HH:mm",
                                                    )}</span
                                                >
                                            {/if}
                                        </div>
                                        <div
                                            class="mt-3 flex items-center gap-2"
                                        >
                                            <span
                                                class="text-sm text-slate-500 dark:text-slate-400"
                                                >Suggested:</span
                                            >
                                            <span
                                                class="px-2 py-1 bg-purple-50 text-purple-700 rounded text-sm font-medium"
                                            >
                                                {suggestion.profile_name ||
                                                    `Profile #${suggestion.payload.predicted_profile_id}`}
                                            </span>
                                        </div>
                                    </div>
                                    <div class="flex gap-2">
                                        <button
                                            on:click={() =>
                                                rejectSuggestion(
                                                    suggestion.suggestion_id,
                                                )}
                                            disabled={processing ===
                                                suggestion.suggestion_id}
                                            class="px-3 py-2 text-sm font-medium text-slate-600 bg-slate-100 rounded-lg hover:bg-slate-200 disabled:opacity-50"
                                        >
                                            Reject
                                        </button>
                                        <button
                                            on:click={() =>
                                                acceptSuggestion(
                                                    suggestion.suggestion_id,
                                                )}
                                            disabled={processing ===
                                                suggestion.suggestion_id}
                                            class="px-3 py-2 text-sm font-medium text-white bg-red-600 rounded-lg hover:bg-red-700 disabled:opacity-50"
                                        >
                                            Accept
                                        </button>
                                    </div>
                                </div>
                            </div>
                        {/each}
                    </div>
                </section>
            {/if}

            <!-- Delete Suggestions -->
            {#if deleteSuggestions.length > 0}
                <section>
                    <div class="flex items-center justify-between mb-4">
                        <div class="flex items-center gap-3">
                            <span
                                class="px-2.5 py-1 rounded-md text-xs font-semibold bg-red-100 text-red-800 uppercase tracking-wider"
                            >
                                Suggested Deletions ({deleteSuggestions.length})
                            </span>
                            <span
                                class="text-sm text-slate-500 dark:text-slate-400"
                                >Learned from your previous deletions</span
                            >
                        </div>
                    </div>
                    <div class="grid grid-cols-1 gap-4">
                        {#each deleteSuggestions as suggestion (suggestion.suggestion_id)}
                            <div
                                class="bg-white dark:bg-slate-800 rounded-xl border border-red-200 dark:border-red-800 shadow-sm p-5 hover:shadow-md transition-shadow"
                            >
                                <div
                                    class="flex justify-between items-start gap-4"
                                >
                                    <div class="flex-1 min-w-0">
                                        <div
                                            class="flex items-center gap-2 mb-2"
                                        >
                                            <span
                                                class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-red-100 text-red-700 border border-red-200"
                                            >
                                                {(
                                                    suggestion.confidence * 100
                                                ).toFixed(0)}% match
                                            </span>
                                            {#if suggestion.block_details}
                                                <span
                                                    class="text-xs text-slate-400"
                                                >
                                                    {suggestion.block_details
                                                        .app_name}
                                                </span>
                                            {/if}
                                        </div>
                                        <h3
                                            class="font-medium text-slate-900 mb-1"
                                        >
                                            {suggestion.block_details
                                                ?.title_summary ||
                                                `Block #${suggestion.entity_id}`}
                                        </h3>
                                        <div
                                            class="flex items-center gap-4 text-sm text-slate-500"
                                        >
                                            {#if suggestion.block_details}
                                                <span
                                                    >{suggestion.block_details.duration_minutes.toFixed(
                                                        0,
                                                    )}
                                                    min</span
                                                >
                                                <span
                                                    >{format(
                                                        parseISO(
                                                            suggestion
                                                                .block_details
                                                                .ts_start,
                                                        ),
                                                        "MMM d, HH:mm",
                                                    )}</span
                                                >
                                            {/if}
                                        </div>
                                        <div
                                            class="mt-3 flex items-center gap-2 text-sm text-red-600"
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
                                                    d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                                                />
                                            </svg>
                                            Similar to previously deleted items
                                        </div>
                                    </div>
                                    <div class="flex gap-2">
                                        <button
                                            on:click={() =>
                                                rejectSuggestion(
                                                    suggestion.suggestion_id,
                                                )}
                                            disabled={processing ===
                                                suggestion.suggestion_id}
                                            class="px-3 py-2 text-sm font-medium text-slate-600 bg-slate-100 rounded-lg hover:bg-slate-200 disabled:opacity-50"
                                        >
                                            Keep
                                        </button>
                                        <button
                                            on:click={() =>
                                                acceptDeleteSuggestion(
                                                    suggestion,
                                                )}
                                            disabled={processing ===
                                                suggestion.suggestion_id}
                                            class="px-3 py-2 text-sm font-medium text-white bg-red-600 rounded-lg hover:bg-red-700 disabled:opacity-50"
                                        >
                                            Delete
                                        </button>
                                    </div>
                                </div>
                            </div>
                        {/each}
                    </div>
                </section>
            {/if}
        </div>
    {/if}
</div>
