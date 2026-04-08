<script>
    import { onMount } from "svelte";
    import { API_ENDPOINTS } from "../../config";
    import {
        appName,
        appDescription,
        aboutText,
        initSiteSettings,
    } from "../../stores/index.js";

    let searchQuery = "";
    let searchFilter = "Semua";
    let loading = true;

    // Real stats from API
    let totalDocs = 0;
    let totalAuthors = 0;
    let totalVisitors = 0;
    let todayVisitors = 0;

    onMount(() => {
        initSiteSettings();
        loadStats();
    });

    async function loadStats() {
        try {
            loading = true;
            const response = await fetch(API_ENDPOINTS.STATS);
            if (response.ok) {
                const data = await response.json();
                totalDocs = data.total_documents || 0;
                totalAuthors = data.total_authors || 0;
                totalVisitors = data.total_visitors || 0;
                todayVisitors = data.today_visitors || 0;
            }
        } catch (e) {
            console.error("Failed to load stats:", e);
        } finally {
            loading = false;
        }
    }

    $: stats = [
        {
            value: loading ? "..." : `${totalDocs}`,
            label: "Total Dokumen",
            icon: "description",
        },
        {
            value: loading ? "..." : `${totalAuthors}`,
            label: "Penulis",
            icon: "group",
        },
        {
            value: loading ? "..." : `${totalVisitors}`,
            label: "Total Pengunjung",
            icon: "visibility",
        },
        {
            value: loading ? "..." : `${todayVisitors}`,
            label: "Pengunjung Hari Ini",
            icon: "today",
        },
    ];

    function handleSearch() {
        if (searchQuery.trim()) {
            window.location.hash = `/browse?q=${encodeURIComponent(searchQuery)}&filter=${searchFilter}`;
        }
    }
</script>

<section class="relative px-4 py-12 lg:py-24 overflow-hidden">
    <!-- Background Gradient Effect -->
    <div
        class="absolute inset-0 bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-blue-900/20 via-background-dark to-background-dark -z-10"
    ></div>
    <div
        class="absolute top-0 right-0 w-1/2 h-full bg-gradient-to-l from-primary/5 to-transparent -z-10"
    ></div>

    <div
        class="container mx-auto max-w-6xl flex flex-col items-center gap-8 text-center"
    >
        <div class="flex flex-col gap-4 max-w-3xl">
            <!-- Badge -->
            <div
                class="inline-flex items-center justify-center gap-2 px-3 py-1 rounded-full bg-primary/10 border border-primary/20 text-primary text-xs font-bold uppercase tracking-wide mb-2 w-fit mx-auto"
            >
                <span class="w-2 h-2 rounded-full bg-primary animate-pulse"
                ></span>
                {$appDescription}
            </div>

            <!-- Title -->
            <h1
                class="text-slate-900 dark:text-white text-4xl md:text-5xl lg:text-6xl font-black leading-tight tracking-[-0.033em]"
            >
                Temukan <span
                    class="text-transparent bg-clip-text bg-gradient-to-r from-primary to-blue-400"
                    >Karya Ilmiah</span
                >
                di {$appName}
            </h1>

            <!-- Subtitle -->
            <h2
                class="text-slate-600 dark:text-slate-400 text-lg md:text-xl font-normal leading-relaxed max-w-2xl mx-auto"
            >
                {#if $aboutText}
                    {$aboutText}
                {:else}
                    Akses ribuan skripsi, tesis, jurnal, dan karya ilmiah dari
                    civitas akademika {$appName}.
                {/if}
            </h2>
        </div>

        <!-- Search Box -->
        <div class="w-full max-w-2xl mt-4 relative z-10">
            <form
                on:submit|preventDefault={handleSearch}
                class="flex flex-col md:flex-row w-full items-stretch rounded-xl shadow-2xl shadow-blue-900/20 bg-white dark:bg-surface-dark border border-slate-200 dark:border-slate-700 overflow-hidden p-2 gap-2"
            >
                <!-- Dropdown Filter -->
                <div class="relative md:w-40 flex-shrink-0">
                    <select
                        bind:value={searchFilter}
                        class="w-full h-12 md:h-full appearance-none bg-slate-50 dark:bg-[#101922] border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300 rounded-lg px-4 pr-10 focus:outline-none focus:ring-2 focus:ring-primary/50 text-sm font-medium cursor-pointer"
                    >
                        <option>Semua</option>
                        <option>Skripsi</option>
                        <option>Tesis</option>
                        <option>Jurnal</option>
                    </select>
                    <div
                        class="pointer-events-none absolute inset-y-0 right-0 flex items-center px-3 text-slate-500"
                    >
                        <span class="material-symbols-outlined text-sm"
                            >expand_more</span
                        >
                    </div>
                </div>

                <!-- Input -->
                <div class="relative flex-1">
                    <div
                        class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none"
                    >
                        <span class="material-symbols-outlined text-slate-400"
                            >search</span
                        >
                    </div>
                    <input
                        bind:value={searchQuery}
                        class="w-full h-12 md:h-full bg-transparent border-none text-slate-900 dark:text-white placeholder:text-slate-400 focus:ring-0 pl-10 pr-4 text-base"
                        placeholder="Cari judul, penulis, atau kata kunci..."
                        type="text"
                    />
                </div>

                <!-- Button -->
                <button
                    type="submit"
                    class="h-12 md:h-auto px-8 bg-primary hover:bg-primary/90 text-white font-bold rounded-lg transition-all flex items-center justify-center gap-2 group"
                >
                    <span>Cari</span>
                    <span
                        class="material-symbols-outlined text-lg group-hover:translate-x-1 transition-transform"
                        >arrow_forward</span
                    >
                </button>
            </form>

            <!-- Quick Links -->
            <div
                class="mt-4 flex flex-wrap justify-center gap-2 text-xs text-slate-500 dark:text-slate-400"
            >
                <span>Populer:</span>
                <a
                    class="hover:text-primary underline decoration-primary/30"
                    href="#/browse?q=skripsi">Skripsi</a
                >
                <span class="w-1 h-1 rounded-full bg-slate-600 mt-1.5"></span>
                <a
                    class="hover:text-primary underline decoration-primary/30"
                    href="#/browse?q=tesis">Tesis</a
                >
                <span class="w-1 h-1 rounded-full bg-slate-600 mt-1.5"></span>
                <a
                    class="hover:text-primary underline decoration-primary/30"
                    href="#/browse?q=jurnal">Jurnal</a
                >
            </div>
        </div>

        <!-- Stats Grid -->
        <div class="grid grid-cols-2 md:grid-cols-4 gap-4 w-full max-w-3xl mt-8">
            {#each stats as stat}
                <div
                    class="flex flex-col items-center justify-center p-4 rounded-xl bg-white dark:bg-surface-highlight border border-slate-200 dark:border-slate-700/50 shadow-sm hover:shadow-md transition-shadow"
                >
                    <span class="material-symbols-outlined text-primary/60 text-xl mb-1">{stat.icon}</span>
                    <p class="text-primary text-2xl md:text-3xl font-bold">
                        {stat.value}
                    </p>
                    <p
                        class="text-slate-500 dark:text-slate-400 text-xs md:text-sm font-medium text-center"
                    >
                        {stat.label}
                    </p>
                </div>
            {/each}
        </div>
    </div>
</section>
