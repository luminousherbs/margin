/// <reference types="astro/client" />

declare namespace App {
  interface Locals {
    user: import("./types").UserProfile | null;
  }
}

interface Window {
  __posthog_initialized?: boolean;
  posthog?: import("posthog-js").PostHog;
}
