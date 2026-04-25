import React from "react";
import ExternalLinkModal from "../modals/ExternalLinkModal";
import { useStore } from "@nanostores/react";
import { $preferences } from "../../store/preferences";

interface RichTextProps {
  text: string;
  className?: string;
}

const MENTION_REGEX =
  /(^|[\s(])@([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*)/g;

const URL_REGEX = /(^|[\s(])(https?:\/\/[^\s]+)/g;

export default function RichText({ text, className }: RichTextProps) {
  const urlParts: { text: string; isUrl: boolean }[] = [];
  let lastUrlIndex = 0;

  for (const match of text.matchAll(URL_REGEX)) {
    const fullMatch = match[0];
    const prefix = match[1];
    const url = match[2];
    const startIndex = match.index!;

    if (startIndex > lastUrlIndex) {
      urlParts.push({
        text: text.slice(lastUrlIndex, startIndex),
        isUrl: false,
      });
    }
    if (prefix) {
      urlParts.push({ text: prefix, isUrl: false });
    }

    urlParts.push({ text: url, isUrl: true });

    lastUrlIndex = startIndex + fullMatch.length;
  }
  if (lastUrlIndex < text.length) {
    urlParts.push({ text: text.slice(lastUrlIndex), isUrl: false });
  }

  if (urlParts.length === 0) {
    urlParts.push({ text, isUrl: false });
  }

  const [showExternalLinkModal, setShowExternalLinkModal] =
    React.useState(false);
  const [externalLinkUrl, setExternalLinkUrl] = React.useState<string | null>(
    null,
  );
  const preferences = useStore($preferences);

  const safeUrlHostname = (url: string | null | undefined) => {
    if (!url) return null;
    try {
      return new URL(url).hostname;
    } catch {
      return null;
    }
  };

  const handleExternalClick = (
    e: React.MouseEvent,
    url: string,
    isBareUrl: boolean = false,
  ) => {
    e.preventDefault();
    e.stopPropagation();

    try {
      const hostname = safeUrlHostname(url);
      if (hostname) {
        if (
          hostname === "margin.at" ||
          hostname.endsWith(".margin.at") ||
          hostname === "semble.so" ||
          hostname.endsWith(".semble.so")
        ) {
          window.open(url, "_blank", "noopener,noreferrer");
          return;
        }

        if (isBareUrl || preferences.disableExternalLinkWarning) {
          window.open(url, "_blank", "noopener,noreferrer");
          return;
        }

        const skipped = preferences.externalLinkSkippedHostnames || [];
        if (skipped.includes(hostname)) {
          window.open(url, "_blank", "noopener,noreferrer");
          return;
        }
      }
    } catch (err) {
      if (err instanceof Error && err.name !== "TypeError") {
        console.debug("Failed to check skipped hostname:", err);
      }
    }

    setExternalLinkUrl(url);
    setShowExternalLinkModal(true);
  };

  const finalParts: React.ReactNode[] = [];

  urlParts.forEach((part, partIndex) => {
    if (part.isUrl) {
      finalParts.push(
        <a
          key={`url-${partIndex}`}
          href={part.text}
          target="_blank"
          rel="noopener noreferrer"
          className="text-primary-600 dark:text-primary-400 hover:underline break-all cursor-pointer"
          onClick={(e) => handleExternalClick(e, part.text, true)}
        >
          {part.text}
        </a>,
      );
    } else {
      let lastMentionIndex = 0;
      const mentionMatches = Array.from(part.text.matchAll(MENTION_REGEX));

      if (mentionMatches.length === 0) {
        finalParts.push(part.text);
      } else {
        for (const match of mentionMatches) {
          const fullMatch = match[0];
          const prefix = match[1];
          const handle = match[2];
          const startIndex = match.index!;

          if (startIndex > lastMentionIndex) {
            finalParts.push(part.text.slice(lastMentionIndex, startIndex));
          }

          if (prefix) {
            finalParts.push(prefix);
          }

          finalParts.push(
            <a
              key={`mention-${partIndex}-${startIndex}`}
              href={`/profile/${handle}`}
              className="text-primary-600 dark:text-primary-400 hover:underline"
              onClick={(e) => e.stopPropagation()}
            >
              @{handle}
            </a>,
          );

          lastMentionIndex = startIndex + fullMatch.length;
        }

        if (lastMentionIndex < part.text.length) {
          finalParts.push(part.text.slice(lastMentionIndex));
        }
      }
    }
  });

  return (
    <>
      <span className={className}>{finalParts}</span>
      <ExternalLinkModal
        isOpen={showExternalLinkModal}
        onClose={() => setShowExternalLinkModal(false)}
        url={externalLinkUrl}
      />
    </>
  );
}
