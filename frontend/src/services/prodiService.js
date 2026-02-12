/**
 * Prodi Service
 * =================
 * Service untuk menangani operasi CRUD Program Studi.
 */

import { API_ENDPOINTS } from "../config";
import { authService } from "./authService.js";

function getAuthHeaders() {
    const token = authService.getToken();
    return {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
    };
}

class ProdiService {
    /**
     * Ambil semua prodi (optional filter by fakultas_id)
     */
    async getAll(fakultasId = null) {
        let url = API_ENDPOINTS.PRODI;
        if (fakultasId) {
            url += `?fakultas_id=${fakultasId}`;
        }
        const response = await fetch(url, {
            headers: getAuthHeaders(),
        });
        if (!response.ok) {
            throw new Error("Gagal mengambil data program studi");
        }
        return response.json();
    }

    /**
     * Ambil prodi berdasarkan ID
     */
    async getById(id) {
        const response = await fetch(API_ENDPOINTS.PRODI_BY_ID(id), {
            headers: getAuthHeaders(),
        });
        if (!response.ok) {
            throw new Error("Gagal mengambil data program studi");
        }
        return response.json();
    }

    /**
     * Buat prodi baru
     */
    async create(data) {
        const response = await fetch(API_ENDPOINTS.PRODI, {
            method: "POST",
            headers: getAuthHeaders(),
            body: JSON.stringify(data),
        });
        if (!response.ok) {
            const text = await response.text();
            throw new Error(text || "Gagal membuat program studi");
        }
        return response.json();
    }

    /**
     * Update prodi
     */
    async update(id, data) {
        const response = await fetch(API_ENDPOINTS.PRODI_BY_ID(id), {
            method: "PUT",
            headers: getAuthHeaders(),
            body: JSON.stringify(data),
        });
        if (!response.ok) {
            const text = await response.text();
            throw new Error(text || "Gagal mengupdate program studi");
        }
        return response.json();
    }

    /**
     * Hapus prodi
     */
    async delete(id) {
        const response = await fetch(API_ENDPOINTS.PRODI_BY_ID(id), {
            method: "DELETE",
            headers: getAuthHeaders(),
        });
        if (!response.ok) {
            const text = await response.text();
            throw new Error(text || "Gagal menghapus program studi");
        }
        return response.json();
    }
}

export const prodiService = new ProdiService();
export default prodiService;
