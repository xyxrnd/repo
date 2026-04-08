<script>
    import { onMount } from "svelte";
    import { querystring } from "svelte-spa-router";
    import { getDocumentById } from "../../services/documentService";
    import authService from "../../services/authService";
    import { API_BASE_URL, API_ENDPOINTS } from "../../config";

    export let id = "";
    export let params = {};

    let doc = null;
    let loading = true;
    let error = "";
    let isLoggedIn = false;
    let isMahasiswa = false;

    // Access Request state (per-dokumen, bukan per-file)
    let showAccessForm = false;
    let accessFormData = { nama: "", email: "", ktm: null };
    let accessFormLoading = false;
    let accessFormSuccess = "";
    let accessFormError = "";

    // OTP Verification state (step sebelum submit access request)
    let otpStep = 1; // 1 = email input, 2 = OTP input, 3 = form lengkap
    let otpEmail = "";
    let otpCode = "";
    let otpLoading = false;
    let otpError = "";
    let otpSuccess = "";
    let otpCountdown = 0;
    let otpTimer = null;

    // Token verification state (otomatis via link email)
    let tokenLoading = false;
    let tokenError = "";
    let tokenSuccess = false;
    let tokenAutoVerified = false;
    let unlockedFiles = {}; // { file_id: { file_path, file_name } }

    // Get ID from route params (prop passed by svelte-spa-router)
    $: if (params && params.id) {
        id = params.id;
        loadDocument();
    }

    // Auto-detect token from URL query parameter
    $: if ($querystring && doc && !tokenAutoVerified) {
        const urlParams = new URLSearchParams($querystring);
        const urlToken = urlParams.get("token");
        if (urlToken) {
            tokenAutoVerified = true;
            autoVerifyToken(urlToken);
        }
    }

    onMount(() => {
        isLoggedIn = authService.isAuthenticated();
        isMahasiswa = authService.isMahasiswa();
    });

    async function loadDocument() {
        try {
            loading = true;
            error = "";
            doc = await getDocumentById(id);
        } catch (e) {
            error = "Dokumen tidak ditemukan atau terjadi kesalahan.";
            console.error("Failed to load document:", e);
        } finally {
            loading = false;
        }
    }

    function handleDownload() {
        // Download semua file dalam ZIP via endpoint download-all
        window.open(API_ENDPOINTS.DOCUMENT_DOWNLOAD_ALL(doc.id), "_blank");
    }

    function handleFileDownload(file) {
        if (file.is_locked && !unlockedFiles[file.id]) {
            if (!isLoggedIn) {
                window.location.hash = "/student-signup";
            } else {
                showAccessForm = true;
            }
            return;
        }
        window.open(`${API_BASE_URL}/${file.file_path}`, "_blank");
    }

    function handleFilePreview(file) {
        if (file.is_locked && !unlockedFiles[file.id]) {
            if (!isLoggedIn) {
                window.location.hash = "/student-signup";
            } else {
                showAccessForm = true;
            }
            return;
        }
        window.open(`${API_BASE_URL}/${file.file_path}`, "_blank");
    }

    // Variabel untuk menyimpan email tersensor sebagai placeholder hint untuk mahasiswa
    let maskedEmail = "";

    // Helper: sensor bagian tengah email (mis: rendhi@gmail.com → r****i@gmail.com)
    function maskEmail(email) {
        if (!email) return "";
        const [local, domain] = email.split("@");
        if (!domain) return email;
        if (local.length <= 2) return local[0] + "***@" + domain;
        const first = local[0];
        const last = local[local.length - 1];
        const stars = "*".repeat(Math.min(local.length - 2, 5));
        return `${first}${stars}${last}@${domain}`;
    }

    function openAccessForm() {
        showAccessForm = true;
        accessFormSuccess = "";
        accessFormError = "";
        otpCode = "";
        otpError = "";
        otpSuccess = "";
        otpCountdown = 0;
        if (otpTimer) clearInterval(otpTimer);

        otpStep = 1;
        otpEmail = "";
        accessFormData = { nama: "", email: "", ktm: null };

        // Jika mahasiswa, siapkan placeholder email tersensor
        if (isMahasiswa) {
            const currentUser = authService.getUser();
            maskedEmail = maskEmail(currentUser?.email || "");
        } else {
            maskedEmail = "";
        }
    }

    function closeAccessForm() {
        showAccessForm = false;
        accessFormData = { nama: "", email: "", ktm: null };
        accessFormSuccess = "";
        accessFormError = "";
        otpStep = 1;
        otpEmail = "";
        otpCode = "";
        otpError = "";
        otpSuccess = "";
        otpCountdown = 0;
        maskedEmail = "";
        if (otpTimer) clearInterval(otpTimer);
    }

    function handleKtmUpload(event) {
        const file = event.target.files[0];
        if (file) {
            accessFormData.ktm = file;
        }
    }

    // Step 1: Kirim OTP ke email
    async function sendOTP() {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!otpEmail.trim()) {
            otpError = "Email wajib diisi.";
            return;
        }
        if (!emailRegex.test(otpEmail.trim())) {
            otpError = "Format email tidak valid.";
            return;
        }

        // Validasi: jika mahasiswa, email harus sesuai dengan email akun
        if (isMahasiswa) {
            const currentUser = authService.getUser();
            if (
                currentUser &&
                otpEmail.trim().toLowerCase() !==
                    currentUser.email.toLowerCase()
            ) {
                otpError =
                    "Email tidak sesuai dengan akun Anda. Gunakan email yang terdaftar.";
                return;
            }
        }

        otpLoading = true;
        otpError = "";
        otpSuccess = "";
        try {
            const response = await fetch(API_ENDPOINTS.SEND_OTP, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    email: otpEmail.trim(),
                    document_id: doc.id,
                }),
            });
            if (response.status === 429) {
                otpError =
                    "Kode OTP sudah dikirim. Tunggu 1 menit sebelum mengirim ulang.";
                return;
            }
            if (!response.ok) {
                const text = await response.text();
                throw new Error(text || "Gagal mengirim OTP");
            }
            otpStep = 2;
            otpSuccess = `Kode OTP telah dikirim ke ${otpEmail}. Cek inbox email Anda.`;
            // Start countdown 60 detik untuk resend
            otpCountdown = 60;
            if (otpTimer) clearInterval(otpTimer);
            otpTimer = setInterval(() => {
                otpCountdown--;
                if (otpCountdown <= 0) {
                    clearInterval(otpTimer);
                    otpTimer = null;
                }
            }, 1000);
        } catch (e) {
            otpError = e.message || "Gagal mengirim OTP.";
        } finally {
            otpLoading = false;
        }
    }

    // Step 2: Verifikasi OTP
    async function verifyOTP() {
        if (!otpCode.trim()) {
            otpError = "Masukkan kode OTP.";
            return;
        }
        if (otpCode.trim().length !== 6) {
            otpError = "Kode OTP harus 6 digit.";
            return;
        }

        otpLoading = true;
        otpError = "";
        try {
            const response = await fetch(API_ENDPOINTS.VERIFY_OTP, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    email: otpEmail.trim(),
                    document_id: doc.id,
                    otp_code: otpCode.trim(),
                }),
            });
            if (!response.ok) {
                const text = await response.text();
                throw new Error(text || "Kode OTP tidak valid");
            }
            otpStep = 3;
            otpSuccess =
                "Email berhasil diverifikasi! Silakan lengkapi formulir di bawah.";
            accessFormData.email = otpEmail.trim();
            // Jika mahasiswa, auto-fill nama dari akun
            if (isMahasiswa) {
                const currentUser = authService.getUser();
                if (currentUser?.name) accessFormData.nama = currentUser.name;
            }
            if (otpTimer) clearInterval(otpTimer);
        } catch (e) {
            otpError = e.message || "Kode OTP tidak valid.";
        } finally {
            otpLoading = false;
        }
    }

    // Auto-verify token from URL (link dari email)
    async function autoVerifyToken(token) {
        tokenLoading = true;
        tokenError = "";
        tokenSuccess = false;
        try {
            await doVerifyToken(token);
            tokenSuccess = true;
        } finally {
            tokenLoading = false;
        }
    }

    async function doVerifyToken(token) {
        const response = await fetch(API_ENDPOINTS.VERIFY_ACCESS_TOKEN, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ document_id: doc.id, token }),
        });
        if (!response.ok) {
            const text = await response.text();
            tokenError = text || "Token tidak valid atau belum disetujui.";
            return;
        }
        const data = await response.json();
        // Unlock all locked files from response
        if (data.files && data.files.length > 0) {
            let newUnlocked = { ...unlockedFiles };
            for (const f of data.files) {
                newUnlocked[f.file_id] = {
                    file_path: f.file_path,
                    file_name: f.file_name,
                };
            }
            unlockedFiles = newUnlocked;
        }
    }

    function isFileUnlocked(fileId) {
        return !!unlockedFiles[fileId];
    }

    function handleUnlockedPreview(fileId) {
        const info = unlockedFiles[fileId];
        if (info) {
            window.open(`${API_BASE_URL}/${info.file_path}`, "_blank");
        }
    }

    function handleUnlockedDownload(fileId) {
        const info = unlockedFiles[fileId];
        if (info) {
            const link = document.createElement("a");
            link.href = `${API_BASE_URL}/${info.file_path}`;
            link.download = info.file_name;
            link.click();
        }
    }

    // Check if document has any locked files
    $: hasLockedFiles = doc && doc.files && doc.files.some((f) => f.is_locked);
    $: allLockedUnlocked =
        doc &&
        doc.files &&
        doc.files.filter((f) => f.is_locked).every((f) => unlockedFiles[f.id]);

    async function submitAccessRequest() {
        if (!accessFormData.nama.trim()) {
            accessFormError = "Nama wajib diisi.";
            return;
        }
        if (!accessFormData.ktm) {
            accessFormError = "Upload KTM wajib dilakukan.";
            return;
        }
        try {
            accessFormLoading = true;
            accessFormError = "";
            const formData = new FormData();
            formData.append("document_id", doc.id);
            formData.append("nama", accessFormData.nama.trim());
            formData.append("email", accessFormData.email.trim());
            if (accessFormData.ktm) {
                formData.append("ktm", accessFormData.ktm);
            }
            /** @type {Record<string, string>} */
            const headers = {};
            if (isMahasiswa) {
                const token = authService.getToken();
                if (token) headers["Authorization"] = `Bearer ${token}`;
            }
            const response = await fetch(API_ENDPOINTS.ACCESS_REQUESTS, {
                method: "POST",
                headers,
                body: formData,
            });
            if (response.status === 409) {
                accessFormError =
                    "Anda sudah memiliki permintaan akses yang masih menunggu persetujuan untuk dokumen ini.";
                return;
            }
            if (!response.ok) {
                const text = await response.text();
                throw new Error(text || "Gagal mengirim permintaan akses");
            }
            accessFormSuccess =
                "Permintaan akses berhasil dikirim! Silakan tunggu persetujuan dari admin. Link akses akan dikirim ke email Anda.";
            accessFormData = { nama: "", email: "", ktm: null };
        } catch (e) {
            accessFormError =
                e.message || "Terjadi kesalahan saat mengirim permintaan.";
        } finally {
            accessFormLoading = false;
        }
    }

    function formatDate(dateString) {
        const date = new Date(dateString);
        return date.toLocaleDateString("id-ID", {
            day: "numeric",
            month: "long",
            year: "numeric",
        });
    }

    function formatFileSize(bytes) {
        if (!bytes || bytes === 0) return "0 B";
        const k = 1024;
        const sizes = ["B", "KB", "MB", "GB"];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + " " + sizes[i];
    }

    function getTypeStyle(jenisFile) {
        const styles = {
            Skripsi: {
                icon: "school",
                bg: "bg-blue-500",
                light: "bg-blue-100 dark:bg-blue-900/30",
                text: "text-blue-600 dark:text-blue-400",
                gradient: "from-blue-500 to-blue-700",
            },
            skripsi: {
                icon: "school",
                bg: "bg-blue-500",
                light: "bg-blue-100 dark:bg-blue-900/30",
                text: "text-blue-600 dark:text-blue-400",
                gradient: "from-blue-500 to-blue-700",
            },
            Tesis: {
                icon: "workspace_premium",
                bg: "bg-purple-500",
                light: "bg-purple-100 dark:bg-purple-900/30",
                text: "text-purple-600 dark:text-purple-400",
                gradient: "from-purple-500 to-purple-700",
            },
            tesis: {
                icon: "workspace_premium",
                bg: "bg-purple-500",
                light: "bg-purple-100 dark:bg-purple-900/30",
                text: "text-purple-600 dark:text-purple-400",
                gradient: "from-purple-500 to-purple-700",
            },
            Jurnal: {
                icon: "article",
                bg: "bg-teal-500",
                light: "bg-teal-100 dark:bg-teal-900/30",
                text: "text-teal-600 dark:text-teal-400",
                gradient: "from-teal-500 to-teal-700",
            },
            jurnal: {
                icon: "article",
                bg: "bg-teal-500",
                light: "bg-teal-100 dark:bg-teal-900/30",
                text: "text-teal-600 dark:text-teal-400",
                gradient: "from-teal-500 to-teal-700",
            },
            Disertasi: {
                icon: "history_edu",
                bg: "bg-amber-500",
                light: "bg-amber-100 dark:bg-amber-900/30",
                text: "text-amber-600 dark:text-amber-400",
                gradient: "from-amber-500 to-amber-700",
            },
        };
        return (
            styles[jenisFile] || {
                icon: "description",
                bg: "bg-slate-500",
                light: "bg-slate-100 dark:bg-slate-700",
                text: "text-slate-600 dark:text-slate-400",
                gradient: "from-slate-500 to-slate-700",
            }
        );
    }

    function getFileIcon(fileName) {
        if (!fileName) return "description";
        const ext = fileName.split(".").pop().toLowerCase();
        const icons = {
            pdf: "picture_as_pdf",
            doc: "description",
            docx: "description",
            xls: "table_chart",
            xlsx: "table_chart",
            ppt: "slideshow",
            pptx: "slideshow",
        };
        return icons[ext] || "description";
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
    {#if loading}
        <!-- Loading State -->
        <div
            class="bg-gradient-to-br from-primary via-blue-600 to-indigo-700 text-white py-16 px-4"
        >
            <div class="container mx-auto max-w-4xl">
                <div
                    class="h-8 w-48 bg-white/20 rounded animate-pulse mb-4"
                ></div>
                <div
                    class="h-12 w-full bg-white/20 rounded animate-pulse mb-4"
                ></div>
                <div class="h-6 w-64 bg-white/20 rounded animate-pulse"></div>
            </div>
        </div>
        <div class="container mx-auto max-w-4xl px-4 py-8">
            <div class="bg-white dark:bg-slate-800 rounded-xl p-8">
                <div class="space-y-4">
                    {#each Array(5) as _}
                        <div
                            class="h-5 bg-slate-200 dark:bg-slate-700 rounded animate-pulse"
                        ></div>
                    {/each}
                </div>
            </div>
        </div>
    {:else if error}
        <!-- Error State -->
        <div
            class="bg-gradient-to-br from-red-500 to-red-700 text-white py-16 px-4"
        >
            <div class="container mx-auto max-w-4xl text-center">
                <span class="material-symbols-outlined text-6xl mb-4 opacity-50"
                    >error</span
                >
                <h1 class="text-3xl font-bold mb-2">Dokumen Tidak Ditemukan</h1>
                <p class="text-red-100 mb-6">{error}</p>
                <a
                    href="#/browse"
                    class="inline-flex items-center gap-2 px-6 py-3 bg-white/20 hover:bg-white/30 backdrop-blur-sm rounded-lg font-medium transition-all"
                >
                    <span class="material-symbols-outlined">arrow_back</span>
                    Kembali ke Jelajah
                </a>
            </div>
        </div>
    {:else if doc}
        <!-- Hero Header -->
        {@const typeStyle = getTypeStyle(doc.jenis_file)}
        <div
            class="bg-gradient-to-br from-primary via-blue-600 to-indigo-700 text-white py-12 lg:py-16 px-4 relative overflow-hidden"
        >
            <!-- Decorative background -->
            <div class="absolute inset-0 opacity-10">
                <div
                    class="absolute -top-24 -right-24 w-96 h-96 rounded-full bg-white/20 blur-3xl"
                ></div>
                <div
                    class="absolute -bottom-24 -left-24 w-64 h-64 rounded-full bg-white/10 blur-2xl"
                ></div>
            </div>

            <div class="container mx-auto max-w-4xl relative">
                <!-- Breadcrumb -->
                <nav class="flex items-center gap-2 text-sm text-blue-200 mb-6">
                    <a href="#/" class="hover:text-white transition-colors"
                        >Home</a
                    >
                    <span>›</span>
                    <a
                        href="#/browse"
                        class="hover:text-white transition-colors">Jelajah</a
                    >
                    <span>›</span>
                    <span class="text-white/70 truncate max-w-[200px]"
                        >{doc.judul}</span
                    >
                </nav>

                <!-- Category badge -->
                <div class="flex items-center gap-3 mb-4">
                    <span
                        class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-white/20 backdrop-blur-sm rounded-full text-sm font-semibold"
                    >
                        <span
                            class="material-symbols-outlined"
                            style="font-size: 16px;">{typeStyle.icon}</span
                        >
                        {doc.jenis_file}
                    </span>
                    {#if doc.view_count > 0}
                        <span
                            class="inline-flex items-center gap-1 px-2.5 py-1 bg-white/10 backdrop-blur-sm rounded-full text-xs font-medium"
                        >
                            <span
                                class="material-symbols-outlined"
                                style="font-size: 14px;">visibility</span
                            >
                            {doc.view_count} kali dilihat
                        </span>
                    {/if}
                </div>

                <!-- Title -->
                <h1
                    class="text-2xl md:text-4xl font-black leading-tight mb-4 max-w-3xl"
                >
                    {doc.judul}
                </h1>

                <!-- Author & Date -->
                <div class="flex flex-wrap items-center gap-4 text-blue-100">
                    <div class="flex items-center gap-2">
                        <div
                            class="w-8 h-8 rounded-full bg-white/20 flex items-center justify-center text-xs font-bold"
                        >
                            {getInitials(doc.penulis)}
                        </div>
                        <span class="font-medium">{doc.penulis}</span>
                    </div>
                    <span class="hidden md:inline text-blue-300">•</span>
                    <div class="flex items-center gap-1.5">
                        <span
                            class="material-symbols-outlined"
                            style="font-size: 16px;">calendar_today</span
                        >
                        {formatDate(doc.created_at)}
                    </div>
                </div>
            </div>
        </div>

        <!-- Content -->
        <div class="container mx-auto max-w-4xl px-4 py-8">
            <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
                <!-- Main Content (Left Column) -->
                <div class="lg:col-span-2 space-y-6">
                    <!-- Informasi Dokumen Card -->
                    <div
                        class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden shadow-sm"
                    >
                        <div
                            class="px-6 py-4 border-b border-slate-100 dark:border-slate-700"
                        >
                            <h2
                                class="text-lg font-bold text-slate-900 dark:text-white flex items-center gap-2"
                            >
                                <span
                                    class="material-symbols-outlined text-primary"
                                    >info</span
                                >
                                Informasi Dokumen
                            </h2>
                        </div>
                        <div class="p-6">
                            <table class="w-full">
                                <tbody>
                                    <tr
                                        class="border-b border-slate-100 dark:border-slate-700/50"
                                    >
                                        <td
                                            class="py-3 pr-4 text-sm font-semibold text-slate-500 dark:text-slate-400 w-40 align-top"
                                            >Judul</td
                                        >
                                        <td
                                            class="py-3 text-sm text-slate-900 dark:text-white font-medium"
                                            >{doc.judul}</td
                                        >
                                    </tr>
                                    <tr
                                        class="border-b border-slate-100 dark:border-slate-700/50"
                                    >
                                        <td
                                            class="py-3 pr-4 text-sm font-semibold text-slate-500 dark:text-slate-400 w-40 align-top"
                                            >Penulis</td
                                        >
                                        <td
                                            class="py-3 text-sm text-slate-900 dark:text-white"
                                            >{doc.penulis}</td
                                        >
                                    </tr>
                                    <tr
                                        class="border-b border-slate-100 dark:border-slate-700/50"
                                    >
                                        <td
                                            class="py-3 pr-4 text-sm font-semibold text-slate-500 dark:text-slate-400 w-40 align-top"
                                            >Jenis Dokumen</td
                                        >
                                        <td class="py-3">
                                            <span
                                                class="inline-flex items-center gap-1.5 px-2.5 py-1 {typeStyle.light} {typeStyle.text} text-xs font-bold rounded-full uppercase"
                                            >
                                                <span
                                                    class="material-symbols-outlined"
                                                    style="font-size: 14px;"
                                                    >{typeStyle.icon}</span
                                                >
                                                {doc.jenis_file}
                                            </span>
                                        </td>
                                    </tr>
                                    {#if doc.fakultas_nama}
                                        <tr
                                            class="border-b border-slate-100 dark:border-slate-700/50"
                                        >
                                            <td
                                                class="py-3 pr-4 text-sm font-semibold text-slate-500 dark:text-slate-400 w-40 align-top"
                                                >Fakultas</td
                                            >
                                            <td
                                                class="py-3 text-sm text-slate-900 dark:text-white"
                                                >{doc.fakultas_nama}</td
                                            >
                                        </tr>
                                    {/if}
                                    {#if doc.prodi_nama}
                                        <tr
                                            class="border-b border-slate-100 dark:border-slate-700/50"
                                        >
                                            <td
                                                class="py-3 pr-4 text-sm font-semibold text-slate-500 dark:text-slate-400 w-40 align-top"
                                                >Program Studi</td
                                            >
                                            <td
                                                class="py-3 text-sm text-slate-900 dark:text-white"
                                                >{doc.prodi_nama}</td
                                            >
                                        </tr>
                                    {/if}
                                    {#if doc.dosen_pembimbing}
                                        <tr
                                            class="border-b border-slate-100 dark:border-slate-700/50"
                                        >
                                            <td
                                                class="py-3 pr-4 text-sm font-semibold text-slate-500 dark:text-slate-400 w-40 align-top"
                                                >Dosen Pembimbing 1</td
                                            >
                                            <td
                                                class="py-3 text-sm text-slate-900 dark:text-white"
                                                >{doc.dosen_pembimbing}</td
                                            >
                                        </tr>
                                    {/if}
                                    {#if doc.dosen_pembimbing_2}
                                        <tr
                                            class="border-b border-slate-100 dark:border-slate-700/50"
                                        >
                                            <td
                                                class="py-3 pr-4 text-sm font-semibold text-slate-500 dark:text-slate-400 w-40 align-top"
                                                >Dosen Pembimbing 2</td
                                            >
                                            <td
                                                class="py-3 text-sm text-slate-900 dark:text-white"
                                                >{doc.dosen_pembimbing_2}</td
                                            >
                                        </tr>
                                    {/if}
                                    {#if doc.kata_kunci}
                                        <tr
                                            class="border-b border-slate-100 dark:border-slate-700/50"
                                        >
                                            <td
                                                class="py-3 pr-4 text-sm font-semibold text-slate-500 dark:text-slate-400 w-40 align-top"
                                                >Kata Kunci</td
                                            >
                                            <td class="py-3">
                                                <div
                                                    class="flex flex-wrap gap-1.5"
                                                >
                                                    {#each doc.kata_kunci
                                                        .split(",")
                                                        .map((k) => k.trim())
                                                        .filter((k) => k) as keyword}
                                                        <span
                                                            class="inline-flex items-center gap-1 px-2.5 py-1 bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400 text-xs font-semibold rounded-full border border-blue-100 dark:border-blue-800/30"
                                                        >
                                                            <span
                                                                class="material-symbols-outlined"
                                                                style="font-size: 12px;"
                                                                >tag</span
                                                            >
                                                            {keyword}
                                                        </span>
                                                    {/each}
                                                </div>
                                            </td>
                                        </tr>
                                    {/if}
                                    <tr>
                                        <td
                                            class="py-3 pr-4 text-sm font-semibold text-slate-500 dark:text-slate-400 w-40 align-top"
                                            >Tanggal Upload</td
                                        >
                                        <td
                                            class="py-3 text-sm text-slate-900 dark:text-white"
                                            >{formatDate(doc.created_at)}</td
                                        >
                                    </tr>
                                </tbody>
                            </table>
                        </div>
                    </div>

                    <!-- Abstrak / Abstract Section -->
                    {#if doc.abstrak}
                        <div
                            class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden shadow-sm"
                        >
                            <div
                                class="px-6 py-4 border-b border-slate-100 dark:border-slate-700"
                            >
                                <h2
                                    class="text-lg font-bold text-slate-900 dark:text-white flex items-center gap-2"
                                >
                                    <span
                                        class="material-symbols-outlined text-primary"
                                        >subject</span
                                    >
                                    Abstract
                                </h2>
                            </div>
                            <div class="p-6">
                                <div
                                    class="prose prose-slate dark:prose-invert max-w-none"
                                >
                                    <p
                                        class="text-sm text-slate-700 dark:text-slate-300 leading-relaxed whitespace-pre-line"
                                    >
                                        {doc.abstrak}
                                    </p>
                                </div>
                            </div>
                        </div>
                    {/if}

                    <!-- Sitasi / Citation Section -->
                    <div
                        class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden shadow-sm"
                    >
                        <div
                            class="px-6 py-4 border-b border-slate-100 dark:border-slate-700"
                        >
                            <h2
                                class="text-lg font-bold text-slate-900 dark:text-white flex items-center gap-2"
                            >
                                <span
                                    class="material-symbols-outlined text-primary"
                                    >format_quote</span
                                >
                                Sitasi
                            </h2>
                        </div>
                        <div class="p-6">
                            <div
                                class="bg-slate-50 dark:bg-slate-900/50 rounded-lg p-4 border border-slate-200 dark:border-slate-700"
                            >
                                <p
                                    class="text-sm text-slate-700 dark:text-slate-300 italic leading-relaxed"
                                >
                                    {doc.penulis} ({doc.tahun && doc.tahun > 0 ? doc.tahun : new Date(
                                        doc.created_at,
                                    ).getFullYear()})
                                    <strong>{doc.judul}</strong>.
                                    {#if doc.jenis_file}
                                        {doc.jenis_file
                                            .charAt(0)
                                            .toUpperCase() +
                                            doc.jenis_file.slice(1)},
                                    {/if}
                                    {#if doc.fakultas_nama}
                                        {doc.fakultas_nama}{#if doc.prodi_nama}, {doc.prodi_nama}{/if}.
                                    {/if}
                                </p>
                            </div>
                            <button
                                on:click={() => {
                                    const citationText = `${doc.penulis} (${doc.tahun && doc.tahun > 0 ? doc.tahun : new Date(doc.created_at).getFullYear()}) ${doc.judul}. ${doc.jenis_file ? doc.jenis_file.charAt(0).toUpperCase() + doc.jenis_file.slice(1) + ", " : ""}${doc.fakultas_nama ? doc.fakultas_nama : ""}${doc.prodi_nama ? ", " + doc.prodi_nama : ""}.`;
                                    navigator.clipboard.writeText(citationText);
                                    alert("Sitasi berhasil disalin!");
                                }}
                                class="mt-3 inline-flex items-center gap-1.5 px-3 py-1.5 bg-primary/10 text-primary hover:bg-primary hover:text-white rounded-lg text-xs font-semibold transition-all"
                            >
                                <span
                                    class="material-symbols-outlined"
                                    style="font-size: 16px;">content_copy</span
                                >
                                Salin Sitasi
                            </button>
                        </div>
                    </div>

                    <!-- Files / Daftar File Card -->
                    {#if doc.files && doc.files.length > 0}
                        <div
                            class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden shadow-sm"
                        >
                            <div
                                class="px-6 py-4 border-b border-slate-100 dark:border-slate-700 flex items-center justify-between"
                            >
                                <h2
                                    class="text-lg font-bold text-slate-900 dark:text-white flex items-center gap-2"
                                >
                                    <span
                                        class="material-symbols-outlined text-primary"
                                        >folder_open</span
                                    >
                                    Daftar File
                                    <span
                                        class="text-xs font-normal text-slate-400 ml-1"
                                        >({doc.files.length} file)</span
                                    >
                                </h2>
                            </div>
                            <div
                                class="divide-y divide-slate-100 dark:divide-slate-700/50"
                            >
                                {#each doc.files as file, index}
                                    <div
                                        class="px-6 py-4 hover:bg-slate-50 dark:hover:bg-slate-700/30 transition-colors group"
                                    >
                                        <div class="flex items-center gap-4">
                                            <!-- File icon -->
                                            <div
                                                class="w-10 h-10 rounded-lg {file.is_locked &&
                                                !isMahasiswa &&
                                                !unlockedFiles[file.id]
                                                    ? 'bg-amber-100 dark:bg-amber-900/30'
                                                    : file.is_locked &&
                                                        (isMahasiswa ||
                                                            unlockedFiles[
                                                                file.id
                                                            ])
                                                      ? 'bg-green-100 dark:bg-green-900/30'
                                                      : 'bg-red-100 dark:bg-red-900/30'} flex items-center justify-center flex-shrink-0"
                                            >
                                                {#if file.is_locked && !isMahasiswa && !unlockedFiles[file.id]}
                                                    <span
                                                        class="material-symbols-outlined text-amber-500"
                                                        style="font-size: 22px;"
                                                        >lock</span
                                                    >
                                                {:else if file.is_locked && (isMahasiswa || unlockedFiles[file.id])}
                                                    <span
                                                        class="material-symbols-outlined text-green-500"
                                                        style="font-size: 22px;"
                                                        >lock_open</span
                                                    >
                                                {:else}
                                                    <span
                                                        class="material-symbols-outlined text-red-500"
                                                        style="font-size: 22px;"
                                                        >{getFileIcon(
                                                            file.file_name,
                                                        )}</span
                                                    >
                                                {/if}
                                            </div>

                                            <!-- File info -->
                                            <div class="flex-1 min-w-0">
                                                <div
                                                    class="flex items-center gap-2"
                                                >
                                                    <p
                                                        class="text-sm font-semibold text-slate-900 dark:text-white truncate"
                                                    >
                                                        {file.file_name}
                                                    </p>
                                                    {#if file.is_locked && !unlockedFiles[file.id]}
                                                        <span
                                                            class="inline-flex items-center gap-0.5 px-1.5 py-0.5 bg-amber-100 dark:bg-amber-900/30 text-amber-600 dark:text-amber-400 text-[10px] font-bold rounded uppercase"
                                                        >
                                                            <span
                                                                class="material-symbols-outlined"
                                                                style="font-size: 11px;"
                                                                >lock</span
                                                            >
                                                            Terkunci
                                                        </span>
                                                    {:else if file.is_locked && unlockedFiles[file.id]}
                                                        <span
                                                            class="inline-flex items-center gap-0.5 px-1.5 py-0.5 bg-green-100 dark:bg-green-900/30 text-green-600 dark:text-green-400 text-[10px] font-bold rounded uppercase"
                                                        >
                                                            <span
                                                                class="material-symbols-outlined"
                                                                style="font-size: 11px;"
                                                                >lock_open</span
                                                            >
                                                            Terbuka
                                                        </span>
                                                    {/if}
                                                </div>
                                                <p
                                                    class="text-xs text-slate-400 dark:text-slate-500 mt-0.5"
                                                >
                                                    {formatFileSize(
                                                        file.file_size,
                                                    )}
                                                </p>
                                            </div>

                                            <!-- Actions -->
                                            <div
                                                class="flex items-center gap-2 flex-shrink-0"
                                            >
                                                {#if file.is_locked && !unlockedFiles[file.id]}
                                                    <!-- File masih terkunci -->
                                                    <span
                                                        class="inline-flex items-center gap-1 px-2 py-1 bg-amber-100 dark:bg-amber-900/30 text-amber-600 dark:text-amber-400 text-[10px] font-bold rounded-lg"
                                                    >
                                                        <span
                                                            class="material-symbols-outlined"
                                                            style="font-size: 13px;"
                                                            >lock</span
                                                        >
                                                        Terkunci
                                                    </span>
                                                {:else if file.is_locked && unlockedFiles[file.id]}
                                                    <!-- File terkunci tapi sudah di-unlock -->
                                                    <button
                                                        on:click={() =>
                                                            handleUnlockedPreview(
                                                                file.id,
                                                            )}
                                                        class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300 hover:bg-primary/10 hover:text-primary rounded-lg text-xs font-semibold transition-all"
                                                        title="Preview file"
                                                    >
                                                        <span
                                                            class="material-symbols-outlined"
                                                            style="font-size: 16px;"
                                                            >visibility</span
                                                        >
                                                        <span
                                                            class="hidden sm:inline"
                                                            >Preview</span
                                                        >
                                                    </button>
                                                    <button
                                                        on:click={() =>
                                                            handleUnlockedDownload(
                                                                file.id,
                                                            )}
                                                        class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-primary/10 text-primary hover:bg-primary hover:text-white rounded-lg text-xs font-semibold transition-all"
                                                        title="Download file"
                                                    >
                                                        <span
                                                            class="material-symbols-outlined"
                                                            style="font-size: 16px;"
                                                            >download</span
                                                        >
                                                        <span
                                                            class="hidden sm:inline"
                                                            >Download</span
                                                        >
                                                    </button>
                                                {:else}
                                                    <!-- File tidak terkunci -->
                                                    <button
                                                        on:click={() =>
                                                            handleFilePreview(
                                                                file,
                                                            )}
                                                        class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300 hover:bg-primary/10 hover:text-primary rounded-lg text-xs font-semibold transition-all"
                                                        title="Preview file"
                                                    >
                                                        <span
                                                            class="material-symbols-outlined"
                                                            style="font-size: 16px;"
                                                            >visibility</span
                                                        >
                                                        <span
                                                            class="hidden sm:inline"
                                                            >Preview</span
                                                        >
                                                    </button>
                                                    <button
                                                        on:click={() =>
                                                            handleFileDownload(
                                                                file,
                                                            )}
                                                        class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-primary/10 text-primary hover:bg-primary hover:text-white rounded-lg text-xs font-semibold transition-all"
                                                        title="Download file"
                                                    >
                                                        <span
                                                            class="material-symbols-outlined"
                                                            style="font-size: 16px;"
                                                            >download</span
                                                        >
                                                        <span
                                                            class="hidden sm:inline"
                                                            >Download</span
                                                        >
                                                    </button>
                                                {/if}
                                            </div>
                                        </div>
                                    </div>
                                {/each}
                            </div>

                            <!-- Document-level Access Section (muncul di bawah daftar file) -->
                            {#if hasLockedFiles && !allLockedUnlocked}
                                <div
                                    class="px-6 py-4 bg-amber-50 dark:bg-amber-900/10 border-t border-amber-200/50 dark:border-amber-800/30"
                                >
                                    {#if !isLoggedIn}
                                        <!-- Belum login: arahkan untuk mendaftar dulu -->
                                        <div class="flex items-start gap-3">
                                            <span
                                                class="material-symbols-outlined text-amber-500 flex-shrink-0 mt-0.5"
                                                style="font-size: 20px;"
                                                >lock</span
                                            >
                                            <div>
                                                <p
                                                    class="text-sm font-semibold text-slate-900 dark:text-white mb-1"
                                                >
                                                    File Terkunci
                                                </p>
                                                <p
                                                    class="text-xs text-slate-600 dark:text-slate-400 mb-3"
                                                >
                                                    Beberapa file pada dokumen
                                                    ini dikunci. Silakan daftar
                                                    atau login terlebih dahulu
                                                    untuk dapat meminta akses.
                                                </p>
                                                <div
                                                    class="flex items-center gap-2"
                                                >
                                                    <a
                                                        href="#/student-signup"
                                                        class="inline-flex items-center gap-1.5 px-4 py-2 bg-gradient-to-r from-primary to-blue-600 hover:from-primary/90 hover:to-blue-700 text-white rounded-lg text-xs font-bold shadow-sm transition-all"
                                                    >
                                                        <span
                                                            class="material-symbols-outlined"
                                                            style="font-size: 16px;"
                                                            >person_add</span
                                                        >
                                                        Daftar Akun
                                                    </a>
                                                    <a
                                                        href="#/login"
                                                        class="inline-flex items-center gap-1.5 px-4 py-2 bg-slate-200 dark:bg-slate-700 hover:bg-slate-300 dark:hover:bg-slate-600 text-slate-700 dark:text-slate-300 rounded-lg text-xs font-bold transition-all"
                                                    >
                                                        <span
                                                            class="material-symbols-outlined"
                                                            style="font-size: 16px;"
                                                            >login</span
                                                        >
                                                        Login
                                                    </a>
                                                </div>
                                            </div>
                                        </div>
                                    {:else if tokenLoading}
                                        <div
                                            class="flex items-center gap-3 justify-center py-2"
                                        >
                                            <span
                                                class="material-symbols-outlined animate-spin text-primary"
                                                style="font-size: 20px;"
                                                >progress_activity</span
                                            >
                                            <p
                                                class="text-sm text-slate-600 dark:text-slate-300 font-medium"
                                            >
                                                Memverifikasi token akses...
                                            </p>
                                        </div>
                                    {:else if tokenSuccess}
                                        <div
                                            class="flex items-start gap-2 p-3 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800/30 rounded-lg"
                                        >
                                            <span
                                                class="material-symbols-outlined text-green-500 flex-shrink-0"
                                                style="font-size: 16px;"
                                                >check_circle</span
                                            >
                                            <p
                                                class="text-xs text-green-700 dark:text-green-400 font-medium"
                                            >
                                                Token valid! Semua file terkunci
                                                pada dokumen ini telah dibuka.
                                            </p>
                                        </div>
                                    {:else if tokenError && !showAccessForm}
                                        <div
                                            class="flex items-start gap-2 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800/30 rounded-lg mb-3"
                                        >
                                            <span
                                                class="material-symbols-outlined text-red-500 flex-shrink-0"
                                                style="font-size: 16px;"
                                                >error</span
                                            >
                                            <p
                                                class="text-xs text-red-600 dark:text-red-400"
                                            >
                                                {tokenError}
                                            </p>
                                        </div>
                                        <button
                                            on:click={openAccessForm}
                                            class="inline-flex items-center gap-1.5 px-4 py-2 bg-gradient-to-r from-amber-500 to-orange-500 hover:from-amber-600 hover:to-orange-600 text-white rounded-lg text-xs font-bold shadow-sm transition-all"
                                        >
                                            <span
                                                class="material-symbols-outlined"
                                                style="font-size: 16px;"
                                                >key</span
                                            >
                                            Minta Akses Dokumen
                                        </button>
                                    {:else if !showAccessForm}
                                        <p
                                            class="text-xs text-amber-600 dark:text-amber-400 flex items-center gap-1.5 mb-3"
                                        >
                                            <span
                                                class="material-symbols-outlined"
                                                style="font-size: 14px;"
                                                >info</span
                                            >
                                            Beberapa file dikunci. Minta akses untuk
                                            membuka semua file terkunci pada dokumen
                                            ini. Link akses akan dikirim ke email
                                            Anda.
                                        </p>
                                        <button
                                            on:click={openAccessForm}
                                            class="inline-flex items-center gap-1.5 px-4 py-2 bg-gradient-to-r from-amber-500 to-orange-500 hover:from-amber-600 hover:to-orange-600 text-white rounded-lg text-xs font-bold shadow-sm transition-all"
                                        >
                                            <span
                                                class="material-symbols-outlined"
                                                style="font-size: 16px;"
                                                >key</span
                                            >
                                            Minta Akses Dokumen
                                        </button>
                                    {/if}

                                    <!-- Access Request Form - 3 Step OTP Flow -->
                                    {#if showAccessForm}
                                        <div
                                            class="mt-4 p-5 bg-gradient-to-br from-amber-50 to-orange-50 dark:from-amber-900/10 dark:to-orange-900/10 rounded-xl border border-amber-200/60 dark:border-amber-800/30 animate-slideDown"
                                        >
                                            <div
                                                class="flex items-center justify-between mb-4"
                                            >
                                                <h4
                                                    class="text-sm font-bold text-slate-900 dark:text-white flex items-center gap-2"
                                                >
                                                    <span
                                                        class="material-symbols-outlined text-amber-500"
                                                        style="font-size: 18px;"
                                                        >key</span
                                                    >
                                                    Permintaan Akses Dokumen
                                                </h4>
                                                <button
                                                    on:click={closeAccessForm}
                                                    class="p-1 hover:bg-amber-200/50 dark:hover:bg-amber-900/30 rounded-lg transition-colors"
                                                    title="Tutup formulir"
                                                >
                                                    <span
                                                        class="material-symbols-outlined text-slate-400"
                                                        style="font-size: 18px;"
                                                        >close</span
                                                    >
                                                </button>
                                            </div>

                                            <!-- Step Indicator -->
                                            <div
                                                class="flex items-center gap-2 mb-5"
                                            >
                                                <div
                                                    class="flex items-center gap-1.5"
                                                >
                                                    <div
                                                        class="w-6 h-6 rounded-full flex items-center justify-center text-[10px] font-bold {otpStep >=
                                                        1
                                                            ? 'bg-amber-500 text-white'
                                                            : 'bg-slate-200 dark:bg-slate-600 text-slate-400'}"
                                                    >
                                                        {#if otpStep > 1}
                                                            <span
                                                                class="material-symbols-outlined"
                                                                style="font-size: 14px;"
                                                                >check</span
                                                            >
                                                        {:else}
                                                            1
                                                        {/if}
                                                    </div>
                                                    <span
                                                        class="text-[10px] font-semibold {otpStep >=
                                                        1
                                                            ? 'text-amber-600 dark:text-amber-400'
                                                            : 'text-slate-400'}"
                                                        >Email</span
                                                    >
                                                </div>
                                                <div
                                                    class="flex-1 h-0.5 {otpStep >=
                                                    2
                                                        ? 'bg-amber-400'
                                                        : 'bg-slate-200 dark:bg-slate-600'} rounded-full"
                                                ></div>
                                                <div
                                                    class="flex items-center gap-1.5"
                                                >
                                                    <div
                                                        class="w-6 h-6 rounded-full flex items-center justify-center text-[10px] font-bold {otpStep >=
                                                        2
                                                            ? 'bg-amber-500 text-white'
                                                            : 'bg-slate-200 dark:bg-slate-600 text-slate-400'}"
                                                    >
                                                        {#if otpStep > 2}
                                                            <span
                                                                class="material-symbols-outlined"
                                                                style="font-size: 14px;"
                                                                >check</span
                                                            >
                                                        {:else}
                                                            2
                                                        {/if}
                                                    </div>
                                                    <span
                                                        class="text-[10px] font-semibold {otpStep >=
                                                        2
                                                            ? 'text-amber-600 dark:text-amber-400'
                                                            : 'text-slate-400'}"
                                                        >OTP</span
                                                    >
                                                </div>
                                                <div
                                                    class="flex-1 h-0.5 {otpStep >=
                                                    3
                                                        ? 'bg-amber-400'
                                                        : 'bg-slate-200 dark:bg-slate-600'} rounded-full"
                                                ></div>
                                                <div
                                                    class="flex items-center gap-1.5"
                                                >
                                                    <div
                                                        class="w-6 h-6 rounded-full flex items-center justify-center text-[10px] font-bold {otpStep >=
                                                        3
                                                            ? 'bg-amber-500 text-white'
                                                            : 'bg-slate-200 dark:bg-slate-600 text-slate-400'}"
                                                    >
                                                        3
                                                    </div>
                                                    <span
                                                        class="text-[10px] font-semibold {otpStep >=
                                                        3
                                                            ? 'text-amber-600 dark:text-amber-400'
                                                            : 'text-slate-400'}"
                                                        >Data</span
                                                    >
                                                </div>
                                            </div>

                                            {#if accessFormSuccess}
                                                <div
                                                    class="flex items-start gap-3 p-4 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800/30 rounded-lg"
                                                >
                                                    <span
                                                        class="material-symbols-outlined text-green-500 flex-shrink-0"
                                                        style="font-size: 20px;"
                                                        >check_circle</span
                                                    >
                                                    <div>
                                                        <p
                                                            class="text-sm font-medium text-green-800 dark:text-green-300"
                                                        >
                                                            {accessFormSuccess}
                                                        </p>
                                                        <button
                                                            on:click={closeAccessForm}
                                                            class="mt-2 text-xs font-semibold text-green-600 hover:text-green-700 underline"
                                                        >
                                                            Tutup
                                                        </button>
                                                    </div>
                                                </div>
                                            {:else}
                                                <div class="space-y-3">
                                                    <!-- ===== STEP 1: Email Input ===== -->
                                                    {#if otpStep === 1}
                                                        <div
                                                            class="animate-slideDown"
                                                        >
                                                            <p
                                                                class="text-xs text-slate-500 dark:text-slate-400 mb-3"
                                                            >
                                                                Masukkan email
                                                                Anda untuk
                                                                verifikasi. Kode
                                                                OTP akan dikirim
                                                                ke email Anda.
                                                            </p>
                                                            <div>
                                                                <label
                                                                    for="otp-email"
                                                                    class="block text-xs font-semibold text-slate-600 dark:text-slate-400 mb-1"
                                                                >
                                                                    Alamat Email <span
                                                                        class="text-red-500"
                                                                        >*</span
                                                                    >
                                                                </label>
                                                                <input
                                                                    id="otp-email"
                                                                    type="email"
                                                                    bind:value={
                                                                        otpEmail
                                                                    }
                                                                    placeholder={maskedEmail ||
                                                                        "contoh@email.com"}
                                                                    class="w-full px-3 py-2.5 text-sm bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-600 rounded-lg focus:ring-2 focus:ring-amber-400 focus:border-amber-400 text-slate-900 dark:text-white placeholder-slate-400 transition-all"
                                                                    on:keydown={(
                                                                        e,
                                                                    ) =>
                                                                        e.key ===
                                                                            "Enter" &&
                                                                        sendOTP()}
                                                                />
                                                                {#if isMahasiswa && maskedEmail}
                                                                    <p
                                                                        class="mt-1 text-[11px] text-amber-600 dark:text-amber-400 flex items-center gap-1"
                                                                    >
                                                                        <span
                                                                            class="material-symbols-outlined"
                                                                            style="font-size: 13px;"
                                                                            >info</span
                                                                        >
                                                                        Masukkan
                                                                        email yang
                                                                        terdaftar
                                                                        di akun Anda
                                                                    </p>
                                                                {/if}
                                                            </div>

                                                            <!-- OTP Error -->
                                                            {#if otpError}
                                                                <div
                                                                    class="flex items-start gap-2 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800/30 rounded-lg"
                                                                >
                                                                    <span
                                                                        class="material-symbols-outlined text-red-500 flex-shrink-0"
                                                                        style="font-size: 16px;"
                                                                        >error</span
                                                                    >
                                                                    <p
                                                                        class="text-xs text-red-600 dark:text-red-400"
                                                                    >
                                                                        {otpError}
                                                                    </p>
                                                                </div>
                                                            {/if}

                                                            <div
                                                                class="flex items-center gap-2 pt-1"
                                                            >
                                                                <button
                                                                    on:click={sendOTP}
                                                                    disabled={otpLoading}
                                                                    class="inline-flex items-center gap-1.5 px-4 py-2.5 bg-gradient-to-r from-blue-500 to-indigo-500 hover:from-blue-600 hover:to-indigo-600 text-white rounded-lg text-xs font-bold shadow-sm hover:shadow-md transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                                                                >
                                                                    {#if otpLoading}
                                                                        <span
                                                                            class="material-symbols-outlined animate-spin"
                                                                            style="font-size: 16px;"
                                                                            >progress_activity</span
                                                                        >
                                                                        Mengirim...
                                                                    {:else}
                                                                        <span
                                                                            class="material-symbols-outlined"
                                                                            style="font-size: 16px;"
                                                                            >mail</span
                                                                        >
                                                                        Kirim Kode
                                                                        OTP
                                                                    {/if}
                                                                </button>
                                                                <button
                                                                    on:click={closeAccessForm}
                                                                    class="inline-flex items-center gap-1 px-3 py-2.5 bg-slate-100 dark:bg-slate-700 text-slate-500 dark:text-slate-400 hover:bg-slate-200 dark:hover:bg-slate-600 rounded-lg text-xs font-semibold transition-all"
                                                                >
                                                                    Batal
                                                                </button>
                                                            </div>
                                                        </div>
                                                    {/if}

                                                    <!-- ===== STEP 2: OTP Input ===== -->
                                                    {#if otpStep === 2}
                                                        <div
                                                            class="animate-slideDown"
                                                        >
                                                            <!-- Email verified indicator -->
                                                            <div
                                                                class="flex items-center gap-2 p-2.5 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800/30 rounded-lg mb-3"
                                                            >
                                                                <span
                                                                    class="material-symbols-outlined text-blue-500"
                                                                    style="font-size: 16px;"
                                                                    >mail</span
                                                                >
                                                                <span
                                                                    class="text-xs text-blue-700 dark:text-blue-300 font-medium truncate"
                                                                    >{otpEmail}</span
                                                                >
                                                                <button
                                                                    on:click={() => {
                                                                        otpStep = 1;
                                                                        otpError =
                                                                            "";
                                                                        otpSuccess =
                                                                            "";
                                                                        otpCode =
                                                                            "";
                                                                    }}
                                                                    class="ml-auto text-[10px] text-blue-500 hover:text-blue-700 font-semibold underline flex-shrink-0"
                                                                >
                                                                    Ubah
                                                                </button>
                                                            </div>

                                                            {#if otpSuccess}
                                                                <div
                                                                    class="flex items-start gap-2 p-3 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800/30 rounded-lg mb-3"
                                                                >
                                                                    <span
                                                                        class="material-symbols-outlined text-green-500 flex-shrink-0"
                                                                        style="font-size: 16px;"
                                                                        >check_circle</span
                                                                    >
                                                                    <p
                                                                        class="text-xs text-green-700 dark:text-green-400 font-medium"
                                                                    >
                                                                        {otpSuccess}
                                                                    </p>
                                                                </div>
                                                            {/if}

                                                            <div>
                                                                <label
                                                                    for="otp-code"
                                                                    class="block text-xs font-semibold text-slate-600 dark:text-slate-400 mb-1"
                                                                >
                                                                    Masukkan
                                                                    Kode OTP <span
                                                                        class="text-red-500"
                                                                        >*</span
                                                                    >
                                                                </label>
                                                                <input
                                                                    id="otp-code"
                                                                    type="text"
                                                                    bind:value={
                                                                        otpCode
                                                                    }
                                                                    placeholder="Masukkan 6 digit kode"
                                                                    maxlength="6"
                                                                    class="w-full px-3 py-2.5 text-center text-lg font-mono font-bold tracking-[6px] bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-600 rounded-lg focus:ring-2 focus:ring-amber-400 focus:border-amber-400 text-slate-900 dark:text-white placeholder-slate-400 placeholder:text-sm placeholder:tracking-normal placeholder:font-normal transition-all"
                                                                    on:keydown={(
                                                                        e,
                                                                    ) =>
                                                                        e.key ===
                                                                            "Enter" &&
                                                                        verifyOTP()}
                                                                />
                                                                <p
                                                                    class="mt-1.5 text-[11px] text-slate-400 dark:text-slate-500"
                                                                >
                                                                    Kode berlaku
                                                                    selama 5
                                                                    menit.
                                                                    {#if otpCountdown > 0}
                                                                        Kirim
                                                                        ulang
                                                                        dalam <span
                                                                            class="font-bold text-amber-500"
                                                                            >{otpCountdown}s</span
                                                                        >
                                                                    {:else}
                                                                        <button
                                                                            on:click={sendOTP}
                                                                            disabled={otpLoading}
                                                                            class="text-blue-500 hover:text-blue-700 font-semibold underline disabled:opacity-50"
                                                                        >
                                                                            Kirim
                                                                            ulang
                                                                            OTP
                                                                        </button>
                                                                    {/if}
                                                                </p>
                                                            </div>

                                                            <!-- OTP Error -->
                                                            {#if otpError}
                                                                <div
                                                                    class="flex items-start gap-2 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800/30 rounded-lg"
                                                                >
                                                                    <span
                                                                        class="material-symbols-outlined text-red-500 flex-shrink-0"
                                                                        style="font-size: 16px;"
                                                                        >error</span
                                                                    >
                                                                    <p
                                                                        class="text-xs text-red-600 dark:text-red-400"
                                                                    >
                                                                        {otpError}
                                                                    </p>
                                                                </div>
                                                            {/if}

                                                            <div
                                                                class="flex items-center gap-2 pt-1"
                                                            >
                                                                <button
                                                                    on:click={verifyOTP}
                                                                    disabled={otpLoading}
                                                                    class="inline-flex items-center gap-1.5 px-4 py-2.5 bg-gradient-to-r from-green-500 to-emerald-500 hover:from-green-600 hover:to-emerald-600 text-white rounded-lg text-xs font-bold shadow-sm hover:shadow-md transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                                                                >
                                                                    {#if otpLoading}
                                                                        <span
                                                                            class="material-symbols-outlined animate-spin"
                                                                            style="font-size: 16px;"
                                                                            >progress_activity</span
                                                                        >
                                                                        Memverifikasi...
                                                                    {:else}
                                                                        <span
                                                                            class="material-symbols-outlined"
                                                                            style="font-size: 16px;"
                                                                            >verified</span
                                                                        >
                                                                        Verifikasi
                                                                        OTP
                                                                    {/if}
                                                                </button>
                                                                <button
                                                                    on:click={closeAccessForm}
                                                                    class="inline-flex items-center gap-1 px-3 py-2.5 bg-slate-100 dark:bg-slate-700 text-slate-500 dark:text-slate-400 hover:bg-slate-200 dark:hover:bg-slate-600 rounded-lg text-xs font-semibold transition-all"
                                                                >
                                                                    Batal
                                                                </button>
                                                            </div>
                                                        </div>
                                                    {/if}

                                                    <!-- ===== STEP 3: Form Lengkap (Nama + KTM) ===== -->
                                                    {#if otpStep === 3}
                                                        <div
                                                            class="animate-slideDown"
                                                        >
                                                            <!-- Email verified badge -->
                                                            <div
                                                                class="flex items-center gap-2 p-2.5 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800/30 rounded-lg mb-3"
                                                            >
                                                                <span
                                                                    class="material-symbols-outlined text-green-500"
                                                                    style="font-size: 16px;"
                                                                    >verified</span
                                                                >
                                                                <span
                                                                    class="text-xs text-green-700 dark:text-green-300 font-medium"
                                                                    >{otpEmail}</span
                                                                >
                                                                <span
                                                                    class="ml-auto inline-flex items-center gap-0.5 px-1.5 py-0.5 bg-green-100 dark:bg-green-800/30 text-green-600 dark:text-green-400 text-[10px] font-bold rounded uppercase"
                                                                >
                                                                    <span
                                                                        class="material-symbols-outlined"
                                                                        style="font-size: 11px;"
                                                                        >check</span
                                                                    >
                                                                    Terverifikasi
                                                                </span>
                                                            </div>

                                                            {#if otpSuccess}
                                                                <div
                                                                    class="flex items-start gap-2 p-3 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800/30 rounded-lg mb-3"
                                                                >
                                                                    <span
                                                                        class="material-symbols-outlined text-green-500 flex-shrink-0"
                                                                        style="font-size: 16px;"
                                                                        >check_circle</span
                                                                    >
                                                                    <p
                                                                        class="text-xs text-green-700 dark:text-green-400 font-medium"
                                                                    >
                                                                        {otpSuccess}
                                                                    </p>
                                                                </div>
                                                            {/if}

                                                            <!-- Nama -->
                                                            <div>
                                                                <label
                                                                    for="access-nama-doc"
                                                                    class="block text-xs font-semibold text-slate-600 dark:text-slate-400 mb-1"
                                                                >
                                                                    Nama Lengkap <span
                                                                        class="text-red-500"
                                                                        >*</span
                                                                    >
                                                                </label>
                                                                <input
                                                                    id="access-nama-doc"
                                                                    type="text"
                                                                    bind:value={
                                                                        accessFormData.nama
                                                                    }
                                                                    placeholder="Masukkan nama lengkap Anda"
                                                                    class="w-full px-3 py-2 text-sm bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-600 rounded-lg focus:ring-2 focus:ring-amber-400 focus:border-amber-400 text-slate-900 dark:text-white placeholder-slate-400 transition-all"
                                                                />
                                                            </div>

                                                            <!-- Upload KTM -->
                                                            <div>
                                                                <label
                                                                    for="access-ktm-doc"
                                                                    class="block text-xs font-semibold text-slate-600 dark:text-slate-400 mb-1"
                                                                >
                                                                    Upload KTM
                                                                    (Kartu Tanda
                                                                    Mahasiswa) <span
                                                                        class="text-red-500"
                                                                        >*</span
                                                                    >
                                                                </label>
                                                                <div
                                                                    class="relative"
                                                                >
                                                                    <input
                                                                        id="access-ktm-doc"
                                                                        type="file"
                                                                        accept="image/*,.pdf"
                                                                        on:change={handleKtmUpload}
                                                                        class="w-full px-3 py-2 text-sm bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-600 rounded-lg focus:ring-2 focus:ring-amber-400 focus:border-amber-400 text-slate-900 dark:text-white file:mr-3 file:py-1 file:px-3 file:rounded-md file:border-0 file:text-xs file:font-semibold file:bg-amber-100 file:text-amber-600 hover:file:bg-amber-200 transition-all"
                                                                    />
                                                                </div>
                                                                {#if accessFormData.ktm}
                                                                    <p
                                                                        class="mt-1 text-xs text-green-600 flex items-center gap-1"
                                                                    >
                                                                        <span
                                                                            class="material-symbols-outlined"
                                                                            style="font-size: 14px;"
                                                                            >check_circle</span
                                                                        >
                                                                        {accessFormData
                                                                            .ktm
                                                                            .name}
                                                                    </p>
                                                                {/if}
                                                                <p
                                                                    class="mt-1 text-[11px] text-slate-400"
                                                                >
                                                                    Format: JPG,
                                                                    PNG, PDF.
                                                                    Maksimum 5
                                                                    MB.
                                                                </p>
                                                            </div>

                                                            <!-- Error Message -->
                                                            {#if accessFormError}
                                                                <div
                                                                    class="flex items-start gap-2 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800/30 rounded-lg"
                                                                >
                                                                    <span
                                                                        class="material-symbols-outlined text-red-500 flex-shrink-0"
                                                                        style="font-size: 16px;"
                                                                        >error</span
                                                                    >
                                                                    <p
                                                                        class="text-xs text-red-600 dark:text-red-400"
                                                                    >
                                                                        {accessFormError}
                                                                    </p>
                                                                </div>
                                                            {/if}

                                                            <!-- Buttons -->
                                                            <div
                                                                class="flex items-center gap-2 pt-1"
                                                            >
                                                                <button
                                                                    on:click={submitAccessRequest}
                                                                    disabled={accessFormLoading}
                                                                    class="inline-flex items-center gap-1.5 px-4 py-2 bg-gradient-to-r from-amber-500 to-orange-500 hover:from-amber-600 hover:to-orange-600 text-white rounded-lg text-xs font-bold shadow-sm hover:shadow-md transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                                                                >
                                                                    {#if accessFormLoading}
                                                                        <span
                                                                            class="material-symbols-outlined animate-spin"
                                                                            style="font-size: 16px;"
                                                                            >progress_activity</span
                                                                        >
                                                                        Mengirim...
                                                                    {:else}
                                                                        <span
                                                                            class="material-symbols-outlined"
                                                                            style="font-size: 16px;"
                                                                            >send</span
                                                                        >
                                                                        Kirim Permintaan
                                                                    {/if}
                                                                </button>
                                                                <button
                                                                    on:click={closeAccessForm}
                                                                    class="inline-flex items-center gap-1 px-3 py-2 bg-slate-100 dark:bg-slate-700 text-slate-500 dark:text-slate-400 hover:bg-slate-200 dark:hover:bg-slate-600 rounded-lg text-xs font-semibold transition-all"
                                                                >
                                                                    Batal
                                                                </button>
                                                            </div>
                                                        </div>
                                                    {/if}
                                                </div>
                                            {/if}
                                        </div>
                                    {/if}
                                </div>
                            {/if}
                        </div>
                    {/if}
                </div>

                <!-- Sidebar (Right Column) -->
                <div class="space-y-6">
                    <!-- Download All Card (hanya tampil jika sudah login) -->
                    {#if isLoggedIn}
                        <div
                            class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-6 shadow-sm"
                        >
                            <h3
                                class="text-sm font-bold text-slate-900 dark:text-white mb-4 flex items-center gap-2"
                            >
                                <span
                                    class="material-symbols-outlined text-primary"
                                    style="font-size: 18px;">download</span
                                >
                                Unduh Dokumen
                            </h3>
                            <button
                                on:click={handleDownload}
                                class="w-full flex items-center justify-center gap-2 px-4 py-3 bg-gradient-to-r {typeStyle.gradient} text-white font-bold rounded-lg shadow-lg hover:shadow-xl hover:opacity-90 transition-all active:scale-[0.98]"
                            >
                                <span class="material-symbols-outlined"
                                    >download</span
                                >
                                Download
                            </button>
                            {#if doc.files && doc.files.length > 1}
                                <p
                                    class="text-xs text-slate-400 mt-2 text-center"
                                >
                                    {doc.files.length} file akan diunduh dalam format
                                    ZIP
                                </p>
                            {/if}
                        </div>
                    {/if}

                    <!-- Metadata Card -->
                    <div
                        class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-6 shadow-sm"
                    >
                        <h3
                            class="text-sm font-bold text-slate-900 dark:text-white mb-4 flex items-center gap-2"
                        >
                            <span
                                class="material-symbols-outlined text-primary"
                                style="font-size: 18px;">label</span
                            >
                            Metadata
                        </h3>
                        <div class="space-y-3">
                            <div class="flex items-center gap-3 text-sm">
                                <span
                                    class="material-symbols-outlined text-slate-400"
                                    style="font-size: 18px;">category</span
                                >
                                <div>
                                    <p
                                        class="text-[11px] uppercase tracking-wider text-slate-400 font-semibold"
                                    >
                                        Jenis
                                    </p>
                                    <p
                                        class="text-slate-900 dark:text-white font-medium"
                                    >
                                        {doc.jenis_file}
                                    </p>
                                </div>
                            </div>
                            <div class="flex items-center gap-3 text-sm">
                                <span
                                    class="material-symbols-outlined text-slate-400"
                                    style="font-size: 18px;">badge</span
                                >
                                <div>
                                    <p
                                        class="text-[11px] uppercase tracking-wider text-slate-400 font-semibold"
                                    >
                                        Status
                                    </p>
                                    <span
                                        class="inline-block px-2 py-0.5 text-xs font-bold rounded-full {doc.status ===
                                        'publish'
                                            ? 'bg-green-100 dark:bg-green-900/30 text-green-600'
                                            : 'bg-slate-100 dark:bg-slate-700 text-slate-500'}"
                                    >
                                        {doc.status === "publish"
                                            ? "Published"
                                            : "Draft"}
                                    </span>
                                </div>
                            </div>
                            {#if doc.fakultas_nama}
                                <div class="flex items-center gap-3 text-sm">
                                    <span
                                        class="material-symbols-outlined text-slate-400"
                                        style="font-size: 18px;">apartment</span
                                    >
                                    <div>
                                        <p
                                            class="text-[11px] uppercase tracking-wider text-slate-400 font-semibold"
                                        >
                                            Fakultas
                                        </p>
                                        <p
                                            class="text-slate-900 dark:text-white font-medium"
                                        >
                                            {doc.fakultas_nama}
                                        </p>
                                    </div>
                                </div>
                            {/if}
                            {#if doc.prodi_nama}
                                <div class="flex items-center gap-3 text-sm">
                                    <span
                                        class="material-symbols-outlined text-slate-400"
                                        style="font-size: 18px;">school</span
                                    >
                                    <div>
                                        <p
                                            class="text-[11px] uppercase tracking-wider text-slate-400 font-semibold"
                                        >
                                            Program Studi
                                        </p>
                                        <p
                                            class="text-slate-900 dark:text-white font-medium"
                                        >
                                            {doc.prodi_nama}
                                        </p>
                                    </div>
                                </div>
                            {/if}

                            {#if doc.view_count > 0}
                                <div class="flex items-center gap-3 text-sm">
                                    <span
                                        class="material-symbols-outlined text-slate-400"
                                        style="font-size: 18px;"
                                        >visibility</span
                                    >
                                    <div>
                                        <p
                                            class="text-[11px] uppercase tracking-wider text-slate-400 font-semibold"
                                        >
                                            Total Dilihat
                                        </p>
                                        <p
                                            class="text-slate-900 dark:text-white font-medium"
                                        >
                                            {doc.view_count} kali
                                        </p>
                                    </div>
                                </div>
                            {/if}
                            {#if doc.kata_kunci}
                                <div class="flex items-start gap-3 text-sm">
                                    <span
                                        class="material-symbols-outlined text-slate-400 mt-0.5"
                                        style="font-size: 18px;">tag</span
                                    >
                                    <div>
                                        <p
                                            class="text-[11px] uppercase tracking-wider text-slate-400 font-semibold"
                                        >
                                            Kata Kunci
                                        </p>
                                        <div class="flex flex-wrap gap-1 mt-1">
                                            {#each doc.kata_kunci
                                                .split(",")
                                                .map((k) => k.trim())
                                                .filter((k) => k) as keyword}
                                                <span
                                                    class="inline-block px-2 py-0.5 text-[11px] font-medium bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400 rounded-full"
                                                >
                                                    {keyword}
                                                </span>
                                            {/each}
                                        </div>
                                    </div>
                                </div>
                            {/if}
                        </div>
                    </div>

                    <!-- Share Card -->
                    <div
                        class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-6 shadow-sm"
                    >
                        <h3
                            class="text-sm font-bold text-slate-900 dark:text-white mb-4 flex items-center gap-2"
                        >
                            <span
                                class="material-symbols-outlined text-primary"
                                style="font-size: 18px;">share</span
                            >
                            Bagikan
                        </h3>
                        <div class="flex items-center gap-2">
                            <input
                                type="text"
                                readonly
                                value={window.location.href}
                                class="flex-1 px-3 py-2 bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300 text-xs rounded-lg border-none truncate"
                            />
                            <button
                                on:click={() => {
                                    navigator.clipboard.writeText(
                                        window.location.href,
                                    );
                                    alert("Link berhasil disalin!");
                                }}
                                class="px-3 py-2 bg-primary/10 text-primary hover:bg-primary hover:text-white rounded-lg text-xs font-bold transition-all"
                                title="Salin link"
                            >
                                <span
                                    class="material-symbols-outlined"
                                    style="font-size: 16px;">content_copy</span
                                >
                            </button>
                        </div>
                    </div>

                    <!-- Back to Browse -->
                    <a
                        href="#/browse"
                        class="flex items-center justify-center gap-2 px-4 py-3 bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-600 rounded-xl font-medium text-sm transition-all"
                    >
                        <span
                            class="material-symbols-outlined"
                            style="font-size: 18px;">arrow_back</span
                        >
                        Kembali ke Jelajah
                    </a>
                </div>
            </div>
        </div>
    {/if}
</div>

<style>
    @keyframes slideDown {
        from {
            opacity: 0;
            transform: translateY(-8px);
            max-height: 0;
        }
        to {
            opacity: 1;
            transform: translateY(0);
            max-height: 500px;
        }
    }
    .animate-slideDown {
        animation: slideDown 0.3s ease-out forwards;
    }
</style>
