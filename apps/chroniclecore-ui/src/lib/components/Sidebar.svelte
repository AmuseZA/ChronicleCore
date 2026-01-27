<script lang="ts">
    import { page } from "$app/stores";
    import { ui } from "$lib/stores/ui";

    function toggleSidebar() {
        $ui.isSidebarCollapsed = !$ui.isSidebarCollapsed;
    }
</script>

<aside
    class="bg-white dark:bg-slate-900 flex flex-col fixed inset-y-0 h-full border-r border-slate-200 dark:border-slate-800 z-50 transition-all duration-300 {$ui.isSidebarCollapsed
        ? 'w-20'
        : 'w-72'}"
>
    <!-- Brand Header -->
    <div
        class="h-20 flex items-center px-4 border-b border-slate-100 dark:border-slate-800 relative"
    >
        <div
            class="flex items-center gap-3 w-full {$ui.isSidebarCollapsed
                ? 'justify-center'
                : ''}"
        >
            <div
                class="h-10 w-10 flex-shrink-0 flex items-center justify-center rounded-xl bg-slate-900 text-white shadow-lg shadow-slate-200"
            >
                <svg
                    class="w-6 h-6"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                >
                    <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                    />
                </svg>
            </div>
            {#if !$ui.isSidebarCollapsed}
                <div class="overflow-hidden whitespace-nowrap">
                    <h1
                        class="font-bold text-xl text-slate-900 dark:text-slate-100 tracking-tight leading-none"
                    >
                        Chronicle
                    </h1>
                    <span
                        class="text-[10px] font-semibold text-slate-400 tracking-widest uppercase"
                        >Core</span
                    >
                </div>
            {/if}
        </div>

        <!-- Toggle Button -->
        <button
            on:click={toggleSidebar}
            class="absolute -right-3 top-1/2 -translate-y-1/2 bg-blue-600 border border-blue-600 rounded-full w-6 h-6 flex items-center justify-center shadow-lg text-white hover:bg-blue-700 cursor-pointer z-50 ring-2 ring-white transition-colors"
        >
            {#if $ui.isSidebarCollapsed}
                <svg
                    class="w-3 h-3"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                    ><path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M13 7l5 5m0 0l-5 5m5-5H6"
                    ></path></svg
                >
            {:else}
                <svg
                    class="w-3 h-3"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                    ><path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M15 19l-7-7 7-7"
                    ></path></svg
                >
            {/if}
        </button>
    </div>

    <!-- Nav -->
    <nav class="flex-1 px-3 py-6 space-y-2 overflow-y-auto overflow-x-hidden">
        {#if !$ui.isSidebarCollapsed}
            <div
                class="px-4 mb-2 text-xs font-bold text-slate-400 uppercase tracking-wider transition-opacity duration-200"
            >
                Overview
            </div>
        {/if}

        <a
            href="/"
            class="group flex items-center gap-3 px-3 py-3 rounded-xl text-sm font-medium transition-all duration-200
           {$page.url.pathname === '/'
                ? 'bg-blue-50 text-blue-700 shadow-sm ring-1 ring-blue-100 dark:bg-blue-900/20 dark:text-blue-400 dark:ring-blue-800'
                : 'text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800 hover:text-slate-900 dark:hover:text-slate-200'}
           {$ui.isSidebarCollapsed ? 'justify-center' : ''}"
            title="Dashboard"
        >
            <div
                class="w-8 h-8 rounded-lg flex-shrink-0 flex items-center justify-center transition-colors
                {$page.url.pathname === '/'
                    ? 'bg-blue-100 text-blue-600 dark:bg-blue-900/40 dark:text-blue-400'
                    : 'bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400 group-hover:bg-white dark:group-hover:bg-slate-700 group-hover:shadow-sm'}"
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
                        d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z"
                    />
                </svg>
            </div>
            {#if !$ui.isSidebarCollapsed}
                <span class="whitespace-nowrap">Dashboard</span>
            {/if}
        </a>

        <a
            href="/review"
            class="group flex items-center gap-3 px-3 py-3 rounded-xl text-sm font-medium transition-all duration-200
           {$page.url.pathname === '/review'
                ? 'bg-amber-50 text-amber-700 shadow-sm ring-1 ring-amber-100 dark:bg-amber-900/20 dark:text-amber-400 dark:ring-amber-800'
                : 'text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800 hover:text-slate-900 dark:hover:text-slate-200'}
           {$ui.isSidebarCollapsed ? 'justify-center' : ''}"
            title="Items to Review"
        >
            <div
                class="w-8 h-8 rounded-lg flex-shrink-0 flex items-center justify-center transition-colors
                {$page.url.pathname === '/review'
                    ? 'bg-amber-100 text-amber-600 dark:bg-amber-900/40 dark:text-amber-400'
                    : 'bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400 group-hover:bg-white dark:group-hover:bg-slate-700 group-hover:shadow-sm'}"
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
                        d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01"
                    />
                </svg>
            </div>
            {#if !$ui.isSidebarCollapsed}
                <span class="whitespace-nowrap">Items to Review</span>
            {/if}
        </a>

        <a
            href="/history"
            class="group flex items-center gap-3 px-3 py-3 rounded-xl text-sm font-medium transition-all duration-200
           {$page.url.pathname === '/history'
                ? 'bg-violet-50 text-violet-700 shadow-sm ring-1 ring-violet-100 dark:bg-violet-900/20 dark:text-violet-400 dark:ring-violet-800'
                : 'text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800 hover:text-slate-900 dark:hover:text-slate-200'}
           {$ui.isSidebarCollapsed ? 'justify-center' : ''}"
            title="History & Reports"
        >
            <div
                class="w-8 h-8 rounded-lg flex-shrink-0 flex items-center justify-center transition-colors
                {$page.url.pathname === '/history'
                    ? 'bg-violet-100 text-violet-600 dark:bg-violet-900/40 dark:text-violet-400'
                    : 'bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400 group-hover:bg-white dark:group-hover:bg-slate-700 group-hover:shadow-sm'}"
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
                        d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
                    />
                </svg>
            </div>
            {#if !$ui.isSidebarCollapsed}
                <span class="whitespace-nowrap">History & Reports</span>
            {/if}
        </a>

        <a
            href="/suggestions"
            class="group flex items-center gap-3 px-3 py-3 rounded-xl text-sm font-medium transition-all duration-200
           {$page.url.pathname === '/suggestions'
                ? 'bg-purple-50 text-purple-700 shadow-sm ring-1 ring-purple-100 dark:bg-purple-900/20 dark:text-purple-400 dark:ring-purple-800'
                : 'text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800 hover:text-slate-900 dark:hover:text-slate-200'}
           {$ui.isSidebarCollapsed ? 'justify-center' : ''}"
            title="ML Suggestions"
        >
            <div
                class="w-8 h-8 rounded-lg flex-shrink-0 flex items-center justify-center transition-colors
                {$page.url.pathname === '/suggestions'
                    ? 'bg-purple-100 text-purple-600 dark:bg-purple-900/40 dark:text-purple-400'
                    : 'bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400 group-hover:bg-white dark:group-hover:bg-slate-700 group-hover:shadow-sm'}"
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
                        d="M13 10V3L4 14h7v7l9-11h-7z"
                    />
                </svg>
            </div>
            {#if !$ui.isSidebarCollapsed}
                <span class="whitespace-nowrap">ML Suggestions</span>
            {/if}
        </a>

        {#if !$ui.isSidebarCollapsed}
            <div
                class="mt-8 px-4 mb-2 text-xs font-bold text-slate-400 uppercase tracking-wider"
            >
                Management
            </div>
        {:else}
            <div class="h-4"></div>
        {/if}

        <a
            href="/profiles"
            class="group flex items-center gap-3 px-3 py-3 rounded-xl text-sm font-medium transition-all duration-200
           {$page.url.pathname.startsWith('/profiles')
                ? 'bg-emerald-50 text-emerald-700 shadow-sm ring-1 ring-emerald-100 dark:bg-emerald-900/20 dark:text-emerald-400 dark:ring-emerald-800'
                : 'text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800 hover:text-slate-900 dark:hover:text-slate-200'}
           {$ui.isSidebarCollapsed ? 'justify-center' : ''}"
            title="Profiles & Rates"
        >
            <div
                class="w-8 h-8 rounded-lg flex-shrink-0 flex items-center justify-center transition-colors
                {$page.url.pathname.startsWith('/profiles')
                    ? 'bg-emerald-100 text-emerald-600 dark:bg-emerald-900/40 dark:text-emerald-400'
                    : 'bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400 group-hover:bg-white dark:group-hover:bg-slate-700 group-hover:shadow-sm'}"
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
                        d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
                    />
                </svg>
            </div>
            {#if !$ui.isSidebarCollapsed}
                <span class="whitespace-nowrap">Profiles & Rates</span>
            {/if}
        </a>

        <a
            href="/settings"
            class="group flex items-center gap-3 px-3 py-3 rounded-xl text-sm font-medium transition-all duration-200
           {$page.url.pathname === '/settings'
                ? 'bg-slate-100 text-slate-900 shadow-sm ring-1 ring-slate-200 dark:bg-slate-800 dark:text-slate-100 dark:ring-slate-700'
                : 'text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800 hover:text-slate-900 dark:hover:text-slate-200'}
           {$ui.isSidebarCollapsed ? 'justify-center' : ''}"
            title="Settings"
        >
            <div
                class="w-8 h-8 rounded-lg flex-shrink-0 flex items-center justify-center transition-colors
                {$page.url.pathname === '/settings'
                    ? 'bg-white text-slate-700 dark:bg-slate-700 dark:text-slate-200'
                    : 'bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400 group-hover:bg-white dark:group-hover:bg-slate-700 group-hover:shadow-sm'}"
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
                        d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
                    />
                    <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                    />
                </svg>
            </div>
            {#if !$ui.isSidebarCollapsed}
                <span class="whitespace-nowrap">Settings</span>
            {/if}
        </a>
    </nav>

    <!-- User Footer -->
    <div
        class="p-6 border-t border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/50"
    >
        <div
            class="flex items-center gap-3 {$ui.isSidebarCollapsed
                ? 'justify-center'
                : ''}"
        >
            <div
                class="w-10 h-10 rounded-full bg-slate-200 dark:bg-slate-700 border-2 border-white dark:border-slate-600 flex-shrink-0 flex items-center justify-center text-sm font-bold text-slate-500 dark:text-slate-300 shadow-sm"
            >
                US
            </div>
            {#if !$ui.isSidebarCollapsed}
                <div class="flex-1 min-w-0 overflow-hidden">
                    <p
                        class="text-sm font-bold text-slate-800 dark:text-slate-200 truncate"
                    >
                        Current User
                    </p>
                    <div class="flex items-center gap-1.5">
                        <span class="w-2 h-2 rounded-full bg-emerald-500"
                        ></span>
                        <p class="text-xs text-slate-500 truncate">Online</p>
                    </div>
                </div>
            {/if}
        </div>
    </div>
</aside>
