import { onMessage } from '@/utils/messaging';
import type { Annotation } from '@/utils/types';
import {
  checkSession,
  getAnnotations,
  createAnnotation,
  createBookmark,
  createHighlight,
  deleteHighlight,
  getUserBookmarks,
  getUserHighlights,
  getUserCollections,
  addToCollection,
  getItemCollections,
  getReplies,
  createReply,
  getUserTags,
  getTrendingTags,
} from '@/utils/api';
import { overlayEnabledItem, apiUrlItem } from '@/utils/storage';

export default defineBackground(() => {
  console.log('Margin extension loaded');

  function getPDFViewerURL(originalUrl: string): string {
    const viewerBase = browser.runtime.getURL('/pdfjs/web/viewer.html' as any);
    try {
      const parsed = new URL(originalUrl);
      const hash = parsed.hash;
      parsed.hash = '';
      return `${viewerBase}?file=${encodeURIComponent(parsed.href)}${hash}`;
    } catch {
      return `${viewerBase}?file=${encodeURIComponent(originalUrl)}`;
    }
  }

  function resolveTabUrl(tabUrl: string): string {
    if (tabUrl.includes('/pdfjs/web/viewer.html')) {
      try {
        const fileParam = new URL(tabUrl).searchParams.get('file');
        if (fileParam) return fileParam;
      } catch {
        /* ignore */
      }
    }
    return tabUrl;
  }

  const annotationCache = new Map<string, { annotations: Annotation[]; timestamp: number }>();
  const CACHE_TTL = 60000;

  onMessage('checkSession', async () => {
    return await checkSession();
  });

  onMessage('getAnnotations', async ({ data }) => {
    return await getAnnotations(data.url, [], data.cacheBust);
  });

  onMessage('activateOnPdf', async ({ data }) => {
    const { tabId, url } = data;
    const viewerUrl = getPDFViewerURL(url);
    await browser.tabs.update(tabId, { url: viewerUrl });
    return { redirected: true };
  });

  onMessage('createAnnotation', async ({ data }) => {
    return await createAnnotation(data);
  });

  onMessage('createBookmark', async ({ data }) => {
    return await createBookmark(data);
  });

  onMessage('createHighlight', async ({ data }) => {
    return await createHighlight(data);
  });

  onMessage('deleteHighlight', async ({ data }) => {
    return await deleteHighlight(data.uri);
  });

  onMessage('convertHighlightToAnnotation', async ({ data }) => {
    const createResult = await createAnnotation({
      url: data.url,
      text: data.text,
      title: data.title,
      selector: data.selector,
    });

    if (!createResult.success) {
      return { success: false, error: createResult.error || 'Failed to create annotation' };
    }

    const deleteResult = await deleteHighlight(data.highlightUri);
    if (!deleteResult.success) {
      console.warn('Created annotation but failed to delete highlight:', deleteResult.error);
    }

    return { success: true };
  });

  onMessage('getUserBookmarks', async ({ data }) => {
    return await getUserBookmarks(data.did);
  });

  onMessage('getUserHighlights', async ({ data }) => {
    return await getUserHighlights(data.did);
  });

  onMessage('getUserCollections', async ({ data }) => {
    return await getUserCollections(data.did);
  });

  onMessage('addToCollection', async ({ data }) => {
    return await addToCollection(data.collectionUri, data.annotationUri);
  });

  onMessage('getItemCollections', async ({ data }) => {
    return await getItemCollections(data.annotationUri);
  });

  onMessage('getReplies', async ({ data }) => {
    return await getReplies(data.uri);
  });

  onMessage('createReply', async ({ data }) => {
    return await createReply(data);
  });

  onMessage('getOverlayEnabled', async () => {
    return await overlayEnabledItem.getValue();
  });

  onMessage('getUserTags', async ({ data }) => {
    return await getUserTags(data.did);
  });

  onMessage('getTrendingTags', async () => {
    return await getTrendingTags();
  });

  onMessage('openAppUrl', async ({ data }) => {
    const apiUrl = await apiUrlItem.getValue();
    await browser.tabs.create({ url: `${apiUrl}${data.path}` });
  });

  onMessage('updateBadge', async ({ data }) => {
    const { count, tabId } = data;
    const text = count > 0 ? String(count > 99 ? '99+' : count) : '';

    if (tabId) {
      await browser.action.setBadgeText({ text, tabId });
      await browser.action.setBadgeBackgroundColor({ color: '#3b82f6', tabId });
    } else {
      const [tab] = await browser.tabs.query({ active: true, currentWindow: true });
      if (tab?.id) {
        await browser.action.setBadgeText({ text, tabId: tab.id });
        await browser.action.setBadgeBackgroundColor({ color: '#3b82f6', tabId: tab.id });
      }
    }
  });

  browser.tabs.onUpdated.addListener(async (tabId, changeInfo) => {
    if (changeInfo.status === 'loading' && changeInfo.url) {
      await browser.action.setBadgeText({ text: '', tabId });
    }
  });

  onMessage('cacheAnnotations', async ({ data }) => {
    const { url, annotations } = data;
    const normalizedUrl = normalizeUrl(url);
    annotationCache.set(normalizedUrl, { annotations, timestamp: Date.now() });
  });

  onMessage('getCachedAnnotations', async ({ data }) => {
    const normalizedUrl = normalizeUrl(data.url);
    const cached = annotationCache.get(normalizedUrl);
    if (cached && Date.now() - cached.timestamp < CACHE_TTL) {
      return cached.annotations;
    }
    return null;
  });

  function normalizeUrl(url: string): string {
    try {
      const u = new URL(url);
      u.hash = '';
      const path = u.pathname.replace(/\/$/, '') || '/';
      return `${u.origin}${path}${u.search}`;
    } catch {
      return url;
    }
  }

  async function ensureContextMenus() {
    await browser.contextMenus.removeAll();

    browser.contextMenus.create({
      id: 'margin-annotate',
      title: 'Annotate "%s"',
      contexts: ['selection'],
    });

    browser.contextMenus.create({
      id: 'margin-highlight',
      title: 'Highlight "%s"',
      contexts: ['selection'],
    });

    browser.contextMenus.create({
      id: 'margin-bookmark',
      title: 'Bookmark this page',
      contexts: ['page'],
    });

    browser.contextMenus.create({
      id: 'margin-open-sidebar',
      title: 'Open Margin Sidebar',
      contexts: ['page', 'selection', 'link'],
    });
  }

  browser.runtime.onInstalled.addListener(async () => {
    await ensureContextMenus();
  });

  browser.runtime.onStartup.addListener(async () => {
    await ensureContextMenus();
  });

  browser.contextMenus.onClicked.addListener((info, tab) => {
    if (info.menuItemId === 'margin-open-sidebar') {
      const browserAny = browser as any;
      if (browserAny.sidePanel && tab?.windowId) {
        browserAny.sidePanel.open({ windowId: tab.windowId }).catch((err: Error) => {
          console.error('Could not open side panel:', err);
        });
      } else if (browserAny.sidebarAction) {
        browserAny.sidebarAction.open().catch((err: Error) => {
          console.warn('Could not open Firefox sidebar:', err);
        });
      }
      return;
    }

    handleContextMenuAction(info, tab);
  });

  async function handleContextMenuAction(info: any, tab?: any) {
    const apiUrl = await apiUrlItem.getValue();

    if (info.menuItemId === 'margin-bookmark' && tab?.url) {
      const session = await checkSession();
      if (!session.authenticated) {
        await browser.tabs.create({ url: `${apiUrl}/login` });
        return;
      }

      const result = await createBookmark({
        url: resolveTabUrl(tab.url),
        title: tab.title,
      });

      if (result.success) {
        showNotification('Margin', 'Page bookmarked!');
      }
      return;
    }

    if (info.menuItemId === 'margin-annotate' && tab?.url && info.selectionText) {
      const session = await checkSession();
      if (!session.authenticated) {
        await browser.tabs.create({ url: `${apiUrl}/login` });
        return;
      }

      try {
        await browser.tabs.sendMessage(tab.id!, {
          type: 'SHOW_INLINE_ANNOTATE',
          data: {
            url: resolveTabUrl(tab.url),
            title: tab.title,
            selector: {
              type: 'TextQuoteSelector',
              exact: info.selectionText,
            },
          },
        });
      } catch {
        let composeUrl = `${apiUrl}/new?url=${encodeURIComponent(resolveTabUrl(tab.url))}`;
        composeUrl += `&selector=${encodeURIComponent(
          JSON.stringify({
            type: 'TextQuoteSelector',
            exact: info.selectionText,
          })
        )}`;
        await browser.tabs.create({ url: composeUrl });
      }
      return;
    }

    if (info.menuItemId === 'margin-highlight' && tab?.url && info.selectionText) {
      const session = await checkSession();
      if (!session.authenticated) {
        await browser.tabs.create({ url: `${apiUrl}/login` });
        return;
      }

      const result = await createHighlight({
        url: resolveTabUrl(tab.url),
        title: tab.title,
        selector: {
          type: 'TextQuoteSelector',
          exact: info.selectionText,
        },
      });

      if (result.success) {
        showNotification('Margin', 'Text highlighted!');
        try {
          await browser.tabs.sendMessage(tab.id!, { type: 'REFRESH_ANNOTATIONS' });
        } catch {
          /* ignore */
        }
      }
      return;
    }
  }

  function showNotification(title: string, message: string) {
    const browserAny = browser as any;
    if (browserAny.notifications) {
      browserAny.notifications.create({
        type: 'basic',
        iconUrl: '/icons/icon-128.png',
        title,
        message,
      });
    }
  }

  let sidePanelOpen = false;

  browser.runtime.onConnect.addListener((port) => {
    if (port.name === 'sidepanel') {
      sidePanelOpen = true;
      port.onDisconnect.addListener(() => {
        sidePanelOpen = false;
      });
    }
  });

  browser.commands?.onCommand.addListener((command) => {
    if (command === 'toggle-sidebar') {
      const browserAny = browser as any;
      if (browserAny.sidePanel) {
        chrome.windows.getCurrent((win) => {
          if (win?.id) {
            if (sidePanelOpen && typeof browserAny.sidePanel.close === 'function') {
              browserAny.sidePanel.close({ windowId: win.id }).catch((err: Error) => {
                console.error('Could not close side panel:', err);
              });
            } else {
              browserAny.sidePanel.open({ windowId: win.id }).catch((err: Error) => {
                console.error('Could not open side panel:', err);
              });
            }
          }
        });
      } else if (browserAny.sidebarAction) {
        browserAny.sidebarAction.toggle().catch((err: Error) => {
          console.warn('Could not toggle Firefox sidebar:', err);
        });
      }
      return;
    }

    handleCommandAsync(command);
  });

  async function handleCommandAsync(command: string) {
    const [tab] = await browser.tabs.query({ active: true, currentWindow: true });

    if (command === 'toggle-overlay') {
      const current = await overlayEnabledItem.getValue();
      await overlayEnabledItem.setValue(!current);
      return;
    }

    if (command === 'bookmark-page' && tab?.url) {
      const session = await checkSession();
      if (!session.authenticated) {
        const apiUrl = await apiUrlItem.getValue();
        await browser.tabs.create({ url: `${apiUrl}/login` });
        return;
      }

      const result = await createBookmark({
        url: resolveTabUrl(tab.url),
        title: tab.title,
      });

      if (result.success) {
        showNotification('Margin', 'Page bookmarked!');
      }
      return;
    }

    if ((command === 'annotate-selection' || command === 'highlight-selection') && tab?.id) {
      try {
        const selection = (await browser.tabs.sendMessage(tab.id, { type: 'GET_SELECTION' })) as
          | { text?: string }
          | undefined;
        if (!selection?.text) return;

        const session = await checkSession();
        if (!session.authenticated) {
          const apiUrl = await apiUrlItem.getValue();
          await browser.tabs.create({ url: `${apiUrl}/login` });
          return;
        }

        if (command === 'annotate-selection') {
          await browser.tabs.sendMessage(tab.id, {
            type: 'SHOW_INLINE_ANNOTATE',
            data: { selector: { exact: selection.text } },
          });
        } else if (command === 'highlight-selection') {
          const result = await createHighlight({
            url: resolveTabUrl(tab.url!),
            title: tab.title,
            selector: {
              type: 'TextQuoteSelector',
              exact: selection.text,
            },
          });

          if (result.success) {
            showNotification('Margin', 'Text highlighted!');
            await browser.tabs.sendMessage(tab.id, { type: 'REFRESH_ANNOTATIONS' });
          }
        }
      } catch (err) {
        console.error('Error handling keyboard shortcut:', err);
      }
    }
  }
});
