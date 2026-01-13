<script lang="ts">
    import { createEventDispatcher, onMount } from "svelte";

    export let value: string = "";
    export let disabled = false;

    const dispatch = createEventDispatcher();

    // Top 20 Traded Currencies + ZAR
    const currencies = [
        { code: "USD", name: "United States Dollar", flag: "🇺🇸" },
        { code: "EUR", name: "Euro", flag: "🇪🇺" },
        { code: "JPY", name: "Japanese Yen", flag: "🇯🇵" },
        { code: "GBP", name: "British Pound", flag: "🇬🇧" },
        { code: "AUD", name: "Australian Dollar", flag: "🇦🇺" },
        { code: "CAD", name: "Canadian Dollar", flag: "🇨🇦" },
        { code: "CHF", name: "Swiss Franc", flag: "🇨🇭" },
        { code: "CNY", name: "Chinese Yuan", flag: "🇨🇳" },
        { code: "HKD", name: "Hong Kong Dollar", flag: "🇭🇰" },
        { code: "NZD", name: "New Zealand Dollar", flag: "🇳🇿" },
        { code: "SEK", name: "Swedish Krona", flag: "🇸🇪" },
        { code: "KRW", name: "South Korean Won", flag: "🇰🇷" },
        { code: "SGD", name: "Singapore Dollar", flag: "🇸🇬" },
        { code: "NOK", name: "Norwegian Krone", flag: "🇳🇴" },
        { code: "MXN", name: "Mexican Peso", flag: "🇲🇽" },
        { code: "INR", name: "Indian Rupee", flag: "🇮🇳" },
        { code: "RUB", name: "Russian Ruble", flag: "🇷🇺" },
        { code: "ZAR", name: "South African Rand", flag: "🇿🇦" },
        { code: "TRY", name: "Turkish Lira", flag: "🇹🇷" },
        { code: "BRL", name: "Brazilian Real", flag: "🇧🇷" },
    ];

    let isOpen = false;
    let search = "";

    $: filtered = currencies.filter(
        (c) =>
            c.code.includes(search.toUpperCase()) ||
            c.name.toLowerCase().includes(search.toLowerCase()),
    );

    function select(code: string) {
        value = code;
        dispatch("change", code);
        isOpen = false;
        search = "";
    }

    // Close on click outside (simplified)
    function backdropClick() {
        isOpen = false;
    }
</script>

<div class="relative">
    <button
        type="button"
        on:click={() => {
            if (!disabled) isOpen = !isOpen;
        }}
        class="w-full bg-white border border-slate-300 rounded-lg px-3 py-2 text-left flex items-center justify-between shadow-sm focus:ring-2 focus:ring-blue-500 focus:outline-none disabled:bg-slate-50 disabled:text-slate-400"
    >
        {#if value}
            {@const curr = currencies.find((c) => c.code === value)}
            <span class="flex items-center gap-2">
                <span>{curr?.flag || "🌐"}</span>
                <span class="font-medium text-slate-700">{value}</span>
            </span>
        {:else}
            <span class="text-slate-400">Select Currency...</span>
        {/if}
        <svg
            class="w-4 h-4 text-slate-400"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
        >
            <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M19 9l-7 7-7-7"
            />
        </svg>
    </button>

    {#if isOpen}
        <div class="fixed inset-0 z-10" on:click={backdropClick}></div>
        <div
            class="absolute z-20 w-full mt-1 bg-white border border-slate-200 rounded-lg shadow-xl max-h-60 overflow-hidden flex flex-col"
        >
            <div class="p-2 border-b border-slate-100">
                <input
                    type="text"
                    bind:value={search}
                    placeholder="Search..."
                    class="w-full px-2 py-1 text-sm border-none focus:ring-0 bg-slate-50 rounded"
                    autoFocus
                />
            </div>
            <div class="overflow-y-auto flex-1">
                {#each filtered as c}
                    <button
                        type="button"
                        on:click={() => select(c.code)}
                        class="w-full px-4 py-2 text-left text-sm hover:bg-slate-50 flex items-center gap-3 transition-colors {value ===
                        c.code
                            ? 'bg-blue-50 text-blue-700'
                            : 'text-slate-700'}"
                    >
                        <span class="text-lg">{c.flag}</span>
                        <div class="flex flex-col leading-tight">
                            <span class="font-medium">{c.code}</span>
                            <span class="text-xs text-slate-400">{c.name}</span>
                        </div>
                    </button>
                {/each}
                {#if filtered.length === 0}
                    <div class="p-4 text-center text-sm text-slate-500">
                        No matches found
                    </div>
                {/if}
            </div>
        </div>
    {/if}
</div>
