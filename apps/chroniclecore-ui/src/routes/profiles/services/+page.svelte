<script lang="ts">
    import { onMount } from "svelte";
    import { fetchApi } from "$lib/api";

    let services: any[] = [];
    let newName = "";
    let creating = false;

    async function load() {
        services = (await fetchApi("/services")) || [];
    }

    async function create() {
        if (!newName.trim()) return;
        creating = true;
        try {
            await fetchApi("/services/create", {
                method: "POST",
                body: JSON.stringify({ name: newName }),
            });
            newName = "";
            await load();
        } catch (e) {
            alert("Failed to create service");
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
        <span class="text-slate-900 font-medium">Services</span>
    </div>

    <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 shadow-sm p-6 mb-8">
        <h2 class="text-lg font-semibold text-slate-900 mb-4">
            Add New Service
        </h2>
        <form on:submit|preventDefault={create} class="flex gap-4">
            <input
                type="text"
                bind:value={newName}
                placeholder="Service Name (e.g. Development, Design)"
                class="flex-1 rounded-lg border border-slate-200 px-4 py-2 focus:ring-2 focus:ring-blue-500 outline-none"
                required
            />
            <button
                type="submit"
                disabled={creating}
                class="bg-slate-900 text-white px-6 py-2 rounded-lg font-medium hover:bg-slate-800 disabled:opacity-50"
            >
                {creating ? "Saving..." : "Create Service"}
            </button>
        </form>
    </div>

    <div
        class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 shadow-sm overflow-hidden"
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
                        >Created</th
                    >
                </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
                {#each services as s}
                    <tr>
                        <td class="px-6 py-4 text-slate-400 font-mono text-xs"
                            >#{s.service_id}</td
                        >
                        <td class="px-6 py-4 font-medium text-slate-900"
                            >{s.name}</td
                        >
                        <td class="px-6 py-4 text-slate-500 text-sm"
                            >{new Date(s.created_at).toLocaleDateString()}</td
                        >
                    </tr>
                {/each}
            </tbody>
        </table>
    </div>
</div>
