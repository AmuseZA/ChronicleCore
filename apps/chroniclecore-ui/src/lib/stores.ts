import { writable } from 'svelte/store';

export const trackingStatus = writable({
    state: 'STOPPED',
    last_active_at: null,
    idle_seconds: 0,
    current_window: null
});

export const todayBlocks = writable([]);
export const profiles = writable([]);
