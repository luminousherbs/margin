const API_URL = process.env.API_URL || "http://localhost:8081";

const CRAWLER_AGENTS = [
  "facebookexternalhit",
  "facebot",
  "twitterbot",
  "linkedinbot",
  "whatsapp",
  "slackbot",
  "telegrambot",
  "discordbot",
  "applebot",
  "bot",
  "crawler",
  "spider",
  "preview",
  "cardyb",
  "bluesky",
];

export function isCrawler(userAgent: string): boolean {
  const ua = userAgent.toLowerCase();
  return CRAWLER_AGENTS.some((bot) => ua.includes(bot));
}

export interface OGData {
  title: string;
  description: string;
  image: string;
  author: string;
  pageURL: string;
}

interface APIAnnotation {
  id?: string;
  uri?: string;
  author?: { did: string; handle?: string };
  creator?: { did: string; handle?: string };
  target?: { source?: string; title?: string; selector?: { exact?: string } };
  body?: string;
  bodyValue?: string;
  text?: string;
  motivation?: string;
  title?: string;
  description?: string;
  url?: string;
  source?: string;
  selector?: { exact?: string };
  selectorJson?: string;
  color?: string;
}

interface APICollection {
  id?: string;
  uri?: string;
  name: string;
  description?: string;
  icon?: string;
  author?: { did: string; handle?: string };
  creator?: { did: string; handle?: string };
}

export async function resolveHandle(handle: string): Promise<string | null> {
  try {
    const res = await fetch(
      `https://public.api.bsky.app/xrpc/com.atproto.identity.resolveHandle?handle=${encodeURIComponent(handle)}`,
    );
    if (!res.ok) return null;
    const data = await res.json();
    return data.did || null;
  } catch {
    return null;
  }
}

async function fetchJSON(path: string): Promise<unknown> {
  const res = await fetch(`${API_URL}${path}`);
  if (!res.ok) return null;
  return res.json();
}

function getAuthorHandle(item: APIAnnotation | APICollection): string {
  const author = item.author || item.creator;
  if (author?.handle) return `@${author.handle}`;
  if (author?.did) return author.did;
  return "someone";
}

function extractDomain(urlStr: string): string {
  try {
    return new URL(urlStr).host;
  } catch {
    return "";
  }
}

function truncate(str: string, max: number): string {
  if (str.length <= max) return str;
  return str.slice(0, max - 3) + "...";
}

const BASE_URL = process.env.BASE_URL || "https://margin.at";

export async function fetchAnnotationOG(uri: string): Promise<OGData | null> {
  const item = (await fetchJSON(
    `/api/annotation?uri=${encodeURIComponent(uri)}`,
  )) as APIAnnotation | null;
  if (!item) return null;

  const itemURI = item.id || item.uri || uri;
  const author = getAuthorHandle(item);
  const source = item.target?.source || item.url || item.source || "";
  const domain = extractDomain(source);
  const selectorText =
    item.target?.selector?.exact || item.selector?.exact || "";

  let title = "Annotation on Margin";
  const targetTitle = item.target?.title || item.title;
  if (targetTitle) title = truncate(`Comment on: ${targetTitle}`, 60);

  let description = item.body || item.bodyValue || item.text || "";
  if (selectorText && description) {
    description = `"${truncate(selectorText, 100)}"\n\n${description}`;
  } else if (selectorText) {
    description = `Highlighted: "${truncate(selectorText, 150)}"`;
  }
  if (!description) {
    description = `An annotation by ${author}`;
    if (domain) description += ` on ${domain}`;
  }
  description = truncate(description, 200);

  return {
    title,
    description,
    image: `${BASE_URL}/og-image?uri=${encodeURIComponent(itemURI)}`,
    author,
    pageURL: `${BASE_URL}/at/${encodeURIComponent(itemURI.slice(5))}`,
  };
}

