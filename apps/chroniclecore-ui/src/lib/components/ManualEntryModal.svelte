<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { fetchApi } from "$lib/api";
    import ProfileSelector from "./ProfileSelector.svelte";

    export let isOpen = false;
    export let preselectedProfileId: number | null = null;

    const dispatch = createEventDispatcher();

    let profileId: number | null = preselectedProfileId;
    let title = "";
    let description = "";
    let date = new Date().toISOString().split("T")[0]; // Today
    let startTime = "09:00";
    let endTime = "10:00";
    let billable = true;
    let loading = false;
    let error = "";

    $: if (preselectedProfileId) {
        profileId = preselectedProfileId;
    }

    function close() {
        isOpen = false;
        resetForm();
        dispatch("close");
    }

    function resetForm() {
        title = "";
        description = "";
        date = new Date().toISOString().split("T")[0];
        startTime = "09:00";
        endTime = "10:00";
        billable = true;
        error = "";
        if (!preselectedProfileId) {
            profileId = null;
        }
    }

    async function submit() {
        error = "";

        if (!profileId) {
            error = "Please select a profile";
            return;
        }
        if (!title.trim()) {
            error = "Please enter a title";
            return;
        }

        // Build ISO timestamps
        const tsStart = `${date}T${startTime}:00Z`;
        const tsEnd = `${date}T${endTime}:00Z`;

        // Validate end > start
        if (new Date(tsEnd) <= new Date(tsStart)) {
            error = "End time must be after start time";
            return;
        }

        loading = true;

        try {
            const result = await fetchApi("/blocks/manual", {
                method: "POST",
                body: JSON.stringify({
                    profile_id: profileId,
                    ts_start: tsStart,
                    ts_end: tsEnd,
                    title: title.trim(),
                    description: description.trim(),
                    billable,
                }),
            });

            dispatch("created", result);
            close();
        } catch (e: any) {
            error = e.message || "Failed to create manual entry";
        } finally {
            loading = false;
        }
    }

    function handleKeydown(e: KeyboardEvent) {
        if (e.key === "Escape") {
            close();
        }
    }
</script>

<svelte:window on:keydown={handleKeydown} />

{#if isOpen}
    <!-- Backdrop -->
    <div
        class="fixed inset-0 bg-black/50 z-40"
        on:click={close}
        on:keypress={() => {}}
        role="button"
        tabindex="-1"
    ></div>

    <!-- Modal -->
    <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div
            class="bg-white rounded-2xl shadow-2xl w-full max-w-lg overflow-hidden"
            on:click|stopPropagation={() => {}}
            on:keypress={() => {}}
            role="dialog"
            tabindex="-1"
        >
            <!-- Header -->
            <div
                class="px-6 py-4 border-b border-slate-200 flex items-center justify-between"
            >
                <h2 class="text-lg font-semibold text-slate-900">
                    Add Manual Time Entry
                </h2>
                <button
                    on:click={close}
                    class="p-1 rounded-lg hover:bg-slate-100 text-slate-400 hover:text-slate-600 transition-colors"
                >
                    <svg
                        class="w-5 h-5"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                    >
                        <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M6 18L18 6M6 6l12 12"
                        />
                    </svg>
                </button>
            </div>

            <!-- Body -->
            <form on:submit|preventDefault={submit} class="p-6 space-y-4">
                {#if error}
                    <div
                        class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm"
                    >
                        {error}
                    </div>
                {/if}

                <!-- Profile -->
                <div>
                    <label
                        class="block text-sm font-medium text-slate-700 mb-1"
                        for="profile"
                    >
                        Profile <span class="text-red-500">*</span>
                    </label>
                    <ProfileSelector
                        bind:value={profileId}
                        placeholder="Select a profile..."
                    />
                </div>

                <!-- Title -->
                <div>
                    <label
                        class="block text-sm font-medium text-slate-700 mb-1"
                        for="title"
                    >
                        Title <span class="text-red-500">*</span>
                    </label>
                    <input
                        id="title"
                        type="text"
                        bind:value={title}
                        placeholder="e.g., Phone call with Client ABC"
                        class="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 text-sm"
                    />
                </div>

                <!-- Description -->
                <div>
                    <label
                        class="block text-sm font-medium text-slate-700 mb-1"
                        for="description"
                    >
                        Description <span class="text-slate-400">(optional)</span>
                    </label>
                    <textarea
                        id="description"
                        bind:value={description}
                        placeholder="Additional details about this activity..."
                        rows="2"
                        class="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 text-sm resize-none"
                    ></textarea>
                </div>

                <!-- Date & Time -->
                <div class="grid grid-cols-3 gap-4">
                    <div>
                        <label
                            class="block text-sm font-medium text-slate-700 mb-1"
                            for="date"
                        >
                            Date
                        </label>
                        <input
                            id="date"
                            type="date"
                            bind:value={date}
                            class="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 text-sm"
                        />
                    </div>
                    <div>
                        <label
                            class="block text-sm font-medium text-slate-700 mb-1"
                            for="startTime"
                        >
                            Start Time
                        </label>
                        <input
                            id="startTime"
                            type="time"
                            bind:value={startTime}
                            class="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 text-sm"
                        />
                    </div>
                    <div>
                        <label
                            class="block text-sm font-medium text-slate-700 mb-1"
                            for="endTime"
                        >
                            End Time
                        </label>
                        <input
                            id="endTime"
                            type="time"
                            bind:value={endTime}
                            class="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 text-sm"
                        />
                    </div>
                </div>

                <!-- Billable Toggle -->
                <div class="flex items-center gap-3">
                    <button
                        type="button"
                        on:click={() => (billable = !billable)}
                        class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors {billable
                            ? 'bg-indigo-600'
                            : 'bg-slate-200'}"
                    >
                        <span
                            class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform {billable
                                ? 'translate-x-6'
                                : 'translate-x-1'}"
                        ></span>
                    </button>
                    <span class="text-sm text-slate-700">Billable</span>
                </div>
            </form>

            <!-- Footer -->
            <div
                class="px-6 py-4 bg-slate-50 border-t border-slate-200 flex justify-end gap-3"
            >
                <button
                    type="button"
                    on:click={close}
                    class="px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100 rounded-lg transition-colors"
                >
                    Cancel
                </button>
                <button
                    type="submit"
                    on:click={submit}
                    disabled={loading}
                    class="px-4 py-2 text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
                >
                    {#if loading}
                        <svg
                            class="animate-spin h-4 w-4"
                            fill="none"
                            viewBox="0 0 24 24"
                        >
                            <circle
                                class="opacity-25"
                                cx="12"
                                cy="12"
                                r="10"
                                stroke="currentColor"
                                stroke-width="4"
                            ></circle>
                            <path
                                class="opacity-75"
                                fill="currentColor"
                                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                            ></path>
                        </svg>
                    {/if}
                    Add Entry
                </button>
            </div>
        </div>
    </div>
{/if}
