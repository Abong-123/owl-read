<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { api } from '$lib/api.js';
  import Cover from '$lib/Cover.svelte';

  let book = null;
  let loading = true;
  let error = '';

  let deleting = false;
  let deleteError = '';

  $: id = $page.params.id;

  onMount(load);

  // Navigasi /read/1 → /read/2 tanpa reload.
  $: if (id) load();

  async function load() {
    if (!id) return;
    loading = true;
    error = '';
    try {
      book = await api.getBook(id);
    } catch (e) {
      error = e.message || 'Buku tidak ditemukan';
      book = null;
    } finally {
      loading = false;
    }
  }

  async function handleDelete() {
    deleteError = '';

    if (!confirm(`Hapus "${book.title}"? Tindakan ini tidak bisa dibatalkan.`)) return;

    const code = prompt('Masukkan kode upload (4 digit):');
    if (code === null) return;
    if (!/^\d{4}$/.test(code)) { deleteError = 'Kode harus 4 digit angka.'; return; }
    if (code !== '1234') { deleteError = 'Kode salah.'; return; }

    deleting = true;
    try {
      await api.deleteBook(book.id);
      goto('/');
    } catch (e) {
      deleteError = e.message || 'Gagal menghapus buku.';
      deleting = false;
    }
  }

  function formatYear(y) {
    if (y === null || y === undefined) return '';
    if (y < 0) return `${-y} SM`;
    return String(y);
  }

  const GENRE_LABEL = {
    novel: 'Novel',
    comic: 'Comic',
    majalah: 'Majalah',
    'non-fiksi': 'Non-fiksi'
  };
</script>

<svelte:head>
  <title>{book ? book.title : 'Memuat…'} — Readowl</title>
</svelte:head>

