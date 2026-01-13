<script lang="ts">
    import { createEventDispatcher, onMount } from "svelte";
    import { fetchApi } from "$lib/api";
    import CurrencySelect from "$lib/components/inputs/CurrencySelect.svelte";
    import { settings } from "$lib/stores/settings";

    export let isOpen = false;
    const dispatch = createEventDispatcher();

    // Data Sources
    let clients: any[] = [];
    let services: any[] = [];

    // Form State
    let clientName = "";
    let serviceName = ""; // Or ID if existing
    let rateAmount = 0;
    let currencyCode = $settings.currencyCode;
    let selectedClientId: number | null = null;
    let selectedServiceId: number | null = null;

    let isLoading = false;
    let error: string | null = null;
    let clientInputFocused = false;
    let serviceInputFocused = false;
    let wasModalOpen = false;

    $: if (isOpen && !wasModalOpen) {
        wasModalOpen = true;
        loadDependencies();
        // Reset form
        clientName = "";
        selectedClientId = null;
        selectedServiceId = null;
        serviceName = "";
        rateAmount = 0;
        if ($settings.currencyCode) {
            currencyCode = $settings.currencyCode;
        }
    } else if (!isOpen) {
        wasModalOpen = false;
    }

    async function loadDependencies() {
        try {
            const [cRes, sRes] = await Promise.all([
                fetchApi("/clients"),
                fetchApi("/services"),
            ]);
            clients = cRes || [];
            services = sRes || [];
        } catch (e) {
            console.error("Failed to load dependencies", e);
        }
    }

    async function handleSubmit() {
        isLoading = true;
        error = null;

        try {
            // 1. Get/Create Client
            let clientId = selectedClientId;
            if (!clientId && clientName) {
                // Check if name matches existing exactly to avoid dupes
                const existing = clients.find(
                    (c) => c.name.toLowerCase() === clientName.toLowerCase(),
                );
                if (existing) {
                    clientId = existing.client_id;
                } else {
                    const newClient = await fetchApi("/clients/create", {
                        method: "POST",
                        body: JSON.stringify({ name: clientName }),
                    });
                    clientId = newClient.client_id;
                }
            }

            if (!clientId) throw new Error("Client is required");

            // 2. Get/Create Service
            let serviceId = selectedServiceId;
            if (!serviceId && serviceName) {
                const existing = services.find(
                    (s) => s.name.toLowerCase() === serviceName.toLowerCase(),
                );
                if (existing) {
                    serviceId = existing.service_id;
                } else {
                    const newService = await fetchApi("/services/create", {
                        method: "POST",
                        body: JSON.stringify({ name: serviceName }),
                    });
                    serviceId = newService.service_id;
                }
            }

            if (!serviceId) throw new Error("Service is required");

            // 3. Create Rate (Always specific to profile for this simplified flow? Or reuse?)
            // Plan says: "Rate Section... Amount input... Currency selector"
            // We create a new rate for this relationship usually.
            const rateRes = await fetchApi("/rates/create", {
                method: "POST",
                body: JSON.stringify({
                    name: `Rate for ${clientName}`,
                    currency_code: currencyCode,
                    hourly_amount: rateAmount,
                }),
            });

            // 4. Create Profile
            await fetchApi("/profiles", {
                method: "POST",
                body: JSON.stringify({
                    client_id: clientId,
                    service_id: serviceId,
                    rate_id: rateRes.rate_id,
                }),
            });

            dispatch("success");
            isOpen = false;
        } catch (e: any) {
            error = e.message || "Failed to create profile";
        } finally {
            isLoading = false;
        }
    }

    // Filtered Client Search
    let clientSearch = "";
    $: filteredClients = clients.filter((c) =>
        c.name.toLowerCase().includes(clientName.toLowerCase()),
    );

    function selectClient(client: any) {
        selectedClientId = client.client_id;
        clientName = client.name;
    }

    // Filtered Service Search
    $: filteredServices = services.filter((s) =>
        s.name.toLowerCase().includes(serviceName.toLowerCase()),
    );

    function selectService(service: any) {
        selectedServiceId = service.service_id;
        serviceName = service.name;
    }
</script>

