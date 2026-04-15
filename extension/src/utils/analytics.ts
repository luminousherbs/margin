import { apiUrlItem } from './storage';

export type ExtensionEvents = {
  extension_installed: { version: string; browser: string };
  extension_updated: { previous_version: string; version: string };
  popup_opened: { authenticated: boolean };
  extension_tab_switched: { tab: string };
  annotation_created: { url: string; tag_count: number; source: 'extension' };
  highlight_created: { url: string; tag_count: number; has_color: boolean; source: 'extension' };
  bookmark_created: { url: string; tag_count: number; source: 'extension' };
  extension_connected: { did: string };
  api_key_created: Record<string, never>;
};

export async function capture<E extends keyof ExtensionEvents>(
  event: E,
  properties: ExtensionEvents[E],
  distinctId?: string
): Promise<void> {
  try {
    const apiUrl = await apiUrlItem.getValue();
    await fetch(`${apiUrl}/api/analytics/capture`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        event,
        distinct_id: distinctId ?? 'anonymous_extension',
        properties: {
          ...properties,
          $lib: 'margin-extension',
        },
      }),
      keepalive: true,
    });
  } catch {
    // ignore
  }
}
