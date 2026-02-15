import { useStore } from "@nanostores/react";
import { clsx } from "clsx";
import {
  Bookmark,
  Clock,
  Highlighter,
  Loader2,
  MessageSquareText,
} from "lucide-react";
import React, { useCallback, useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { getFeed } from "../../api/client";
import Card from "../../components/common/Card";
import { Button, EmptyState, Tabs } from "../../components/ui";
import LayoutToggle from "../../components/ui/LayoutToggle";
import { $user } from "../../store/auth";
import { $feedLayout } from "../../store/feedLayout";
import type { AnnotationItem } from "../../types";

interface FeedProps {
  initialType?: string;
  motivation?: string;
  showTabs?: boolean;
  emptyMessage?: string;
}

function FeedContent({
  type,
  motivation,
  emptyMessage,
  layout,
  tag,
}: {
  type: string;
  motivation?: string;
  emptyMessage: string;
  layout: "list" | "mosaic";
  tag?: string;
}) {
  const [items, setItems] = useState<AnnotationItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [hasMore, setHasMore] = useState(false);
  const [offset, setOffset] = useState(0);

  const LIMIT = 50;

  useEffect(() => {
    let cancelled = false;

    getFeed({ type, motivation, tag, limit: LIMIT, offset: 0 })
      .then((data) => {
        if (cancelled) return;
        const fetched = data.items;
        setItems(fetched);
        setHasMore(data.hasMore);
        setOffset(data.fetchedCount);
        setLoading(false);
      })
      .catch((e) => {
        if (cancelled) return;
        console.error(e);
        setItems([]);
        setHasMore(false);
        setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [type, motivation, tag]);

  const loadMore = useCallback(async () => {
    setLoadingMore(true);
    try {
      const data = await getFeed({
        type,
        motivation,
        tag,
        limit: LIMIT,
        offset,
      });
      const fetched = data?.items || [];
      setItems((prev) => [...prev, ...fetched]);
      setHasMore(data.hasMore);
      setOffset((prev) => prev + data.fetchedCount);
    } catch (e) {
      console.error(e);
    } finally {
      setLoadingMore(false);
    }
  }, [type, motivation, tag, offset]);

  const handleDelete = (uri: string) => {
    setItems((prev) => prev.filter((i) => i.uri !== uri));
  };

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center py-20 gap-3">
        <Loader2
          className="animate-spin text-primary-600 dark:text-primary-400"
          size={32}
        />
        <p className="text-sm text-surface-400 dark:text-surface-500">
          Loading feed...
        </p>
      </div>
    );
  }

  if (items.length === 0) {
    return (
      <EmptyState
        icon={<Clock size={48} />}
        title="Nothing here yet"
        message={emptyMessage}
      />
    );
  }

  const loadMoreButton = hasMore && (
    <div className="flex justify-center py-6">
      <button
        onClick={loadMore}
        disabled={loadingMore}
        className="inline-flex items-center gap-2 px-5 py-2.5 text-sm font-medium rounded-xl bg-surface-100 dark:bg-surface-800 text-surface-600 dark:text-surface-300 hover:bg-surface-200 dark:hover:bg-surface-700 transition-colors disabled:opacity-50"
      >
        {loadingMore ? (
          <>
            <Loader2 size={16} className="animate-spin" />
            Loading...
          </>
        ) : (
          "Load more"
        )}
      </button>
    </div>
  );

  if (layout === "mosaic") {
    return (
      <>
        <div className="columns-1 sm:columns-2 xl:columns-3 2xl:columns-4 gap-4 animate-fade-in">
          {items.map((item) => (
            <div key={item.uri || item.cid} className="break-inside-avoid mb-4">
              <Card item={item} onDelete={handleDelete} layout="mosaic" />
            </div>
          ))}
        </div>
        {loadMoreButton}
      </>
    );
  }

  return (
    <>
      <div className="space-y-3 animate-fade-in">
        {items.map((item) => (
          <Card
            key={item.uri || item.cid}
            item={item}
            onDelete={handleDelete}
          />
        ))}
      </div>
      {loadMoreButton}
    </>
  );
}

export default function Feed({
  initialType = "all",
  motivation,
  showTabs = true,
  emptyMessage = "No items found.",
}: FeedProps) {
  const [searchParams, setSearchParams] = useSearchParams();
  const tag = searchParams.get("tag") || undefined;
  const user = useStore($user);
  const layout = useStore($feedLayout);
  const [activeTab, setActiveTab] = useState(initialType);
  const [activeFilter, setActiveFilter] = useState<string | undefined>(
    motivation,
  );

  const handleTabChange = (id: string) => {
    if (id === activeTab) return;
    setActiveTab(id);
    setSearchParams((prev) => {
      const newParams = new URLSearchParams(prev);
      newParams.delete("tag");
      return newParams;
    });
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  const handleFilterChange = (id: string) => {
    const next = id === "all" ? undefined : id;
    if (next === activeFilter) return;
    setActiveFilter(next);
    setSearchParams((prev) => {
      const newParams = new URLSearchParams(prev);
      newParams.delete("tag");
      return newParams;
    });
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  const tabs = [
    { id: "all", label: "Recent" },
    { id: "popular", label: "Popular" },
    { id: "shelved", label: "Shelved" },
    { id: "margin", label: "Margin" },
    { id: "semble", label: "Semble" },
  ];

  const filters = [
    { id: "all", label: "All", icon: null },
    { id: "commenting", label: "Annotations", icon: MessageSquareText },
    { id: "highlighting", label: "Highlights", icon: Highlighter },
    { id: "bookmarking", label: "Bookmarks", icon: Bookmark },
  ];

  return (
    <div className="mx-auto max-w-2xl xl:max-w-none">
      {!user && (
        <div className="text-center py-10 px-6 mb-4 animate-fade-in">
          <h1 className="text-2xl font-display font-bold mb-2 tracking-tight text-surface-900 dark:text-white">
            Welcome to Margin
          </h1>
          <p className="text-surface-500 dark:text-surface-400 mb-4 max-w-md mx-auto">
            Annotate, highlight, and bookmark anything on the web.
          </p>
          <div className="flex gap-3 justify-center">
            <Button onClick={() => (window.location.href = "/login")}>
              Get Started
            </Button>
            <Button
              variant="secondary"
              onClick={() => window.open("/about", "_blank")}
            >
              Learn More
            </Button>
          </div>
        </div>
      )}

      {showTabs && (
        <div className="sticky top-0 z-10 bg-white/95 dark:bg-surface-800/95 backdrop-blur-sm pb-3 mb-2 -mx-1 px-1 pt-1 space-y-2">
          {!tag && (
            <Tabs
              tabs={tabs}
              activeTab={activeTab}
              onChange={handleTabChange}
            />
          )}
          {tag && (
            <div className="flex items-center justify-between mb-2">
              <h2 className="text-xl font-bold flex items-center gap-2">
                <span className="text-surface-500 font-normal">
                  Items with tag:
                </span>
                <span className="bg-primary-50 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400 px-2 py-0.5 rounded-lg">
                  #{tag}
                </span>
              </h2>
              <button
                onClick={() => {
                  setSearchParams((prev) => {
                    const newParams = new URLSearchParams(prev);
                    newParams.delete("tag");
                    return newParams;
                  });
                }}
                className="text-sm text-surface-500 hover:text-surface-900 dark:hover:text-white"
              >
                Clear filter
              </button>
            </div>
          )}
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
            <div className="ml-auto">
              <LayoutToggle className="hidden sm:inline-flex" />
            </div>
          </div>
        </div>
      )}

      <FeedContent
        key={`${activeTab}-${activeFilter || "all"}-${tag || ""}`}
        type={activeTab}
        motivation={activeFilter}
        tag={tag}
        emptyMessage={emptyMessage}
        layout={layout}
      />
    </div>
  );
}
