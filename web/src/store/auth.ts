import { atom } from "nanostores";
import { checkSession } from "../api/client";
import { loadPreferences } from "./preferences";
import type { UserProfile } from "../types";

export const $user = atom<UserProfile | null>(null);
export const $isLoading = atom<boolean>(true);

export async function initAuth() {
  $isLoading.set(true);
  const session = await checkSession();
  $user.set(session);
  $isLoading.set(false);
  if (session) {
    loadPreferences();
  }
}

export function logout() {
  fetch("/auth/logout", { method: "POST" }).then(() => {
    window.location.href = "/";
  });
}
