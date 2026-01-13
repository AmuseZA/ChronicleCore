import { writable } from 'svelte/store';
import { fetchApi } from '$lib/api';

interface SystemSettings {
    currencyCode: string;
    locale: string;
}

function createSettingsStore() {
    const { subscribe, set, update } = writable<SystemSettings>({
        currencyCode: 'ZAR',
        locale: 'en-ZA'
    });

    return {
        subscribe,
        detectLocale: async () => {
            try {
                // 1. Try Backend
                const sys = await fetchApi('/system/locale');
                if (sys && sys.currency_code) {
                    set({
                        currencyCode: sys.currency_code,
                        locale: sys.locale || 'en-US'
                    });
                    return;
                }
            } catch (e) {
                console.warn("Backend locale detection failed, falling back to browser.", e);
            }

            // 2. Try Browser
            try {
                const browserCurrency = new Intl.NumberFormat().resolvedOptions().currency;
                if (browserCurrency) {
                    set({
                        currencyCode: browserCurrency,
                        locale: navigator.language
                    });
                    return;
                }
            } catch (e) {
                console.warn("Browser currency detection failed.", e);
            }

            // 3. Fallback to ZAR (Default)
            set({ currencyCode: 'ZAR', locale: 'en-ZA' });
        }
    };
}

export const settings = createSettingsStore();
