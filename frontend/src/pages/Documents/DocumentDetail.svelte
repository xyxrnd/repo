<script>
    import { onMount } from "svelte";
    import { params } from "svelte-spa-router";
    import {
        getDocumentById,
        downloadDocument,
    } from "../../services/documentService";
    import authService from "../../services/authService";
    import { API_BASE_URL, API_ENDPOINTS } from "../../config";

    export let id = "";

    let doc = null;
    let loading = true;
    let error = "";
    let isLoggedIn = false;

    // Access Request state
    let showAccessForm = null; // file_id yang sedang diminta akses
    let accessFormData = { nama: "", email: "", ktm: null };
    let accessFormLoading = false;
    let accessFormSuccess = "";
    let accessFormError = "";

    // Token verification state
    let showTokenInput = null; // file_id yang sedang input token
    let tokenValue = "";
    let tokenLoading = false;
    let tokenError = "";
    let unlockedFiles = {}; // { file_id: { file_path, file_name } }

    // Get ID from route params
    $: if ($params && $params.id) {
        id = $params.id;
        loadDocument();
    }

    onMount(() => {
        isLoggedIn = authService.isAuthenticated();
        if (id) loadDocument();
    });

    async function loadDocument() {
        try {
            loading = true;
            error = "";
            doc = await getDocumentById(id);
        } catch (e) {
            error = "Dokumen tidak ditemukan atau terjadi kesalahan.";
            console.error("Failed to load document:", e);
        } finally {
            loading = false;
        }
    }

    function handleDownload() {
        downloadDocument(doc.id);
    }

    function handleFileDownload(file) {
        if (file.is_locked) {
            showAccessForm = file.id;
            accessFormError = "";
            accessFormSuccess = "";
            return;
        }
        window.open(`${API_BASE_URL}/${file.file_path}`, "_blank");
    }

    function handleFilePreview(file) {
        if (file.is_locked) {
            showAccessForm = file.id;
            accessFormError = "";
            accessFormSuccess = "";
            return;
        }
        window.open(`${API_BASE_URL}/${file.file_path}`, "_blank");
    }

    function openAccessForm(fileId) {
        showAccessForm = fileId;
        accessFormData = { nama: "", email: "", ktm: null };
        accessFormSuccess = "";
        accessFormError = "";
    }

    function closeAccessForm() {
        showAccessForm = null;
        accessFormData = { nama: "", email: "", ktm: null };
        accessFormSuccess = "";
        accessFormError = "";
    }

    function handleKtmUpload(event) {
        const file = event.target.files[0];
        if (file) {
            accessFormData.ktm = file;
        }
    }

    // Token verification functions
    function openTokenInput(fileId) {
        showTokenInput = fileId;
        tokenValue = "";
        tokenError = "";
    }

    function closeTokenInput() {
        showTokenInput = null;
        tokenValue = "";
        tokenError = "";
    }

    async function verifyToken() {
        if (!tokenValue.trim()) {
            tokenError = "Masukkan token akses.";
            return;
        }

        try {
            tokenLoading = true;
            tokenError = "";

            const response = await fetch(API_ENDPOINTS.VERIFY_ACCESS_TOKEN, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    file_id: showTokenInput,
                    token: tokenValue.trim(),
                }),
            });

            if (!response.ok) {
                const text = await response.text();
                throw new Error(text || "Token tidak valid");
            }

            const data = await response.json();
            // File unlocked! Store the info
            unlockedFiles = {
                ...unlockedFiles,
                [showTokenInput]: {
                    file_path: data.file_path,
                    file_name: data.file_name,
                },
            };
            showTokenInput = null;
            tokenValue = "";
        } catch (e) {
            tokenError = e.message || "Token tidak valid atau belum disetujui.";
        } finally {
            tokenLoading = false;
        }
    }

    function isFileUnlocked(fileId) {
        return !!unlockedFiles[fileId];
    }

    function handleUnlockedPreview(fileId) {
        const info = unlockedFiles[fileId];
        if (info) {
            window.open(`${API_BASE_URL}/${info.file_path}`, "_blank");
        }
    }

    function handleUnlockedDownload(fileId) {
        const info = unlockedFiles[fileId];
        if (info) {
            const link = document.createElement("a");
            link.href = `${API_BASE_URL}/${info.file_path}`;
            link.download = info.file_name;
            link.click();
        }
    }

    async function submitAccessRequest() {
        if (!accessFormData.nama.trim() || !accessFormData.email.trim()) {
            accessFormError = "Nama dan email wajib diisi.";
            return;
        }

        // Simple email validation
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!emailRegex.test(accessFormData.email.trim())) {
            accessFormError = "Format email tidak valid.";
            return;
        }

        if (!accessFormData.ktm) {
            accessFormError = "Upload KTM wajib dilakukan.";
            return;
        }

        try {
            accessFormLoading = true;
            accessFormError = "";

            const formData = new FormData();
            formData.append("document_id", doc.id);
            formData.append("file_id", showAccessForm);
            formData.append("nama", accessFormData.nama.trim());
            formData.append("email", accessFormData.email.trim());
            if (accessFormData.ktm) {
                formData.append("ktm", accessFormData.ktm);
            }

            const response = await fetch(API_ENDPOINTS.ACCESS_REQUESTS, {
                method: "POST",
                body: formData,
            });

            if (response.status === 409) {
                accessFormError =
                    "Anda sudah memiliki permintaan akses yang masih menunggu persetujuan untuk file ini.";
                return;
            }

            if (!response.ok) {
                const text = await response.text();
                throw new Error(text || "Gagal mengirim permintaan akses");
            }

            accessFormSuccess =
                "Permintaan akses berhasil dikirim! Silakan tunggu persetujuan dari admin.";
            accessFormData = { nama: "", email: "", ktm: null };
        } catch (e) {
            accessFormError =
                e.message || "Terjadi kesalahan saat mengirim permintaan.";
        } finally {
            accessFormLoading = false;
        }
    }

    function formatDate(dateString) {
        const date = new Date(dateString);
        return date.toLocaleDateString("id-ID", {
            day: "numeric",
            month: "long",
            year: "numeric",
        });
    }

    function formatFileSize(bytes) {
        if (!bytes || bytes === 0) return "0 B";
        const k = 1024;
        const sizes = ["B", "KB", "MB", "GB"];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + " " + sizes[i];
    }

    function getTypeStyle(jenisFile) {
        const styles = {
            Skripsi: {
                icon: "school",
                bg: "bg-blue-500",
                light: "bg-blue-100 dark:bg-blue-900/30",
                text: "text-blue-600 dark:text-blue-400",
                gradient: "from-blue-500 to-blue-700",
            },
            skripsi: {
                icon: "school",
                bg: "bg-blue-500",
                light: "bg-blue-100 dark:bg-blue-900/30",
                text: "text-blue-600 dark:text-blue-400",
                gradient: "from-blue-500 to-blue-700",
            },
            Tesis: {
                icon: "workspace_premium",
                bg: "bg-purple-500",
                light: "bg-purple-100 dark:bg-purple-900/30",
                text: "text-purple-600 dark:text-purple-400",
                gradient: "from-purple-500 to-purple-700",
            },
            tesis: {
                icon: "workspace_premium",
                bg: "bg-purple-500",
                light: "bg-purple-100 dark:bg-purple-900/30",
                text: "text-purple-600 dark:text-purple-400",
                gradient: "from-purple-500 to-purple-700",
            },
            Jurnal: {
                icon: "article",
                bg: "bg-teal-500",
                light: "bg-teal-100 dark:bg-teal-900/30",
                text: "text-teal-600 dark:text-teal-400",
                gradient: "from-teal-500 to-teal-700",
            },
            jurnal: {
                icon: "article",
                bg: "bg-teal-500",
                light: "bg-teal-100 dark:bg-teal-900/30",
                text: "text-teal-600 dark:text-teal-400",
                gradient: "from-teal-500 to-teal-700",
            },
            Disertasi: {
                icon: "history_edu",
                bg: "bg-amber-500",
                light: "bg-amber-100 dark:bg-amber-900/30",
                text: "text-amber-600 dark:text-amber-400",
                gradient: "from-amber-500 to-amber-700",
            },
        };
        return (
            styles[jenisFile] || {
                icon: "description",
                bg: "bg-slate-500",
                light: "bg-slate-100 dark:bg-slate-700",
                text: "text-slate-600 dark:text-slate-400",
                gradient: "from-slate-500 to-slate-700",
            }
        );
    }

    function getFileIcon(fileName) {
        if (!fileName) return "description";
        const ext = fileName.split(".").pop().toLowerCase();
        const icons = {
            pdf: "picture_as_pdf",
            doc: "description",
            docx: "description",
            xls: "table_chart",
            xlsx: "table_chart",
            ppt: "slideshow",
            pptx: "slideshow",
        };
        return icons[ext] || "description";
    }

    function getInitials(name) {
        if (!name) return "?";
        return name
            .split(" ")
            .map((n) => n[0])
            .join("")
            .substring(0, 2)
            .toUpperCase();
    }
