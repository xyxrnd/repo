<script>
    import { onMount } from "svelte";
    import { link } from "svelte-spa-router";
    import authService from "../../services/authService";
    import { getDocuments } from "../../services/documentService";
    import { API_ENDPOINTS } from "../../config/index.js";

    let user = null;
    let documents = [];
    let loading = true;
    let currentTime = new Date();

    // Stats
    let totalDocs = 0;
    let publishedDocs = 0;
    let draftDocs = 0;
    let totalUsers = 0;
    let recentDocs = [];

    onMount(() => {
        user = authService.getUser();
        loadData();

        // Update time every minute
        const interval = setInterval(() => {
            currentTime = new Date();
        }, 60000);

        return () => clearInterval(interval);
    });

    async function loadData() {
        try {
            loading = true;
            // Fetch documents
            documents = await getDocuments();
            totalDocs = documents.length;
            publishedDocs = documents.filter(
                (d) => d.status === "publish",
            ).length;
            draftDocs = documents.filter((d) => d.status === "draft").length;
            recentDocs = documents.slice(0, 5);

            // Fetch stats (termasuk total_users)
            try {
                const token = authService.getToken();
                const res = await fetch(API_ENDPOINTS.STATS, {
                    headers: token ? { Authorization: `Bearer ${token}` } : {},
                });
                if (res.ok) {
                    const stats = await res.json();
                    totalUsers = stats.total_users || 0;
                }
            } catch (e) {
                console.error("Failed to load stats:", e);
            }
        } catch (e) {
            console.error("Failed to load data:", e);
        } finally {
            loading = false;
        }
    }

    function formatDate(dateString) {
        const date = new Date(dateString);
        return date.toLocaleDateString("id-ID", {
            day: "numeric",
            month: "short",
            year: "numeric",
        });
    }

    function formatTime(date) {
        return date.toLocaleTimeString("id-ID", {
            hour: "2-digit",
            minute: "2-digit",
        });
    }

    function formatFullDate(date) {
        return date.toLocaleDateString("id-ID", {
            weekday: "long",
            day: "numeric",
            month: "long",
            year: "numeric",
        });
    }

    function getGreeting() {
        const hour = currentTime.getHours();
        if (hour < 12) return "Selamat Pagi";
        if (hour < 15) return "Selamat Siang";
        if (hour < 18) return "Selamat Sore";
        return "Selamat Malam";
    }

    function getFileIcon(jenisFile) {
        const icons = {
            pdf: { icon: "picture_as_pdf", color: "text-red-500" },
            docx: { icon: "description", color: "text-blue-500" },
            xlsx: { icon: "table_chart", color: "text-green-500" },
            pptx: { icon: "slideshow", color: "text-orange-500" },
        };
        return (
            icons[jenisFile] || { icon: "description", color: "text-slate-500" }
        );
    }
</script>

