<script lang="ts">
    import { onMount } from "svelte";
    import { fetchApi } from "$lib/api";
    import { settings } from "$lib/stores/settings";
    import UnifiedProfileModal from "$lib/components/profiles/UnifiedProfileModal.svelte";

    let profiles: any[] = [];
    let loading = true;
    let error: string | null = null;
    let isModalOpen = false;

    async function loadProfiles() {
        try {
            loading = true;
            // Ensure settings are loaded for currency formatting
            await settings.detectLocale();

            const res = await fetchApi("/profiles");
            profiles = res || [];
        } catch (err: any) {
            error = err.message;
        } finally {
            loading = false;
        }
    }

    async function deleteProfile(id: number) {
        if (!confirm("Are you sure you want to delete this profile?")) return;
        try {
            await fetchApi(`/profiles/${id}`, { method: "DELETE" });
            profiles = profiles.filter((p) => p.profile_id !== id);
        } catch (err: any) {
            alert(err.message);
        }
    }

    function formatCurrency(amount: number, code: string) {
        try {
            return new Intl.NumberFormat($settings.locale, {
                style: "currency",
                currency: code,
            }).format(amount);
        } catch (e) {
            return `${code} ${amount.toFixed(2)}`;
        }
    }

    function getColor(name: string) {
        const colors = [
            "bg-red-100 text-red-700",
            "bg-blue-100 text-blue-700",
            "bg-emerald-100 text-emerald-700",
            "bg-amber-100 text-amber-700",
            "bg-purple-100 text-purple-700",
            "bg-indigo-100 text-indigo-700",
            "bg-pink-100 text-pink-700",
        ];
        return colors[name.length % colors.length];
    }

    onMount(loadProfiles);
</script>

<div class="max-w-7xl mx-auto space-y-6">
    <div
        class="flex justify-between items-center bg-white p-6 rounded-2xl border border-slate-200 shadow-sm"
    >
        <div>
            <h1 class="text-2xl font-bold text-slate-900 tracking-tight">
                Profiles
            </h1>
            <p class="text-slate-500 mt-1">
                Manage your clients, projects, and billing rates in one place.
            </p>
        </div>
        <button
            on:click={() => (isModalOpen = true)}
            class="bg-blue-600 hover:bg-blue-700 text-white px-5 py-2.5 rounded-xl font-semibold shadow-lg shadow-blue-200 transition-all hover:-translate-y-0.5 flex items-center gap-2"
        >
            <svg
                class="w-5 h-5"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                ><path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M12 6v6m0 0v6m0-6h6m-6 0H6"
                ></path></svg
            >
            Create Profile
        </button>
    </div>

    {#if error}
        <div
            class="bg-red-50 text-red-700 p-4 rounded-xl border border-red-100 flex items-center gap-2"
        >
            <svg
                class="w-5 h-5"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                ><path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                /></svg
            >
            {error}
        </div>
    {/if}

    <div
        class="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden"
    >
        {#if loading}
            <div class="p-12 text-center text-slate-500 animate-pulse">
                Loading profiles...
            </div>
        {:else if profiles.length === 0}
            <div
                class="p-16 text-center flex flex-col items-center justify-center"
            >
                <div
                    class="w-16 h-16 bg-slate-100 text-slate-400 rounded-full flex items-center justify-center mb-4"
                >
                    <svg
                        class="w-8 h-8"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                        ><path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"
                        /></svg
                    >
                </div>
                <h3 class="text-lg font-bold text-slate-900">
                    No profiles yet
                </h3>
                <p class="text-slate-500 mt-1 max-w-sm">
                    Create your first profile to start tracking time against
                    clients and rates.
                </p>
                <button
                    on:click={() => (isModalOpen = true)}
                    class="mt-6 text-blue-600 font-medium hover:underline"
                    >Create Profile</button
                >
            </div>
        {:else}
            <table class="w-full text-left">
                <thead class="bg-slate-50/50 border-b border-slate-200">
                    <tr>
                        <th
                            class="px-6 py-4 text-xs font-bold text-slate-400 uppercase tracking-wider"
                            >Client & Project</th
                        >
                        <th
                            class="px-6 py-4 text-xs font-bold text-slate-400 uppercase tracking-wider"
                            >Service</th
                        >
                        <th
                            class="px-6 py-4 text-xs font-bold text-slate-400 uppercase tracking-wider"
                            >Billable Rate</th
                        >
                        <th
                            class="px-6 py-4 text-right text-xs font-bold text-slate-400 uppercase tracking-wider"
                            >Actions</th
                        >
                    </tr>
                </thead>
                <tbody class="divide-y divide-slate-100">
                    {#each profiles as p}
                        <tr class="hover:bg-slate-50 transition-colors group cursor-pointer" on:click={() => window.location.href = `/profiles/${p.profile_id}`}>
                            <td class="px-6 py-4">
                                <div class="flex items-center gap-3">
                                    <div
                                        class="w-10 h-10 rounded-full flex items-center justify-center font-bold text-sm {getColor(
                                            p.client_name,
                                        )}"
                                    >
                                        {p.client_name
                                            .slice(0, 2)
                                            .toUpperCase()}
                                    </div>
                                    <div>
                                        <div
                                            class="font-semibold text-slate-900"
                                        >
                                            {p.client_name}
                                        </div>
                                        {#if p.project_name}
                                            <div class="text-xs text-slate-500">
                                                {p.project_name}
                                            </div>
                                        {/if}
                                    </div>
                                </div>
                            </td>
                            <td class="px-6 py-4">
                                <div
                                    class="inline-flex px-2.5 py-1 rounded-md bg-slate-100 text-slate-600 text-xs font-medium border border-slate-200"
                                >
                                    {p.service_name}
                                </div>
                            </td>
                            <td class="px-6 py-4">
                                <div
                                    class="font-mono text-sm text-slate-700 font-medium tracking-tight"
                                >
                                    {formatCurrency(
                                        p.rate_amount,
                                        p.currency_code || "USD",
                                    )}
                                    <span class="text-slate-400 text-xs ml-1"
                                        >/hr</span
                                    >
                                </div>
                                <div class="text-[10px] text-slate-400 mt-0.5">
                                    {p.rate_name}
                                </div>
                            </td>
                            <td
                                class="px-6 py-4 text-right"
                            >
                                <div class="flex items-center justify-end gap-3">
                                    <a
                                        href="/profiles/{p.profile_id}"
                                        class="text-blue-600 hover:text-blue-700 font-medium text-sm hover:underline"
                                        on:click|stopPropagation
                                    >
                                        View Details
                                    </a>
                                    <button
                                        on:click|stopPropagation={() => deleteProfile(p.profile_id)}
                                        class="text-red-600 hover:text-red-700 font-medium text-sm hover:underline opacity-0 group-hover:opacity-100 transition-opacity"
                                    >
                                        Remove
                                    </button>
                                </div>
                            </td>
                        </tr>
                    {/each}
                </tbody>
            </table>
        {/if}
    </div>
</div>

<UnifiedProfileModal bind:isOpen={isModalOpen} on:success={() => { isModalOpen = false; loadProfiles(); }} />
