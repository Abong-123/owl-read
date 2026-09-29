<script>
  import { api } from '$lib/api.js';
  import { goto } from '$app/navigation';

  // Props
  export let book = null;           // null = create, != null = edit
  export let mode = 'create';       // 'create' | 'edit'

  const GENRES = [
    { key: 'novel',     label: 'Novel' },
    { key: 'comic',     label: 'Comic' },
    { key: 'majalah',   label: 'Majalah' },
    { key: 'non-fiksi', label: 'Non-fiksi' }
  ];

  const MAX_PDF = 50 * 1024 * 1024;
  const MAX_COVER = 5 * 1024 * 1024;
  const MAX_COVER_TEXT = 18;

  let title = book?.title ?? '';
  let author = book?.author ?? '';
  let year = book?.year ?? '';
  let synopsis = book?.synopsis ?? '';
  let genre = book?.genre ?? '';

  let coverMode = book?.cover_type ?? '';
  let coverFile = null;
  let coverColor = book?.cover_color || '#9c7450';
  let coverText = book?.cover_text ?? '';

  let pdfFile = null;

  let errors = {};
  let submitting = false;
  let serverError = '';
  let successMessage = '';

  $: existingCoverURL =
    book?.cover_type === 'uploaded' && book?.cover_path ? '/' + book.cover_path : '';

  // ---------- Validasi ----------
  function validate() {
    const e = {};

    if (!title.trim()) e.title = 'Judul wajib diisi';
    else if (title.length > 300) e.title = 'Judul maksimal 300 karakter';

    if (author.length > 200) e.author = 'Penulis maksimal 200 karakter';

    if (year !== '' && year !== null && year !== undefined) {
      const y = Number(year);
      if (!Number.isInteger(y) || y < -3000 || y > 9999)
        e.year = 'Tahun harus antara -3000 (SM) dan 9999';
    }

    if (synopsis.length > 5000) e.synopsis = 'Sinopsis maksimal 5000 karakter';
    if (!genre) e.genre = 'Pilih salah satu genre';
    if (!coverMode) e.cover = 'Pilih mode cover';

    if (coverMode === 'uploaded') {
      const hasExistingCover = book?.cover_type === 'uploaded' && book?.cover_path;
      if (!coverFile && !hasExistingCover) e.coverFile = 'Cover wajib diupload';
      else if (coverFile && coverFile.size > MAX_COVER) e.coverFile = 'Cover maksimal 5 MB';
    }

    if (coverMode === 'generated') {
      if (!coverColor || !/^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$/.test(coverColor))
        e.coverColor = 'Warna hex tidak valid';
      if (!coverText.trim()) e.coverText = 'Teks cover wajib diisi';
      else if ([...coverText].length > MAX_COVER_TEXT)
        e.coverText = `Teks cover maksimal ${MAX_COVER_TEXT} karakter`;
    }

    const hasExistingPDF = book?.pdf_path;
    if (!pdfFile && !hasExistingPDF) e.pdfFile = 'PDF wajib diupload';
    else if (pdfFile && pdfFile.size > MAX_PDF) e.pdfFile = 'PDF maksimal 50 MB';
    else if (pdfFile && !pdfFile.type.includes('pdf') && !pdfFile.name.toLowerCase().endsWith('.pdf'))
      e.pdfFile = 'File harus berformat PDF';

    errors = e;
    return Object.keys(e).length === 0;
  }

  // ---------- File handlers ----------
  function onCoverChange(event) {
    coverFile = event.target.files?.[0] || null;
  }
  function onPdfChange(event) {
    pdfFile = event.target.files?.[0] || null;
  }

  // ---------- Preview generated cover ----------
  $: genFontSize = calcFontSize([...coverText].length);
  function calcFontSize(len) {
    if (len <= 4) return 26;
    if (len <= 8) return 20;
    if (len <= 12) return 16;
    if (len <= 15) return 14;
    return 12;
  }
  $: textColor = isLight(coverColor) ? '#463d36' : '#fffdf9';
  function isLight(hex) {
    const c = (hex || '#9c7450').replace('#', '');
    const full = c.length === 3 ? c.split('').map((x) => x + x).join('') : c;
    const r = parseInt(full.slice(0, 2), 16);
    const g = parseInt(full.slice(2, 4), 16);
    const b = parseInt(full.slice(4, 6), 16);
    return 0.299 * r + 0.587 * g + 0.114 * b > 160;
  }

  // ---------- Preview uploaded cover ----------
  let coverPreviewURL = '';
  $: {
    if (coverPreviewURL) URL.revokeObjectURL(coverPreviewURL);
    coverPreviewURL = coverFile ? URL.createObjectURL(coverFile) : '';
  }

  // ---------- Kode upload ----------
  function askCode() {
    const code = prompt('Masukkan kode upload (4 digit):');
    if (code === null) return { cancelled: true };
    if (!/^\d{4}$/.test(code)) return { error: 'Kode harus 4 digit angka.' };
    if (code !== '1234') return { error: 'Kode salah. Silakan coba lagi.' };
    return { ok: true };
  }

  // ---------- Submit ----------
  async function handleSubmit(event) {
    event.preventDefault();
    serverError = '';
    successMessage = '';

    if (!validate()) {
      const firstKey = Object.keys(errors)[0];
      document.querySelector(`[data-field="${firstKey}"]`)?.scrollIntoView({
        behavior: 'smooth', block: 'center'
      });
      return;
    }

    const c = askCode();
    if (c.cancelled) return;
    if (c.error) { serverError = c.error; return; }

    const fd = new FormData();
    fd.append('title', title.trim());
    fd.append('author', author.trim());
    fd.append('year', year === '' || year === null ? '' : String(year));
    fd.append('synopsis', synopsis.trim());
    fd.append('genre', genre);
    fd.append('cover_type', coverMode);

    if (coverMode === 'uploaded') {
      if (coverFile) fd.append('cover_file', coverFile);
    } else {
      fd.append('cover_color', coverColor);
      fd.append('cover_text', coverText.trim());
    }

    if (pdfFile) fd.append('pdf_file', pdfFile);

    submitting = true;
    try {
      if (mode === 'edit') {
        await api.updateBook(book.id, fd);
        successMessage = `Perubahan pada "${title}" tersimpan.`;
      } else {
        await api.createBook(fd);
        successMessage = `Berhasil menambahkan "${title}".`;
      }
      setTimeout(() => {
        goto(mode === 'edit' ? `/read/${book.id}` : '/');
      }, 900);
    } catch (e) {
      serverError = e.message || 'Gagal menyimpan.';
    } finally {
      submitting = false;
    }
  }
