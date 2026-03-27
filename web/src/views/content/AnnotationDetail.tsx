import React, { useEffect, useRef, useState } from "react";
import { useStore } from "@nanostores/react";
import { $user } from "../../store/auth";
import {
  getAnnotation,
  getReplies,
  resolveHandle,
  createReply,
  deleteReply,
} from "../../api/client";
import type { AnnotationItem } from "../../types";
import Card from "../../components/common/Card";
import ReplyList from "../../components/feed/ReplyList";
import {
  Loader2,
  MessageSquare,
  ArrowLeft,
  X,
  AlertTriangle,
} from "lucide-react";
import { getAvatarUrl } from "../../api/client";

interface AnnotationDetailProps {
  handle?: string;
  rkey?: string;
  type?: string;
  uri?: string;
  did?: string;
  initialAnnotation?: AnnotationItem | null;
  initialReplies?: AnnotationItem[];
  resolvedUri?: string;
}

export default function AnnotationDetail({
  handle,
  rkey,
  type,
  uri,
  did,
  initialAnnotation,
  initialReplies,
  resolvedUri,
}: AnnotationDetailProps) {
  const user = useStore($user);

  const [annotation, setAnnotation] = useState<AnnotationItem | null>(
    initialAnnotation || null,
  );
  const [replies, setReplies] = useState<AnnotationItem[]>(
    initialReplies || [],
  );
  const [loading, setLoading] = useState(!initialAnnotation);
  const [error, setError] = useState<string | null>(null);

  const [replyText, setReplyText] = useState("");
  const [posting, setPosting] = useState(false);
  const [replyingTo, setReplyingTo] = useState<AnnotationItem | null>(null);

  const [targetUri, setTargetUri] = useState<string | null>(
    resolvedUri || uri || null,
  );
  const skipInitialFetch = useRef(!!initialAnnotation);

  useEffect(() => {
    if (resolvedUri) return;

    async function resolve() {
      if (uri) {
        setTargetUri(decodeURIComponent(uri));
        return;
      }

      if (handle && rkey) {
        let collection = "at.margin.annotation";
        if (type === "highlight") collection = "at.margin.highlight";
        if (type === "bookmark") collection = "at.margin.bookmark";

        try {
          const resolvedDid = await resolveHandle(handle);
          if (resolvedDid) {
            setTargetUri(`at://${resolvedDid}/${collection}/${rkey}`);
          } else {
            throw new Error("Could not resolve handle");
          }
        } catch (e) {
          setError(
            "Failed to resolve handle: " +
              (e instanceof Error ? e.message : "Unknown error"),
          );
          setLoading(false);
        }
      } else if (did && rkey) {
        setTargetUri(`at://${did}/at.margin.annotation/${rkey}`);
      }
    }
    resolve();
  }, [uri, did, rkey, handle, type, resolvedUri]);

  const refreshReplies = async () => {
    if (!targetUri) return;
    const repliesData = await getReplies(targetUri);
    setReplies(repliesData.items || []);
  };

  useEffect(() => {
    if (skipInitialFetch.current) {
      skipInitialFetch.current = false;
      return;
    }

    async function fetchData() {
      if (!targetUri) return;

      try {
        setLoading(true);
        const [annData, repliesData] = await Promise.all([
          getAnnotation(targetUri),
          getReplies(targetUri).catch(() => ({
            items: [] as AnnotationItem[],
          })),
        ]);

        if (!annData) {
          setError("Annotation not found");
        } else {
          setAnnotation(annData);
          setReplies(repliesData.items || []);
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : "Unknown error");
      } finally {
        setLoading(false);
      }
    }
    fetchData();
  }, [targetUri]);

  const handleReply = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    if (!replyText.trim() || !annotation || !targetUri) return;

    try {
      setPosting(true);
      const parentUri = replyingTo
        ? replyingTo.uri || replyingTo.id
        : targetUri;
      const parentCid = replyingTo ? replyingTo.cid : annotation.cid;

      if (!parentUri || !parentCid || !annotation.cid)
        throw new Error("Missing parent info");

      await createReply(
        parentUri,
        parentCid,
        targetUri,
        annotation.cid,
        replyText,
      );

      setReplyText("");
      setReplyingTo(null);
      await refreshReplies();
    } catch (err) {
      alert(
        "Failed to post reply: " +
          (err instanceof Error ? err.message : "Unknown error"),
      );
    } finally {
      setPosting(false);
    }
  };

  const handleDeleteReply = async (reply: AnnotationItem) => {
    if (!window.confirm("Delete this reply?")) return;
    try {
      await deleteReply(reply.uri || reply.id!);
      await refreshReplies();
    } catch (err) {
      alert(
        "Failed to delete: " +
          (err instanceof Error ? err.message : "Unknown error"),
      );
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center py-20">
        <Loader2
          className="animate-spin text-primary-600 dark:text-primary-400"
          size={32}
        />
      </div>
    );
  }

  if (error || !annotation) {
    return (
      <div className="max-w-md mx-auto py-12 px-4 text-center">
        <div className="w-14 h-14 bg-surface-100 dark:bg-surface-800 rounded-full flex items-center justify-center mx-auto mb-4 text-surface-400 dark:text-surface-500">
          <AlertTriangle size={28} />
        </div>
        <h3 className="text-xl font-bold text-surface-900 dark:text-white mb-2">
          Not found
        </h3>
        <p className="text-surface-500 dark:text-surface-400 text-sm mb-6">
          {error || "This may have been deleted."}
        </p>
        <a
          href="/home"
          className="px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white font-medium rounded-lg transition-colors"
        >
          Back to Feed
        </a>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto pb-20">
      <div className="mb-4">
        <a
          href="/home"
          className="inline-flex items-center gap-1.5 text-sm font-medium text-surface-500 dark:text-surface-400 hover:text-surface-900 dark:hover:text-white transition-colors"
        >
          <ArrowLeft size={16} />
          Back
        </a>
      </div>

      <Card
        item={annotation}
        onDelete={() => {
          window.location.href = "/home";
        }}
      />

      {annotation.type !== "Bookmark" &&
        annotation.type !== "Highlight" &&
        !annotation.motivation?.includes("bookmark") &&
        !annotation.motivation?.includes("highlight") && (
          <div className="mt-6">
            <h3 className="flex items-center gap-2 text-sm font-semibold text-surface-500 dark:text-surface-400 uppercase tracking-wider mb-4">
              <MessageSquare size={16} />
              Replies ({replies.length})
            </h3>

            {user ? (
              <div className="bg-white dark:bg-surface-900 rounded-xl ring-1 ring-black/5 dark:ring-white/5 p-4 mb-4">
                {replyingTo && (
                  <div className="flex items-center justify-between bg-surface-50 dark:bg-surface-800 px-3 py-2 rounded-lg mb-3 border border-surface-200 dark:border-surface-700">
                    <span className="text-sm text-surface-600 dark:text-surface-300">
                      Replying to{" "}
                      <span className="font-medium text-surface-900 dark:text-white">
                        @
                        {(replyingTo.author || replyingTo.creator)?.handle ||
                          "unknown"}
                      </span>
                    </span>
                    <button
                      onClick={() => setReplyingTo(null)}
                      className="text-surface-400 dark:text-surface-500 hover:text-surface-900 dark:hover:text-white p-1"
                    >
                      <X size={14} />
                    </button>
                  </div>
                )}
                <div className="flex gap-3">
                  {getAvatarUrl(user.did, user.avatar) ? (
                    <img
                      src={getAvatarUrl(user.did, user.avatar)}
                      alt=""
                      className="w-8 h-8 rounded-full object-cover bg-surface-100 dark:bg-surface-800"
                    />
                  ) : (
                    <div className="w-8 h-8 rounded-full bg-surface-100 dark:bg-surface-800 flex items-center justify-center text-xs font-bold text-surface-400 dark:text-surface-500">
                      {user.handle?.[0]?.toUpperCase()}
                    </div>
                  )}
                  <div className="flex-1">
                    <textarea
                      value={replyText}
                      onChange={(e) => setReplyText(e.target.value)}
                      placeholder="Write a reply..."
                      className="w-full p-3 bg-surface-50 dark:bg-surface-800 border border-surface-200 dark:border-surface-700 rounded-lg text-surface-900 dark:text-white placeholder:text-surface-400 dark:placeholder:text-surface-500 focus:ring-2 focus:ring-primary-500/20 focus:border-primary-500 dark:focus:border-primary-400 outline-none resize-none min-h-[80px]"
                      rows={2}
                      disabled={posting}
                    />
                    <div className="flex justify-end mt-2 pt-2 border-t border-surface-100 dark:border-surface-800">
                      <button
                        className="px-4 py-1.5 bg-primary-600 hover:bg-primary-700 text-white text-sm font-medium rounded-full transition-colors disabled:opacity-50"
                        disabled={posting || !replyText.trim()}
                        onClick={() => handleReply()}
                      >
                        {posting ? "..." : "Reply"}
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            ) : (
              <div className="bg-surface-50 dark:bg-surface-800/50 rounded-xl p-5 text-center mb-4 border border-dashed border-surface-200 dark:border-surface-700">
                <p className="text-surface-500 dark:text-surface-400 text-sm mb-2">
                  Sign in to reply
                </p>
                <a
                  href="/login"
                  className="text-primary-600 dark:text-primary-400 font-medium hover:underline text-sm"
                >
                  Log in
                </a>
              </div>
            )}

            <ReplyList
              replies={replies}
              rootUri={targetUri || ""}
              user={user}
              onReply={(reply) => setReplyingTo(reply)}
              onDelete={handleDeleteReply}
              isInline={false}
            />
          </div>
        )}
    </div>
  );
}
