/**
 * Fakultas Service
 * =================
 * Service untuk menangani operasi CRUD Fakultas.
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

class FakultasService {
    /**
     * Ambil semua fakultas
     */
    async getAll() {
        const response = await fetch(API_ENDPOINTS.FAKULTAS, {
            headers: getAuthHeaders(),
        });
        if (!response.ok) {
            throw new Error("Gagal mengambil data fakultas");
        }
        return response.json();
    }

    /**
     * Ambil fakultas berdasarkan ID
     */
    async getById(id) {
        const response = await fetch(API_ENDPOINTS.FAKULTAS_BY_ID(id), {
            headers: getAuthHeaders(),
        });
        if (!response.ok) {
            throw new Error("Gagal mengambil data fakultas");
        }
        return response.json();
    }

    /**
     * Buat fakultas baru
     */
    async create(data) {
        const response = await fetch(API_ENDPOINTS.FAKULTAS, {
            method: "POST",
            headers: getAuthHeaders(),
            body: JSON.stringify(data),
        });
        if (!response.ok) {
            const text = await response.text();
            throw new Error(text || "Gagal membuat fakultas");
        }
        return response.json();
    }

    /**
     * Update fakultas
     */
    async update(id, data) {
        const response = await fetch(API_ENDPOINTS.FAKULTAS_BY_ID(id), {
            method: "PUT",
            headers: getAuthHeaders(),
            body: JSON.stringify(data),
        });
        if (!response.ok) {
            const text = await response.text();
            throw new Error(text || "Gagal mengupdate fakultas");
        }
        return response.json();
    }

    /**
     * Hapus fakultas
     */
    async delete(id) {
        const response = await fetch(API_ENDPOINTS.FAKULTAS_BY_ID(id), {
            method: "DELETE",
            headers: getAuthHeaders(),
        });
        if (!response.ok) {
            const text = await response.text();
            throw new Error(text || "Gagal menghapus fakultas");
        }
        return response.json();
    }
}

export const fakultasService = new FakultasService();
export default fakultasService;