</script>

<form on:submit={handleSubmit} novalidate>

  <!-- INFO DASAR -->
  <section class="group">
    <div class="field" data-field="title">
      <label for="title">Judul</label>
      <input id="title" type="text" placeholder="Judul buku" autocomplete="off"
             bind:value={title} class:invalid={errors.title} />
      {#if errors.title}<div class="err">{errors.title}</div>{/if}
    </div>

    <div class="row">
      <div class="field" data-field="author">
        <label for="author">Penulis <span class="opt">(opsional)</span></label>
        <input id="author" type="text" placeholder="Nama penulis" autocomplete="off"
               bind:value={author} class:invalid={errors.author} />
        {#if errors.author}<div class="err">{errors.author}</div>{/if}
      </div>

      <div class="field" data-field="year">
        <label for="year">Tahun <span class="opt">(opsional)</span></label>
        <input id="year" type="number" inputmode="numeric" placeholder="2024 atau -800 (SM)"
               bind:value={year} class:invalid={errors.year} />
        {#if errors.year}<div class="err">{errors.year}</div>{/if}
      </div>
    </div>

    <div class="field" data-field="genre">
      <span class="label" id="genre-label">Jenis buku</span>
      <div class="genre-grid" role="group" aria-labelledby="genre-label">
        {#each GENRES as g (g.key)}
          <button type="button" class="genre-card" class:active={genre === g.key}
                  aria-pressed={genre === g.key}
                  on:click={() => { genre = g.key; errors.genre = ''; }}>
            {g.label}
          </button>
        {/each}
      </div>
      {#if errors.genre}<div class="err">{errors.genre}</div>{/if}
    </div>
  </section>

  <!-- COVER -->
  <section class="group">
    <div class="field" data-field="cover">
      <span class="label" id="cover-label">Cover</span>
      <div class="segmented" role="group" aria-labelledby="cover-label">
        <button type="button" class:active={coverMode === 'uploaded'}
                aria-pressed={coverMode === 'uploaded'}
                on:click={() => { coverMode = 'uploaded'; errors.cover = ''; }}>
          Upload gambar
        </button>
        <button type="button" class:active={coverMode === 'generated'}
                aria-pressed={coverMode === 'generated'}
                on:click={() => { coverMode = 'generated'; errors.cover = ''; }}>
          Buat otomatis
        </button>
      </div>
      {#if errors.cover}<div class="err">{errors.cover}</div>{/if}
    </div>

    {#if coverMode === 'uploaded'}
      <div class="cover-panel">
        <div class="cover-inputs" data-field="coverFile">
          <input type="file" accept=".png,.jpg,.jpeg,image/png,image/jpeg"
                 on:change={onCoverChange} aria-label="File cover" />
          <div class="hint">
            PNG atau JPG, maksimal 5 MB.
            {#if mode === 'edit' && existingCoverURL && !coverFile}
              Kosongkan untuk mempertahankan cover lama.
            {/if}
          </div>
          {#if errors.coverFile}<div class="err">{errors.coverFile}</div>{/if}
        </div>

        {#if coverPreviewURL || existingCoverURL}
          <div class="preview">
            <img src={coverPreviewURL || existingCoverURL}
                 alt={coverPreviewURL ? 'Preview cover' : 'Cover saat ini'} />
          </div>
        {/if}
      </div>
    {/if}

    {#if coverMode === 'generated'}
      <div class="cover-panel">
        <div class="cover-inputs">
          <div class="sub-field" data-field="coverColor">
            <label class="mini-label" for="coverColor">Warna latar</label>
            <div class="color-row">
              <input id="coverColor" type="color" bind:value={coverColor} />
              <input type="text" class="hex" bind:value={coverColor}
                     placeholder="#9c7450" aria-label="Kode warna hex" />
            </div>
            {#if errors.coverColor}<div class="err">{errors.coverColor}</div>{/if}
          </div>

          <div class="sub-field" data-field="coverText">
            <label class="mini-label" for="coverText">
              <span>Teks cover</span>
              <span class="counter">{[...coverText].length}/{MAX_COVER_TEXT}</span>
            </label>
            <input id="coverText" type="text" placeholder="Judul singkat atau inisial"
                   bind:value={coverText} maxlength={MAX_COVER_TEXT} />
            {#if errors.coverText}<div class="err">{errors.coverText}</div>{/if}
          </div>
        </div>

        <div class="preview">
          <div class="gen-cover"
               style="background:{coverColor}; color:{textColor}; font-size:{genFontSize}px;">
            {coverText || 'Preview'}
          </div>
        </div>
      </div>
    {/if}
  </section>

  <!-- FILE & SINOPSIS -->
  <section class="group">
    <div class="field" data-field="pdfFile">
      <label for="pdf">File PDF</label>
      <input id="pdf" type="file" accept=".pdf,application/pdf" on:change={onPdfChange} />
      <div class="hint">
        Maksimal 50 MB.
        {#if mode === 'edit' && book?.pdf_path && !pdfFile}
          Kosongkan untuk mempertahankan PDF lama.
        {/if}
      </div>
      {#if pdfFile}
        <div class="file-info">{pdfFile.name} ({(pdfFile.size / 1024 / 1024).toFixed(2)} MB)</div>
      {:else if mode === 'edit' && book?.pdf_path}
        <div class="file-info">{book.pdf_path.split('/').pop()} (file saat ini)</div>
      {/if}
      {#if errors.pdfFile}<div class="err">{errors.pdfFile}</div>{/if}
    </div>

    <div class="field" data-field="synopsis">
      <label for="synopsis">Sinopsis <span class="opt">(opsional)</span></label>
      <textarea id="synopsis" rows="5" placeholder="Ringkasan singkat isi buku"
                bind:value={synopsis} class:invalid={errors.synopsis}></textarea>
      {#if errors.synopsis}<div class="err">{errors.synopsis}</div>{/if}
    </div>
  </section>

  {#if serverError}<div class="banner error" role="alert">{serverError}</div>{/if}
  {#if successMessage}<div class="banner success" role="status">{successMessage}</div>{/if}

  <div class="actions">
    <a href={mode === 'edit' ? `/read/${book.id}` : '/'} class="btn cancel">Batal</a>
    <button type="submit" class="btn primary" disabled={submitting}>
      {submitting ? 'Menyimpan…' : (mode === 'edit' ? 'Simpan perubahan' : 'Upload buku')}
    </button>
  </div>
</form>

<style>
  form { display: flex; flex-direction: column; gap: 32px; }

  .group { display: flex; flex-direction: column; gap: 20px; }
  .row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }

  .field { display: flex; flex-direction: column; gap: 6px; min-width: 0; }
  label, .label { font-size: 13px; font-weight: 500; color: var(--text); }
  .opt { color: var(--text-secondary); font-weight: 400; font-size: 12px; }

  input[type="text"], input[type="number"], textarea {
    width: 100%;
    box-sizing: border-box;
    padding: 11px 12px;
    background: var(--white);
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    font-size: 16px; /* cegah auto-zoom iOS */
    font-family: inherit;
    color: var(--text);
    outline: none;
    transition: border-color 0.15s;
  }
  input[type="text"]:focus, input[type="number"]:focus, textarea:focus {
    border-color: var(--accent);
  }
  input.invalid, textarea.invalid { border-color: #c97a7a; }
  textarea { resize: vertical; min-height: 110px; font-family: var(--font-body); line-height: 1.5; }
  input[type="file"] { font-size: 13px; padding: 6px 0; max-width: 100%; }

  .hint { font-size: 12px; color: var(--text-secondary); }
  .err { font-size: 12px; color: #b04a4a; }
  .file-info {
    font-size: 12.5px;
    color: var(--text);
    background: var(--surface);
    padding: 6px 10px;
    border-radius: 6px;
    align-self: flex-start;
    max-width: 100%;
    overflow-wrap: anywhere;
  }

  /* Segmented control */
  .segmented {
    display: grid;
    grid-template-columns: 1fr 1fr;
    padding: 3px;
    background: var(--surface);
    border-radius: var(--radius-sm);
    gap: 3px;
  }
  .segmented button {
    padding: 9px 10px;
    border-radius: calc(var(--radius-sm) - 2px);
    font-size: 13.5px;
    color: var(--text-secondary);
    transition: all 0.15s;
  }
  .segmented button.active {
    background: var(--white);
    color: var(--accent-dark);
    font-weight: 500;
    box-shadow: 0 1px 2px rgba(70, 61, 54, 0.15);
  }

  /* Panel cover */
  .cover-panel {
    display: grid;
    grid-template-columns: 1fr auto;
    gap: 20px;
    align-items: start;
  }
  .cover-inputs { display: flex; flex-direction: column; gap: 14px; min-width: 0; }
  .sub-field { display: flex; flex-direction: column; gap: 6px; }
  .mini-label {
    font-size: 12.5px;
    color: var(--text-secondary);
    font-weight: 400;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .counter { font-size: 11.5px; }
  .color-row { display: flex; gap: 10px; align-items: center; }
  input[type="color"] {
    width: 46px;
    height: 44px;
    padding: 2px;
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    background: var(--white);
    cursor: pointer;
    flex-shrink: 0;
  }
  .hex { flex: 1; }

  .preview { display: flex; justify-content: center; }
  .preview img,
  .gen-cover {
    width: 96px;
    height: 144px;
    border-radius: 4px;
    box-shadow: 0 1px 2px rgba(70, 61, 54, 0.18), 0 6px 14px -8px rgba(70, 61, 54, 0.35);
  }
  .preview img { object-fit: cover; display: block; }
  .gen-cover {
    display: flex;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 8px;
    box-sizing: border-box;
    font-family: var(--font-heading);
    font-weight: 500;
    line-height: 1.15;
    word-break: break-word;
  }

  /* Genre */
  .genre-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 8px; }
  .genre-card {
    padding: 11px 6px;
    background: var(--white);
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    font-size: 14px;
    color: var(--text);
    transition: all 0.15s;
  }
  .genre-card:hover { border-color: var(--accent); }
  .genre-card.active {
    background: var(--surface);
    border-color: var(--accent);
    color: var(--accent-dark);
    font-weight: 500;
  }

  .banner { padding: 12px 16px; border-radius: var(--radius-sm); font-size: 13.5px; }
  .banner.error { background: #fbeaea; color: #a44; border: 1px solid #f0cfcf; }
  .banner.success { background: #eaf3ea; color: #4a7a4a; border: 1px solid #cfe4cf; }

  .actions {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
    padding-top: 20px;
    border-top: 1px solid var(--border);
  }
  .btn {
    padding: 11px 20px;
    border-radius: var(--radius-sm);
    font-size: 14px;
    font-weight: 500;
    text-align: center;
    transition: all 0.15s;
  }
  .btn.cancel {
    background: var(--white);
    border: 1px solid var(--border);
    color: var(--text-secondary);
  }
  .btn.cancel:hover { border-color: var(--text-secondary); color: var(--text); }
  .btn.primary { background: var(--accent); color: var(--white); }
  .btn.primary:hover { background: var(--accent-dark); }
  .btn.primary:disabled { opacity: 0.6; cursor: not-allowed; }

  @media (max-width: 720px) {
    .row { grid-template-columns: 1fr; }
    .genre-grid { grid-template-columns: repeat(2, 1fr); }
    .cover-panel { grid-template-columns: 1fr; }
    .preview { justify-content: flex-start; }
    .actions { flex-direction: column-reverse; }
    .btn { padding: 13px 20px; }
  }
</style>