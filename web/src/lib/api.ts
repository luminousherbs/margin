import type {
  AnnotationItem,
  UserProfile,
  FeedResponse,
  Collection,
  NotificationItem,
  Target,
  Selector,
} from "../types";

const API_URL =
  process.env.API_URL || `http://localhost:${process.env.API_PORT || 8081}`;

interface RawItem {
  type?: string;
  collectionUri?: string;
  annotation?: RawItem;
  highlight?: RawItem;
  bookmark?: RawItem;
  uri?: string;
  id?: string;
  cid?: string;
  author?: UserProfile;
  creator?: UserProfile;
  collection?: { uri: string; name: string; icon?: string };
  context?: { uri: string; name: string; icon?: string }[];
  created?: string;
  createdAt?: string;
  target?: string | { source?: string; title?: string; selector?: Selector };
  url?: string;
  targetUrl?: string;
  title?: string;
  selector?: Selector;
  viewer?: { like?: string; [key: string]: unknown };
  viewerHasLiked?: boolean;
  motivation?: string;
  [key: string]: unknown;
}

export function normalizeItem(raw: RawItem): AnnotationItem {
  if (raw.type === "CollectionItem" || raw.collectionUri) {
    const inner = raw.annotation || raw.highlight || raw.bookmark || {};
    const normalizedInner = normalizeItem(inner);
    return {
      ...normalizedInner,
      uri: normalizedInner.uri || raw.uri || "",
      cid: normalizedInner.cid || raw.cid || "",
      author: (normalizedInner.author ||
        raw.author ||
        raw.creator) as UserProfile,
      collection: raw.collection
        ? {
            uri: raw.collection.uri,
            name: raw.collection.name,
            icon: raw.collection.icon,
          }
        : undefined,
      context: raw.context?.map((c) => ({
        uri: c.uri,
        name: c.name,
        icon: c.icon,
      })),
      addedBy: raw.creator || raw.author,
      createdAt:
        normalizedInner.createdAt ||
        raw.created ||
        raw.createdAt ||
        new Date().toISOString(),
      collectionItemUri: raw.id || raw.uri,
    };
  }

  let target: Target | undefined;
  if (raw.target) {
    if (typeof raw.target === "string") {
      target = { source: raw.target, title: raw.title, selector: raw.selector };
    } else {
      target = {
        source: raw.target.source || "",
        title: raw.target.title || raw.title,
        selector: raw.target.selector || raw.selector,
      };
    }
  }
  if (!target || !target.source) {
    const url =
      raw.url ||
      raw.targetUrl ||
      (typeof raw.target === "string" ? raw.target : raw.target?.source);
    if (url) {
      target = {
        source: url,
        title:
          raw.title ||
          (typeof raw.target !== "string" ? raw.target?.title : undefined),
        selector:
          raw.selector ||
          (typeof raw.target !== "string" ? raw.target?.selector : undefined),
      };
    }
  }

  return {
    ...raw,
    uri: raw.id || raw.uri || "",
    cid: raw.cid || "",
    author: (raw.creator || raw.author) as UserProfile,
    createdAt: raw.created || raw.createdAt || new Date().toISOString(),
    target,
    viewer: raw.viewer || { like: raw.viewerHasLiked ? "true" : undefined },
    motivation: raw.motivation || "highlighting",
    parentUri: (raw as Record<string, unknown>).inReplyTo as string | undefined,
  };
}

async function serverFetch(path: string, cookie?: string): Promise<Response> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };
  if (cookie) headers["Cookie"] = cookie;
  return fetch(`${API_URL}${path}`, { headers });
}

const sessionCache = new Map<string, { user: UserProfile; expires: number }>();

