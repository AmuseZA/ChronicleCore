import { writable, get } from 'svelte/store';
import { fetchApi } from '$lib/api';

export interface Suggestion {
    profile_id: number;
    profile_name: string;
    score: number;
    reason_codes: string[];
}

export interface SuggestionMap {
    [block_id: number]: Suggestion[];
}

function createSuggestionsStore() {
    const { subscribe, update, set } = writable<SuggestionMap>({});

    return {
        subscribe,
        loadForBlocks: async (endpoint: string) => {
            // In a real implementation, we might batch fetch suggestions
            // For now we assume they might come attached to blocks or we fetch all pending
            try {
                const suggestions = await fetchApi('/ml/suggestions');
                if (suggestions) {
                    const map: SuggestionMap = {};
                    suggestions.forEach((s: any) => {
                        if (!map[s.block_id]) map[s.block_id] = [];
                        map[s.block_id].push({
                            profile_id: s.suggested_profile_id,
                            profile_name: s.profile_name, // Backend should return this or we join
                            score: s.score,
                            reason_codes: s.global_reasons || []
                        });
                    });
                    set(map);
                }
            } catch (e) {
                console.error("Failed to load suggestions", e);
            }
        },
        accept: async (blockId: number, profileId: number) => {
            await fetchApi('/ml/suggestions/accept', {
                method: 'POST',
                body: JSON.stringify({ block_id: blockId, profile_id: profileId })
            });
            // Optimistic update: remove suggestion for this block
            update(s => {
                const ns = { ...s };
                delete ns[blockId];
                return ns;
            });
        }
    };
}

export const suggestions = createSuggestionsStore();
export const mlEnabled = writable<boolean>(false); // Settings toggle
