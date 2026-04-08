<script>
    import { onMount } from "svelte";
    import { push } from "svelte-spa-router";
    import authService from "../services/authService";
    import {
        appName,
        appDescription,
        aboutText,
        footerText,
        logoFullUrl,
        contactInfo,
        initSiteSettings,
    } from "../stores/index.js";

    let user = null;
    let mobileMenuOpen = false;

    onMount(() => {
        user = authService.getUser();
        initSiteSettings();
    });

    function handleLogout() {
        authService.clearAuth();
        user = null;
        push("/");
    }

    function toggleMobileMenu() {
        mobileMenuOpen = !mobileMenuOpen;
    }

    function closeMobileMenu() {
        mobileMenuOpen = false;
    }
</script>

<div class="relative flex min-h-screen w-full flex-col">
    <!-- Navbar -->
    <header
        class="sticky top-0 z-50 flex items-center justify-between whitespace-nowrap border-b border-solid border-slate-200 dark:border-[#233648] bg-white/80 dark:bg-[#101922]/90 backdrop-blur-md px-4 py-3 md:px-6 md:py-4 lg:px-10"
    >
        <a
            href="#/"
            class="flex items-center gap-3 md:gap-4 text-slate-900 dark:text-white cursor-pointer hover:opacity-90 transition-opacity"
        >
            {#if $logoFullUrl}
                <img
                    src={$logoFullUrl}
                    alt="Logo"
                    class="w-8 h-8 object-contain"
                />
            {:else}
                <div class="size-8 text-primary">
                    <span class="material-symbols-outlined text-4xl"
                        >local_library</span
                    >
                </div>
            {/if}
            <h2
                class="text-lg md:text-xl font-bold leading-tight tracking-[-0.015em]"
            >
                {$appName}
            </h2>
        </a>

        <div class="hidden lg:flex flex-1 justify-end gap-8">
            <nav class="flex items-center gap-9">
                <a
                    class="text-slate-600 dark:text-slate-300 hover:text-primary dark:hover:text-primary text-sm font-medium leading-normal transition-colors"
                    href="#/">Home</a
                >
                <a
                    class="text-slate-600 dark:text-slate-300 hover:text-primary dark:hover:text-primary text-sm font-medium leading-normal transition-colors"
                    href="#/browse">Browse</a
                >
                <a
                    class="text-slate-600 dark:text-slate-300 hover:text-primary dark:hover:text-primary text-sm font-medium leading-normal transition-colors"
                    href="#/about">About</a
                >
            </nav>
            <div class="flex gap-3 items-center">
                {#if user}
                    <a
                        href="#/admin"
                        class="flex items-center gap-2 text-slate-600 dark:text-slate-300 hover:text-primary text-sm font-medium transition-colors"
                    >
                        <span class="material-symbols-outlined text-xl"
                            >dashboard</span
                        >
                        Dashboard
                    </a>
                    <div
                        class="flex items-center gap-3 pl-4 border-l border-slate-200 dark:border-slate-700"
                    >
                        <div
                            class="w-8 h-8 rounded-full bg-gradient-to-br from-primary to-blue-600 flex items-center justify-center text-white font-bold text-sm"
                        >
                            {user.name?.charAt(0).toUpperCase() || "U"}
                        </div>
                        <span
                            class="text-sm font-medium text-slate-700 dark:text-slate-300"
                            >{user.name}</span
                        >
                        <button
                            on:click={handleLogout}
                            class="p-2 text-slate-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors"
                            title="Logout"
                        >
                            <span class="material-symbols-outlined text-xl"
                                >logout</span
                            >
                        </button>
                    </div>
                {:else}
                    <a
                        href="#/login"
                        class="flex min-w-[84px] cursor-pointer items-center justify-center overflow-hidden rounded-lg h-10 px-6 bg-slate-200 dark:bg-[#233648] hover:bg-slate-300 dark:hover:bg-[#2f455a] text-slate-900 dark:text-white text-sm font-bold leading-normal tracking-[0.015em] transition-all"
                    >
                        <span class="truncate">Login</span>
                    </a>
                    <a
                        href="#/register"
                        class="flex min-w-[84px] cursor-pointer items-center justify-center overflow-hidden rounded-lg h-10 px-6 bg-primary hover:bg-primary/90 text-white text-sm font-bold leading-normal tracking-[0.015em] transition-all shadow-lg shadow-primary/25"
                    >
                        <span class="truncate">Sign Up</span>
                    </a>
                {/if}
            </div>
        </div>

        <!-- Mobile Menu Button -->
        <button
            class="lg:hidden p-2 text-slate-900 dark:text-white hover:bg-slate-100 dark:hover:bg-slate-800 rounded-lg transition-colors"
            on:click={toggleMobileMenu}
            aria-label="Toggle menu"
        >
            <span class="material-symbols-outlined"
                >{mobileMenuOpen ? "close" : "menu"}</span
            >
        </button>
    </header>

    <!-- Mobile Menu Panel -->
    {#if mobileMenuOpen}
        <div
            class="lg:hidden fixed inset-x-0 top-[57px] z-40 bg-white dark:bg-[#101922] border-b border-slate-200 dark:border-[#233648] shadow-xl animate-slideDown"
        >
            <nav class="flex flex-col p-4 gap-1">
                <a
                    class="flex items-center gap-3 px-4 py-3 rounded-lg text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 font-medium transition-colors"
                    href="#/"
                    on:click={closeMobileMenu}
                >
                    <span class="material-symbols-outlined text-xl">home</span>
                    Home
                </a>
                <a
                    class="flex items-center gap-3 px-4 py-3 rounded-lg text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 font-medium transition-colors"
                    href="#/browse"
                    on:click={closeMobileMenu}
                >
                    <span class="material-symbols-outlined text-xl"
                        >explore</span
                    >
                    Browse
                </a>
                <a
                    class="flex items-center gap-3 px-4 py-3 rounded-lg text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 font-medium transition-colors"
                    href="#/about"
                    on:click={closeMobileMenu}
                >
                    <span class="material-symbols-outlined text-xl">info</span>
                    About
                </a>

                <hr class="border-slate-200 dark:border-slate-700 my-2" />

                {#if user}
                    <!-- <a
                        class="flex items-center gap-3 px-4 py-3 rounded-lg text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 font-medium transition-colors"
                        href="#/admin"
                        on:click={closeMobileMenu}
                    >
                        <span class="material-symbols-outlined text-xl"
                            >dashboard</span
                        >
                        Dashboard
                    </a> -->
                    <div class="flex items-center gap-3 px-4 py-3">
                        <div
                            class="w-8 h-8 rounded-full bg-gradient-to-br from-primary to-blue-600 flex items-center justify-center text-white font-bold text-sm"
                        >
                            {user.name?.charAt(0).toUpperCase() || "U"}
                        </div>
                        <span
                            class="text-sm font-medium text-slate-700 dark:text-slate-300 flex-1"
                            >{user.name}</span
                        >
                        <button
                            on:click={() => {
                                handleLogout();
                                closeMobileMenu();
                            }}
                            class="p-2 text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors"
                        >
                            <span class="material-symbols-outlined text-xl"
                                >logout</span
                            >
                        </button>
                    </div>
                {:else}
                    <div class="flex gap-3 px-4 py-3">
                        <a
                            href="#/login"
                            on:click={closeMobileMenu}
                            class="flex-1 flex items-center justify-center py-2.5 bg-slate-200 dark:bg-[#233648] hover:bg-slate-300 dark:hover:bg-[#2f455a] text-slate-900 dark:text-white text-sm font-bold rounded-lg transition-all"
                        >
                            Login
                        </a>
                        <a
                            href="#/register"
                            on:click={closeMobileMenu}
                            class="flex-1 flex items-center justify-center py-2.5 bg-primary hover:bg-primary/90 text-white text-sm font-bold rounded-lg transition-all shadow-lg shadow-primary/25"
                        >
                            Sign Up
                        </a>
                    </div>
                {/if}
            </nav>
        </div>
    {/if}

    <main class="flex-1">
        <slot />
    </main>

    <!-- Footer -->
    <footer
        class="bg-white dark:bg-[#0b1219] border-t border-slate-200 dark:border-slate-800 pt-16 pb-8"
    >
        <div class="container mx-auto max-w-6xl px-4">
            <div class="grid grid-cols-1 md:grid-cols-4 gap-12 mb-12">
                <div class="col-span-1 md:col-span-1">
                    <div class="flex items-center gap-2 mb-4">
                        {#if $logoFullUrl}
                            <img
                                src={$logoFullUrl}
                                alt="Logo"
                                class="w-6 h-6 object-contain"
                            />
                        {:else}
                            <div class="size-6 text-primary">
                                <span class="material-symbols-outlined text-3xl"
                                    >local_library</span
                                >
                            </div>
                        {/if}
                        <h3
                            class="text-xl font-bold text-slate-900 dark:text-white"
                        >
                            {$appName}
                        </h3>
                    </div>
                    <p
                        class="text-slate-500 dark:text-slate-400 text-sm leading-relaxed mb-6"
                    >
                        {#if $aboutText}
                            {$aboutText}
                        {:else}
                            Demokratisasi akses ke pengetahuan ilmiah dunia.
                            Membangun infrastruktur untuk sains terbuka.
                        {/if}
                    </p>
                    <!-- <div class="flex gap-4">
                        <a
                            class="w-8 h-8 flex items-center justify-center rounded-full bg-slate-100 dark:bg-slate-800 text-slate-500 hover:bg-primary hover:text-white transition-colors"
                            href="#"
                        >
                            <span class="material-symbols-outlined text-sm"
                                >share</span
                            >
                        </a>
                        <a
                            class="w-8 h-8 flex items-center justify-center rounded-full bg-slate-100 dark:bg-slate-800 text-slate-500 hover:bg-primary hover:text-white transition-colors"
                            href="#"
                        >
                            <span class="material-symbols-outlined text-sm"
                                >mail</span
                            >
                        </a>
                    </div> -->
                </div>
                <div>
                    <h4 class="font-bold text-slate-900 dark:text-white mb-4">
                        Repositori
                    </h4>
                    <ul
                        class="space-y-3 text-sm text-slate-500 dark:text-slate-400"
                    >
                        <li>
                            <a
                                class="hover:text-primary transition-colors"
                                href="#/browse">Jelajahi Koleksi</a
                            >
                        </li>
                        <!-- <li>
                            <a
                                class="hover:text-primary transition-colors"
                                href="#">Penulis A-Z</a
                            >
                        </li> -->
                        <!-- <li>
                            <a
                                class="hover:text-primary transition-colors"
                                href="#">Pencarian Lanjutan</a
                            >
                        </li>
                        <li>
                            <a
                                class="hover:text-primary transition-colors"
                                href="#">Submit Dokumen</a
                            >
                        </li> -->
                    </ul>
                </div>
                <div>
                    <h4 class="font-bold text-slate-900 dark:text-white mb-4">
                        Sumber Daya
                    </h4>
                    <ul
                        class="space-y-3 text-sm text-slate-500 dark:text-slate-400"
                    >
                        <li>
                            <a
                                class="hover:text-primary transition-colors"
                                href="https://kbk2988048.perpustakaandigital.com/"
                                >Perpustakaan Digital</a
                            >
                        </li>
                        <li>
                            <a
                                class="hover:text-primary transition-colors"
                                href="https://lib.unsub.ac.id/">Library Unsub</a
                            >
                        </li>
                        <li>
                            <a
                                class="hover:text-primary transition-colors"
                                href="https://ejournal.unsub.ac.id/"
                                >E-Journal Unsub</a
                            >
                        </li>
                        <!-- <li>
                            <a
                                class="hover:text-primary transition-colors"
                                href="#">Untuk Pustakawan</a
                            >
                        </li>
                        <li>
                            <a
                                class="hover:text-primary transition-colors"
                                href="#">Kebijakan Akses Terbuka</a
                            >
                        </li>
                        <li>
                            <a
                                class="hover:text-primary transition-colors"
                                href="#">Panduan Sitasi</a
                            >
                        </li> -->
                    </ul>
                </div>
                <div>
                    <h4 class="font-bold text-slate-900 dark:text-white mb-4">
                        Dukungan
                    </h4>
                    <ul
                        class="space-y-3 text-sm text-slate-500 dark:text-slate-400"
                    >
                        <!-- <li>
                            <a
                                class="hover:text-primary transition-colors"
                                href="#">Pusat Bantuan</a
                            >
                        </li> -->
                        <li>
                            <a
                                class="hover:text-primary transition-colors"
                                href="#">Hubungi Kami</a
                            >
                        </li>
                        <!-- <li>
                            <a
                                class="hover:text-primary transition-colors"
                                href="#">Status Sistem</a
                            >
                        </li>
                        <li>
                            <a
                                class="hover:text-primary transition-colors"
                                href="#">Laporkan Masalah</a
                            >
                        </li> -->
                    </ul>
                </div>
            </div>
            <div
                class="border-t border-slate-200 dark:border-slate-800 pt-8 flex flex-col md:flex-row items-center justify-between gap-4"
            >
                <p
                    class="text-slate-400 dark:text-slate-600 text-sm text-center md:text-left"
                >
                    {#if $footerText}
                        {$footerText}
                    {:else}
                        © 2026 {$appName} Digital Repository. All rights reserved.
                    {/if}
                </p>
            </div>
        </div>
    </footer>
</div>

<style>
    .size-8 {
        width: 2rem;
        height: 2rem;
    }
    .size-6 {
        width: 1.5rem;
        height: 1.5rem;
    }
    @keyframes slideDown {
        from {
            opacity: 0;
            transform: translateY(-10px);
        }
        to {
            opacity: 1;
            transform: translateY(0);
        }
    }
    :global(.animate-slideDown) {
        animation: slideDown 0.2s ease-out;
    }
</style>