export async function getSession(cookie: string): Promise<UserProfile | null> {
  try {
    const cacheKey = cookie.match(/margin_session=([^;]+)/)?.[1] || "";
    const cached = sessionCache.get(cacheKey);
    if (cached && Date.now() < cached.expires) {
      return cached.user;
    }

    const res = await serverFetch("/auth/session", cookie);
    if (!res.ok) return null;
    const data = await res.json();
    if (!data.authenticated && !data.did) return null;

    const profile: UserProfile = {
      did: data.did,
      handle: data.handle,
      displayName: data.displayName,
      avatar: data.avatar,
      description: data.description,
      website: data.website,
      links: data.links,
      followersCount: data.followersCount,
      followsCount: data.followsCount,
      postsCount: data.postsCount,
    };

    // Fetch bsky profile and margin profile in parallel with a 3s timeout
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), 3000);

    const [bskyRes, marginRes] = await Promise.allSettled([
      fetch(
        `https://public.api.bsky.app/xrpc/app.bsky.actor.getProfile?actor=${encodeURIComponent(data.did)}`,
        { signal: controller.signal },
      ),
      serverFetch(`/api/profile/${data.did}`, cookie),
    ]);

    clearTimeout(timeout);

    if (bskyRes.status === "fulfilled" && bskyRes.value.ok) {
      try {
        const bsky = await bskyRes.value.json();
        if (bsky.avatar) profile.avatar = bsky.avatar;
        if (bsky.displayName) profile.displayName = bsky.displayName;
      } catch {
        /* ignore */
      }
    }

    if (marginRes.status === "fulfilled" && marginRes.value.ok) {
      try {
        const mp = await marginRes.value.json();
        if (mp?.description) profile.description = mp.description;
        if (mp?.followersCount) profile.followersCount = mp.followersCount;
        if (mp?.followsCount) profile.followsCount = mp.followsCount;
        if (mp?.postsCount) profile.postsCount = mp.postsCount;
        if (mp?.website) profile.website = mp.website;
        if (mp?.links) profile.links = mp.links;
      } catch {
        /* ignore */
      }
    }

    if (cacheKey) {
      sessionCache.set(cacheKey, {
        user: profile,
        expires: Date.now() + 30_000,
      });
      // Evict old entries
      if (sessionCache.size > 100) {
        const now = Date.now();
        for (const [k, v] of sessionCache) {
          if (now > v.expires) sessionCache.delete(k);
        }
      }
    }

    return profile;
  } catch {
    return null;
  }
}

export interface GetFeedParams {
  type?: string;
  limit?: number;
  offset?: number;
  motivation?: string;
  tag?: string;
  creator?: string;
  source?: string;
}

function groupFeedItems(items: AnnotationItem[]): AnnotationItem[] {
  if (items.length === 0) return items;
  const grouped: AnnotationItem[] = [items[0]];
  for (let i = 1; i < items.length; i++) {
    const prev = grouped[grouped.length - 1];
    const curr = items[i];
    if (
      prev.collection &&
      curr.collection &&
      prev.uri === curr.uri &&
      prev.addedBy?.did === curr.addedBy?.did
    ) {
      if (!prev.context) prev.context = [prev.collection];
      if (!prev.context.find((c) => c.uri === curr.collection!.uri)) {
        prev.context.push(curr.collection);
      }
      continue;
    }
    grouped.push(curr);
  }
  return grouped;
}

export async function getFeed(
  cookie: string,
  params: GetFeedParams = {},
): Promise<FeedResponse> {
  const qs = new URLSearchParams();
  if (params.source) qs.append("source", params.source);
  if (params.type) qs.append("type", params.type);
  if (params.limit) qs.append("limit", params.limit.toString());
  if (params.offset) qs.append("offset", params.offset.toString());
  if (params.motivation) qs.append("motivation", params.motivation);
  if (params.tag) qs.append("tag", params.tag);
  if (params.creator) qs.append("creator", params.creator);

  const endpoint = params.source ? "/api/targets" : "/api/annotations/feed";
  try {
    const res = await serverFetch(`${endpoint}?${qs.toString()}`, cookie);
    if (!res.ok) return { items: [], hasMore: false, fetchedCount: 0 };
    const data = await res.json();
    const items = (data.items || []).map(normalizeItem);
    return {
      items: groupFeedItems(items),
      hasMore: items.length >= (params.limit || 50),
      fetchedCount: items.length,
    };
  } catch {
    return { items: [], hasMore: false, fetchedCount: 0 };
  }
}

export async function searchItems(
  cookie: string,
  query: string,
  options: { creator?: string; limit?: number; offset?: number } = {},
): Promise<FeedResponse> {
  const qs = new URLSearchParams();
  qs.append("q", query);
  if (options.creator) qs.append("creator", options.creator);
  if (options.limit) qs.append("limit", options.limit.toString());
  if (options.offset) qs.append("offset", options.offset.toString());

  try {
    const res = await serverFetch(`/api/search?${qs.toString()}`, cookie);
    if (!res.ok) return { items: [], hasMore: false, fetchedCount: 0 };
    const data = await res.json();
    const items = (data.items || []).map(normalizeItem);
    return {
      items,
      hasMore: items.length >= (options.limit || 50),
      fetchedCount: items.length,
    };
  } catch {
    return { items: [], hasMore: false, fetchedCount: 0 };
  }
}

