<script>
  import { onMount } from "svelte";
  import { createDocument } from "../../services/documentService.js";
  import fakultasService from "../../services/fakultasService.js";
  import prodiService from "../../services/prodiService.js";
  import { link } from "svelte-spa-router";

  let title = "";
  let author = "";
  let fileType = "";
  let status = "draft";
  let fakultasId = "";
  let prodiId = "";
  let dosenPembimbing = "";
  let files = [];
  let fileLocks = []; // track lock state per file
  let confirmCheck = false;

  let loading = false;
  let error = "";

  let fakultasList = [];
  let prodiList = [];
  let loadingFakultas = true;
  let loadingProdi = false;

  onMount(async () => {
    try {
      fakultasList = await fakultasService.getAll();
    } catch (e) {
      console.error("Gagal memuat fakultas:", e);
    } finally {
      loadingFakultas = false;
    }
  });

  // Ketika fakultas berubah, muat prodi yang sesuai
  async function onFakultasChange() {
    prodiId = "";
    prodiList = [];

    if (!fakultasId) return;

    try {
      loadingProdi = true;
      prodiList = await prodiService.getAll(fakultasId);
    } catch (e) {
      console.error("Gagal memuat prodi:", e);
    } finally {
      loadingProdi = false;
    }
  }

  function handleFileChange(e) {
    const newFiles = Array.from(e.target.files);
    files = newFiles;
    fileLocks = newFiles.map(() => false);
  }

  function removeFile(index) {
    files = files.filter((_, i) => i !== index);
    fileLocks = fileLocks.filter((_, i) => i !== index);
    /** @type {HTMLInputElement | null} */
    const input = /** @type {HTMLInputElement} */ (
      document.getElementById("file-input")
    );
    if (input) input.value = "";
  }

  function toggleLock(index) {
    fileLocks[index] = !fileLocks[index];
    fileLocks = fileLocks; // trigger reactivity
  }

  function formatFileSize(bytes) {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + " " + sizes[i];
  }

  async function handleSubmit() {
    error = "";

    if (!title || !author || !fileType || files.length === 0 || !confirmCheck) {
      error =
        "Semua field wajib diisi, minimal 1 file diunggah, dan dikonfirmasi.";
      return;
    }

    loading = true;

    const formData = new FormData();
    formData.append("title", title);
    formData.append("author", author);
    formData.append("category", fileType);
    formData.append("status", status);
    formData.append("dosen_pembimbing", dosenPembimbing);

    if (fakultasId) {
      formData.append("fakultas_id", fakultasId);
    }
    if (prodiId) {
      formData.append("prodi_id", prodiId);
    }

    // Append multiple files
    for (const file of files) {
      formData.append("files", file);
    }

    // Append lock status per file
    formData.append(
      "file_locks",
      fileLocks.map((l) => (l ? "true" : "false")).join(","),
    );

    try {
      await createDocument(formData);
      alert("Dokumen berhasil ditambahkan");
      window.location.href = "#/documents";
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }
</script>

<div class="max-w-4xl mx-auto flex flex-col gap-6">
  <!-- Breadcrumb -->
  <nav class="flex items-center gap-2 text-sm">
    <a
      href="#/"
      use:link
      class="text-slate-500 hover:text-primary transition-colors">Dashboard</a
    >
    <span class="text-slate-400">/</span>
    <a
      href="#/documents"
      use:link
      class="text-slate-500 hover:text-primary transition-colors"
      >Kelola Dokumen</a
    >
    <span class="text-slate-400">/</span>
    <span class="text-slate-900 dark:text-white font-medium"
      >Tambah Dokumen</span
    >
  </nav>

  <!-- Header -->
  <div class="flex flex-col gap-1">
    <h2
      class="text-3xl font-black text-slate-900 dark:text-white tracking-tight"
    >
      Tambah Dokumen Baru
    </h2>
    <p class="text-slate-500 dark:text-slate-400">
      Lengkapi detail informasi dokumen di bawah ini
    </p>
  </div>

  {#if error}
    <div
      class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-xl p-4 text-red-700 dark:text-red-400"
    >
      <p class="font-semibold">Error:</p>
      <p>{error}</p>
    </div>
  {/if}

  <!-- Form -->
  <div
    class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden"
  >
    <div class="p-6 md:p-8">
      <h3 class="text-lg font-bold mb-6 flex items-center gap-2">
        <span class="material-symbols-outlined text-primary">description</span>
        Detail Informasi Dokumen
      </h3>

      <form class="space-y-6" on:submit|preventDefault={handleSubmit}>
        <!-- Judul -->
        <div>
          <label
            for="input-judul"
            class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
          >
            Judul Dokumen <span class="text-red-500">*</span>
          </label>
          <input
            id="input-judul"
            bind:value={title}
            class="w-full px-4 py-2.5 bg-slate-100 dark:bg-slate-800 border-none rounded-lg focus:ring-2 focus:ring-primary/50 text-sm transition-all"
            placeholder="Contoh: Analisis Pengaruh Digitalisasi Terhadap Efisiensi..."
            type="text"
          />
        </div>

        <!-- Penulis + Jenis -->
        <div class="grid md:grid-cols-2 gap-6">
          <div>
            <label
              for="input-penulis"
              class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
            >
              Penulis <span class="text-red-500">*</span>
            </label>
            <input
              id="input-penulis"
              bind:value={author}
              class="w-full px-4 py-2.5 bg-slate-100 dark:bg-slate-800 border-none rounded-lg focus:ring-2 focus:ring-primary/50 text-sm transition-all"
              placeholder="Masukkan nama penulis"
              type="text"
            />
          </div>

          <div>
            <label
              for="input-jenis"
              class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
            >
              Jenis Dokumen <span class="text-red-500">*</span>
            </label>
            <select
              id="input-jenis"
              bind:value={fileType}
              class="w-full px-4 py-2.5 bg-slate-100 dark:bg-slate-800 border-none rounded-lg focus:ring-2 focus:ring-primary/50 text-sm transition-all"
            >
              <option value="">Pilih jenis dokumen</option>
              <option value="skripsi">Skripsi</option>
              <option value="tesis">Tesis</option>
              <option value="jurnal">Jurnal</option>
            </select>
          </div>
        </div>

        <!-- Fakultas + Prodi -->
        <div class="grid md:grid-cols-2 gap-6">
          <div>
            <label
              for="input-fakultas"
              class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
            >
              Fakultas
            </label>
            {#if loadingFakultas}
              <div
                class="w-full px-4 py-2.5 bg-slate-100 dark:bg-slate-800 rounded-lg text-sm text-slate-400"
              >
                Memuat data fakultas...
              </div>
            {:else}
              <select
                id="input-fakultas"
                bind:value={fakultasId}
                on:change={onFakultasChange}
                class="w-full px-4 py-2.5 bg-slate-100 dark:bg-slate-800 border-none rounded-lg focus:ring-2 focus:ring-primary/50 text-sm transition-all"
              >
                <option value="">Pilih Fakultas (Opsional)</option>
                {#each fakultasList as fak}
                  <option value={fak.id}>{fak.nama}</option>
                {/each}
              </select>
              {#if fakultasList.length === 0}
                <p class="mt-1 text-xs text-amber-500">
                  Belum ada data fakultas. Tambahkan di menu Fakultas terlebih
                  dahulu.
                </p>
              {/if}
            {/if}
          </div>

          <div>
            <label
              for="input-prodi"
              class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
            >
              Program Studi
            </label>
            {#if loadingProdi}
              <div
                class="w-full px-4 py-2.5 bg-slate-100 dark:bg-slate-800 rounded-lg text-sm text-slate-400"
              >
                Memuat data prodi...
              </div>
            {:else}
              <select
                id="input-prodi"
                bind:value={prodiId}
                class="w-full px-4 py-2.5 bg-slate-100 dark:bg-slate-800 border-none rounded-lg focus:ring-2 focus:ring-primary/50 text-sm transition-all"
                disabled={!fakultasId}
              >
                <option value=""
                  >{fakultasId
                    ? "Pilih Program Studi (Opsional)"
                    : "Pilih Fakultas terlebih dahulu"}</option
                >
                {#each prodiList as prodi}
                  <option value={prodi.id}>{prodi.nama}</option>
                {/each}
              </select>
              {#if fakultasId && prodiList.length === 0 && !loadingProdi}
                <p class="mt-1 text-xs text-amber-500">
                  Belum ada prodi untuk fakultas ini.
                </p>
              {/if}
            {/if}
          </div>
        </div>

        <!-- Dosen Pembimbing -->
        <div>
          <label
            for="input-dosen"
            class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
          >
            Dosen Pembimbing
          </label>
          <input
            id="input-dosen"
            bind:value={dosenPembimbing}
            class="w-full px-4 py-2.5 bg-slate-100 dark:bg-slate-800 border-none rounded-lg focus:ring-2 focus:ring-primary/50 text-sm transition-all"
            placeholder="Contoh: Prof. Dr. Ahmad, M.Si"
            type="text"
          />
        </div>

        <!-- Status -->
        <div>
          <label
            for="input-status"
            class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
          >
            Status <span class="text-red-500">*</span>
          </label>
          <select
            id="input-status"
            bind:value={status}
            class="w-full px-4 py-2.5 bg-slate-100 dark:bg-slate-800 border-none rounded-lg focus:ring-2 focus:ring-primary/50 text-sm transition-all"
          >
            <option value="draft">Draft</option>
            <option value="publish">Publish</option>
          </select>
        </div>

        <!-- Multiple File Upload -->
        <div>
          <label
            for="file-input"
            class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
          >
            Unggah File <span class="text-red-500">*</span>
          </label>
          <div
            class="mb-2 p-3 bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-300 rounded-lg text-sm flex items-center gap-2"
          >
            <span class="material-symbols-outlined text-base">info</span>
            Anda dapat mengunggah lebih dari 1 file sekaligus. Pilih beberapa file
            dengan Ctrl+Click atau Shift+Click.
          </div>
          <input
            id="file-input"
            type="file"
            on:change={handleFileChange}
            accept=".pdf,.doc,.docx"
            multiple
            class="w-full px-4 py-2.5 bg-slate-100 dark:bg-slate-800 border-none rounded-lg focus:ring-2 focus:ring-primary/50 text-sm transition-all file:mr-4 file:py-2 file:px-4 file:rounded-lg file:border-0 file:bg-primary file:text-white file:font-semibold hover:file:bg-primary/90"
          />
          <p class="mt-1 text-xs text-slate-400">
            Format yang didukung: PDF, DOC, DOCX (Maks 50MB per file)
          </p>

          <!-- File list preview -->
          {#if files.length > 0}
            <div class="mt-3 space-y-2">
              <p class="text-sm font-bold text-slate-700 dark:text-slate-300">
                {files.length} file dipilih:
              </p>
              {#each files as file, i}
                <div
                  class="flex items-center justify-between px-3 py-2 rounded-lg border {fileLocks[
                    i
                  ]
                    ? 'bg-amber-50 dark:bg-amber-900/20 border-amber-200 dark:border-amber-800'
                    : 'bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800'}"
                >
                  <div class="flex items-center gap-2 min-w-0">
                    <span
                      class="material-symbols-outlined text-base shrink-0 {fileLocks[
                        i
                      ]
                        ? 'text-amber-600'
                        : 'text-green-600'}"
                      >{fileLocks[i] ? "lock" : "check_circle"}</span
                    >
                    <span
                      class="text-sm truncate {fileLocks[i]
                        ? 'text-amber-700 dark:text-amber-400'
                        : 'text-green-700 dark:text-green-400'}"
                      >{file.name}</span
                    >
                    <span
                      class="text-xs shrink-0 {fileLocks[i]
                        ? 'text-amber-500'
                        : 'text-green-500'}">({formatFileSize(file.size)})</span
                    >
                  </div>
                  <div class="flex items-center gap-1 shrink-0 ml-2">
                    <!-- Lock Toggle -->
                    <button
                      type="button"
                      on:click={() => toggleLock(i)}
                      class="flex items-center gap-1 px-2 py-1 rounded text-xs font-semibold transition-all {fileLocks[
                        i
                      ]
                        ? 'bg-amber-100 dark:bg-amber-900/40 text-amber-700 dark:text-amber-400 hover:bg-amber-200'
                        : 'bg-green-100 dark:bg-green-900/40 text-green-700 dark:text-green-400 hover:bg-green-200'}"
                      title={fileLocks[i]
                        ? "Klik untuk membuka kunci"
                        : "Klik untuk mengunci file"}
                    >
                      <span class="material-symbols-outlined text-sm">
                        {fileLocks[i] ? "lock" : "lock_open"}
                      </span>
                      {fileLocks[i] ? "Terkunci" : "Terbuka"}
                    </button>
                    <!-- Remove -->
                    <button
                      type="button"
                      on:click={() => removeFile(i)}
                      class="text-red-400 hover:text-red-600 transition-colors"
                      title="Hapus file"
                    >
                      <span class="material-symbols-outlined text-base"
                        >close</span
                      >
                    </button>
                  </div>
                </div>
              {/each}
              <p class="text-xs text-slate-400 flex items-center gap-1 mt-1">
                <span class="material-symbols-outlined text-sm">info</span>
                File yang dikunci hanya bisa didownload oleh pengguna yang sudah
                login.
              </p>
            </div>
          {/if}
        </div>

        <!-- Confirm -->
        <div
          class="flex gap-3 items-start bg-primary/5 dark:bg-primary/10 p-4 rounded-lg border border-primary/20"
        >
          <input
            type="checkbox"
            bind:checked={confirmCheck}
            id="confirm-check"
            class="mt-0.5 w-5 h-5 text-primary rounded focus:ring-2 focus:ring-primary/50"
          />
          <label
            for="confirm-check"
            class="text-sm text-slate-700 dark:text-slate-300"
          >
            Saya mengonfirmasi bahwa dokumen ini adalah versi terbaru dan
            informasi yang dimasukkan sudah benar.
          </label>
        </div>

        <!-- Action -->
        <div
          class="flex justify-end gap-3 pt-4 border-t border-slate-200 dark:border-slate-800"
        >
          <a
            href="#/documents"
            use:link
            class="px-6 py-2.5 bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 rounded-lg font-bold hover:bg-slate-200 dark:hover:bg-slate-700 transition-all"
          >
            Batal
          </a>

          <button
            type="submit"
            class="flex items-center gap-2 px-6 py-2.5 bg-primary text-white font-bold rounded-lg shadow-lg shadow-primary/25 hover:bg-primary/90 transition-all active:scale-95 disabled:opacity-50 disabled:cursor-not-allowed"
            disabled={loading}
          >
            <span class="material-symbols-outlined text-xl">save</span>
            <span>{loading ? "Menyimpan..." : "Simpan Dokumen"}</span>
          </button>
        </div>
      </form>
    </div>
  </div>
</div>
