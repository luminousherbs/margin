import React, { useState, useEffect, useCallback, useRef } from "react";
import {
  Search as SearchIcon,
  Loader2,
  SlidersHorizontal,
  MessageSquareText,
  Highlighter,
  Bookmark,
} from "lucide-react";
import { clsx } from "clsx";
import { useStore } from "@nanostores/react";
import { searchItems } from "../../api/client";
import type { AnnotationItem } from "../../types";
import Card from "../../components/common/Card";
import { EmptyState } from "../../components/ui";
import LayoutToggle from "../../components/ui/LayoutToggle";
import { $user } from "../../store/auth";
import { $feedLayout } from "../../store/feedLayout";

const searchCache = new Map<
  string,
  {
    results: AnnotationItem[];
    hasMore: boolean;
    offset: number;
    timestamp: number;
  }
>();

interface SearchProps {
  initialQuery?: string;
  initialResults?: AnnotationItem[];
  initialHasMore?: boolean;
}

export default function Search({
  initialQuery = "",
  initialResults,
  initialHasMore,
}: SearchProps) {
  const user = useStore($user);
  const layout = useStore($feedLayout);

  const [query, setQuery] = useState(initialQuery);
  const [results, setResults] = useState<AnnotationItem[]>(
    initialResults || [],
  );
  const [loading, setLoading] = useState(false);
  const [hasMore, setHasMore] = useState(initialHasMore ?? false);
  const [offset, setOffset] = useState(initialResults?.length ?? 0);
  const [myItemsOnly, setMyItemsOnly] = useState(false);
  const [activeFilter, setActiveFilter] = useState<string | undefined>(
    undefined,
  );
  const [platform, setPlatform] = useState<"all" | "margin" | "semble">("all");
  const inputRef = useRef<HTMLInputElement>(null);
  const myItemsRef = useRef(myItemsOnly);
  const fetchIdRef = useRef(0);

  useEffect(() => {
    myItemsRef.current = myItemsOnly;
  }, [myItemsOnly]);

  const filters = [
    { id: "all", label: "All", icon: null },
    { id: "commenting", label: "Annotations", icon: MessageSquareText },
    { id: "highlighting", label: "Highlights", icon: Highlighter },
    { id: "bookmarking", label: "Bookmarks", icon: Bookmark },
  ];

  const doSearch = useCallback(
    async (q: string, newOffset = 0, append = false) => {
      if (!q.trim()) {
        setResults([]);
        return;
      }

      const cacheKey = JSON.stringify({
        q: q.trim(),
        myItemsOnly: myItemsRef.current,
      });

      if (!append && newOffset === 0) {
        const cached = searchCache.get(cacheKey);
        if (cached && Date.now() - cached.timestamp < 5 * 60 * 1000) {
          setResults(cached.results);
          setHasMore(cached.hasMore);
          setOffset(cached.offset);
          setLoading(false);

          const id = ++fetchIdRef.current;
          searchItems(q.trim(), {
            creator: myItemsRef.current && user ? user.did : undefined,
            limit: 30,
            offset: newOffset,
          })
            .then((data) => {
              if (id !== fetchIdRef.current) return;
              setResults(data.items);
              setHasMore(data.hasMore);
              setOffset(newOffset + data.items.length);
              searchCache.set(cacheKey, {
                results: data.items,
                hasMore: data.hasMore,
                offset: newOffset + data.items.length,
                timestamp: Date.now(),
              });
            })
            .catch(console.error);

          return;
        }
      }

      const id = ++fetchIdRef.current;
      setLoading(true);
      const data = await searchItems(q.trim(), {
        creator: myItemsRef.current && user ? user.did : undefined,
        limit: 30,
        offset: newOffset,
      });
      if (id !== fetchIdRef.current) return;
      if (append) {
        setResults((prev) => {
          const newResults = [...prev, ...data.items];
          searchCache.set(cacheKey, {
            results: newResults,
            hasMore: data.hasMore,
            offset: newOffset + data.items.length,
            timestamp: Date.now(),
          });
          return newResults;
        });
      } else {
        setResults(data.items);
        searchCache.set(cacheKey, {
          results: data.items,
          hasMore: data.hasMore,
          offset: newOffset + data.items.length,
          timestamp: Date.now(),
        });
      }
      setHasMore(data.hasMore);
      setOffset(newOffset + data.items.length);
      setLoading(false);
    },
    [user],
  );

  const skipInitialSearch = useRef(!!initialResults);
  useEffect(() => {
    if (skipInitialSearch.current) {
      skipInitialSearch.current = false;
      return;
    }
    if (initialQuery) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      doSearch(initialQuery);
    }
  }, [initialQuery, doSearch]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (query.trim()) {
      const url = new URL(window.location.href);
      url.searchParams.set("q", query.trim());
      window.history.replaceState({}, "", url.toString());
      doSearch(query.trim());
    }
  };

  const handleDelete = (uri: string) => {
    setResults((prev) => prev.filter((item) => item.uri !== uri));
  };

  const handleFilterChange = (id: string) => {
    setActiveFilter(id === "all" ? undefined : id);
  };

  const filteredResults = results.filter((item) => {
    if (activeFilter && item.motivation !== activeFilter) return false;
    if (platform === "margin" && item.uri?.includes("network.cosmik"))
      return false;
    if (platform === "semble" && !item.uri?.includes("network.cosmik"))
      return false;
    return true;
  });

  return (
    <div className="mx-auto max-w-2xl xl:max-w-none">
      <form onSubmit={handleSubmit} className="mb-4">
        <div className="relative">
          <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
            <SearchIcon
              className="text-surface-400 dark:text-surface-500"
              size={18}
            />
          </div>
          <input
            ref={inputRef}
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search annotations, highlights, bookmarks..."
            autoFocus
            className="w-full pl-11 pr-4 py-3 bg-surface-50 dark:bg-surface-800 border border-surface-200 dark:border-surface-700 rounded-xl text-sm focus:outline-none focus:border-primary-400 focus:ring-2 focus:ring-primary-400/20 placeholder:text-surface-400"
          />
        </div>
      </form>

      {initialQuery && (
        <div className="sticky top-0 z-10 bg-white/90 dark:bg-surface-800/90 backdrop-blur-md pb-3 mb-2 -mx-1 px-1 pt-2 space-y-2">
          <div className="flex items-center gap-1.5 flex-wrap">
            {filters.map((f) => {
              const isActive =
                f.id === "all" ? !activeFilter : activeFilter === f.id;
              return (
                <button
                  key={f.id}
                  onClick={() => handleFilterChange(f.id)}
                  className={clsx(
                    "inline-flex items-center gap-1.5 px-3 py-1 text-xs font-medium rounded-full border transition-all",
                    isActive
                      ? "bg-primary-600 dark:bg-primary-500 text-white border-transparent shadow-sm"
                      : "bg-white dark:bg-surface-900 text-surface-500 dark:text-surface-400 border-surface-200 dark:border-surface-700 hover:border-primary-300 dark:hover:border-primary-700 hover:text-primary-600 dark:hover:text-primary-400",
                  )}
                >
                  {f.icon && <f.icon size={12} />}
                  {f.label}
                </button>
              );
            })}

            {user && (
              <button
                type="button"
                onClick={() => {
                  const next = !myItemsOnly;
                  setMyItemsOnly(next);
                  myItemsRef.current = next;
                  if (initialQuery) {
                    doSearch(initialQuery);
                  }
                }}
                className={clsx(
                  "inline-flex items-center gap-1.5 px-3 py-1 text-xs font-medium rounded-full border transition-all",
                  myItemsOnly
                    ? "bg-primary-600 dark:bg-primary-500 text-white border-transparent shadow-sm"
                    : "bg-white dark:bg-surface-900 text-surface-500 dark:text-surface-400 border-surface-200 dark:border-surface-700 hover:border-primary-300 dark:hover:border-primary-700 hover:text-primary-600 dark:hover:text-primary-400",
                )}
              >
                <SlidersHorizontal size={12} />
                Mine
              </button>
            )}

            <div className="ml-auto flex items-center gap-1.5">
              <div className="inline-flex items-center rounded-lg border border-surface-200 dark:border-surface-700 p-0.5 bg-surface-50 dark:bg-surface-800/60 hidden sm:inline-flex">
                <button
                  onClick={() =>
                    setPlatform(platform === "margin" ? "all" : "margin")
                  }
                  title="Margin only"
                  className={clsx(
                    "relative flex items-center justify-center w-7 h-7 rounded-md transition-all group",
                    platform === "margin"
                      ? "bg-white dark:bg-surface-700 shadow-sm"
                      : "hover:bg-surface-100 dark:hover:bg-surface-700/50",
                  )}
                >
                  {platform === "margin" ? (
                    <img
                      src="/logo.svg"
                      alt="Margin"
                      className="w-4 h-4 transition-all"
                    />
                  ) : (
                    <>
                      <img
                        src="/logo.svg"
                        alt="Margin"
                        className="w-4 h-4 transition-all opacity-0 group-hover:opacity-100 absolute"
                      />
                      <div
                        className="w-4 h-4 bg-surface-400 dark:bg-surface-500 group-hover:opacity-0 transition-all"
                        style={{
                          maskImage: "url(/logo.svg)",
                          WebkitMaskImage: "url(/logo.svg)",
                          maskSize: "contain",
                          WebkitMaskSize: "contain",
                          maskRepeat: "no-repeat",
                          WebkitMaskRepeat: "no-repeat",
                          maskPosition: "center",
                          WebkitMaskPosition: "center",
                        }}
                      />
                    </>
                  )}
                </button>
                <button
                  onClick={() =>
                    setPlatform(platform === "semble" ? "all" : "semble")
                  }
                  title="Semble only"
                  className={clsx(
                    "relative flex items-center justify-center w-7 h-7 rounded-md transition-all group",
                    platform === "semble"
                      ? "bg-white dark:bg-surface-700 shadow-sm"
                      : "hover:bg-surface-100 dark:hover:bg-surface-700/50",
                  )}
                >
                  {platform === "semble" ? (
                    <img
                      src="/semble-logo.svg"
                      alt="Semble"
                      className="w-4 h-4 transition-all"
                    />
                  ) : (
                    <>
                      <img
                        src="/semble-logo.svg"
                        alt="Semble"
                        className="w-4 h-4 transition-all opacity-0 group-hover:opacity-100 absolute"
                      />
                      <div
                        className="w-4 h-4 bg-surface-400 dark:bg-surface-500 group-hover:opacity-0 transition-all"
                        style={{
                          maskImage: "url(/semble-logo.svg)",
                          WebkitMaskImage: "url(/semble-logo.svg)",
                          maskSize: "contain",
                          WebkitMaskSize: "contain",
                          maskRepeat: "no-repeat",
                          WebkitMaskRepeat: "no-repeat",
                          maskPosition: "center",
                          WebkitMaskPosition: "center",
                        }}
                      />
                    </>
                  )}
                </button>
              </div>
              <LayoutToggle className="hidden sm:inline-flex" />
            </div>
          </div>
        </div>
      )}

      {loading && results.length === 0 && (
        <div className="flex items-center justify-center py-20 animate-fade-in">
          <Loader2 className="animate-spin text-surface-400" size={24} />
        </div>
      )}

      {loading && results.length > 0 && (
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 z-20">
          <div className="bg-white/90 dark:bg-surface-800/90 shadow-lg rounded-full p-3 backdrop-blur-sm animate-in fade-in zoom-in-95">
            <Loader2
              className="animate-spin text-primary-600 dark:text-primary-400"
              size={24}
            />
          </div>
        </div>
      )}

      {!loading && initialQuery && filteredResults.length === 0 && (
        <EmptyState
          icon={<SearchIcon size={48} />}
          title="No results found"
          message={`Nothing matched "${initialQuery}". Try different keywords.`}
        />
      )}

      {filteredResults.length > 0 && (
        <div
          className={clsx(
            "transition-opacity duration-200 relative",
            loading ? "opacity-40 pointer-events-none" : "opacity-100",
          )}
        >
          <p className="text-xs text-surface-400 dark:text-surface-500 font-medium mb-3 px-1">
            {filteredResults.length}
            {hasMore ? "+" : ""} results for &ldquo;{initialQuery}&rdquo;
          </p>

          {layout === "mosaic" ? (
            <div className="columns-1 sm:columns-2 gap-3 space-y-3">
              {filteredResults.map((item) => (
                <div key={item.uri} className="break-inside-avoid">
                  <Card item={item} onDelete={handleDelete} layout="mosaic" />
                </div>
              ))}
            </div>
          ) : (
            <div className="space-y-3">
              {filteredResults.map((item) => (
                <Card
                  key={item.uri}
                  item={item}
                  onDelete={handleDelete}
                  layout="list"
                />
              ))}
            </div>
          )}

          {hasMore && (
            <button
              onClick={() => doSearch(initialQuery, offset, true)}
              disabled={loading}
              className="w-full py-3 mt-3 text-sm font-medium text-primary-600 dark:text-primary-400 bg-surface-50 dark:bg-surface-800 rounded-xl hover:bg-surface-100 dark:hover:bg-surface-700 transition-colors disabled:opacity-50"
            >
              {loading ? (
                <Loader2 className="animate-spin mx-auto" size={16} />
              ) : (
                "Load more"
              )}
            </button>
          )}
        </div>
      )}

      {!initialQuery && !loading && (
        <EmptyState
          icon={<SearchIcon size={48} />}
          title="Search your library"
          message="Find annotations, highlights, and bookmarks by keyword, URL, or tag."
        />
      )}
    </div>
  );
}
