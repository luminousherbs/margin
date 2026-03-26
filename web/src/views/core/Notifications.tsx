import React, { useEffect, useState } from "react";

import { getNotifications, markNotificationsRead } from "../../api/client";
import type { NotificationItem, AnnotationItem } from "../../types";
import {
  Heart,
  MessageCircle,
  Bell,
  PenTool,
  Bookmark,
  UserPlus,
  AtSign,
  ExternalLink,
} from "lucide-react";
import { formatDistanceToNow } from "date-fns";
import { clsx } from "clsx";
import { Avatar, EmptyState, Skeleton } from "../../components/ui";

function getContentType(
  uri: string,
): "annotation" | "highlight" | "bookmark" | "reply" | "unknown" {
  if (uri.includes("/at.margin.annotation/")) return "annotation";
  if (uri.includes("/at.margin.highlight/")) return "highlight";
  if (uri.includes("/at.margin.bookmark/")) return "bookmark";
  if (uri.includes("/at.margin.reply/")) return "reply";
  return "unknown";
}

function getNotificationVerb(
  notifType: string,
  contentType: string,
  subject?: AnnotationItem,
): string {
  switch (notifType) {
    case "like":
      switch (contentType) {
        case "annotation":
          return "liked your annotation";
        case "highlight":
          return "liked your highlight";
        case "bookmark":
          return "liked your bookmark";
        case "reply":
          return "liked your reply";
        default:
          return "liked your post";
      }
    case "reply": {
      const parentUri = subject?.inReplyTo;
      const parentIsReply = parentUri
        ? getContentType(parentUri) === "reply"
        : false;
      return parentIsReply
        ? "replied to your reply"
        : "replied to your annotation";
    }
    case "mention":
      return "mentioned you in an annotation";
    case "follow":
      return "followed you";
    case "highlight":
      return "highlighted your page";
    default:
      return notifType;
  }
}

const NotificationIcon = ({ type }: { type: string }) => {
  const base = "p-2 rounded-full";
  switch (type) {
    case "like":
      return (
        <div className={clsx(base, "bg-red-100 dark:bg-red-900/30")}>
          <Heart size={15} className="text-red-500" />
        </div>
      );
    case "reply":
      return (
        <div className={clsx(base, "bg-blue-100 dark:bg-blue-900/30")}>
          <MessageCircle size={15} className="text-blue-500" />
        </div>
      );
    case "highlight":
      return (
        <div className={clsx(base, "bg-yellow-100 dark:bg-yellow-900/30")}>
          <PenTool size={15} className="text-yellow-600" />
        </div>
      );
    case "bookmark":
      return (
        <div className={clsx(base, "bg-green-100 dark:bg-green-900/30")}>
          <Bookmark size={15} className="text-green-600" />
        </div>
      );
    case "follow":
      return (
        <div className={clsx(base, "bg-purple-100 dark:bg-purple-900/30")}>
          <UserPlus size={15} className="text-purple-500" />
        </div>
      );
    case "mention":
      return (
        <div className={clsx(base, "bg-indigo-100 dark:bg-indigo-900/30")}>
          <AtSign size={15} className="text-indigo-500" />
        </div>
      );
    default:
      return (
        <div className={clsx(base, "bg-surface-100 dark:bg-surface-800")}>
          <Bell size={15} className="text-surface-500" />
        </div>
      );
  }
};

