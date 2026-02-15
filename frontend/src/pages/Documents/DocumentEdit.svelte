<script>
    import { onMount } from "svelte";
    import { link } from "svelte-spa-router";
    import {
        getDocumentById,
        updateDocument,
    } from "../../services/documentService.js";
    import fakultasService from "../../services/fakultasService.js";
    import prodiService from "../../services/prodiService.js";

    export let params = {};

    let title = "";
    let author = "";
    let abstrak = "";
    let fileType = "";
    let status = "draft";
    let fakultasId = "";
    let prodiId = "";
    let dosenPembimbing = "";
    let files = [];
    let existingFiles = [];
    let confirmCheck = false;

    let loading = false;
    let loadingData = true;
    let error = "";

    let fakultasList = [];
    let prodiList = [];
    let loadingFakultas = true;
    let loadingProdi = false;

    onMount(async () => {
        await Promise.all([loadFakultas(), loadDocument()]);
    });

    async function loadFakultas() {
        try {
            fakultasList = await fakultasService.getAll();
        } catch (e) {
            console.error("Gagal memuat fakultas:", e);
        } finally {
            loadingFakultas = false;
        }
    }

    async function loadDocument() {
        try {
            loadingData = true;
            const doc = await getDocumentById(params.id);
            title = doc.judul;
            author = doc.penulis;
            abstrak = doc.abstrak || "";
            fileType = doc.jenis_file;
            status = doc.status;
            fakultasId = doc.fakultas_id || "";
            prodiId = doc.prodi_id || "";
            dosenPembimbing = doc.dosen_pembimbing || "";
            existingFiles = doc.files || [];

            // Muat prodi berdasarkan fakultas yang sudah ada
            if (fakultasId) {
                await loadProdi(fakultasId);
            }
        } catch (e) {
            error = "Gagal memuat data dokumen: " + e.message;
        } finally {
            loadingData = false;
        }
    }

    async function loadProdi(fakId) {
        try {
            loadingProdi = true;
            prodiList = await prodiService.getAll(fakId);
        } catch (e) {
            console.error("Gagal memuat prodi:", e);
        } finally {
            loadingProdi = false;
        }
    }

    async function onFakultasChange() {
        prodiId = "";
        prodiList = [];
        if (fakultasId) {
            await loadProdi(fakultasId);
        }
    }

    function handleFileChange(e) {
        files = Array.from(e.target.files);
    }

    function removeNewFile(index) {
        files = files.filter((_, i) => i !== index);
        /** @type {HTMLInputElement | null} */
        const input = /** @type {HTMLInputElement} */ (
            document.getElementById("file-input-edit")
        );
        if (input) input.value = "";
    }

    function formatFileSize(bytes) {
        if (!bytes || bytes === 0) return "0 B";
        const k = 1024;
        const sizes = ["B", "KB", "MB", "GB"];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + " " + sizes[i];
    }

    async function handleSubmit() {
        error = "";

        if (!title || !author || !fileType || !confirmCheck) {
            error = "Semua field wajib diisi dan dikonfirmasi.";
            return;
        }

        loading = true;

        const formData = new FormData();
        formData.append("title", title);
        formData.append("author", author);
        formData.append("abstrak", abstrak);
        formData.append("category", fileType);
        formData.append("status", status);
        formData.append("dosen_pembimbing", dosenPembimbing);

        if (fakultasId) {
            formData.append("fakultas_id", fakultasId);
        }
        if (prodiId) {
            formData.append("prodi_id", prodiId);
        }

        // Append new files jika ada
        if (files.length > 0) {
            for (const file of files) {
                formData.append("files", file);
            }
        }

        try {
            await updateDocument(params.id, formData);
            alert("Dokumen berhasil diupdate");
            window.location.href = "#/documents";
        } catch (e) {
            error = e.message;
        } finally {
            loading = false;
        }
    }
</script>

