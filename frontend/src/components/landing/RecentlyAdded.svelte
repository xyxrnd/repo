<script>
    import { onMount } from "svelte";
    import { getDocuments } from "../../services/documentService";

    let documents = [];
    let loading = true;

    onMount(() => {
        loadDocuments();
    });

    async function loadDocuments() {
        try {
            loading = true;
            const allDocs = await getDocuments();
            // Get the 6 most recent documents
            documents = allDocs.slice(0, 6);
        } catch (e) {
            console.error("Failed to load documents:", e);
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

    function getTypeStyle(jenisFile) {
        const styles = {
            PDF: {
                icon: "picture_as_pdf",
                color: "bg-red-100 dark:bg-red-900/30 text-red-600 dark:text-red-400",
            },
            Skripsi: {
                icon: "school",
                color: "bg-blue-100 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400",
            },
            Tesis: {
                icon: "menu_book",
                color: "bg-purple-100 dark:bg-purple-900/30 text-purple-600 dark:text-purple-400",
            },
            Jurnal: {
                icon: "article",
                color: "bg-teal-100 dark:bg-teal-900/30 text-teal-600 dark:text-teal-400",
            },
            Disertasi: {
                icon: "history_edu",
                color: "bg-amber-100 dark:bg-amber-900/30 text-amber-600 dark:text-amber-400",
            },
        };
        return (
            styles[jenisFile] || {
                icon: "description",
                color: "bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-400",
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

<section
    class="py-12 bg-slate-50 dark:bg-background-dark border-y border-slate-200 dark:border-slate-800"
>
    <div class="container mx-auto max-w-6xl px-4">
        <div class="flex items-center justify-between mb-8">
            <div>
                <h2
                    class="text-2xl md:text-3xl font-bold text-slate-900 dark:text-white mb-2"
                >
                    Baru Ditambahkan
                </h2>
                <p class="text-slate-500 dark:text-slate-400">
                    Dokumen terbaru yang diupload ke repositori.
                </p>
            </div>
            <a
                href="#/documents"
                class="hidden md:flex items-center gap-1 text-primary font-medium hover:underline"
            >
                Lihat Semua <span class="material-symbols-outlined text-sm"
                    >arrow_forward_ios</span
                >
            </a>
        </div>

        {#if loading}
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {#each Array(3) as _}
                    <div
                        class="bg-white dark:bg-surface-highlight rounded-xl p-5 border border-slate-200 dark:border-slate-700"
                    >
                        <div class="flex items-start justify-between mb-3">
                            <div
                                class="w-10 h-10 bg-slate-200 dark:bg-slate-700 rounded-lg animate-pulse"
                            ></div>
                            <div
                                class="w-16 h-6 bg-slate-200 dark:bg-slate-700 rounded animate-pulse"
                            ></div>
                        </div>
                        <div
                            class="h-6 w-3/4 bg-slate-200 dark:bg-slate-700 rounded animate-pulse mb-2"
                        ></div>
                        <div
                            class="h-4 w-full bg-slate-100 dark:bg-slate-800 rounded animate-pulse mb-4"
                        ></div>
                        <div
                            class="flex items-center gap-3 pt-4 border-t border-slate-100 dark:border-slate-700"
                        >
                            <div
                                class="w-8 h-8 bg-slate-200 dark:bg-slate-700 rounded-full animate-pulse"
                            ></div>
                            <div class="flex-1">
                                <div
                                    class="h-4 w-24 bg-slate-200 dark:bg-slate-700 rounded animate-pulse mb-1"
                                ></div>
                                <div
                                    class="h-3 w-16 bg-slate-100 dark:bg-slate-800 rounded animate-pulse"
                                ></div>
                            </div>
                        </div>
                    </div>
                {/each}
            </div>
        {:else if documents.length === 0}
            <div class="text-center py-12">
                <span
                    class="material-symbols-outlined text-5xl text-slate-300 dark:text-slate-700 mb-4"
                    >folder_off</span
                >
                <p class="text-slate-500">Belum ada dokumen</p>
            </div>
        {:else}
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {#each documents as doc}
                    <article
                        class="bg-white dark:bg-surface-highlight rounded-xl p-5 border border-slate-200 dark:border-slate-700 hover:border-primary dark:hover:border-primary transition-all hover:shadow-lg group"
                    >
                        <div class="flex items-start justify-between mb-3">
                            <div
                                class="p-2 {getTypeStyle(doc.jenis_file)
                                    .color} rounded-lg"
                            >
                                <span class="material-symbols-outlined"
                                    >{getTypeStyle(doc.jenis_file).icon}</span
                                >
                            </div>
                            <span
                                class="px-2 py-1 bg-slate-100 dark:bg-slate-700 text-slate-500 dark:text-slate-300 text-xs font-bold uppercase rounded"
                            >
                                {doc.jenis_file}
                            </span>
                        </div>

                        <a href="#/document/{doc.id}" class="block">
                            <h3
                                class="text-lg font-bold text-slate-900 dark:text-white mb-2 line-clamp-2 group-hover:text-primary transition-colors"
                            >
                                {doc.judul}
                            </h3>
                        </a>

                        <p
                            class="text-sm text-slate-500 dark:text-slate-400 mb-4 line-clamp-2"
                        >
                            {#if doc.abstrak}
                                {doc.abstrak}
                            {:else}
                                {doc.status === "publish"
                                    ? "Dipublikasikan"
                                    : "Draft"} • {doc.jenis_file
                                    ? doc.jenis_file.charAt(0).toUpperCase() +
                                      doc.jenis_file.slice(1)
                                    : ""}
                            {/if}
                        </p>

                        <div
                            class="flex items-center gap-3 pt-4 border-t border-slate-100 dark:border-slate-700"
                        >
                            <div
                                class="w-8 h-8 rounded-full bg-gradient-to-br from-primary to-blue-600 flex items-center justify-center text-white text-xs font-bold"
                            >
                                {getInitials(doc.penulis)}
                            </div>
                            <div class="flex-1 min-w-0">
                                <p
                                    class="text-sm font-medium text-slate-900 dark:text-white truncate"
                                >
                                    {doc.penulis}
                                </p>
                                <p
                                    class="text-xs text-slate-500 dark:text-slate-500 truncate"
                                >
                                    {formatDate(doc.created_at)}
                                </p>
                            </div>
                        </div>
                    </article>
                {/each}
            </div>
        {/if}

        <div class="md:hidden mt-6 text-center">
            <a
                class="inline-flex items-center gap-1 text-primary font-medium hover:underline"
                href="#/documents"
            >
                Lihat Semua <span class="material-symbols-outlined text-sm"
                    >arrow_forward_ios</span
                >
            </a>
        </div>
    </div>
</section>

<style>
    .bg-surface-highlight {
        background-color: #233648;
    }
    .line-clamp-2 {
        display: -webkit-box;
        -webkit-line-clamp: 2;
        line-clamp: 2;
        -webkit-box-orient: vertical;
        overflow: hidden;
    }
</style>