function SubjectPreview({
  subject,
  subjectUri,
}: {
  subject: AnnotationItem | unknown;
  subjectUri: string;
}) {
  const item = subject as AnnotationItem | undefined;
  if (!item?.uri && !subjectUri) return null;

  const contentType = getContentType(subjectUri);
  const href = `/annotation/${encodeURIComponent(subjectUri)}`;

  let preview: React.ReactNode = null;

  if (contentType === "annotation") {
    const quote = item?.target?.selector?.exact;
    const body = item?.text || item?.body?.value;
    preview = (
      <>
        {quote && (
          <p className="text-surface-500 dark:text-surface-400 text-xs italic line-clamp-2 mb-1">
            &ldquo;{quote}&rdquo;
          </p>
        )}
        {body && (
          <p className="text-surface-700 dark:text-surface-300 text-sm line-clamp-2">
            {body}
          </p>
        )}
      </>
    );
  } else if (contentType === "highlight") {
    const quote = item?.target?.selector?.exact;
    preview = quote ? (
      <p className="text-surface-500 dark:text-surface-400 text-xs italic line-clamp-2">
        &ldquo;{quote}&rdquo;
      </p>
    ) : null;
  } else if (contentType === "bookmark") {
    const title = item?.title || item?.target?.title;
    const source = item?.source || item?.target?.source;
    preview = (
      <>
        {title && (
          <p className="text-surface-700 dark:text-surface-300 text-sm font-medium line-clamp-1">
            {title}
          </p>
        )}
        {source && (
          <p className="text-surface-400 dark:text-surface-500 text-xs line-clamp-1 mt-0.5 flex items-center gap-1">
            <ExternalLink size={10} className="shrink-0" />
            {(() => {
              try {
                return new URL(source).hostname;
              } catch {
                return source;
              }
            })()}
          </p>
        )}
      </>
    );
  } else if (contentType === "reply") {
    const text = item?.text;
    const parentUri = item?.inReplyTo;
    const parentIsReply = parentUri
      ? getContentType(parentUri) === "reply"
      : false;
    preview = (
      <>
        {text && (
          <p className="text-surface-700 dark:text-surface-300 text-sm line-clamp-2">
            {text}
          </p>
        )}
        {parentUri && (
          <p className="text-surface-400 dark:text-surface-500 text-xs mt-1">
            in reply to{" "}
            <a
              href={`/annotation/${encodeURIComponent(parentUri)}`}
              className="hover:underline text-primary-500"
              onClick={(e) => e.stopPropagation()}
            >
              {parentIsReply ? "a reply" : "an annotation"}
            </a>
          </p>
        )}
      </>
    );
  }

  if (!preview) return null;

  return (
    <a
      href={href}
      className="block mt-2 pl-3 border-l-2 border-surface-200 dark:border-surface-700 hover:border-primary-400 dark:hover:border-primary-500 transition-colors group"
    >
      {preview}
    </a>
  );
}

export default function Notifications() {
  const [notifications, setNotifications] = useState<NotificationItem[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const load = async () => {
      setLoading(true);
      const data = await getNotifications();
      setNotifications(data);
      setLoading(false);
      markNotificationsRead();
    };
    load();
  }, []);

  if (loading) {
    return (
      <div className="max-w-2xl mx-auto animate-fade-in">
        <h1 className="text-3xl font-display font-bold text-surface-900 dark:text-white mb-6">
          Activity
        </h1>
        <div className="space-y-3">
          {[1, 2, 3].map((i) => (
            <div key={i} className="card p-4 flex gap-3">
              <Skeleton variant="circular" className="w-10 h-10" />
              <div className="flex-1 space-y-2">
                <Skeleton width="60%" />
                <Skeleton width="40%" />
              </div>
            </div>
          ))}
        </div>
      </div>
    );
  }

  if (notifications.length === 0) {
    return (
      <div className="max-w-2xl mx-auto animate-fade-in">
        <h1 className="text-3xl font-display font-bold text-surface-900 dark:text-white mb-6">
          Activity
        </h1>
        <EmptyState
          icon={<Bell size={48} />}
          title="No activity yet"
          message="Interactions with your content will appear here."
        />
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto animate-slide-up">
      <h1 className="text-3xl font-display font-bold text-surface-900 dark:text-white mb-6">
        Activity
      </h1>
      <div className="space-y-2">
        {notifications.map((n) => {
          const contentType = getContentType(n.subjectUri || "");
          const verb = getNotificationVerb(
            n.type,
            contentType,
            n.subject as AnnotationItem,
          );
          const timeAgo = formatDistanceToNow(new Date(n.createdAt), {
            addSuffix: false,
          });

          return (
            <div
              key={n.id}
              className={clsx(
                "card p-4 transition-all",
                !n.readAt &&
                  "ring-2 ring-primary-500/20 dark:ring-primary-400/20 bg-primary-50/30 dark:bg-primary-900/10",
              )}
            >
              <div className="flex gap-3">
                <div className="shrink-0 mt-0.5">
                  <NotificationIcon type={n.type} />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-start gap-2 flex-wrap">
                    <a href={`/profile/${n.actor.did}`} className="shrink-0">
                      <Avatar src={n.actor.avatar} size="xs" />
                    </a>
                    <div className="flex-1 min-w-0">
                      <span className="text-surface-500 dark:text-surface-400 text-sm">
                        <a
                          href={`/profile/${n.actor.did}`}
                          className="font-semibold text-surface-900 dark:text-white hover:underline"
                        >
                          {n.actor.displayName || `@${n.actor.handle}`}
                        </a>{" "}
                        {n.type !== "follow" && n.subjectUri ? (
                          <a
                            href={`/annotation/${encodeURIComponent(n.subjectUri)}`}
                            className="hover:underline"
                          >
                            {verb}
                          </a>
                        ) : (
                          verb
                        )}
                      </span>
                      <span className="text-surface-400 dark:text-surface-500 text-xs ml-1.5">
                        {timeAgo}
                      </span>
                    </div>
                  </div>

                  {n.subject !== undefined && n.subject !== null && (
                    <SubjectPreview
                      subject={n.subject}
                      subjectUri={n.subjectUri || ""}
                    />
                  )}
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
