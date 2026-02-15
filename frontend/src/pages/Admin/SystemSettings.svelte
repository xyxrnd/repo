<script>
    import { onMount } from "svelte";
    import { link } from "svelte-spa-router";
    import {
        getSiteSettings,
        updateSiteSettings,
        uploadSiteLogo,
    } from "../../services/siteSettingsService.js";
    import { API_BASE_URL } from "../../config/index.js";
    import { refreshSiteSettings } from "../../stores/index.js";

    let settings = {
        app_name: "",
        app_description: "",
        about_text: "",
        visi: "",
        misi: "",
        logo_url: "",
        address: "",
        email: "",
        phone: "",
        footer_text: "",
    };

    let loading = true;
    let saving = false;
    let error = "";
    let successMsg = "";

    // Logo upload state
    let logoFile = null;
    let logoPreview = "";
    let uploadingLogo = false;

    onMount(async () => {
        await loadSettings();
    });

    async function loadSettings() {
        try {
            loading = true;
            error = "";
            const data = await getSiteSettings();
            settings = { ...settings, ...data };
            if (settings.logo_url) {
                logoPreview = settings.logo_url.startsWith("http")
                    ? settings.logo_url
                    : `${API_BASE_URL}${settings.logo_url}`;
            }
        } catch (e) {
            error = "Gagal memuat pengaturan: " + e.message;
        } finally {
            loading = false;
        }
    }

    function handleLogoSelect(e) {
        const file = e.target.files[0];
        if (!file) return;

        // Validate
        const allowed = [
            "image/png",
            "image/jpeg",
            "image/svg+xml",
            "image/webp",
            "image/x-icon",
        ];
        if (!allowed.includes(file.type)) {
            error =
                "Format file tidak didukung. Gunakan PNG, JPG, SVG, WEBP, atau ICO.";
            return;
        }
        if (file.size > 5 * 1024 * 1024) {
            error = "Ukuran file maksimal 5MB.";
            return;
        }

        logoFile = file;
        logoPreview = URL.createObjectURL(file);
        error = "";
    }

    async function handleUploadLogo() {
        if (!logoFile) return;
        uploadingLogo = true;
        error = "";
        try {
            const result = await uploadSiteLogo(logoFile);
            settings.logo_url = result.logo_url;
            logoPreview = `${API_BASE_URL}${result.logo_url}`;
            logoFile = null;
            showSuccess("Logo berhasil diupload!");
        } catch (e) {
            error = "Gagal upload logo: " + e.message;
        } finally {
            uploadingLogo = false;
        }
    }

    function removeLogo() {
        logoFile = null;
        logoPreview = "";
        settings.logo_url = "";
    }

    async function handleSave() {
        saving = true;
        error = "";
        try {
            // Upload logo first if a new one is selected
            if (logoFile) {
                await handleUploadLogo();
            }

            // Save other settings
            const { logo_url, ...otherSettings } = settings;
            await updateSiteSettings(otherSettings);

            // Refresh global store so landing page and other components update
            await refreshSiteSettings();

            showSuccess("Pengaturan berhasil disimpan!");
        } catch (e) {
            error = "Gagal menyimpan: " + e.message;
        } finally {
            saving = false;
        }
    }

    function showSuccess(msg) {
        successMsg = msg;
        setTimeout(() => {
            successMsg = "";
        }, 3000);
    }
</script>

