import "../../../../chunks/index-server.js";
import { C as attr, d as unsubscribe_stores, l as store_get, o as head, w as escape_html } from "../../../../chunks/server.js";
import { t as api } from "../../../../chunks/api.js";
import { t as page } from "../../../../chunks/stores.js";
import { t as BookForm } from "../../../../chunks/BookForm.js";
//#region src/routes/edit/[id]/+page.svelte
function _page($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		var $$store_subs;
		let id;
		let book = null;
		let loading = true;
		let error = "";
		async function load() {
			if (!id) return;
			loading = true;
			error = "";
			try {
				book = await api.getBook(id);
			} catch (e) {
				error = e.message || "Buku tidak ditemukan";
				book = null;
			} finally {
				loading = false;
			}
		}
		$: id = store_get($$store_subs ??= {}, "$page", page).params.id;
		$: if (id) load();
		head("1ul2axt", $$renderer, ($$renderer) => {
			$$renderer.title(($$renderer) => {
				$$renderer.push(`<title>Edit buku — Readowl</title>`);
			});
		});
		$$renderer.push(`<div class="container narrow svelte-1ul2axt"><a${attr("href", book ? `/read/${book.id}` : "/")} class="back svelte-1ul2axt">← Kembali</a> `);
		if (loading) $$renderer.push(`<!--[0--><div class="state svelte-1ul2axt">Memuat…</div>`);
		else if (error) $$renderer.push(`<!--[1--><div class="state error svelte-1ul2axt">${escape_html(error)}</div>`);
		else if (book) {
			$$renderer.push(`<!--[2--><header class="head svelte-1ul2axt"><h1 class="svelte-1ul2axt">Edit buku</h1> <p class="svelte-1ul2axt">Perbarui informasi buku.</p></header> `);
			BookForm($$renderer, {
				mode: "edit",
				book
			});
			$$renderer.push(`<!---->`);
		} else $$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--></div>`);
		if ($$store_subs) unsubscribe_stores($$store_subs);
	});
}
//#endregion
export { _page as default };
