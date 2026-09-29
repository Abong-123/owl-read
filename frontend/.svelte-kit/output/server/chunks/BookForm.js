import { C as attr, a as ensure_array_like, kt as fallback, n as attr_style, r as bind_props, t as attr_class, u as stringify, w as escape_html } from "./server.js";
import "./navigation.js";
//#region src/lib/BookForm.svelte
function BookForm($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		let existingCoverURL, genFontSize, textColor;
		let book = fallback($$props["book"], null);
		let mode = fallback($$props["mode"], "create");
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
		const MAX_COVER_TEXT = 18;
		let title = book?.title ?? "";
		let author = book?.author ?? "";
		let year = book?.year ?? "";
		let synopsis = book?.synopsis ?? "";
		let genre = book?.genre ?? "";
		let coverMode = book?.cover_type ?? "";
		let coverColor = book?.cover_color || "#9c7450";
		let coverText = book?.cover_text ?? "";
		let errors = {};
		let submitting = false;
		function calcFontSize(len) {
			if (len <= 4) return 26;
			if (len <= 8) return 20;
			if (len <= 12) return 16;
			if (len <= 15) return 14;
			return 12;
		}
		function isLight(hex) {
			const c = (hex || "#9c7450").replace("#", "");
			const full = c.length === 3 ? c.split("").map((x) => x + x).join("") : c;
			const r = parseInt(full.slice(0, 2), 16);
			const g = parseInt(full.slice(2, 4), 16);
			const b = parseInt(full.slice(4, 6), 16);
			return .299 * r + .587 * g + .114 * b > 160;
		}
		let coverPreviewURL = "";
		$: existingCoverURL = book?.cover_type === "uploaded" && book?.cover_path ? "/" + book.cover_path : "";
		$: genFontSize = calcFontSize([...coverText].length);
		$: textColor = isLight(coverColor) ? "#463d36" : "#fffdf9";
		$: {
			if (coverPreviewURL) URL.revokeObjectURL(coverPreviewURL);
			coverPreviewURL = "";
		}
		$$renderer.push(`<form novalidate="" class="svelte-79z2zc"><section class="group svelte-79z2zc"><div class="field svelte-79z2zc" data-field="title"><label for="title" class="svelte-79z2zc">Judul</label> <input id="title" type="text" placeholder="Judul buku" autocomplete="off"${attr("value", title)}${attr_class("svelte-79z2zc", void 0, { "invalid": errors.title })}/> `);
		if (errors.title) $$renderer.push(`<!--[0--><div class="err svelte-79z2zc">${escape_html(errors.title)}</div>`);
		else $$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--></div> <div class="row svelte-79z2zc"><div class="field svelte-79z2zc" data-field="author"><label for="author" class="svelte-79z2zc">Penulis <span class="opt svelte-79z2zc">(opsional)</span></label> <input id="author" type="text" placeholder="Nama penulis" autocomplete="off"${attr("value", author)}${attr_class("svelte-79z2zc", void 0, { "invalid": errors.author })}/> `);
		if (errors.author) $$renderer.push(`<!--[0--><div class="err svelte-79z2zc">${escape_html(errors.author)}</div>`);
		else $$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--></div> <div class="field svelte-79z2zc" data-field="year"><label for="year" class="svelte-79z2zc">Tahun <span class="opt svelte-79z2zc">(opsional)</span></label> <input id="year" type="number" inputmode="numeric" placeholder="2024 atau -800 (SM)"${attr("value", year)}${attr_class("svelte-79z2zc", void 0, { "invalid": errors.year })}/> `);
		if (errors.year) $$renderer.push(`<!--[0--><div class="err svelte-79z2zc">${escape_html(errors.year)}</div>`);
		else $$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--></div></div> <div class="field svelte-79z2zc" data-field="genre"><span class="label svelte-79z2zc" id="genre-label">Jenis buku</span> <div class="genre-grid svelte-79z2zc" role="group" aria-labelledby="genre-label"><!--[-->`);
		const each_array = ensure_array_like(GENRES);
		for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
			let g = each_array[$$index];
			$$renderer.push(`<button type="button"${attr_class("genre-card svelte-79z2zc", void 0, { "active": genre === g.key })}${attr("aria-pressed", genre === g.key)}>${escape_html(g.label)}</button>`);
		}
		$$renderer.push(`<!--]--></div> `);
		if (errors.genre) $$renderer.push(`<!--[0--><div class="err svelte-79z2zc">${escape_html(errors.genre)}</div>`);
		else $$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--></div></section> <section class="group svelte-79z2zc"><div class="field svelte-79z2zc" data-field="cover"><span class="label svelte-79z2zc" id="cover-label">Cover</span> <div class="segmented svelte-79z2zc" role="group" aria-labelledby="cover-label"><button type="button"${attr("aria-pressed", coverMode === "uploaded")}${attr_class("svelte-79z2zc", void 0, { "active": coverMode === "uploaded" })}>Upload gambar</button> <button type="button"${attr("aria-pressed", coverMode === "generated")}${attr_class("svelte-79z2zc", void 0, { "active": coverMode === "generated" })}>Buat otomatis</button></div> `);
		if (errors.cover) $$renderer.push(`<!--[0--><div class="err svelte-79z2zc">${escape_html(errors.cover)}</div>`);
		else $$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--></div> `);
		if (coverMode === "uploaded") {
			$$renderer.push(`<!--[0--><div class="cover-panel svelte-79z2zc"><div class="cover-inputs svelte-79z2zc" data-field="coverFile"><input type="file" accept=".png,.jpg,.jpeg,image/png,image/jpeg" aria-label="File cover" class="svelte-79z2zc"/> <div class="hint svelte-79z2zc">PNG atau JPG, maksimal 5 MB. `);
			if (mode === "edit" && existingCoverURL && true) $$renderer.push(`<!--[0-->Kosongkan untuk mempertahankan cover lama.`);
			else $$renderer.push("<!--[-1-->");
			$$renderer.push(`<!--]--></div> `);
			if (errors.coverFile) $$renderer.push(`<!--[0--><div class="err svelte-79z2zc">${escape_html(errors.coverFile)}</div>`);
			else $$renderer.push("<!--[-1-->");
			$$renderer.push(`<!--]--></div> `);
			if (coverPreviewURL || existingCoverURL) $$renderer.push(`<!--[0--><div class="preview svelte-79z2zc"><img${attr("src", coverPreviewURL || existingCoverURL)}${attr("alt", coverPreviewURL ? "Preview cover" : "Cover saat ini")} class="svelte-79z2zc"/></div>`);
			else $$renderer.push("<!--[-1-->");
			$$renderer.push(`<!--]--></div>`);
		} else $$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--> `);
		if (coverMode === "generated") {
			$$renderer.push(`<!--[0--><div class="cover-panel svelte-79z2zc"><div class="cover-inputs svelte-79z2zc"><div class="sub-field svelte-79z2zc" data-field="coverColor"><label class="mini-label svelte-79z2zc" for="coverColor">Warna latar</label> <div class="color-row svelte-79z2zc"><input id="coverColor" type="color"${attr("value", coverColor)} class="svelte-79z2zc"/> <input type="text" class="hex svelte-79z2zc"${attr("value", coverColor)} placeholder="#9c7450" aria-label="Kode warna hex"/></div> `);
			if (errors.coverColor) $$renderer.push(`<!--[0--><div class="err svelte-79z2zc">${escape_html(errors.coverColor)}</div>`);
			else $$renderer.push("<!--[-1-->");
			$$renderer.push(`<!--]--></div> <div class="sub-field svelte-79z2zc" data-field="coverText"><label class="mini-label svelte-79z2zc" for="coverText"><span>Teks cover</span> <span class="counter svelte-79z2zc">${escape_html([...coverText].length)}/18</span></label> <input id="coverText" type="text" placeholder="Judul singkat atau inisial"${attr("value", coverText)}${attr("maxlength", MAX_COVER_TEXT)} class="svelte-79z2zc"/> `);
			if (errors.coverText) $$renderer.push(`<!--[0--><div class="err svelte-79z2zc">${escape_html(errors.coverText)}</div>`);
			else $$renderer.push("<!--[-1-->");
			$$renderer.push(`<!--]--></div></div> <div class="preview svelte-79z2zc"><div class="gen-cover svelte-79z2zc"${attr_style(`background:${stringify(coverColor)}; color:${stringify(textColor)}; font-size:${stringify(genFontSize)}px;`)}>${escape_html(coverText || "Preview")}</div></div></div>`);
		} else $$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--></section> <section class="group svelte-79z2zc"><div class="field svelte-79z2zc" data-field="pdfFile"><label for="pdf" class="svelte-79z2zc">File PDF</label> <input id="pdf" type="file" accept=".pdf,application/pdf" class="svelte-79z2zc"/> <div class="hint svelte-79z2zc">Maksimal 50 MB. `);
		if (mode === "edit" && book?.pdf_path && true) $$renderer.push(`<!--[0-->Kosongkan untuk mempertahankan PDF lama.`);
		else $$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--></div> `);
		if (mode === "edit" && book?.pdf_path) $$renderer.push(`<!--[1--><div class="file-info svelte-79z2zc">${escape_html(book.pdf_path.split("/").pop())} (file saat ini)</div>`);
		else $$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--> `);
		if (errors.pdfFile) $$renderer.push(`<!--[0--><div class="err svelte-79z2zc">${escape_html(errors.pdfFile)}</div>`);
		else $$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--></div> <div class="field svelte-79z2zc" data-field="synopsis"><label for="synopsis" class="svelte-79z2zc">Sinopsis <span class="opt svelte-79z2zc">(opsional)</span></label> <textarea id="synopsis" rows="5" placeholder="Ringkasan singkat isi buku"${attr_class("svelte-79z2zc", void 0, { "invalid": errors.synopsis })}>`);
		const $$body = escape_html(synopsis);
		if ($$body) $$renderer.push(`${$$body}`);
		$$renderer.push(`</textarea> `);
		if (errors.synopsis) $$renderer.push(`<!--[0--><div class="err svelte-79z2zc">${escape_html(errors.synopsis)}</div>`);
		else $$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--></div></section> `);
		$$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--> `);
		$$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--> <div class="actions svelte-79z2zc"><a${attr("href", mode === "edit" ? `/read/${book.id}` : "/")} class="btn cancel svelte-79z2zc">Batal</a> <button type="submit" class="btn primary svelte-79z2zc"${attr("disabled", submitting, true)}>${escape_html(mode === "edit" ? "Simpan perubahan" : "Upload buku")}</button></div></form>`);
		bind_props($$props, {
			book,
			mode
		});
	});
}
//#endregion
export { BookForm as t };
