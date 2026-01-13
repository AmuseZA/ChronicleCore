<script lang="ts">
    import { onMount } from "svelte";
    import { fetchApi } from "$lib/api";

    let rates: any[] = [];
    let form = { name: "", amount: 0, currency: "USD" };
    let creating = false;

    async function load() {
        rates = (await fetchApi("/rates")) || [];
    }

    async function create() {
        if (!form.name.trim()) return;
        creating = true;
        try {
            await fetchApi("/rates/create", {
                method: "POST",
                body: JSON.stringify({
                    name: form.name,
                    hourly_amount: Number(form.amount),
                    currency: form.currency,
                }),
            });
            form = { name: "", amount: 0, currency: "USD" };
            await load();
        } catch (e) {
            alert("Failed to create rate");
        } finally {
            creating = false;
        }
    }

    onMount(load);
</script>

<div class="max-w-4xl mx-auto">
    <div class="mb-6 flex items-center gap-2 text-sm text-slate-500">
        <a href="/profiles" class="hover:text-blue-600">Profiles</a>
        <span>/</span>
        <span class="text-slate-900 font-medium">Rates</span>
    </div>

    <div class="bg-white rounded-xl border border-slate-200 shadow-sm p-6 mb-8">
        <h2 class="text-lg font-semibold text-slate-900 mb-4">Add New Rate</h2>
        <form on:submit|preventDefault={create} class="flex gap-4 items-end">
            <div class="flex-1">
                <label class="block text-xs font-medium text-slate-500 mb-1"
                    >Rate Name</label
                >
                <input
                    type="text"
                    bind:value={form.name}
                    placeholder="e.g. Standard, Rush"
                    class="w-full rounded-lg border border-slate-200 px-3 py-2 text-sm"
                    required
                />
            </div>
            <div class="w-32">
                <label class="block text-xs font-medium text-slate-500 mb-1"
                    >Amount</label
                >
                <input
                    type="number"
                    step="0.01"
                    bind:value={form.amount}
                    class="w-full rounded-lg border border-slate-200 px-3 py-2 text-sm"
                    required
                />
            </div>
            <div class="w-24">
                <label class="block text-xs font-medium text-slate-500 mb-1"
                    >Currency</label
                >
                <select
                    bind:value={form.currency}
                    class="w-full rounded-lg border border-slate-200 px-3 py-2 text-sm"
                >
                    <option>USD</option>
                    <option>EUR</option>
                    <option>GBP</option>
                </select>
            </div>
            <button
                type="submit"
                disabled={creating}
                class="bg-slate-900 text-white px-6 py-2 rounded-lg font-medium hover:bg-slate-800 disabled:opacity-50"
            >
                Add
            </button>
        </form>
    </div>

    <div
        class="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden"
    >
        <table class="w-full text-left">
            <thead class="bg-slate-50 border-b border-slate-200">
                <tr>
                    <th
                        class="px-6 py-3 text-xs font-medium text-slate-500 uppercase"
                        >ID</th
                    >
                    <th
                        class="px-6 py-3 text-xs font-medium text-slate-500 uppercase"
                        >Name</th
                    >
                    <th
                        class="px-6 py-3 text-xs font-medium text-slate-500 uppercase"
                        >Rate</th
                    >
                </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
                {#each rates as r}
                    <tr>
                        <td class="px-6 py-4 text-slate-400 font-mono text-xs"
                            >#{r.rate_id}</td
                        >
                        <td class="px-6 py-4 font-medium text-slate-900"
                            >{r.name}</td
                        >
                        <td class="px-6 py-4 text-slate-900 font-mono"
                            >{r.hourly_amount} {r.currency}</td
                        >
                    </tr>
                {/each}
            </tbody>
        </table>
    </div>
</div>
