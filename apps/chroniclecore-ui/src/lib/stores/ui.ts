import { writable } from 'svelte/store';

export const ui = writable({
    isSidebarCollapsed: false
});
