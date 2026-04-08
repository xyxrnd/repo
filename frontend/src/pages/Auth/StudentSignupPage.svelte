<script>
    import { API_ENDPOINTS } from "../../config";

    // Step state: 1 = email + OTP, 2 = verify OTP, 3 = fill form
    let step = 1;

    // OTP fields
    let otpEmail = "";
    let otpCode = "";
    let otpLoading = false;
    let otpError = "";
    let otpSuccess = "";
    let otpCountdown = 0;
    let otpTimer = null;

    // Form fields
    let name = "";
    let email = "";
    let password = "";
    let confirmPassword = "";
    let ktmFile = null;
    let ktmPreview = null;
    let loading = false;
    let error = "";
    let success = "";
    let showPassword = false;

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

        otpLoading = true;
        otpError = "";
        otpSuccess = "";
        try {
            const response = await fetch(API_ENDPOINTS.SEND_OTP, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    email: otpEmail.trim(),
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
            step = 2;
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
                    otp_code: otpCode.trim(),
                }),
            });
            if (!response.ok) {
                const text = await response.text();
                throw new Error(text || "Kode OTP tidak valid");
            }
            step = 3;
            email = otpEmail.trim();
            otpSuccess =
                "Email berhasil diverifikasi! Silakan lengkapi formulir pendaftaran.";
            if (otpTimer) clearInterval(otpTimer);
        } catch (e) {
            otpError = e.message || "Kode OTP tidak valid.";
        } finally {
            otpLoading = false;
        }
    }

    function handleKtmUpload(event) {
        const file = event.target.files[0];
        if (file) {
            // Validate file type
            const allowedTypes = [
                "image/jpeg",
                "image/png",
                "image/jpg",
                "application/pdf",
            ];
            if (!allowedTypes.includes(file.type)) {
                error = "Format file KTM harus JPG, PNG, atau PDF";
                event.target.value = "";
                return;
            }
            // Validate file size (max 5MB)
            if (file.size > 5 * 1024 * 1024) {
                error = "Ukuran file KTM maksimal 5MB";
                event.target.value = "";
                return;
            }
            ktmFile = file;
            error = "";
            // Create preview for images
            if (file.type.startsWith("image/")) {
                const reader = new FileReader();
                reader.onload = (e) => {
                    ktmPreview = e.target.result;
                };
                reader.readAsDataURL(file);
            } else {
                ktmPreview = null;
            }
        }
    }

    function removeKtm() {
        ktmFile = null;
        ktmPreview = null;
        // Reset file input
        const input = /** @type {HTMLInputElement} */ (
            document.getElementById("ktm-upload")
        );
        if (input) input.value = "";
    }

    async function handleSubmit() {
        error = "";
        success = "";

        // Validation
        if (!name.trim()) {
            error = "Nama lengkap wajib diisi";
            return;
        }
        if (!password) {
            error = "Password wajib diisi";
            return;
        }
        if (password.length < 6) {
            error = "Password minimal 6 karakter";
            return;
        }
        if (password !== confirmPassword) {
            error = "Konfirmasi password tidak cocok";
            return;
        }
        if (!ktmFile) {
            error = "Upload KTM wajib dilakukan";
            return;
        }

        loading = true;

        try {
            const formData = new FormData();
            formData.append("name", name.trim());
            formData.append("email", email.trim());
            formData.append("password", password);
            formData.append("ktm", ktmFile);

            const response = await fetch(API_ENDPOINTS.STUDENT_SIGNUP, {
                method: "POST",
                body: formData,
            });

            const text = await response.text();
            let data;
            try {
                data = JSON.parse(text);
            } catch {
                throw new Error(
                    !response.ok
                        ? `Server error (${response.status}): ${text.substring(0, 100)}`
                        : "Gagal memproses respons dari server",
                );
            }

            if (!response.ok) {
                throw new Error(data.error || "Gagal mendaftar");
            }
            success =
                data.message ||
                "Pendaftaran berhasil! Silakan tunggu verifikasi dari admin.";

            // Reset form
            name = "";
            email = "";
            password = "";
            confirmPassword = "";
            ktmFile = null;
            ktmPreview = null;
        } catch (err) {
            error = err.message;
        } finally {
            loading = false;
        }
    }
