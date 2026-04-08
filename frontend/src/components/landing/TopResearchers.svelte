<script>
    import { onMount } from "svelte";
    import { API_ENDPOINTS } from "../../config";

    let fakultasList = [];
    let loading = true;
    let maxDocs = 0;

    onMount(() => {
        loadTopFakultas();
    });

    async function loadTopFakultas() {
        try {
            loading = true;
            const response = await fetch(API_ENDPOINTS.TOP_FAKULTAS);
            if (response.ok) {
                fakultasList = await response.json();
                if (fakultasList.length > 0) {
                    maxDocs = fakultasList[0].total_documents;
                }
            }
        } catch (e) {
            console.error("Failed to load top fakultas:", e);
        } finally {
            loading = false;
        }
    }

    function getIcon(index) {
        const icons = [
            "school",
            "biotech",
            "engineering",
            "balance",
            "psychology",
            "computer",
        ];
        return icons[index % icons.length];
    }

    function getGradient(index) {
        const gradients = [
            "from-blue-500 to-indigo-600",
            "from-emerald-500 to-teal-600",
            "from-violet-500 to-purple-600",
            "from-amber-500 to-orange-600",
            "from-rose-500 to-pink-600",
            "from-cyan-500 to-sky-600",
        ];
        return gradients[index % gradients.length];
    }

    function getBarColor(index) {
        const colors = [
            "bg-blue-500",
            "bg-emerald-500",
            "bg-violet-500",
            "bg-amber-500",
            "bg-rose-500",
            "bg-cyan-500",
        ];
        return colors[index % colors.length];
    }
</script>

<section class="py-12 lg:py-16 bg-white dark:bg-[#131d27]">
    <div class="container mx-auto max-w-6xl px-4">
        <div class="text-center mb-10">
            <h2 class="text-3xl font-bold text-slate-900 dark:text-white mb-3">
                Fakultas Terpopuler
            </h2>
            <p class="text-slate-500 dark:text-slate-400 max-w-2xl mx-auto">
                Fakultas dengan kontribusi dokumen terbanyak dalam repositori.
            </p>
        </div>

        {#if loading}
            <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5">
                {#each Array(6) as _}
                    <div
                        class="bg-slate-50 dark:bg-[#1a2836] p-5 rounded-xl border border-slate-100 dark:border-slate-700/50 flex items-center gap-4"
                    >
                        <div
                            class="w-12 h-12 rounded-xl bg-slate-200 dark:bg-slate-700 animate-pulse shrink-0"
                        ></div>
                        <div class="flex-1">
                            <div
                                class="h-4 w-32 bg-slate-200 dark:bg-slate-700 rounded animate-pulse mb-2"
                            ></div>
                            <div
                                class="h-3 w-full bg-slate-100 dark:bg-slate-800 rounded-full animate-pulse"
                            ></div>
                        </div>
                    </div>
                {/each}
            </div>
        {:else if fakultasList.length === 0}
            <div class="text-center py-12">
                <span
                    class="material-symbols-outlined text-5xl text-slate-300 dark:text-slate-700 mb-4"
                    >apartment</span
                >
                <p class="text-slate-500">Belum ada data fakultas</p>
            </div>
        {:else}
            <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5">
                {#each fakultasList as fak, index}
                    <a
                        href="#/browse?fakultas={fak.id}"
                        class="group bg-slate-50 dark:bg-[#1a2836] p-5 rounded-xl border border-slate-100 dark:border-slate-700/50 hover:border-primary/30 dark:hover:border-primary/30 hover:shadow-lg hover:shadow-primary/5 transition-all flex items-center gap-4"
                    >
                        <!-- Icon -->
                        <div
                            class="w-12 h-12 rounded-xl bg-gradient-to-br {getGradient(index)} flex items-center justify-center text-white shrink-0 shadow-lg group-hover:scale-105 transition-transform"
                        >
                            <span
                                class="material-symbols-outlined"
                                style="font-size: 24px;">{getIcon(index)}</span
                            >
                        </div>

                        <!-- Info -->
                        <div class="flex-1 min-w-0">
                            <div class="flex items-center justify-between mb-1.5">
                                <h3
                                    class="font-bold text-slate-900 dark:text-white text-sm truncate pr-2 group-hover:text-primary transition-colors"
                                >
                                    {fak.nama}
                                </h3>
                                <span
                                    class="text-xs font-bold text-primary shrink-0"
                                >
                                    {fak.total_documents}
                                </span>
                            </div>

                            <!-- Progress bar -->
                            <div
                                class="w-full h-2 bg-slate-200 dark:bg-slate-700 rounded-full overflow-hidden"
                            >
                                <div
                                    class="h-full rounded-full {getBarColor(index)} transition-all duration-700"
                                    style="width: {maxDocs > 0
                                        ? (fak.total_documents / maxDocs) * 100
                                        : 0}%"
                                ></div>
                            </div>

                            <p
                                class="text-[11px] text-slate-400 dark:text-slate-500 mt-1"
                            >
                                {fak.total_documents} dokumen
                            </p>
                        </div>
                    </a>
                {/each}
            </div>
        {/if}
    </div>
</section>
