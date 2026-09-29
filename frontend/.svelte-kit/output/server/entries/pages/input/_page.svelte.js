import { o as head } from "../../../chunks/server.js";
import { t as BookForm } from "../../../chunks/BookForm.js";
//#region src/routes/input/+page.svelte
function _page($$renderer) {
	head("rx1dtf", $$renderer, ($$renderer) => {
		$$renderer.title(($$renderer) => {
			$$renderer.push(`<title>Tambah buku — Readowl</title>`);
		});
	});
	$$renderer.push(`<div class="container narrow svelte-rx1dtf"><a href="/" class="back svelte-rx1dtf">← Library</a> <header class="head svelte-rx1dtf"><h1 class="svelte-rx1dtf">Tambah buku</h1> <p class="svelte-rx1dtf">Isi informasi buku untuk menambahkannya ke perpustakaan.</p></header> `);
	BookForm($$renderer, {
		mode: "create",
		book: null
	});
	$$renderer.push(`<!----></div>`);
}
//#endregion
export { _page as default };