</script>

<section
    class="min-h-screen flex items-center justify-center px-4 py-12 bg-gradient-to-br from-background-dark via-surface-dark to-background-dark relative"
>
    <!-- Background decorations -->
    <div class="absolute inset-0 overflow-hidden">
        <div
            class="absolute top-0 left-1/4 w-96 h-96 bg-emerald-500/10 rounded-full blur-3xl"
        ></div>
        <div
            class="absolute bottom-0 right-1/4 w-96 h-96 bg-primary/10 rounded-full blur-3xl"
        ></div>
        <div
            class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] bg-blue-500/5 rounded-full blur-3xl"
        ></div>
    </div>

    <div class="relative w-full max-w-lg">
        <!-- Card -->
        <div
            class="bg-white dark:bg-surface-highlight border border-slate-200 dark:border-slate-700 rounded-2xl shadow-2xl overflow-hidden"
        >
            <!-- Header with gradient -->
            <div
                class="bg-gradient-to-r from-emerald-500 via-teal-500 to-primary p-6 text-center relative overflow-hidden"
            >
                <div class="absolute inset-0 opacity-10">
                    <div
                        class="absolute -top-12 -right-12 w-48 h-48 rounded-full bg-white/20 blur-xl"
                    ></div>
                    <div
                        class="absolute -bottom-12 -left-12 w-32 h-32 rounded-full bg-white/10 blur-lg"
                    ></div>
                </div>
                <div class="relative">
                    <div
                        class="w-16 h-16 mx-auto mb-3 bg-white/20 backdrop-blur-sm rounded-2xl flex items-center justify-center"
                    >
                        <span
                            class="material-symbols-outlined text-white text-3xl"
                            >school</span
                        >
                    </div>
                    <h1 class="text-2xl font-bold text-white">
                        Daftar Mahasiswa
                    </h1>
                    <p class="mt-1 text-white/80 text-sm">
                        Daftarkan diri Anda untuk mengakses dokumen akademik
                    </p>
                </div>
            </div>

            <div class="p-6 sm:p-8">
                <!-- Step Indicator -->
                {#if !success}
                    <div class="flex items-center gap-2 mb-6">
                        <div class="flex items-center gap-1.5">
                            <div
                                class="w-7 h-7 rounded-full flex items-center justify-center text-xs font-bold {step >=
                                1
                                    ? 'bg-emerald-500 text-white'
                                    : 'bg-slate-200 dark:bg-slate-600 text-slate-400'}"
                            >
                                {#if step > 1}
                                    <span
                                        class="material-symbols-outlined"
                                        style="font-size: 14px;">check</span
                                    >
                                {:else}
                                    1
                                {/if}
                            </div>
                            <span
                                class="text-xs font-semibold {step >= 1
                                    ? 'text-emerald-600 dark:text-emerald-400'
                                    : 'text-slate-400'}">Email</span
                            >
                        </div>
                        <div
                            class="flex-1 h-0.5 {step >= 2
                                ? 'bg-emerald-400'
                                : 'bg-slate-200 dark:bg-slate-600'} rounded-full"
                        ></div>
                        <div class="flex items-center gap-1.5">
                            <div
                                class="w-7 h-7 rounded-full flex items-center justify-center text-xs font-bold {step >=
                                2
                                    ? 'bg-emerald-500 text-white'
                                    : 'bg-slate-200 dark:bg-slate-600 text-slate-400'}"
                            >
                                {#if step > 2}
                                    <span
                                        class="material-symbols-outlined"
                                        style="font-size: 14px;">check</span
                                    >
                                {:else}
                                    2
                                {/if}
                            </div>
                            <span
                                class="text-xs font-semibold {step >= 2
                                    ? 'text-emerald-600 dark:text-emerald-400'
                                    : 'text-slate-400'}">OTP</span
                            >
                        </div>
                        <div
                            class="flex-1 h-0.5 {step >= 3
                                ? 'bg-emerald-400'
                                : 'bg-slate-200 dark:bg-slate-600'} rounded-full"
                        ></div>
                        <div class="flex items-center gap-1.5">
                            <div
                                class="w-7 h-7 rounded-full flex items-center justify-center text-xs font-bold {step >=
                                3
                                    ? 'bg-emerald-500 text-white'
                                    : 'bg-slate-200 dark:bg-slate-600 text-slate-400'}"
                            >
                                3
                            </div>
                            <span
                                class="text-xs font-semibold {step >= 3
                                    ? 'text-emerald-600 dark:text-emerald-400'
                                    : 'text-slate-400'}">Data</span
                            >
                        </div>
                    </div>
                {/if}

                <!-- Success Message -->
                {#if success}
                    <div
                        class="mb-6 p-4 bg-emerald-50 dark:bg-emerald-900/20 border border-emerald-200 dark:border-emerald-800 rounded-xl"
                    >
                        <div class="flex items-start gap-3">
                            <div
                                class="w-10 h-10 rounded-full bg-emerald-100 dark:bg-emerald-900/40 flex items-center justify-center flex-shrink-0"
                            >
                                <span
                                    class="material-symbols-outlined text-emerald-500"
                                    >check_circle</span
                                >
                            </div>
                            <div>
                                <p
                                    class="text-emerald-700 dark:text-emerald-300 text-sm font-semibold mb-1"
                                >
                                    Pendaftaran Berhasil!
                                </p>
                                <p
                                    class="text-emerald-600 dark:text-emerald-400 text-sm"
                                >
                                    {success}
                                </p>
                            </div>
                        </div>
                        <div class="mt-4 flex gap-3">
                            <a
                                href="#/login"
                                class="inline-flex items-center gap-1.5 px-4 py-2 bg-emerald-500 hover:bg-emerald-600 text-white text-sm font-medium rounded-lg transition-colors"
                            >
                                <span
                                    class="material-symbols-outlined"
                                    style="font-size: 16px;">login</span
                                >
                                Ke Halaman Login
                            </a>
                            <a
                                href="#/"
                                class="inline-flex items-center gap-1.5 px-4 py-2 bg-slate-100 dark:bg-slate-700 hover:bg-slate-200 dark:hover:bg-slate-600 text-slate-700 dark:text-slate-300 text-sm font-medium rounded-lg transition-colors"
                            >
                                Kembali ke Beranda
                            </a>
                        </div>
                    </div>
                {/if}

                <!-- Error Message (global) -->
                {#if error}
                    <div
                        class="mb-6 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-xl"
                    >
                        <p
                            class="text-red-600 dark:text-red-400 text-sm flex items-center gap-2"
                        >
                            <span class="material-symbols-outlined text-lg"
                                >error</span
                            >
                            {error}
                        </p>
                    </div>
                {/if}

                {#if !success}
                    <!-- ===== STEP 1: Email Input ===== -->
                    {#if step === 1}
                        <div class="space-y-5">
                            <p
                                class="text-sm text-slate-500 dark:text-slate-400"
                            >
                                Masukkan email aktif Anda. Kode OTP akan dikirim
                                ke email ini untuk verifikasi.
                            </p>

                            <div>
                                <label
                                    for="otp-email"
                                    class="block text-sm font-semibold text-slate-700 dark:text-slate-300 mb-2"
                                >
                                    Email Aktif <span class="text-red-400"
                                        >*</span
                                    >
                                </label>
                                <div class="relative">
                                    <div
                                        class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none"
                                    >
                                        <span
                                            class="material-symbols-outlined text-slate-400 text-xl"
                                            >email</span
                                        >
                                    </div>
                                    <input
                                        type="email"
                                        id="otp-email"
                                        bind:value={otpEmail}
                                        class="w-full h-12 pl-10 pr-4 bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-600 rounded-xl text-slate-900 dark:text-white placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-emerald-500/50 focus:border-emerald-500 transition-all"
                                        placeholder="email@example.com"
                                        on:keydown={(e) =>
                                            e.key === "Enter" && sendOTP()}
                                    />
                                </div>
                                <p class="mt-1 text-xs text-slate-400">
                                    Pastikan email aktif untuk menerima kode
                                    verifikasi
                                </p>
                            </div>

                            <!-- OTP Error -->
                            {#if otpError}
                                <div
                                    class="p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800/30 rounded-lg"
                                >
                                    <p
                                        class="text-sm text-red-600 dark:text-red-400 flex items-center gap-2"
                                    >
                                        <span
                                            class="material-symbols-outlined"
                                            style="font-size: 16px;">error</span
                                        >
                                        {otpError}
                                    </p>
                                </div>
                            {/if}

                            <button
                                on:click={sendOTP}
                                disabled={otpLoading}
                                class="w-full h-12 bg-gradient-to-r from-emerald-500 to-teal-500 hover:from-emerald-600 hover:to-teal-600 disabled:from-emerald-300 disabled:to-teal-300 text-white font-bold rounded-xl transition-all flex items-center justify-center gap-2 shadow-lg shadow-emerald-500/25"
                            >
                                {#if otpLoading}
                                    <svg
                                        class="animate-spin h-5 w-5"
                                        viewBox="0 0 24 24"
                                    >
                                        <circle
                                            class="opacity-25"
                                            cx="12"
                                            cy="12"
                                            r="10"
                                            stroke="currentColor"
                                            stroke-width="4"
                                            fill="none"
                                        ></circle>
                                        <path
                                            class="opacity-75"
                                            fill="currentColor"
                                            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                                        ></path>
                                    </svg>
                                    <span>Mengirim...</span>
                                {:else}
                                    <span
                                        class="material-symbols-outlined text-lg"
                                        >email</span
                                    >
                                    <span>Kirim Kode OTP</span>
                                {/if}
                            </button>
                        </div>
                    {/if}

                    <!-- ===== STEP 2: OTP Verification ===== -->
                    {#if step === 2}
                        <div class="space-y-5">
                            <!-- Email indicator -->
                            <div
                                class="flex items-center gap-2 p-3 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800/30 rounded-lg"
                            >
                                <span
                                    class="material-symbols-outlined text-blue-500"
                                    style="font-size: 18px;">email</span
                                >
                                <span
                                    class="text-sm text-blue-700 dark:text-blue-300 font-medium truncate"
                                    >{otpEmail}</span
                                >
                                <button
                                    on:click={() => {
                                        step = 1;
                                        otpError = "";
                                        otpSuccess = "";
                                        otpCode = "";
                                    }}
                                    class="ml-auto text-xs text-blue-500 hover:text-blue-700 font-semibold underline flex-shrink-0"
                                >
                                    Ubah
                                </button>
                            </div>

                            {#if otpSuccess}
                                <div
                                    class="p-3 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800/30 rounded-lg"
                                >
                                    <p
                                        class="text-sm text-green-700 dark:text-green-400 font-medium flex items-center gap-2"
                                    >
                                        <span
                                            class="material-symbols-outlined"
                                            style="font-size: 16px;"
                                            >check_circle</span
                                        >
                                        {otpSuccess}
                                    </p>
                                </div>
                            {/if}

                            <div>
                                <label
                                    for="otp-code"
                                    class="block text-sm font-semibold text-slate-700 dark:text-slate-300 mb-2"
                                >
                                    Masukkan Kode OTP <span class="text-red-400"
                                        >*</span
                                    >
                                </label>
                                <input
                                    id="otp-code"
                                    type="text"
                                    bind:value={otpCode}
                                    placeholder="Masukkan 6 digit kode"
                                    maxlength="6"
                                    class="w-full h-14 px-4 text-center text-2xl font-mono font-bold tracking-[8px] bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-600 rounded-xl focus:ring-2 focus:ring-emerald-500/50 focus:border-emerald-500 text-slate-900 dark:text-white placeholder:text-sm placeholder:tracking-normal placeholder:font-normal transition-all"
                                    on:keydown={(e) =>
                                        e.key === "Enter" && verifyOTP()}
                                />
                                <p
                                    class="mt-2 text-xs text-slate-400 dark:text-slate-500"
                                >
                                    Kode berlaku selama 5 menit.
                                    {#if otpCountdown > 0}
                                        Kirim ulang dalam <span
                                            class="font-bold text-emerald-500"
                                            >{otpCountdown}s</span
                                        >
                                    {:else}
                                        <button
                                            on:click={sendOTP}
                                            disabled={otpLoading}
                                            class="text-emerald-500 hover:text-emerald-700 font-semibold underline disabled:opacity-50"
                                        >
                                            Kirim ulang OTP
                                        </button>
                                    {/if}
                                </p>
                            </div>

                            <!-- OTP Error -->
                            {#if otpError}
                                <div
                                    class="p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800/30 rounded-lg"
                                >
                                    <p
                                        class="text-sm text-red-600 dark:text-red-400 flex items-center gap-2"
                                    >
                                        <span
                                            class="material-symbols-outlined"
                                            style="font-size: 16px;">error</span
                                        >
                                        {otpError}
                                    </p>
                                </div>
                            {/if}

                            <button
                                on:click={verifyOTP}
                                disabled={otpLoading}
                                class="w-full h-12 bg-gradient-to-r from-emerald-500 to-teal-500 hover:from-emerald-600 hover:to-teal-600 disabled:from-emerald-300 disabled:to-teal-300 text-white font-bold rounded-xl transition-all flex items-center justify-center gap-2 shadow-lg shadow-emerald-500/25"
                            >
                                {#if otpLoading}
                                    <svg
                                        class="animate-spin h-5 w-5"
                                        viewBox="0 0 24 24"
                                    >
                                        <circle
                                            class="opacity-25"
                                            cx="12"
                                            cy="12"
                                            r="10"
                                            stroke="currentColor"
                                            stroke-width="4"
                                            fill="none"
                                        ></circle>
                                        <path
                                            class="opacity-75"
                                            fill="currentColor"
                                            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                                        ></path>
                                    </svg>
                                    <span>Memverifikasi...</span>
                                {:else}
                                    <span
                                        class="material-symbols-outlined text-lg"
                                        >verified</span
                                    >
                                    <span>Verifikasi OTP</span>
                                {/if}
                            </button>
                        </div>
                    {/if}

                    <!-- ===== STEP 3: Registration Form ===== -->
                    {#if step === 3}
                        <div class="space-y-5">
                            <!-- Email verified badge -->
                            <div
                                class="flex items-center gap-2 p-3 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800/30 rounded-lg"
                            >
                                <span
                                    class="material-symbols-outlined text-green-500"
                                    style="font-size: 18px;">verified</span
                                >
                                <span
                                    class="text-sm text-green-700 dark:text-green-300 font-medium truncate"
                                    >{email}</span
                                >
                                <span
                                    class="ml-auto inline-flex items-center gap-0.5 px-2 py-0.5 bg-green-100 dark:bg-green-800/30 text-green-600 dark:text-green-400 text-[10px] font-bold rounded uppercase"
                                >
                                    <span
                                        class="material-symbols-outlined"
                                        style="font-size: 11px;">check</span
                                    >
                                    Terverifikasi
                                </span>
                            </div>

                            {#if otpSuccess}
                                <div
                                    class="p-3 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800/30 rounded-lg"
                                >
                                    <p
                                        class="text-sm text-green-700 dark:text-green-400 font-medium flex items-center gap-2"
                                    >
                                        <span
                                            class="material-symbols-outlined"
                                            style="font-size: 16px;"
                                            >check_circle</span
                                        >
                                        {otpSuccess}
                                    </p>
                                </div>
                            {/if}

                            <form
                                on:submit|preventDefault={handleSubmit}
                                class="space-y-5"
                            >
                                <!-- Nama Lengkap -->
                                <div>
                                    <label
                                        for="name"
                                        class="block text-sm font-semibold text-slate-700 dark:text-slate-300 mb-2"
                                    >
                                        Nama Lengkap <span class="text-red-400"
                                            >*</span
                                        >
                                    </label>
                                    <div class="relative">
                                        <div
                                            class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none"
                                        >
                                            <span
                                                class="material-symbols-outlined text-slate-400 text-xl"
                                                >person</span
                                            >
                                        </div>
                                        <input
                                            type="text"
                                            id="name"
                                            bind:value={name}
                                            class="w-full h-12 pl-10 pr-4 bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-600 rounded-xl text-slate-900 dark:text-white placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-emerald-500/50 focus:border-emerald-500 transition-all"
                                            placeholder="Masukkan nama lengkap"
                                            required
                                        />
                                    </div>
                                </div>

                                <!-- Password -->
                                <div>
                                    <label
                                        for="password"
                                        class="block text-sm font-semibold text-slate-700 dark:text-slate-300 mb-2"
                                    >
                                        Password <span class="text-red-400"
                                            >*</span
                                        >
                                    </label>
                                    <div class="relative">
                                        <div
                                            class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none"
                                        >
                                            <span
                                                class="material-symbols-outlined text-slate-400 text-xl"
                                                >lock</span
                                            >
                                        </div>
                                        {#if showPassword}
                                            <input
                                                type="text"
                                                id="password"
                                                bind:value={password}
                                                class="w-full h-12 pl-10 pr-12 bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-600 rounded-xl text-slate-900 dark:text-white placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-emerald-500/50 focus:border-emerald-500 transition-all"
                                                placeholder="Minimal 6 karakter"
                                                required
                                            />
                                        {:else}
                                            <input
                                                type="password"
                                                id="password"
                                                bind:value={password}
                                                class="w-full h-12 pl-10 pr-12 bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-600 rounded-xl text-slate-900 dark:text-white placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-emerald-500/50 focus:border-emerald-500 transition-all"
                                                placeholder="Minimal 6 karakter"
                                                required
                                            />
                                        {/if}
                                        <button
                                            type="button"
                                            on:click={() =>
                                                (showPassword = !showPassword)}
                                            class="absolute inset-y-0 right-0 pr-3 flex items-center text-slate-400 hover:text-slate-600 dark:hover:text-slate-300"
                                        >
                                            <span
                                                class="material-symbols-outlined text-xl"
                                            >
                                                {showPassword
                                                    ? "visibility_off"
                                                    : "visibility"}
                                            </span>
                                        </button>
                                    </div>
                                </div>

                                <!-- Confirm Password -->
                                <div>
                                    <label
                                        for="confirmPassword"
                                        class="block text-sm font-semibold text-slate-700 dark:text-slate-300 mb-2"
                                    >
                                        Konfirmasi Password <span
                                            class="text-red-400">*</span
                                        >
                                    </label>
                                    <div class="relative">
                                        <div
                                            class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none"
                                        >
                                            <span
                                                class="material-symbols-outlined text-slate-400 text-xl"
                                                >lock_clock</span
                                            >
                                        </div>
                                        {#if showPassword}
                                            <input
                                                type="text"
                                                id="confirmPassword"
                                                bind:value={confirmPassword}
                                                class="w-full h-12 pl-10 pr-4 bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-600 rounded-xl text-slate-900 dark:text-white placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-emerald-500/50 focus:border-emerald-500 transition-all {confirmPassword &&
                                                password !== confirmPassword
                                                    ? 'border-red-400 focus:ring-red-500/50 focus:border-red-500'
                                                    : ''}"
                                                placeholder="Ulangi password"
                                                required
                                            />
                                        {:else}
                                            <input
                                                type="password"
                                                id="confirmPassword"
                                                bind:value={confirmPassword}
                                                class="w-full h-12 pl-10 pr-4 bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-600 rounded-xl text-slate-900 dark:text-white placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-emerald-500/50 focus:border-emerald-500 transition-all {confirmPassword &&
                                                password !== confirmPassword
                                                    ? 'border-red-400 focus:ring-red-500/50 focus:border-red-500'
                                                    : ''}"
                                                placeholder="Ulangi password"
                                                required
                                            />
                                        {/if}
                                    </div>
                                    {#if confirmPassword && password !== confirmPassword}
                                        <p
                                            class="mt-1 text-xs text-red-400 flex items-center gap-1"
                                        >
                                            <span
                                                class="material-symbols-outlined"
                                                style="font-size: 12px;"
                                                >error</span
                                            >
                                            Password tidak cocok
                                        </p>
                                    {/if}
                                </div>

                                <!-- Upload KTM -->
                                <div>
                                    <label
                                        for="ktm-upload"
                                        class="block text-sm font-semibold text-slate-700 dark:text-slate-300 mb-2"
                                    >
                                        Upload KTM <span class="text-red-400"
                                            >*</span
                                        >
                                    </label>

                                    {#if ktmFile}
                                        <!-- KTM Preview -->
                                        <div
                                            class="relative border-2 border-emerald-300 dark:border-emerald-700 rounded-xl overflow-hidden bg-slate-50 dark:bg-slate-800"
                                        >
                                            {#if ktmPreview}
                                                <img
                                                    src={ktmPreview}
                                                    alt="KTM Preview"
                                                    class="w-full h-40 object-contain p-2"
                                                />
                                            {:else}
                                                <div
                                                    class="flex items-center justify-center h-24 gap-2 text-slate-500"
                                                >
                                                    <span
                                                        class="material-symbols-outlined"
                                                        >picture_as_pdf</span
                                                    >
                                                    <span class="text-sm"
                                                        >{ktmFile.name}</span
                                                    >
                                                </div>
                                            {/if}
                                            <div
                                                class="px-3 py-2 bg-emerald-50 dark:bg-emerald-900/20 border-t border-emerald-200 dark:border-emerald-800 flex items-center justify-between"
                                            >
                                                <div
                                                    class="flex items-center gap-2"
                                                >
                                                    <span
                                                        class="material-symbols-outlined text-emerald-500"
                                                        style="font-size: 16px;"
                                                        >check_circle</span
                                                    >
                                                    <span
                                                        class="text-xs text-emerald-600 dark:text-emerald-400 font-medium truncate max-w-[200px]"
                                                        >{ktmFile.name}</span
                                                    >
                                                </div>
                                                <button
                                                    type="button"
                                                    on:click={removeKtm}
                                                    class="text-red-400 hover:text-red-500 transition-colors"
                                                >
                                                    <span
                                                        class="material-symbols-outlined"
                                                        style="font-size: 18px;"
                                                        >close</span
                                                    >
                                                </button>
                                            </div>
                                        </div>
                                    {:else}
                                        <!-- Upload Area -->
                                        <label
                                            for="ktm-upload"
                                            class="flex flex-col items-center justify-center h-32 border-2 border-dashed border-slate-300 dark:border-slate-600 rounded-xl cursor-pointer hover:border-emerald-400 dark:hover:border-emerald-600 hover:bg-emerald-50/50 dark:hover:bg-emerald-900/10 transition-all group"
                                        >
                                            <div
                                                class="w-12 h-12 rounded-xl bg-slate-100 dark:bg-slate-700 flex items-center justify-center mb-2 group-hover:bg-emerald-100 dark:group-hover:bg-emerald-900/30 transition-colors"
                                            >
                                                <span
                                                    class="material-symbols-outlined text-slate-400 group-hover:text-emerald-500 transition-colors text-2xl"
                                                    >badge</span
                                                >
                                            </div>
                                            <p
                                                class="text-sm text-slate-500 dark:text-slate-400 font-medium"
                                            >
                                                Klik untuk upload <span
                                                    class="text-emerald-500"
                                                    >KTM</span
                                                >
                                            </p>
                                            <p
                                                class="text-xs text-slate-400 mt-0.5"
                                            >
                                                JPG, PNG, atau PDF (maks. 5MB)
                                            </p>
                                        </label>
                                    {/if}
                                    <input
                                        type="file"
                                        id="ktm-upload"
                                        accept="image/*,.pdf"
                                        on:change={handleKtmUpload}
                                        class="hidden"
                                    />
                                </div>

                                <!-- Info box -->
                                <div
                                    class="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-xl p-4"
                                >
                                    <div class="flex items-start gap-3">
                                        <span
                                            class="material-symbols-outlined text-blue-500 flex-shrink-0"
                                            style="font-size: 20px;">info</span
                                        >
                                        <div
                                            class="text-xs text-blue-600 dark:text-blue-400 leading-relaxed"
                                        >
                                            <p class="font-semibold mb-1">
                                                Proses Pendaftaran:
                                            </p>
                                            <ol
                                                class="list-decimal list-inside space-y-0.5"
                                            >
                                                <li>
                                                    Isi form dan upload foto KTM
                                                    Anda
                                                </li>
                                                <li>
                                                    Admin akan memverifikasi
                                                    data Anda
                                                </li>
                                                <li>
                                                    Jika disetujui, informasi
                                                    akun dikirim via email
                                                </li>
                                                <li>
                                                    Login dengan email &
                                                    password yang didaftarkan
                                                </li>
                                            </ol>
                                        </div>
                                    </div>
                                </div>

                                <!-- Submit Button -->
                                <button
                                    type="submit"
                                    disabled={loading}
                                    class="w-full h-12 bg-gradient-to-r from-emerald-500 to-teal-500 hover:from-emerald-600 hover:to-teal-600 disabled:from-emerald-300 disabled:to-teal-300 text-white font-bold rounded-xl transition-all flex items-center justify-center gap-2 shadow-lg shadow-emerald-500/25"
                                >
                                    {#if loading}
                                        <svg
                                            class="animate-spin h-5 w-5"
                                            viewBox="0 0 24 24"
                                        >
                                            <circle
                                                class="opacity-25"
                                                cx="12"
                                                cy="12"
                                                r="10"
                                                stroke="currentColor"
                                                stroke-width="4"
                                                fill="none"
                                            ></circle>
                                            <path
                                                class="opacity-75"
                                                fill="currentColor"
                                                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                                            ></path>
                                        </svg>
                                        <span>Mendaftar...</span>
                                    {:else}
                                        <span
                                            class="material-symbols-outlined text-lg"
                                            >how_to_reg</span
                                        >
                                        <span>Daftar Sekarang</span>
                                    {/if}
                                </button>
                            </form>
                        </div>
                    {/if}
                {/if}

                <!-- Divider -->
                <div class="my-6 flex items-center gap-4">
                    <div
                        class="flex-1 h-px bg-slate-200 dark:bg-slate-700"
                    ></div>
                    <span class="text-sm text-slate-400">atau</span>
                    <div
                        class="flex-1 h-px bg-slate-200 dark:bg-slate-700"
                    ></div>
                </div>

                <!-- Links -->
                <div class="text-center space-y-2">
                    <p class="text-slate-600 dark:text-slate-400 text-sm">
                        Sudah punya akun?
                        <a
                            href="#/login"
                            class="text-primary font-semibold hover:underline"
                            >Login di sini</a
                        >
                    </p>
                </div>
            </div>
        </div>

        <!-- Back to home -->
        <p class="mt-6 text-center">
            <a
                href="#/"
                class="text-slate-500 dark:text-slate-400 hover:text-primary transition-colors inline-flex items-center gap-1 text-sm"
            >
                <span class="material-symbols-outlined text-sm">arrow_back</span
                >
                Kembali ke beranda
            </a>
        </p>
    </div>
</section>

<style>
    .bg-surface-dark {
        background-color: #192633;
    }
    .bg-surface-highlight {
        background-color: #233648;
    }
</style>
