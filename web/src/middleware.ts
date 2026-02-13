import type { APIContext } from "astro";

const API_PORT = process.env.API_PORT || 8081;
const API_URL = process.env.API_URL || `http://localhost:${API_PORT}`;

const PROXY_PATHS = ["/api/", "/auth/", "/client-metadata.json", "/jwks.json"];

export async function onRequest(
  { request, url }: APIContext,
  next: () => Promise<Response>,
): Promise<Response> {
  const shouldProxy = PROXY_PATHS.some(
    (p) => url.pathname.startsWith(p) || url.pathname === p.replace(/\/$/, ""),
  );

  if (!shouldProxy) {
    return next();
  }

  const target = new URL(url.pathname + url.search, API_URL);

  const headers = new Headers(request.headers);
  headers.delete("host");

  const init: RequestInit = {
    method: request.method,
    headers,
  };

  if (request.method !== "GET" && request.method !== "HEAD" && request.body) {
    init.body = request.body;
    // @ts-expect-error
    init.duplex = "half";
  }

  try {
    const res = await fetch(target.toString(), init);
    return new Response(res.body, {
      status: res.status,
      statusText: res.statusText,
      headers: res.headers,
    });
  } catch {
    return new Response("Backend unavailable", { status: 502 });
  }
}
