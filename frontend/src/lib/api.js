// Wrapper fetch ke backend.
// - Otomatis sertakan credentials (cookie CSRF).
// - Ambil token CSRF dari cookie & kirim sebagai header saat write.
// - Sertakan X-Upload-Code saat write.

const UPLOAD_CODE = '1234'; // samakan dengan middleware.go

// Baca cookie di browser.
function readCookie(name) {
  if (typeof document === 'undefined') return '';
  const match = document.cookie.match(new RegExp('(^|;\\s*)' + name + '=([^;]*)'));
  return match ? decodeURIComponent(match[2]) : '';
}

// Pastikan cookie CSRF sudah ada. Panggil sekali sebelum write.
export async function ensureCSRF() {
  if (readCookie('csrf_token')) return;
  await fetch('/api/csrf', { credentials: 'include' });
}

// GET sederhana.
export async function apiGet(path) {
  const res = await fetch(path, { credentials: 'include' });
  if (!res.ok) throw new Error((await safeError(res)) || 'Gagal memuat data');
  return res.json();
}

// POST/PUT/DELETE dengan CSRF + upload code.
export async function apiWrite(method, path, body) {
  await ensureCSRF();
  const token = readCookie('csrf_token');

  const opts = {
    method,
    credentials: 'include',
    headers: {
      'X-CSRF-Token': token,
      'X-Upload-Code': UPLOAD_CODE
    }
  };

  // Kalau body FormData, biarkan browser set Content-Type.
  // Kalau JSON, set manual.
  if (body instanceof FormData) {
    opts.body = body;
  } else if (body !== undefined) {
    opts.headers['Content-Type'] = 'application/json';
    opts.body = JSON.stringify(body);
  }

  const res = await fetch(path, opts);
  if (!res.ok) throw new Error((await safeError(res)) || 'Operasi gagal');
  return res.json();
}

async function safeError(res) {
  try {
    const j = await res.json();
    return j.error;
  } catch {
    return null;
  }
}

// Shortcut
export const api = {
  listBooks:   ()      => apiGet('/api/books'),
  getBook:     (id)    => apiGet('/api/books/' + id),
  stats:       ()      => apiGet('/api/stats'),
  createBook:  (form)  => apiWrite('POST', '/api/books', form),
  updateBook:  (id, f) => apiWrite('PUT', '/api/books/' + id, f),
  deleteBook:  (id)    => apiWrite('DELETE', '/api/books/' + id)
};