<div class="max-w-7xl mx-auto flex flex-col gap-8">
    <!-- Welcome Header -->
    <div
        class="relative overflow-hidden rounded-2xl bg-gradient-to-br from-primary via-blue-600 to-indigo-700 p-8 text-white shadow-xl"
    >
        <!-- Decorative Elements -->
        <div
            class="absolute top-0 right-0 w-64 h-64 bg-white/10 rounded-full -translate-y-1/2 translate-x-1/2 blur-3xl"
        ></div>
        <div
            class="absolute bottom-0 left-0 w-48 h-48 bg-white/10 rounded-full translate-y-1/2 -translate-x-1/2 blur-2xl"
        ></div>

        <div
            class="relative z-10 flex flex-col md:flex-row justify-between items-start md:items-center gap-6"
        >
            <div class="flex flex-col gap-2">
                <p class="text-white/70 text-sm font-medium">
                    {formatFullDate(currentTime)}
                </p>
                <h1 class="text-3xl md:text-4xl font-black tracking-tight">
                    {getGreeting()}, {user?.name?.split(" ")[0] || "Admin"}! 👋
                </h1>
                <p class="text-white/80 text-lg">
                    Selamat datang kembali di panel administrasi.
                </p>
            </div>
            <div class="flex flex-col items-end gap-1">
                <div class="text-5xl font-black tracking-tight">
                    {formatTime(currentTime)}
                </div>
                <span
                    class="px-3 py-1 bg-white/20 backdrop-blur-sm rounded-full text-xs font-semibold"
                >
                    {user?.role === "admin" ? "Administrator" : "User"}
                </span>
            </div>
        </div>
    </div>

    <!-- Quick Stats -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <!-- Total Documents -->
        <div
            class="group relative bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 p-6 shadow-sm hover:shadow-lg hover:border-primary/30 transition-all duration-300"
        >
            <div class="flex items-start justify-between">
                <div class="flex flex-col gap-1">
                    <span
                        class="text-sm font-medium text-slate-500 dark:text-slate-400"
                    >
                        Total Dokumen
                    </span>
                    {#if loading}
                        <div
                            class="h-9 w-16 bg-slate-200 dark:bg-slate-700 rounded animate-pulse"
                        ></div>
                    {:else}
                        <span
                            class="text-3xl font-black text-slate-900 dark:text-white"
                        >
                            {totalDocs}
                        </span>
                    {/if}
                </div>
                <div
                    class="p-3 bg-gradient-to-br from-blue-500 to-blue-600 rounded-xl shadow-lg shadow-blue-500/25 group-hover:scale-110 transition-transform"
                >
                    <span class="material-symbols-outlined text-white text-2xl"
                        >folder</span
                    >
                </div>
            </div>
            <div class="mt-4 flex items-center gap-2 text-sm">
                <span
                    class="text-emerald-500 font-semibold flex items-center gap-1"
                >
                    <span class="material-symbols-outlined text-base"
                        >trending_up</span
                    >
                    +12%
                </span>
                <span class="text-slate-400">dari bulan lalu</span>
            </div>
        </div>

        <!-- Published -->
        <div
            class="group relative bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 p-6 shadow-sm hover:shadow-lg hover:border-emerald-500/30 transition-all duration-300"
        >
            <div class="flex items-start justify-between">
                <div class="flex flex-col gap-1">
                    <span
                        class="text-sm font-medium text-slate-500 dark:text-slate-400"
                    >
                        Dipublikasi
                    </span>
                    {#if loading}
                        <div
                            class="h-9 w-16 bg-slate-200 dark:bg-slate-700 rounded animate-pulse"
                        ></div>
                    {:else}
                        <span
                            class="text-3xl font-black text-slate-900 dark:text-white"
                        >
                            {publishedDocs}
                        </span>
                    {/if}
                </div>
                <div
                    class="p-3 bg-gradient-to-br from-emerald-500 to-emerald-600 rounded-xl shadow-lg shadow-emerald-500/25 group-hover:scale-110 transition-transform"
                >
                    <span class="material-symbols-outlined text-white text-2xl"
                        >check_circle</span
                    >
                </div>
            </div>
            <div class="mt-4 flex items-center gap-2 text-sm">
                <span
                    class="text-emerald-500 font-semibold flex items-center gap-1"
                >
                    <span class="material-symbols-outlined text-base"
                        >trending_up</span
                    >
                    +8%
                </span>
                <span class="text-slate-400">dari bulan lalu</span>
            </div>
        </div>

        <!-- Drafts -->
        <div
            class="group relative bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 p-6 shadow-sm hover:shadow-lg hover:border-amber-500/30 transition-all duration-300"
        >
            <div class="flex items-start justify-between">
                <div class="flex flex-col gap-1">
                    <span
                        class="text-sm font-medium text-slate-500 dark:text-slate-400"
                    >
                        Draft
                    </span>
                    {#if loading}
                        <div
                            class="h-9 w-16 bg-slate-200 dark:bg-slate-700 rounded animate-pulse"
                        ></div>
                    {:else}
                        <span
                            class="text-3xl font-black text-slate-900 dark:text-white"
                        >
                            {draftDocs}
                        </span>
                    {/if}
                </div>
                <div
                    class="p-3 bg-gradient-to-br from-amber-500 to-amber-600 rounded-xl shadow-lg shadow-amber-500/25 group-hover:scale-110 transition-transform"
                >
                    <span class="material-symbols-outlined text-white text-2xl"
                        >edit_note</span
                    >
                </div>
            </div>
            <div class="mt-4 flex items-center gap-2 text-sm">
                <span
                    class="text-amber-500 font-semibold flex items-center gap-1"
                >
                    <span class="material-symbols-outlined text-base"
                        >schedule</span
                    >
                    Menunggu
                </span>
                <span class="text-slate-400">review</span>
            </div>
        </div>

        <!-- Users (Admin only) -->
        <div
            class="group relative bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 p-6 shadow-sm hover:shadow-lg hover:border-purple-500/30 transition-all duration-300"
        >
            <div class="flex items-start justify-between">
                <div class="flex flex-col gap-1">
                    <span
                        class="text-sm font-medium text-slate-500 dark:text-slate-400"
                    >
                        Pengguna Aktif
                    </span>
                    {#if loading}
                        <div
                            class="h-9 w-16 bg-slate-200 dark:bg-slate-700 rounded animate-pulse"
                        ></div>
                    {:else}
                        <span
                            class="text-3xl font-black text-slate-900 dark:text-white"
                        >
                            {totalUsers}
                        </span>
                    {/if}
                </div>
                <div
                    class="p-3 bg-gradient-to-br from-purple-500 to-purple-600 rounded-xl shadow-lg shadow-purple-500/25 group-hover:scale-110 transition-transform"
                >
                    <span class="material-symbols-outlined text-white text-2xl"
                        >group</span
                    >
                </div>
            </div>
            <div class="mt-4 flex items-center gap-2 text-sm">
                <span
                    class="text-purple-500 font-semibold flex items-center gap-1"
                >
                    <span class="material-symbols-outlined text-base"
                        >person_add</span
                    >
                    +3
                </span>
                <span class="text-slate-400">user baru hari ini</span>
            </div>
        </div>
    </div>

    <!-- Main Content Grid -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Recent Documents -->
        <div
            class="lg:col-span-2 bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden"
        >
            <div
                class="p-6 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between"
            >
                <div>
                    <h2
                        class="text-lg font-bold text-slate-900 dark:text-white"
                    >
                        Dokumen Terbaru
                    </h2>
                    <p class="text-sm text-slate-500">
                        5 dokumen terakhir yang diupload
                    </p>
                </div>
                <a
                    href="/documents"
                    use:link
                    class="text-sm font-semibold text-primary hover:text-primary/80 flex items-center gap-1 transition-colors"
                >
                    Lihat Semua
                    <span class="material-symbols-outlined text-lg"
                        >arrow_forward</span
                    >
                </a>
            </div>

            <div class="divide-y divide-slate-100 dark:divide-slate-800">
                {#if loading}
                    {#each Array(5) as _}
                        <div class="p-4 flex items-center gap-4">
                            <div
                                class="w-10 h-10 bg-slate-200 dark:bg-slate-700 rounded-lg animate-pulse"
                            ></div>
                            <div class="flex-1 flex flex-col gap-2">
                                <div
                                    class="h-4 w-48 bg-slate-200 dark:bg-slate-700 rounded animate-pulse"
                                ></div>
                                <div
                                    class="h-3 w-32 bg-slate-100 dark:bg-slate-800 rounded animate-pulse"
                                ></div>
                            </div>
                        </div>
                    {/each}
                {:else if recentDocs.length === 0}
                    <div class="p-12 text-center">
                        <span
                            class="material-symbols-outlined text-5xl text-slate-300 dark:text-slate-700"
                        >
                            folder_off
                        </span>
                        <p class="mt-2 text-slate-500">Belum ada dokumen</p>
                    </div>
                {:else}
                    {#each recentDocs as doc}
                        <div
                            class="p-4 flex items-center gap-4 hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors"
                        >
                            <div
                                class="w-10 h-10 rounded-lg bg-slate-100 dark:bg-slate-800 flex items-center justify-center {getFileIcon(
                                    doc.jenis_file,
                                ).color}"
                            >
                                <span class="material-symbols-outlined">
                                    {getFileIcon(doc.jenis_file).icon}
                                </span>
                            </div>
                            <div class="flex-1 min-w-0">
                                <h3
                                    class="font-semibold text-slate-900 dark:text-white truncate"
                                >
                                    {doc.judul}
                                </h3>
                                <p class="text-sm text-slate-500 truncate">
                                    {doc.penulis} • {formatDate(doc.created_at)}
                                </p>
                            </div>
                            <span
                                class="px-2.5 py-1 text-xs font-bold uppercase rounded-full {doc.status ===
                                'publish'
                                    ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
                                    : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'}"
                            >
                                {doc.status}
                            </span>
                        </div>
                    {/each}
                {/if}
            </div>
        </div>

        <!-- Quick Actions -->
        <div class="flex flex-col gap-6">
            <!-- Quick Actions Card -->
            <div
                class="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 shadow-sm p-6"
            >
                <h2
                    class="text-lg font-bold text-slate-900 dark:text-white mb-4"
                >
                    Aksi Cepat
                </h2>
                <div class="flex flex-col gap-3">
                    <a
                        href="/documents/add"
                        use:link
                        class="flex items-center gap-3 p-4 rounded-xl bg-gradient-to-r from-primary to-blue-600 text-white hover:shadow-lg hover:shadow-primary/25 transition-all group"
                    >
                        <div
                            class="p-2 bg-white/20 rounded-lg group-hover:scale-110 transition-transform"
                        >
                            <span class="material-symbols-outlined">add</span>
                        </div>
                        <div>
                            <span class="font-bold">Upload Dokumen</span>
                            <p class="text-sm text-white/80">
                                Tambah dokumen baru
                            </p>
                        </div>
                    </a>

                    <a
                        href="/documents"
                        use:link
                        class="flex items-center gap-3 p-4 rounded-xl bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 transition-all group"
                    >
                        <div
                            class="p-2 bg-white dark:bg-slate-700 rounded-lg shadow-sm group-hover:scale-110 transition-transform"
                        >
                            <span
                                class="material-symbols-outlined text-slate-600 dark:text-slate-300"
                                >folder_open</span
                            >
                        </div>
                        <div>
                            <span
                                class="font-bold text-slate-900 dark:text-white"
                                >Kelola Dokumen</span
                            >
                            <p class="text-sm text-slate-500">
                                Lihat semua file
                            </p>
                        </div>
                    </a>

                    {#if user?.role === "admin"}
                        <a
                            href="/users"
                            use:link
                            class="flex items-center gap-3 p-4 rounded-xl bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 transition-all group"
                        >
                            <div
                                class="p-2 bg-white dark:bg-slate-700 rounded-lg shadow-sm group-hover:scale-110 transition-transform"
                            >
                                <span
                                    class="material-symbols-outlined text-slate-600 dark:text-slate-300"
                                    >manage_accounts</span
                                >
                            </div>
                            <div>
                                <span
                                    class="font-bold text-slate-900 dark:text-white"
                                    >Manajemen User</span
                                >
                                <p class="text-sm text-slate-500">
                                    Kelola pengguna
                                </p>
                            </div>
                        </a>
                    {/if}
                </div>
            </div>

            <!-- System Info -->
            <div
                class="bg-gradient-to-br from-slate-900 to-slate-800 dark:from-slate-800 dark:to-slate-900 rounded-2xl p-6 text-white"
            >
                <div class="flex items-center gap-3 mb-4">
                    <div class="p-2 bg-white/10 rounded-lg">
                        <span class="material-symbols-outlined">info</span>
                    </div>
                    <h2 class="font-bold">Info Sistem</h2>
                </div>
                <div class="space-y-3 text-sm">
                    <div class="flex justify-between">
                        <span class="text-white/60">Versi Aplikasi</span>
                        <span class="font-semibold">1.0.0</span>
                    </div>
                    <div class="flex justify-between">
                        <span class="text-white/60">Status Server</span>
                        <span
                            class="flex items-center gap-1.5 font-semibold text-emerald-400"
                        >
                            <span
                                class="w-2 h-2 bg-emerald-400 rounded-full animate-pulse"
                            ></span>
                            Online
                        </span>
                    </div>
                    <div class="flex justify-between">
                        <span class="text-white/60">Database</span>
                        <span class="font-semibold text-emerald-400"
                            >Terhubung</span
                        >
                    </div>
                    <div class="flex justify-between">
                        <span class="text-white/60">Penyimpanan</span>
                        <span class="font-semibold">45% terpakai</span>
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>
