<script lang="ts">
    import { onMount } from "svelte";
    import { fetchApi } from "$lib/api";

    export let value: number | null = null;
    export let disabled = false;
    export let placeholder = "Select Profile...";

    let profiles: any[] = [];
    let loading = true;

    onMount(async () => {
        try {
            const res = await fetchApi("/profiles");
            profiles = res || [];
        } catch (e) {
            console.error(e);
        } finally {
            loading = false;
        }
    });

    function handleChange(e: Event) {
        const target = e.target as HTMLSelectElement;
        const val = parseInt(target.value);
        value = isNaN(val) ? null : val;
        dispatch("change", value);
    }

    import { createEventDispatcher } from "svelte";
    const dispatch = createEventDispatcher();
</script>

<div class="relative">
    <select
        class="w-full pl-3 pr-10 py-2 text-sm border-slate-200 rounded-lg shadow-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 disabled:bg-slate-50 disabled:text-slate-500"
        {disabled}
        value={value ?? ""}
        on:change={handleChange}
    >
        <option value="">{placeholder}</option>
        {#each profiles as p}
            <option value={p.profile_id}>
                {p.client_name} - {p.service_name} ({p.currency}
                {p.rate_amount}/hr)
            </option>
        {/each}
    </select>
    {#if loading}
        <div class="absolute right-3 top-2.5">
            <div
                class="w-4 h-4 border-2 border-slate-200 border-t-blue-500 rounded-full animate-spin"
            ></div>
        </div>
    {/if}
</div>
