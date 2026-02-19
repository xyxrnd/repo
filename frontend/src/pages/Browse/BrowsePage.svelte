<script>
    import { onMount } from "svelte";
    import { getDocuments } from "../../services/documentService";
    import { querystring } from "svelte-spa-router";

    let documents = [];
    let filteredDocuments = [];
    let loading = true;
    let searchQuery = "";
    let selectedCategory = "Semua";
    let selectedSort = "terbaru";
    let viewMode = "grid";

    // Get categories from documents
    let categories = ["Semua"];

    // Parse query params
    $: {
        const params = new URLSearchParams($querystring);
        const q = params.get("q");
        const cat = params.get("category");
        if (q) searchQuery = q;
        if (cat) selectedCategory = cat;
    }

    onMount(() => {
        loadDocuments();
    });

    async function loadDocuments() {
        try {
            loading = true;
            documents = await getDocuments();

            // Extract unique categories
            const cats = new Set(documents.map((d) => d.jenis_file));
            categories = ["Semua", ...Array.from(cats)];

            filterDocuments();
        } catch (e) {
            console.error("Failed to load documents:", e);
        } finally {
            loading = false;
        }
    }

    function filterDocuments() {
        let result = [...documents];

        // Filter by search
        if (searchQuery) {
            const query = searchQuery.toLowerCase();
            result = result.filter(
                (doc) =>
                    doc.judul?.toLowerCase().includes(query) ||
                    doc.penulis?.toLowerCase().includes(query) ||
                    doc.jenis_file?.toLowerCase().includes(query),
            );
        }

        // Filter by category
        if (selectedCategory && selectedCategory !== "Semua") {
            result = result.filter(
                (doc) => doc.jenis_file === selectedCategory,
            );
        }

        // Sort
        if (selectedSort === "terbaru") {
            result.sort(
                (a, b) =>
                    new Date(b.created_at).getTime() -
                    new Date(a.created_at).getTime(),
            );
        } else if (selectedSort === "terlama") {
            result.sort(
                (a, b) =>
                    new Date(a.created_at).getTime() -
                    new Date(b.created_at).getTime(),
            );
        } else if (selectedSort === "judul") {
            result.sort((a, b) => (a.judul || "").localeCompare(b.judul || ""));
        }

        filteredDocuments = result;
    }

    $: searchQuery, selectedCategory, selectedSort, filterDocuments();

    function formatDate(dateString) {
        const date = new Date(dateString);
        return date.toLocaleDateString("id-ID", {
            day: "numeric",
            month: "long",
            year: "numeric",
        });
    }

    function getTypeStyle(jenisFile) {
        const styles = {
            PDF: {
                icon: "picture_as_pdf",
                color: "text-red-500 bg-red-100 dark:bg-red-900/30",
            },
            Skripsi: {
                icon: "school",
                color: "text-blue-500 bg-blue-100 dark:bg-blue-900/30",
            },
            Tesis: {
                icon: "menu_book",
                color: "text-purple-500 bg-purple-100 dark:bg-purple-900/30",
            },
            Jurnal: {
                icon: "article",
                color: "text-teal-500 bg-teal-100 dark:bg-teal-900/30",
            },
            Disertasi: {
                icon: "history_edu",
                color: "text-amber-500 bg-amber-100 dark:bg-amber-900/30",
            },
        };
        return (
            styles[jenisFile] || {
                icon: "description",
                color: "text-slate-500 bg-slate-100 dark:bg-slate-700",
            }
        );
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
    <!-- Hero Header -->
    <div
        class="bg-gradient-to-br from-primary via-blue-600 to-indigo-700 text-white py-12 px-4"
    >
        <div class="container mx-auto max-w-6xl">
            <h1 class="text-3xl md:text-4xl font-black mb-4">
                Jelajahi Repositori
            </h1>
            <p class="text-blue-100 text-lg max-w-2xl mb-8">
                Temukan ribuan karya ilmiah dari berbagai bidang dan kategori.
            </p>

            <!-- Search Bar -->
            <div class="flex flex-col md:flex-row gap-4 max-w-3xl">
                <div class="relative flex-1">
                    <span
                        class="material-symbols-outlined absolute left-4 top-1/2 -translate-y-1/2 text-slate-400"
                        >search</span
                    >
                    <input
                        type="text"
                        bind:value={searchQuery}
                        placeholder="Cari judul, penulis, atau kata kunci..."
                        class="w-full pl-12 pr-4 py-4 rounded-xl bg-white/10 backdrop-blur-sm border border-white/20 text-white placeholder:text-white/60 focus:bg-white/20 focus:outline-none focus:ring-2 focus:ring-white/50 transition-all"
                    />
                </div>
                <select
                    bind:value={selectedCategory}
                    class="px-4 py-4 rounded-xl bg-white/10 backdrop-blur-sm border border-white/20 text-white focus:bg-white/20 focus:outline-none focus:ring-2 focus:ring-white/50 cursor-pointer"
                >
                    {#each categories as cat}
                        <option value={cat} class="text-slate-900">{cat}</option
                        >
                    {/each}
                </select>
            </div>
        </div>
    </div>

    <!-- Content -->
    <div class="container mx-auto max-w-6xl px-4 py-8">
        <!-- Toolbar -->
        <div class="flex flex-wrap items-center justify-between gap-4 mb-6">
            <div
                class="flex items-center gap-2 text-slate-600 dark:text-slate-400"
            >
                <span class="material-symbols-outlined">folder</span>
                <span class="font-medium"
                    >{filteredDocuments.length} dokumen ditemukan</span
                >
            </div>

            <div class="flex items-center gap-4">
                <!-- Sort -->
                <select
                    bind:value={selectedSort}
                    class="px-4 py-2 rounded-lg bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300 focus:outline-none focus:ring-2 focus:ring-primary/50"
                >
                    <option value="terbaru">Terbaru</option>
                    <option value="terlama">Terlama</option>
                    <option value="judul">Judul A-Z</option>
                </select>

                <!-- View Mode -->
                <div
                    class="flex bg-white dark:bg-slate-800 rounded-lg border border-slate-200 dark:border-slate-700 overflow-hidden"
                >
                    <button
                        on:click={() => (viewMode = "grid")}
                        class="p-2 {viewMode === 'grid'
                            ? 'bg-primary text-white'
                            : 'text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-700'}"
                        aria-label="Grid view"
                    >
                        <span class="material-symbols-outlined">grid_view</span>
                    </button>
                    <button
                        on:click={() => (viewMode = "list")}
                        class="p-2 {viewMode === 'list'
                            ? 'bg-primary text-white'
                            : 'text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-700'}"
                        aria-label="List view"
                    >
                        <span class="material-symbols-outlined">view_list</span>
                    </button>
                </div>
            </div>
        </div>

        <!-- Results -->
        {#if loading}
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {#each Array(6) as _}
                    <div
                        class="bg-white dark:bg-slate-800 rounded-xl p-6 border border-slate-200 dark:border-slate-700"
                    >
                        <div
                            class="w-12 h-12 bg-slate-200 dark:bg-slate-700 rounded-lg animate-pulse mb-4"
                        ></div>
                        <div
                            class="h-6 bg-slate-200 dark:bg-slate-700 rounded animate-pulse mb-2"
                        ></div>
                        <div
                            class="h-4 bg-slate-100 dark:bg-slate-600 rounded animate-pulse mb-4"
                        ></div>
                        <div
                            class="h-4 w-1/2 bg-slate-100 dark:bg-slate-600 rounded animate-pulse"
                        ></div>
                    </div>
                {/each}
            </div>
        {:else if filteredDocuments.length === 0}
            <div class="text-center py-16">
                <div
                    class="w-24 h-24 mx-auto mb-6 rounded-full bg-slate-100 dark:bg-slate-800 flex items-center justify-center"
                >
                    <span
                        class="material-symbols-outlined text-5xl text-slate-300 dark:text-slate-600"
                        >search_off</span
                    >
                </div>
                <h3
                    class="text-xl font-bold text-slate-700 dark:text-slate-300 mb-2"
                >
                    Tidak ada hasil
                </h3>
                <p class="text-slate-500">
                    Coba ubah kata kunci atau filter pencarian Anda.
                </p>
            </div>
        {:else if viewMode === "grid"}
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {#each filteredDocuments as doc}
                    <article
                        class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden hover:shadow-xl hover:border-primary transition-all group"
                    >
                        <!-- Header -->
                        <div class="p-6">
                            <div class="flex items-start justify-between mb-4">
                                <div
                                    class="p-3 rounded-xl {getTypeStyle(
                                        doc.jenis_file,
                                    ).color}"
                                >
                                    <span class="material-symbols-outlined"
                                        >{getTypeStyle(doc.jenis_file)
                                            .icon}</span
                                    >
                                </div>
                                <span
                                    class="px-3 py-1 text-xs font-bold uppercase rounded-full bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-400"
                                >
                                    {doc.jenis_file}
                                </span>
                            </div>

                            <a href="#/document/{doc.id}">
                                <h3
                                    class="text-lg font-bold text-slate-900 dark:text-white mb-2 line-clamp-2 group-hover:text-primary transition-colors"
                                >
                                    {doc.judul}
                                </h3>
                            </a>

                            <div
                                class="flex items-center gap-2 text-sm text-slate-500 dark:text-slate-400 mb-4"
                            >
                                <div
                                    class="w-6 h-6 rounded-full bg-gradient-to-br from-primary to-blue-600 flex items-center justify-center text-white text-xs font-bold"
                                >
                                    {getInitials(doc.penulis)}
                                </div>
                                <span class="truncate">{doc.penulis}</span>
                            </div>

                            <p
                                class="text-sm text-slate-500 dark:text-slate-400"
                            >
                                <span
                                    class="material-symbols-outlined text-sm align-middle mr-1"
                                    >calendar_today</span
                                >
                                {formatDate(doc.created_at)}
                            </p>
                        </div>

                        <!-- Footer -->
                        <div
                            class="px-6 py-4 bg-slate-50 dark:bg-slate-900/50 border-t border-slate-100 dark:border-slate-700 flex items-center justify-between"
                        >
                            <a
                                href="#/document/{doc.id}"
                                class="text-primary font-medium text-sm hover:underline"
                            >
                                Lihat Detail
                            </a>
                        </div>
                    </article>
                {/each}
            </div>
        {:else}
            <!-- List View -->
            <div class="flex flex-col gap-4">
                {#each filteredDocuments as doc}
                    <article
                        class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-6 hover:shadow-lg hover:border-primary transition-all group flex gap-6"
                    >
                        <div
                            class="p-4 rounded-xl {getTypeStyle(doc.jenis_file)
                                .color} h-fit"
                        >
                            <span class="material-symbols-outlined text-2xl"
                                >{getTypeStyle(doc.jenis_file).icon}</span
                            >
                        </div>

                        <div class="flex-1 min-w-0">
                            <div
                                class="flex items-start justify-between gap-4 mb-2"
                            >
                                <a href="#/document/{doc.id}">
                                    <h3
                                        class="text-lg font-bold text-slate-900 dark:text-white group-hover:text-primary transition-colors"
                                    >
                                        {doc.judul}
                                    </h3>
                                </a>
                                <span
                                    class="px-3 py-1 text-xs font-bold uppercase rounded-full bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-400 flex-shrink-0"
                                >
                                    {doc.jenis_file}
                                </span>
                            </div>

                            <div
                                class="flex flex-wrap items-center gap-4 text-sm text-slate-500 dark:text-slate-400"
                            >
                                <div class="flex items-center gap-2">
                                    <div
                                        class="w-6 h-6 rounded-full bg-gradient-to-br from-primary to-blue-600 flex items-center justify-center text-white text-xs font-bold"
                                    >
                                        {getInitials(doc.penulis)}
                                    </div>
                                    <span>{doc.penulis}</span>
                                </div>
                                <span>•</span>
                                <span>{formatDate(doc.created_at)}</span>
                            </div>
                        </div>

                        <div class="flex items-center gap-2 flex-shrink-0">
                            <a
                                href="#/document/{doc.id}"
                                class="px-4 py-2 bg-primary/10 text-primary font-medium rounded-lg hover:bg-primary/20 transition-colors"
                            >
                                Detail
                            </a>
                        </div>
                    </article>
                {/each}
            </div>
        {/if}
    </div>
</div>

<style>
    .line-clamp-2 {
        display: -webkit-box;
        -webkit-line-clamp: 2;
        line-clamp: 2;
        -webkit-box-orient: vertical;
        overflow: hidden;
    }
</style>
