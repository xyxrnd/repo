/**
 * Site Settings Store
 * ====================
 * Global reactive store untuk pengaturan situs.
 * Data diambil dari API dan tersedia di seluruh aplikasi.
 * Ketika admin mengubah settings, panggil refreshSiteSettings()
 * agar semua komponen yang subscribe otomatis ter-update.
 */

import { writable, derived } from "svelte/store";
import { getSiteSettings } from "../services/siteSettingsService.js";
import { API_BASE_URL } from "../config/index.js";

// Default values
const DEFAULT_SETTINGS = {
    app_name: "ScholarHub",
    app_description: "Repository Dokumen Akademik",
    about_text: "",
    visi: "",
    misi: "",
    logo_url: "",
    address: "",
    email: "",
    phone: "",
    footer_text: "",
};

// Internal writable store
const _siteSettings = writable({ ...DEFAULT_SETTINGS });

// Loading state
export const siteSettingsLoading = writable(true);

// Track if already fetched
let _fetched = false;

/**
 * Fetch site settings dari API dan update store
 */
export async function refreshSiteSettings() {
    try {
        siteSettingsLoading.set(true);
        const data = await getSiteSettings();
        _siteSettings.set({ ...DEFAULT_SETTINGS, ...data });
        _fetched = true;
    } catch (e) {
        console.error("Failed to load site settings:", e);
        // Keep defaults on error
    } finally {
        siteSettingsLoading.set(false);
    }
}

/**
 * Initialize store — only fetches once unless forced
 */
export async function initSiteSettings() {
    if (!_fetched) {
        await refreshSiteSettings();
    }
}

// Derived read-only store for consumers
export const siteSettings = derived(_siteSettings, ($s) => $s);

// Derived convenience stores
export const appName = derived(
    _siteSettings,
    ($s) => $s.app_name || DEFAULT_SETTINGS.app_name,
);
export const appDescription = derived(
    _siteSettings,
    ($s) => $s.app_description || DEFAULT_SETTINGS.app_description,
);
export const aboutText = derived(_siteSettings, ($s) => $s.about_text || "");
export const visi = derived(_siteSettings, ($s) => $s.visi || "");
export const misi = derived(_siteSettings, ($s) => $s.misi || "");
export const footerText = derived(_siteSettings, ($s) => $s.footer_text || "");
export const logoFullUrl = derived(_siteSettings, ($s) => {
    if (!$s.logo_url) return "";
    return $s.logo_url.startsWith("http")
        ? $s.logo_url
        : `${API_BASE_URL}${$s.logo_url}`;
});
export const contactInfo = derived(_siteSettings, ($s) => ({
    address: $s.address || "",
    email: $s.email || "",
    phone: $s.phone || "",
}));
