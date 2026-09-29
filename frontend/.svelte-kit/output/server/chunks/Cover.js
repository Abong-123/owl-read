import { C as attr, kt as fallback, n as attr_style, r as bind_props, u as stringify, w as escape_html } from "./server.js";
//#region src/lib/Cover.svelte
function Cover($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		let height, isUploaded, text, fontSize, bg, textColor;
		let book = $$props["book"];
		let size = fallback($$props["size"], 118);
		function calcFontSize(len, w) {
			if (len <= 4) return Math.round(w * .2);
			if (len <= 8) return Math.round(w * .16);
			if (len <= 12) return Math.round(w * .13);
			if (len <= 15) return Math.round(w * .11);
			return Math.round(w * .095);
		}
		function isLight(hex) {
			const c = (hex || "#9c7450").replace("#", "");
			const full = c.length === 3 ? c.split("").map((x) => x + x).join("") : c;
			const r = parseInt(full.slice(0, 2), 16);
			const g = parseInt(full.slice(2, 4), 16);
			const b = parseInt(full.slice(4, 6), 16);
			return .299 * r + .587 * g + .114 * b > 160;
		}
		$: height = Math.round(size * 1.5);
		$: isUploaded = book?.cover_type === "uploaded" && book?.cover_path;
		$: text = book?.cover_text || "";
		$: fontSize = calcFontSize([...text].length, size);
		$: bg = book?.cover_color || "#9c7450";
		$: textColor = isLight(bg) ? "#463d36" : "#fffdf9";
		$$renderer.push(`<div class="cover svelte-1470o3e"${attr_style(`width:${stringify(size)}px; height:${stringify(height)}px;`)}>`);
		if (isUploaded) $$renderer.push(`<!--[0--><img${attr("src", `/${stringify(book.cover_path)}`)}${attr("alt", book.title)} loading="lazy" class="svelte-1470o3e"/>`);
		else $$renderer.push(`<!--[-1--><div class="generated svelte-1470o3e"${attr_style(`background:${stringify(bg)}; color:${stringify(textColor)}; font-size:${stringify(fontSize)}px;`)}>${escape_html(text)}</div>`);
		$$renderer.push(`<!--]--></div>`);
		bind_props($$props, {
			book,
			size
		});
	});
}
//#endregion
export { Cover as t };
