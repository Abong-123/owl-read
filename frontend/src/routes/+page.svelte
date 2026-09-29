<script>
  import { onMount } from 'svelte';
  import { api } from '$lib/api.js';
  import Cover from '$lib/Cover.svelte';

  const GENRES = [
    { key: 'novel',     label: 'Novel' },
    { key: 'comic',     label: 'Comic' },
    { key: 'majalah',   label: 'Majalah' },
    { key: 'non-fiksi', label: 'Non-fiksi' }
  ];

  let books = [];
  let loading = true;
  let error = '';
  let query = '';
  let activeGenre = ''; // '' = semua

  onMount(load);

  async function load() {
    loading = true;
    error = '';
    try {
      const data = await api.listBooks();
      books = data.books || [];
    } catch (e) {
      error = e.message || 'Gagal memuat buku';
    } finally {
      loading = false;
    }
  }

  // Filter client-side: judul + penulis.
  $: filtered = query.trim()
    ? books.filter((b) => {
        const q = query.toLowerCase();
        return (
          b.title.toLowerCase().includes(q) ||
          (b.author || '').toLowerCase().includes(q)
        );
      })
    : books;

  $: shelves = GENRES
    .filter((g) => !activeGenre || g.key === activeGenre)
    .map((g) => ({ ...g, books: filtered.filter((b) => b.genre === g.key) }))
    .filter((s) => s.books.length > 0);

  $: noResults = !loading && !error && books.length > 0 && shelves.length === 0;
  $: isEmpty = !loading && !error && books.length === 0;
</script>

<svelte:head>
  <title>Readowl Library</title>
</svelte:head>

