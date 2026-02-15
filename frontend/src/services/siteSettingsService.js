/**
 * Site Settings Service
 * =====================
 * Service untuk mengelola pengaturan situs (nama, logo, about, kontak)
 */

import { API_BASE_URL } from "../config/index.js";

const SETTINGS_URL = `${API_BASE_URL}/api/site-settings`;
const LOGO_UPLOAD_URL = `${API_BASE_URL}/api/site-settings/logo`;

/**
 * Mengambil semua pengaturan situs
 */
export async function getSiteSettings() {
    const res = await fetch(SETTINGS_URL);
    if (!res.ok) throw new Error("Gagal mengambil pengaturan situs");
    return await res.json();
}

/**
 * Memperbarui pengaturan situs (memerlukan auth admin)
 * @param {Object} settings - key-value pairs yang ingin diupdate
 */
export async function updateSiteSettings(settings) {
    const token = localStorage.getItem("auth_token");
    const res = await fetch(SETTINGS_URL, {
        method: "PUT",
        headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(settings),
    });
    if (!res.ok) {
        const text = await res.text();
        throw new Error(text || "Gagal menyimpan pengaturan");
    }
    return await res.json();
}

/**
 * Upload logo situs (memerlukan auth admin)
 * @param {File} file - File gambar logo
 */
export async function uploadSiteLogo(file) {
    const token = localStorage.getItem("auth_token");
    const formData = new FormData();
    formData.append("logo", file);

    const res = await fetch(LOGO_UPLOAD_URL, {
        method: "POST",
        headers: {
            Authorization: `Bearer ${token}`,
        },
        body: formData,
    });
    if (!res.ok) {
        const text = await res.text();
        throw new Error(text || "Gagal mengupload logo");
    }
    return await res.json();
}

export default {
    getSiteSettings,
    updateSiteSettings,
    uploadSiteLogo,
};
