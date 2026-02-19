<script>
    import { onMount } from "svelte";
    import { link } from "svelte-spa-router";
    import {
        getDocuments,
        deleteDocument,
    } from "../../services/documentService.js";
    import { API_BASE_URL, API_ENDPOINTS } from "../../config/index.js";

    let documents = [];
    let loading = true;
    let error = "";
    let searchQuery = "";
    let filterStatus = "all";
    let filterType = "all";
    let filterFakultas = "all";
    let filterProdi = "all";
    let filterTahun = "all";

    // Data for filter options
    let fakultasList = [];
    let prodiList = [];
    let tahunList = [];

    // Modal state
    let showFileModal = false;
    let selectedDoc = null;

    onMount(async () => {
        await Promise.all([loadDocuments(), loadFakultas(), loadProdi()]);
    });

    async function loadDocuments() {
        try {
            loading = true;
            error = "";
            documents = await getDocuments();
            buildTahunList();
        } catch (e) {
            error = e.message;
        } finally {
            loading = false;
        }
    }

    async function loadFakultas() {
        try {
            const res = await fetch(API_ENDPOINTS.FAKULTAS);
            if (res.ok) fakultasList = await res.json();
        } catch (e) {
            console.error("Failed to load fakultas:", e);
        }
    }

    async function loadProdi() {
        try {
            const res = await fetch(API_ENDPOINTS.PRODI);
            if (res.ok) prodiList = await res.json();
        } catch (e) {
            console.error("Failed to load prodi:", e);
        }
    }

    function buildTahunList() {
        const years = new Set();
        documents.forEach((doc) => {
            if (doc.created_at) {
                years.add(new Date(doc.created_at).getFullYear());
            }
        });
        tahunList = [...years].sort((a, b) => b - a);
    }

    // Reset prodi when fakultas changes
    $: if (filterFakultas) {
        filterProdi = "all";
    }

    // Get prodi filtered by selected fakultas
    $: filteredProdiList =
        filterFakultas === "all"
            ? prodiList
            : prodiList.filter((p) => p.fakultas_id === filterFakultas);

    async function handleDelete(id) {
        if (!confirm("Apakah Anda yakin ingin menghapus dokumen ini?")) return;

        try {
            await deleteDocument(id);
            await loadDocuments();
        } catch (e) {
            alert("Gagal menghapus dokumen: " + e.message);
        }
    }

    function handleDownloadAll(id) {
        window.open(API_ENDPOINTS.DOCUMENT_DOWNLOAD_ALL(id), "_blank");
    }

    function openFileModal(doc) {
        selectedDoc = doc;
        showFileModal = true;
    }

    function closeFileModal() {
        showFileModal = false;
        selectedDoc = null;
    }

    function formatFileSize(bytes) {
        if (!bytes || bytes === 0) return "0 B";
        const k = 1024;
        const sizes = ["B", "KB", "MB", "GB"];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + " " + sizes[i];
    }

    function getFileIcon(fileName) {
        if (!fileName) return "description";
        const ext = fileName.split(".").pop().toLowerCase();
        if (ext === "pdf") return "picture_as_pdf";
        if (["doc", "docx"].includes(ext)) return "article";
        if (["xls", "xlsx"].includes(ext)) return "table_chart";
        if (["ppt", "pptx"].includes(ext)) return "slideshow";
        if (["jpg", "jpeg", "png", "gif"].includes(ext)) return "image";
        return "description";
    }

    $: filteredDocuments = documents.filter((doc) => {
        const matchesSearch =
            doc.judul.toLowerCase().includes(searchQuery.toLowerCase()) ||
            doc.penulis.toLowerCase().includes(searchQuery.toLowerCase()) ||
            (doc.dosen_pembimbing || "")
                .toLowerCase()
                .includes(searchQuery.toLowerCase());
        const matchesStatus =
            filterStatus === "all" || doc.status === filterStatus;
        const matchesType =
            filterType === "all" || doc.jenis_file === filterType;
        const matchesFakultas =
            filterFakultas === "all" || doc.fakultas_id === filterFakultas;
        const matchesProdi =
            filterProdi === "all" || doc.prodi_id === filterProdi;
        const matchesTahun =
            filterTahun === "all" ||
            (doc.created_at &&
                new Date(doc.created_at).getFullYear() === Number(filterTahun));
        return (
            matchesSearch &&
            matchesStatus &&
            matchesType &&
            matchesFakultas &&
            matchesProdi &&
            matchesTahun
        );
    });

    let activeFilterCount;
    $: activeFilterCount =
        (filterStatus !== "all" ? 1 : 0) +
        (filterType !== "all" ? 1 : 0) +
        (filterFakultas !== "all" ? 1 : 0) +
        (filterProdi !== "all" ? 1 : 0) +
        (filterTahun !== "all" ? 1 : 0);

    function resetFilters() {
        searchQuery = "";
        filterStatus = "all";
        filterType = "all";
        filterFakultas = "all";
        filterProdi = "all";
        filterTahun = "all";
    }

    function formatDate(dateString) {
        const date = new Date(dateString);
        return date.toLocaleDateString("id-ID", {
            day: "numeric",
            month: "short",
            year: "numeric",
            hour: "2-digit",
            minute: "2-digit",
        });
    }

    function getDocTypeInfo(jenisFile) {
        const types = {
            skripsi: {
                icon: "school",
                iconBg: "bg-blue-50 dark:bg-blue-900/20",
                iconText: "text-blue-600",
                badgeBg: "bg-blue-100 dark:bg-blue-900/30",
                badgeText: "text-blue-700 dark:text-blue-400",
                label: "Skripsi",
            },
            tesis: {
                icon: "workspace_premium",
                iconBg: "bg-purple-50 dark:bg-purple-900/20",
                iconText: "text-purple-600",
                badgeBg: "bg-purple-100 dark:bg-purple-900/30",
                badgeText: "text-purple-700 dark:text-purple-400",
                label: "Tesis",
            },
            jurnal: {
                icon: "article",
                iconBg: "bg-emerald-50 dark:bg-emerald-900/20",
                iconText: "text-emerald-600",
                badgeBg: "bg-emerald-100 dark:bg-emerald-900/30",
                badgeText: "text-emerald-700 dark:text-emerald-400",
                label: "Jurnal",
            },
        };
        return (
            types[jenisFile] || {
                icon: "description",
                iconBg: "bg-slate-50 dark:bg-slate-800/20",
                iconText: "text-slate-600",
                badgeBg: "bg-slate-100 dark:bg-slate-800/30",
                badgeText: "text-slate-700 dark:text-slate-400",
                label: jenisFile,
            }
        );
    }
