import { Clock, Loader2 } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";
import { type GetFeedParams, getFeed } from "../../api/client";
import Card from "../../components/common/Card";
import { EmptyState } from "../../components/ui";
import type { AnnotationItem } from "../../types";

const LIMIT = 50;

const feedCache = new Map<
  string,
  {
    items: AnnotationItem[];
    hasMore: boolean;
    offset: number;
    timestamp: number;
  }
>();

export interface FeedItemsProps extends Omit<
  GetFeedParams,
  "limit" | "offset"
> {
  layout: "list" | "mosaic";
  emptyMessage: string;
  initialItems?: AnnotationItem[];
  initialHasMore?: boolean;
}

export default function FeedItems({
  creator,
  source,
  tag,
  type,
  motivation,
  emptyMessage,
  layout,
  initialItems,
  initialHasMore,
}: FeedItemsProps) {
  const [items, setItems] = useState<AnnotationItem[]>(initialItems || []);
  const [loading, setLoading] = useState(!initialItems);
  const [loadingMore, setLoadingMore] = useState(false);
  const [hasMore, setHasMore] = useState(initialHasMore ?? false);
  const [offset, setOffset] = useState(initialItems?.length ?? 0);
  const skipInitialFetch = useRef(!!initialItems);

  useEffect(() => {
    if (skipInitialFetch.current) {
      skipInitialFetch.current = false;
      return;
    }

    let cancelled = false;
    const cacheKey = JSON.stringify({ type, motivation, tag, creator, source });
    const cached = feedCache.get(cacheKey);

    if (cached && Date.now() - cached.timestamp < 5 * 60 * 1000) {
      setItems(cached.items);
      setHasMore(cached.hasMore);
      setOffset(cached.offset);
      setLoading(false);

      getFeed({
        type,
        motivation,
        tag,
        creator,
        source,
        limit: LIMIT,
        offset: 0,
      })
        .then((data) => {
          if (cancelled) return;
          const fetched = data.items;
          setItems(fetched);
          setHasMore(data.hasMore);
          setOffset(data.fetchedCount);
          feedCache.set(cacheKey, {
            items: fetched,
            hasMore: data.hasMore,
            offset: data.fetchedCount,
            timestamp: Date.now(),
          });
        })
        .catch(console.error);

      return () => {
        cancelled = true;
      };
    }

    setLoading(true);
    getFeed({ type, motivation, tag, creator, source, limit: LIMIT, offset: 0 })
      .then((data) => {
        if (cancelled) return;
        const fetched = data.items;
        setItems(fetched);
        setHasMore(data.hasMore);
        setOffset(data.fetchedCount);
        setLoading(false);
        feedCache.set(cacheKey, {
          items: fetched,
          hasMore: data.hasMore,
          offset: data.fetchedCount,
          timestamp: Date.now(),
        });
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
  }, [type, motivation, tag, creator, source]);

  const loadMore = useCallback(async () => {
    setLoadingMore(true);
    try {
      const cacheKey = JSON.stringify({
        type,
        motivation,
        tag,
        creator,
        source,
      });
      const data = await getFeed({
        type,
        motivation,
        tag,
        creator,
        source,
        limit: LIMIT,
        offset,
      });
      const fetched = data?.items || [];
      const newItems = [...items, ...fetched];
      setItems(newItems);
      setHasMore(data.hasMore);
      const newOffset = offset + data.fetchedCount;
      setOffset(newOffset);
      feedCache.set(cacheKey, {
        items: newItems,
        hasMore: data.hasMore,
        offset: newOffset,
        timestamp: Date.now(),
      });
    } catch (e) {
      console.error(e);
    } finally {
      setLoadingMore(false);
    }
  }, [type, motivation, tag, creator, source, offset, items]);

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
          Loading...
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
        type="button"
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
