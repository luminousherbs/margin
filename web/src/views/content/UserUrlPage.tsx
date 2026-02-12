import React, { useState, useEffect } from "react";
import { useParams } from "react-router-dom";
import { getUserTargetItems } from "../../api/client";
import type { AnnotationItem, UserProfile } from "../../types";
import Card from "../../components/common/Card";
import {
  PenTool,
  Highlighter,
  Search,
  AlertTriangle,
  ExternalLink,
} from "lucide-react";
import { clsx } from "clsx";
import { getAvatarUrl } from "../../api/client";

export default function UserUrlPage() {
  const params = useParams();
  const handle = params.handle;
  const urlPath = params["*"];
  const targetUrl = urlPath || "";

  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [annotations, setAnnotations] = useState<AnnotationItem[]>([]);
  const [highlights, setHighlights] = useState<AnnotationItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<
    "all" | "annotations" | "highlights"
  >("all");

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

        const data = await getUserTargetItems(did, decodedUrl);
        setAnnotations(data.annotations || []);
        setHighlights(data.highlights || []);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Unknown error");
      } finally {
        setLoading(false);
      }
    }
    fetchData();
  }, [handle, targetUrl]);

  const displayName = profile?.displayName || profile?.handle || handle;
  const displayHandle =
    profile?.handle || (handle?.startsWith("did:") ? null : handle);
  const avatarUrl = getAvatarUrl(profile?.did, profile?.avatar);

  const getInitial = () => {
    return (displayName || displayHandle || "??")
      ?.substring(0, 2)
      .toUpperCase();
  };

  const totalItems = annotations.length + highlights.length;
  const bskyProfileUrl = displayHandle
    ? `https://bsky.app/profile/${displayHandle}`
    : `https://bsky.app/profile/${handle}`;

  const renderResults = () => {
    if (activeTab === "annotations" && annotations.length === 0) {
      return (
        <div className="flex flex-col items-center justify-center p-12 text-center bg-surface-50 border border-dashed border-surface-200 rounded-2xl">
          <div className="w-12 h-12 bg-surface-100 rounded-full flex items-center justify-center text-surface-400 mb-4">
            <PenTool size={24} />
          </div>
          <h3 className="text-lg font-medium text-surface-600">
            No annotations
          </h3>
        </div>
      );
    }

    if (activeTab === "highlights" && highlights.length === 0) {
      return (
        <div className="flex flex-col items-center justify-center p-12 text-center bg-surface-50 border border-dashed border-surface-200 rounded-2xl">
          <div className="w-12 h-12 bg-surface-100 rounded-full flex items-center justify-center text-surface-400 mb-4">
            <Highlighter size={24} />
          </div>
          <h3 className="text-lg font-medium text-surface-600">
            No highlights
          </h3>
        </div>
      );
    }

    return (
      <div className="space-y-6">
        {(activeTab === "all" || activeTab === "annotations") &&
          annotations.map((a) => <Card key={a.uri} item={a} />)}
        {(activeTab === "all" || activeTab === "highlights") &&
          highlights.map((h) => <Card key={h.uri} item={h} />)}
      </div>
    );
  };

  if (!targetUrl) {
    return (
      <div className="max-w-2xl mx-auto py-20 text-center">
        <div className="w-16 h-16 bg-surface-100 rounded-full flex items-center justify-center mx-auto mb-4 text-surface-400">
          <Search size={32} />
        </div>
        <h3 className="text-xl font-bold text-surface-900 mb-2">
          No URL specified
        </h3>
        <p className="text-surface-500">
          Please provide a URL to view annotations.
        </p>
      </div>
    );
  }

  return (
    <div className="max-w-3xl mx-auto pb-20">
      <header className="flex items-center gap-6 mb-8 p-6 bg-white dark:bg-surface-800 rounded-2xl border border-surface-200 dark:border-surface-700 shadow-sm">
        <a
          href={bskyProfileUrl}
          target="_blank"
          rel="noopener noreferrer"
          className="shrink-0 hover:opacity-80 transition-opacity"
        >
          {avatarUrl ? (
            <img
              src={avatarUrl}
              alt={displayName}
              className="w-20 h-20 rounded-full object-cover border-4 border-surface-50 dark:border-surface-700"
            />
          ) : (
            <div className="w-20 h-20 rounded-full bg-surface-100 dark:bg-surface-700 flex items-center justify-center text-2xl font-bold text-surface-500 dark:text-surface-400 border-4 border-surface-50 dark:border-surface-700">
              {getInitial()}
            </div>
          )}
        </a>
        <div className="flex-1">
          <h1 className="text-2xl font-bold text-surface-900 dark:text-white mb-1">
            {displayName}
          </h1>
          {displayHandle && (
            <a
              href={bskyProfileUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="text-surface-500 dark:text-surface-400 hover:text-primary-600 transition-colors bg-surface-50 dark:bg-surface-700 hover:bg-primary-50 dark:hover:bg-primary-900/30 px-2 py-1 rounded-md text-sm inline-flex items-center gap-1"
            >
              @{displayHandle} <ExternalLink size={12} />
            </a>
          )}
        </div>
      </header>

      <div className="mb-8 p-4 bg-surface-50 dark:bg-surface-800 border border-surface-200 dark:border-surface-700 rounded-xl flex flex-col sm:flex-row sm:items-center gap-4">
        <span className="text-sm font-semibold text-surface-500 dark:text-surface-400 uppercase tracking-wide">
          Annotations on:
        </span>
        <a
          href={decodeURIComponent(targetUrl)}
          target="_blank"
          rel="noopener noreferrer"
          className="text-primary-600 hover:text-primary-700 hover:underline font-medium truncate flex-1 block"
        >
          {decodeURIComponent(targetUrl)}
        </a>
      </div>

      {loading && (
        <div className="flex justify-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
        </div>
      )}

      {error && (
        <div className="mb-8 bg-red-50 text-red-600 p-4 rounded-xl flex items-start gap-3 border border-red-100">
          <AlertTriangle className="shrink-0 mt-0.5" size={18} />
          <p>{error}</p>
        </div>
      )}

      {!loading && !error && totalItems === 0 && (
        <div className="text-center py-16 bg-surface-50 dark:bg-surface-800 rounded-2xl border border-dashed border-surface-200 dark:border-surface-700">
          <div className="w-12 h-12 bg-surface-100 dark:bg-surface-700 rounded-full flex items-center justify-center mx-auto mb-4 text-surface-400">
            <PenTool size={24} />
          </div>
          <h3 className="text-lg font-bold text-surface-900 dark:text-white mb-1">
            No items found
          </h3>
          <p className="text-surface-500 dark:text-surface-400">
            {displayName} hasn&apos;t annotated this page yet.
          </p>
        </div>
      )}

      {!loading && !error && totalItems > 0 && (
        <div className="animate-fade-in">
          <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-6">
            <h2 className="text-xl font-bold text-surface-900 dark:text-white">
              {totalItems} item{totalItems !== 1 ? "s" : ""}
            </h2>
            <div className="flex bg-surface-100 dark:bg-surface-800 p-1 rounded-xl self-start md:self-auto">
              <button
                className={clsx(
                  "px-4 py-1.5 rounded-lg text-sm font-medium transition-all",
                  activeTab === "all"
                    ? "bg-white dark:bg-surface-700 text-surface-900 dark:text-white shadow-sm"
                    : "text-surface-500 dark:text-surface-400 hover:text-surface-700 dark:hover:text-surface-200",
                )}
                onClick={() => setActiveTab("all")}
              >
                All ({totalItems})
              </button>
              <button
                className={clsx(
                  "px-4 py-1.5 rounded-lg text-sm font-medium transition-all",
                  activeTab === "annotations"
                    ? "bg-white dark:bg-surface-700 text-surface-900 dark:text-white shadow-sm"
                    : "text-surface-500 dark:text-surface-400 hover:text-surface-700 dark:hover:text-surface-200",
                )}
                onClick={() => setActiveTab("annotations")}
              >
                Annotations ({annotations.length})
              </button>
              <button
                className={clsx(
                  "px-4 py-1.5 rounded-lg text-sm font-medium transition-all",
                  activeTab === "highlights"
                    ? "bg-white dark:bg-surface-700 text-surface-900 dark:text-white shadow-sm"
                    : "text-surface-500 dark:text-surface-400 hover:text-surface-700 dark:hover:text-surface-200",
                )}
                onClick={() => setActiveTab("highlights")}
              >
                Highlights ({highlights.length})
              </button>
            </div>
          </div>
          <div className="space-y-6">{renderResults()}</div>
        </div>
      )}
    </div>
  );
}
