<script>
  export let book;
  export let size = 118; // lebar cover (px)

  $: height = Math.round(size * 1.5); // rasio 2:3
  $: isUploaded = book?.cover_type === 'uploaded' && book?.cover_path;

  $: text = book?.cover_text || '';
  $: fontSize = calcFontSize([...text].length, size);
  $: bg = book?.cover_color || '#9c7450';
  $: textColor = isLight(bg) ? '#463d36' : '#fffdf9';

  function calcFontSize(len, w) {
    if (len <= 4)  return Math.round(w * 0.20);
    if (len <= 8)  return Math.round(w * 0.16);
    if (len <= 12) return Math.round(w * 0.13);
    if (len <= 15) return Math.round(w * 0.11);
    return Math.round(w * 0.095);
  }

  function isLight(hex) {
    const c = (hex || '#9c7450').replace('#', '');
    const full = c.length === 3 ? c.split('').map((x) => x + x).join('') : c;
    const r = parseInt(full.slice(0, 2), 16);
    const g = parseInt(full.slice(2, 4), 16);
    const b = parseInt(full.slice(4, 6), 16);
    return 0.299 * r + 0.587 * g + 0.114 * b > 160;
  }
</script>

<div class="cover" style="width:{size}px; height:{height}px;">
  {#if isUploaded}
    <img src="/{book.cover_path}" alt={book.title} loading="lazy" />
  {:else}
    <div
      class="generated"
      style="background:{bg}; color:{textColor}; font-size:{fontSize}px;"
    >
      {text}
    </div>
  {/if}
</div>

<style>
  .cover {
    position: relative;
    flex-shrink: 0;
    border-radius: 4px;
    overflow: hidden;
    background: var(--surface);
    box-shadow: 0 1px 2px rgba(70, 61, 54, 0.18), 0 6px 14px -8px rgba(70, 61, 54, 0.35);
    transition: box-shadow 0.2s ease;
  }

  /* Efek lipatan punggung buku tipis */
  .cover::after {
    content: '';
    position: absolute;
    inset: 0 auto 0 0;
    width: 4px;
    background: linear-gradient(to right, rgba(0, 0, 0, 0.12), transparent);
    pointer-events: none;
  }

  .cover img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }

  .generated {
    width: 100%;
    height: 100%;
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
</style>