import {
  AlertTriangle,
  ExternalLink,
  Highlighter,
  Loader2,
  PenTool,
  Search,
} from "lucide-react";
import React, { useCallback, useEffect, useState } from "react";
import { getUserTargetItems } from "../../api/client";
import Card from "../../components/common/Card";
import Avatar from "../../components/ui/Avatar";
import { EmptyState, Tabs } from "../../components/ui";
import type { AnnotationItem, UserProfile } from "../../types";

interface UserUrlPageProps {
  handle?: string;
  urlPath?: string;
}

export default function UserUrlPage({ handle, urlPath }: UserUrlPageProps) {
  const targetUrl = urlPath || "";

  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [annotations, setAnnotations] = useState<AnnotationItem[]>([]);
  const [highlights, setHighlights] = useState<AnnotationItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [loadMoreError, setLoadMoreError] = useState<string | null>(null);
  const [hasMore, setHasMore] = useState(false);
  const [offset, setOffset] = useState(0);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<
    "all" | "annotations" | "highlights"
  >("all");

  const LIMIT = 50;
  const [resolvedDid, setResolvedDid] = useState<string | null>(null);

  useEffect(() => {
    async function fetchData() {
      if (!targetUrl || !handle) {
        setLoading(false);
        return;
      }

      try {
        setLoading(true);
        setError(null);

        const profileRes = await fetch(
          `https://public.api.bsky.app/xrpc/app.bsky.actor.getProfile?actor=${encodeURIComponent(handle)}`,
        );

        let did = handle;
        if (profileRes.ok) {
          const profileData = await profileRes.json();
          setProfile(profileData);
          did = profileData.did;
        }

        const decodedUrl = decodeURIComponent(targetUrl);
        setResolvedDid(did);

        const data = await getUserTargetItems(did, decodedUrl, LIMIT, 0);
        const fetchedAnnotations = data.annotations || [];
        const fetchedHighlights = data.highlights || [];
        setAnnotations(fetchedAnnotations);
        setHighlights(fetchedHighlights);
        const totalFetched =
          fetchedAnnotations.length + fetchedHighlights.length;
        setHasMore(totalFetched >= LIMIT);
        setOffset(totalFetched);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Unknown error");
      } finally {
        setLoading(false);
      }
    }
    fetchData();
  }, [handle, targetUrl]);

  const loadMore = useCallback(async () => {
    if (!resolvedDid) return;
    setLoadingMore(true);
    setLoadMoreError(null);
    try {
      const decodedUrl = decodeURIComponent(targetUrl);
      const data = await getUserTargetItems(
        resolvedDid,
        decodedUrl,
        LIMIT,
        offset,
      );
      const fetchedAnnotations = data.annotations || [];
      const fetchedHighlights = data.highlights || [];
      setAnnotations((prev) => [...prev, ...fetchedAnnotations]);
      setHighlights((prev) => [...prev, ...fetchedHighlights]);
      const totalFetched = fetchedAnnotations.length + fetchedHighlights.length;
      setHasMore(totalFetched >= LIMIT);
      setOffset((prev) => prev + totalFetched);
    } catch (err) {
      console.error("Failed to load more:", err);
      const msg = err instanceof Error ? err.message : "Something went wrong";
      setLoadMoreError(msg);
      setTimeout(() => setLoadMoreError(null), 5000);
    } finally {
      setLoadingMore(false);
    }
  }, [resolvedDid, targetUrl, offset]);

  const displayName = profile?.displayName || profile?.handle || handle;
  const displayHandle =
    profile?.handle || (handle?.startsWith("did:") ? null : handle);

  const totalItems = annotations.length + highlights.length;
  const decodedTargetUrl = decodeURIComponent(targetUrl);

  const items = [
    ...(activeTab === "all" || activeTab === "annotations" ? annotations : []),
    ...(activeTab === "all" || activeTab === "highlights" ? highlights : []),
  ];

  if (activeTab === "all") {
    items.sort((a, b) => {
      const dateA = new Date(a.createdAt).getTime();
      const dateB = new Date(b.createdAt).getTime();
      return dateB - dateA;
    });
  }

  if (!targetUrl) {
    return (
      <EmptyState
        icon={<Search size={48} />}
        title="No URL specified"
        message="Please provide a URL to view annotations."
      />
    );
  }

  return (
    <div className="max-w-2xl mx-auto pb-20 animate-fade-in">
      <div className="card p-5 mb-4">
        <div className="flex items-start gap-4">
          <a
            href={`/profile/${displayHandle || handle}`}
            className="shrink-0 hover:opacity-80 transition-opacity"
          >
            <Avatar
              did={profile?.did}
              avatar={profile?.avatar}
              size="lg"
              className="ring-4 ring-surface-100 dark:ring-surface-800"
            />
          </a>
          <div className="flex-1 min-w-0">
            <a
              href={`/profile/${displayHandle || handle}`}
              className="hover:underline"
            >
              <h1 className="text-xl font-bold text-surface-900 dark:text-white truncate">
                {displayName}
              </h1>
            </a>
            {displayHandle && (
              <p className="text-surface-500 dark:text-surface-400">
                @{displayHandle}
              </p>
            )}
          </div>
        </div>

        <div className="mt-4 pt-4 border-t border-surface-100 dark:border-surface-700">
          <div className="flex items-center gap-2 text-sm">
            <span className="text-surface-400 dark:text-surface-500 font-medium shrink-0">
              on
            </span>
            <a
              href={decodedTargetUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="text-primary-600 dark:text-primary-400 hover:underline truncate flex items-center gap-1"
            >
              <span className="truncate">{decodedTargetUrl}</span>
              <ExternalLink size={12} className="shrink-0" />
            </a>
          </div>
        </div>
      </div>

      {loading && (
        <div className="flex flex-col items-center justify-center py-20">
          <Loader2
            className="animate-spin text-primary-600 dark:text-primary-400 mb-4"
            size={32}
          />
          <p className="text-surface-500 dark:text-surface-400">
            Loading annotations...
          </p>
        </div>
      )}

      {error && (
        <div className="bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 p-4 rounded-xl flex items-start gap-3 border border-red-100 dark:border-red-900/30 mb-6">
          <AlertTriangle className="shrink-0 mt-0.5" size={18} />
          <p>{error}</p>
        </div>
      )}

      {!loading && !error && totalItems === 0 && (
        <EmptyState
          icon={<PenTool size={32} />}
          title="No items found"
          message={`${displayName} hasn't annotated this page yet.`}
        />
      )}

      {!loading && !error && totalItems > 0 && (
        <div>
          <div className="mb-6">
            <Tabs
              tabs={[
                { id: "all", label: "All" },
                { id: "annotations", label: "Annotations" },
                { id: "highlights", label: "Highlights" },
              ]}
              activeTab={activeTab}
              onChange={(id: string) =>
                setActiveTab(id as "all" | "annotations" | "highlights")
              }
            />
          </div>

          <div className="space-y-4">
            {activeTab === "annotations" && annotations.length === 0 && (
              <EmptyState
                icon={<PenTool size={32} />}
                title="No annotations"
                message={`${displayName} hasn't annotated this page yet.`}
              />
            )}
            {activeTab === "highlights" && highlights.length === 0 && (
              <EmptyState
                icon={<Highlighter size={32} />}
                title="No highlights"
                message={`${displayName} hasn't highlighted this page yet.`}
              />
            )}

            {items.map((item) => (
              <Card key={item.uri} item={item} />
            ))}
          </div>

          {hasMore && (
            <div className="flex flex-col items-center gap-2 py-6">
              {loadMoreError && (
                <p className="text-sm text-red-500 dark:text-red-400">
                  Failed to load more: {loadMoreError}
                </p>
              )}
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
          )}
        </div>
      )}
    </div>
  );
}
