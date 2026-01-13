<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import SuggestionBadge from "./SuggestionBadge.svelte";
    import type { Suggestion } from "$lib/stores/suggestions";

    export let suggestion: Suggestion | undefined;
    export let blockId: number;

    const dispatch = createEventDispatcher();

    function apply() {
        if (suggestion) {
            dispatch("apply", { blockId, profileId: suggestion.profile_id });
        }
    }
</script>

<div class="flex items-center justify-between min-h-[3rem]">
    {#if suggestion}
        <div class="flex flex-col">
            <span class="text-sm font-medium text-slate-900"
                >{suggestion.profile_name}</span
            >
            <div class="flex items-center gap-2 mt-0.5">
                <SuggestionBadge score={suggestion.score} />
                {#if suggestion.score >= 0.85}
                    <span
                        class="text-[10px] text-emerald-600 font-bold uppercase tracking-wider"
                        >Top Match</span
                    >
                {/if}
            </div>
        </div>

        {#if suggestion.score >= 0.85}
            <button
                on:click={apply}
                class="ml-4 p-1.5 rounded-lg bg-emerald-50 text-emerald-600 hover:bg-emerald-100 hover:text-emerald-700 transition-colors"
                title="Apply Suggestion"
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
                        d="M5 13l4 4L19 7"
                    />
                </svg>
            </button>
        {/if}
    {:else}
        <span class="text-slate-300 text-sm">—</span>
    {/if}
</div>
