type QueueEntry = {
  url: string;
  resolve: (data: Record<string, string> | null) => void;
};

const MAX_CONCURRENT = 3;
const queue: QueueEntry[] = [];
let active = 0;
const inflight = new Map<string, Promise<Record<string, string> | null>>();

function processQueue() {
  while (active < MAX_CONCURRENT && queue.length > 0) {
    const entry = queue.shift()!;
    active++;

    doFetch(entry.url)
      .then(entry.resolve)
      .finally(() => {
        active--;
        processQueue();
      });
  }
}

function doFetch(url: string): Promise<Record<string, string> | null> {
  const existing = inflight.get(url);
  if (existing) return existing;

  const promise = fetch(`/api/url-metadata?url=${encodeURIComponent(url)}`)
    .then((res) => (res.ok ? res.json() : null))
    .catch(() => null)
    .finally(() => {
      inflight.delete(url);
    });

  inflight.set(url, promise);
  return promise;
}

export function fetchMetadata(
  url: string,
): Promise<Record<string, string> | null> {
  try {
    const cached = sessionStorage.getItem(`og:${url}`);
    if (cached) return Promise.resolve(JSON.parse(cached));
  } catch {
    /* ignore */
  }

  return new Promise((resolve) => {
    queue.push({ url, resolve });
    processQueue();
  }).then((data) => {
    if (data) {
      try {
        sessionStorage.setItem(`og:${url}`, JSON.stringify(data));
      } catch {
        /* ignore */
      }
    }
    return data as Record<string, string> | null;
  });
}
