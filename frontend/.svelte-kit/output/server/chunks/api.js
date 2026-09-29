//#region src/lib/api.js
var UPLOAD_CODE = "1234";
function readCookie(name) {
	if (typeof document === "undefined") return "";
	const match = document.cookie.match(new RegExp("(^|;\\s*)" + name + "=([^;]*)"));
	return match ? decodeURIComponent(match[2]) : "";
}
async function ensureCSRF() {
	if (readCookie("csrf_token")) return;
	await fetch("/api/csrf", { credentials: "include" });
}
async function apiGet(path) {
	const res = await fetch(path, { credentials: "include" });
	if (!res.ok) throw new Error(await safeError(res) || "Gagal memuat data");
	return res.json();
}
async function apiWrite(method, path, body) {
	await ensureCSRF();
	const opts = {
		method,
		credentials: "include",
		headers: {
			"X-CSRF-Token": readCookie("csrf_token"),
			"X-Upload-Code": UPLOAD_CODE
		}
	};
	if (body instanceof FormData) opts.body = body;
	else if (body !== void 0) {
		opts.headers["Content-Type"] = "application/json";
		opts.body = JSON.stringify(body);
	}
	const res = await fetch(path, opts);
	if (!res.ok) throw new Error(await safeError(res) || "Operasi gagal");
	return res.json();
}
async function safeError(res) {
	try {
		return (await res.json()).error;
	} catch {
		return null;
	}
}
var api = {
	listBooks: () => apiGet("/api/books"),
	getBook: (id) => apiGet("/api/books/" + id),
	stats: () => apiGet("/api/stats"),
	createBook: (form) => apiWrite("POST", "/api/books", form),
	updateBook: (id, f) => apiWrite("PUT", "/api/books/" + id, f),
	deleteBook: (id) => apiWrite("DELETE", "/api/books/" + id)
};
//#endregion
export { api as t };
