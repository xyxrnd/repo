<script>
    import { onMount } from "svelte";
    import { getPopularDocuments } from "../../services/documentService";

    let loading = true;
    let popularDocs = [];

    onMount(() => {
        loadPopularDocuments();
    });

    async function loadPopularDocuments() {
        try {
            loading = true;
            const documents = await getPopularDocuments(6);
            popularDocs = documents || [];
        } catch (e) {
            console.error("Failed to load popular documents:", e);
            popularDocs = [];
        } finally {
            loading = false;
        }
    }

    // Get full class string for card based on jenis_file
    function getCardClasses(jenisFile) {
        const map = {
            Skripsi:
                "from-blue-600/20 to-blue-900/40 hover:border-blue-500/40 hover:shadow-blue-500/10",
            Tesis: "from-purple-600/20 to-purple-900/40 hover:border-purple-500/40 hover:shadow-purple-500/10",
            Jurnal: "from-teal-600/20 to-teal-900/40 hover:border-teal-500/40 hover:shadow-teal-500/10",
            Disertasi:
                "from-amber-600/20 to-amber-900/40 hover:border-amber-500/40 hover:shadow-amber-500/10",
        };
        return (
            map[jenisFile] ||
            "from-slate-600/20 to-slate-900/40 hover:border-slate-500/40 hover:shadow-slate-500/10"
        );
    }

    function getBadgeBg(jenisFile) {
        const map = {
            Skripsi: "bg-blue-500/90",
            Tesis: "bg-purple-500/90",
            Jurnal: "bg-teal-500/90",
            Disertasi: "bg-amber-500/90",
        };
        return map[jenisFile] || "bg-slate-500/90";
    }

    function getIconBg(jenisFile) {
        const map = {
            Skripsi: "bg-blue-500",
            Tesis: "bg-purple-500",
            Jurnal: "bg-teal-500",
            Disertasi: "bg-amber-500",
        };
        return map[jenisFile] || "bg-slate-500";
    }

    function getCategoryIcon(jenisFile) {
        const icons = {
            Skripsi: "school",
            Tesis: "workspace_premium",
            Jurnal: "article",
            Disertasi: "history_edu",
        };
        return icons[jenisFile] || "description";
    }

    function formatViews(count) {
        if (!count) return "0";
        if (count >= 1000) {
            return (count / 1000).toFixed(1) + "rb";
        }
        return count.toString();
    }

    function getRankBg(index) {
        if (index === 0)
            return "bg-gradient-to-br from-yellow-400 to-yellow-600";
        if (index === 1) return "bg-gradient-to-br from-slate-300 to-slate-500";
        return "bg-gradient-to-br from-amber-600 to-amber-800";
    }
</script>

