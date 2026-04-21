<script>
    import { onMount } from "svelte";
    import { querystring } from "svelte-spa-router";
    import { getDocumentById } from "../../services/documentService";
    import authService from "../../services/authService";
    import { API_BASE_URL, API_ENDPOINTS } from "../../config";

    export let id = "";
    export let params = {};

    let doc = null;
    let loading = true;
    let error = "";
    let isLoggedIn = false;

    // Token verification state (backward compatibility untuk link lama dari email)
    let tokenLoading = false;
    let tokenError = "";
    let tokenSuccess = false;
    let tokenAutoVerified = false;
    let unlockedFiles = {}; // { file_id: { file_path, file_name } }

    // Get ID from route params (prop passed by svelte-spa-router)
    $: if (params && params.id) {
        id = params.id;
        loadDocument();
    }

    // Auto-detect token from URL query parameter (backward compatibility)
    $: if ($querystring && doc && !tokenAutoVerified) {
        const urlParams = new URLSearchParams($querystring);
        const urlToken = urlParams.get("token");
        if (urlToken) {
            tokenAutoVerified = true;
            autoVerifyToken(urlToken);
        }
    }

    onMount(() => {
        isLoggedIn = authService.isAuthenticated();
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

    // Helper: cek apakah file_path adalah Google Drive ID (bukan path lokal)
    function isGDriveId(path) {
        if (!path) return false;
        return !path.includes('/') && !path.includes('\\') && !path.includes('.');
    }

    // Helper: dapatkan URL file yang benar
    function getFileUrl(filePath) {
        if (!filePath) return '#';
        if (isGDriveId(filePath)) {
            return `${API_BASE_URL}/api/gdrive-proxy/${filePath}`;
        }
        return `${API_BASE_URL}/${filePath}`;
    }

    function handleDownload() {
        // Download semua file dalam ZIP via endpoint download-all
        window.open(API_ENDPOINTS.DOCUMENT_DOWNLOAD_ALL(doc.id), "_blank");
    }

    function handleFileDownload(file) {
        // Jika file terkunci dan user belum login, arahkan ke login
        if (file.is_locked && !isLoggedIn && !unlockedFiles[file.id]) {
            // Simpan halaman saat ini agar setelah login kembali ke sini
            sessionStorage.setItem("redirectAfterLogin", `#/document/${doc.id}`);
            window.location.hash = "/login";
            return;
        }
        window.open(getFileUrl(file.file_path), "_blank");
    }

    function handleFilePreview(file) {
        // Jika file terkunci dan user belum login, arahkan ke login
        if (file.is_locked && !isLoggedIn && !unlockedFiles[file.id]) {
            // Simpan halaman saat ini agar setelah login kembali ke sini
            sessionStorage.setItem("redirectAfterLogin", `#/document/${doc.id}`);
            window.location.hash = "/login";
            return;
        }
        window.open(getFileUrl(file.file_path), "_blank");
    }

    // Auto-verify token from URL (backward compatibility - link dari email lama)
    async function autoVerifyToken(token) {
        tokenLoading = true;
        tokenError = "";
        tokenSuccess = false;
        try {
            await doVerifyToken(token);
            tokenSuccess = true;
        } finally {
            tokenLoading = false;
        }
    }

    async function doVerifyToken(token) {
        const response = await fetch(API_ENDPOINTS.VERIFY_ACCESS_TOKEN, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ document_id: doc.id, token }),
        });
        if (!response.ok) {
            const text = await response.text();
            tokenError = text || "Token tidak valid atau belum disetujui.";
            return;
        }
        const data = await response.json();
        // Unlock all locked files from response
        if (data.files && data.files.length > 0) {
            let newUnlocked = { ...unlockedFiles };
            for (const f of data.files) {
                newUnlocked[f.file_id] = {
                    file_path: f.file_path,
                    file_name: f.file_name,
                };
            }
            unlockedFiles = newUnlocked;
        }
    }

    function isFileUnlocked(fileId) {
        return !!unlockedFiles[fileId];
    }

    function handleUnlockedPreview(fileId) {
        const info = unlockedFiles[fileId];
        if (info) {
            window.open(getFileUrl(info.file_path), "_blank");
        }
    }

    function handleUnlockedDownload(fileId) {
        const info = unlockedFiles[fileId];
        if (info) {
            const link = document.createElement("a");
            link.href = getFileUrl(info.file_path);
            link.download = info.file_name;
            link.click();
        }
    }

    // Check if document has any locked files
    $: hasLockedFiles = doc && doc.files && doc.files.some((f) => f.is_locked);
    // All locked files are "unlocked" if user is logged in OR has token-based unlock
    $: allLockedUnlocked =
        isLoggedIn ||
        (doc &&
        doc.files &&
        doc.files.filter((f) => f.is_locked).every((f) => unlockedFiles[f.id]));

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
                                                >Dosen Pembimbing 1</td
                                            >
                                            <td
                                                class="py-3 text-sm text-slate-900 dark:text-white"
                                                >{doc.dosen_pembimbing}</td
                                            >
                                        </tr>
                                    {/if}
                                    {#if doc.dosen_pembimbing_2}
                                        <tr
                                            class="border-b border-slate-100 dark:border-slate-700/50"
                                        >
                                            <td
                                                class="py-3 pr-4 text-sm font-semibold text-slate-500 dark:text-slate-400 w-40 align-top"
                                                >Dosen Pembimbing 2</td
                                            >
                                            <td
                                                class="py-3 text-sm text-slate-900 dark:text-white"
                                                >{doc.dosen_pembimbing_2}</td
                                            >
                                        </tr>
                                    {/if}
                                    {#if doc.kata_kunci}
                                        <tr
                                            class="border-b border-slate-100 dark:border-slate-700/50"
                                        >
                                            <td
                                                class="py-3 pr-4 text-sm font-semibold text-slate-500 dark:text-slate-400 w-40 align-top"
                                                >Kata Kunci</td
                                            >
                                            <td class="py-3">
                                                <div
                                                    class="flex flex-wrap gap-1.5"
                                                >
                                                    {#each doc.kata_kunci
                                                        .split(",")
                                                        .map((k) => k.trim())
                                                        .filter((k) => k) as keyword}
                                                        <span
                                                            class="inline-flex items-center gap-1 px-2.5 py-1 bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400 text-xs font-semibold rounded-full border border-blue-100 dark:border-blue-800/30"
                                                        >
                                                            <span
                                                                class="material-symbols-outlined"
                                                                style="font-size: 12px;"
                                                                >tag</span
                                                            >
                                                            {keyword}
                                                        </span>
                                                    {/each}
                                                </div>
                                            </td>
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
                                    {doc.penulis} ({doc.tahun && doc.tahun > 0 ? doc.tahun : new Date(
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
                                    const citationText = `${doc.penulis} (${doc.tahun && doc.tahun > 0 ? doc.tahun : new Date(doc.created_at).getFullYear()}) ${doc.judul}. ${doc.jenis_file ? doc.jenis_file.charAt(0).toUpperCase() + doc.jenis_file.slice(1) + ", " : ""}${doc.fakultas_nama ? doc.fakultas_nama : ""}${doc.prodi_nama ? ", " + doc.prodi_nama : ""}.`;
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
                                                class="w-10 h-10 rounded-lg {file.is_locked &&
                                                !isLoggedIn &&
                                                !unlockedFiles[file.id]
                                                    ? 'bg-amber-100 dark:bg-amber-900/30'
                                                    : file.is_locked &&
                                                        (isLoggedIn ||
                                                            unlockedFiles[
                                                                file.id
                                                            ])
                                                      ? 'bg-green-100 dark:bg-green-900/30'
                                                      : 'bg-red-100 dark:bg-red-900/30'} flex items-center justify-center flex-shrink-0"
                                            >
                                                {#if file.is_locked && !isLoggedIn && !unlockedFiles[file.id]}
                                                    <span
                                                        class="material-symbols-outlined text-amber-500"
                                                        style="font-size: 22px;"
                                                        >lock</span
                                                    >
                                                {:else if file.is_locked && (isLoggedIn || unlockedFiles[file.id])}
                                                    <span
                                                        class="material-symbols-outlined text-green-500"
                                                        style="font-size: 22px;"
                                                        >lock_open</span
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
                                                    {#if file.is_locked && !isLoggedIn && !unlockedFiles[file.id]}
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
                                                    {:else if file.is_locked && (isLoggedIn || unlockedFiles[file.id])}
                                                        <span
                                                            class="inline-flex items-center gap-0.5 px-1.5 py-0.5 bg-green-100 dark:bg-green-900/30 text-green-600 dark:text-green-400 text-[10px] font-bold rounded uppercase"
                                                        >
                                                            <span
                                                                class="material-symbols-outlined"
                                                                style="font-size: 11px;"
                                                                >lock_open</span
                                                            >
                                                            Terbuka
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
                                                {#if file.is_locked && !isLoggedIn && !unlockedFiles[file.id]}
                                                    <!-- File terkunci & belum login -->
                                                    <button
                                                        on:click={() =>
                                                            handleFilePreview(
                                                                file,
                                                            )}
                                                        class="inline-flex items-center gap-1 px-2 py-1 bg-amber-100 dark:bg-amber-900/30 text-amber-600 dark:text-amber-400 text-[10px] font-bold rounded-lg hover:bg-amber-200 dark:hover:bg-amber-900/50 transition-all"
                                                        title="Login untuk membuka"
                                                    >
                                                        <span
                                                            class="material-symbols-outlined"
                                                            style="font-size: 13px;"
                                                            >lock</span
                                                        >
                                                        Login untuk Akses
                                                    </button>
                                                {:else}
                                                    <!-- File tidak terkunci ATAU sudah login/token unlock -->
                                                    <button
                                                        on:click={() =>
                                                            file.is_locked && unlockedFiles[file.id]
                                                                ? handleUnlockedPreview(file.id)
                                                                : handleFilePreview(file)}
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
                                                            file.is_locked && unlockedFiles[file.id]
                                                                ? handleUnlockedDownload(file.id)
                                                                : handleFileDownload(file)}
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
                                    </div>
                                {/each}
                            </div>

                            <!-- Document-level Access Section (muncul di bawah daftar file) -->
                            {#if hasLockedFiles && !allLockedUnlocked}
                                <div
                                    class="px-6 py-4 bg-amber-50 dark:bg-amber-900/10 border-t border-amber-200/50 dark:border-amber-800/30"
                                >
                                    {#if !isLoggedIn}
                                        <!-- Belum login: arahkan untuk login -->
                                        <div class="flex items-start gap-3">
                                            <span
                                                class="material-symbols-outlined text-amber-500 flex-shrink-0 mt-0.5"
                                                style="font-size: 20px;"
                                                >lock</span
                                            >
                                            <div>
                                                <p
                                                    class="text-sm font-semibold text-slate-900 dark:text-white mb-1"
                                                >
                                                    File Terkunci
                                                </p>
                                                <p
                                                    class="text-xs text-slate-600 dark:text-slate-400 mb-3"
                                                >
                                                    Beberapa file pada dokumen
                                                    ini dikunci. Silakan login
                                                    untuk membuka semua file
                                                    terkunci pada dokumen ini.
                                                </p>
                                                <div
                                                    class="flex items-center gap-2"
                                                >
                                                    <a
                                                        href="#/login"
                                                        on:click={() => sessionStorage.setItem('redirectAfterLogin', `#/document/${doc.id}`)}
                                                        class="inline-flex items-center gap-1.5 px-4 py-2 bg-gradient-to-r from-primary to-blue-600 hover:from-primary/90 hover:to-blue-700 text-white rounded-lg text-xs font-bold shadow-sm transition-all"
                                                    >
                                                        <span
                                                            class="material-symbols-outlined"
                                                            style="font-size: 16px;"
                                                            >login</span
                                                        >
                                                        Login untuk Akses
                                                    </a>
                                                    <a
                                                        href="#/student-signup"
                                                        class="inline-flex items-center gap-1.5 px-4 py-2 bg-slate-200 dark:bg-slate-700 hover:bg-slate-300 dark:hover:bg-slate-600 text-slate-700 dark:text-slate-300 rounded-lg text-xs font-bold transition-all"
                                                    >
                                                        <span
                                                            class="material-symbols-outlined"
                                                            style="font-size: 16px;"
                                                            >person_add</span
                                                        >
                                                        Belum punya akun? Daftar
                                                    </a>
                                                </div>
                                            </div>
                                        </div>
                                    {:else if tokenLoading}
                                        <div
                                            class="flex items-center gap-3 justify-center py-2"
                                        >
                                            <span
                                                class="material-symbols-outlined animate-spin text-primary"
                                                style="font-size: 20px;"
                                                >progress_activity</span
                                            >
                                            <p
                                                class="text-sm text-slate-600 dark:text-slate-300 font-medium"
                                            >
                                                Memverifikasi token akses...
                                            </p>
                                        </div>
                                    {:else if tokenSuccess}
                                        <div
                                            class="flex items-start gap-2 p-3 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800/30 rounded-lg"
                                        >
                                            <span
                                                class="material-symbols-outlined text-green-500 flex-shrink-0"
                                                style="font-size: 16px;"
                                                >check_circle</span
                                            >
                                            <p
                                                class="text-xs text-green-700 dark:text-green-400 font-medium"
                                            >
                                                Token valid! Semua file terkunci
                                                pada dokumen ini telah dibuka.
                                            </p>
                                        </div>
                                    {:else if tokenError}
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
                                                {tokenError}
                                            </p>
                                        </div>
                                    {/if}
                                </div>
                            {/if}
                        </div>
                    {/if}
                </div>

                <!-- Sidebar (Right Column) -->
                <div class="space-y-6">
                    <!-- Download All Card (hanya tampil jika sudah login) -->
                    {#if isLoggedIn}
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
                                <p
                                    class="text-xs text-slate-400 mt-2 text-center"
                                >
                                    {doc.files.length} file akan diunduh dalam format
                                    ZIP
                                </p>
                            {/if}
                        </div>
                    {/if}

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
                            {#if doc.kata_kunci}
                                <div class="flex items-start gap-3 text-sm">
                                    <span
                                        class="material-symbols-outlined text-slate-400 mt-0.5"
                                        style="font-size: 18px;">tag</span
                                    >
                                    <div>
                                        <p
                                            class="text-[11px] uppercase tracking-wider text-slate-400 font-semibold"
                                        >
                                            Kata Kunci
                                        </p>
                                        <div class="flex flex-wrap gap-1 mt-1">
                                            {#each doc.kata_kunci
                                                .split(",")
                                                .map((k) => k.trim())
                                                .filter((k) => k) as keyword}
                                                <span
                                                    class="inline-block px-2 py-0.5 text-[11px] font-medium bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400 rounded-full"
                                                >
                                                    {keyword}
                                                </span>
                                            {/each}
                                        </div>
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
