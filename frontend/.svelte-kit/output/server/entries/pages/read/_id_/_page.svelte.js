import "../../../../chunks/index-server.js";
import { C as attr, d as unsubscribe_stores, l as store_get, o as head, u as stringify, w as escape_html } from "../../../../chunks/server.js";
import { t as api } from "../../../../chunks/api.js";
import { t as Cover } from "../../../../chunks/Cover.js";
import { t as page } from "../../../../chunks/stores.js";
import "../../../../chunks/navigation.js";
//#region src/routes/read/[id]/+page.svelte
function _page($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		var $$store_subs;
		let id;
		let book = null;
		let loading = true;
		let error = "";
		let deleting = false;
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
		function formatYear(y) {
			if (y === null || y === void 0) return "";
			if (y < 0) return `${-y} SM`;
			return String(y);
		}
		const GENRE_LABEL = {
			novel: "Novel",
			comic: "Comic",
			majalah: "Majalah",
			"non-fiksi": "Non-fiksi"
		};
		$: id = store_get($$store_subs ??= {}, "$page", page).params.id;
		$: if (id) load();
		head("hze91l", $$renderer, ($$renderer) => {
			$$renderer.title(($$renderer) => {
				$$renderer.push(`<title>${escape_html(book ? book.title : "Memuat…")} — Readowl</title>`);
			});
		});
		$$renderer.push(`<div class="container"><a href="/" class="back svelte-hze91l">← Library</a> `);
		if (loading) $$renderer.push(`<!--[0--><div class="state svelte-hze91l">Memuat buku…</div>`);
		else if (error) $$renderer.push(`<!--[1--><div class="state error svelte-hze91l">${escape_html(error)}</div>`);
		else if (book) {
			$$renderer.push(`<!--[2--><header class="book-head svelte-hze91l"><div class="cover-col svelte-hze91l">`);
			Cover($$renderer, {
				book,
				size: 150
			});
			$$renderer.push(`<!----></div> <div class="meta-col svelte-hze91l"><div class="genre-tag svelte-hze91l">${escape_html(GENRE_LABEL[book.genre] || book.genre)}</div> <h1 class="title svelte-hze91l">${escape_html(book.title)}</h1> `);
			if (book.author) $$renderer.push(`<!--[0--><div class="author svelte-hze91l">oleh ${escape_html(book.author)}</div>`);
			else $$renderer.push("<!--[-1-->");
			$$renderer.push(`<!--]--> <dl class="details svelte-hze91l">`);
			if (book.year !== null && book.year !== void 0 && book.year !== "") $$renderer.push(`<!--[0--><div class="detail svelte-hze91l"><dt class="svelte-hze91l">Tahun</dt> <dd class="svelte-hze91l">${escape_html(formatYear(book.year))}</dd></div>`);
			else $$renderer.push("<!--[-1-->");
			$$renderer.push(`<!--]--> <div class="detail svelte-hze91l"><dt class="svelte-hze91l">Ditambahkan</dt> <dd class="svelte-hze91l">${escape_html(new Date(book.created_at).toLocaleDateString("id-ID", {
				day: "numeric",
				month: "long",
				year: "numeric"
			}))}</dd></div></dl> <div class="actions svelte-hze91l"><a class="btn primary svelte-hze91l"${attr("href", `/${stringify(book.pdf_path)}`)} target="_blank" rel="noopener">Buka PDF</a> <a class="btn svelte-hze91l"${attr("href", `/${stringify(book.pdf_path)}`)} download="">Download</a> <a class="btn svelte-hze91l"${attr("href", `/edit/${stringify(book.id)}`)}>Edit</a> <button class="btn danger svelte-hze91l"${attr("disabled", deleting, true)}>${escape_html("Hapus")}</button></div> `);
			$$renderer.push("<!--[-1-->");
			$$renderer.push(`<!--]--></div></header> `);
			if (book.synopsis) $$renderer.push(`<!--[0--><section class="section svelte-hze91l"><h2 class="section-title svelte-hze91l">Sinopsis</h2> <p class="synopsis svelte-hze91l">${escape_html(book.synopsis)}</p></section>`);
			else $$renderer.push("<!--[-1-->");
			$$renderer.push(`<!--]--> <section class="section svelte-hze91l"><h2 class="section-title svelte-hze91l">Baca</h2> <div class="viewer svelte-hze91l"><iframe${attr("src", `/${stringify(book.pdf_path)}#toolbar=1&view=FitH`)} title="PDF viewer" class="svelte-hze91l"></iframe></div> <div class="viewer-fallback svelte-hze91l">PDF tidak muncul? <a${attr("href", `/${stringify(book.pdf_path)}`)} target="_blank" rel="noopener" class="svelte-hze91l">Buka di tab baru</a>.</div></section>`);
		} else $$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--></div>`);
		if ($$store_subs) unsubscribe_stores($$store_subs);
	});
}
//#endregion
export { _page as default };
