import "../../chunks/index-server.js";
import { C as attr, a as ensure_array_like, o as head, t as attr_class, w as escape_html } from "../../chunks/server.js";
import "../../chunks/api.js";
import "../../chunks/Cover.js";
//#region src/routes/+page.svelte
function _page($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		let filtered;
		const GENRES = [
			{
				key: "novel",
				label: "Novel"
			},
			{
				key: "comic",
				label: "Comic"
			},
			{
				key: "majalah",
				label: "Majalah"
			},
			{
				key: "non-fiksi",
				label: "Non-fiksi"
			}
		];
		let books = [];
		let query = "";
		let activeGenre = "";
		$: filtered = (query.trim(), books);
		$: GENRES.filter((g) => true).map((g) => ({
			...g,
			books: filtered.filter((b) => b.genre === g.key)
		})).filter((s) => s.books.length > 0);
		head("1uha8ag", $$renderer, ($$renderer) => {
			$$renderer.title(($$renderer) => {
				$$renderer.push(`<title>Readowl Library</title>`);
			});
		});
		$$renderer.push(`<div class="container"><header class="topbar svelte-1uha8ag"><a href="/" class="brand svelte-1uha8ag">Readowl</a> <a href="/input" class="upload-btn svelte-1uha8ag">+ Upload</a></header> <div class="intro svelte-1uha8ag"><h1 class="svelte-1uha8ag">Perpustakaan</h1> <span class="total svelte-1uha8ag">${escape_html(books.length)} buku</span></div> <div class="search-wrap svelte-1uha8ag"><input type="search" class="search svelte-1uha8ag" placeholder="Cari judul atau penulis"${attr("value", query)} aria-label="Cari buku"/> `);
		$$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--></div> <div class="chips svelte-1uha8ag" role="tablist" aria-label="Filter genre"><button${attr_class("chip svelte-1uha8ag", void 0, { "active": true })}>Semua</button> <!--[-->`);
		const each_array = ensure_array_like(GENRES);
		for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
			let g = each_array[$$index];
			$$renderer.push(`<button${attr_class("chip svelte-1uha8ag", void 0, { "active": activeGenre === g.key })}>${escape_html(g.label)}</button>`);
		}
		$$renderer.push(`<!--]--></div> `);
		$$renderer.push(`<!--[0--><div class="state svelte-1uha8ag">Memuat perpustakaan…</div>`);
		$$renderer.push(`<!--]--></div>`);
	});
}
//#endregion
export { _page as default };