</script>

<div class="max-w-7xl mx-auto flex flex-col gap-6">
    <!-- Header Actions -->
    <div class="flex flex-wrap items-end justify-between gap-4">
        <div class="flex flex-col gap-1">
            <h2
                class="text-3xl font-black text-slate-900 dark:text-white tracking-tight"
            >
                Kelola Dokumen
            </h2>
            <p class="text-slate-500 dark:text-slate-400">
                Pusat kontrol untuk semua dokumen akademik.
            </p>
        </div>
        <a
            href="/documents/add"
            use:link
            class="flex items-center gap-2 px-5 py-2.5 bg-primary text-white font-bold rounded-lg shadow-lg shadow-primary/25 hover:bg-primary/90 transition-all active:scale-95"
        >
            <span class="material-symbols-outlined text-xl">add</span>
            <span>Tambah Dokumen</span>
        </a>
    </div>

    <!-- Filters & Search -->
    <div
        class="bg-white dark:bg-slate-900 p-4 rounded-xl border border-slate-200 dark:border-slate-800 shadow-sm flex flex-col gap-4"
    >
        <!-- Row 1: Search + Reset -->
        <div class="flex flex-col md:flex-row gap-4 items-center">
            <div class="relative w-full md:flex-1">
                <span
                    class="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 text-xl"
                    >search</span
                >
                <input
                    bind:value={searchQuery}
                    class="w-full pl-10 pr-4 py-2 bg-slate-100 dark:bg-slate-800 border-none rounded-lg focus:ring-2 focus:ring-primary/50 text-sm transition-all"
                    placeholder="Cari judul, penulis, atau dosen pembimbing..."
                    type="text"
                />
            </div>
            {#if activeFilterCount > 0}
                <button
                    on:click={resetFilters}
                    class="flex items-center gap-1.5 px-4 py-2 text-sm font-medium text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors whitespace-nowrap"
                >
                    <span
                        class="material-symbols-outlined"
                        style="font-size: 18px;">filter_alt_off</span
                    >
                    Reset Filter ({activeFilterCount})
                </button>
            {/if}
        </div>

        <!-- Row 2: Filter dropdowns -->
        <div class="flex flex-wrap gap-3 items-center">
            <div class="flex items-center gap-2 text-slate-400">
                <span class="material-symbols-outlined" style="font-size: 20px;"
                    >tune</span
                >
                <span class="text-xs font-semibold uppercase tracking-wider"
                    >Filter</span
                >
            </div>

            <select
                bind:value={filterStatus}
                class="px-3 py-2 bg-slate-100 dark:bg-slate-800 border-none rounded-lg text-sm font-medium text-slate-700 dark:text-slate-300"
            >
                <option value="all">Semua Status</option>
                <option value="draft">Draft</option>
                <option value="publish">Published</option>
            </select>

            <select
                bind:value={filterType}
                class="px-3 py-2 bg-slate-100 dark:bg-slate-800 border-none rounded-lg text-sm font-medium text-slate-700 dark:text-slate-300"
            >
                <option value="all">Semua Jenis</option>
                <option value="skripsi">Skripsi</option>
                <option value="tesis">Tesis</option>
                <option value="jurnal">Jurnal</option>
            </select>

            <select
                bind:value={filterFakultas}
                class="px-3 py-2 bg-slate-100 dark:bg-slate-800 border-none rounded-lg text-sm font-medium text-slate-700 dark:text-slate-300"
            >
                <option value="all">Semua Fakultas</option>
                {#each fakultasList as fak}
                    <option value={fak.id}>{fak.nama}</option>
                {/each}
            </select>

            <select
                bind:value={filterProdi}
                class="px-3 py-2 bg-slate-100 dark:bg-slate-800 border-none rounded-lg text-sm font-medium text-slate-700 dark:text-slate-300"
                disabled={filterFakultas === "all" &&
                    filteredProdiList.length === 0}
            >
                <option value="all">Semua Prodi</option>
                {#each filteredProdiList as prodi}
                    <option value={prodi.id}>{prodi.nama}</option>
                {/each}
            </select>

            <select
                bind:value={filterTahun}
                class="px-3 py-2 bg-slate-100 dark:bg-slate-800 border-none rounded-lg text-sm font-medium text-slate-700 dark:text-slate-300"
            >
                <option value="all">Semua Tahun</option>
                {#each tahunList as tahun}
                    <option value={tahun}>{tahun}</option>
                {/each}
            </select>
        </div>
    </div>

    {#if error}
        <div
            class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-xl p-4 text-red-700 dark:text-red-400"
        >
            <p class="font-semibold">Error:</p>
            <p>{error}</p>
        </div>
    {/if}

    {#if loading}
        <div class="flex justify-center items-center py-20">
            <div
                class="animate-spin rounded-full h-16 w-16 border-4 border-primary border-t-transparent"
            ></div>
        </div>
    {:else if filteredDocuments.length === 0}
        <div
            class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-12 text-center"
        >
            <span
                class="material-symbols-outlined text-6xl text-slate-300 dark:text-slate-700 mb-4"
                >folder_off</span
            >
            <h3
                class="text-xl font-bold text-slate-700 dark:text-slate-300 mb-2"
            >
                Tidak ada dokumen
            </h3>
            <p class="text-slate-500 dark:text-slate-400 mb-6">
                {searchQuery || activeFilterCount > 0
                    ? "Tidak ada dokumen yang sesuai dengan filter Anda"
                    : "Mulai dengan menambahkan dokumen baru"}
            </p>
            {#if !searchQuery && activeFilterCount === 0}
                <a
                    href="/documents/add"
                    use:link
                    class="inline-flex items-center gap-2 px-5 py-2.5 bg-primary text-white font-bold rounded-lg shadow-lg shadow-primary/25 hover:bg-primary/90 transition-all"
                >
                    <span class="material-symbols-outlined">add</span>
                    <span>Tambah Dokumen Pertama</span>
                </a>
            {/if}
        </div>
    {:else}
        <!-- Data Table -->
        <div
            class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden"
        >
            <div class="overflow-x-auto">
                <table class="w-full text-left border-collapse">
                    <thead>
                        <tr
                            class="bg-slate-50 dark:bg-slate-800/50 border-b border-slate-200 dark:border-slate-800"
                        >
                            <th
                                class="px-6 py-4 text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider"
                            >
                                Judul Dokumen
                            </th>
                            <th
                                class="px-6 py-4 text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider"
                            >
                                Penulis
                            </th>
                            <th
                                class="px-6 py-4 text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider"
                            >
                                Fakultas
                            </th>
                            <th
                                class="px-6 py-4 text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider"
                            >
                                Dosen Pembimbing
                            </th>
                            <th
                                class="px-6 py-4 text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider"
                            >
                                Jenis
                            </th>
                            <th
                                class="px-6 py-4 text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider"
                            >
                                File
                            </th>
                            <th
                                class="px-6 py-4 text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider"
                            >
                                Status
                            </th>
                            <th
                                class="px-6 py-4 text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider text-right"
                            >
                                Aksi
                            </th>
                        </tr>
                    </thead>
                    <tbody
                        class="divide-y divide-slate-100 dark:divide-slate-800"
                    >
                        {#each filteredDocuments as doc (doc.id)}
                            <tr
                                class="hover:bg-slate-50/80 dark:hover:bg-slate-800/30 transition-colors"
                            >
                                <td class="px-6 py-4">
                                    <div class="flex items-center gap-3">
                                        <div
                                            class="{getDocTypeInfo(
                                                doc.jenis_file,
                                            ).iconBg} {getDocTypeInfo(
                                                doc.jenis_file,
                                            ).iconText} p-2 rounded-lg"
                                        >
                                            <span
                                                class="material-symbols-outlined"
                                                >{getDocTypeInfo(doc.jenis_file)
                                                    .icon}</span
                                            >
                                        </div>
                                        <div class="flex flex-col">
                                            <span
                                                class="text-sm font-bold text-slate-900 dark:text-white"
                                                >{doc.judul}</span
                                            >
                                            <span class="text-xs text-slate-500"
                                                >ID: {doc.id.substring(
                                                    0,
                                                    8,
                                                )}</span
                                            >
                                        </div>
                                    </div>
                                </td>
                                <td
                                    class="px-6 py-4 text-sm text-slate-600 dark:text-slate-400"
                                >
                                    {doc.penulis}
                                </td>
                                <td class="px-6 py-4">
                                    {#if doc.fakultas_nama}
                                        <div class="flex flex-col gap-0.5">
                                            <span
                                                class="px-2 py-1 bg-indigo-100 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-400 text-xs font-bold rounded inline-block w-fit"
                                            >
                                                {doc.fakultas_nama}
                                            </span>
                                            {#if doc.prodi_nama}
                                                <span
                                                    class="text-xs text-slate-500 dark:text-slate-400 pl-1"
                                                >
                                                    {doc.prodi_nama}
                                                </span>
                                            {/if}
                                        </div>
                                    {:else}
                                        <span class="text-xs text-slate-400"
                                            >-</span
                                        >
                                    {/if}
                                </td>
                                <td
                                    class="px-6 py-4 text-sm text-slate-600 dark:text-slate-400"
                                >
                                    {doc.dosen_pembimbing || "-"}
                                </td>
                                <td class="px-6 py-4">
                                    <span
                                        class="px-2.5 py-1 {getDocTypeInfo(
                                            doc.jenis_file,
                                        ).badgeBg} {getDocTypeInfo(
                                            doc.jenis_file,
                                        )
                                            .badgeText} text-[10px] font-bold uppercase rounded leading-none"
                                    >
                                        {getDocTypeInfo(doc.jenis_file).label}
                                    </span>
                                </td>
                                <td class="px-6 py-4">
                                    {#if doc.files && doc.files.length > 0}
                                        <button
                                            on:click={() => openFileModal(doc)}
                                            class="px-2 py-1 bg-teal-100 dark:bg-teal-900/30 text-teal-700 dark:text-teal-400 text-xs font-bold rounded hover:bg-teal-200 dark:hover:bg-teal-900/50 transition-colors cursor-pointer flex items-center gap-1"
                                            title="Klik untuk melihat daftar file"
                                        >
                                            <span
                                                class="material-symbols-outlined text-sm"
                                                >folder_open</span
                                            >
                                            {doc.files.length} file
                                        </button>
                                    {:else}
                                        <span class="text-xs text-slate-400"
                                            >-</span
                                        >
                                    {/if}
                                </td>
                                <td class="px-6 py-4">
                                    <span
                                        class="px-2.5 py-1 text-[10px] font-bold uppercase rounded leading-none {doc.status ===
                                        'publish'
                                            ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
                                            : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'}"
                                    >
                                        {doc.status}
                                    </span>
                                </td>
                                <td class="px-6 py-4 text-right">
                                    <div class="flex justify-end gap-1">
                                        <button
                                            on:click={() => openFileModal(doc)}
                                            class="p-1.5 text-teal-600 hover:bg-teal-50 dark:hover:bg-teal-900/20 rounded transition-colors"
                                            title="Lihat File"
                                        >
                                            <span
                                                class="material-symbols-outlined text-xl"
                                                >visibility</span
                                            >
                                        </button>
                                        <button
                                            on:click={() =>
                                                handleDownloadAll(doc.id)}
                                            class="p-1.5 text-green-600 hover:bg-green-50 dark:hover:bg-green-900/20 rounded transition-colors"
                                            title="Download Semua File (ZIP)"
                                        >
                                            <span
                                                class="material-symbols-outlined text-xl"
                                                >download</span
                                            >
                                        </button>
                                        <a
                                            href="/documents/edit/{doc.id}"
                                            use:link
                                            class="p-1.5 text-primary hover:bg-primary/10 rounded transition-colors"
                                            title="Edit"
                                        >
                                            <span
                                                class="material-symbols-outlined text-xl"
                                                >edit_square</span
                                            >
                                        </a>
                                        <button
                                            on:click={() =>
                                                handleDelete(doc.id)}
                                            class="p-1.5 text-slate-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 rounded transition-colors"
                                            title="Hapus"
                                        >
                                            <span
                                                class="material-symbols-outlined text-xl"
                                                >delete</span
                                            >
                                        </button>
                                    </div>
                                </td>
                            </tr>
                        {/each}
                    </tbody>
                </table>
            </div>

            <!-- Pagination -->
            <div
                class="px-6 py-4 bg-slate-50 dark:bg-slate-800/50 flex items-center justify-between"
            >
                <span class="text-sm text-slate-500">
                    Menampilkan {filteredDocuments.length} dari {documents.length}
                    dokumen
                </span>
            </div>
        </div>
    {/if}
</div>

<!-- File List Modal -->
{#if showFileModal && selectedDoc}
    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-noninteractive-element-interactions -->
    <div
        role="dialog"
        class="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center p-4"
        on:click={closeFileModal}
    >
        <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-noninteractive-element-interactions -->
        <div
            role="document"
            class="bg-white dark:bg-slate-900 rounded-2xl shadow-2xl w-full max-w-lg max-h-[80vh] flex flex-col overflow-hidden"
            on:click|stopPropagation
        >
            <!-- Modal Header -->
            <div
                class="px-6 py-4 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between shrink-0"
            >
                <div class="flex flex-col gap-0.5">
                    <h3
                        class="text-lg font-bold text-slate-900 dark:text-white flex items-center gap-2"
                    >
                        <span class="material-symbols-outlined text-primary"
                            >folder_open</span
                        >
                        Daftar File
                    </h3>
                    <p
                        class="text-sm text-slate-500 truncate max-w-sm"
                        title={selectedDoc.judul}
                    >
                        {selectedDoc.judul}
                    </p>
                </div>
                <button
                    on:click={closeFileModal}
                    class="p-1.5 text-slate-400 hover:text-slate-600 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-lg transition-colors"
                >
                    <span class="material-symbols-outlined">close</span>
                </button>
            </div>

            <!-- Modal Body -->
            <div class="px-6 py-4 overflow-y-auto flex-1">
                {#if selectedDoc.files && selectedDoc.files.length > 0}
                    <div class="space-y-3">
                        {#each selectedDoc.files as file, i}
                            <div
                                class="flex items-center gap-3 p-3 bg-slate-50 dark:bg-slate-800/50 rounded-xl border border-slate-200 dark:border-slate-700 hover:border-primary/30 transition-colors"
                            >
                                <!-- File Icon -->
                                <div
                                    class="p-2.5 bg-primary/10 rounded-lg shrink-0"
                                >
                                    <span
                                        class="material-symbols-outlined text-primary text-xl"
                                    >
                                        {getFileIcon(file.file_name)}
                                    </span>
                                </div>

                                <!-- File Info -->
                                <div class="flex-1 min-w-0">
                                    <p
                                        class="text-sm font-semibold text-slate-900 dark:text-white truncate"
                                        title={file.file_name}
                                    >
                                        {file.file_name}
                                    </p>
                                    <div class="flex items-center gap-3 mt-0.5">
                                        <span class="text-xs text-slate-500">
                                            {formatFileSize(file.file_size)}
                                        </span>
                                        <span class="text-xs text-slate-400"
                                            >•</span
                                        >
                                        <span class="text-xs text-slate-500">
                                            File #{i + 1}
                                        </span>
                                    </div>
                                </div>

                                <!-- Download Button -->
                                <a
                                    href="{API_BASE_URL}/{file.file_path}"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    class="p-2 text-green-600 hover:bg-green-50 dark:hover:bg-green-900/20 rounded-lg transition-colors shrink-0"
                                    title="Download {file.file_name}"
                                >
                                    <span
                                        class="material-symbols-outlined text-xl"
                                        >download</span
                                    >
                                </a>
                            </div>
                        {/each}
                    </div>
                {:else}
                    <div class="text-center py-8">
                        <span
                            class="material-symbols-outlined text-5xl text-slate-300 dark:text-slate-700 mb-3"
                        >
                            folder_off
                        </span>
                        <p class="text-slate-500 dark:text-slate-400 text-sm">
                            Tidak ada file yang terdaftar untuk dokumen ini.
                        </p>
                        <p class="text-slate-400 text-xs mt-1">
                            File mungkin hanya tersimpan sebagai file_path
                            utama.
                        </p>
                    </div>
                {/if}
            </div>

            <!-- Modal Footer -->
            <div
                class="px-6 py-4 border-t border-slate-200 dark:border-slate-800 flex items-center justify-between shrink-0"
            >
                <span class="text-xs text-slate-400">
                    {selectedDoc.files ? selectedDoc.files.length : 0} file terdaftar
                </span>
                <button
                    on:click={closeFileModal}
                    class="px-4 py-2 bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 rounded-lg text-sm font-semibold hover:bg-slate-200 dark:hover:bg-slate-700 transition-colors"
                >
                    Tutup
                </button>
            </div>
        </div>
    </div>
{/if}