export async function fetchHighlightOG(uri: string): Promise<OGData | null> {
  const item = (await fetchJSON(
    `/api/annotation?uri=${encodeURIComponent(uri)}`,
  )) as APIAnnotation | null;
  if (!item) return null;

  const itemURI = item.id || item.uri || uri;
  const author = getAuthorHandle(item);
  const source = item.target?.source || item.url || item.source || "";
  const domain = extractDomain(source);
  const selectorText =
    item.target?.selector?.exact || item.selector?.exact || "";

  let title = "Highlight on Margin";
  const targetTitle = item.target?.title || item.title;
  if (targetTitle) title = truncate(`Highlight on: ${targetTitle}`, 60);

  let description = "";
  if (selectorText) {
    description = `"${truncate(selectorText, 180)}"`;
  }
  if (!description) {
    description = `A highlight by ${author}`;
    if (domain) description += ` on ${domain}`;
  }

  return {
    title,
    description,
    image: `${BASE_URL}/og-image?uri=${encodeURIComponent(itemURI)}`,
    author,
    pageURL: `${BASE_URL}/at/${encodeURIComponent(itemURI.slice(5))}`,
  };
}

export async function fetchBookmarkOG(uri: string): Promise<OGData | null> {
  const item = (await fetchJSON(
    `/api/annotation?uri=${encodeURIComponent(uri)}`,
  )) as APIAnnotation | null;
  if (!item) return null;

  const itemURI = item.id || item.uri || uri;
  const author = getAuthorHandle(item);
  const source = item.target?.source || item.url || item.source || "";
  const domain = extractDomain(source);

  const title = item.title || item.target?.title || "Bookmark on Margin";
  let description = item.description || item.body || item.bodyValue || "";
  if (!description) description = "A saved bookmark on Margin";
  if (domain) description += ` from ${domain}`;
  description = truncate(description, 200);

  return {
    title,
    description,
    image: `${BASE_URL}/og-image?uri=${encodeURIComponent(itemURI)}`,
    author,
    pageURL: `${BASE_URL}/at/${encodeURIComponent(itemURI.slice(5))}`,
  };
}

export async function fetchCollectionOG(uri: string): Promise<OGData | null> {
  const item = (await fetchJSON(
    `/api/collection?uri=${encodeURIComponent(uri)}`,
  )) as APICollection | null;
  if (!item) return null;

  const itemURI = item.id || item.uri || uri;
  const author = getAuthorHandle(item);
  const icon = item.icon || "📁";
  const title = `${icon} ${item.name}`;

  let description = "";
  if (item.description) {
    description = `By ${author} · ${truncate(item.description, 170)}`;
  } else {
    description = `A collection by ${author}`;
  }

  return {
    title,
    description,
    image: `${BASE_URL}/og-image?uri=${encodeURIComponent(itemURI)}`,
    author,
    pageURL: `${BASE_URL}/collection/${encodeURIComponent(itemURI)}`,
  };
}

export async function fetchOGByURI(uri: string): Promise<OGData | null> {
  if (uri.includes("/at.margin.annotation/")) return fetchAnnotationOG(uri);
  if (uri.includes("/at.margin.highlight/")) return fetchHighlightOG(uri);
  if (uri.includes("/at.margin.bookmark/")) return fetchBookmarkOG(uri);
  if (uri.includes("/at.margin.collection/")) return fetchCollectionOG(uri);

  return fetchAnnotationOG(uri);
}

export async function fetchOGForRoute(
  did: string,
  rkey: string,
  collectionType?: string,
): Promise<OGData | null> {
  if (collectionType) {
    const uri = `at://${did}/${collectionType}/${rkey}`;
    return fetchOGByURI(uri);
  }

  for (const type of [
    "at.margin.annotation",
    "at.margin.highlight",
    "at.margin.bookmark",
  ]) {
    const uri = `at://${did}/${type}/${rkey}`;
    const data = await fetchOGByURI(uri);
    if (data) return data;
  }

  const colUri = `at://${did}/at.margin.collection/${rkey}`;
  return fetchCollectionOG(colUri);
}
