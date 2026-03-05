import React, { useState, useEffect, useCallback, useRef } from "react";
import { Loader2, ExternalLink, Compass, Tag } from "lucide-react";
import { useStore } from "@nanostores/react";
import { clsx } from "clsx";
import { getDocuments, getRecommendations } from "../../api/client";
import type { DocumentItem } from "../../api/client";
import { Tabs, EmptyState } from "../../components/ui";
import LayoutToggle from "../../components/ui/LayoutToggle";
import { $user } from "../../store/auth";
import { $feedLayout } from "../../store/feedLayout";
import { formatDistanceToNow } from "date-fns";

export default function Discover() {
  const user = useStore($user);
  const layout = useStore($feedLayout);
  const [activeTab, setActiveTab] = useState("new");
  const [items, setItems] = useState<DocumentItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [hasMore, setHasMore] = useState(false);
  const [offset, setOffset] = useState(0);
  const [recommendationsUnavailable, setRecommendationsUnavailable] =
    useState(false);
  const fetchIdRef = useRef(0);
  const limit = 30;

  const tabs = [
    { id: "new", label: "New" },
    { id: "popular", label: "Popular" },
    ...(user ? [{ id: "recommended", label: "For You" }] : []),
  ];

  const fetchItems = useCallback(
    async (tab: string, newOffset = 0, append = false) => {
      const id = ++fetchIdRef.current;
      setLoading(true);

      let data: { items: DocumentItem[]; totalItems: number };
      if (tab === "recommended") {
        const res = await getRecommendations(limit);
        if ("unavailable" in res && res.unavailable) {
          setRecommendationsUnavailable(true);
          setLoading(false);
          return;
        }
        setRecommendationsUnavailable(false);
        data = res;
      } else {
        data = await getDocuments({ sort: tab, limit, offset: newOffset });
      }

      if (id !== fetchIdRef.current) return;

      setItems((prev) => (append ? [...prev, ...data.items] : data.items));
      setHasMore(
        tab !== "recommended" &&
          newOffset + data.items.length < data.totalItems,
      );
      setOffset(newOffset + data.items.length);
      setLoading(false);
    },
    [limit],
  );

  useEffect(() => {
    queueMicrotask(() => fetchItems(activeTab, 0));
  }, [activeTab, fetchItems]);

  const handleTabChange = (id: string) => {
    if (id === activeTab) return;
    setActiveTab(id);
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  const loadMore = () => {
    fetchItems(activeTab, offset, true);
  };

  return (
    <div className="mx-auto max-w-2xl xl:max-w-none">
      <div className="sticky top-0 z-10 bg-white/95 dark:bg-surface-800/95 backdrop-blur-sm pb-3 mb-2 -mx-1 px-1 pt-1 space-y-2">
        <div className="flex items-center gap-2">
          <Tabs tabs={tabs} activeTab={activeTab} onChange={handleTabChange} />
          <LayoutToggle className="hidden sm:inline-flex ml-auto" />
        </div>
      </div>

      {loading && items.length === 0 ? (
        <div className="flex justify-center py-20">
          <Loader2 className="w-6 h-6 animate-spin text-surface-400" />
        </div>
      ) : activeTab === "recommended" && recommendationsUnavailable ? (
        <EmptyState
          icon={<Compass size={40} />}
          title="Coming soon"
          message="Personalized recommendations aren't available on this server yet."
        />
      ) : items.length === 0 ? (
        <EmptyState
          icon={<Compass size={40} />}
          title="Nothing here yet"
          message={
            activeTab === "recommended"
              ? "Start annotating and highlighting to get personalized recommendations."
              : "No documents have been discovered yet. Check back soon!"
          }
        />
      ) : (
        <div
          className={clsx(
            layout === "mosaic"
              ? "columns-1 sm:columns-2 xl:columns-3 2xl:columns-4 gap-4"
              : "space-y-3",
            "animate-fade-in",
          )}
        >
          {items.map((doc) => (
            <div
              key={doc.uri}
              className={
                layout === "mosaic" ? "break-inside-avoid mb-4" : undefined
              }
            >
              <DocumentCard doc={doc} />
            </div>
          ))}

          {loading && (
            <div className="flex justify-center py-6">
              <Loader2 className="w-5 h-5 animate-spin text-surface-400" />
            </div>
          )}

          {hasMore && !loading && (
            <button
              onClick={loadMore}
              className="w-full py-3 text-sm font-medium text-surface-500 hover:text-surface-700 dark:text-surface-400 dark:hover:text-surface-200 hover:bg-surface-100 dark:hover:bg-surface-800 rounded-lg transition-colors"
            >
              Load more
            </button>
          )}
        </div>
      )}
    </div>
  );
}

function DocumentCard({ doc }: { doc: DocumentItem }) {
  const [ogData, setOgData] = useState<{
    title?: string;
    description?: string;
    image?: string;
    icon?: string;
  } | null>(() => {
    try {
      const cached = sessionStorage.getItem(`og:${doc.canonicalUrl}`);
      return cached ? JSON.parse(cached) : null;
    } catch {
      return null;
    }
  });

  useEffect(() => {
    if (!doc.canonicalUrl || ogData) return;
    fetch(`/api/url-metadata?url=${encodeURIComponent(doc.canonicalUrl)}`)
      .then((res) => (res.ok ? res.json() : null))
      .then((data) => {
        if (data) {
          setOgData(data);
          try {
            sessionStorage.setItem(
              `og:${doc.canonicalUrl}`,
              JSON.stringify(data),
            );
          } catch {
            /* quota exceeded */
          }
        }
      })
      .catch(() => {});
  }, [doc.canonicalUrl, ogData]);

  const displayUrl = doc.canonicalUrl
    .replace(/^https?:\/\//, "")
    .replace(/\/$/, "");

  const hostname = (() => {
    try {
      return new URL(doc.canonicalUrl).hostname;
    } catch {
      return null;
    }
  })();

  return (
    <a
      href={doc.canonicalUrl}
      target="_blank"
      rel="noopener noreferrer"
      className="card block hover:ring-1 hover:ring-black/10 dark:hover:ring-white/10 transition-all group overflow-hidden"
    >
      {ogData?.image && (
        <div className="w-full h-40 bg-surface-100 dark:bg-surface-800 overflow-hidden">
          <img
            src={ogData.image}
            alt=""
            className="w-full h-full object-cover"
            onError={(e) => (e.currentTarget.style.display = "none")}
          />
        </div>
      )}
      <div className="p-4">
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0 flex-1">
            <h3 className="font-display font-semibold text-surface-900 dark:text-white group-hover:text-primary-600 dark:group-hover:text-primary-400 transition-colors line-clamp-2">
              {doc.title || displayUrl}
            </h3>
            {doc.description && (
              <p className="mt-1 text-sm text-surface-500 dark:text-surface-400 line-clamp-2">
                {doc.description}
              </p>
            )}
            <div className="mt-2 flex items-center gap-3 text-xs text-surface-400 dark:text-surface-500">
              <span className="flex items-center gap-1 truncate">
                {ogData?.icon ? (
                  <img
                    src={ogData.icon}
                    alt=""
                    className="w-3 h-3 rounded-sm"
                    onError={(e) => (e.currentTarget.style.display = "none")}
                  />
                ) : (
                  <ExternalLink size={12} />
                )}
                {hostname || displayUrl}
              </span>
              {doc.publishedAt && (
                <span>
                  {formatDistanceToNow(new Date(doc.publishedAt), {
                    addSuffix: true,
                  })}
                </span>
              )}
            </div>
            {doc.tags && doc.tags.length > 0 && (
              <div className="mt-2 flex flex-wrap gap-1.5">
                {doc.tags.slice(0, 5).map((tag) => (
                  <span
                    key={tag}
                    className="inline-flex items-center gap-1 px-2 py-1 text-xs font-medium rounded-md bg-surface-100 dark:bg-surface-800 text-surface-600 dark:text-surface-400"
                  >
                    <Tag size={10} />
                    {tag}
                  </span>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </a>
  );
}
