import React, { useState, useRef, useEffect } from "react";
import {
  Copy,
  ExternalLink,
  Check,
  Share2,
  MoreHorizontal,
} from "lucide-react";
import {
  AturiIcon,
  BlueskyIcon,
  BlackskyIcon,
  WitchskyIcon,
  CatskyIcon,
  DeerIcon,
} from "../common/Icons";
import { analytics } from "../../lib/analytics";

const SembleLogo = () => (
  <img src="/semble-logo.svg" alt="Semble" className="w-4 h-4 opacity-90" />
);

const BLUESKY_COLOR = "#1185fe";

interface ShareMenuProps {
  uri: string;
  text?: string;
  customUrl?: string;
  handle?: string;
  type?: string;
  url?: string;
}

export default function ShareMenu({
  uri,
  text,
  customUrl,
  handle,
  type,
  url,
}: ShareMenuProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [copied, setCopied] = useState<string | null>(null);
  const menuRef = useRef<HTMLDivElement>(null);
  const buttonRef = useRef<HTMLButtonElement>(null);
  const [menuPosition, setMenuPosition] = useState({
    top: 0,
    left: 0,
    alignRight: false,
  });

  const getShareUrl = () => {
    if (customUrl) return customUrl;
    if (!uri) return "";

    const origin = typeof window !== "undefined" ? window.location.origin : "";
    const uriParts = uri.split("/");
    const rkey = uriParts[uriParts.length - 1];
    const did = uriParts[2];

    if (uri.includes("network.cosmik.card"))
      return `${origin}/at/${did}/${rkey}`;
    if (handle && type)
      return `${origin}/${handle}/${type.toLowerCase()}/${rkey}`;
    return `${origin}/at/${did}/${rkey}`;
  };

  const shareUrl = getShareUrl();
  const isSemble = uri && uri.includes("network.cosmik");

  const sembleUrl = (() => {
    if (!isSemble) return "";
    const parts = (uri || "").split("/");
    const rkey = parts[parts.length - 1];
    const userHandle = handle || (parts.length > 2 ? parts[2] : "");

    if (uri.includes("network.cosmik.collection"))
      return `https://semble.so/profile/${userHandle}/collections/${rkey}`;
    if (uri.includes("network.cosmik.card") && url)
      return `https://semble.so/url?id=${encodeURIComponent(url)}`;
    return `https://semble.so/profile/${userHandle}`;
  })();

  const handleCopy = async (textToCopy: string, key: string) => {
    try {
      await navigator.clipboard.writeText(textToCopy);
      setCopied(key);
      analytics.capture("item_shared", {
        method: "copy_link",
        destination: key,
        item_type: type,
      });
      setTimeout(() => {
        setCopied(null);
        setIsOpen(false);
      }, 1000);
    } catch {
      prompt("Copy this link:", textToCopy);
    }
  };

  const handleShareToFork = (domain: string) => {
    const composeText = text
      ? `${text.substring(0, 200)}...\n\n${shareUrl}`
      : shareUrl;
    const composeUrl = `https://${domain}/intent/compose?text=${encodeURIComponent(composeText)}`;
    analytics.capture("item_shared", {
      method: "social_app",
      destination: domain,
      item_type: type,
    });
    window.open(composeUrl, "_blank");
    setIsOpen(false);
  };

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (
        menuRef.current &&
        !menuRef.current.contains(e.target as Node) &&
        !buttonRef.current?.contains(e.target as Node)
      ) {
        setIsOpen(false);
      }
    };
    if (isOpen) {
      document.addEventListener("mousedown", handleClickOutside);
      window.addEventListener("scroll", () => setIsOpen(false), true);
      window.addEventListener("resize", () => setIsOpen(false));
    }
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
      window.removeEventListener("scroll", () => setIsOpen(false), true);
      window.removeEventListener("resize", () => setIsOpen(false));
    };
  }, [isOpen]);

  const calculatePosition = () => {
    if (!buttonRef.current) return;
    const rect = buttonRef.current.getBoundingClientRect();
    const menuWidth = 240;

    let top = rect.bottom + 8;
    let left = rect.left;
    let alignRight = false;

    if (left + menuWidth > window.innerWidth - 16) {
      left = rect.right - menuWidth;
      alignRight = true;
    }

    if (top + 300 > window.innerHeight) {
      top = rect.top - 8;
    }

    setMenuPosition({ top, left, alignRight });
  };

  const toggleMenu = () => {
    if (!isOpen) calculatePosition();
    setIsOpen(!isOpen);
  };

  const renderMenuItem = (
    label: string,
    icon: React.ReactNode,
    onClick: () => void,
    isCopied: boolean = false,
    highlight: boolean = false,
  ) => (
    <button
      onClick={onClick}
      className={`w-full flex items-center gap-3 px-3 py-2.5 text-[14px] font-medium transition-colors rounded-lg group
                ${
                  highlight
                    ? "text-primary-700 dark:text-primary-400 bg-primary-50/50 dark:bg-primary-900/20 hover:bg-primary-50 dark:hover:bg-primary-900/30"
                    : "text-surface-700 dark:text-surface-200 hover:bg-surface-50 dark:hover:bg-surface-800 hover:text-surface-900 dark:hover:text-white"
                }`}
    >
      <span
        className={`flex items-center justify-center w-5 h-5 ${highlight ? "text-primary-600 dark:text-primary-400" : "text-surface-400 dark:text-surface-500 group-hover:text-surface-600 dark:group-hover:text-surface-300"}`}
      >
        {isCopied ? (
          <Check size={16} className="text-green-600 dark:text-green-400" />
        ) : (
          icon
        )}
      </span>
      <span className="flex-1 text-left">{isCopied ? "Copied!" : label}</span>
    </button>
  );

  const shareForks = [
    {
      name: "Bluesky",
      domain: "bsky.app",
      icon: <BlueskyIcon size={18} color={BLUESKY_COLOR} />,
    },
    {
      name: "Witchsky",
      domain: "witchsky.app",
      icon: <WitchskyIcon size={18} />,
    },
    {
      name: "Blacksky",
      domain: "blacksky.community",
      icon: <BlackskyIcon size={18} />,
    },
    { name: "Catsky", domain: "catsky.social", icon: <CatskyIcon size={18} /> },
    { name: "Deer", domain: "deer.social", icon: <DeerIcon size={18} /> },
  ];

  return (
    <div className="relative inline-block">
      <button
        ref={buttonRef}
        onClick={toggleMenu}
        className={`flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg transition-all ${isOpen ? "text-primary-600 dark:text-primary-400 bg-primary-50 dark:bg-primary-900/20" : "text-surface-400 dark:text-surface-500 hover:text-primary-600 dark:hover:text-primary-400 hover:bg-primary-50 dark:hover:bg-primary-900/20"}`}
        title="Share"
      >
        <Share2 size={16} />
      </button>

      {isOpen && (
        <div
          ref={menuRef}
          className="fixed z-[1000] w-[260px] bg-white dark:bg-surface-900 rounded-xl shadow-xl ring-1 ring-black/5 dark:ring-white/5 p-1.5 animate-in fade-in zoom-in-95 duration-150 origin-top-left"
          style={{
            top: menuPosition.top,
            left: menuPosition.left,
            transformOrigin: menuPosition.alignRight ? "top right" : "top left",
          }}
        >
          <div className="flex flex-col gap-0.5">
            {isSemble ? (
              <>
                <div className="px-3 py-2 text-[11px] font-bold text-surface-400 dark:text-surface-500 uppercase tracking-wider flex items-center gap-1.5 select-none">
                  <SembleLogo />
                  Semble Integration
                </div>
                {renderMenuItem(
                  "Open on Semble",
                  <ExternalLink size={16} />,
                  () => window.open(sembleUrl, "_blank"),
                  false,
                  true,
                )}
                {renderMenuItem(
                  "Copy Semble Link",
                  <Copy size={16} />,
                  () => handleCopy(sembleUrl, "semble"),
                  copied === "semble",
                )}
                <div className="h-px bg-surface-100 dark:bg-surface-800 my-1 mx-2" />
              </>
            ) : null}

            {renderMenuItem(
              "Copy Link",
              <Copy size={16} />,
              () => handleCopy(shareUrl, "link"),
              copied === "link",
            )}

            <div className="px-3 pt-3 pb-1 text-[11px] font-bold text-surface-400 dark:text-surface-500 uppercase tracking-wider select-none">
              Share via App
            </div>

            <div className="grid grid-cols-5 gap-1 px-1 mb-1">
              {shareForks.map((fork) => (
                <button
                  key={fork.domain}
                  onClick={() => handleShareToFork(fork.domain)}
                  className="flex items-center justify-center p-2 rounded-lg hover:bg-surface-50 dark:hover:bg-surface-800 hover:scale-105 transition-all text-surface-400 dark:text-surface-500 hover:text-surface-900 dark:hover:text-white"
                  title={`Share to ${fork.name}`}
                >
                  {fork.icon}
                </button>
              ))}
            </div>

            <div className="h-px bg-surface-100 dark:bg-surface-800 my-1 mx-2" />

            {renderMenuItem(
              "Copy Universal Link",
              <AturiIcon size={16} />,
              () =>
                handleCopy(uri.replace("at://", "https://aturi.to/"), "aturi"),
              copied === "aturi",
            )}

            {typeof navigator !== "undefined" &&
              navigator.share &&
              renderMenuItem(
                "More Options...",
                <MoreHorizontal size={16} />,
                () => {
                  navigator
                    .share({ title: "Margin", text, url: shareUrl })
                    .catch(() => {});
                  setIsOpen(false);
                },
              )}
          </div>
        </div>
      )}
    </div>
  );
}
