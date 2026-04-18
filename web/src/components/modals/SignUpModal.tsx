import React, { useState, useEffect, useMemo } from "react";
import { X, ChevronRight, Loader2, AlertCircle } from "lucide-react";
import {
  BlackskyIcon,
  NorthskyIcon,
  BlueskyIcon,
  TophhieIcon,
  MarginIcon,
} from "../common/Icons";
import { startSignup } from "../../api/client";
import { analytics } from "../../lib/analytics";

interface Provider {
  id: string;
  name: string;
  service: string;
  Icon: React.ComponentType<{ size?: number }> | null;
  description: string;
  custom?: boolean;
  wide?: boolean;
}

const MARGIN_PROVIDER: Provider = {
  id: "margin",
  name: "Margin",
  service: "https://margin.cafe",
  Icon: MarginIcon,
  description: "The easiest way to get started",
};

const OTHER_PROVIDERS: Provider[] = [
  {
    id: "bluesky",
    name: "Bluesky",
    service: "https://bsky.social",
    Icon: BlueskyIcon,
    description: "The largest and most popular community",
  },
  {
    id: "blacksky",
    name: "Blacksky",
    service: "https://blacksky.app",
    Icon: BlackskyIcon,
    description: "For the Culture — a safe space for users and allies",
  },
    {
    id: "eurosky",
    name: "Eurosky",
    service: "https://eurosky.social",
    Icon: null,
    description: "Eurosky is your European home on the Atmosphere",
  },
  {
    id: "selfhosted.social",
    name: "selfhosted.social",
    service: "https://selfhosted.social",
    Icon: null,
    description: "A home for builders, tinkerers, and the curious",
  },
  {
    id: "northsky",
    name: "Northsky",
    service: "https://northsky.social",
    Icon: NorthskyIcon,
    description: "A Canadian worker-owned cooperative",
  },
  {
    id: "tophhie",
    name: "Tophhie",
    service: "https://tophhie.social",
    Icon: TophhieIcon,
    description: "A welcoming and friendly community",
  },
  {
    id: "custom",
    name: "Use a custom PDS",
    service: "",
    custom: true,
    Icon: null,
    description: "Already have a PDS? Enter its address.",
  },
];

function shuffleArray<T>(arr: T[]): T[] {
  const shuffled = [...arr];
  for (let i = shuffled.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
  }
  return shuffled;
}

const inviteStatusPromise: Promise<Record<string, boolean>> = (async () => {
  const results: Record<string, boolean> = {};
  await Promise.allSettled(
    [MARGIN_PROVIDER, ...OTHER_PROVIDERS]
      .filter((p) => p.service && !p.custom)
      .map(async (p) => {
        try {
          const res = await fetch(
            `${p.service}/xrpc/com.atproto.server.describeServer`,
          );
          if (res.ok) {
            const data = await res.json();
            results[p.id] = !!data.inviteCodeRequired;
          }
        } catch {
          // ignore unreachable providers
        }
      }),
  );
  return results;
})();

interface SignUpModalProps {
  onClose: () => void;
}

