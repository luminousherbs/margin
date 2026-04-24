/// <reference types="astro/client" />

declare module "virtual:i18n-languages" {
  export const languages: { code: string; name: string; nativeName: string }[];
}

declare namespace App {
  interface Locals {
    user: import("./types").UserProfile | null;
  }
}

interface Window {
  __posthog_initialized?: boolean;
  posthog?: import("posthog-js").PostHog;
}
