import React, { useState, useEffect, useRef } from "react";
import { useSearchParams, Navigate } from "react-router-dom";
import { AtSign } from "lucide-react";
import SignUpModal from "../../components/modals/SignUpModal";
import {
  searchActors,
  startLogin,
  type ActorSearchItem,
} from "../../api/client";
import { Avatar } from "../../components/ui";
import { useStore } from "@nanostores/react";
import { $theme } from "../../store/theme";
import { $user } from "../../store/auth";

export default function Login() {
  useStore($theme); // ensure theme is applied on this page
  const user = useStore($user);
  const [searchParams] = useSearchParams();
  const [handle, setHandle] = useState("");
  const [suggestions, setSuggestions] = useState<ActorSearchItem[]>([]);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(
    searchParams.get("error") || null,
  );
  const [selectedIndex, setSelectedIndex] = useState(-1);
  const [showSignUp, setShowSignUp] = useState(false);

  const inputRef = useRef<HTMLInputElement>(null);
  const suggestionsRef = useRef<HTMLDivElement>(null);
  const isSelectionRef = useRef(false);

  const [providerIndex, setProviderIndex] = useState(0);
  const [morphClass, setMorphClass] = useState(
    "opacity-100 translate-y-0 blur-0",
  );
  const providers = [
    "AT Protocol",
    "Margin",
    "Bluesky",
    "Blacksky",
    "Tangled",
    "Northsky",
    "witchcraft.systems",
    "tophhie.social",
    "altq.net",
  ];

  const [selectedAvatar, setSelectedAvatar] = useState<string | null>(null);

  useEffect(() => {
    const cycleText = () => {
      setMorphClass("opacity-0 translate-y-2 blur-sm");
      setTimeout(() => {
        setProviderIndex((prev) => (prev + 1) % providers.length);
        setMorphClass("opacity-100 translate-y-0 blur-0");
      }, 400);
    };
    const interval = setInterval(cycleText, 3000);
    return () => clearInterval(interval);
  }, [providers.length]);

  useEffect(() => {
    if (handle.length >= 3) {
      if (isSelectionRef.current) {
        isSelectionRef.current = false;
        return;
      }
      const timer = setTimeout(async () => {
        try {
          if (!handle.includes(".")) {
            const data = await searchActors(handle);
            setSuggestions(data.actors || []);

            const exactMatch = data.actors?.find((s) => s.handle === handle);
            if (exactMatch) {
              setSelectedAvatar(exactMatch.avatar || null);
            }

            setShowSuggestions(true);
            setSelectedIndex(-1);
          }
        } catch (e) {
          console.error("Search failed:", e);
        }
      }, 300);
      return () => clearTimeout(timer);
    }
  }, [handle]);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (
        suggestionsRef.current &&
        !suggestionsRef.current.contains(e.target as Node) &&
        inputRef.current &&
        !inputRef.current.contains(e.target as Node)
      ) {
        setShowSuggestions(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (!showSuggestions || suggestions.length === 0) return;

    if (e.key === "ArrowDown") {
      e.preventDefault();
      setSelectedIndex((prev) => Math.min(prev + 1, suggestions.length - 1));
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setSelectedIndex((prev) => Math.max(prev - 1, -1));
    } else if (e.key === "Enter" && selectedIndex >= 0) {
      e.preventDefault();
      selectSuggestion(suggestions[selectedIndex]);
    } else if (e.key === "Escape") {
      setShowSuggestions(false);
    }
  };

  const selectSuggestion = (actor: ActorSearchItem) => {
    isSelectionRef.current = true;
    setHandle(actor.handle);
    setSelectedAvatar(actor.avatar || null);
    setSuggestions([]);
    setShowSuggestions(false);
    inputRef.current?.blur();
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!handle.trim()) return;

    setLoading(true);
    setError(null);

    try {
      const result = await startLogin(handle.trim());
      if (result.authorizationUrl) {
        window.location.href = result.authorizationUrl;
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : "Unknown error";
      setError(message || "Failed to initiate login. Please try again.");
      setLoading(false);
    }
  };

  if (user) {
    return <Navigate to="/home" replace />;
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-surface-100 dark:bg-surface-800 p-4">
      <div className="w-full max-w-[440px] bg-white dark:bg-surface-900 rounded-2xl border border-surface-200/60 dark:border-surface-800 p-8 shadow-sm dark:shadow-none">
        <div className="flex flex-col items-center mb-8">
          <h1 className="text-2xl font-bold font-display text-surface-900 dark:text-white text-center leading-snug">
            Sign in with your <br />
            <span
              className={`inline-block transition-all duration-400 ease-out text-transparent bg-clip-text bg-gradient-to-r from-[#027bff] to-[#0285FF] ${morphClass}`}
            >
              {providers[providerIndex]}
            </span>{" "}
            handle
          </h1>
        </div>

        <form onSubmit={handleSubmit} className="w-full flex flex-col gap-4">
          <div className="relative group">
            <div className="absolute left-4 top-1/2 -translate-y-1/2 text-surface-400 dark:text-surface-500 transition-colors pointer-events-none">
              {selectedAvatar ? (
                <Avatar
                  src={selectedAvatar}
                  size="xs"
                  className="ring-2 ring-white dark:ring-surface-900 shadow-sm"
                />
              ) : (
                <AtSign
                  size={20}
                  className="stroke-[2.5] group-focus-within:text-[#027bff]"
                />
              )}
            </div>
            <input
              ref={inputRef}
              type="text"
              value={handle}
              onChange={(e) => {
                const val = e.target.value;
                setHandle(val);
                if (selectedAvatar) setSelectedAvatar(null);
                if (val.length < 3) {
                  setSuggestions([]);
                  setShowSuggestions(false);
                }
              }}
              onKeyDown={handleKeyDown}
              onFocus={() =>
                handle.length >= 3 &&
                suggestions.length > 0 &&
                !handle.includes(".") &&
                setShowSuggestions(true)
              }
              placeholder="handle.margin.cafe"
              className="w-full pl-12 pr-4 py-3.5 bg-surface-50 dark:bg-surface-950 border border-surface-200 dark:border-surface-700 rounded-xl focus:border-[#027bff] dark:focus:border-[#027bff] outline-none focus:ring-4 focus:ring-[#027bff]/10 transition-all font-medium text-lg text-surface-900 dark:text-white placeholder:text-surface-400 dark:placeholder:text-surface-500"
              autoCapitalize="none"
              autoCorrect="off"
              autoComplete="off"
              spellCheck={false}
              disabled={loading}
            />

            {showSuggestions && suggestions.length > 0 && (
              <div
                ref={suggestionsRef}
                className="absolute top-[calc(100%+8px)] left-0 right-0 bg-white dark:bg-surface-900 border border-surface-200 dark:border-surface-700 rounded-xl shadow-xl overflow-hidden z-50 animate-fade-in max-h-[300px] overflow-y-auto"
              >
                {suggestions.map((actor, index) => (
                  <button
                    key={actor.did}
                    type="button"
                    className={`w-full flex items-center gap-3 px-4 py-3 border-b border-surface-100 dark:border-surface-800 last:border-0 hover:bg-surface-50 dark:hover:bg-surface-800 transition-colors text-left ${index === selectedIndex ? "bg-surface-50 dark:bg-surface-800" : ""}`}
                    onClick={() => selectSuggestion(actor)}
                  >
                    <Avatar src={actor.avatar} size="sm" />
                    <div className="min-w-0">
                      <div className="font-semibold text-surface-900 dark:text-white truncate text-sm">
                        {actor.displayName || actor.handle}
                      </div>
                      <div className="text-surface-500 dark:text-surface-400 text-xs truncate">
                        @{actor.handle}
                      </div>
                    </div>
                  </button>
                ))}
              </div>
            )}
          </div>

          {error && (
            <div className="p-3 bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 text-sm rounded-lg border border-red-100 dark:border-red-800 text-center font-medium animate-fade-in">
              {error}
            </div>
          )}

          <button
            type="submit"
            disabled={loading || !handle}
            className="w-full py-3.5 bg-[#027bff] hover:bg-[#0269d9] active:scale-[0.98] text-white rounded-xl font-bold text-lg shadow-md shadow-[#027bff]/20 focus:outline-none focus:ring-4 focus:ring-[#027bff]/20 disabled:opacity-50 disabled:cursor-not-allowed transition-all flex items-center justify-center gap-2 mt-2"
          >
            {loading ? "Connecting..." : "Continue"}
          </button>

          <p className="text-center text-sm text-surface-400 dark:text-surface-500 mt-2 leading-relaxed">
            By signing in, you agree to our{" "}
            <a
              href="/terms"
              className="text-surface-900 dark:text-white hover:underline font-medium hover:text-[#027bff] dark:hover:text-[#027bff] transition-colors"
            >
              Terms of Service
            </a>{" "}
            and{" "}
            <a
              href="/privacy"
              className="text-surface-900 dark:text-white hover:underline font-medium hover:text-[#027bff] dark:hover:text-[#027bff] transition-colors"
            >
              Privacy Policy
            </a>
          </p>

          <div className="flex items-center gap-4 py-2 opacity-50">
            <div className="h-px bg-surface-200 dark:bg-surface-700 flex-1" />
            <span className="text-xs font-bold text-surface-400 dark:text-surface-500 uppercase tracking-wider">
              or
            </span>
            <div className="h-px bg-surface-200 dark:bg-surface-700 flex-1" />
          </div>

          <button
            type="button"
            onClick={() => setShowSignUp(true)}
            className="w-full py-3.5 bg-transparent border border-surface-200 dark:border-surface-700 hover:bg-surface-50 dark:hover:bg-surface-800 text-surface-600 dark:text-surface-300 hover:text-surface-900 dark:hover:text-white rounded-xl font-bold transition-all text-sm"
          >
            Create New Account
          </button>
        </form>
      </div>

      {showSignUp && <SignUpModal onClose={() => setShowSignUp(false)} />}
    </div>
  );
}