export default function SignUpModal({ onClose }: SignUpModalProps) {
  const [showCustomInput, setShowCustomInput] = useState(false);
  const [customService, setCustomService] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [inviteStatus, setInviteStatus] = useState<Record<string, boolean>>({});
  const [statusLoaded, setStatusLoaded] = useState(false);

  useEffect(() => {
    inviteStatusPromise.then((status) => {
      setInviteStatus(status);
      setStatusLoaded(true);
    });
  }, []);

  const providers = useMemo(() => {
    const nonCustom = OTHER_PROVIDERS.filter((p) => !p.custom);
    const custom = OTHER_PROVIDERS.find((p) => p.custom);

    if (!statusLoaded) {
      return [
        MARGIN_PROVIDER,
        ...shuffleArray(nonCustom),
        ...(custom ? [custom] : []),
      ];
    }

    const open = nonCustom.filter((p) => !inviteStatus[p.id]);
    const inviteOnly = nonCustom.filter((p) => inviteStatus[p.id]);
    return [
      MARGIN_PROVIDER,
      ...shuffleArray(open),
      ...shuffleArray(inviteOnly),
      ...(custom ? [custom] : []),
    ];
  }, [statusLoaded, inviteStatus]);

  useEffect(() => {
    document.body.style.overflow = "hidden";
    return () => {
      document.body.style.overflow = "unset";
    };
  }, []);

  const handleProviderSelect = async (provider: Provider) => {
    if (provider.custom) {
      setShowCustomInput(true);
      return;
    }

    setLoading(true);
    setError(null);

    try {
      analytics.capture("signup_initiated", { provider: provider.id });
      const result = await startSignup(provider.service);
      if (result.authorizationUrl) {
        window.location.assign(result.authorizationUrl);
      }
    } catch (err) {
      console.error(err);
      analytics.captureException(err);
      setError("Could not connect to this provider. Please try again.");
      setLoading(false);
    }
  };

  const handleCustomSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!customService.trim()) return;

    setLoading(true);
    setError(null);

    let serviceUrl = customService.trim();
    if (!serviceUrl.startsWith("http")) {
      serviceUrl = `https://${serviceUrl}`;
    }

    try {
      analytics.capture("signup_initiated", { provider: "custom" });
      const result = await startSignup(serviceUrl);
      if (result.authorizationUrl) {
        const url = new URL(result.authorizationUrl);
        if (url.protocol !== "https:")
          throw new Error("Invalid authorization URL");
        window.location.href = result.authorizationUrl;
      }
    } catch (err) {
      console.error(err);
      analytics.captureException(err);
      setError("Couldn't connect to that PDS. Double-check the address.");
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-[100] flex items-end sm:items-center justify-center sm:p-4 bg-black/60 backdrop-blur-sm animate-fade-in">
      <div className="w-full sm:max-w-md bg-white dark:bg-surface-900 rounded-t-3xl sm:rounded-3xl shadow-2xl overflow-hidden animate-slide-up max-h-[90vh] sm:max-h-[85vh] flex flex-col">
        <div className="p-3 sm:p-4 flex justify-end flex-shrink-0">
          <button
            onClick={onClose}
            className="p-2 text-surface-400 dark:text-surface-500 hover:text-surface-900 dark:hover:text-white hover:bg-surface-50 dark:hover:bg-surface-800 rounded-full transition-colors"
          >
            <X size={20} />
          </button>
        </div>

        <div className="px-5 sm:px-8 pb-8 sm:pb-10 overflow-y-auto">
          {loading ? (
            <div className="text-center py-10">
              <Loader2
                size={40}
                className="animate-spin text-primary-600 dark:text-primary-400 mx-auto mb-4"
              />
              <p className="text-surface-600 dark:text-surface-400 font-medium">
                Connecting...
              </p>
            </div>
          ) : showCustomInput ? (
            <div>
              <h2 className="text-2xl font-display font-bold text-surface-900 dark:text-white mb-2">
                Use a custom PDS
              </h2>
              <p className="text-sm text-surface-500 dark:text-surface-400 mb-6">
                Enter the address of the PDS hosting your account.
              </p>
              <form onSubmit={handleCustomSubmit} className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-surface-700 dark:text-surface-300 mb-1">
                    PDS address
                  </label>
                  <input
                    type="text"
                    className="w-full px-4 py-3 bg-surface-50 dark:bg-surface-800 border border-surface-200 dark:border-surface-700 rounded-xl text-surface-900 dark:text-white placeholder:text-surface-400 dark:placeholder:text-surface-500 focus:border-primary-500 dark:focus:border-primary-400 focus:ring-4 focus:ring-primary-500/10 dark:focus:ring-primary-400/10 outline-none transition-all"
                    value={customService}
                    onChange={(e) => setCustomService(e.target.value)}
                    placeholder="pds.example.com"
                    autoFocus
                  />
                </div>

                {error && (
                  <div className="p-3 bg-red-50 dark:bg-red-950/40 text-red-600 dark:text-red-400 text-sm rounded-lg flex items-center gap-2 border border-red-100 dark:border-red-900/40">
                    <AlertCircle size={16} />
                    {error}
                  </div>
                )}

                <div className="flex gap-3 pt-4">
                  <button
                    type="button"
                    className="flex-1 py-3 bg-white dark:bg-surface-800 border border-surface-200 dark:border-surface-700 text-surface-700 dark:text-surface-300 font-semibold rounded-xl hover:bg-surface-50 dark:hover:bg-surface-700 transition-colors"
                    onClick={() => {
                      setShowCustomInput(false);
                      setError(null);
                    }}
                  >
                    Back
                  </button>
                  <button
                    type="submit"
                    className="flex-1 py-3 bg-primary-600 dark:bg-primary-500 text-white font-semibold rounded-xl hover:bg-primary-700 dark:hover:bg-primary-400 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                    disabled={!customService.trim()}
                  >
                    Continue
                  </button>
                </div>
              </form>
            </div>
          ) : (
            <div>
              <h2 className="text-2xl font-display font-bold text-surface-900 dark:text-white mb-2">
                Create your account
              </h2>
              <p className="text-surface-500 dark:text-surface-400 mb-6">
                Margin adheres to the{" "}
                <a
                  href="https://atproto.com"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-primary-600 dark:text-primary-400 hover:underline"
                >
                  AT Protocol
                </a>
                . Choose a provider to host your account.
              </p>

              {error && (
                <div className="mb-4 p-3 bg-red-50 dark:bg-red-950/40 text-red-600 dark:text-red-400 text-sm rounded-lg flex items-center gap-2 border border-red-100 dark:border-red-900/40">
                  <AlertCircle size={16} />
                  {error}
                </div>
              )}

              <div className="space-y-2">
                {providers.map((p) => (
                  <button
                    key={p.id}
                    className={`w-full flex items-center gap-3 p-3 rounded-xl transition-all text-left group ${
                      p.id === "margin"
                        ? "bg-primary-50/80 dark:bg-primary-900/20 border border-primary-200/60 dark:border-primary-800/40 hover:border-primary-300 dark:hover:border-primary-700"
                        : "bg-surface-50 dark:bg-surface-800/60 hover:bg-surface-100 dark:hover:bg-surface-800 border border-transparent"
                    }`}
                    onClick={() => handleProviderSelect(p)}
                  >
                    <div
                      className={`w-9 h-9 flex items-center justify-center rounded-full flex-shrink-0 ${
                        p.id === "margin"
                          ? "bg-primary-100 dark:bg-primary-900/40 text-primary-600 dark:text-primary-400"
                          : "bg-white dark:bg-surface-700 shadow-sm dark:shadow-none text-surface-600 dark:text-surface-300"
                      }`}
                    >
                      {p.Icon ? (
                        <p.Icon size={18} />
                      ) : (
                        <span className="font-bold text-xs">{p.name[0]}</span>
                      )}
                    </div>
                    <div className="flex-1 min-w-0">
                      <h3 className="text-sm font-bold text-surface-900 dark:text-white">
                        {p.name}
                      </h3>
                      <p className="text-xs text-surface-500 dark:text-surface-400 line-clamp-1">
                        {p.description}
                      </p>
                    </div>
                    {inviteStatus[p.id] && (
                      <span className="text-[10px] font-medium text-surface-400 dark:text-surface-500 bg-surface-100 dark:bg-surface-800 px-1.5 py-0.5 rounded-md flex-shrink-0">
                        Invite
                      </span>
                    )}
                    <ChevronRight
                      size={16}
                      className="text-surface-300 dark:text-surface-600 group-hover:text-surface-600 dark:group-hover:text-surface-400"
                    />
                  </button>
                ))}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