<div class="max-w-4xl mx-auto flex flex-col gap-6">
    <!-- Breadcrumb -->
    <nav class="flex items-center gap-2 text-sm">
        <a
            href="#/"
            use:link
            class="text-slate-500 hover:text-primary transition-colors"
            >Dashboard</a
        >
        <span class="text-slate-400">/</span>
        <span class="text-slate-900 dark:text-white font-medium"
            >Pengaturan Sistem</span
        >
    </nav>

    <!-- Header -->
    <div class="flex flex-col gap-1">
        <h2
            class="text-3xl font-black text-slate-900 dark:text-white tracking-tight"
        >
            Pengaturan Sistem
        </h2>
        <p class="text-slate-500 dark:text-slate-400">
            Kelola identitas dan informasi sistem repositori.
        </p>
    </div>

    <!-- Toast Messages -->
    {#if successMsg}
        <div
            class="fixed top-4 right-4 z-50 px-6 py-3 rounded-lg shadow-lg flex items-center gap-3 bg-emerald-500 text-white animate-slide-in"
        >
            <span class="material-symbols-outlined">check_circle</span>
            {successMsg}
        </div>
    {/if}

    {#if error}
        <div
            class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-xl p-4 text-red-700 dark:text-red-400 flex items-center gap-3"
        >
            <span class="material-symbols-outlined">error</span>
            <p>{error}</p>
        </div>
    {/if}

    {#if loading}
        <div class="flex justify-center items-center py-20">
            <div
                class="animate-spin rounded-full h-16 w-16 border-4 border-primary border-t-transparent"
            ></div>
        </div>
    {:else}
        <form class="space-y-6" on:submit|preventDefault={handleSave}>
            <!-- Logo & Identity Card -->
            <div
                class="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden"
            >
                <div
                    class="p-6 border-b border-slate-200 dark:border-slate-800"
                >
                    <h3
                        class="text-lg font-bold text-slate-900 dark:text-white flex items-center gap-2"
                    >
                        <span class="material-symbols-outlined text-primary"
                            >image</span
                        >
                        Logo & Identitas
                    </h3>
                    <p class="text-sm text-slate-500 mt-1">
                        Atur logo dan nama sistem yang ditampilkan di seluruh
                        halaman.
                    </p>
                </div>
                <div class="p-6 space-y-6">
                    <!-- Logo Upload -->
                    <div>
                        <label
                            for="logo-upload"
                            class="block text-sm font-bold mb-3 text-slate-700 dark:text-slate-300"
                        >
                            Logo Aplikasi
                        </label>
                        <div class="flex items-start gap-6">
                            <!-- Preview -->
                            <div class="shrink-0">
                                {#if logoPreview}
                                    <div
                                        class="relative group w-24 h-24 rounded-xl border-2 border-dashed border-slate-300 dark:border-slate-600 overflow-hidden bg-slate-50 dark:bg-slate-800"
                                    >
                                        <img
                                            src={logoPreview}
                                            alt="Logo preview"
                                            class="w-full h-full object-contain p-2"
                                        />
                                        <button
                                            type="button"
                                            on:click={removeLogo}
                                            class="absolute inset-0 bg-red-500/80 text-white flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity"
                                        >
                                            <span
                                                class="material-symbols-outlined"
                                                >delete</span
                                            >
                                        </button>
                                    </div>
                                {:else}
                                    <div
                                        class="w-24 h-24 rounded-xl border-2 border-dashed border-slate-300 dark:border-slate-600 flex items-center justify-center bg-slate-50 dark:bg-slate-800"
                                    >
                                        <span
                                            class="material-symbols-outlined text-3xl text-slate-400"
                                            >add_photo_alternate</span
                                        >
                                    </div>
                                {/if}
                            </div>

                            <!-- Upload Controls -->
                            <div class="flex-1">
                                <label
                                    for="logo-upload"
                                    class="inline-flex items-center gap-2 px-4 py-2.5 bg-primary text-white rounded-lg font-semibold text-sm cursor-pointer hover:bg-primary/90 transition-colors shadow-lg shadow-primary/25"
                                >
                                    <span
                                        class="material-symbols-outlined text-lg"
                                        >upload</span
                                    >
                                    Pilih Logo
                                </label>
                                <input
                                    id="logo-upload"
                                    type="file"
                                    accept="image/png,image/jpeg,image/svg+xml,image/webp"
                                    on:change={handleLogoSelect}
                                    class="hidden"
                                />
                                <p
                                    class="mt-2 text-xs text-slate-400 leading-relaxed"
                                >
                                    Format: PNG, JPG, SVG, WEBP (Maks 5MB)<br />
                                    Rekomendasi: 200×200px atau lebih, latar transparan
                                </p>
                                {#if logoFile}
                                    <div
                                        class="mt-2 flex items-center gap-2 text-sm text-emerald-600 dark:text-emerald-400"
                                    >
                                        <span
                                            class="material-symbols-outlined text-base"
                                            >check_circle</span
                                        >
                                        <span class="font-medium"
                                            >{logoFile.name}</span
                                        >
                                        <span class="text-xs text-slate-400">
                                            ({Math.round(
                                                logoFile.size / 1024,
                                            )}KB)
                                        </span>
                                    </div>
                                {/if}
                            </div>
                        </div>
                    </div>

                    <!-- App Name -->
                    <div class="grid md:grid-cols-2 gap-6">
                        <div>
                            <label
                                for="sys-app-name"
                                class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
                            >
                                Nama Aplikasi <span class="text-red-500">*</span
                                >
                            </label>
                            <input
                                id="sys-app-name"
                                type="text"
                                bind:value={settings.app_name}
                                class="w-full px-4 py-3 bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent transition-all"
                                placeholder="Contoh: ScholarHub"
                            />
                            <p class="mt-1 text-xs text-slate-400">
                                Nama ini ditampilkan di navbar, footer, dan
                                judul halaman.
                            </p>
                        </div>
                        <div>
                            <label
                                for="sys-app-desc"
                                class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
                            >
                                Deskripsi Singkat
                            </label>
                            <input
                                id="sys-app-desc"
                                type="text"
                                bind:value={settings.app_description}
                                class="w-full px-4 py-3 bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent transition-all"
                                placeholder="Contoh: Repository Dokumen Akademik"
                            />
                            <p class="mt-1 text-xs text-slate-400">
                                Subtitle / tagline yang muncul di halaman utama.
                            </p>
                        </div>
                    </div>
                </div>
            </div>

            <!-- About Card -->
            <div
                class="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden"
            >
                <div
                    class="p-6 border-b border-slate-200 dark:border-slate-800"
                >
                    <h3
                        class="text-lg font-bold text-slate-900 dark:text-white flex items-center gap-2"
                    >
                        <span class="material-symbols-outlined text-primary"
                            >info</span
                        >
                        Tentang Sistem
                    </h3>
                    <p class="text-sm text-slate-500 mt-1">
                        Informasi tentang sistem yang ditampilkan di halaman
                        About.
                    </p>
                </div>
                <div class="p-6 space-y-6">
                    <div>
                        <label
                            for="sys-about"
                            class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
                        >
                            Teks About
                        </label>
                        <textarea
                            id="sys-about"
                            bind:value={settings.about_text}
                            rows="5"
                            class="w-full px-4 py-3 bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent transition-all min-h-[140px] resize-y"
                            placeholder="Deskripsikan tentang sistem repositori ini..."
                        ></textarea>
                        <p class="mt-1 text-xs text-slate-400">
                            Teks ini akan ditampilkan di halaman About dan
                            bagian hero description.
                        </p>
                    </div>

                    <div>
                        <label
                            for="sys-footer"
                            class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
                        >
                            Teks Footer
                        </label>
                        <input
                            id="sys-footer"
                            type="text"
                            bind:value={settings.footer_text}
                            class="w-full px-4 py-3 bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent transition-all"
                            placeholder="Contoh: Hak Cipta © 2024 Universitas Negeri"
                        />
                    </div>
                </div>
            </div>

            <!-- Halaman About Card -->
            <div
                class="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden"
            >
                <div
                    class="p-6 border-b border-slate-200 dark:border-slate-800"
                >
                    <h3
                        class="text-lg font-bold text-slate-900 dark:text-white flex items-center gap-2"
                    >
                        <span class="material-symbols-outlined text-primary"
                            >description</span
                        >
                        Halaman About
                    </h3>
                    <p class="text-sm text-slate-500 mt-1">
                        Atur konten Visi & Misi yang ditampilkan di halaman
                        About.
                    </p>
                </div>
                <div class="p-6 space-y-6">
                    <!-- Visi -->
                    <div>
                        <label
                            for="sys-visi"
                            class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
                        >
                            <span
                                class="material-symbols-outlined text-base align-middle mr-1"
                                >visibility</span
                            >
                            Visi
                        </label>
                        <textarea
                            id="sys-visi"
                            bind:value={settings.visi}
                            rows="3"
                            class="w-full px-4 py-3 bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent transition-all min-h-[100px] resize-y"
                            placeholder="Contoh: Menjadi pusat informasi dan dokumentasi karya ilmiah terdepan..."
                        ></textarea>
                        <p class="mt-1 text-xs text-slate-400">
                            Teks visi yang ditampilkan di halaman About.
                        </p>
                    </div>

                    <!-- Misi -->
                    <div>
                        <label
                            for="sys-misi"
                            class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
                        >
                            <span
                                class="material-symbols-outlined text-base align-middle mr-1"
                                >flag</span
                            >
                            Misi
                        </label>
                        <textarea
                            id="sys-misi"
                            bind:value={settings.misi}
                            rows="5"
                            class="w-full px-4 py-3 bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent transition-all min-h-[140px] resize-y"
                            placeholder="Tulis setiap poin misi di baris baru, contoh:
Menyediakan akses mudah terhadap karya ilmiah
Mendokumentasikan hasil penelitian civitas akademika
Mendukung pengembangan riset dan inovasi"
                        ></textarea>
                        <p class="mt-1 text-xs text-slate-400">
                            Tulis setiap poin misi di baris baru. Setiap baris
                            akan ditampilkan sebagai item terpisah di halaman
                            About.
                        </p>
                    </div>
                </div>
            </div>

            <!-- Contact Info Card -->
            <div
                class="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden"
            >
                <div
                    class="p-6 border-b border-slate-200 dark:border-slate-800"
                >
                    <h3
                        class="text-lg font-bold text-slate-900 dark:text-white flex items-center gap-2"
                    >
                        <span class="material-symbols-outlined text-primary"
                            >contact_mail</span
                        >
                        Informasi Kontak
                    </h3>
                    <p class="text-sm text-slate-500 mt-1">
                        Informasi kontak yang ditampilkan di halaman About dan
                        footer.
                    </p>
                </div>
                <div class="p-6 space-y-6">
                    <div class="grid md:grid-cols-3 gap-6">
                        <div>
                            <label
                                for="sys-address"
                                class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
                            >
                                <span
                                    class="material-symbols-outlined text-base align-middle mr-1"
                                    >location_on</span
                                >
                                Alamat
                            </label>
                            <input
                                id="sys-address"
                                type="text"
                                bind:value={settings.address}
                                class="w-full px-4 py-3 bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent transition-all"
                                placeholder="Jl. Universitas No. 1"
                            />
                        </div>
                        <div>
                            <label
                                for="sys-email"
                                class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
                            >
                                <span
                                    class="material-symbols-outlined text-base align-middle mr-1"
                                    >email</span
                                >
                                Email
                            </label>
                            <input
                                id="sys-email"
                                type="email"
                                bind:value={settings.email}
                                class="w-full px-4 py-3 bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent transition-all"
                                placeholder="repository@univ.ac.id"
                            />
                        </div>
                        <div>
                            <label
                                for="sys-phone"
                                class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
                            >
                                <span
                                    class="material-symbols-outlined text-base align-middle mr-1"
                                    >call</span
                                >
                                Telepon
                            </label>
                            <input
                                id="sys-phone"
                                type="text"
                                bind:value={settings.phone}
                                class="w-full px-4 py-3 bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent transition-all"
                                placeholder="(021) 1234567"
                            />
                        </div>
                    </div>
                </div>
            </div>

            <!-- Preview Card -->
            <div
                class="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden"
            >
                <div
                    class="p-6 border-b border-slate-200 dark:border-slate-800"
                >
                    <h3
                        class="text-lg font-bold text-slate-900 dark:text-white flex items-center gap-2"
                    >
                        <span class="material-symbols-outlined text-primary"
                            >preview</span
                        >
                        Pratinjau
                    </h3>
                    <p class="text-sm text-slate-500 mt-1">
                        Tampilan yang akan terlihat di halaman publik.
                    </p>
                </div>
                <div class="p-6">
                    <div
                        class="border border-slate-200 dark:border-slate-700 rounded-xl overflow-hidden"
                    >
                        <!-- Mock Navbar -->
                        <div
                            class="flex items-center gap-3 px-5 py-3 bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-700"
                        >
                            {#if logoPreview}
                                <img
                                    src={logoPreview}
                                    alt="Logo"
                                    class="w-8 h-8 object-contain"
                                />
                            {:else}
                                <span
                                    class="material-symbols-outlined text-primary text-2xl"
                                    >local_library</span
                                >
                            {/if}
                            <span
                                class="text-lg font-bold text-slate-900 dark:text-white"
                            >
                                {settings.app_name || "ScholarHub"}
                            </span>
                            <span
                                class="text-sm text-slate-400 ml-2 hidden sm:inline"
                            >
                                — {settings.app_description ||
                                    "Repository Dokumen Akademik"}
                            </span>
                        </div>

                        <!-- Mock Hero -->
                        <div
                            class="bg-gradient-to-br from-primary via-blue-600 to-indigo-700 p-8 text-white text-center"
                        >
                            <h4 class="text-2xl font-bold mb-2">
                                {settings.app_name || "ScholarHub"}
                            </h4>
                            <p class="text-blue-100 text-sm max-w-md mx-auto">
                                {settings.about_text ||
                                    "Platform digital untuk menyimpan, mengelola, dan menyebarluaskan karya ilmiah."}
                            </p>
                        </div>

                        <!-- Mock Footer -->
                        <div
                            class="bg-slate-900 px-5 py-3 text-center text-xs text-slate-400"
                        >
                            {settings.footer_text ||
                                `© 2025 ${settings.app_name || "ScholarHub"} Digital Repository`}
                        </div>
                    </div>
                </div>
            </div>

            <!-- Actions -->
            <div class="flex justify-end gap-3 py-4">
                <a
                    href="#/dashboard"
                    use:link
                    class="px-6 py-3 bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 rounded-xl font-bold hover:bg-slate-200 dark:hover:bg-slate-700 transition-all"
                >
                    Batal
                </a>
                <button
                    type="submit"
                    class="flex items-center gap-2 px-6 py-3 bg-primary text-white font-bold rounded-xl shadow-lg shadow-primary/25 hover:bg-primary/90 transition-all active:scale-95 disabled:opacity-50 disabled:cursor-not-allowed"
                    disabled={saving}
                >
                    {#if saving}
                        <div
                            class="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin"
                        ></div>
                    {:else}
                        <span class="material-symbols-outlined text-xl"
                            >save</span
                        >
                    {/if}
                    <span>{saving ? "Menyimpan..." : "Simpan Pengaturan"}</span>
                </button>
            </div>
        </form>
    {/if}
</div>

<style>
    @keyframes slide-in {
        from {
            transform: translateX(100%);
            opacity: 0;
        }
        to {
            transform: translateX(0);
            opacity: 1;
        }
    }
    .animate-slide-in {
        animation: slide-in 0.3s ease-out;
    }
</style>
