import React, { useState, useEffect, useCallback } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { useStore } from "@nanostores/react";
import { $user } from "../../store/auth";
import { getByTarget, searchActors } from "../../api/client";
import type { AnnotationItem } from "../../types";
import Card from "../../components/common/Card";
import {
  Search,
  PenTool,
  Highlighter,
  Loader2,
  AlertTriangle,
  Copy,
  Check,
  Clock,
  Globe,
} from "lucide-react";

import { EmptyState, Tabs, Input, Button } from "../../components/ui";

export default function UrlPage() {
  const user = useStore($user);
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const query = searchParams.get("q");

  const [annotations, setAnnotations] = useState<AnnotationItem[]>([]);
  const [highlights, setHighlights] = useState<AnnotationItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<
    "all" | "annotations" | "highlights"
  >("all");
  const [copied, setCopied] = useState(false);
  const [recentSearches, setRecentSearches] = useState<string[]>([]);
  const [searched, setSearched] = useState(false);

  useEffect(() => {
    const stored = localStorage.getItem("margin-recent-searches");
    if (stored) {
      try {
        setRecentSearches(JSON.parse(stored).slice(0, 5));
      } catch (e) {
        console.warn("Failed to parse recent searches", e);
      }
    }
  }, []);

  const saveRecentSearch = useCallback((q: string) => {
    setRecentSearches((prev) => {
      const updated = [q, ...prev.filter((s) => s !== q)].slice(0, 5);
      localStorage.setItem("margin-recent-searches", JSON.stringify(updated));
      return updated;
    });
  }, []);

  useEffect(() => {
    const performSearch = async (urlOrHandle: string) => {
      if (!urlOrHandle.trim()) return;

      setLoading(true);
      setError(null);
      setSearched(true);
      setAnnotations([]);
      setHighlights([]);

      const isProtocol =
        urlOrHandle.startsWith("http://") || urlOrHandle.startsWith("https://");

      if (isProtocol) {
        try {
          const data = await getByTarget(urlOrHandle);
          setAnnotations(data.annotations || []);
          setHighlights(data.highlights || []);
          saveRecentSearch(urlOrHandle);
        } catch (err) {
          setError(err instanceof Error ? err.message : "Search failed");
        } finally {
          setLoading(false);
        }
      } else {
        try {
          const actorRes = await searchActors(urlOrHandle);
          if (actorRes?.actors?.length > 0) {
            const match = actorRes.actors[0];
            navigate(`/profile/${encodeURIComponent(match.handle)}`, {
              replace: true,
            });
            return;
          } else {
            setError(
              "User not found. To search for a URL, please include 'http://' or 'https://'.",
            );
            setLoading(false);
          }
        } catch {
          setError("Failed to search user.");
          setLoading(false);
        }
      }
    };

    if (query) {
      performSearch(query);
    } else {
      setSearched(false);
      setAnnotations([]);
      setHighlights([]);
      setLoading(false);
    }
  }, [query, navigate, saveRecentSearch]);

  const myAnnotations = user
    ? annotations.filter((a) => (a.author?.did || a.creator?.did) === user.did)
    : [];
  const myHighlights = user
    ? highlights.filter((h) => (h.author?.did || h.creator?.did) === user.did)
    : [];
  const myItemsCount = myAnnotations.length + myHighlights.length;

  const getShareUrl = () => {
    if (!user?.handle || !query) return null;
    return `${window.location.origin}/${user.handle}/url/${encodeURIComponent(query)}`;
  };

  const handleCopyShareLink = async () => {
    const shareUrl = getShareUrl();
    if (!shareUrl) return;
    try {
      await navigator.clipboard.writeText(shareUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error("Failed to copy link:", err);
    }
  };

  const totalItems = annotations.length + highlights.length;

  const renderResults = () => {
    if (activeTab === "annotations" && annotations.length === 0) {
      return (
        <EmptyState
          icon={<PenTool size={32} />}
          title="No annotations"
          message="There are no annotations for this URL yet."
        />
      );
    }

    if (activeTab === "highlights" && highlights.length === 0) {
      return (
        <EmptyState
          icon={<Highlighter size={32} />}
          title="No highlights"
          message="There are no highlights for this URL yet."
        />
      );
    }

    return (
      <div className="space-y-4">
        {(activeTab === "all" || activeTab === "annotations") &&
          annotations.map((a) => <Card key={a.uri} item={a} />)}
        {(activeTab === "all" || activeTab === "highlights") &&
          highlights.map((h) => <Card key={h.uri} item={h} />)}
      </div>
    );
  };

  const handleRecentClick = (q: string) => {
    navigate(`/url?q=${encodeURIComponent(q)}`);
  };

  return (
    <div className="max-w-2xl mx-auto pb-20 animate-fade-in">
      {!query && (
        <div className="text-center py-10">
          <div className="w-16 h-16 bg-primary-50 dark:bg-primary-900/20 rounded-2xl flex items-center justify-center mx-auto mb-6 rotate-3">
            <Search
              size={32}
              className="text-primary-600 dark:text-primary-400"
            />
          </div>
          <h1 className="text-3xl font-display font-bold text-surface-900 dark:text-white mb-3">
            Explore
          </h1>
          <p className="text-surface-500 dark:text-surface-400 max-w-md mx-auto mb-8">
            Search for any URL in the sidebar to see specific annotations and
            highlights.
          </p>

          <form
            onSubmit={(e) => {
              e.preventDefault();
              const formData = new FormData(e.currentTarget);
              const q = formData.get("q") as string;
              if (q?.trim()) {
                navigate(`/url?q=${encodeURIComponent(q.trim())}`);
              }
            }}
            className="max-w-md mx-auto mb-8 flex gap-2"
          >
            <div className="flex-1">
              <Input
                name="q"
                placeholder="https://example.com"
                className="w-full bg-surface-50 dark:bg-surface-800"
                autoFocus
              />
            </div>
            <Button type="submit">Search</Button>
          </form>

          {recentSearches.length > 0 && (
            <div className="text-left max-w-lg mx-auto bg-surface-50 dark:bg-surface-800/50 rounded-2xl p-6 border border-surface-100 dark:border-surface-800">
              <h3 className="text-sm font-bold text-surface-900 dark:text-white mb-4 flex items-center gap-2">
                <Clock size={16} className="text-primary-500" />
                Recent Searches
              </h3>
              <div className="flex flex-wrap gap-2">
                {recentSearches.map((q, i) => (
                  <button
                    key={i}
                    onClick={() => handleRecentClick(q)}
                    className="px-3 py-1.5 bg-white dark:bg-surface-700 hover:bg-surface-50 dark:hover:bg-surface-600 rounded-lg text-sm text-surface-700 dark:text-surface-200 transition-colors shadow-sm ring-1 ring-black/5 dark:ring-white/5 flex items-center gap-2"
                  >
                    <Globe size={12} className="opacity-50" />
                    <span className="truncate max-w-[200px]">
                      {q.replace(/^https?:\/\//, "")}
                    </span>
                  </button>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {loading && (
        <div className="flex flex-col items-center justify-center py-20">
          <Loader2
            className="animate-spin text-primary-600 dark:text-primary-400 mb-4"
            size={32}
          />
          <p className="text-surface-500 dark:text-surface-400">Searching...</p>
        </div>
      )}

      {error && (
        <div className="bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 p-4 rounded-xl flex items-start gap-3 border border-red-100 dark:border-red-900/30 mb-6">
          <AlertTriangle className="shrink-0 mt-0.5" size={18} />
          <p>{error}</p>
        </div>
      )}

      {searched && !loading && !error && totalItems === 0 && (
        <EmptyState
          icon={<Search size={48} />}
          title="No results found"
          message="We couldn't find any annotations for this URL. Be the first to add one!"
        />
      )}

      {searched && !loading && !error && totalItems > 0 && (
        <div>
          <div className="flex items-center justify-between gap-4 mb-6">
            <div>
              <h1 className="text-2xl font-bold text-surface-900 dark:text-white truncate max-w-md">
                {query?.replace(/^https?:\/\//, "")}
              </h1>
              <p className="text-surface-500 dark:text-surface-400 text-sm">
                {totalItems} result{totalItems !== 1 ? "s" : ""} found
              </p>
            </div>

            {user && myItemsCount > 0 && (
              <button
                onClick={handleCopyShareLink}
                className="flex items-center gap-1.5 px-3 py-1.5 bg-surface-100 dark:bg-surface-800 hover:bg-surface-200 dark:hover:bg-surface-700 text-surface-900 dark:text-white text-sm font-medium rounded-lg transition-colors"
              >
                {copied ? <Check size={14} /> : <Copy size={14} />}
                {copied ? "Copied" : "Share your thoughts on this URL"}
              </button>
            )}
          </div>

          <div className="mb-6">
            <Tabs
              tabs={[
                { id: "all", label: `All (${totalItems})` },
                {
                  id: "annotations",
                  label: `Annotations (${annotations.length})`,
                },
                {
                  id: "highlights",
                  label: `Highlights (${highlights.length})`,
                },
              ]}
              activeTab={activeTab}
              onChange={(id: string) =>
                setActiveTab(id as "all" | "annotations" | "highlights")
              }
            />
          </div>

          {renderResults()}
        </div>
      )}
    </div>
  );
}
