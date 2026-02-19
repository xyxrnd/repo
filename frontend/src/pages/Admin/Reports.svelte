<script>
    import { onMount, onDestroy } from "svelte";
    import { getDocuments } from "../../services/documentService";

    let loading = true;
    let documents = [];

    // Report data
    let totalDocs = 0;
    let publishedDocs = 0;
    let draftDocs = 0;
    let totalViews = 0;
    let docsByType = {};
    let docsByMonth = [];
    let recentActivity = [];

    // Filter
    let selectedPeriod = "all";
    let selectedType = "all";

    // Real-time
    let refreshInterval;
    let lastUpdated = null;
    let nextRefreshIn = 30;
    let countdownInterval;
    let isRefreshing = false;

    // Export
    let showExportMenu = false;
    let exportLoading = false;

    onMount(() => {
        loadData();
        // Auto-refresh tiap 30 detik
        refreshInterval = setInterval(() => {
            loadData(true);
        }, 30000);
        // Countdown timer
        countdownInterval = setInterval(() => {
            nextRefreshIn = Math.max(0, nextRefreshIn - 1);
        }, 1000);
    });

    onDestroy(() => {
        if (refreshInterval) clearInterval(refreshInterval);
        if (countdownInterval) clearInterval(countdownInterval);
    });

    async function loadData(silent = false) {
        try {
            if (!silent) loading = true;
            isRefreshing = true;
            documents = await getDocuments();
            calculateStats();
            buildActivityLog();
            lastUpdated = new Date();
            nextRefreshIn = 30;
        } catch (e) {
            console.error("Failed to load data:", e);
        } finally {
            loading = false;
            isRefreshing = false;
        }
    }

    function getFilteredDocuments() {
        let filtered = [...documents];

        // Filter by period
        if (selectedPeriod !== "all") {
            const now = new Date();
            filtered = filtered.filter((doc) => {
                const created = new Date(doc.created_at);
                switch (selectedPeriod) {
                    case "today":
                        return created.toDateString() === now.toDateString();
                    case "week": {
                        const weekAgo = new Date(
                            now.getTime() - 7 * 24 * 60 * 60 * 1000,
                        );
                        return created >= weekAgo;
                    }
                    case "month":
                        return (
                            created.getMonth() === now.getMonth() &&
                            created.getFullYear() === now.getFullYear()
                        );
                    case "year":
                        return created.getFullYear() === now.getFullYear();
                    default:
                        return true;
                }
            });
        }

        // Filter by type
        if (selectedType !== "all") {
            filtered = filtered.filter(
                (d) => d.jenis_file.toLowerCase() === selectedType,
            );
        }

        return filtered;
    }

    $: selectedPeriod, selectedType, calculateStats();

    function calculateStats() {
        const filtered = getFilteredDocuments();
        totalDocs = filtered.length;
        publishedDocs = filtered.filter((d) => d.status === "publish").length;
        draftDocs = filtered.filter((d) => d.status === "draft").length;
        totalViews = filtered.reduce((sum, d) => sum + (d.view_count || 0), 0);

        // Group by type
        docsByType = filtered.reduce((acc, doc) => {
            const type = doc.jenis_file
                ? doc.jenis_file.charAt(0).toUpperCase() +
                  doc.jenis_file.slice(1).toLowerCase()
                : "Lainnya";
            acc[type] = (acc[type] || 0) + 1;
            return acc;
        }, {});

        // Group by month (last 6 months)
        const months = [];
        for (let i = 5; i >= 0; i--) {
            const date = new Date();
            date.setMonth(date.getMonth() - i);
            const monthName = date.toLocaleDateString("id-ID", {
                month: "short",
            });

            const count = filtered.filter((doc) => {
                const docDate = new Date(doc.created_at);
                return (
                    docDate.getFullYear() === date.getFullYear() &&
                    docDate.getMonth() === date.getMonth()
                );
            }).length;

            months.push({ name: monthName, count });
        }
        docsByMonth = months;
    }

    function buildActivityLog() {
        // Build real activity log from documents data
        const activities = [];

        // Sort docs by creation date (newest first)
        const sorted = [...documents].sort(
            (a, b) =>
                new Date(b.created_at).getTime() -
                new Date(a.created_at).getTime(),
        );

        sorted.slice(0, 10).forEach((doc) => {
            activities.push({
                type: doc.status === "publish" ? "publish" : "upload",
                icon: doc.status === "publish" ? "publish" : "upload_file",
                color: doc.status === "publish" ? "emerald" : "blue",
                title:
                    doc.status === "publish"
                        ? "Dokumen dipublikasikan"
                        : "Dokumen diupload",
                description: doc.judul,
                author: doc.penulis,
                time: new Date(doc.created_at),
            });
        });

        recentActivity = activities;
    }

    function formatTimeAgo(date) {
        const now = new Date();
        const diff = now.getTime() - date.getTime();
        const minutes = Math.floor(diff / 60000);
        const hours = Math.floor(diff / 3600000);
        const days = Math.floor(diff / 86400000);

        if (minutes < 1) return "Baru saja";
        if (minutes < 60) return `${minutes} menit lalu`;
        if (hours < 24) return `${hours} jam lalu`;
        if (days < 7) return `${days} hari lalu`;
        return date.toLocaleDateString("id-ID", {
            day: "numeric",
            month: "short",
            year: "numeric",
        });
    }

    function getMaxCount() {
        return Math.max(...docsByMonth.map((m) => m.count), 1);
    }

    function getTypeColor(type) {
        const t = type.toLowerCase();
        const colors = {
            skripsi: {
                bg: "bg-blue-500",
                text: "text-blue-500",
                light: "bg-blue-100 dark:bg-blue-900/30",
            },
            tesis: {
                bg: "bg-purple-500",
                text: "text-purple-500",
                light: "bg-purple-100 dark:bg-purple-900/30",
            },
            jurnal: {
                bg: "bg-teal-500",
                text: "text-teal-500",
                light: "bg-teal-100 dark:bg-teal-900/30",
            },
            disertasi: {
                bg: "bg-amber-500",
                text: "text-amber-500",
                light: "bg-amber-100 dark:bg-amber-900/30",
            },
        };
        return (
            colors[t] || {
                bg: "bg-slate-500",
                text: "text-slate-500",
                light: "bg-slate-100 dark:bg-slate-800",
            }
        );
    }

    function getTypeIcon(type) {
        const t = type.toLowerCase();
        const icons = {
            skripsi: "school",
            tesis: "history_edu",
            jurnal: "article",
            disertasi: "workspace_premium",
        };
        return icons[t] || "description";
    }

    // ============================================
    // EXPORT FUNCTIONS
    // ============================================

    async function exportToExcel() {
        exportLoading = true;
        showExportMenu = false;
        try {
            const XLSX = await import("xlsx");
            const filtered = getFilteredDocuments();

            // Sheet 1: Ringkasan
            const summaryData = [
                ["LAPORAN DOKUMEN REPOSITORY"],
                [
                    "Tanggal Export",
                    new Date().toLocaleDateString("id-ID", {
                        day: "numeric",
                        month: "long",
                        year: "numeric",
                        hour: "2-digit",
                        minute: "2-digit",
                    }),
                ],
                ["Periode", getPeriodLabel()],
                [""],
                ["RINGKASAN"],
                ["Total Dokumen", totalDocs],
                ["Dipublikasi", publishedDocs],
                ["Draft", draftDocs],
                ["Total Views", totalViews],
                [""],
                ["DISTRIBUSI TIPE"],
                ...Object.entries(docsByType).map(([type, count]) => [
                    type,
                    count,
                    `${Math.round((count / totalDocs) * 100)}%`,
                ]),
                [""],
                ["TREN UPLOAD (6 Bulan Terakhir)"],
                ...docsByMonth.map((m) => [m.name, m.count]),
            ];

            // Sheet 2: Detail Dokumen
            const detailData = [
                [
                    "No",
                    "Judul",
                    "Penulis",
                    "Jenis",
                    "Status",
                    "Fakultas",
                    "Prodi",
                    "Dosen Pembimbing",
                    "Views",
                    "Tanggal Upload",
                ],
                ...filtered.map((doc, i) => [
                    i + 1,
                    doc.judul,
                    doc.penulis,
                    doc.jenis_file,
                    doc.status === "publish" ? "Dipublikasi" : "Draft",
                    doc.fakultas_nama || "-",
                    doc.prodi_nama || "-",
                    doc.dosen_pembimbing || "-",
                    doc.view_count || 0,
                    new Date(doc.created_at).toLocaleDateString("id-ID"),
                ]),
            ];

            const wb = XLSX.utils.book_new();

            const ws1 = XLSX.utils.aoa_to_sheet(summaryData);
            ws1["!cols"] = [{ wch: 25 }, { wch: 40 }, { wch: 15 }];
            XLSX.utils.book_append_sheet(wb, ws1, "Ringkasan");

            const ws2 = XLSX.utils.aoa_to_sheet(detailData);
            ws2["!cols"] = [
                { wch: 5 },
                { wch: 40 },
                { wch: 25 },
                { wch: 12 },
                { wch: 12 },
                { wch: 20 },
                { wch: 20 },
                { wch: 25 },
                { wch: 8 },
                { wch: 15 },
            ];
            XLSX.utils.book_append_sheet(wb, ws2, "Detail Dokumen");

            const fileName = `Laporan_Repository_${new Date().toISOString().slice(0, 10)}.xlsx`;
            XLSX.writeFile(wb, fileName);
        } catch (e) {
            console.error("Export Excel error:", e);
            alert("Gagal export Excel: " + e.message);
        } finally {
            exportLoading = false;
        }
    }

    async function exportToPDF() {
        exportLoading = true;
        showExportMenu = false;
        try {
            const { default: jsPDF } = await import("jspdf");
            const autoTableModule = await import("jspdf-autotable");
            const autoTable =
                autoTableModule.default || autoTableModule.autoTable;

            const pdf = new jsPDF("p", "mm", "a4");
            const filtered = getFilteredDocuments();
            const pageWidth = pdf.internal.pageSize.getWidth();

            // Header
            pdf.setFontSize(18);
            pdf.setFont("helvetica", "bold");
            pdf.text("LAPORAN DOKUMEN REPOSITORY", pageWidth / 2, 20, {
                align: "center",
            });

            pdf.setFontSize(10);
            pdf.setFont("helvetica", "normal");
            pdf.setTextColor(100);
            pdf.text(
                `Digenerate: ${new Date().toLocaleDateString("id-ID", { day: "numeric", month: "long", year: "numeric", hour: "2-digit", minute: "2-digit" })}`,
                pageWidth / 2,
                28,
                { align: "center" },
            );
            pdf.text(`Periode: ${getPeriodLabel()}`, pageWidth / 2, 33, {
                align: "center",
            });

            // Summary boxes
            pdf.setDrawColor(200);
            pdf.setFillColor(240, 245, 255);
            pdf.roundedRect(14, 40, 42, 25, 2, 2, "FD");
            pdf.roundedRect(60, 40, 42, 25, 2, 2, "FD");
            pdf.roundedRect(106, 40, 42, 25, 2, 2, "FD");
            pdf.roundedRect(152, 40, 42, 25, 2, 2, "FD");

            pdf.setTextColor(100);
            pdf.setFontSize(8);
            pdf.text("Total Dokumen", 35, 47, { align: "center" });
            pdf.text("Dipublikasi", 81, 47, { align: "center" });
            pdf.text("Draft", 127, 47, { align: "center" });
            pdf.text("Total Views", 173, 47, { align: "center" });

            pdf.setTextColor(30);
            pdf.setFontSize(16);
            pdf.setFont("helvetica", "bold");
            pdf.text(String(totalDocs), 35, 58, { align: "center" });
            pdf.text(String(publishedDocs), 81, 58, { align: "center" });
            pdf.text(String(draftDocs), 127, 58, { align: "center" });
            pdf.text(String(totalViews), 173, 58, { align: "center" });

            // Distribusi Tipe
            pdf.setFontSize(12);
            pdf.setFont("helvetica", "bold");
            pdf.setTextColor(30);
            pdf.text("Distribusi Tipe Dokumen", 14, 78);

            const typeData = Object.entries(docsByType).map(([type, count]) => [
                type,
                count,
                `${totalDocs > 0 ? Math.round((Number(count) / totalDocs) * 100) : 0}%`,
            ]);

            autoTable(pdf, {
                startY: 82,
                head: [["Tipe", "Jumlah", "Persentase"]],
                body: typeData,
                theme: "grid",
                headStyles: { fillColor: [17, 115, 212], fontSize: 9 },
                bodyStyles: { fontSize: 9 },
                margin: { left: 14, right: 14 },
                tableWidth: 80,
            });

            // Tren Upload
            const trendStartY =
                /** @type {any} */ (pdf).previousAutoTable.finalY + 12;
            pdf.setFontSize(12);
            pdf.setFont("helvetica", "bold");
            pdf.text("Tren Upload (6 Bulan Terakhir)", 14, trendStartY);

            autoTable(pdf, {
                startY: trendStartY + 4,
                head: [["Bulan", "Jumlah Upload"]],
                body: docsByMonth.map((m) => [m.name, m.count]),
                theme: "grid",
                headStyles: { fillColor: [17, 115, 212], fontSize: 9 },
                bodyStyles: { fontSize: 9 },
                margin: { left: 14, right: 14 },
                tableWidth: 80,
            });

            // Detail Dokumen (new page)
            pdf.addPage();
            pdf.setFontSize(12);
            pdf.setFont("helvetica", "bold");
            pdf.setTextColor(30);
            pdf.text("Detail Dokumen", 14, 15);

            autoTable(pdf, {
                startY: 20,
                head: [
                    [
                        "No",
                        "Judul",
                        "Penulis",
                        "Jenis",
                        "Status",
                        "Views",
                        "Tanggal",
                    ],
                ],
                body: filtered.map((d, i) => [
                    i + 1,
                    d.judul.length > 35
                        ? d.judul.substring(0, 35) + "..."
                        : d.judul,
                    d.penulis,
                    d.jenis_file,
                    d.status === "publish" ? "Publik" : "Draft",
                    d.view_count || 0,
                    new Date(d.created_at).toLocaleDateString("id-ID"),
                ]),
                theme: "grid",
                headStyles: { fillColor: [17, 115, 212], fontSize: 8 },
                bodyStyles: { fontSize: 7 },
                columnStyles: {
                    0: { cellWidth: 10 },
                    1: { cellWidth: 55 },
                    2: { cellWidth: 30 },
                    3: { cellWidth: 20 },
                    4: { cellWidth: 15 },
                    5: { cellWidth: 15 },
                    6: { cellWidth: 22 },
                },
                margin: { left: 14, right: 14 },
            });

            // Footer on each page
            const totalPages = pdf.getNumberOfPages();
            for (let i = 1; i <= totalPages; i++) {
                pdf.setPage(i);
                pdf.setFontSize(8);
                pdf.setFont("helvetica", "normal");
                pdf.setTextColor(150);
                pdf.text(
                    `Halaman ${i} dari ${totalPages}`,
                    pageWidth / 2,
                    pdf.internal.pageSize.getHeight() - 10,
                    { align: "center" },
                );
            }

            const fileName = `Laporan_Repository_${new Date().toISOString().slice(0, 10)}.pdf`;
            pdf.save(fileName);
        } catch (e) {
            console.error("Export PDF error:", e);
            alert("Gagal export PDF: " + e.message);
        } finally {
            exportLoading = false;
        }
    }

    function getPeriodLabel() {
        const labels = {
            all: "Semua Waktu",
            today: "Hari Ini",
            week: "Minggu Ini",
            month: "Bulan Ini",
            year: "Tahun Ini",
        };
        return labels[selectedPeriod] || "Semua Waktu";
    }

    function handleManualRefresh() {
        loadData(true);
    }

    function handleClickOutside(e) {
        if (showExportMenu && !e.target.closest(".export-menu-container")) {
            showExportMenu = false;
        }
    }
