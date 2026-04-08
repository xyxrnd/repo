<script>
    import { createEventDispatcher } from "svelte";
    const dispatch = createEventDispatcher();

    function toggleSidebar() {
        dispatch("toggleSidebar");
    }
</script>

<header
    class="h-14 md:h-16 bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between px-4 md:px-8 shrink-0"
>
    <div class="flex items-center gap-2">
        <!-- Mobile hamburger -->
        <button
            on:click={toggleSidebar}
            class="lg:hidden p-2 -ml-2 text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-lg transition-colors"
            aria-label="Toggle menu"
        >
            <span class="material-symbols-outlined">menu</span>
        </button>

        <!-- <span class="text-slate-400 material-symbols-outlined hidden md:block"
            >search</span
        >
        <input
            class="bg-transparent border-none focus:ring-0 text-sm text-slate-600 placeholder:text-slate-400 w-40 md:w-64 hidden md:block"
            placeholder="Cepat temukan fitur..."
            type="text"
        /> -->
    </div>

    <!-- <div class="flex items-center gap-2 md:gap-4">
        <button
            class="p-2 text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-lg relative"
        >
            <span class="material-symbols-outlined">notifications</span>
            <span
                class="absolute top-2 right-2 size-2 bg-red-500 rounded-full border-2 border-white dark:border-slate-900"
            ></span>
        </button>

        <div
            class="h-6 w-px bg-slate-200 dark:border-slate-800 hidden md:block"
        ></div>

        <button
            class="hidden md:flex items-center gap-2 text-sm font-medium text-slate-600 dark:text-slate-300"
        >
            <span>Bahasa Indonesia</span>
            <span class="material-symbols-outlined text-lg">expand_more</span>
        </button>
    </div> -->
</header>