<div class="container">
  <a href="/" class="back">← Library</a>

  {#if loading}
    <div class="state">Memuat buku…</div>
  {:else if error}
    <div class="state error">{error}</div>
  {:else if book}
    <header class="book-head">
      <div class="cover-col">
        <Cover {book} size={150} />
      </div>

      <div class="meta-col">
        <div class="genre-tag">{GENRE_LABEL[book.genre] || book.genre}</div>
        <h1 class="title">{book.title}</h1>
        {#if book.author}
          <div class="author">oleh {book.author}</div>
        {/if}

        <dl class="details">
          {#if book.year !== null && book.year !== undefined && book.year !== ''}
            <div class="detail">
              <dt>Tahun</dt>
              <dd>{formatYear(book.year)}</dd>
            </div>
          {/if}
          <div class="detail">
            <dt>Ditambahkan</dt>
            <dd>
              {new Date(book.created_at).toLocaleDateString('id-ID', {
                day: 'numeric', month: 'long', year: 'numeric'
              })}
            </dd>
          </div>
        </dl>

        <div class="actions">
          <a class="btn primary" href="/{book.pdf_path}" target="_blank" rel="noopener">
            Buka PDF
          </a>
          <a class="btn" href="/{book.pdf_path}" download>Download</a>
          <a class="btn" href="/edit/{book.id}">Edit</a>
          <button class="btn danger" on:click={handleDelete} disabled={deleting}>
            {deleting ? 'Menghapus…' : 'Hapus'}
          </button>
        </div>

        {#if deleteError}
          <div class="delete-error" role="alert">{deleteError}</div>
        {/if}
      </div>
    </header>

    {#if book.synopsis}
      <section class="section">
        <h2 class="section-title">Sinopsis</h2>
        <p class="synopsis">{book.synopsis}</p>
      </section>
    {/if}

    <section class="section">
      <h2 class="section-title">Baca</h2>
      <div class="viewer">
        <iframe
          src="/{book.pdf_path}#toolbar=1&view=FitH"
          title="PDF viewer"
        ></iframe>
      </div>
      <div class="viewer-fallback">
        PDF tidak muncul? <a href="/{book.pdf_path}" target="_blank" rel="noopener">Buka di tab baru</a>.
      </div>
    </section>
  {/if}
</div>

<style>
  .back {
    display: inline-block;
    font-size: 13px;
    color: var(--text-secondary);
    margin-bottom: 24px;
    transition: color 0.15s;
  }
  .back:hover { color: var(--accent); }

  .state { padding: 60px 0; text-align: center; color: var(--text-secondary); }
  .state.error { color: #a44; }

  /* ---------- HEAD ---------- */
  .book-head {
    display: grid;
    grid-template-columns: 150px 1fr;
    gap: 32px;
    align-items: start;
    margin-bottom: 36px;
    padding-bottom: 32px;
    border-bottom: 1px solid var(--border);
  }
  .cover-col { display: flex; justify-content: center; }
  .meta-col { display: flex; flex-direction: column; gap: 10px; min-width: 0; }

  .genre-tag { font-size: 13px; color: var(--accent); font-weight: 500; }
  .title {
    font-size: 30px;
    line-height: 1.15;
    margin: 0;
    overflow-wrap: anywhere;
  }
  .author { font-size: 15px; color: var(--text-secondary); }

  .details { display: flex; flex-wrap: wrap; gap: 8px 28px; margin: 6px 0 0; }
  .detail { display: flex; flex-direction: column; gap: 2px; }
  .detail dt { font-size: 12px; color: var(--text-secondary); }
  .detail dd { margin: 0; font-size: 14px; color: var(--text); }

  .actions { display: flex; gap: 8px; margin-top: 12px; flex-wrap: wrap; }
  .btn {
    padding: 10px 16px;
    border-radius: var(--radius-sm);
    font-size: 14px;
    font-weight: 500;
    border: 1px solid var(--border);
    background: var(--white);
    color: var(--text);
    text-align: center;
    transition: all 0.15s;
  }
  .btn:hover { border-color: var(--accent); color: var(--accent); }
  .btn.primary { background: var(--accent); color: var(--white); border-color: var(--accent); }
  .btn.primary:hover { background: var(--accent-dark); border-color: var(--accent-dark); color: var(--white); }
  .btn.danger { border-color: #d99; color: #a44; }
  .btn.danger:hover { background: #a44; border-color: #a44; color: var(--white); }
  .btn:disabled { opacity: 0.6; cursor: not-allowed; }

  .delete-error {
    margin-top: 8px;
    padding: 10px 14px;
    background: #fbeaea;
    border: 1px solid #f0cfcf;
    color: #a44;
    border-radius: var(--radius-sm);
    font-size: 13px;
  }

  /* ---------- SECTIONS ---------- */
  .section { margin-bottom: 36px; }
  .section-title {
    font-size: 18px;
    margin-bottom: 12px;
  }
  .synopsis {
    font-size: 15px;
    line-height: 1.7;
    color: var(--text);
    white-space: pre-wrap;
    max-width: 68ch;
    margin: 0;
  }

  /* ---------- VIEWER ---------- */
  .viewer {
    width: 100%;
    height: 80vh;
    min-height: 480px;
    border-radius: var(--radius-sm);
    overflow: hidden;
    background: var(--surface);
    border: 1px solid var(--border);
  }
  .viewer iframe { width: 100%; height: 100%; border: 0; display: block; }
  .viewer-fallback {
    font-size: 13px;
    color: var(--text-secondary);
    margin-top: 10px;
    text-align: center;
  }
  .viewer-fallback a { color: var(--accent); text-decoration: underline; }

  /* ---------- RESPONSIVE ---------- */
  @media (max-width: 720px) {
    .book-head {
      grid-template-columns: 1fr;
      gap: 20px;
      text-align: center;
    }
    .meta-col { align-items: center; }
    .details { justify-content: center; }
    .title { font-size: 24px; }

    /* Tombol utama penuh, sisanya berbagi baris */
    .actions {
      width: 100%;
      display: grid;
      grid-template-columns: repeat(3, 1fr);
      gap: 8px;
    }
    .actions .btn.primary { grid-column: 1 / -1; padding: 13px 16px; }

    .viewer { height: 70vh; min-height: 400px; }
  }
</style>