</script>

<svelte:window on:click={handleClickOutside} />

<div class="max-w-7xl mx-auto flex flex-col gap-6">
    <!-- Header -->
    <div class="flex flex-wrap items-end justify-between gap-4">
        <div class="flex flex-col gap-1">
            <h2
                class="text-3xl font-black text-slate-900 dark:text-white tracking-tight"
            >
                Laporan & Statistik
            </h2>
            <div class="flex items-center gap-3">
                <p class="text-slate-500 dark:text-slate-400">
                    Analisis dan ringkasan data dokumen sistem.
                </p>
                {#if lastUpdated}
                    <div
                        class="flex items-center gap-1.5 text-xs text-slate-400 dark:text-slate-500"
                    >
                        <span class="relative flex h-2 w-2">
                            <span
                                class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"
                            ></span>
                            <span
                                class="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"
                            ></span>
                        </span>
                        <span>Live</span>
                        <span class="text-slate-300 dark:text-slate-600">•</span
                        >
                        <span>Refresh dalam {nextRefreshIn}s</span>
                    </div>
                {/if}
            </div>
        </div>
        <div class="flex gap-3">
            <!-- Refresh button -->
            <button
                on:click={handleManualRefresh}
                disabled={isRefreshing}
                class="flex items-center gap-2 px-4 py-2.5 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300 font-medium rounded-lg hover:bg-slate-50 dark:hover:bg-slate-700 transition-all disabled:opacity-50"
                title="Refresh data"
            >
                <span
                    class="material-symbols-outlined text-xl {isRefreshing
                        ? 'animate-spin'
                        : ''}">refresh</span
                >
                <span class="hidden sm:inline">Refresh</span>
            </button>

            <!-- Print -->
            <button
                on:click={() => window.print()}
                class="flex items-center gap-2 px-4 py-2.5 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300 font-medium rounded-lg hover:bg-slate-50 dark:hover:bg-slate-700 transition-all"
            >
                <span class="material-symbols-outlined text-xl">print</span>
                <span class="hidden sm:inline">Cetak</span>
            </button>

            <!-- Export dropdown -->
            <div class="relative export-menu-container">
                <button
                    on:click|stopPropagation={() =>
                        (showExportMenu = !showExportMenu)}
                    disabled={exportLoading}
                    class="flex items-center gap-2 px-5 py-2.5 bg-primary text-white font-bold rounded-lg shadow-lg shadow-primary/25 hover:bg-primary/90 transition-all active:scale-95 disabled:opacity-50"
                >
                    {#if exportLoading}
                        <span
                            class="material-symbols-outlined text-xl animate-spin"
                            >progress_activity</span
                        >
                        <span>Exporting...</span>
                    {:else}
                        <span class="material-symbols-outlined text-xl"
                            >download</span
                        >
                        <span>Export</span>
                        <span class="material-symbols-outlined text-lg"
                            >expand_more</span
                        >
                    {/if}
                </button>
                {#if showExportMenu}
                    <div
                        class="absolute right-0 top-full mt-2 w-56 bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 shadow-xl z-50 overflow-hidden animate-fade-in"
                    >
                        <button
                            on:click={exportToExcel}
                            class="w-full flex items-center gap-3 px-4 py-3 text-left hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors"
                        >
                            <div
                                class="w-9 h-9 rounded-lg bg-emerald-100 dark:bg-emerald-900/30 flex items-center justify-center"
                            >
                                <span
                                    class="material-symbols-outlined text-emerald-600"
                                    style="font-size: 20px;">table_view</span
                                >
                            </div>
                            <div>
                                <p
                                    class="font-semibold text-sm text-slate-900 dark:text-white"
                                >
                                    Export Excel
                                </p>
                                <p class="text-xs text-slate-500">
                                    .xlsx — Microsoft Excel
                                </p>
                            </div>
                        </button>
                        <div
                            class="border-t border-slate-100 dark:border-slate-700"
                        ></div>
                        <button
                            on:click={exportToPDF}
                            class="w-full flex items-center gap-3 px-4 py-3 text-left hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors"
                        >
                            <div
                                class="w-9 h-9 rounded-lg bg-red-100 dark:bg-red-900/30 flex items-center justify-center"
                            >
                                <span
                                    class="material-symbols-outlined text-red-600"
                                    style="font-size: 20px;"
                                    >picture_as_pdf</span
                                >
                            </div>
                            <div>
                                <p
                                    class="font-semibold text-sm text-slate-900 dark:text-white"
                                >
                                    Export PDF
                                </p>
                                <p class="text-xs text-slate-500">
                                    .pdf — Portable Document
                                </p>
                            </div>
                        </button>
                    </div>
                {/if}
            </div>
        </div>
    </div>

    <!-- Filter Bar -->
    <div
        class="bg-white dark:bg-slate-900 p-4 rounded-xl border border-slate-200 dark:border-slate-800 shadow-sm flex flex-wrap gap-4 items-center"
    >
        <div class="flex items-center gap-2">
            <span class="material-symbols-outlined text-slate-400"
                >filter_alt</span
            >
            <span class="text-sm font-medium text-slate-600 dark:text-slate-400"
                >Filter:</span
            >
        </div>
        <select
            bind:value={selectedPeriod}
            class="px-4 py-2 bg-slate-100 dark:bg-slate-800 border-none rounded-lg text-sm font-medium text-slate-700 dark:text-slate-300"
        >
            <option value="all">Semua Waktu</option>
            <option value="today">Hari Ini</option>
            <option value="week">Minggu Ini</option>
            <option value="month">Bulan Ini</option>
            <option value="year">Tahun Ini</option>
        </select>
        <select
            bind:value={selectedType}
            class="px-4 py-2 bg-slate-100 dark:bg-slate-800 border-none rounded-lg text-sm font-medium text-slate-700 dark:text-slate-300"
        >
            <option value="all">Semua Tipe</option>
            <option value="skripsi">Skripsi</option>
            <option value="tesis">Tesis</option>
            <option value="jurnal">Jurnal</option>
        </select>

        {#if isRefreshing}
            <div class="ml-auto flex items-center gap-2 text-xs text-primary">
                <span class="material-symbols-outlined text-base animate-spin"
                    >progress_activity</span
                >
                Memperbarui data...
            </div>
        {/if}
    </div>

    <!-- Summary Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">
        <!-- Total Documents -->
        <div
            class="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 p-6 shadow-sm hover:shadow-md transition-shadow"
        >
            <div class="flex items-center justify-between mb-4">
                <span class="text-sm font-medium text-slate-500"
                    >Total Dokumen</span
                >
                <div class="p-2 bg-blue-100 dark:bg-blue-900/30 rounded-lg">
                    <span class="material-symbols-outlined text-blue-600"
                        >folder</span
                    >
                </div>
            </div>
            {#if loading}
                <div
                    class="h-10 w-20 bg-slate-200 dark:bg-slate-700 rounded animate-pulse"
                ></div>
            {:else}
                <div class="text-4xl font-black text-slate-900 dark:text-white">
                    {totalDocs}
                </div>
            {/if}
            <p class="text-sm text-slate-500 mt-2">dokumen tersimpan</p>
        </div>

        <!-- Published -->
        <div
            class="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 p-6 shadow-sm hover:shadow-md transition-shadow"
        >
            <div class="flex items-center justify-between mb-4">
                <span class="text-sm font-medium text-slate-500"
                    >Dipublikasi</span
                >
                <div
                    class="p-2 bg-emerald-100 dark:bg-emerald-900/30 rounded-lg"
                >
                    <span class="material-symbols-outlined text-emerald-600"
                        >check_circle</span
                    >
                </div>
            </div>
            {#if loading}
                <div
                    class="h-10 w-20 bg-slate-200 dark:bg-slate-700 rounded animate-pulse"
                ></div>
            {:else}
                <div class="text-4xl font-black text-slate-900 dark:text-white">
                    {publishedDocs}
                </div>
            {/if}
            <p class="text-sm text-slate-500 mt-2">
                {totalDocs > 0
                    ? Math.round((publishedDocs / totalDocs) * 100)
                    : 0}% dari total
            </p>
        </div>

        <!-- Draft -->
        <div
            class="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 p-6 shadow-sm hover:shadow-md transition-shadow"
        >
            <div class="flex items-center justify-between mb-4">
                <span class="text-sm font-medium text-slate-500">Draft</span>
                <div class="p-2 bg-amber-100 dark:bg-amber-900/30 rounded-lg">
                    <span class="material-symbols-outlined text-amber-600"
                        >edit_note</span
                    >
                </div>
            </div>
            {#if loading}
                <div
                    class="h-10 w-20 bg-slate-200 dark:bg-slate-700 rounded animate-pulse"
                ></div>
            {:else}
                <div class="text-4xl font-black text-slate-900 dark:text-white">
                    {draftDocs}
                </div>
            {/if}
            <p class="text-sm text-slate-500 mt-2">menunggu publikasi</p>
        </div>

        <!-- Total Views -->
        <div
            class="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 p-6 shadow-sm hover:shadow-md transition-shadow"
        >
            <div class="flex items-center justify-between mb-4">
                <span class="text-sm font-medium text-slate-500"
                    >Total Views</span
                >
                <div class="p-2 bg-purple-100 dark:bg-purple-900/30 rounded-lg">
                    <span class="material-symbols-outlined text-purple-600"
                        >visibility</span
                    >
                </div>
            </div>
            {#if loading}
                <div
                    class="h-10 w-20 bg-slate-200 dark:bg-slate-700 rounded animate-pulse"
                ></div>
            {:else}
                <div class="text-4xl font-black text-slate-900 dark:text-white">
                    {totalViews.toLocaleString("id-ID")}
                </div>
            {/if}
            <p class="text-sm text-slate-500 mt-2">total kunjungan</p>
        </div>
    </div>

    <!-- Charts Row -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- Monthly Chart -->
        <div
            class="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 p-6 shadow-sm"
        >
            <h3 class="text-lg font-bold text-slate-900 dark:text-white mb-6">
                Tren Upload Dokumen
            </h3>
            <div class="flex items-end gap-4 h-48">
                {#if loading}
                    {#each Array(6) as _}
                        <div class="flex-1 flex flex-col items-center gap-2">
                            <div
                                class="w-full bg-slate-200 dark:bg-slate-700 rounded-t animate-pulse"
                                style="height: 60%"
                            ></div>
                            <div
                                class="h-4 w-8 bg-slate-200 dark:bg-slate-700 rounded animate-pulse"
                            ></div>
                        </div>
                    {/each}
                {:else}
                    {#each docsByMonth as month, i}
                        <div
                            class="flex-1 flex flex-col items-center gap-2 group"
                        >
                            <div class="w-full flex flex-col justify-end h-40">
                                <div
                                    class="w-full bg-gradient-to-t from-primary to-blue-400 rounded-t transition-all duration-500 group-hover:from-primary/90 group-hover:to-blue-300"
                                    style="height: {(month.count /
                                        getMaxCount()) *
                                        100}%; min-height: 4px; animation: barGrow 0.6s ease-out {i *
                                        0.1}s both;"
                                ></div>
                            </div>
                            <span class="text-xs font-medium text-slate-500"
                                >{month.name}</span
                            >
                            <span
                                class="text-sm font-bold text-slate-900 dark:text-white"
                                >{month.count}</span
                            >
                        </div>
                    {/each}
                {/if}
            </div>
        </div>

        <!-- Document Types -->
        <div
            class="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 p-6 shadow-sm"
        >
            <h3 class="text-lg font-bold text-slate-900 dark:text-white mb-6">
                Distribusi Tipe Dokumen
            </h3>
            {#if loading}
                <div class="space-y-4">
                    {#each Array(4) as _}
                        <div class="flex items-center gap-4">
                            <div
                                class="w-12 h-12 bg-slate-200 dark:bg-slate-700 rounded-lg animate-pulse"
                            ></div>
                            <div
                                class="flex-1 h-4 bg-slate-200 dark:bg-slate-700 rounded animate-pulse"
                            ></div>
                        </div>
                    {/each}
                </div>
            {:else if Object.keys(docsByType).length === 0}
                <div
                    class="flex flex-col items-center justify-center h-48 text-slate-400"
                >
                    <span class="material-symbols-outlined text-4xl mb-2"
                        >pie_chart</span
                    >
                    <p>Belum ada data</p>
                </div>
            {:else}
                <div class="space-y-4">
                    {#each Object.entries(docsByType) as [type, count]}
                        <div class="flex items-center gap-4 group">
                            <div
                                class="{getTypeColor(type)
                                    .light} p-3 rounded-lg group-hover:scale-110 transition-transform"
                            >
                                <span
                                    class="material-symbols-outlined {getTypeColor(
                                        type,
                                    ).text}">{getTypeIcon(type)}</span
                                >
                            </div>
                            <div class="flex-1">
                                <div class="flex justify-between mb-1">
                                    <span
                                        class="font-semibold text-slate-900 dark:text-white"
                                        >{type}</span
                                    >
                                    <span class="text-sm text-slate-500"
                                        >{count} dokumen</span
                                    >
                                </div>
                                <div
                                    class="h-2 bg-slate-100 dark:bg-slate-800 rounded-full overflow-hidden"
                                >
                                    <div
                                        class="{getTypeColor(type)
                                            .bg} h-full rounded-full transition-all duration-700"
                                        style="width: {(count / totalDocs) *
                                            100}%"
                                    ></div>
                                </div>
                            </div>
                            <span
                                class="text-sm font-bold text-slate-600 dark:text-slate-400 min-w-[40px] text-right"
                            >
                                {Math.round((count / totalDocs) * 100)}%
                            </span>
                        </div>
                    {/each}
                </div>
            {/if}
        </div>
    </div>

    <!-- Activity Log -->
    <div
        class="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden"
    >
        <div
            class="p-6 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between"
        >
            <div>
                <h3 class="text-lg font-bold text-slate-900 dark:text-white">
                    Aktivitas Terbaru
                </h3>
                <p class="text-sm text-slate-500">
                    Riwayat dokumen terbaru berdasarkan data real-time
                </p>
            </div>
            {#if lastUpdated}
                <span class="text-xs text-slate-400">
                    Update: {lastUpdated.toLocaleTimeString("id-ID")}
                </span>
            {/if}
        </div>
        <div
            class="divide-y divide-slate-100 dark:divide-slate-800 max-h-[400px] overflow-y-auto"
        >
            {#if loading}
                {#each Array(5) as _}
                    <div class="p-4 flex items-center gap-4">
                        <div
                            class="w-10 h-10 bg-slate-200 dark:bg-slate-700 rounded-full animate-pulse"
                        ></div>
                        <div class="flex-1 space-y-2">
                            <div
                                class="h-4 w-48 bg-slate-200 dark:bg-slate-700 rounded animate-pulse"
                            ></div>
                            <div
                                class="h-3 w-32 bg-slate-100 dark:bg-slate-800 rounded animate-pulse"
                            ></div>
                        </div>
                    </div>
                {/each}
            {:else if recentActivity.length === 0}
                <div class="p-8 text-center text-slate-400">
                    <span class="material-symbols-outlined text-4xl mb-2"
                        >inbox</span
                    >
                    <p>Belum ada aktivitas</p>
                </div>
            {:else}
                {#each recentActivity as activity}
                    <div
                        class="p-4 flex items-center gap-4 hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors"
                    >
                        <div
                            class="w-10 h-10 rounded-full bg-{activity.color}-100 dark:bg-{activity.color}-900/30 flex items-center justify-center flex-shrink-0"
                        >
                            <span
                                class="material-symbols-outlined text-{activity.color}-600"
                                style="font-size: 20px;">{activity.icon}</span
                            >
                        </div>
                        <div class="flex-1 min-w-0">
                            <p
                                class="font-medium text-slate-900 dark:text-white text-sm"
                            >
                                {activity.title}
                            </p>
                            <p
                                class="text-sm text-slate-500 truncate"
                                title={activity.description}
                            >
                                {activity.description} — oleh {activity.author}
                            </p>
                        </div>
                        <span class="text-xs text-slate-400 flex-shrink-0">
                            {formatTimeAgo(activity.time)}
                        </span>
                    </div>
                {/each}
            {/if}
        </div>
    </div>
</div>

<style>
    @keyframes barGrow {
        from {
            height: 0;
        }
    }
    @keyframes fade-in {
        from {
            opacity: 0;
            transform: translateY(-4px);
        }
        to {
            opacity: 1;
            transform: translateY(0);
        }
    }
    .animate-fade-in {
        animation: fade-in 0.15s ease-out;
    }
</style>
