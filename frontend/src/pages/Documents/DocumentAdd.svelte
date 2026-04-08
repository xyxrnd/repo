<script>
  import { onMount } from "svelte";
  import { createDocument } from "../../services/documentService.js";
  import fakultasService from "../../services/fakultasService.js";
  import prodiService from "../../services/prodiService.js";
  import { link } from "svelte-spa-router";
  import * as pdfjsLib from "pdfjs-dist";

  pdfjsLib.GlobalWorkerOptions.workerSrc = "/pdf.worker.min.mjs";

  let title = "";
  let author = "";
  let abstrak = "";
  let fileType = "";
  let status = "draft";
  let fakultasId = "";
  let prodiId = "";
  let dosenPembimbing1 = "";
  let dosenPembimbing2 = "";
  let kataKunci = "";
  let tahun = "";
  let files = [];
  let fileLocks = [];
  let confirmCheck = false;

  let loading = false;
  let error = "";
  let extracting = false;
  let extractResult = null; // { success: bool, fields: string[] }

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

  // ===== PDF TEXT EXTRACTION =====
  async function extractTextFromPDF(file, maxPages = 10) {
    const arrayBuffer = await file.arrayBuffer();
    const pdf = await pdfjsLib.getDocument({ data: arrayBuffer }).promise;
    const totalPages = Math.min(pdf.numPages, maxPages);
    let fullText = "";
    const pageTexts = [];

    for (let i = 1; i <= totalPages; i++) {
      const page = await pdf.getPage(i);
      const textContent = await page.getTextContent();

      // Reconstruct lines based on Y-coordinate position
      const items = textContent.items.filter((item) => "str" in item);
      const textItems = items.map((item) => {
        const t = /** @type {any} */ (item);
        return {
          str: t.str,
          x: t.transform ? t.transform[4] : 0,
          y: t.transform ? t.transform[5] : 0,
          fontSize: t.transform ? Math.abs(t.transform[0]) : 12,
        };
      });

      // Sort by Y (descending since PDF Y=0 is bottom) then X (ascending)
      textItems.sort((a, b) => {
        const yDiff = b.y - a.y;
        if (Math.abs(yDiff) > 3) return yDiff; // different line
        return a.x - b.x; // same line, sort left to right
      });

      // Group items into lines based on Y proximity
      const lines = [];
      let currentLine = [];
      let currentY = null;

      for (const item of textItems) {
        if (currentY === null || Math.abs(item.y - currentY) > 3) {
          if (currentLine.length > 0) {
            lines.push(currentLine.map((it) => it.str).join(" "));
          }
          currentLine = [item];
          currentY = item.y;
        } else {
          currentLine.push(item);
        }
      }
      if (currentLine.length > 0) {
        lines.push(currentLine.map((it) => it.str).join(" "));
      }

      const pageText = lines.join("\n");
      pageTexts.push(pageText);
      fullText += pageText + "\n\n--- PAGE BREAK ---\n\n";
    }

    return { fullText, pageTexts, numPages: pdf.numPages };
  }

  function cleanText(text) {
    return text.replace(/\s+/g, " ").trim();
  }

  function extractTitle(text) {
    const lines = text
      .split(/\n/)
      .map((l) => l.trim())
      .filter(Boolean);

    // Strategy 1: Find title between doc type label and "Oleh:"
    // Skripsi Indonesia: SKRIPSI -> JUDUL -> (untuk memenuhi...) -> Oleh:
    const typeLabels =
      /^\s*(SKRIPSI|TESIS|DISERTASI|JURNAL|TUGAS AKHIR|LAPORAN)\s*$/i;
    const olehPattern =
      /\b(o\s*leh|disusun\s+oleh|penulis|diajukan\s+oleh)\s*:?\s*$/i;
    const skipPatterns =
      /^(untuk|sebagai|guna|diajukan|memenuhi|memperoleh|salah satu|syarat|gelar|sarjana|magister|doktor|strata|S-?\d|dengan|pada|di|yang)/i;

    let titleParts = [];
    let capturing = false;

    for (let idx = 0; idx < lines.length; idx++) {
      const line = lines[idx];

      // Start capturing after SKRIPSI/TESIS label
      if (typeLabels.test(line) && !capturing) {
        capturing = true;
        continue;
      }

      if (capturing) {
        // Stop at "Oleh:" or similar
        if (olehPattern.test(line)) break;

        const cleaned = cleanText(line);
        // Skip short/label lines
        if (cleaned.length <= 3) continue;
        // Skip "untuk memenuhi..." type lines
        if (skipPatterns.test(cleaned)) continue;
        // Skip page break markers
        if (cleaned.includes("PAGE BREAK")) continue;
        // Stop if we hit university/faculty info
        if (/^(UNIVERSITAS|FAKULTAS|PROGRAM STUDI|JURUSAN)/i.test(cleaned))
          break;

        titleParts.push(cleaned);
      }
    }

    if (titleParts.length > 0) {
      return titleParts.join(" ").replace(/\s+/g, " ").trim();
    }

    // Strategy 2: Look for consecutive uppercase lines that look like a title
    // (skip known non-title blocks)
    const nonTitleStart =
      /^(UNIVERSITAS|FAKULTAS|PROGRAM|KEMENTERIAN|DEPARTEMEN|SKRIPSI|TESIS|JURNAL|HALAMAN|LEMBAR|KATA PENGANTAR|DAFTAR|ABSTRAK|ABSTRACT|BAB|OLEH|PERNYATAAN|PENGESAHAN|PERSETUJUAN)/i;

    for (let i = 0; i < lines.length; i++) {
      const cl = cleanText(lines[i]);
      // Find uppercase text > 20 chars that doesn't start with known patterns
      if (
        cl.length > 20 &&
        cl === cl.toUpperCase() &&
        /[A-Z]/.test(cl) &&
        !nonTitleStart.test(cl)
      ) {
        // Collect consecutive uppercase lines
        let collected = [cl];
        for (let j = i + 1; j < lines.length && j < i + 5; j++) {
          const nextCl = cleanText(lines[j]);
          if (
            nextCl.length > 5 &&
            nextCl === nextCl.toUpperCase() &&
            /[A-Z]/.test(nextCl) &&
            !nonTitleStart.test(nextCl) &&
            !/^(UNTUK|SEBAGAI|GUNA|DIAJUKAN)/i.test(nextCl)
          ) {
            collected.push(nextCl);
          } else {
            break;
          }
        }
        return collected.join(" ");
      }
    }

    return "";
  }

  function extractAuthor(text) {
    const lines = text
      .split(/\n/)
      .map((l) => l.trim())
      .filter(Boolean);

    // Find "Oleh :" or "Disusun oleh" line, then get the next non-empty line
    for (let i = 0; i < lines.length; i++) {
      if (
        /^\s*(o\s*leh|disusun\s+oleh|penulis|diajukan\s+oleh)\s*:?\s*$/i.test(
          lines[i],
        )
      ) {
        // Next non-empty line should be the author name
        for (let j = i + 1; j < lines.length && j <= i + 3; j++) {
          const candidate = cleanText(lines[j]);
          if (!candidate) continue;
          // Skip if it's a NIM/NPM line
          if (/^\d{5,}/.test(candidate)) continue;
          // Skip if it's a label
          if (/^(NIM|NPM|NRP|No\.|Nomor)/i.test(candidate)) continue;
          // Should be a name (at least two words, no digits)
          let name = candidate.replace(/\s*\d{5,}.*$/, "").trim(); // remove trailing NIM
          name = name.replace(/\s*\(.*\)$/, "").trim(); // remove trailing (NIM)
          if (
            name.length > 3 &&
            !/universitas|fakultas|program|jurusan/i.test(name)
          ) {
            return name;
          }
        }
      }

      // Also handle inline: "Oleh : Nama Lengkap"
      const inlineMatch = lines[i].match(
        /(?:o\s*leh|disusun\s+oleh|penulis)\s*:\s*(.{3,50})/i,
      );
      if (inlineMatch) {
        let name = cleanText(inlineMatch[1]);
        name = name.replace(/\s*\d{5,}.*$/, "").trim();
        name = name.replace(/\s*\(.*\)$/, "").trim();
        if (
          name.length > 3 &&
          !/universitas|fakultas|program|jurusan|NIM|NPM/i.test(name)
        ) {
          return name;
        }
      }
    }
    return "";
  }

  function extractDosenPembimbing(text) {
    const result = { pembimbing1: "", pembimbing2: "" };
    const lines = text
      .split(/\n/)
      .map((l) => l.trim())
      .filter(Boolean);

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i];

      // "Pembimbing I" or "Pembimbing 1" or "Pembimbing Utama"
      if (
        /pembimbing\s*(?:I|1|utama|pertama)\s*:?\s*$/i.test(line) &&
        !result.pembimbing1
      ) {
        // Name is on next line(s)
        for (let j = i + 1; j < lines.length && j <= i + 2; j++) {
          const candidate = cleanText(lines[j]);
          if (candidate.length > 5 && /[A-Za-z]/.test(candidate)) {
            result.pembimbing1 = candidate
              .replace(/\s*NIP\.?\s*:?.*$/i, "")
              .replace(/\s*NIDN\.?\s*:?.*$/i, "")
              .trim();
            break;
          }
        }
      }

      // Inline: "Pembimbing I : Dr. ..."
      const match1 = line.match(
        /pembimbing\s*(?:I|1|utama|pertama)\s*:\s*(.+)/i,
      );
      if (match1 && !result.pembimbing1) {
        result.pembimbing1 = cleanText(match1[1])
          .replace(/\s*NIP\.?\s*:?.*$/i, "")
          .replace(/\s*NIDN\.?\s*:?.*$/i, "")
          .trim();
      }

      // "Pembimbing II" or "Pembimbing 2"
      if (
        /pembimbing\s*(?:II|2|kedua|pendamping)\s*:?\s*$/i.test(line) &&
        !result.pembimbing2
      ) {
        for (let j = i + 1; j < lines.length && j <= i + 2; j++) {
          const candidate = cleanText(lines[j]);
          if (candidate.length > 5 && /[A-Za-z]/.test(candidate)) {
            result.pembimbing2 = candidate
              .replace(/\s*NIP\.?\s*:?.*$/i, "")
              .replace(/\s*NIDN\.?\s*:?.*$/i, "")
              .trim();
            break;
          }
        }
      }

      const match2 = line.match(
        /pembimbing\s*(?:II|2|kedua|pendamping)\s*:\s*(.+)/i,
      );
      if (match2 && !result.pembimbing2) {
        result.pembimbing2 = cleanText(match2[1])
          .replace(/\s*NIP\.?\s*:?.*$/i, "")
          .replace(/\s*NIDN\.?\s*:?.*$/i, "")
          .trim();
      }
    }

    // Fallback: single "Dosen Pembimbing" without number
    if (!result.pembimbing1) {
      for (let i = 0; i < lines.length; i++) {
        const m = lines[i].match(
          /(?:dosen\s+pembimbing|pembimbing\s+akademik)\s*:\s*(.+)/i,
        );
        if (m) {
          result.pembimbing1 = cleanText(m[1])
            .replace(/\s*NIP\.?\s*:?.*$/i, "")
            .trim();
          break;
        }
      }
    }

    return result;
  }

  function extractAbstrak(text) {
    const lines = text.split(/\n/);

    // Find the ABSTRAK heading line
    let abstrakStart = -1;
    for (let i = 0; i < lines.length; i++) {
      if (/^\s*ABSTRAK\s*$/i.test(lines[i].trim())) {
        abstrakStart = i + 1;
        break;
      }
    }

    if (abstrakStart >= 0) {
      // Collect lines until "Kata Kunci", "ABSTRACT", "Keywords", or PAGE BREAK
      const abstrakLines = [];
      for (let i = abstrakStart; i < lines.length; i++) {
        const trimmed = lines[i].trim();
        // Stop conditions
        if (/^\s*(Kata\s*[Kk]unci|Keywords?)\s*:?\s*/i.test(trimmed)) break;
        if (/^\s*ABSTRACT\s*$/i.test(trimmed)) break;
        if (/PAGE BREAK/i.test(trimmed)) break;
        if (/^\s*(BAB|DAFTAR|KATA PENGANTAR|LEMBAR)\s/i.test(trimmed)) break;
        if (trimmed) abstrakLines.push(trimmed);
      }
      if (abstrakLines.length > 0) {
        return abstrakLines.join(" ").replace(/\s+/g, " ").trim();
      }
    }

    // Try ABSTRACT (English)
    let abstractStart = -1;
    for (let i = 0; i < lines.length; i++) {
      if (/^\s*ABSTRACT\s*$/i.test(lines[i].trim())) {
        abstractStart = i + 1;
        break;
      }
    }

    if (abstractStart >= 0) {
      const abstractLines = [];
      for (let i = abstractStart; i < lines.length; i++) {
        const trimmed = lines[i].trim();
        if (/^\s*(Keywords?|Kata\s*[Kk]unci)\s*:?\s*/i.test(trimmed)) break;
        if (/PAGE BREAK/i.test(trimmed)) break;
        if (/^\s*(BAB|DAFTAR|KATA PENGANTAR|LEMBAR)\s/i.test(trimmed)) break;
        if (trimmed) abstractLines.push(trimmed);
      }
      if (abstractLines.length > 0) {
        return abstractLines.join(" ").replace(/\s+/g, " ").trim();
      }
    }

    return "";
  }

  function extractKataKunci(text) {
    const lines = text.split(/\n/);

    for (let i = 0; i < lines.length; i++) {
      const trimmed = lines[i].trim();

      // Match "Kata Kunci : keyword1, keyword2, ..."
      const match = trimmed.match(/(?:Kata\s*[Kk]unci|Keywords?)\s*:?\s*(.+)/i);
      if (match && match[1]) {
        let keywords = cleanText(match[1]);
        // If keywords continue to next line (short match)
        if (keywords.length < 15 && i + 1 < lines.length) {
          const nextLine = lines[i + 1].trim();
          if (
            nextLine &&
            !/^(ABSTRACT|KATA PENGANTAR|DAFTAR|BAB|PAGE BREAK)/i.test(nextLine)
          ) {
            keywords += " " + nextLine;
          }
        }
        keywords = keywords
          .replace(
            /\s*(ABSTRACT|KATA PENGANTAR|DAFTAR|BAB\s|PAGE BREAK).*$/i,
            "",
          )
          .trim();
        keywords = keywords.replace(/\.+$/, "").trim();
        return keywords;
      }
    }
    return "";
  }

  function extractFakultasProdi(text) {
    const result = { fakultas: "", prodi: "" };
    const lines = text
      .split(/\n/)
      .map((l) => l.trim())
      .filter(Boolean);

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i];

      // Match "FAKULTAS TEKNIK" or "Fakultas : Teknik"
      if (!result.fakultas) {
        const fakMatch = line.match(/^FAKULTAS\s+(.+)/i);
        if (fakMatch) {
          let nama = cleanText(fakMatch[1]);
          // Remove trailing university or year
          nama = nama.replace(/\s*(UNIVERSITAS|TAHUN|\d{4}).*$/i, "").trim();
          if (nama.length > 2) result.fakultas = nama;
        }
      }

      // Match "PROGRAM STUDI INFORMATIKA" or "JURUSAN INFORMATIKA"
      if (!result.prodi) {
        const prodiMatch = line.match(
          /^(?:PROGRAM\s+STUDI|PRODI|JURUSAN)\s+:?\s*(.+)/i,
        );
        if (prodiMatch) {
          let nama = cleanText(prodiMatch[1]);
          nama = nama
            .replace(/\s*(FAKULTAS|UNIVERSITAS|TAHUN|\d{4}).*$/i, "")
            .trim();
          if (nama.length > 2) result.prodi = nama;
        }
      }
    }

    return result;
  }

  function detectDocType(text, fileName) {
    const lower = (text + " " + fileName).toLowerCase();
    if (/\bskripsi\b/.test(lower)) return "skripsi";
    if (/\btesis\b/.test(lower)) return "tesis";
    if (/\bjurnal\b|journal\b|artikel\b/.test(lower)) return "jurnal";
    return "";
  }

  function matchFakultasList(extractedName) {
    if (!extractedName || fakultasList.length === 0) return "";
    const normalized = extractedName.toLowerCase().replace(/\s+/g, " ");
    for (const fak of fakultasList) {
      const fakNorm = fak.nama.toLowerCase().replace(/\s+/g, " ");
      // Check if extracted name contains or is contained by list name
      if (fakNorm.includes(normalized) || normalized.includes(fakNorm)) {
        return fak.id;
      }
      // Partial match (at least 60% of words match)
      const extractedWords = normalized.split(" ");
      const fakWords = fakNorm.split(" ");
      const matches = extractedWords.filter((w) => fakWords.includes(w)).length;
      if (matches >= Math.ceil(fakWords.length * 0.6)) {
        return fak.id;
      }
    }
    return "";
  }

  async function matchProdiList(extractedName, fakId) {
    if (!extractedName || !fakId) return "";
    try {
      const prodiItems = await prodiService.getAll(fakId);
      prodiList = prodiItems;
      const normalized = extractedName.toLowerCase().replace(/\s+/g, " ");
      for (const p of prodiItems) {
        const pNorm = p.nama.toLowerCase().replace(/\s+/g, " ");
        if (pNorm.includes(normalized) || normalized.includes(pNorm)) {
          return p.id;
        }
        const extractedWords = normalized.split(" ");
        const pWords = pNorm.split(" ");
        const matches = extractedWords.filter((w) => pWords.includes(w)).length;
        if (matches >= Math.ceil(pWords.length * 0.6)) {
          return p.id;
        }
      }
    } catch (e) {
      console.error("Gagal memuat prodi untuk matching:", e);
    }
    return "";
  }

  async function autoExtractMetadata() {
    if (files.length === 0) return;

    const pdfFiles = files.filter((f) => f.name.toLowerCase().endsWith(".pdf"));
    if (pdfFiles.length === 0) return;

    extracting = true;
    extractResult = null;

    try {
      const filledFields = [];

      // Extract from all PDF files and combine text
      let allText = "";
      for (const pdfFile of pdfFiles) {
        try {
          const { fullText } = await extractTextFromPDF(pdfFile, 8);
          allText += fullText + "\n\n";
        } catch (e) {
          console.warn(`Gagal membaca ${pdfFile.name}:`, e);
        }
      }

      if (!allText.trim()) {
        extractResult = { success: false, fields: [] };
        return;
      }

      // Detect document type
      const detectedType = detectDocType(
        allText,
        pdfFiles.map((f) => f.name).join(" "),
      );
      if (detectedType && !fileType) {
        fileType = detectedType;
        filledFields.push("Jenis Dokumen");
      }

      // Extract author
      const extractedAuthor = extractAuthor(allText);
      if (extractedAuthor && !author) {
        author = extractedAuthor;
        filledFields.push("Penulis");
      }

      // Extract advisors
      const advisors = extractDosenPembimbing(allText);
      if (advisors.pembimbing1 && !dosenPembimbing1) {
        dosenPembimbing1 = advisors.pembimbing1;
        filledFields.push("Dosen Pembimbing 1");
      }
      if (advisors.pembimbing2 && !dosenPembimbing2) {
        dosenPembimbing2 = advisors.pembimbing2;
        filledFields.push("Dosen Pembimbing 2");
      }

      // Extract abstract
      const extractedAbstrak = extractAbstrak(allText);
      if (extractedAbstrak && !abstrak) {
        abstrak = extractedAbstrak;
        filledFields.push("Abstrak");
      }

      // Extract keywords
      const extractedKeywords = extractKataKunci(allText);
      if (extractedKeywords && !kataKunci) {
        kataKunci = extractedKeywords;
        filledFields.push("Kata Kunci");
      }

      // Extract faculty & program
      const fakProdi = extractFakultasProdi(allText);
      if (fakProdi.fakultas && !fakultasId) {
        const matchedFakId = matchFakultasList(fakProdi.fakultas);
        if (matchedFakId) {
          fakultasId = matchedFakId;
          filledFields.push("Fakultas");

          // Try to match prodi
          if (fakProdi.prodi && !prodiId) {
            const matchedProdiId = await matchProdiList(
              fakProdi.prodi,
              matchedFakId,
            );
            if (matchedProdiId) {
              prodiId = matchedProdiId;
              filledFields.push("Program Studi");
            }
          }
        }
      }

      extractResult = {
        success: filledFields.length > 0,
        fields: filledFields,
      };
    } catch (e) {
      console.error("Error extracting metadata:", e);
      extractResult = { success: false, fields: [] };
    } finally {
      extracting = false;
    }
  }

  // ===== FILE HANDLING =====
  async function handleFileChange(e) {
    const newFiles = Array.from(e.target.files);
    files = newFiles;
    fileLocks = newFiles.map(() => false);

    // Auto-extract metadata from PDFs
    await autoExtractMetadata();
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
    fileLocks = fileLocks;
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
    formData.append("abstrak", abstrak);
    formData.append("category", fileType);
    formData.append("status", status);
    formData.append("dosen_pembimbing", dosenPembimbing1);
    formData.append("dosen_pembimbing_2", dosenPembimbing2);
    formData.append("kata_kunci", kataKunci);
    if (tahun) formData.append("tahun", tahun);

    if (fakultasId) {
      formData.append("fakultas_id", fakultasId);
    }
    if (prodiId) {
      formData.append("prodi_id", prodiId);
    }

    for (const file of files) {
      formData.append("files", file);
    }

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
      Unggah file terlebih dahulu — sistem akan otomatis mengisi kolom dari PDF
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
      <form class="space-y-6" on:submit|preventDefault={handleSubmit}>
        <!-- STEP 1: File Upload (MOVED TO TOP) -->
        <div
          class="bg-gradient-to-br from-primary/5 to-blue-50 dark:from-primary/10 dark:to-blue-900/20 rounded-xl p-5 border border-primary/20"
        >
          <h3 class="text-lg font-bold mb-3 flex items-center gap-2">
            <span
              class="flex items-center justify-center w-7 h-7 rounded-full bg-primary text-white text-sm font-bold"
              >1</span
            >
            <span>Unggah File Dokumen</span>
            <span class="text-red-500 text-sm">*</span>
          </h3>
          <div
            class="mb-3 p-3 bg-white/80 dark:bg-slate-800/80 text-slate-600 dark:text-slate-300 rounded-lg text-sm flex items-start gap-2"
          >
            <span
              class="material-symbols-outlined text-primary text-lg mt-0.5 shrink-0"
              >auto_awesome</span
            >
            <div>
              <p class="font-semibold text-slate-700 dark:text-slate-200">
                Auto-Fill dari PDF
              </p>
              <p class="text-xs mt-0.5">
                Unggah file PDF skripsi/tesis/jurnal — sistem akan membaca teks
                dan mengisi otomatis judul, penulis, dosen pembimbing, abstrak,
                kata kunci, dan lainnya.
              </p>
            </div>
          </div>
          <input
            id="file-input"
            type="file"
            on:change={handleFileChange}
            accept=".pdf,.doc,.docx"
            multiple
            class="w-full px-4 py-2.5 bg-white dark:bg-slate-800 border-none rounded-lg focus:ring-2 focus:ring-primary/50 text-sm transition-all file:mr-4 file:py-2 file:px-4 file:rounded-lg file:border-0 file:bg-primary file:text-white file:font-semibold hover:file:bg-primary/90"
          />
          <p class="mt-1.5 text-xs text-slate-400">
            Format: PDF, DOC, DOCX (Maks 50MB per file). Pilih beberapa file
            dengan Ctrl+Click.
          </p>

          <!-- Extracting indicator -->
          {#if extracting}
            <div
              class="mt-3 flex items-center gap-3 p-3 bg-blue-50 dark:bg-blue-900/30 rounded-lg border border-blue-200 dark:border-blue-800 animate-pulse"
            >
              <div
                class="animate-spin rounded-full h-5 w-5 border-2 border-primary border-t-transparent shrink-0"
              ></div>
              <div>
                <p
                  class="text-sm font-semibold text-blue-700 dark:text-blue-300"
                >
                  Menganalisis file PDF...
                </p>
                <p class="text-xs text-blue-500 dark:text-blue-400">
                  Membaca teks dan mengekstrak metadata dokumen
                </p>
              </div>
            </div>
          {/if}

          <!-- Extract result -->
          {#if extractResult}
            {#if extractResult.success}
              <div
                class="mt-3 p-3 bg-emerald-50 dark:bg-emerald-900/20 rounded-lg border border-emerald-200 dark:border-emerald-800"
              >
                <div class="flex items-center gap-2 mb-1.5">
                  <span
                    class="material-symbols-outlined text-emerald-600 text-lg"
                    >check_circle</span
                  >
                  <p
                    class="text-sm font-bold text-emerald-700 dark:text-emerald-300"
                  >
                    Berhasil mengisi {extractResult.fields.length} kolom otomatis!
                  </p>
                </div>
                <div class="flex flex-wrap gap-1.5">
                  {#each extractResult.fields as field}
                    <span
                      class="inline-flex items-center gap-1 px-2 py-0.5 bg-emerald-100 dark:bg-emerald-900/40 text-emerald-700 dark:text-emerald-400 rounded-full text-xs font-medium"
                    >
                      <span class="material-symbols-outlined text-xs"
                        >check</span
                      >
                      {field}
                    </span>
                  {/each}
                </div>
                <p
                  class="mt-2 text-xs text-emerald-600/80 dark:text-emerald-400/80"
                >
                  Silakan periksa dan koreksi data yang sudah terisi di bawah
                  sebelum menyimpan.
                </p>
              </div>
            {:else}
              <div
                class="mt-3 p-3 bg-amber-50 dark:bg-amber-900/20 rounded-lg border border-amber-200 dark:border-amber-800"
              >
                <div class="flex items-center gap-2">
                  <span class="material-symbols-outlined text-amber-600 text-lg"
                    >info</span
                  >
                  <p class="text-sm text-amber-700 dark:text-amber-400">
                    Tidak dapat mengekstrak metadata dari file. Silakan isi
                    kolom secara manual.
                  </p>
                </div>
              </div>
            {/if}
          {/if}

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
                    : 'bg-white dark:bg-slate-800 border-green-200 dark:border-green-800'}"
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

        <!-- STEP 2: Detail Information -->
        <div>
          <h3 class="text-lg font-bold mb-5 flex items-center gap-2">
            <span
              class="flex items-center justify-center w-7 h-7 rounded-full bg-primary text-white text-sm font-bold"
              >2</span
            >
            <span>Detail Informasi Dokumen</span>
          </h3>
        </div>

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

        <!-- Abstrak -->
        <div>
          <label
            for="input-abstrak"
            class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
          >
            Abstrak / Ringkasan
          </label>
          <textarea
            id="input-abstrak"
            bind:value={abstrak}
            class="w-full px-4 py-2.5 bg-slate-100 dark:bg-slate-800 border-none rounded-lg focus:ring-2 focus:ring-primary/50 text-sm transition-all min-h-[160px] resize-y"
            placeholder="Masukkan abstrak atau ringkasan dokumen..."
            rows="6"
          ></textarea>
          <p class="mt-1 text-xs text-slate-400">
            Abstrak akan ditampilkan pada halaman detail dokumen.
          </p>
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

        <!-- Dosen Pembimbing 1 & 2 -->
        <div class="grid md:grid-cols-2 gap-6">
          <div>
            <label
              for="input-dosen1"
              class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
            >
              Dosen Pembimbing 1
            </label>
            <input
              id="input-dosen1"
              bind:value={dosenPembimbing1}
              class="w-full px-4 py-2.5 bg-slate-100 dark:bg-slate-800 border-none rounded-lg focus:ring-2 focus:ring-primary/50 text-sm transition-all"
              placeholder="Contoh: Prof. Dr. Ahmad, M.Si"
              type="text"
            />
          </div>
          <div>
            <label
              for="input-dosen2"
              class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
            >
              Dosen Pembimbing 2
            </label>
            <input
              id="input-dosen2"
              bind:value={dosenPembimbing2}
              class="w-full px-4 py-2.5 bg-slate-100 dark:bg-slate-800 border-none rounded-lg focus:ring-2 focus:ring-primary/50 text-sm transition-all"
              placeholder="Contoh: Dr. Budi Santoso, M.Kom"
              type="text"
            />
          </div>
        </div>

        <!-- Kata Kunci + Tahun -->
        <div class="grid md:grid-cols-2 gap-6">
          <div>
            <label
              for="input-kata-kunci"
              class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
            >
              Kata Kunci
            </label>
            <input
              id="input-kata-kunci"
              bind:value={kataKunci}
              class="w-full px-4 py-2.5 bg-slate-100 dark:bg-slate-800 border-none rounded-lg focus:ring-2 focus:ring-primary/50 text-sm transition-all"
              placeholder="Contoh: machine learning, deep learning"
              type="text"
            />
            <p class="mt-1 text-xs text-slate-400">
              Pisahkan setiap kata kunci dengan tanda koma.
            </p>
          </div>
          <div>
            <label
              for="input-tahun"
              class="block text-sm font-bold mb-2 text-slate-700 dark:text-slate-300"
            >
              Tahun Dokumen
            </label>
            <input
              id="input-tahun"
              bind:value={tahun}
              class="w-full px-4 py-2.5 bg-slate-100 dark:bg-slate-800 border-none rounded-lg focus:ring-2 focus:ring-primary/50 text-sm transition-all"
              placeholder="Contoh: 2024"
              type="number"
              min="1900"
              max="2099"
            />
            <p class="mt-1 text-xs text-slate-400">
              Tahun penerbitan/penyusunan dokumen.
            </p>
          </div>
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
            disabled={loading || extracting}
          >
            <span class="material-symbols-outlined text-xl">save</span>
            <span>{loading ? "Menyimpan..." : "Simpan Dokumen"}</span>
          </button>
        </div>
      </form>
    </div>
  </div>
</div>
