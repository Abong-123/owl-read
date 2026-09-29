<script>
  import BookForm from '$lib/BookForm.svelte';
</script>

<svelte:head>
  <title>Tambah buku — Readowl</title>
</svelte:head>

<div class="container narrow">
  <a href="/" class="back">← Library</a>
  <header class="head">
    <h1>Tambah buku</h1>
    <p>Isi informasi buku untuk menambahkannya ke perpustakaan.</p>
  </header>

  <BookForm mode="create" book={null} />
</div>

<style>
  .narrow { max-width: 640px; }
  .back {
    display: inline-block;
    font-size: 13px;
    color: var(--text-secondary);
    margin-bottom: 20px;
    transition: color 0.15s;
  }
  .back:hover { color: var(--accent); }
  .head { margin-bottom: 28px; }
  .head h1 { font-size: 28px; margin-bottom: 6px; }
  .head p { color: var(--text-secondary); font-size: 14px; margin: 0; }
</style>