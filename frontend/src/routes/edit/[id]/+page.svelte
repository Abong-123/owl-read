<script>
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { api } from '$lib/api.js';
  import BookForm from '$lib/BookForm.svelte';

  let book = null;
  let loading = true;
  let error = '';

  $: id = $page.params.id;
  onMount(load);
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
</script>

<svelte:head>
  <title>Edit buku — Readowl</title>
</svelte:head>

<div class="container narrow">
  <a href={book ? `/read/${book.id}` : '/'} class="back">← Kembali</a>

  {#if loading}
    <div class="state">Memuat…</div>
  {:else if error}
    <div class="state error">{error}</div>
  {:else if book}
    <header class="head">
      <h1>Edit buku</h1>
      <p>Perbarui informasi buku.</p>
    </header>
    <BookForm mode="edit" {book} />
  {/if}
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
  .state { padding: 60px 0; text-align: center; color: var(--text-secondary); }
  .state.error { color: #a44; }
</style>