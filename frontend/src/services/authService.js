/**
 * Auth Service
 * =============
 * Service untuk menangani autentikasi (login, register, logout).
 * 
 * Contoh penggunaan:
 * ```javascript
 * import { authService } from '@/services';
 * 
 * // Login
 * await authService.login('email@example.com', 'password');
 * 
 * // Cek status login
 * if (authService.isAuthenticated()) {
 *   console.log('User sudah login');
 * }
 * ```
 */

import { API_ENDPOINTS, APP_CONFIG } from "../config";

class AuthService {
    constructor() {
        this.tokenKey = APP_CONFIG.TOKEN_KEY;
        this.userKey = APP_CONFIG.USER_KEY;
    }

    /**
     * Ambil token dari localStorage
     * @returns {string|null} - JWT token atau null
     */
    getToken() {
        return localStorage.getItem(this.tokenKey);
    }

    /**
     * Ambil data user dari localStorage
     * @returns {Object|null} - User object atau null
     */
    getUser() {
        const user = localStorage.getItem(this.userKey);
        return user ? JSON.parse(user) : null;
    }

    /**
     * Cek apakah user sudah login
     * @returns {boolean}
     */
    isAuthenticated() {
        return !!this.getToken();
    }

    /**
     * Cek apakah user adalah admin
     * @returns {boolean}
     */
    isAdmin() {
        const user = this.getUser();
        return user && user.role === "admin";
    }

    /**
     * Cek apakah user adalah mahasiswa
     * @returns {boolean}
     */
    isMahasiswa() {
        const user = this.getUser();
        return user && user.role === "mahasiswa";
    }

    /**
     * Simpan token dan user ke localStorage
     * @param {string} token - JWT token
     * @param {Object} user - User object
     */
    setAuth(token, user) {
        localStorage.setItem(this.tokenKey, token);
        localStorage.setItem(this.userKey, JSON.stringify(user));
    }

    /**
     * Hapus token dan user dari localStorage
     */
    clearAuth() {
        localStorage.removeItem(this.tokenKey);
        localStorage.removeItem(this.userKey);
    }

    /**
     * Dapatkan headers untuk request yang butuh auth
     * @returns {Object} - Headers dengan Authorization
     */
    getAuthHeaders() {
        const token = this.getToken();
        return token ? { Authorization: `Bearer ${token}` } : {};
    }

    /**
     * Login user
     * @param {string} email - Email user
     * @param {string} password - Password user
     * @returns {Promise<Object>} - Response dengan token dan user
     * @throws {Error} - Jika login gagal
     */
    async login(email, password) {
        const response = await fetch(API_ENDPOINTS.AUTH_LOGIN, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ email, password }),
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || "Login failed");
        }

        const data = await response.json();
        this.setAuth(data.token, data.user);
        return data;
    }

    /**
     * Register user baru
     * @param {string} name - Nama user
     * @param {string} email - Email user
     * @param {string} password - Password user
     * @returns {Promise<Object>} - Response dengan token dan user
     * @throws {Error} - Jika registrasi gagal
     */
    async register(name, email, password) {
        const response = await fetch(API_ENDPOINTS.AUTH_REGISTER, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ name, email, password }),
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || "Registration failed");
        }

        const data = await response.json();
        this.setAuth(data.token, data.user);
        return data;
    }

    /**
     * Ambil data user yang sedang login
     * @returns {Promise<Object>} - User object
     * @throws {Error} - Jika session expired
     */
    async getMe() {
        const response = await fetch(API_ENDPOINTS.AUTH_ME, {
            headers: this.getAuthHeaders(),
        });

        if (!response.ok) {
            this.clearAuth();
            throw new Error("Session expired");
        }

        return response.json();
    }

    /**
     * Logout user
     */
    logout() {
        this.clearAuth();
        window.location.hash = "/";
        window.location.reload();
    }
}

export const authService = new AuthService();
export default authService;