<section class="py-12 bg-white dark:bg-[#131d27]">
    <div class="container mx-auto max-w-6xl px-4">
        <div class="flex items-center justify-between mb-8">
            <div>
                <h2
                    class="text-2xl md:text-3xl font-bold text-slate-900 dark:text-white mb-2"
                >
                    <span
                        class="material-symbols-outlined text-primary align-middle mr-2"
                        style="font-size: 1.8rem;">trending_up</span
                    >
                    Dokumen Unggulan
                </h2>
                <p class="text-slate-500 dark:text-slate-400">
                    Dokumen paling sering dibuka oleh pengunjung.
                </p>
            </div>
            <a
                class="hidden md:flex items-center gap-1 text-primary font-medium hover:underline"
                href="#/browse"
            >
                Lihat Semua <span class="material-symbols-outlined text-sm"
                    >arrow_forward_ios</span
                >
            </a>
        </div>

        {#if loading}
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {#each Array(6) as _}
                    <div
                        class="h-48 bg-slate-200 dark:bg-slate-700 rounded-xl animate-pulse"
                    ></div>
                {/each}
            </div>
        {:else if popularDocs.length === 0}
            <div class="text-center py-12">
                <span
                    class="material-symbols-outlined text-5xl text-slate-300 dark:text-slate-700 mb-4"
                    >folder_off</span
                >
                <p class="text-slate-500">Belum ada dokumen tersedia</p>
            </div>
        {:else}
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {#each popularDocs as doc, index}
                    <a
                        href="#/document/{doc.id}"
                        class="group relative overflow-hidden rounded-xl bg-gradient-to-br {getCardClasses(
                            doc.jenis_file,
                        )} border border-slate-200/10 dark:border-slate-700/50 transition-all duration-300 hover:shadow-lg hover:-translate-y-1 block"
                    >
                        <!-- Ranking badge -->
                        {#if index < 3}
                            <div
                                class="absolute top-3 left-3 z-10 w-8 h-8 rounded-full flex items-center justify-center text-white text-xs font-bold shadow-lg {getRankBg(
                                    index,
                                )}"
                            >
                                #{index + 1}
                            </div>
                        {/if}

                        <!-- View count badge -->
                        <div
                            class="absolute top-3 right-3 z-10 flex items-center gap-1 px-2.5 py-1 bg-black/40 backdrop-blur-sm rounded-full text-white text-xs font-medium"
                        >
                            <span
                                class="material-symbols-outlined"
                                style="font-size: 14px;">visibility</span
                            >
                            {formatViews(doc.view_count)}
                        </div>

                        <div class="p-5">
                            <!-- Icon & Category -->
                            <div class="flex items-start gap-3 mb-3">
                                <div
                                    class="w-11 h-11 rounded-lg {getIconBg(
                                        doc.jenis_file,
                                    )} flex items-center justify-center flex-shrink-0 shadow-md"
                                >
                                    <span
                                        class="material-symbols-outlined text-white"
                                        style="font-size: 22px;"
                                        >{getCategoryIcon(doc.jenis_file)}</span
                                    >
                                </div>
                                <div class="flex-1 min-w-0">
                                    <span
                                        class="inline-block px-2 py-0.5 {getBadgeBg(
                                            doc.jenis_file,
                                        )} text-white text-[10px] font-bold rounded-full mb-1.5 uppercase tracking-wider"
                                    >
                                        {doc.jenis_file}
                                    </span>
                                    <h3
                                        class="text-sm font-bold text-slate-900 dark:text-white leading-tight line-clamp-2 group-hover:text-primary transition-colors"
                                    >
                                        {doc.judul}
                                    </h3>
                                </div>
                            </div>

                            <!-- Author -->
                            <div
                                class="flex items-center gap-2 text-xs text-slate-500 dark:text-slate-400 mb-2"
                            >
                                <span
                                    class="material-symbols-outlined"
                                    style="font-size: 14px;">person</span
                                >
                                <span class="truncate">{doc.penulis}</span>
                            </div>

                            <!-- Prodi / Fakultas -->
                            {#if doc.prodi_nama || doc.fakultas_nama}
                                <div
                                    class="flex items-center gap-2 text-xs text-slate-400 dark:text-slate-500"
                                >
                                    <span
                                        class="material-symbols-outlined"
                                        style="font-size: 14px;">apartment</span
                                    >
                                    <span class="truncate">
                                        {doc.prodi_nama || doc.fakultas_nama}
                                    </span>
                                </div>
                            {/if}

                            <!-- Bottom bar -->
                            <div
                                class="flex items-center justify-between mt-4 pt-3 border-t border-slate-200/10 dark:border-slate-700/30"
                            >
                                <div
                                    class="flex items-center gap-1 text-xs text-slate-400 dark:text-slate-500"
                                >
                                    <span
                                        class="material-symbols-outlined"
                                        style="font-size: 14px;"
                                        >calendar_today</span
                                    >
                                    {new Date(
                                        doc.created_at,
                                    ).toLocaleDateString("id-ID", {
                                        day: "numeric",
                                        month: "short",
                                        year: "numeric",
                                    })}
                                </div>
                                <span
                                    class="material-symbols-outlined text-primary opacity-0 group-hover:opacity-100 transition-all duration-300 -translate-x-2 group-hover:translate-x-0"
                                    >arrow_right_alt</span
                                >
                            </div>
                        </div>
                    </a>
                {/each}
            </div>
        {/if}

        <div class="md:hidden mt-6 text-center">
            <a
                class="inline-flex items-center gap-1 text-primary font-medium hover:underline"
                href="#/browse"
            >
                Lihat Semua <span class="material-symbols-outlined text-sm"
                    >arrow_forward_ios</span
                >
            </a>
        </div>
    </div>
</section>

<style>
    .line-clamp-2 {
        display: -webkit-box;
        -webkit-line-clamp: 2;
        line-clamp: 2;
        -webkit-box-orient: vertical;
        overflow: hidden;
    }
</style>
