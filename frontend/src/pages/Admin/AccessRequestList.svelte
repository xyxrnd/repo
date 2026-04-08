<script>
    import { onMount } from "svelte";
    import authService from "../../services/authService";
    import { API_ENDPOINTS, API_BASE_URL } from "../../config";

    let requests = [];
    let loading = true;
    let error = "";
    let actionLoading = null; // ID yang sedang diproses
    let previewKtm = null; // KTM yang sedang dilihat
    let deleteConfirm = null;
    let filterStatus = "all";

    // Helper: cek apakah ktm_path adalah Google Drive file ID
    function isGDriveId(path) {
        if (!path) return false;
        if (path.includes("/") || path.includes("\\")) return false;
        if (path.includes(".")) return false;
        return path.length > 20;
    }

    // Helper: dapatkan URL KTM (Google Drive atau lokal)
    function getKtmUrl(path) {
        if (isGDriveId(path)) {
            return `https://drive.google.com/file/d/${path}/view`;
        }
        return `${API_BASE_URL}/${path}`;
    }

    function getKtmImageUrl(path) {
        if (isGDriveId(path)) {
            return `${API_BASE_URL}/api/gdrive-proxy/${path}`;
        }
        return `${API_BASE_URL}/${path}`;
    }

    const token = authService.getToken();

    onMount(async () => {
        await loadRequests();
    });

    async function loadRequests() {
        loading = true;
        error = "";
        try {
            const response = await fetch(API_ENDPOINTS.ACCESS_REQUESTS, {
                headers: { Authorization: `Bearer ${token}` },
            });
            if (!response.ok) throw new Error("Gagal mengambil data");
            requests = await response.json();
        } catch (err) {
            error = err.message;
        } finally {
            loading = false;
        }
    }

    async function handleApprove(req) {
        actionLoading = req.id;
        try {
            const response = await fetch(
                API_ENDPOINTS.ACCESS_REQUEST_BY_ID(req.id),
                {
                    method: "PUT",
                    headers: {
                        "Content-Type": "application/json",
                        Authorization: `Bearer ${token}`,
                    },
                    body: JSON.stringify({ status: "approved" }),
                },
            );

            if (!response.ok) throw new Error("Gagal menyetujui permintaan");

            const data = await response.json();
            alert(
                data.message ||
                    "Permintaan disetujui & token dikirim ke email.",
            );
            await loadRequests();
        } catch (err) {
            alert("Error: " + err.message);
        } finally {
            actionLoading = null;
        }
    }

    async function handleReject(req) {
        if (!confirm(`Tolak permintaan dari ${req.nama}?`)) return;

        actionLoading = req.id;
        try {
            const response = await fetch(
                API_ENDPOINTS.ACCESS_REQUEST_BY_ID(req.id),
                {
                    method: "PUT",
                    headers: {
                        "Content-Type": "application/json",
                        Authorization: `Bearer ${token}`,
                    },
                    body: JSON.stringify({ status: "rejected" }),
                },
            );

            if (!response.ok) throw new Error("Gagal menolak permintaan");
            await loadRequests();
        } catch (err) {
            alert("Error: " + err.message);
        } finally {
            actionLoading = null;
        }
    }

    async function handleDelete(req) {
        deleteConfirm = req;
    }

    async function confirmDelete() {
        if (!deleteConfirm) return;
        actionLoading = deleteConfirm.id;
        try {
            const response = await fetch(
                API_ENDPOINTS.ACCESS_REQUEST_BY_ID(deleteConfirm.id),
                {
                    method: "DELETE",
                    headers: { Authorization: `Bearer ${token}` },
                },
            );
            if (!response.ok) throw new Error("Gagal menghapus");
            deleteConfirm = null;
            await loadRequests();
        } catch (err) {
            alert("Error: " + err.message);
        } finally {
            actionLoading = null;
        }
    }

    function formatDate(dateString) {
        return new Date(dateString).toLocaleDateString("id-ID", {
            year: "numeric",
            month: "short",
            day: "numeric",
            hour: "2-digit",
            minute: "2-digit",
        });
    }

    function getStatusStyle(status) {
        switch (status) {
            case "pending":
                return {
                    bg: "bg-amber-100 dark:bg-amber-900/30",
                    text: "text-amber-700 dark:text-amber-400",
                    icon: "hourglass_top",
                    label: "Menunggu",
                };
            case "approved":
                return {
                    bg: "bg-green-100 dark:bg-green-900/30",
                    text: "text-green-700 dark:text-green-400",
                    icon: "check_circle",
                    label: "Disetujui",
                };
            case "rejected":
                return {
                    bg: "bg-red-100 dark:bg-red-900/30",
                    text: "text-red-700 dark:text-red-400",
                    icon: "cancel",
                    label: "Ditolak",
                };
            default:
                return {
                    bg: "bg-slate-100",
                    text: "text-slate-500",
                    icon: "help",
                    label: status,
                };
        }
    }

    $: filteredRequests =
        filterStatus === "all"
            ? requests
            : requests.filter((r) => r.status === filterStatus);

    $: pendingCount = requests.filter((r) => r.status === "pending").length;
    $: approvedCount = requests.filter((r) => r.status === "approved").length;
    $: rejectedCount = requests.filter((r) => r.status === "rejected").length;