{#if isOpen}
    <div
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/50 backdrop-blur-sm"
    >
        <div
            class="bg-white rounded-2xl shadow-xl w-full max-w-lg overflow-hidden"
        >
            <div
                class="px-6 py-4 border-b border-slate-100 flex justify-between items-center"
            >
                <h2 class="text-lg font-bold text-slate-900">New Profile</h2>
                <button
                    class="text-slate-400 hover:text-slate-600"
                    on:click={() => (isOpen = false)}
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
                            d="M6 18L18 6M6 6l12 12"
                        /></svg
                    >
                </button>
            </div>

            <form on:submit|preventDefault={handleSubmit} class="p-6 space-y-6">
                {#if error}
                    <div
                        class="bg-red-50 text-red-600 text-sm p-3 rounded-lg border border-red-100"
                    >
                        {error}
                    </div>
                {/if}

                <!-- Client Input with Auto-complete -->
                <div>
                    <label
                        for="client-input"
                        class="block text-sm font-medium text-slate-700 mb-1"
                        >Client</label
                    >
                    <div class="relative">
                        <input
                            id="client-input"
                            type="text"
                            bind:value={clientName}
                            on:input={() => (selectedClientId = null)}
                            on:focus={() => (clientInputFocused = true)}
                            on:blur={() =>
                                setTimeout(
                                    () => (clientInputFocused = false),
                                    200,
                                )}
                            placeholder="e.g. Acme Corp"
                            class="w-full rounded-lg border-slate-300 focus:ring-blue-500 focus:border-blue-500"
                            required
                            autocomplete="off"
                        />
                        <!-- Dropdown suggestions -->
                        {#if clientInputFocused && !selectedClientId && filteredClients.length > 0}
                            <div
                                class="absolute z-10 w-full bg-white border border-slate-200 mt-1 rounded-lg shadow-lg max-h-40 overflow-auto"
                            >
                                {#each filteredClients as fc}
                                    <button
                                        type="button"
                                        class="w-full text-left px-4 py-2 hover:bg-slate-50 text-sm"
                                        on:click={() => selectClient(fc)}
                                    >
                                        {fc.name}
                                    </button>
                                {/each}
                            </div>
                        {/if}
                        {#if clientName && !selectedClientId && filteredClients.length === 0}
                            <div
                                class="absolute right-3 top-2.5 text-xs text-blue-600 font-medium pointer-events-none"
                            >
                                New Client
                            </div>
                        {/if}
                    </div>
                </div>

                <!-- Service Input with Auto-complete -->
                <div>
                    <label
                        for="service-input"
                        class="block text-sm font-medium text-slate-700 mb-1"
                        >Service</label
                    >
                    <div class="relative">
                        <input
                            id="service-input"
                            type="text"
                            bind:value={serviceName}
                            on:input={() => (selectedServiceId = null)}
                            on:focus={() => (serviceInputFocused = true)}
                            on:blur={() =>
                                setTimeout(
                                    () => (serviceInputFocused = false),
                                    200,
                                )}
                            placeholder="e.g. Consulting"
                            class="w-full rounded-lg border-slate-300 focus:ring-blue-500 focus:border-blue-500"
                            required
                            autocomplete="off"
                        />
                        <!-- Dropdown suggestions -->
                        {#if serviceInputFocused && !selectedServiceId}
                            <div
                                class="absolute z-10 w-full bg-white border border-slate-200 mt-1 rounded-lg shadow-lg max-h-40 overflow-auto"
                            >
                                {#if filteredServices.length > 0}
                                    {#each filteredServices as fs}
                                        <button
                                            type="button"
                                            class="w-full text-left px-4 py-2 hover:bg-slate-50 text-sm"
                                            on:click={() => selectService(fs)}
                                        >
                                            {fs.name}
                                        </button>
                                    {/each}
                                {/if}

                                {#if serviceName && filteredServices.length === 0}
                                    <div
                                        class="px-4 py-2 text-sm text-blue-600 bg-blue-50 font-medium"
                                    >
                                        Create new service "{serviceName}"
                                    </div>
                                {/if}
                            </div>
                        {/if}
                    </div>
                    <p class="text-xs text-slate-400 mt-1">
                        Defines the type of work.
                    </p>
                </div>

                <!-- Rate & Currency -->
                <div class="grid grid-cols-2 gap-4">
                    <div>
                        <label
                            class="block text-sm font-medium text-slate-700 mb-1"
                            >Hourly Rate</label
                        >
                        <input
                            type="number"
                            bind:value={rateAmount}
                            min="0"
                            step="0.01"
                            class="w-full rounded-lg border-slate-300 focus:ring-blue-500 focus:border-blue-500"
                        />
                    </div>
                    <div>
                        <label
                            class="block text-sm font-medium text-slate-700 mb-1"
                            >Currency</label
                        >
                        <CurrencySelect bind:value={currencyCode} />
                    </div>
                </div>

                <div
                    class="pt-4 flex justify-end gap-3 border-t border-slate-50"
                >
                    <button
                        type="button"
                        class="px-4 py-2 text-slate-600 font-medium hover:bg-slate-50 rounded-lg"
                        on:click={() => (isOpen = false)}>Cancel</button
                    >
                    <button
                        type="submit"
                        class="px-6 py-2 bg-blue-600 text-white font-medium rounded-lg hover:bg-blue-700 shadow-sm shadow-blue-200 disabled:opacity-70"
                        disabled={isLoading}
                    >
                        {isLoading ? "Creating..." : "Create Profile"}
                    </button>
                </div>
            </form>
        </div>
    </div>
{/if}