</script>

<div class="min-h-screen bg-slate-50 dark:bg-background-dark">
    {#if loading}
        <!-- Loading State -->
        <div
            class="bg-gradient-to-br from-primary via-blue-600 to-indigo-700 text-white py-16 px-4"
        >
            <div class="container mx-auto max-w-4xl">
                <div
                    class="h-8 w-48 bg-white/20 rounded animate-pulse mb-4"
                ></div>
                <div
                    class="h-12 w-full bg-white/20 rounded animate-pulse mb-4"
                ></div>
                <div class="h-6 w-64 bg-white/20 rounded animate-pulse"></div>
            </div>
        </div>
        <div class="container mx-auto max-w-4xl px-4 py-8">
            <div class="bg-white dark:bg-slate-800 rounded-xl p-8">
                <div class="space-y-4">
                    {#each Array(5) as _}
                        <div
                            class="h-5 bg-slate-200 dark:bg-slate-700 rounded animate-pulse"
                        ></div>
                    {/each}
                </div>
            </div>
        </div>
    {:else if error}
        <!-- Error State -->
        <div
            class="bg-gradient-to-br from-red-500 to-red-700 text-white py-16 px-4"
        >
            <div class="container mx-auto max-w-4xl text-center">
                <span class="material-symbols-outlined text-6xl mb-4 opacity-50"
                    >error</span
                >
                <h1 class="text-3xl font-bold mb-2">Dokumen Tidak Ditemukan</h1>
                <p class="text-red-100 mb-6">{error}</p>
                <a
                    href="#/browse"
                    class="inline-flex items-center gap-2 px-6 py-3 bg-white/20 hover:bg-white/30 backdrop-blur-sm rounded-lg font-medium transition-all"
                >
                    <span class="material-symbols-outlined">arrow_back</span>
                    Kembali ke Jelajah
                </a>
            </div>
        </div>
    {:else if doc}
        <!-- Hero Header -->
        {@const typeStyle = getTypeStyle(doc.jenis_file)}
        <div
            class="bg-gradient-to-br from-primary via-blue-600 to-indigo-700 text-white py-12 lg:py-16 px-4 relative overflow-hidden"
        >
            <!-- Decorative background -->
            <div class="absolute inset-0 opacity-10">
                <div
                    class="absolute -top-24 -right-24 w-96 h-96 rounded-full bg-white/20 blur-3xl"
                ></div>
                <div
                    class="absolute -bottom-24 -left-24 w-64 h-64 rounded-full bg-white/10 blur-2xl"
                ></div>
            </div>

            <div class="container mx-auto max-w-4xl relative">
                <!-- Breadcrumb -->
                <nav class="flex items-center gap-2 text-sm text-blue-200 mb-6">
                    <a href="#/" class="hover:text-white transition-colors"
                        >Home</a
                    >
                    <span>›</span>
                    <a
                        href="#/browse"
                        class="hover:text-white transition-colors">Jelajah</a
                    >
                    <span>›</span>
                    <span class="text-white/70 truncate max-w-[200px]"
                        >{doc.judul}</span
                    >
                </nav>

                <!-- Category badge -->
                <div class="flex items-center gap-3 mb-4">
                    <span
                        class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-white/20 backdrop-blur-sm rounded-full text-sm font-semibold"
                    >
                        <span
                            class="material-symbols-outlined"
                            style="font-size: 16px;">{typeStyle.icon}</span
                        >
                        {doc.jenis_file}
                    </span>
                    {#if doc.view_count > 0}
                        <span
                            class="inline-flex items-center gap-1 px-2.5 py-1 bg-white/10 backdrop-blur-sm rounded-full text-xs font-medium"
                        >
                            <span
                                class="material-symbols-outlined"
                                style="font-size: 14px;">visibility</span
                            >
                            {doc.view_count} kali dilihat
                        </span>
                    {/if}
                </div>

                <!-- Title -->
                <h1
                    class="text-2xl md:text-4xl font-black leading-tight mb-4 max-w-3xl"
                >
                    {doc.judul}
                </h1>

                <!-- Author & Date -->
                <div class="flex flex-wrap items-center gap-4 text-blue-100">
                    <div class="flex items-center gap-2">
                        <div
                            class="w-8 h-8 rounded-full bg-white/20 flex items-center justify-center text-xs font-bold"
                        >
                            {getInitials(doc.penulis)}
                        </div>
                        <span class="font-medium">{doc.penulis}</span>
                    </div>
                    <span class="hidden md:inline text-blue-300">•</span>
                    <div class="flex items-center gap-1.5">
                        <span
                            class="material-symbols-outlined"
                            style="font-size: 16px;">calendar_today</span
                        >
                        {formatDate(doc.created_at)}
                    </div>
                </div>
            </div>
        </div>

        <!-- Content -->
        <div class="container mx-auto max-w-4xl px-4 py-8">
            <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
                <!-- Main Content (Left Column) -->
                <div class="lg:col-span-2 space-y-6">
                    <!-- Informasi Dokumen Card -->
                    <div
                        class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden shadow-sm"
                    >
                        <div
                            class="px-6 py-4 border-b border-slate-100 dark:border-slate-700"
                        >
                            <h2
                                class="text-lg font-bold text-slate-900 dark:text-white flex items-center gap-2"
                            >
                                <span
                                    class="material-symbols-outlined text-primary"
                                    >info</span
                                >
                                Informasi Dokumen
                            </h2>
                        </div>
                        <div class="p-6">
                            <table class="w-full">
                                <tbody>
                                    <tr
                                        class="border-b border-slate-100 dark:border-slate-700/50"
                                    >
                                        <td
                                            class="py-3 pr-4 text-sm font-semibold text-slate-500 dark:text-slate-400 w-40 align-top"
                                            >Judul</td
                                        >
                                        <td
                                            class="py-3 text-sm text-slate-900 dark:text-white font-medium"
                                            >{doc.judul}</td
                                        >
                                    </tr>
                                    <tr
                                        class="border-b border-slate-100 dark:border-slate-700/50"
                                    >
                                        <td
                                            class="py-3 pr-4 text-sm font-semibold text-slate-500 dark:text-slate-400 w-40 align-top"
                                            >Penulis</td
                                        >
                                        <td
                                            class="py-3 text-sm text-slate-900 dark:text-white"
                                            >{doc.penulis}</td
                                        >
                                    </tr>
                                    <tr
                                        class="border-b border-slate-100 dark:border-slate-700/50"
                                    >
                                        <td
                                            class="py-3 pr-4 text-sm font-semibold text-slate-500 dark:text-slate-400 w-40 align-top"
                                            >Jenis Dokumen</td
                                        >
                                        <td class="py-3">
                                            <span
                                                class="inline-flex items-center gap-1.5 px-2.5 py-1 {typeStyle.light} {typeStyle.text} text-xs font-bold rounded-full uppercase"
                                            >
                                                <span
                                                    class="material-symbols-outlined"
                                                    style="font-size: 14px;"
                                                    >{typeStyle.icon}</span
                                                >
                                                {doc.jenis_file}
                                            </span>
                                        </td>
                                    </tr>
                                    {#if doc.fakultas_nama}
                                        <tr
                                            class="border-b border-slate-100 dark:border-slate-700/50"
                                        >
                                            <td
                                                class="py-3 pr-4 text-sm font-semibold text-slate-500 dark:text-slate-400 w-40 align-top"
                                                >Fakultas</td
                                            >
                                            <td
                                                class="py-3 text-sm text-slate-900 dark:text-white"
                                                >{doc.fakultas_nama}</td
                                            >
                                        </tr>
                                    {/if}
                                    {#if doc.prodi_nama}
                                        <tr
                                            class="border-b border-slate-100 dark:border-slate-700/50"
                                        >
                                            <td
                                                class="py-3 pr-4 text-sm font-semibold text-slate-500 dark:text-slate-400 w-40 align-top"
                                                >Program Studi</td
                                            >
                                            <td
                                                class="py-3 text-sm text-slate-900 dark:text-white"
                                                >{doc.prodi_nama}</td
                                            >
                                        </tr>
                                    {/if}
                                    {#if doc.dosen_pembimbing}
                                        <tr
                                            class="border-b border-slate-100 dark:border-slate-700/50"
                                        >
                                            <td
                                                class="py-3 pr-4 text-sm font-semibold text-slate-500 dark:text-slate-400 w-40 align-top"
                                                >Dosen Pembimbing</td
                                            >
                                            <td
                                                class="py-3 text-sm text-slate-900 dark:text-white"
                                                >{doc.dosen_pembimbing}</td
                                            >
                                        </tr>
                                    {/if}
                                    <tr>
                                        <td
                                            class="py-3 pr-4 text-sm font-semibold text-slate-500 dark:text-slate-400 w-40 align-top"
                                            >Tanggal Upload</td
                                        >
                                        <td
                                            class="py-3 text-sm text-slate-900 dark:text-white"
                                            >{formatDate(doc.created_at)}</td
                                        >
                                    </tr>
                                </tbody>
                            </table>
                        </div>
                    </div>

                    <!-- Abstrak / Abstract Section -->
                    {#if doc.abstrak}
                        <div
                            class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden shadow-sm"
                        >
                            <div
                                class="px-6 py-4 border-b border-slate-100 dark:border-slate-700"
                            >
                                <h2
                                    class="text-lg font-bold text-slate-900 dark:text-white flex items-center gap-2"
                                >
                                    <span
                                        class="material-symbols-outlined text-primary"
                                        >subject</span
                                    >
                                    Abstract
                                </h2>
                            </div>
                            <div class="p-6">
                                <div
                                    class="prose prose-slate dark:prose-invert max-w-none"
                                >
                                    <p
                                        class="text-sm text-slate-700 dark:text-slate-300 leading-relaxed whitespace-pre-line"
                                    >
                                        {doc.abstrak}
                                    </p>
                                </div>
                            </div>
                        </div>
                    {/if}

                    <!-- Sitasi / Citation Section -->
                    <div
                        class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden shadow-sm"
                    >
                        <div
                            class="px-6 py-4 border-b border-slate-100 dark:border-slate-700"
                        >
                            <h2
                                class="text-lg font-bold text-slate-900 dark:text-white flex items-center gap-2"
                            >
                                <span
                                    class="material-symbols-outlined text-primary"
                                    >format_quote</span
                                >
                                Sitasi
                            </h2>
                        </div>
                        <div class="p-6">
                            <div
                                class="bg-slate-50 dark:bg-slate-900/50 rounded-lg p-4 border border-slate-200 dark:border-slate-700"
                            >
                                <p
                                    class="text-sm text-slate-700 dark:text-slate-300 italic leading-relaxed"
                                >
                                    {doc.penulis} ({new Date(
                                        doc.created_at,
                                    ).getFullYear()})
                                    <strong>{doc.judul}</strong>.
                                    {#if doc.jenis_file}
                                        {doc.jenis_file
                                            .charAt(0)
                                            .toUpperCase() +
                                            doc.jenis_file.slice(1)},
                                    {/if}
                                    {#if doc.fakultas_nama}
                                        {doc.fakultas_nama}{#if doc.prodi_nama}, {doc.prodi_nama}{/if}.
                                    {/if}
                                </p>
                            </div>
                            <button
                                on:click={() => {
                                    const citationText = `${doc.penulis} (${new Date(doc.created_at).getFullYear()}) ${doc.judul}. ${doc.jenis_file ? doc.jenis_file.charAt(0).toUpperCase() + doc.jenis_file.slice(1) + ", " : ""}${doc.fakultas_nama ? doc.fakultas_nama : ""}${doc.prodi_nama ? ", " + doc.prodi_nama : ""}.`;
                                    navigator.clipboard.writeText(citationText);
                                    alert("Sitasi berhasil disalin!");
                                }}
                                class="mt-3 inline-flex items-center gap-1.5 px-3 py-1.5 bg-primary/10 text-primary hover:bg-primary hover:text-white rounded-lg text-xs font-semibold transition-all"
                            >
                                <span
                                    class="material-symbols-outlined"
                                    style="font-size: 16px;">content_copy</span
                                >
                                Salin Sitasi
                            </button>
                        </div>
                    </div>

                    <!-- Files / Daftar File Card -->
                    {#if doc.files && doc.files.length > 0}
                        <div
                            class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden shadow-sm"
                        >
                            <div
                                class="px-6 py-4 border-b border-slate-100 dark:border-slate-700 flex items-center justify-between"
                            >
                                <h2
                                    class="text-lg font-bold text-slate-900 dark:text-white flex items-center gap-2"
                                >
                                    <span
                                        class="material-symbols-outlined text-primary"
                                        >folder_open</span
                                    >
                                    Daftar File
                                    <span
                                        class="text-xs font-normal text-slate-400 ml-1"
                                        >({doc.files.length} file)</span
                                    >
                                </h2>
                            </div>
                            <div
                                class="divide-y divide-slate-100 dark:divide-slate-700/50"
                            >
                                {#each doc.files as file, index}
                                    <div
                                        class="px-6 py-4 hover:bg-slate-50 dark:hover:bg-slate-700/30 transition-colors group"
                                    >
                                        <div class="flex items-center gap-4">
                                            <!-- File icon -->
                                            <div
                                                class="w-10 h-10 rounded-lg {file.is_locked
                                                    ? 'bg-amber-100 dark:bg-amber-900/30'
                                                    : 'bg-red-100 dark:bg-red-900/30'} flex items-center justify-center flex-shrink-0"
                                            >
                                                {#if file.is_locked}
                                                    <span
                                                        class="material-symbols-outlined text-amber-500"
                                                        style="font-size: 22px;"
                                                        >lock</span
                                                    >
                                                {:else}
                                                    <span
                                                        class="material-symbols-outlined text-red-500"
                                                        style="font-size: 22px;"
                                                        >{getFileIcon(
                                                            file.file_name,
                                                        )}</span
                                                    >
                                                {/if}
                                            </div>

                                            <!-- File info -->
                                            <div class="flex-1 min-w-0">
                                                <div
                                                    class="flex items-center gap-2"
                                                >
                                                    <p
                                                        class="text-sm font-semibold text-slate-900 dark:text-white truncate"
                                                    >
                                                        {file.file_name}
                                                    </p>
                                                    {#if file.is_locked}
                                                        <span
                                                            class="inline-flex items-center gap-0.5 px-1.5 py-0.5 bg-amber-100 dark:bg-amber-900/30 text-amber-600 dark:text-amber-400 text-[10px] font-bold rounded uppercase"
                                                        >
                                                            <span
                                                                class="material-symbols-outlined"
                                                                style="font-size: 11px;"
                                                                >lock</span
                                                            >
                                                            Terkunci
                                                        </span>
                                                    {/if}
                                                </div>
                                                <p
                                                    class="text-xs text-slate-400 dark:text-slate-500 mt-0.5"
                                                >
                                                    {formatFileSize(
                                                        file.file_size,
                                                    )}
                                                </p>
                                            </div>

                                            <!-- Actions -->
                                            <div
                                                class="flex items-center gap-2 flex-shrink-0"
                                            >
                                                {#if file.is_locked && !isFileUnlocked(file.id)}
                                                    <!-- Tombol untuk file terkunci -->
                                                    <button
                                                        on:click={() =>
                                                            openAccessForm(
                                                                file.id,
                                                            )}
                                                        class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-amber-100 dark:bg-amber-900/30 text-amber-600 dark:text-amber-400 hover:bg-amber-200 dark:hover:bg-amber-900/50 rounded-lg text-xs font-semibold transition-all"
                                                        title="Minta akses file"
                                                    >
                                                        <span
                                                            class="material-symbols-outlined"
                                                            style="font-size: 16px;"
                                                            >key</span
                                                        >
                                                        <span
                                                            class="hidden sm:inline"
                                                            >Minta Akses</span
                                                        >
                                                    </button>
                                                    <button
                                                        on:click={() =>
                                                            openTokenInput(
                                                                file.id,
                                                            )}
                                                        class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-sky-100 dark:bg-sky-900/30 text-sky-600 dark:text-sky-400 hover:bg-sky-200 dark:hover:bg-sky-900/50 rounded-lg text-xs font-semibold transition-all"
                                                        title="Masukkan token akses"
                                                    >
                                                        <span
                                                            class="material-symbols-outlined"
                                                            style="font-size: 16px;"
                                                            >vpn_key</span
                                                        >
                                                        <span
                                                            class="hidden sm:inline"
                                                            >Masukkan Token</span
                                                        >
                                                    </button>
                                                {:else if file.is_locked && isFileUnlocked(file.id)}
                                                    <!-- File terkunci tapi sudah di-unlock dengan token -->
                                                    <span
                                                        class="inline-flex items-center gap-1 px-2 py-1 bg-green-100 dark:bg-green-900/30 text-green-600 dark:text-green-400 text-[10px] font-bold rounded-lg"
                                                    >
                                                        <span
                                                            class="material-symbols-outlined"
                                                            style="font-size: 13px;"
                                                            >lock_open</span
                                                        >
                                                        Terbuka
                                                    </span>
                                                    <button
                                                        on:click={() =>
                                                            handleUnlockedPreview(
                                                                file.id,
                                                            )}
                                                        class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300 hover:bg-primary/10 hover:text-primary rounded-lg text-xs font-semibold transition-all"
                                                        title="Preview file"
                                                    >
                                                        <span
                                                            class="material-symbols-outlined"
                                                            style="font-size: 16px;"
                                                            >visibility</span
                                                        >
                                                        <span
                                                            class="hidden sm:inline"
                                                            >Preview</span
                                                        >
                                                    </button>
                                                    <button
                                                        on:click={() =>
                                                            handleUnlockedDownload(
                                                                file.id,
                                                            )}
                                                        class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-primary/10 text-primary hover:bg-primary hover:text-white rounded-lg text-xs font-semibold transition-all"
                                                        title="Download file"
                                                    >
                                                        <span
                                                            class="material-symbols-outlined"
                                                            style="font-size: 16px;"
                                                            >download</span
                                                        >
                                                        <span
                                                            class="hidden sm:inline"
                                                            >Download</span
                                                        >
                                                    </button>
                                                {:else}
                                                    <!-- Tombol Preview dan Download untuk file tidak terkunci -->
                                                    <button
                                                        on:click={() =>
                                                            handleFilePreview(
                                                                file,
                                                            )}
                                                        class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300 hover:bg-primary/10 hover:text-primary rounded-lg text-xs font-semibold transition-all"
                                                        title="Preview file"
                                                    >
                                                        <span
                                                            class="material-symbols-outlined"
                                                            style="font-size: 16px;"
                                                            >visibility</span
                                                        >
                                                        <span
                                                            class="hidden sm:inline"
                                                            >Preview</span
                                                        >
                                                    </button>
                                                    <button
                                                        on:click={() =>
                                                            handleFileDownload(
                                                                file,
                                                            )}
                                                        class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-primary/10 text-primary hover:bg-primary hover:text-white rounded-lg text-xs font-semibold transition-all"
                                                        title="Download file"
                                                    >
                                                        <span
                                                            class="material-symbols-outlined"
                                                            style="font-size: 16px;"
                                                            >download</span
                                                        >
                                                        <span
                                                            class="hidden sm:inline"
                                                            >Download</span
                                                        >
                                                    </button>
                                                {/if}
                                            </div>
                                        </div>

                                        <!-- Access Request Form (inline, muncul di bawah file terkunci) -->
                                        {#if showAccessForm === file.id}
                                            <div
                                                class="mt-4 p-5 bg-gradient-to-br from-amber-50 to-orange-50 dark:from-amber-900/10 dark:to-orange-900/10 rounded-xl border border-amber-200/60 dark:border-amber-800/30 animate-slideDown"
                                            >
                                                <div
                                                    class="flex items-center justify-between mb-4"
                                                >
                                                    <h4
                                                        class="text-sm font-bold text-slate-900 dark:text-white flex items-center gap-2"
                                                    >
                                                        <span
                                                            class="material-symbols-outlined text-amber-500"
                                                            style="font-size: 18px;"
                                                            >key</span
                                                        >
                                                        Formulir Permintaan Akses
                                                    </h4>
                                                    <button
                                                        on:click={closeAccessForm}
                                                        class="p-1 hover:bg-amber-200/50 dark:hover:bg-amber-900/30 rounded-lg transition-colors"
                                                        title="Tutup formulir"
                                                    >
                                                        <span
                                                            class="material-symbols-outlined text-slate-400"
                                                            style="font-size: 18px;"
                                                            >close</span
                                                        >
                                                    </button>
                                                </div>

                                                {#if accessFormSuccess}
                                                    <div
                                                        class="flex items-start gap-3 p-4 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800/30 rounded-lg"
                                                    >
                                                        <span
                                                            class="material-symbols-outlined text-green-500 flex-shrink-0"
                                                            style="font-size: 20px;"
                                                            >check_circle</span
                                                        >
                                                        <div>
                                                            <p
                                                                class="text-sm font-medium text-green-800 dark:text-green-300"
                                                            >
                                                                {accessFormSuccess}
                                                            </p>
                                                            <button
                                                                on:click={closeAccessForm}
                                                                class="mt-2 text-xs font-semibold text-green-600 hover:text-green-700 underline"
                                                            >
                                                                Tutup
                                                            </button>
                                                        </div>
                                                    </div>
                                                {:else}
                                                    <div class="space-y-3">
                                                        <!-- Nama -->
                                                        <div>
                                                            <label
                                                                for="access-nama-{file.id}"
                                                                class="block text-xs font-semibold text-slate-600 dark:text-slate-400 mb-1"
                                                            >
                                                                Nama Lengkap <span
                                                                    class="text-red-500"
                                                                    >*</span
                                                                >
                                                            </label>
                                                            <input
                                                                id="access-nama-{file.id}"
                                                                type="text"
                                                                bind:value={
                                                                    accessFormData.nama
                                                                }
                                                                placeholder="Masukkan nama lengkap Anda"
                                                                class="w-full px-3 py-2 text-sm bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-600 rounded-lg focus:ring-2 focus:ring-amber-400 focus:border-amber-400 text-slate-900 dark:text-white placeholder-slate-400 transition-all"
                                                            />
                                                        </div>

                                                        <!-- Email -->
                                                        <div>
                                                            <label
                                                                for="access-email-{file.id}"
                                                                class="block text-xs font-semibold text-slate-600 dark:text-slate-400 mb-1"
                                                            >
                                                                Alamat Email <span
                                                                    class="text-red-500"
                                                                    >*</span
                                                                >
                                                            </label>
                                                            <input
                                                                id="access-email-{file.id}"
                                                                type="email"
                                                                bind:value={
                                                                    accessFormData.email
                                                                }
                                                                placeholder="contoh@email.com"
                                                                class="w-full px-3 py-2 text-sm bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-600 rounded-lg focus:ring-2 focus:ring-amber-400 focus:border-amber-400 text-slate-900 dark:text-white placeholder-slate-400 transition-all"
                                                            />
                                                        </div>

                                                        <!-- Upload KTM -->
                                                        <div>
                                                            <label
                                                                for="access-ktm-{file.id}"
                                                                class="block text-xs font-semibold text-slate-600 dark:text-slate-400 mb-1"
                                                            >
                                                                Upload KTM
                                                                (Kartu Tanda
                                                                Mahasiswa) <span
                                                                    class="text-red-500"
                                                                    >*</span
                                                                >
                                                            </label>
                                                            <div
                                                                class="relative"
                                                            >
                                                                <input
                                                                    id="access-ktm-{file.id}"
                                                                    type="file"
                                                                    accept="image/*,.pdf"
                                                                    on:change={handleKtmUpload}
                                                                    class="w-full px-3 py-2 text-sm bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-600 rounded-lg focus:ring-2 focus:ring-amber-400 focus:border-amber-400 text-slate-900 dark:text-white file:mr-3 file:py-1 file:px-3 file:rounded-md file:border-0 file:text-xs file:font-semibold file:bg-amber-100 file:text-amber-600 hover:file:bg-amber-200 transition-all"
                                                                />
                                                            </div>
                                                            {#if accessFormData.ktm}
                                                                <p
                                                                    class="mt-1 text-xs text-green-600 flex items-center gap-1"
                                                                >
                                                                    <span
                                                                        class="material-symbols-outlined"
                                                                        style="font-size: 14px;"
                                                                        >check_circle</span
                                                                    >
                                                                    {accessFormData
                                                                        .ktm
                                                                        .name}
                                                                </p>
                                                            {/if}
                                                            <p
                                                                class="mt-1 text-[11px] text-slate-400"
                                                            >
                                                                Format: JPG,
                                                                PNG, PDF.
                                                                Maksimum 5 MB.
                                                            </p>
                                                        </div>

                                                        <!-- Error Message -->
                                                        {#if accessFormError}
                                                            <div
                                                                class="flex items-start gap-2 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800/30 rounded-lg"
                                                            >
                                                                <span
                                                                    class="material-symbols-outlined text-red-500 flex-shrink-0"
                                                                    style="font-size: 16px;"
                                                                    >error</span
                                                                >
                                                                <p
                                                                    class="text-xs text-red-600 dark:text-red-400"
                                                                >
                                                                    {accessFormError}
                                                                </p>
                                                            </div>
                                                        {/if}

                                                        <!-- Buttons -->
                                                        <div
                                                            class="flex items-center gap-2 pt-1"
                                                        >
                                                            <button
                                                                on:click={submitAccessRequest}
                                                                disabled={accessFormLoading}
                                                                class="inline-flex items-center gap-1.5 px-4 py-2 bg-gradient-to-r from-amber-500 to-orange-500 hover:from-amber-600 hover:to-orange-600 text-white rounded-lg text-xs font-bold shadow-sm hover:shadow-md transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                                                            >
                                                                {#if accessFormLoading}
                                                                    <span
                                                                        class="material-symbols-outlined animate-spin"
                                                                        style="font-size: 16px;"
                                                                        >progress_activity</span
                                                                    >
                                                                    Mengirim...
                                                                {:else}
                                                                    <span
                                                                        class="material-symbols-outlined"
                                                                        style="font-size: 16px;"
                                                                        >send</span
                                                                    >
                                                                    Kirim Permintaan
                                                                {/if}
                                                            </button>
                                                            <button
                                                                on:click={closeAccessForm}
                                                                class="inline-flex items-center gap-1 px-3 py-2 bg-slate-100 dark:bg-slate-700 text-slate-500 dark:text-slate-400 hover:bg-slate-200 dark:hover:bg-slate-600 rounded-lg text-xs font-semibold transition-all"
                                                            >
                                                                Batal
                                                            </button>
                                                        </div>
                                                    </div>
                                                {/if}
                                            </div>
                                        {/if}

                                        <!-- Token Input Form (inline, muncul di bawah file terkunci) -->
                                        {#if showTokenInput === file.id}
                                            <div
                                                class="mt-4 p-4 bg-gradient-to-br from-sky-50 to-blue-50 dark:from-sky-900/10 dark:to-blue-900/10 rounded-xl border border-sky-200/60 dark:border-sky-800/30 animate-slideDown"
                                            >
                                                <div
                                                    class="flex items-center justify-between mb-3"
                                                >
                                                    <h4
                                                        class="text-sm font-bold text-slate-900 dark:text-white flex items-center gap-2"
                                                    >
                                                        <span
                                                            class="material-symbols-outlined text-sky-500"
                                                            style="font-size: 18px;"
                                                            >vpn_key</span
                                                        >
                                                        Masukkan Token Akses
                                                    </h4>
                                                    <button
                                                        on:click={closeTokenInput}
                                                        class="p-1 hover:bg-sky-200/50 dark:hover:bg-sky-900/30 rounded-lg transition-colors"
                                                        title="Tutup"
                                                    >
                                                        <span
                                                            class="material-symbols-outlined text-slate-400"
                                                            style="font-size: 18px;"
                                                            >close</span
                                                        >
                                                    </button>
                                                </div>
                                                <p
                                                    class="text-xs text-slate-500 dark:text-slate-400 mb-3"
                                                >
                                                    Masukkan token yang telah
                                                    dikirimkan ke email Anda
                                                    untuk membuka akses file
                                                    ini.
                                                </p>
                                                <div
                                                    class="flex items-center gap-2"
                                                >
                                                    <input
                                                        type="text"
                                                        bind:value={tokenValue}
                                                        placeholder="Masukkan token akses Anda"
                                                        class="flex-1 px-3 py-2.5 text-sm bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-600 rounded-lg focus:ring-2 focus:ring-sky-400 focus:border-sky-400 text-slate-900 dark:text-white placeholder-slate-400 transition-all font-mono tracking-wider"
                                                        on:keydown={(e) =>
                                                            e.key === "Enter" &&
                                                            verifyToken()}
                                                    />
                                                    <button
                                                        on:click={verifyToken}
                                                        disabled={tokenLoading}
                                                        class="inline-flex items-center gap-1.5 px-4 py-2.5 bg-gradient-to-r from-sky-500 to-blue-500 hover:from-sky-600 hover:to-blue-600 text-white rounded-lg text-xs font-bold shadow-sm hover:shadow-md transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                                                    >
                                                        {#if tokenLoading}
                                                            <span
                                                                class="material-symbols-outlined animate-spin"
                                                                style="font-size: 16px;"
                                                                >progress_activity</span
                                                            >
                                                            Verifikasi...
                                                        {:else}
                                                            <span
                                                                class="material-symbols-outlined"
                                                                style="font-size: 16px;"
                                                                >lock_open</span
                                                            >
                                                            Buka
                                                        {/if}
                                                    </button>
                                                </div>
                                                {#if tokenError}
                                                    <div
                                                        class="flex items-start gap-2 mt-2 p-2.5 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800/30 rounded-lg"
                                                    >
                                                        <span
                                                            class="material-symbols-outlined text-red-500 flex-shrink-0"
                                                            style="font-size: 14px;"
                                                            >error</span
                                                        >
                                                        <p
                                                            class="text-xs text-red-600 dark:text-red-400"
                                                        >
                                                            {tokenError}
                                                        </p>
                                                    </div>
                                                {/if}
                                            </div>
                                        {/if}
                                    </div>
                                {/each}
                            </div>

                            {#if doc.files.some((f) => f.is_locked)}
                                <div
                                    class="px-6 py-3 bg-amber-50 dark:bg-amber-900/10 border-t border-amber-200/50 dark:border-amber-800/30"
                                >
                                    <p
                                        class="text-xs text-amber-600 dark:text-amber-400 flex items-center gap-1.5"
                                    >
                                        <span
                                            class="material-symbols-outlined"
                                            style="font-size: 14px;">info</span
                                        >
                                        Beberapa file dikunci. Klik
                                        <strong>"Minta Akses"</strong>
                                        untuk mengajukan permintaan, atau klik
                                        <strong>"Masukkan Token"</strong> jika Anda
                                        sudah memiliki token akses dari email.
                                    </p>
                                </div>
                            {/if}
                        </div>
                    {/if}
                </div>

                <!-- Sidebar (Right Column) -->
                <div class="space-y-6">
                    <!-- Download All Card -->
                    <div
                        class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-6 shadow-sm"
                    >
                        <h3
                            class="text-sm font-bold text-slate-900 dark:text-white mb-4 flex items-center gap-2"
                        >
                            <span
                                class="material-symbols-outlined text-primary"
                                style="font-size: 18px;">download</span
                            >
                            Unduh Dokumen
                        </h3>
                        <button
                            on:click={handleDownload}
                            class="w-full flex items-center justify-center gap-2 px-4 py-3 bg-gradient-to-r {typeStyle.gradient} text-white font-bold rounded-lg shadow-lg hover:shadow-xl hover:opacity-90 transition-all active:scale-[0.98]"
                        >
                            <span class="material-symbols-outlined"
                                >download</span
                            >
                            Download
                        </button>
                        {#if doc.files && doc.files.length > 1}
                            <p class="text-xs text-slate-400 mt-2 text-center">
                                {doc.files.length} file akan diunduh dalam format
                                ZIP
                            </p>
                        {/if}
                    </div>

                    <!-- Metadata Card -->
                    <div
                        class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-6 shadow-sm"
                    >
                        <h3
                            class="text-sm font-bold text-slate-900 dark:text-white mb-4 flex items-center gap-2"
                        >
                            <span
                                class="material-symbols-outlined text-primary"
                                style="font-size: 18px;">label</span
                            >
                            Metadata
                        </h3>
                        <div class="space-y-3">
                            <div class="flex items-center gap-3 text-sm">
                                <span
                                    class="material-symbols-outlined text-slate-400"
                                    style="font-size: 18px;">category</span
                                >
                                <div>
                                    <p
                                        class="text-[11px] uppercase tracking-wider text-slate-400 font-semibold"
                                    >
                                        Jenis
                                    </p>
                                    <p
                                        class="text-slate-900 dark:text-white font-medium"
                                    >
                                        {doc.jenis_file}
                                    </p>
                                </div>
                            </div>
                            <div class="flex items-center gap-3 text-sm">
                                <span
                                    class="material-symbols-outlined text-slate-400"
                                    style="font-size: 18px;">badge</span
                                >
                                <div>
                                    <p
                                        class="text-[11px] uppercase tracking-wider text-slate-400 font-semibold"
                                    >
                                        Status
                                    </p>
                                    <span
                                        class="inline-block px-2 py-0.5 text-xs font-bold rounded-full {doc.status ===
                                        'publish'
                                            ? 'bg-green-100 dark:bg-green-900/30 text-green-600'
                                            : 'bg-slate-100 dark:bg-slate-700 text-slate-500'}"
                                    >
                                        {doc.status === "publish"
                                            ? "Published"
                                            : "Draft"}
                                    </span>
                                </div>
                            </div>
                            {#if doc.fakultas_nama}
                                <div class="flex items-center gap-3 text-sm">
                                    <span
                                        class="material-symbols-outlined text-slate-400"
                                        style="font-size: 18px;">apartment</span
                                    >
                                    <div>
                                        <p
                                            class="text-[11px] uppercase tracking-wider text-slate-400 font-semibold"
                                        >
                                            Fakultas
                                        </p>
                                        <p
                                            class="text-slate-900 dark:text-white font-medium"
                                        >
                                            {doc.fakultas_nama}
                                        </p>
                                    </div>
                                </div>
                            {/if}
                            {#if doc.prodi_nama}
                                <div class="flex items-center gap-3 text-sm">
                                    <span
                                        class="material-symbols-outlined text-slate-400"
                                        style="font-size: 18px;">school</span
                                    >
                                    <div>
                                        <p
                                            class="text-[11px] uppercase tracking-wider text-slate-400 font-semibold"
                                        >
                                            Program Studi
                                        </p>
                                        <p
                                            class="text-slate-900 dark:text-white font-medium"
                                        >
                                            {doc.prodi_nama}
                                        </p>
                                    </div>
                                </div>
                            {/if}
                            {#if doc.dosen_pembimbing}
                                <div class="flex items-center gap-3 text-sm">
                                    <span
                                        class="material-symbols-outlined text-slate-400"
                                        style="font-size: 18px;"
                                        >supervisor_account</span
                                    >
                                    <div>
                                        <p
                                            class="text-[11px] uppercase tracking-wider text-slate-400 font-semibold"
                                        >
                                            Dosen Pembimbing
                                        </p>
                                        <p
                                            class="text-slate-900 dark:text-white font-medium"
                                        >
                                            {doc.dosen_pembimbing}
                                        </p>
                                    </div>
                                </div>
                            {/if}
                            {#if doc.view_count > 0}
                                <div class="flex items-center gap-3 text-sm">
                                    <span
                                        class="material-symbols-outlined text-slate-400"
                                        style="font-size: 18px;"
                                        >visibility</span
                                    >
                                    <div>
                                        <p
                                            class="text-[11px] uppercase tracking-wider text-slate-400 font-semibold"
                                        >
                                            Total Dilihat
                                        </p>
                                        <p
                                            class="text-slate-900 dark:text-white font-medium"
                                        >
                                            {doc.view_count} kali
                                        </p>
                                    </div>
                                </div>
                            {/if}
                        </div>
                    </div>

                    <!-- Share Card -->
                    <div
                        class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-6 shadow-sm"
                    >
                        <h3
                            class="text-sm font-bold text-slate-900 dark:text-white mb-4 flex items-center gap-2"
                        >
                            <span
                                class="material-symbols-outlined text-primary"
                                style="font-size: 18px;">share</span
                            >
                            Bagikan
                        </h3>
                        <div class="flex items-center gap-2">
                            <input
                                type="text"
                                readonly
                                value={window.location.href}
                                class="flex-1 px-3 py-2 bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300 text-xs rounded-lg border-none truncate"
                            />
                            <button
                                on:click={() => {
                                    navigator.clipboard.writeText(
                                        window.location.href,
                                    );
                                    alert("Link berhasil disalin!");
                                }}
                                class="px-3 py-2 bg-primary/10 text-primary hover:bg-primary hover:text-white rounded-lg text-xs font-bold transition-all"
                                title="Salin link"
                            >
                                <span
                                    class="material-symbols-outlined"
                                    style="font-size: 16px;">content_copy</span
                                >
                            </button>
                        </div>
                    </div>

                    <!-- Back to Browse -->
                    <a
                        href="#/browse"
                        class="flex items-center justify-center gap-2 px-4 py-3 bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-600 rounded-xl font-medium text-sm transition-all"
                    >
                        <span
                            class="material-symbols-outlined"
                            style="font-size: 18px;">arrow_back</span
                        >
                        Kembali ke Jelajah
                    </a>
                </div>
            </div>
        </div>
    {/if}
</div>

<style>
    @keyframes slideDown {
        from {
            opacity: 0;
            transform: translateY(-8px);
            max-height: 0;
        }
        to {
            opacity: 1;
            transform: translateY(0);
            max-height: 500px;
        }
    }
    .animate-slideDown {
        animation: slideDown 0.3s ease-out forwards;
    }
</style>
