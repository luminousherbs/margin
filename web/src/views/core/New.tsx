import React, { useState } from "react";
import { useStore } from "@nanostores/react";
import { $user } from "../../store/auth";
import Composer from "../../components/feed/Composer";
import type { Selector } from "../../types";

interface NewAnnotationProps {
  initialUrl?: string;
  initialSelectorJson?: string;
  initialQuote?: string;
}

export default function NewAnnotationPage({
  initialUrl: propUrl,
  initialSelectorJson,
  initialQuote,
}: NewAnnotationProps) {
  const user = useStore($user);

  const initialUrl = propUrl || "";

  let initialSelector: Selector | null = null;
  if (initialSelectorJson) {
    try {
      initialSelector = JSON.parse(initialSelectorJson);
    } catch (e) {
      console.error("Failed to parse selector:", e);
    }
  }

  if (initialQuote && !initialSelector) {
    initialSelector = {
      type: "TextQuoteSelector",
      exact: initialQuote,
    };
  }

  const [url, setUrl] = useState(initialUrl);

  if (!user) {
    return (
      <div className="max-w-sm mx-auto py-16 px-4">
        <div className="bg-white dark:bg-surface-900 rounded-xl ring-1 ring-black/5 dark:ring-white/5 p-6 text-center">
          <h2 className="text-xl font-bold text-surface-900 dark:text-white mb-2">
            Sign in to create
          </h2>
          <p className="text-surface-500 dark:text-surface-400 text-sm mb-5">
            You need a Bluesky account
          </p>
          <a
            href="/login"
            className="block w-full py-2.5 px-4 bg-primary-600 hover:bg-primary-700 text-white font-medium rounded-lg transition-colors"
          >
            Sign in with Bluesky
          </a>
        </div>
      </div>
    );
  }

  const handleSuccess = () => {
    window.location.href = "/home";
  };

  return (
    <div className="max-w-2xl mx-auto pb-20">
      <div className="mb-6 text-center sm:text-left">
        <h1 className="text-2xl font-display font-bold text-surface-900 dark:text-white mb-1">
          New Annotation
        </h1>
        <p className="text-surface-500 dark:text-surface-400">
          Write in the margins of the web
        </p>
      </div>

      {!initialUrl && (
        <div className="mb-4">
          <label
            htmlFor="url-input"
            className="block text-sm font-medium text-surface-700 dark:text-surface-300 mb-1.5"
          >
            URL to annotate
          </label>
          <input
            id="url-input"
            type="url"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            placeholder="https://example.com/article"
            className="w-full p-3 bg-white dark:bg-surface-900 border border-surface-200 dark:border-surface-700 rounded-lg text-surface-900 dark:text-white placeholder:text-surface-400 dark:placeholder:text-surface-500 focus:ring-2 focus:ring-primary-500/20 focus:border-primary-500 dark:focus:border-primary-400 outline-none transition-all"
            required
          />
        </div>
      )}

      <div className="bg-white dark:bg-surface-900 rounded-xl ring-1 ring-black/5 dark:ring-white/5 p-5">
        <Composer
          url={
            (url || initialUrl) && !/^(?:f|ht)tps?:\/\//.test(url || initialUrl)
              ? `https://${url || initialUrl}`
              : url || initialUrl
          }
          selector={initialSelector}
          onSuccess={handleSuccess}
          onCancel={() => window.history.back()}
        />
      </div>
    </div>
  );
}