<div class="max-w-4xl mx-auto flex flex-col gap-6">
    <!-- Breadcrumb -->
    <nav class="flex items-center gap-2 text-sm">
        <a
            href="#/"
            use:link
            class="text-slate-500 hover:text-primary transition-colors"
            >Dashboard</a
        >
        <span class="text-slate-400">/</span>
        <a
            href="#/documents"
            use:link
            class="text-slate-500 hover:text-primary transition-colors"
            >Kelola Dokumen</a
        >
        <span class="text-slate-400">/</span>
        <span class="text-slate-900 dark:text-white font-medium"
            >Edit Dokumen</span
        >
    </nav>

    <!-- Header -->
    <div class="flex flex-col gap-1">
        <h2
            class="text-3xl font-black text-slate-900 dark:text-white tracking-tight"
        >
            Edit Dokumen
        </h2>
        <p class="text-slate-500 dark:text-slate-400">
            Perbarui informasi dokumen yang sudah ada
        </p>
    </div>

    {#if loadingData}
        <div class="flex justify-center items-center py-20">
            <div
                class="animate-spin rounded-full h-16 w-16 border-4 border-primary border-t-transparent"
            ></div>
        </div>
    {:else}
        {#if error}
            <div
                class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-xl p-4 text-red-700 dark:text-red-400"
            >
                <p class="font-semibold">Error:</p>
                <p>{error}</p>
            </div>
        {/if}

        <div
            class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden"
        >
            <div class="p-6 md:p-8">
                <h3 class="text-lg font-bold mb-6 flex items-center gap-2">
                    <span class="material-symbols-outlined text-primary"
                        >edit_document</span
                    >
                    Detail Informasi Dokumen
                </h3>

                <form class="space-y-6" on:submit|preventDefault={handleSubmit}>
                    <!-- Judul -->
                    <div>
                        <label
                            for="edit-judul"
                            class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
                        >
                            Judul Dokumen <span class="text-red-500">*</span>
                        </label>
                        <input
                            id="edit-judul"
                            bind:value={title}
                            class="w-full px-4 py-2.5 bg-slate-100 dark:bg-slate-800 border-none rounded-lg focus:ring-2 focus:ring-primary/50 text-sm transition-all"
                            placeholder="Contoh: Analisis Pengaruh Digitalisasi..."
                            type="text"
                        />
                    </div>

                    <!-- Abstrak -->
                    <div>
                        <label
                            for="edit-abstrak"
                            class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
                        >
                            Abstrak / Ringkasan
                        </label>
                        <textarea
                            id="edit-abstrak"
                            bind:value={abstrak}
                            class="w-full px-4 py-2.5 bg-slate-100 dark:bg-slate-800 border-none rounded-lg focus:ring-2 focus:ring-primary/50 text-sm transition-all min-h-[160px] resize-y"
                            placeholder="Masukkan abstrak atau ringkasan dokumen..."
                            rows="6"
                        ></textarea>
                        <p class="mt-1 text-xs text-slate-400">
                            Abstrak akan ditampilkan pada halaman detail
                            dokumen.
                        </p>
                    </div>

                    <!-- Penulis + Jenis -->
                    <div class="grid md:grid-cols-2 gap-6">
                        <div>
                            <label
                                for="edit-penulis"
                                class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
                            >
                                Penulis <span class="text-red-500">*</span>
                            </label>
                            <input
                                id="edit-penulis"
                                bind:value={author}
                                class="w-full px-4 py-2.5 bg-slate-100 dark:bg-slate-800 border-none rounded-lg focus:ring-2 focus:ring-primary/50 text-sm transition-all"
                                placeholder="Masukkan nama penulis"
                                type="text"
                            />
                        </div>

                        <div>
                            <label
                                for="edit-jenis"
                                class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
                            >
                                Jenis Dokumen <span class="text-red-500">*</span
                                >
                            </label>
                            <select
                                id="edit-jenis"
                                bind:value={fileType}
                                class="w-full px-4 py-2.5 bg-slate-100 dark:bg-slate-800 border-none rounded-lg focus:ring-2 focus:ring-primary/50 text-sm transition-all"
                            >
                                <option value="">Pilih jenis dokumen</option>
                                <option value="skripsi">Skripsi</option>
                                <option value="tesis">Tesis</option>
                                <option value="jurnal">Jurnal</option>
                            </select>
                        </div>
                    </div>

                    <!-- Fakultas + Prodi -->
                    <div class="grid md:grid-cols-2 gap-6">
                        <div>
                            <label
                                for="edit-fakultas"
                                class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
                            >
                                Fakultas
                            </label>
                            {#if loadingFakultas}
                                <div
                                    class="w-full px-4 py-2.5 bg-slate-100 dark:bg-slate-800 rounded-lg text-sm text-slate-400"
                                >
                                    Memuat data fakultas...
                                </div>
                            {:else}
                                <select
                                    id="edit-fakultas"
                                    bind:value={fakultasId}
                                    on:change={onFakultasChange}
                                    class="w-full px-4 py-2.5 bg-slate-100 dark:bg-slate-800 border-none rounded-lg focus:ring-2 focus:ring-primary/50 text-sm transition-all"
                                >
                                    <option value=""
                                        >Pilih Fakultas (Opsional)</option
                                    >
                                    {#each fakultasList as fak}
                                        <option value={fak.id}
                                            >{fak.nama}</option
                                        >
                                    {/each}
                                </select>
                            {/if}
                        </div>

                        <div>
                            <label
                                for="edit-prodi"
                                class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
                            >
                                Program Studi
                            </label>
                            {#if loadingProdi}
                                <div
                                    class="w-full px-4 py-2.5 bg-slate-100 dark:bg-slate-800 rounded-lg text-sm text-slate-400"
                                >
                                    Memuat data prodi...
                                </div>
                            {:else}
                                <select
                                    id="edit-prodi"
                                    bind:value={prodiId}
                                    class="w-full px-4 py-2.5 bg-slate-100 dark:bg-slate-800 border-none rounded-lg focus:ring-2 focus:ring-primary/50 text-sm transition-all"
                                    disabled={!fakultasId}
                                >
                                    <option value=""
                                        >{fakultasId
                                            ? "Pilih Program Studi (Opsional)"
                                            : "Pilih Fakultas terlebih dahulu"}</option
                                    >
                                    {#each prodiList as prodi}
                                        <option value={prodi.id}
                                            >{prodi.nama}</option
                                        >
                                    {/each}
                                </select>
                            {/if}
                        </div>
                    </div>

                    <!-- Dosen Pembimbing -->
                    <div>
                        <label
                            for="edit-dosen"
                            class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
                        >
                            Dosen Pembimbing
                        </label>
                        <input
                            id="edit-dosen"
                            bind:value={dosenPembimbing}
                            class="w-full px-4 py-2.5 bg-slate-100 dark:bg-slate-800 border-none rounded-lg focus:ring-2 focus:ring-primary/50 text-sm transition-all"
                            placeholder="Contoh: Prof. Dr. Ahmad, M.Si"
                            type="text"
                        />
                    </div>

                    <!-- Status -->
                    <div>
                        <label
                            for="edit-status"
                            class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
                        >
                            Status <span class="text-red-500">*</span>
                        </label>
                        <select
                            id="edit-status"
                            bind:value={status}
                            class="w-full px-4 py-2.5 bg-slate-100 dark:bg-slate-800 border-none rounded-lg focus:ring-2 focus:ring-primary/50 text-sm transition-all"
                        >
                            <option value="draft">Draft</option>
                            <option value="publish">Publish</option>
                        </select>
                    </div>

                    <!-- Existing Files -->
                    {#if existingFiles.length > 0}
                        <div>
                            <p
                                class="text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
                            >
                                File Saat Ini
                            </p>
                            <div class="space-y-2">
                                {#each existingFiles as ef}
                                    <div
                                        class="flex items-center gap-2 px-3 py-2 bg-blue-50 dark:bg-blue-900/20 rounded-lg border border-blue-200 dark:border-blue-800"
                                    >
                                        <span
                                            class="material-symbols-outlined text-blue-600 text-base"
                                            >description</span
                                        >
                                        <span
                                            class="text-sm text-blue-700 dark:text-blue-400 truncate"
                                            >{ef.file_name}</span
                                        >
                                        <span
                                            class="text-xs text-blue-500 shrink-0"
                                            >({formatFileSize(
                                                ef.file_size,
                                            )})</span
                                        >
                                    </div>
                                {/each}
                            </div>
                        </div>
                    {/if}

                    <!-- Upload File (Optional - replace) -->
                    <div>
                        <label
                            for="file-input-edit"
                            class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
                        >
                            Ganti File (Opsional)
                        </label>
                        <div
                            class="mb-2 p-3 bg-amber-50 dark:bg-amber-900/20 text-amber-700 dark:text-amber-300 rounded-lg text-sm flex items-center gap-2"
                        >
                            <span class="material-symbols-outlined text-base"
                                >warning</span
                            >
                            Jika Anda mengunggah file baru, file lama akan diganti.
                            Biarkan kosong jika tidak ingin mengganti.
                        </div>
                        <input
                            id="file-input-edit"
                            type="file"
                            on:change={handleFileChange}
                            accept=".pdf,.doc,.docx"
                            multiple
                            class="w-full px-4 py-2.5 bg-slate-100 dark:bg-slate-800 border-none rounded-lg focus:ring-2 focus:ring-primary/50 text-sm transition-all file:mr-4 file:py-2 file:px-4 file:rounded-lg file:border-0 file:bg-primary file:text-white file:font-semibold hover:file:bg-primary/90"
                        />
                        <p class="mt-1 text-xs text-slate-400">
                            Format: PDF, DOC, DOCX (Maks 50MB per file). Anda
                            dapat memilih beberapa file.
                        </p>

                        {#if files.length > 0}
                            <div class="mt-3 space-y-2">
                                <p
                                    class="text-sm font-bold text-slate-700 dark:text-slate-300"
                                >
                                    {files.length} file baru dipilih:
                                </p>
                                {#each files as file, i}
                                    <div
                                        class="flex items-center justify-between px-3 py-2 bg-green-50 dark:bg-green-900/20 rounded-lg border border-green-200 dark:border-green-800"
                                    >
                                        <div
                                            class="flex items-center gap-2 min-w-0"
                                        >
                                            <span
                                                class="material-symbols-outlined text-green-600 text-base shrink-0"
                                                >check_circle</span
                                            >
                                            <span
                                                class="text-sm text-green-700 dark:text-green-400 truncate"
                                                >{file.name}</span
                                            >
                                            <span
                                                class="text-xs text-green-500 shrink-0"
                                                >({formatFileSize(
                                                    file.size,
                                                )})</span
                                            >
                                        </div>
                                        <button
                                            type="button"
                                            on:click={() => removeNewFile(i)}
                                            class="text-red-400 hover:text-red-600 transition-colors shrink-0 ml-2"
                                            title="Hapus file"
                                        >
                                            <span
                                                class="material-symbols-outlined text-base"
                                                >close</span
                                            >
                                        </button>
                                    </div>
                                {/each}
                            </div>
                        {/if}
                    </div>

                    <!-- Confirm -->
                    <div
                        class="flex gap-3 items-start bg-primary/5 dark:bg-primary/10 p-4 rounded-lg border border-primary/20"
                    >
                        <input
                            type="checkbox"
                            bind:checked={confirmCheck}
                            id="edit-confirm"
                            class="mt-0.5 w-5 h-5 text-primary rounded focus:ring-2 focus:ring-primary/50"
                        />
                        <label
                            for="edit-confirm"
                            class="text-sm text-slate-700 dark:text-slate-300"
                        >
                            Saya mengonfirmasi bahwa perubahan data ini sudah
                            benar.
                        </label>
                    </div>

                    <!-- Action -->
                    <div
                        class="flex justify-end gap-3 pt-4 border-t border-slate-200 dark:border-slate-800"
                    >
                        <a
                            href="#/documents"
                            use:link
                            class="px-6 py-2.5 bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 rounded-lg font-bold hover:bg-slate-200 dark:hover:bg-slate-700 transition-all"
                        >
                            Batal
                        </a>

                        <button
                            type="submit"
                            class="flex items-center gap-2 px-6 py-2.5 bg-primary text-white font-bold rounded-lg shadow-lg shadow-primary/25 hover:bg-primary/90 transition-all active:scale-95 disabled:opacity-50 disabled:cursor-not-allowed"
                            disabled={loading}
                        >
                            <span class="material-symbols-outlined text-xl"
                                >save</span
                            >
                            <span
                                >{loading
                                    ? "Menyimpan..."
                                    : "Simpan Perubahan"}</span
                            >
                        </button>
                    </div>
                </form>
            </div>
        </div>
    {/if}
</div>
