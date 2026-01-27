<script lang="ts">
    import { onMount } from "svelte";
    import { fetchApi } from "$lib/api";
    import { goto } from "$app/navigation";

    let clients: any[] = [];
    let services: any[] = [];
    let rates: any[] = [];

    let form = { client_id: "", service_id: "", rate_id: "" };

    onMount(async () => {
        const [c, s, r] = await Promise.all([
            fetchApi("/clients"),
            fetchApi("/services"),
            fetchApi("/rates"),
        ]);
        clients = c || [];
        services = s || [];
        rates = r || [];
    });

    async function save() {
        try {
            await fetchApi("/profiles", {
                method: "POST",
                body: JSON.stringify({
                    client_id: Number(form.client_id),
                    service_id: Number(form.service_id),
                    rate_id: Number(form.rate_id),
                }),
            });
            goto("/profiles");
        } catch (e: any) {
            alert(e.message || "Failed to create profile");
        }
    }
</script>

<div class="max-w-xl mx-auto mt-12">
    <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 shadow-sm p-8">
        <h1 class="text-2xl font-bold text-slate-900 mb-6">
            Create New Profile
        </h1>

        <div class="space-y-6">
            <div>
                <label class="block text-sm font-medium text-slate-700 mb-2"
                    >Client</label
                >
                <select
                    bind:value={form.client_id}
                    class="w-full rounded-lg border-slate-200 px-3 py-2"
                >
                    <option value="">Select Client...</option>
                    {#each clients as c}
                        <option value={c.client_id}>{c.name}</option>
                    {/each}
                </select>
                <div class="mt-1 text-xs text-right">
                    <a
                        href="/profiles/clients"
                        class="text-blue-600 hover:underline">Add new client</a
                    >
                </div>
            </div>

            <div>
                <label class="block text-sm font-medium text-slate-700 mb-2"
                    >Service</label
                >
                <select
                    bind:value={form.service_id}
                    class="w-full rounded-lg border-slate-200 px-3 py-2"
                >
                    <option value="">Select Service...</option>
                    {#each services as s}
                        <option value={s.service_id}>{s.name}</option>
                    {/each}
                </select>
                <div class="mt-1 text-xs text-right">
                    <a
                        href="/profiles/services"
                        class="text-blue-600 hover:underline">Add new service</a
                    >
                </div>
            </div>

            <div>
                <label class="block text-sm font-medium text-slate-700 mb-2"
                    >Rate</label
                >
                <select
                    bind:value={form.rate_id}
                    class="w-full rounded-lg border-slate-200 px-3 py-2"
                >
                    <option value="">Select Rate...</option>
                    {#each rates as r}
                        <option value={r.rate_id}
                            >{r.name} ({r.hourly_amount} {r.currency})</option
                        >
                    {/each}
                </select>
                <div class="mt-1 text-xs text-right">
                    <a
                        href="/profiles/rates"
                        class="text-blue-600 hover:underline">Add new rate</a
                    >
                </div>
            </div>

            <div class="pt-6 flex gap-3">
                <button
                    on:click={save}
                    class="flex-1 bg-slate-900 text-white py-2.5 rounded-lg font-medium hover:bg-slate-800"
                >
                    Create Profile
                </button>
                <a
                    href="/profiles"
                    class="px-6 py-2.5 rounded-lg border border-slate-200 text-slate-700 hover:bg-slate-50 font-medium"
                >
                    Cancel
                </a>
            </div>
        </div>
    </div>
</div>