export async function getAnnotation(
  cookie: string,
  uri: string,
): Promise<AnnotationItem | null> {
  try {
    const res = await serverFetch(
      `/api/annotation?uri=${encodeURIComponent(uri)}`,
      cookie,
    );
    if (!res.ok) return null;
    const data = await res.json();
    return normalizeItem(data);
  } catch {
    return null;
  }
}

export async function getReplies(
  cookie: string,
  uri: string,
): Promise<AnnotationItem[]> {
  try {
    const res = await serverFetch(
      `/api/replies?uri=${encodeURIComponent(uri)}`,
      cookie,
    );
    if (!res.ok) return [];
    const data = await res.json();
    return (data.items || []).map(normalizeItem);
  } catch {
    return [];
  }
}

export async function getProfile(
  cookie: string,
  did: string,
): Promise<UserProfile | null> {
  try {
    const res = await serverFetch(
      `/api/profile/${encodeURIComponent(did)}`,
      cookie,
    );
    if (!res.ok) return null;
    return await res.json();
  } catch {
    return null;
  }
}

export async function getCollections(
  cookie: string,
  author?: string,
): Promise<Collection[]> {
  const qs = author ? `?author=${encodeURIComponent(author)}` : "";
  try {
    const res = await serverFetch(`/api/collections${qs}`, cookie);
    if (!res.ok) return [];
    const data = await res.json();
    return data.collections || data || [];
  } catch {
    return [];
  }
}

export async function getCollection(
  cookie: string,
  uri: string,
): Promise<Collection | null> {
  try {
    const res = await serverFetch(
      `/api/collection?uri=${encodeURIComponent(uri)}`,
      cookie,
    );
    if (!res.ok) return null;
    return await res.json();
  } catch {
    return null;
  }
}

export async function getCollectionItems(
  cookie: string,
  uri: string,
): Promise<AnnotationItem[]> {
  try {
    const res = await serverFetch(
      `/api/collections/${encodeURIComponent(uri)}/items`,
      cookie,
    );
    if (!res.ok) return [];
    const data = await res.json();
    return (data.items || data || []).map(normalizeItem);
  } catch {
    return [];
  }
}

export async function getNotifications(
  cookie: string,
  limit = 50,
  offset = 0,
): Promise<{ items: NotificationItem[]; hasMore: boolean }> {
  try {
    const res = await serverFetch(
      `/api/notifications?limit=${limit}&offset=${offset}`,
      cookie,
    );
    if (!res.ok) return { items: [], hasMore: false };
    const data = await res.json();
    return {
      items: data.notifications || [],
      hasMore: (data.notifications || []).length >= limit,
    };
  } catch {
    return { items: [], hasMore: false };
  }
}

export async function getTrendingTags(
  limit = 10,
): Promise<{ tag: string; count: number }[]> {
  try {
    const res = await serverFetch(`/api/tags/trending?limit=${limit}`);
    if (!res.ok) return [];
    return await res.json();
  } catch {
    return [];
  }
}

export async function getRecommendations(
  cookie: string,
  params: { sort?: string; limit?: number; offset?: number } = {},
): Promise<{ items: AnnotationItem[]; hasMore: boolean }> {
  const qs = new URLSearchParams();
  if (params.sort) qs.append("sort", params.sort);
  if (params.limit) qs.append("limit", params.limit.toString());
  if (params.offset) qs.append("offset", params.offset.toString());

  try {
    const res = await serverFetch(`/api/documents?${qs.toString()}`, cookie);
    if (!res.ok) return { items: [], hasMore: false };
    const data = await res.json();
    return {
      items: data.items || [],
      hasMore: (data.items || []).length >= (params.limit || 20),
    };
  } catch {
    return { items: [], hasMore: false };
  }
}

export async function getByTarget(
  cookie: string,
  url: string,
  limit = 50,
  offset = 0,
): Promise<{ annotations: AnnotationItem[]; highlights: AnnotationItem[] }> {
  try {
    const res = await serverFetch(
      `/api/targets?source=${encodeURIComponent(url)}&limit=${limit}&offset=${offset}`,
      cookie,
    );
    if (!res.ok) return { annotations: [], highlights: [] };
    const data = await res.json();
    const items = (data.items || []).map(normalizeItem);
    return {
      annotations: items.filter(
        (i: AnnotationItem) => i.motivation === "commenting",
      ),
      highlights: items.filter(
        (i: AnnotationItem) => i.motivation === "highlighting",
      ),
    };
  } catch {
    return { annotations: [], highlights: [] };
  }
}

export async function resolveHandle(handle: string): Promise<string | null> {
  if (handle.startsWith("did:")) return handle;
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
