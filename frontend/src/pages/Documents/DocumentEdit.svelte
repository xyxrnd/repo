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
    let kataKunci = "";
    let tahun = "";

    // File management: unified list of existing + new files
    // Each item: { type: 'existing'|'new', id?, file_name, file_size, file?, order }
    let fileList = [];
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
            kataKunci = doc.kata_kunci || "";
            tahun = doc.tahun || "";

            // Initialize fileList from existing files
            fileList = (doc.files || []).map((f, i) => ({
                type: "existing",
                id: f.id,
                file_name: f.file_name,
                file_size: f.file_size,
                order: f.file_order ?? i,
            }));

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

    function removeFile(index) {
        fileList = fileList.filter((_, i) => i !== index);
        // Re-order
        fileList = fileList.map((f, i) => ({ ...f, order: i }));
    }

    function addNewFiles(e) {
        const newFiles = Array.from(e.target.files);
        for (const file of newFiles) {
            fileList = [
                ...fileList,
                {
                    type: "new",
                    file_name: file.name,
                    file_size: file.size,
                    file: file,
                    order: fileList.length,
                },
            ];
        }
        // Reset input
        e.target.value = "";
    }

    function replaceFile(index, e) {
        const file = e.target.files[0];
        if (!file) return;

        fileList[index] = {
            type: "new",
            file_name: file.name,
            file_size: file.size,
            file: file,
            order: fileList[index].order,
        };
        fileList = fileList; // trigger reactivity
        e.target.value = "";
    }

    function moveFile(index, direction) {
        const newIndex = index + direction;
        if (newIndex < 0 || newIndex >= fileList.length) return;

        const temp = fileList[index];
        fileList[index] = fileList[newIndex];
        fileList[newIndex] = temp;

        // Re-order
        fileList = fileList.map((f, i) => ({ ...f, order: i }));
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

        if (fileList.length === 0) {
            error = "Minimal harus ada 1 file.";
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
        formData.append("kata_kunci", kataKunci);
        if (tahun) formData.append("tahun", String(tahun));

        if (fakultasId) formData.append("fakultas_id", fakultasId);
        if (prodiId) formData.append("prodi_id", prodiId);

        // Build existing_files array (files to keep with their order)
        const existingFiles = fileList
            .filter((f) => f.type === "existing")
            .map((f) => ({ id: f.id, order: f.order }));
        formData.append("existing_files", JSON.stringify(existingFiles));

        // Append new files and their orders
        const newFiles = fileList.filter((f) => f.type === "new");
        const newFileOrders = [];
        for (const nf of newFiles) {
            formData.append("files", nf.file);
            newFileOrders.push(nf.order);
        }
        formData.append("new_file_orders", newFileOrders.join(","));

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

                    <!-- Dosen Pembimbing + Kata Kunci -->
                    <div class="grid md:grid-cols-2 gap-6">
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
                        <div>
                            <label
                                for="edit-katakunci"
                                class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
                            >
                                Kata Kunci
                            </label>
                            <input
                                id="edit-katakunci"
                                bind:value={kataKunci}
                                class="w-full px-4 py-2.5 bg-slate-100 dark:bg-slate-800 border-none rounded-lg focus:ring-2 focus:ring-primary/50 text-sm transition-all"
                                placeholder="Pisahkan dengan koma: AI, Machine Learning, Data"
                                type="text"
                            />
                            <p class="mt-1 text-xs text-slate-400">
                                Pisahkan kata kunci dengan tanda koma.
                            </p>
                        </div>
                    </div>

                    <!-- Tahun -->
                    <div class="w-full md:w-1/2">
                        <label
                            for="edit-tahun"
                            class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
                        >
                            Tahun Dokumen
                        </label>
                        <input
                            id="edit-tahun"
                            bind:value={tahun}
                            class="w-full px-4 py-2.5 bg-slate-100 dark:bg-slate-800 border-none rounded-lg focus:ring-2 focus:ring-primary/50 text-sm transition-all"
                            placeholder="Contoh: 2024"
                            type="number"
                            min="1900"
                            max="2099"
                        />
                        <p class="mt-1 text-xs text-slate-400">
                            Tahun penerbitan/penyusunan dokumen.
                        </p>
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

                    <!-- ====== FILE MANAGEMENT ====== -->
                    <div>
                        <div class="flex items-center justify-between mb-3">
                            <p
                                class="text-sm font-bold text-slate-700 dark:text-slate-300 flex items-center gap-2"
                            >
                                <span
                                    class="material-symbols-outlined text-primary"
                                    style="font-size: 18px;">folder</span
                                >
                                Kelola File ({fileList.length} file)
                            </p>
                            <label
                                class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-primary/10 text-primary hover:bg-primary hover:text-white rounded-lg text-xs font-bold cursor-pointer transition-all"
                            >
                                <span
                                    class="material-symbols-outlined"
                                    style="font-size: 16px;">add</span
                                >
                                Tambah File
                                <input
                                    type="file"
                                    accept=".pdf,.doc,.docx"
                                    multiple
                                    on:change={addNewFiles}
                                    class="hidden"
                                />
                            </label>
                        </div>

                        {#if fileList.length === 0}
                            <div
                                class="flex flex-col items-center justify-center py-8 bg-slate-50 dark:bg-slate-800/50 rounded-xl border-2 border-dashed border-slate-200 dark:border-slate-700"
                            >
                                <span
                                    class="material-symbols-outlined text-4xl text-slate-300 dark:text-slate-600 mb-2"
                                    >upload_file</span
                                >
                                <p
                                    class="text-sm text-slate-400 dark:text-slate-500"
                                >
                                    Belum ada file. Klik "Tambah File" untuk
                                    mengunggah.
                                </p>
                            </div>
                        {:else}
                            <div class="space-y-2">
                                {#each fileList as item, index (item.order + '-' + item.file_name)}
                                    <div
                                        class="flex items-center gap-3 px-4 py-3 rounded-xl border transition-all {item.type ===
                                        'existing'
                                            ? 'bg-blue-50/50 dark:bg-blue-900/10 border-blue-200 dark:border-blue-800/40'
                                            : 'bg-green-50/50 dark:bg-green-900/10 border-green-200 dark:border-green-800/40'}"
                                    >
                                        <!-- Order number -->
                                        <div
                                            class="w-7 h-7 rounded-lg flex items-center justify-center text-xs font-bold shrink-0 {item.type ===
                                            'existing'
                                                ? 'bg-blue-100 dark:bg-blue-900/30 text-blue-600'
                                                : 'bg-green-100 dark:bg-green-900/30 text-green-600'}"
                                        >
                                            {index + 1}
                                        </div>

                                        <!-- File icon -->
                                        <span
                                            class="material-symbols-outlined shrink-0 {item.type ===
                                            'existing'
                                                ? 'text-blue-500'
                                                : 'text-green-500'}"
                                            style="font-size: 20px;"
                                            >{item.type === "existing"
                                                ? "description"
                                                : "note_add"}</span
                                        >

                                        <!-- File info -->
                                        <div class="flex-1 min-w-0">
                                            <p
                                                class="text-sm font-medium truncate {item.type ===
                                                'existing'
                                                    ? 'text-blue-700 dark:text-blue-400'
                                                    : 'text-green-700 dark:text-green-400'}"
                                            >
                                                {item.file_name}
                                            </p>
                                            <p class="text-[11px] text-slate-400">
                                                {formatFileSize(item.file_size)}
                                                {#if item.type === "new"}
                                                    <span
                                                        class="ml-1 text-green-500 font-semibold"
                                                        >• Baru</span
                                                    >
                                                {/if}
                                            </p>
                                        </div>

                                        <!-- Actions -->
                                        <div
                                            class="flex items-center gap-1 shrink-0"
                                        >
                                            <!-- Move up -->
                                            <button
                                                type="button"
                                                on:click={() =>
                                                    moveFile(index, -1)}
                                                disabled={index === 0}
                                                class="p-1.5 rounded-lg hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-400 hover:text-slate-600 transition-colors disabled:opacity-30 disabled:cursor-not-allowed"
                                                title="Pindah ke atas"
                                            >
                                                <span
                                                    class="material-symbols-outlined"
                                                    style="font-size: 16px;"
                                                    >arrow_upward</span
                                                >
                                            </button>

                                            <!-- Move down -->
                                            <button
                                                type="button"
                                                on:click={() =>
                                                    moveFile(index, 1)}
                                                disabled={index ===
                                                    fileList.length - 1}
                                                class="p-1.5 rounded-lg hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-400 hover:text-slate-600 transition-colors disabled:opacity-30 disabled:cursor-not-allowed"
                                                title="Pindah ke bawah"
                                            >
                                                <span
                                                    class="material-symbols-outlined"
                                                    style="font-size: 16px;"
                                                    >arrow_downward</span
                                                >
                                            </button>

                                            <!-- Replace -->
                                            <label
                                                class="p-1.5 rounded-lg hover:bg-amber-100 dark:hover:bg-amber-900/30 text-slate-400 hover:text-amber-600 transition-colors cursor-pointer"
                                                title="Ganti file ini"
                                            >
                                                <span
                                                    class="material-symbols-outlined"
                                                    style="font-size: 16px;"
                                                    >swap_horiz</span
                                                >
                                                <input
                                                    type="file"
                                                    accept=".pdf,.doc,.docx"
                                                    on:change={(e) =>
                                                        replaceFile(index, e)}
                                                    class="hidden"
                                                />
                                            </label>

                                            <!-- Delete -->
                                            <button
                                                type="button"
                                                on:click={() =>
                                                    removeFile(index)}
                                                class="p-1.5 rounded-lg hover:bg-red-100 dark:hover:bg-red-900/30 text-slate-400 hover:text-red-500 transition-colors"
                                                title="Hapus file ini"
                                            >
                                                <span
                                                    class="material-symbols-outlined"
                                                    style="font-size: 16px;"
                                                    >delete</span
                                                >
                                            </button>
                                        </div>
                                    </div>
                                {/each}
                            </div>
                        {/if}

                        <p class="mt-2 text-xs text-slate-400">
                            Format: PDF, DOC, DOCX (Maks 50MB per file). Gunakan
                            tombol ↑↓ untuk mengubah urutan, ⇄ untuk mengganti
                            file, dan 🗑️ untuk menghapus.
                        </p>
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