</script>

<div class="space-y-6">
    <!-- Header -->
    <div
        class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4"
    >
        <div>
            <h1 class="text-2xl font-bold text-slate-900 dark:text-white">
                Permintaan Akses
            </h1>
            <p class="text-slate-500 dark:text-slate-400 mt-1">
                Kelola permintaan akses file terkunci dari pengunjung
            </p>
        </div>
        <button
            on:click={loadRequests}
            class="inline-flex items-center gap-2 px-4 py-2.5 bg-slate-100 dark:bg-slate-700 hover:bg-slate-200 dark:hover:bg-slate-600 text-slate-700 dark:text-slate-300 font-medium rounded-lg transition-all"
        >
            <span class="material-symbols-outlined text-xl">refresh</span>
            Refresh
        </button>
    </div>

    <!-- Stats Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <button
            on:click={() => (filterStatus = "pending")}
            class="p-4 rounded-xl border-2 transition-all {filterStatus ===
            'pending'
                ? 'border-amber-400 bg-amber-50 dark:bg-amber-900/20'
                : 'border-slate-200 dark:border-slate-700 bg-white dark:bg-[#192633]'}"
        >
            <div class="flex items-center gap-3">
                <div
                    class="w-10 h-10 rounded-lg bg-amber-100 dark:bg-amber-900/30 flex items-center justify-center"
                >
                    <span
                        class="material-symbols-outlined text-amber-500"
                        style="font-size: 22px;">hourglass_top</span
                    >
                </div>
                <div class="text-left">
                    <p
                        class="text-2xl font-bold text-slate-900 dark:text-white"
                    >
                        {pendingCount}
                    </p>
                    <p class="text-xs text-slate-500 dark:text-slate-400">
                        Menunggu
                    </p>
                </div>
            </div>
        </button>
        <button
            on:click={() => (filterStatus = "approved")}
            class="p-4 rounded-xl border-2 transition-all {filterStatus ===
            'approved'
                ? 'border-green-400 bg-green-50 dark:bg-green-900/20'
                : 'border-slate-200 dark:border-slate-700 bg-white dark:bg-[#192633]'}"
        >
            <div class="flex items-center gap-3">
                <div
                    class="w-10 h-10 rounded-lg bg-green-100 dark:bg-green-900/30 flex items-center justify-center"
                >
                    <span
                        class="material-symbols-outlined text-green-500"
                        style="font-size: 22px;">check_circle</span
                    >
                </div>
                <div class="text-left">
                    <p
                        class="text-2xl font-bold text-slate-900 dark:text-white"
                    >
                        {approvedCount}
                    </p>
                    <p class="text-xs text-slate-500 dark:text-slate-400">
                        Disetujui
                    </p>
                </div>
            </div>
        </button>
        <button
            on:click={() => (filterStatus = "rejected")}
            class="p-4 rounded-xl border-2 transition-all {filterStatus ===
            'rejected'
                ? 'border-red-400 bg-red-50 dark:bg-red-900/20'
                : 'border-slate-200 dark:border-slate-700 bg-white dark:bg-[#192633]'}"
        >
            <div class="flex items-center gap-3">
                <div
                    class="w-10 h-10 rounded-lg bg-red-100 dark:bg-red-900/30 flex items-center justify-center"
                >
                    <span
                        class="material-symbols-outlined text-red-500"
                        style="font-size: 22px;">cancel</span
                    >
                </div>
                <div class="text-left">
                    <p
                        class="text-2xl font-bold text-slate-900 dark:text-white"
                    >
                        {rejectedCount}
                    </p>
                    <p class="text-xs text-slate-500 dark:text-slate-400">
                        Ditolak
                    </p>
                </div>
            </div>
        </button>
    </div>

    <!-- Filter bar -->
    <div class="flex items-center gap-2">
        <button
            on:click={() => (filterStatus = "all")}
            class="px-3 py-1.5 text-xs font-semibold rounded-lg transition-all {filterStatus ===
            'all'
                ? 'bg-primary text-white'
                : 'bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300 hover:bg-slate-200'}"
        >
            Semua ({requests.length})
        </button>
        <button
            on:click={() => (filterStatus = "pending")}
            class="px-3 py-1.5 text-xs font-semibold rounded-lg transition-all {filterStatus ===
            'pending'
                ? 'bg-amber-500 text-white'
                : 'bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300 hover:bg-slate-200'}"
        >
            Menunggu ({pendingCount})
        </button>
        <button
            on:click={() => (filterStatus = "approved")}
            class="px-3 py-1.5 text-xs font-semibold rounded-lg transition-all {filterStatus ===
            'approved'
                ? 'bg-green-500 text-white'
                : 'bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300 hover:bg-slate-200'}"
        >
            Disetujui ({approvedCount})
        </button>
        <button
            on:click={() => (filterStatus = "rejected")}
            class="px-3 py-1.5 text-xs font-semibold rounded-lg transition-all {filterStatus ===
            'rejected'
                ? 'bg-red-500 text-white'
                : 'bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300 hover:bg-slate-200'}"
        >
            Ditolak ({rejectedCount})
        </button>
    </div>

    <!-- Error -->
    {#if error}
        <div
            class="p-4 bg-red-100 dark:bg-red-900/30 border border-red-200 dark:border-red-800 rounded-lg"
        >
            <p class="text-red-600 dark:text-red-400 flex items-center gap-2">
                <span class="material-symbols-outlined">error</span>
                {error}
            </p>
        </div>
    {/if}

    <!-- Table -->
    <div
        class="bg-white dark:bg-[#192633] border border-slate-200 dark:border-slate-700 rounded-xl overflow-hidden shadow-sm"
    >
        {#if loading}
            <div class="p-12 text-center">
                <div class="inline-flex items-center gap-3 text-slate-500">
                    <svg class="animate-spin h-6 w-6" viewBox="0 0 24 24">
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
                    Loading...
                </div>
            </div>
        {:else if filteredRequests.length === 0}
            <div class="p-12 text-center">
                <span
                    class="material-symbols-outlined text-6xl text-slate-300 dark:text-slate-600"
                    >inbox</span
                >
                <p class="mt-4 text-slate-500 dark:text-slate-400">
                    Tidak ada permintaan akses
                    {filterStatus !== "all"
                        ? `dengan status "${filterStatus}"`
                        : ""}
                </p>
            </div>
        {:else}
            <div class="overflow-x-auto">
                <table class="w-full">
                    <thead class="bg-slate-50 dark:bg-slate-800/50">
                        <tr>
                            <th
                                class="px-6 py-4 text-left text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider"
                                >Pemohon</th
                            >
                            <th
                                class="px-6 py-4 text-left text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider"
                                >Dokumen</th
                            >
                            <th
                                class="px-6 py-4 text-left text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider"
                                >KTM</th
                            >
                            <th
                                class="px-6 py-4 text-left text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider"
                                >Status</th
                            >
                            <th
                                class="px-6 py-4 text-left text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider"
                                >Tanggal</th
                            >
                            <th
                                class="px-6 py-4 text-right text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider"
                                >Aksi</th
                            >
                        </tr>
                    </thead>
                    <tbody
                        class="divide-y divide-slate-100 dark:divide-slate-700"
                    >
                        {#each filteredRequests as req}
                            {@const statusStyle = getStatusStyle(req.status)}
                            <tr
                                class="hover:bg-slate-50 dark:hover:bg-slate-800/30 transition-colors"
                            >
                                <!-- Pemohon -->
                                <td class="px-6 py-4">
                                    <div>
                                        <p
                                            class="font-semibold text-slate-900 dark:text-white text-sm"
                                        >
                                            {req.nama}
                                        </p>
                                        <p
                                            class="text-xs text-slate-500 dark:text-slate-400 mt-0.5"
                                        >
                                            {req.email}
                                        </p>
                                    </div>
                                </td>

                                <!-- Dokumen -->
                                <td class="px-6 py-4">
                                    <p
                                        class="text-sm text-slate-900 dark:text-white font-medium truncate max-w-[250px]"
                                        title={req.doc_judul}
                                    >
                                        {req.doc_judul || "-"}
                                    </p>
                                </td>

                                <!-- KTM -->
                                <td class="px-6 py-4">
                                    {#if req.ktm_path}
                                        <button
                                            on:click={() =>
                                                (previewKtm = req.ktm_path)}
                                            class="inline-flex items-center gap-1 px-2 py-1 bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400 text-xs font-medium rounded-lg hover:bg-blue-100 dark:hover:bg-blue-900/40 transition-colors"
                                        >
                                            <span
                                                class="material-symbols-outlined"
                                                style="font-size: 14px;"
                                                >badge</span
                                            >
                                            Lihat KTM
                                        </button>
                                    {:else}
                                        <span
                                            class="text-xs text-slate-400 italic"
                                            >Tidak ada</span
                                        >
                                    {/if}
                                </td>

                                <!-- Status -->
                                <td class="px-6 py-4">
                                    <span
                                        class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-semibold {statusStyle.bg} {statusStyle.text}"
                                    >
                                        <span
                                            class="material-symbols-outlined"
                                            style="font-size: 14px;"
                                            >{statusStyle.icon}</span
                                        >
                                        {statusStyle.label}
                                    </span>
                                    {#if req.status === "approved" && req.access_token}
                                        <p
                                            class="text-[10px] text-green-600 dark:text-green-400 mt-1 font-mono"
                                        >
                                            Token: {req.access_token}
                                        </p>
                                    {/if}
                                </td>

                                <!-- Tanggal -->
                                <td
                                    class="px-6 py-4 text-sm text-slate-500 dark:text-slate-400 whitespace-nowrap"
                                >
                                    {formatDate(req.created_at)}
                                </td>

                                <!-- Aksi -->
                                <td class="px-6 py-4 text-right">
                                    <div
                                        class="flex items-center justify-end gap-1"
                                    >
                                        {#if req.status === "pending"}
                                            <button
                                                on:click={() =>
                                                    handleApprove(req)}
                                                disabled={actionLoading ===
                                                    req.id}
                                                class="inline-flex items-center gap-1 px-3 py-1.5 bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 hover:bg-green-200 dark:hover:bg-green-900/50 rounded-lg text-xs font-semibold transition-all disabled:opacity-50"
                                                title="Setujui & kirim token via email"
                                            >
                                                {#if actionLoading === req.id}
                                                    <span
                                                        class="material-symbols-outlined animate-spin"
                                                        style="font-size: 14px;"
                                                        >progress_activity</span
                                                    >
                                                {:else}
                                                    <span
                                                        class="material-symbols-outlined"
                                                        style="font-size: 14px;"
                                                        >check</span
                                                    >
                                                {/if}
                                                Setujui
                                            </button>
                                            <button
                                                on:click={() =>
                                                    handleReject(req)}
                                                disabled={actionLoading ===
                                                    req.id}
                                                class="inline-flex items-center gap-1 px-3 py-1.5 bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400 hover:bg-red-200 dark:hover:bg-red-900/50 rounded-lg text-xs font-semibold transition-all disabled:opacity-50"
                                                title="Tolak permintaan"
                                            >
                                                <span
                                                    class="material-symbols-outlined"
                                                    style="font-size: 14px;"
                                                    >close</span
                                                >
                                                Tolak
                                            </button>
                                        {/if}
                                        <button
                                            on:click={() => handleDelete(req)}
                                            class="p-1.5 text-slate-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors"
                                            title="Hapus"
                                        >
                                            <span
                                                class="material-symbols-outlined"
                                                style="font-size: 18px;"
                                                >delete</span
                                            >
                                        </button>
                                    </div>
                                </td>
                            </tr>
                        {/each}
                    </tbody>
                </table>
            </div>
        {/if}
    </div>
</div>

<!-- KTM Preview Modal -->
{#if previewKtm}
    <div
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm"
        on:click|self={() => (previewKtm = null)}
        on:keydown={(e) => e.key === "Escape" && (previewKtm = null)}
        role="button"
        tabindex="0"
    >
        <div
            class="bg-white dark:bg-[#192633] border border-slate-200 dark:border-slate-700 rounded-2xl shadow-2xl w-full max-w-lg overflow-hidden"
        >
            <div
                class="px-6 py-4 border-b border-slate-200 dark:border-slate-700 flex items-center justify-between"
            >
                <h3
                    class="text-lg font-bold text-slate-900 dark:text-white flex items-center gap-2"
                >
                    <span class="material-symbols-outlined text-primary"
                        >badge</span
                    >
                    Preview KTM
                </h3>
                <button
                    on:click={() => (previewKtm = null)}
                    class="p-1 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors"
                >
                    <span class="material-symbols-outlined text-slate-400"
                        >close</span
                    >
                </button>
            </div>
            <div class="p-6">
                {#if isGDriveId(previewKtm)}
                    <img
                        src={getKtmImageUrl(previewKtm)}
                        alt="KTM"
                        class="w-full rounded-lg border border-slate-200 dark:border-slate-600 object-contain max-h-96"
                    />
                    <a
                        href={getKtmUrl(previewKtm)}
                        target="_blank"
                        rel="noopener noreferrer"
                        class="mt-3 inline-flex items-center gap-2 px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white text-sm font-medium rounded-lg transition-colors"
                    >
                        <span
                            class="material-symbols-outlined"
                            style="font-size: 16px;">open_in_new</span
                        >
                        Buka di Google Drive
                    </a>
                {:else if previewKtm.endsWith(".pdf")}
                    <iframe
                        src="{API_BASE_URL}/{previewKtm}"
                        class="w-full h-96 rounded-lg border border-slate-200 dark:border-slate-600"
                        title="KTM Preview"
                    ></iframe>
                {:else}
                    <img
                        src="{API_BASE_URL}/{previewKtm}"
                        alt="KTM"
                        class="w-full rounded-lg border border-slate-200 dark:border-slate-600 object-contain max-h-96"
                    />
                {/if}
            </div>
        </div>
    </div>
{/if}

<!-- Delete Confirmation Modal -->
{#if deleteConfirm}
    <div
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm"
    >
        <div
            class="bg-white dark:bg-[#192633] border border-slate-200 dark:border-slate-700 rounded-2xl shadow-2xl w-full max-w-sm p-6"
        >
            <div class="text-center">
                <div
                    class="w-16 h-16 mx-auto mb-4 bg-red-100 dark:bg-red-900/30 rounded-full flex items-center justify-center"
                >
                    <span
                        class="material-symbols-outlined text-4xl text-red-500"
                        >warning</span
                    >
                </div>
                <h3
                    class="text-xl font-bold text-slate-900 dark:text-white mb-2"
                >
                    Hapus Permintaan?
                </h3>
                <p class="text-slate-500 dark:text-slate-400 mb-6 text-sm">
                    Hapus permintaan akses dari <strong
                        class="text-slate-900 dark:text-white"
                        >{deleteConfirm.nama}</strong
                    >? Tindakan ini tidak dapat dibatalkan.
                </p>
                <div class="flex gap-3">
                    <button
                        on:click={() => (deleteConfirm = null)}
                        class="flex-1 h-11 border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-300 font-medium rounded-lg hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors"
                    >
                        Batal
                    </button>
                    <button
                        on:click={confirmDelete}
                        class="flex-1 h-11 bg-red-500 hover:bg-red-600 text-white font-medium rounded-lg transition-colors"
                    >
                        Hapus
                    </button>
                </div>
            </div>
        </div>
    </div>
{/if}