<div class="container">
  <header class="topbar">
    <a href="/" class="brand">Readowl</a>
    <a href="/input" class="upload-btn">+ Upload</a>
  </header>

  <div class="intro">
    <h1>Perpustakaan</h1>
    <span class="total">{books.length} buku</span>
  </div>

  <div class="search-wrap">
    <input
      type="search"
      class="search"
      placeholder="Cari judul atau penulis"
      bind:value={query}
      aria-label="Cari buku"
    />
    {#if query}
      <button class="clear" on:click={() => (query = '')} aria-label="Hapus pencarian">×</button>
    {/if}
  </div>

  <div class="chips" role="tablist" aria-label="Filter genre">
    <button class="chip" class:active={activeGenre === ''} on:click={() => (activeGenre = '')}>
      Semua
    </button>
    {#each GENRES as g (g.key)}
      <button
        class="chip"
        class:active={activeGenre === g.key}
        on:click={() => (activeGenre = activeGenre === g.key ? '' : g.key)}
      >
        {g.label}
      </button>
    {/each}
  </div>

  {#if loading}
    <div class="state">Memuat perpustakaan…</div>
  {:else if error}
    <div class="state error">
      {error}
      <button class="retry" on:click={load}>Coba lagi</button>
    </div>
  {:else if isEmpty}
    <div class="state">
      <div class="state-title">Perpustakaan masih kosong</div>
      <p>Upload buku pertama untuk mulai.</p>
      <a class="cta" href="/input">Upload buku</a>
    </div>
  {:else if noResults}
    <div class="state">
      <div class="state-title">Buku tidak ditemukan</div>
      <p>Coba judul, penulis, atau genre lain.</p>
    </div>
  {:else}
    {#each shelves as shelf (shelf.key)}
      <section class="shelf-section">
        <h2 class="shelf-title">
          {shelf.label}
          <span class="shelf-count">{shelf.books.length}</span>
        </h2>

        <div class="shelf" class:grid={!!activeGenre}>
          {#each shelf.books as book (book.id)}
            <a class="book-card" href="/read/{book.id}">
              <Cover {book} size={112} />
              <div class="book-title" title={book.title}>{book.title}</div>
              {#if book.author}
                <div class="book-author" title={book.author}>{book.author}</div>
              {/if}
            </a>
          {/each}
        </div>
      </section>
    {/each}
  {/if}
</div>

<style>
  /* ---------- TOPBAR ---------- */
  .topbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 28px;
  }
  .brand {
    font-family: var(--font-heading);
    font-size: 24px;
    color: var(--accent-dark);
    letter-spacing: -0.01em;
  }
  .upload-btn {
    padding: 8px 14px;
    border: 1px solid var(--accent);
    border-radius: var(--radius-sm);
    color: var(--accent);
    font-size: 14px;
    font-weight: 500;
    transition: background 0.15s, color 0.15s;
  }
  .upload-btn:hover,
  .upload-btn:focus-visible {
    background: var(--accent);
    color: var(--white);
  }

  /* ---------- INTRO ---------- */
  .intro {
    display: flex;
    align-items: baseline;
    gap: 10px;
    margin-bottom: 16px;
  }
  .intro h1 { font-size: 28px; line-height: 1.15; }
  .total { font-size: 13px; color: var(--text-secondary); }

  /* ---------- SEARCH ---------- */
  .search-wrap { position: relative; margin-bottom: 14px; }
  .search {
    width: 100%;
    box-sizing: border-box;
    padding: 11px 38px 11px 14px;
    background: var(--white);
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    font-size: 16px; /* cegah auto-zoom iOS */
    outline: none;
    transition: border-color 0.15s;
    -webkit-appearance: none;
    appearance: none;
  }
  .search::-webkit-search-cancel-button { display: none; }
  .search:focus { border-color: var(--accent); }
  .clear {
    position: absolute;
    right: 6px;
    top: 50%;
    transform: translateY(-50%);
    width: 30px;
    height: 30px;
    border-radius: 50%;
    color: var(--text-secondary);
    font-size: 18px;
    line-height: 1;
  }
  .clear:hover { background: var(--surface); color: var(--text); }

  /* ---------- CHIPS ---------- */
  .chips {
    display: flex;
    gap: 8px;
    margin-bottom: 32px;
    overflow-x: auto;
    padding-bottom: 2px;
    scrollbar-width: none;
  }
  .chips::-webkit-scrollbar { display: none; }
  .chip {
    flex: 0 0 auto;
    padding: 6px 14px;
    border: 1px solid var(--border);
    border-radius: 999px;
    background: var(--white);
    font-size: 13px;
    color: var(--text-secondary);
    transition: all 0.15s;
  }
  .chip:hover { border-color: var(--accent); color: var(--accent); }
  .chip.active {
    background: var(--accent);
    border-color: var(--accent);
    color: var(--white);
  }

  /* ---------- SHELF ---------- */
  .shelf-section { margin-bottom: 36px; }
  .shelf-title {
    display: flex;
    align-items: baseline;
    gap: 8px;
    font-size: 18px;
    margin-bottom: 14px;
  }
  .shelf-count {
    font-family: var(--font-body);
    font-size: 12.5px;
    font-weight: 400;
    color: var(--text-secondary);
  }

  .shelf {
    display: flex;
    gap: 18px;
    overflow-x: auto;
    padding: 4px 2px 12px;
    scroll-snap-type: x proximity;
    scrollbar-width: thin;
    -webkit-overflow-scrolling: touch;
  }
  .shelf.grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(112px, 1fr));
    gap: 22px 16px;
    overflow: visible;
  }

  .book-card {
    flex: 0 0 112px;
    width: 112px;
    scroll-snap-align: start;
    display: flex;
    flex-direction: column;
    gap: 2px;
    transition: transform 0.2s ease;
  }
  .grid .book-card { flex: initial; width: auto; align-items: center; }
  .book-card:hover,
  .book-card:focus-visible { transform: translateY(-4px); }
  .book-card:hover :global(.cover) {
    box-shadow: 0 12px 20px -8px rgba(70, 61, 54, 0.45);
  }

  .book-title {
    margin-top: 8px;
    font-size: 13px;
    font-weight: 600;
    line-height: 1.3;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
  .grid .book-title,
  .grid .book-author { text-align: center; width: 100%; }
  .book-author {
    font-size: 12px;
    color: var(--text-secondary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  /* ---------- STATE ---------- */
  .state {
    padding: 56px 16px;
    text-align: center;
    color: var(--text-secondary);
    font-size: 14px;
  }
  .state.error { color: #a44; }
  .state-title {
    font-family: var(--font-heading);
    font-size: 20px;
    color: var(--text);
    margin-bottom: 6px;
  }
  .state p { margin: 0 0 16px; }
  .cta,
  .retry {
    display: inline-block;
    margin-top: 8px;
    padding: 9px 18px;
    background: var(--accent);
    color: var(--white);
    border-radius: var(--radius-sm);
    font-size: 14px;
    font-weight: 500;
  }
  .cta:hover, .retry:hover { background: var(--accent-dark); }

  /* ---------- RESPONSIVE ---------- */
  @media (min-width: 721px) {
    .shelf { gap: 24px; }
    .intro h1 { font-size: 32px; }
  }
</style>