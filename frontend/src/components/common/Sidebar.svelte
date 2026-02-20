<script>
    import { link } from "svelte-spa-router";
    import { onMount, createEventDispatcher } from "svelte";
    import authService from "../../services/authService";

    const dispatch = createEventDispatcher();
    let user = null;

    onMount(() => {
        user = authService.getUser();
    });

    function handleLogout() {
        authService.logout();
    }

    function handleNavClick() {
        dispatch("navigate");
    }
</script>

<aside
    class="w-72 bg-white dark:bg-slate-900 border-r border-slate-200 dark:border-slate-800 flex flex-col h-full shrink-0"
>
    <div class="p-6 flex flex-col gap-6">
        <!-- Brand -->
        <div class="flex items-center gap-3">
            <div
                class="bg-primary size-10 rounded-lg flex items-center justify-center text-white shadow-lg shadow-primary/20"
            >
                <span class="material-symbols-outlined">folder_managed</span>
            </div>
            <div class="flex flex-col">
                <h1
                    class="text-slate-900 dark:text-white text-base font-bold leading-tight"
                >
                    Admin Panel
                </h1>
                <p
                    class="text-slate-500 dark:text-slate-400 text-xs font-normal"
                >
                    Repository Akademik
                </p>
            </div>
        </div>

        <!-- Navigation -->
        <nav class="flex flex-col gap-1">
            <p
                class="text-[10px] font-bold text-slate-400 dark:text-slate-500 uppercase tracking-wider px-3 mb-1"
            >
                Utama
            </p>
            <a
                href="#/dashboard"
                use:link
                on:click={handleNavClick}
                class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            >
                <span class="material-symbols-outlined text-xl">home</span>
                <span class="text-sm font-medium">Beranda</span>
            </a>
            <a
                href="#/documents"
                use:link
                on:click={handleNavClick}
                class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            >
                <span class="material-symbols-outlined text-xl"
                    >description</span
                >
                <span class="text-sm font-medium">Kelola Dokumen</span>
            </a>

            {#if user?.role === "admin"}
                <hr class="border-slate-200 dark:border-slate-800 my-2" />
                <p
                    class="text-[10px] font-bold text-slate-400 dark:text-slate-500 uppercase tracking-wider px-3 mb-1"
                >
                    Master Data
                </p>
                <a
                    href="#/fakultas"
                    use:link
                    on:click={handleNavClick}
                    class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
                >
                    <span class="material-symbols-outlined text-xl"
                        >account_balance</span
                    >
                    <span class="text-sm font-medium">Fakultas</span>
                </a>
                <a
                    href="#/prodi"
                    use:link
                    on:click={handleNavClick}
                    class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
                >
                    <span class="material-symbols-outlined text-xl"
                        >local_library</span
                    >
                    <span class="text-sm font-medium">Program Studi</span>
                </a>

                <hr class="border-slate-200 dark:border-slate-800 my-2" />
                <p
                    class="text-[10px] font-bold text-slate-400 dark:text-slate-500 uppercase tracking-wider px-3 mb-1"
                >
                    Administrasi
                </p>
                <a
                    href="#/users"
                    use:link
                    on:click={handleNavClick}
                    class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
                >
                    <span class="material-symbols-outlined text-xl">group</span>
                    <span class="text-sm font-medium">Manajemen User</span>
                </a>
                <a
                    href="#/access-requests"
                    use:link
                    on:click={handleNavClick}
                    class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
                >
                    <span class="material-symbols-outlined text-xl">key</span>
                    <span class="text-sm font-medium">Permintaan Akses</span>
                </a>
            {/if}
            <a
                href="#/reports"
                use:link
                on:click={handleNavClick}
                class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            >
                <span class="material-symbols-outlined text-xl">assessment</span
                >
                <span class="text-sm font-medium">Laporan</span>
            </a>
        </nav>

        <hr class="border-slate-200 dark:border-slate-800" />

        <nav class="flex flex-col gap-1">
            <a
                href="#/settings"
                use:link
                on:click={handleNavClick}
                class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            >
                <span class="material-symbols-outlined text-xl">settings</span>
                <span class="text-sm font-medium">Pengaturan</span>
            </a>
            {#if user?.role === "admin"}
                <a
                    href="#/system-settings"
                    use:link
                    on:click={handleNavClick}
                    class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
                >
                    <span class="material-symbols-outlined text-xl">tune</span>
                    <span class="text-sm font-medium">Pengaturan Sistem</span>
                </a>
            {/if}
        </nav>
    </div>

    <div class="mt-auto p-6 border-t border-slate-200 dark:border-slate-800">
        {#if user}
            <div class="flex items-center gap-3 px-2">
                <div
                    class="size-9 rounded-full bg-gradient-to-br from-primary to-blue-600 overflow-hidden shrink-0 flex items-center justify-center text-white font-bold text-sm"
                >
                    {user.name?.charAt(0).toUpperCase() || "U"}
                </div>
                <div class="flex flex-col min-w-0">
                    <span
                        class="text-sm font-bold truncate text-slate-900 dark:text-white"
                        >{user.name || "User"}</span
                    >
                    <span class="text-xs text-slate-500 truncate"
                        >{user.email || ""}</span
                    >
                </div>
                <button
                    on:click={handleLogout}
                    class="ml-auto text-slate-400 hover:text-red-500 transition-colors"
                    title="Logout"
                >
                    <span class="material-symbols-outlined">logout</span>
                </button>
            </div>
        {:else}
            <a
                href="#/login"
                class="flex items-center justify-center gap-2 px-4 py-2.5 bg-primary hover:bg-primary/90 text-white font-medium rounded-lg transition-colors"
            >
                <span class="material-symbols-outlined text-xl">login</span>
                <span>Login</span>
            </a>
        {/if}
    </div>
</aside>
