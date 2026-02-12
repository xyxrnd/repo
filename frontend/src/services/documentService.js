/**
 * Document Service
 * =================
 * Service untuk menangani operasi CRUD dokumen.
 * 
 * Contoh penggunaan:
 * ```javascript
 * import { documentService } from '@/services';
 * 
 * // Ambil semua dokumen
 * const docs = await documentService.getAll();
 * 
 * // Ambil dokumen by ID
 * const doc = await documentService.getById('123');
 * ```
 */

import { API_ENDPOINTS } from "../config";

class DocumentService {
  /**
   * Ambil semua dokumen
   * @returns {Promise<Array>} - Array of documents
   * @throws {Error} - Jika gagal mengambil data
   */
  async getAll() {
    const response = await fetch(API_ENDPOINTS.DOCUMENTS);
    if (!response.ok) {
      throw new Error("Gagal mengambil data dokumen");
    }
    return response.json();
  }

  /**
   * Ambil dokumen berdasarkan ID
   * @param {string} id - Document ID
   * @returns {Promise<Object>} - Document object
   * @throws {Error} - Jika dokumen tidak ditemukan
   */
  async getById(id) {
    const response = await fetch(API_ENDPOINTS.DOCUMENT_BY_ID(id));
    if (!response.ok) {
      throw new Error("Gagal mengambil data dokumen");
    }
    return response.json();
  }

  /**
   * Buat dokumen baru
   * @param {FormData} formData - Form data berisi file dan metadata
   * @returns {Promise<Object>} - Created document
   * @throws {Error} - Jika gagal membuat dokumen
   */
  async create(formData) {
    const response = await fetch(API_ENDPOINTS.DOCUMENTS, {
      method: "POST",
      body: formData,
    });
    if (!response.ok) {
      throw new Error("Gagal membuat dokumen");
    }
    return response.json();
  }

  /**
   * Update dokumen
   * @param {string} id - Document ID
   * @param {FormData} formData - Form data berisi file dan metadata
   * @returns {Promise<Object>} - Updated document
   * @throws {Error} - Jika gagal mengupdate
   */
  async update(id, formData) {
    const response = await fetch(API_ENDPOINTS.DOCUMENT_BY_ID(id), {
      method: "PUT",
      body: formData,
    });
    if (!response.ok) {
      throw new Error("Gagal mengupdate dokumen");
    }
    return response.json();
  }

  /**
   * Hapus dokumen
   * @param {string} id - Document ID
   * @returns {Promise<Object>} - Response message
   * @throws {Error} - Jika gagal menghapus
   */
  async delete(id) {
    const response = await fetch(API_ENDPOINTS.DOCUMENT_BY_ID(id), {
      method: "DELETE",
    });
    if (!response.ok) {
      throw new Error("Gagal menghapus dokumen");
    }
    return response.json();
  }

  /**
   * Download dokumen
   * Membuka tab baru untuk download file
   * @param {string} id - Document ID
   */
  download(id) {
    window.open(API_ENDPOINTS.DOCUMENT_DOWNLOAD(id), "_blank");
  }
}

// Export functions untuk backward compatibility
export const getDocuments = () => documentService.getAll();
export const getDocumentById = (id) => documentService.getById(id);
export const createDocument = (formData) => documentService.create(formData);
export const updateDocument = (id, formData) => documentService.update(id, formData);
export const deleteDocument = (id) => documentService.delete(id);
export const downloadDocument = (id) => documentService.download(id);

export const documentService = new DocumentService();
export default documentService;
