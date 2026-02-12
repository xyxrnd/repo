<script>
    import { onMount } from "svelte";
    import { link } from "svelte-spa-router";
    import prodiService from "../../services/prodiService.js";
    import fakultasService from "../../services/fakultasService.js";

    let prodiList = [];
    let fakultasList = [];
    let loading = true;
    let error = "";
    let searchQuery = "";
    let filterFakultas = "all";

    // Modal state
    let showModal = false;
    let modalMode = "add"; // 'add' or 'edit'
    let editId = "";
    let formNama = "";
    let formKode = "";
    let formFakultasId = "";
    let formLoading = false;
    let formError = "";

    onMount(async () => {
        await loadData();
    });

    async function loadData() {
        try {
            loading = true;
            error = "";
            [prodiList, fakultasList] = await Promise.all([
                prodiService.getAll(),
                fakultasService.getAll(),
            ]);
        } catch (e) {
            error = e.message;
        } finally {
            loading = false;
        }
    }

    function openAddModal() {
        modalMode = "add";
        editId = "";
        formNama = "";
        formKode = "";
        formFakultasId = fakultasList.length > 0 ? fakultasList[0].id : "";
        formError = "";
        showModal = true;
    }

    function openEditModal(item) {
        modalMode = "edit";
        editId = item.id;
        formNama = item.nama;
        formKode = item.kode;
        formFakultasId = item.fakultas_id;
        formError = "";
        showModal = true;
    }

    function closeModal() {
        showModal = false;
        formError = "";
    }

    async function handleSubmit() {
        formError = "";

        if (!formNama || !formKode || !formFakultasId) {
            formError = "Semua field wajib diisi.";
            return;
        }

        formLoading = true;

        try {
            const data = {
                nama: formNama,
                kode: formKode,
                fakultas_id: formFakultasId,
            };

            if (modalMode === "add") {
                await prodiService.create(data);
            } else {
                await prodiService.update(editId, data);
            }
            closeModal();
            await loadData();
        } catch (e) {
            formError = e.message;
        } finally {
            formLoading = false;
        }
    }

    async function handleDelete(id, nama) {
        if (!confirm(`Apakah Anda yakin ingin menghapus prodi "${nama}"?`))
            return;

        try {
            await prodiService.delete(id);
            await loadData();
        } catch (e) {
            alert("Gagal menghapus: " + e.message);
        }
    }

    $: filteredList = prodiList.filter((p) => {
        const matchesSearch =
            p.nama.toLowerCase().includes(searchQuery.toLowerCase()) ||
            p.kode.toLowerCase().includes(searchQuery.toLowerCase());
        const matchesFakultas =
            filterFakultas === "all" || p.fakultas_id === filterFakultas;
        return matchesSearch && matchesFakultas;
    });

    function formatDate(dateString) {
        const date = new Date(dateString);
        return date.toLocaleDateString("id-ID", {
            day: "numeric",
            month: "short",
            year: "numeric",
        });
    }
</script>

