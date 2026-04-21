/**
 * Konfigurasi aplikasi
 * =====================
 * File ini berisi semua konfigurasi yang digunakan di aplikasi.
 * Ubah nilai di sini untuk menyesuaikan dengan environment.
 */

// URL API Backend
// Development: Vite proxy atau localhost
// Production: auto-detect dari origin browser (saat di-serve dari Go backend)
function getApiBaseUrl() {
  // Jika di-serve dari Go backend (production), gunakan origin yang sama
  if (
    typeof window !== "undefined" &&
    !window.location.port.startsWith("517")
  ) {
    return window.location.origin;
  }
  // Development (Vite dev server)
  return "http://localhost:8080";
}
export const API_BASE_URL = getApiBaseUrl();

// API Endpoints
export const API_ENDPOINTS = {
  // Auth
  AUTH_LOGIN: `${API_BASE_URL}/api/auth/login`,
  AUTH_REGISTER: `${API_BASE_URL}/api/auth/register`,
  AUTH_ME: `${API_BASE_URL}/api/auth/me`,

  // Users
  USERS: `${API_BASE_URL}/api/users`,
  USER_BY_ID: (id) => `${API_BASE_URL}/api/users/${id}`,

  // Documents
  DOCUMENTS: `${API_BASE_URL}/api/documents`,
  DOCUMENT_BY_ID: (id) => `${API_BASE_URL}/api/documents/${id}`,
  DOCUMENT_DOWNLOAD: (id) => `${API_BASE_URL}/download/${id}`,
  DOCUMENT_DOWNLOAD_ALL: (id) =>
    `${API_BASE_URL}/api/documents/${id}/download-all`,
  DOCUMENTS_POPULAR: `${API_BASE_URL}/api/documents/popular`,
  STATS: `${API_BASE_URL}/api/stats`,
  TOP_FAKULTAS: `${API_BASE_URL}/api/stats/top-fakultas`,

  // Fakultas
  FAKULTAS: `${API_BASE_URL}/api/fakultas`,
  FAKULTAS_BY_ID: (id) => `${API_BASE_URL}/api/fakultas/${id}`,

  // Prodi
  PRODI: `${API_BASE_URL}/api/prodi`,
  PRODI_BY_ID: (id) => `${API_BASE_URL}/api/prodi/${id}`,

  // Site Settings
  SITE_SETTINGS: `${API_BASE_URL}/api/site-settings`,
  SITE_LOGO_UPLOAD: `${API_BASE_URL}/api/site-settings/logo`,

  // Access Requests
  ACCESS_REQUESTS: `${API_BASE_URL}/api/access-requests`,
  ACCESS_REQUEST_BY_ID: (id) => `${API_BASE_URL}/api/access-requests/${id}`,
  VERIFY_ACCESS_TOKEN: `${API_BASE_URL}/api/verify-access-token`,

  // Email OTP Verification
  SEND_OTP: `${API_BASE_URL}/api/send-otp`,
  VERIFY_OTP: `${API_BASE_URL}/api/verify-otp`,

  // Student Registration
  STUDENT_SIGNUP: `${API_BASE_URL}/api/student-signup`,
  OCR_KTM: `${API_BASE_URL}/api/ocr-ktm`,
  STUDENT_REGISTRATIONS: `${API_BASE_URL}/api/student-registrations`,
  STUDENT_REGISTRATION_BY_ID: (id) =>
    `${API_BASE_URL}/api/student-registrations/${id}`,
};

// App Configuration
export const APP_CONFIG = {
  APP_NAME: "ScholarHub",
  APP_DESCRIPTION: "Repository Dokumen Akademik",

  // Storage keys
  TOKEN_KEY: "auth_token",
  USER_KEY: "auth_user",

  // Pagination
  DEFAULT_PAGE_SIZE: 10,

  // File upload
  MAX_FILE_SIZE: 50 * 1024 * 1024, // 50 MB
  ALLOWED_FILE_TYPES: [".pdf", ".doc", ".docx"],

  // Jenis Dokumen
  DOCUMENT_TYPES: ["skripsi", "tesis", "jurnal"],
};

// Route paths
export const ROUTES = {
  // Public
  HOME: "/",
  LANDING: "/landing",
  BROWSE: "/browse",
  ABOUT: "/about",

  // Auth
  LOGIN: "/login",
  REGISTER: "/register",

  // Admin
  ADMIN: "/admin",
  DASHBOARD: "/admin/dashboard",
  DOCUMENTS: "/admin/documents",
  DOCUMENTS_ADD: "/admin/documents/add",
  DOCUMENTS_EDIT: (id) => `/admin/documents/edit/${id}`,
  USERS: "/admin/users",
  FAKULTAS: "/admin/fakultas",
  PRODI: "/admin/prodi",
  REPORTS: "/reports",
  SETTINGS: "/settings",
  SYSTEM_SETTINGS: "/system-settings",
};