<div class="max-w-7xl mx-auto flex flex-col gap-6">
    <!-- Header -->
    <div class="flex flex-wrap items-end justify-between gap-4">
        <div class="flex flex-col gap-1">
            <h2
                class="text-3xl font-black text-slate-900 dark:text-white tracking-tight"
            >
                Kelola Program Studi
            </h2>
            <p class="text-slate-500 dark:text-slate-400">
                Kelola data program studi yang tersedia di sistem.
            </p>
        </div>
        <button
            on:click={openAddModal}
            class="flex items-center gap-2 px-5 py-2.5 bg-primary text-white font-bold rounded-lg shadow-lg shadow-primary/25 hover:bg-primary/90 transition-all active:scale-95"
        >
            <span class="material-symbols-outlined text-xl">add</span>
            <span>Tambah Prodi</span>
        </button>
    </div>

    <!-- Search & Filter -->
    <div
        class="bg-white dark:bg-slate-900 p-4 rounded-xl border border-slate-200 dark:border-slate-800 shadow-sm flex flex-col md:flex-row gap-4 items-center"
    >
        <div class="relative w-full md:w-96">
            <span
                class="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 text-xl"
                >search</span
            >
            <input
                bind:value={searchQuery}
                class="w-full pl-10 pr-4 py-2 bg-slate-100 dark:bg-slate-800 border-none rounded-lg focus:ring-2 focus:ring-primary/50 text-sm transition-all"
                placeholder="Cari nama atau kode prodi..."
                type="text"
            />
        </div>
        <select
            bind:value={filterFakultas}
            class="px-4 py-2 bg-slate-100 dark:bg-slate-800 border-none rounded-lg text-sm font-medium"
        >
            <option value="all">Semua Fakultas</option>
            {#each fakultasList as fak}
                <option value={fak.id}>{fak.nama}</option>
            {/each}
        </select>
    </div>

    {#if error}
        <div
            class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-xl p-4 text-red-700 dark:text-red-400"
        >
            <p class="font-semibold">Error:</p>
            <p>{error}</p>
        </div>
    {/if}

    {#if loading}
        <div class="flex justify-center items-center py-20">
            <div
                class="animate-spin rounded-full h-16 w-16 border-4 border-primary border-t-transparent"
            ></div>
        </div>
    {:else if filteredList.length === 0}
        <div
            class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-12 text-center"
        >
            <span
                class="material-symbols-outlined text-6xl text-slate-300 dark:text-slate-700 mb-4"
                >local_library</span
            >
            <h3
                class="text-xl font-bold text-slate-700 dark:text-slate-300 mb-2"
            >
                Belum ada program studi
            </h3>
            <p class="text-slate-500 dark:text-slate-400 mb-6">
                {searchQuery || filterFakultas !== "all"
                    ? "Tidak ada prodi yang sesuai dengan filter"
                    : "Mulai dengan menambahkan program studi baru"}
            </p>
            {#if !searchQuery && filterFakultas === "all"}
                <button
                    on:click={openAddModal}
                    class="inline-flex items-center gap-2 px-5 py-2.5 bg-primary text-white font-bold rounded-lg shadow-lg shadow-primary/25 hover:bg-primary/90 transition-all"
                >
                    <span class="material-symbols-outlined">add</span>
                    <span>Tambah Prodi Pertama</span>
                </button>
            {/if}
        </div>
    {:else}
        <!-- Data Table -->
        <div
            class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden"
        >
            <div class="overflow-x-auto">
                <table class="w-full text-left border-collapse">
                    <thead>
                        <tr
                            class="bg-slate-50 dark:bg-slate-800/50 border-b border-slate-200 dark:border-slate-800"
                        >
                            <th
                                class="px-6 py-4 text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider"
                            >
                                No
                            </th>
                            <th
                                class="px-6 py-4 text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider"
                            >
                                Kode
                            </th>
                            <th
                                class="px-6 py-4 text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider"
                            >
                                Nama Program Studi
                            </th>
                            <th
                                class="px-6 py-4 text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider"
                            >
                                Fakultas
                            </th>
                            <th
                                class="px-6 py-4 text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider"
                            >
                                Tanggal Dibuat
                            </th>
                            <th
                                class="px-6 py-4 text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider text-right"
                            >
                                Aksi
                            </th>
                        </tr>
                    </thead>
                    <tbody
                        class="divide-y divide-slate-100 dark:divide-slate-800"
                    >
                        {#each filteredList as item, i (item.id)}
                            <tr
                                class="hover:bg-slate-50/80 dark:hover:bg-slate-800/30 transition-colors"
                            >
                                <td class="px-6 py-4 text-sm text-slate-500">
                                    {i + 1}
                                </td>
                                <td class="px-6 py-4">
                                    <span
                                        class="px-2.5 py-1 bg-primary/10 text-primary text-xs font-bold rounded"
                                    >
                                        {item.kode}
                                    </span>
                                </td>
                                <td class="px-6 py-4">
                                    <div class="flex items-center gap-3">
                                        <div
                                            class="bg-teal-50 dark:bg-teal-900/20 text-teal-600 p-2 rounded-lg"
                                        >
                                            <span
                                                class="material-symbols-outlined"
                                                >local_library</span
                                            >
                                        </div>
                                        <span
                                            class="text-sm font-bold text-slate-900 dark:text-white"
                                            >{item.nama}</span
                                        >
                                    </div>
                                </td>
                                <td class="px-6 py-4">
                                    <span
                                        class="px-2.5 py-1 bg-indigo-100 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-400 text-xs font-bold rounded"
                                    >
                                        {item.fakultas_nama || "-"}
                                    </span>
                                </td>
                                <td
                                    class="px-6 py-4 text-sm text-slate-600 dark:text-slate-400"
                                >
                                    {formatDate(item.created_at)}
                                </td>
                                <td class="px-6 py-4 text-right">
                                    <div class="flex justify-end gap-2">
                                        <button
                                            on:click={() => openEditModal(item)}
                                            class="p-1.5 text-primary hover:bg-primary/10 rounded transition-colors"
                                            title="Edit"
                                        >
                                            <span
                                                class="material-symbols-outlined text-xl"
                                                >edit_square</span
                                            >
                                        </button>
                                        <button
                                            on:click={() =>
                                                handleDelete(
                                                    item.id,
                                                    item.nama,
                                                )}
                                            class="p-1.5 text-slate-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 rounded transition-colors"
                                            title="Hapus"
                                        >
                                            <span
                                                class="material-symbols-outlined text-xl"
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

            <div
                class="px-6 py-4 bg-slate-50 dark:bg-slate-800/50 flex items-center justify-between"
            >
                <span class="text-sm text-slate-500">
                    Menampilkan {filteredList.length} dari {prodiList.length} program
                    studi
                </span>
            </div>
        </div>
    {/if}
</div>

<!-- Add/Edit Modal -->
{#if showModal}
    <div
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm"
        on:click={closeModal}
    >
        <div
            class="bg-white dark:bg-slate-900 rounded-xl shadow-2xl w-full max-w-lg overflow-hidden"
            on:click|stopPropagation
        >
            <div
                class="p-6 border-b border-slate-200 dark:border-slate-800 flex justify-between items-center"
            >
                <h3 class="text-xl font-bold text-slate-900 dark:text-white">
                    {modalMode === "add"
                        ? "Tambah Program Studi Baru"
                        : "Edit Program Studi"}
                </h3>
                <button
                    on:click={closeModal}
                    class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200"
                >
                    <span class="material-symbols-outlined">close</span>
                </button>
            </div>

            <form class="p-6 space-y-4" on:submit|preventDefault={handleSubmit}>
                {#if formError}
                    <div
                        class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-3 text-red-700 dark:text-red-400 text-sm"
                    >
                        {formError}
                    </div>
                {/if}

                <div>
                    <label
                        class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
                    >
                        Nama Program Studi <span class="text-red-500">*</span>
                    </label>
                    <input
                        bind:value={formNama}
                        class="w-full px-4 py-2.5 bg-slate-100 dark:bg-slate-800 border-none rounded-lg focus:ring-2 focus:ring-primary/50 text-sm transition-all"
                        placeholder="Contoh: Teknik Informatika"
                        type="text"
                    />
                </div>

                <div>
                    <label
                        class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
                    >
                        Kode Program Studi <span class="text-red-500">*</span>
                    </label>
                    <input
                        bind:value={formKode}
                        class="w-full px-4 py-2.5 bg-slate-100 dark:bg-slate-800 border-none rounded-lg focus:ring-2 focus:ring-primary/50 text-sm transition-all"
                        placeholder="Contoh: TI"
                        type="text"
                    />
                </div>

                <div>
                    <label
                        class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
                    >
                        Fakultas <span class="text-red-500">*</span>
                    </label>
                    <select
                        bind:value={formFakultasId}
                        class="w-full px-4 py-2.5 bg-slate-100 dark:bg-slate-800 border-none rounded-lg focus:ring-2 focus:ring-primary/50 text-sm transition-all"
                    >
                        <option value="">Pilih Fakultas</option>
                        {#each fakultasList as fak}
                            <option value={fak.id}>{fak.nama}</option>
                        {/each}
                    </select>
                    {#if fakultasList.length === 0}
                        <p class="mt-1 text-xs text-amber-500">
                            Belum ada fakultas. Silakan tambahkan fakultas
                            terlebih dahulu.
                        </p>
                    {/if}
                </div>

                <div
                    class="flex justify-end gap-3 pt-4 border-t border-slate-200 dark:border-slate-800"
                >
                    <button
                        type="button"
                        on:click={closeModal}
                        class="px-5 py-2.5 bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 rounded-lg font-bold hover:bg-slate-200 dark:hover:bg-slate-700 transition-all"
                    >
                        Batal
                    </button>
                    <button
                        type="submit"
                        class="flex items-center gap-2 px-5 py-2.5 bg-primary text-white font-bold rounded-lg shadow-lg shadow-primary/25 hover:bg-primary/90 transition-all active:scale-95 disabled:opacity-50"
                        disabled={formLoading}
                    >
                        <span class="material-symbols-outlined text-xl"
                            >save</span
                        >
                        <span>{formLoading ? "Menyimpan..." : "Simpan"}</span>
                    </button>
                </div>
            </form>
        </div>
    </div>
{/if}
