import type { APIRoute } from "astro";
import satori from "satori";
import { Resvg } from "@resvg/resvg-js";
import { readFileSync } from "node:fs";
import { join } from "node:path";

function getPublicDir(): string {
  if (import.meta.env.PROD) {
    return join(process.cwd(), "dist", "client");
  }
  return join(process.cwd(), "public");
}

let fontsLoaded: { regular: ArrayBuffer; bold: ArrayBuffer } | null = null;

function loadFonts() {
  if (fontsLoaded) return fontsLoaded;
  const publicDir = getPublicDir();
  fontsLoaded = {
    regular: readFileSync(join(publicDir, "fonts", "Inter-Regular.ttf"))
      .buffer as ArrayBuffer,
    bold: readFileSync(join(publicDir, "fonts", "Inter-Bold.ttf"))
      .buffer as ArrayBuffer,
  };
  return fontsLoaded;
}

const API_URL = process.env.API_URL || "http://localhost:8081";

interface RecordData {
  type: "annotation" | "highlight" | "bookmark" | "collection";
  author: string;
  displayName: string;
  avatarURL: string;
  text: string;
  quote: string;
  source: string;
  title: string;
  icon: string;
  description: string;
  color: string;
}

async function resolveAvatarUrl(did: string): Promise<string> {
  if (!did) return "";
  try {
    const res = await fetch(
      `https://public.api.bsky.app/xrpc/app.bsky.actor.getProfile?actor=${encodeURIComponent(did)}`,
    );
    if (!res.ok) return "";
    const profile = await res.json();
    const avatar = profile.avatar || "";
    if (!avatar) return "";
    return (
      avatar.replace(/@[a-z]+$/, "@jpeg") +
      (/@[a-z]+$/.test(avatar) ? "" : "@jpeg")
    );
  } catch {
    return "";
  }
}

async function fetchAvatarDataUri(did: string): Promise<string> {
  const url = await resolveAvatarUrl(did);
  if (!url) return "";
  try {
    const res = await fetch(url, { headers: { "User-Agent": "margin.at/og" } });
    if (!res.ok) return "";
    const buf = await res.arrayBuffer();
    const mime = res.headers.get("content-type") || "image/jpeg";
    const b64 = Buffer.from(buf).toString("base64");
    return `data:${mime};base64,${b64}`;
  } catch {
    return "";
  }
}

async function fetchRecordData(uri: string): Promise<RecordData | null> {
  try {
    const res = await fetch(
      `${API_URL}/api/note?uri=${encodeURIComponent(uri)}`,
    );
    if (res.ok) {
      const item = await res.json();
      const author = item.author || item.creator || {};
      const handle = author.handle || "";
      const did = author.did || "";
      const authorName = handle ? `@${handle}` : did || "someone";
      const displayName = author.displayName || handle || did || "someone";
      const avatarURL = await fetchAvatarDataUri(author.did || "");
      const targetSource = item.target?.source || item.url || item.source || "";
      const domain = targetSource
        ? (() => {
            try {
              return new URL(targetSource).hostname.replace(/^www\./, "");
            } catch {
              return "";
            }
          })()
        : "";
      const selectorText =
        item.target?.selector?.exact || item.selector?.exact || "";
      const bodyText =
        extractBody(item.body) || item.bodyValue || item.text || "";
      const motivation = item.motivation || "";
      const targetTitle = item.target?.title || item.title || "";

      if (
        motivation === "highlighting" ||
        uri.includes("/at.margin.highlight/")
      ) {
        return {
          type: "highlight",
          author: authorName,
          displayName,
          avatarURL,
          text: targetTitle,
          quote: selectorText,
          source: domain,
          title: "",
          icon: "",
          description: "",
          color: item.color || "",
        };
      }

      if (uri.includes("/at.margin.bookmark/")) {
        return {
          type: "bookmark",
          author: authorName,
          displayName,
          avatarURL,
          text: item.title || targetTitle || "Bookmark",
          quote: item.description || bodyText || "",
          source: domain,
          title: "",
          icon: "",
          description: "",
          color: "",
        };
      }

      return {
        type: "annotation",
        author: authorName,
        displayName,
        avatarURL,
        text: bodyText,
        quote: selectorText,
        source: domain,
        title: "",
        icon: "",
        description: "",
        color: "",
      };
    }
  } catch {
    /* fall through */
  }

  try {
    const res = await fetch(
      `${API_URL}/api/collection?uri=${encodeURIComponent(uri)}`,
    );
    if (res.ok) {
      const item = await res.json();
      const author = item.author || item.creator || {};
      const handle = author.handle || "";
      const did = author.did || "";
      const authorName = handle ? `@${handle}` : did || "someone";
      const displayName = author.displayName || handle || did || "someone";
      const avatarURL = await fetchAvatarDataUri(author.did || "");

      return {
        type: "collection",
        author: authorName,
        displayName,
        avatarURL,
        text: "",
        quote: "",
        source: "",
        title: item.name || "Collection",
        icon: item.icon || "",
        description: item.description || "",
        color: "",
      };
    }
  } catch {
    /* fall through */
  }

  return null;
}

function truncate(str: string, max: number): string {
  if (str.length <= max) return str;
  return str.slice(0, max - 3) + "\u2026";
}

function extractBody(body: unknown): string {
  if (!body) return "";
  if (typeof body === "string") return body;
  if (typeof body === "object" && body !== null && "value" in body) {
    return String((body as { value: unknown }).value || "");
  }
  return "";
}

const C = {
  bg: "#18181b",
  bgCard: "#27272a",
  text: "#fafafa",
  textSecondary: "#a1a1aa",
  textMuted: "#71717a",
  textFaint: "#3f3f46",
  border: "#3f3f46",
};

const typeAccent: Record<string, string> = {
  annotation: "#3b82f6",
  highlight: "#facc15",
  bookmark: "#22c55e",
  collection: "#a78bfa",
};

const typeLabels: Record<string, string> = {
  annotation: "Annotation",
  highlight: "Highlight",
  bookmark: "Bookmark",
  collection: "Collection",
};

const namedColors: Record<string, string> = {
  yellow: "#facc15",
  green: "#4ade80",
  red: "#f87171",
  blue: "#60a5fa",
  purple: "#a78bfa",
  pink: "#f472b6",
  orange: "#fb923c",
};

function resolveHighlightColor(color: string): string {
  if (!color) return "#facc15";
  if (color.startsWith("#")) return color;
  return namedColors[color] || "#facc15";
}

function sanitizeColor(color: string): string {
  if (/^#[0-9a-fA-F]{3,8}$/.test(color)) return color;
  if (/^rgba?\([0-9,.\s]+\)$/.test(color)) return color;
  return "#888888";
}

function coloredLogoUri(color: string): string {
  const safe = sanitizeColor(color);
  const publicDir = getPublicDir();
  const svg = readFileSync(join(publicDir, "logo.svg"), "utf-8");
  const recolored = svg.replace(/fill="[^"]*"/, `fill="${safe}"`);
  return `data:image/svg+xml,${encodeURIComponent(recolored)}`;
}

function logoElement(height: number, color: string): unknown | null {
  const uri = coloredLogoUri(color);
  if (!uri) return null;
  const width = Math.round(height * (265 / 231));
  return {
    type: "img",
    props: {
      src: uri,
      width,
      height,
      style: { flexShrink: 0, opacity: 0.9 },
    },
  };
}

const lucideFileMap: Record<string, string> = {
  file: "file-text",
  pin: "map-pin",
  trending: "trending-up",
};

function lucideIconUri(iconName: string, color: string): string | null {
  try {
    const safe = sanitizeColor(color);
    const fileName = lucideFileMap[iconName] || iconName;
    const iconPath = join(
      process.cwd(),
      "node_modules",
      "lucide-react",
      "dist",
      "esm",
      "icons",
      `${fileName}.js`,
    );
    const src = readFileSync(iconPath, "utf-8");
    const paths = [...src.matchAll(/d:\s*"([^"]+)"/g)].map((m) => m[1]);
    if (!paths.length) return null;
    const pathEls = paths
      .map(
        (d) =>
          `<path d="${d}" fill="none" stroke="${safe}" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>`,
      )
      .join("");
    const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24">${pathEls}</svg>`;
    return `data:image/svg+xml,${encodeURIComponent(svg)}`;
  } catch {
    return null;
  }
}

function avatarElement(url: string, name: string, size: number): unknown {
  if (url) {
    return {
      type: "img",
      props: {
        src: url,
        width: size,
        height: size,
        style: {
          borderRadius: size / 2,
          flexShrink: 0,
          border: `2px solid ${C.border}`,
        },
      },
    };
  }
  const letter =
    name[0] === "@"
      ? (name[1] || "?").toUpperCase()
      : (name[0] || "?").toUpperCase();
  return {
    type: "div",
    props: {
      style: {
        width: size,
        height: size,
        borderRadius: size / 2,
        background: C.bgCard,
        border: `2px solid ${C.border}`,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        color: C.textMuted,
        fontSize: Math.round(size * 0.45),
        fontWeight: 700,
        flexShrink: 0,
      },
      children: letter,
    },
  };
}

function accentBar(color: string): unknown {
  return {
    type: "div",
    props: {
      style: {
        position: "absolute",
        left: 0,
        top: 0,
        bottom: 0,
        width: 5,
        background: color,
        borderRadius: "3px 0 0 3px",
      },
    },
  };
}

function bgPattern(): unknown {
  return {
    type: "div",
    props: {
      style: {
        position: "absolute",
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        backgroundImage:
          "radial-gradient(circle at 1px 1px, rgba(255,255,255,0.03) 1px, transparent 0)",
        backgroundSize: "32px 32px",
      },
    },
  };
}

function header(data: RecordData, accentColor: string): unknown {
  return {
    type: "div",
    props: {
      style: { display: "flex", alignItems: "center", width: "100%" },
      children: [
        avatarElement(data.avatarURL, data.displayName || data.author, 64),
        {
          type: "div",
          props: {
            style: {
              display: "flex",
              flexDirection: "column",
              marginLeft: 18,
              flex: 1,
            },
            children: [
              {
                type: "span",
                props: {
                  style: { color: C.text, fontSize: 28, fontWeight: 700 },
                  children: truncate(data.displayName || data.author, 30),
                },
              },
              {
                type: "span",
                props: {
                  style: {
                    color: C.textMuted,
                    fontSize: 20,
                    marginTop: 3,
                  },
                  children: data.author,
                },
              },
            ],
          },
        },
        {
          type: "div",
          props: {
            style: {
              display: "flex",
              alignItems: "center",
              gap: 10,
              background: "rgba(255,255,255,0.06)",
              borderRadius: 24,
              padding: "8px 18px 8px 14px",
            },
            children: [
              logoElement(22, accentColor),
              {
                type: "span",
                props: {
                  style: {
                    color: accentColor,
                    fontSize: 18,
                    fontWeight: 600,
                    letterSpacing: 0.5,
                  },
                  children: typeLabels[data.type] || data.type,
                },
              },
            ].filter(Boolean) as unknown[],
          },
        },
      ],
    },
  };
}

function footerEl(source: string): unknown | null {
  if (!source) return null;
  return {
    type: "div",
    props: {
      style: {
        display: "flex",
        alignItems: "center",
        marginTop: "auto",
        paddingTop: 24,
      },
      children: [
        {
          type: "div",
          props: {
            style: {
              width: 8,
              height: 8,
              borderRadius: 4,
              background: C.textFaint,
              marginRight: 12,
            },
          },
        },
        {
          type: "span",
          props: {
            style: { color: C.textMuted, fontSize: 20 },
            children: source,
          },
        },
      ],
    },
  };
}

function card(children: unknown[], accentColor: string): unknown {
  return {
    type: "div",
    props: {
      style: {
        display: "flex",
        width: "100%",
        height: "100%",
        background: C.bg,
        padding: 0,
        fontFamily: "Inter",
        position: "relative",
      },
      children: [
        bgPattern(),
        accentBar(accentColor),
        {
          type: "div",
          props: {
            style: {
              display: "flex",
              flexDirection: "column",
              width: "100%",
              height: "100%",
              padding: "48px 56px 48px 60px",
              position: "relative",
            },
            children,
          },
        },
      ],
    },
  };
}

function buildAnnotationImage(data: RecordData) {
  const accent = typeAccent.annotation;
  const children: unknown[] = [header(data, accent)];

  if (data.text) {
    const len = data.text.length;
    children.push({
      type: "div",
      props: {
        style: {
          color: C.text,
          fontSize: len > 200 ? 32 : len > 100 ? 38 : 46,
          fontWeight: 400,
          lineHeight: 1.5,
          marginTop: 36,
          overflow: "hidden",
        },
        children: truncate(data.text, 240),
      },
    });
  }

  if (data.quote) {
    children.push({
      type: "div",
      props: {
        style: {
          display: "flex",
          marginTop: 28,
          background: "rgba(59, 130, 246, 0.06)",
          borderRadius: 14,
          padding: "20px 24px",
        },
        children: [
          {
            type: "div",
            props: {
              style: {
                width: 4,
                borderRadius: 2,
                background: accent,
                flexShrink: 0,
                opacity: 0.7,
              },
            },
          },
          {
            type: "div",
            props: {
              style: {
                color: C.textSecondary,
                fontSize: 24,
                lineHeight: 1.6,
                paddingLeft: 18,
                fontStyle: "italic",
                overflow: "hidden",
              },
              children: truncate(data.quote, 180),
            },
          },
        ],
      },
    });
  }

  const f = footerEl(data.source);
  if (f) children.push(f);
  return card(children, accent);
}

function buildHighlightImage(data: RecordData) {
  const highlightColor = resolveHighlightColor(data.color);
  const quoteText = data.quote || data.text || "Highlighted passage";
  const len = quoteText.length;

  const children: unknown[] = [
    header(data, highlightColor),
    {
      type: "div",
      props: {
        style: {
          color: highlightColor,
          fontSize: 120,
          fontWeight: 700,
          lineHeight: 1,
          marginTop: 28,
          opacity: 0.5,
        },
        children: "\u201C",
      },
    },
    {
      type: "div",
      props: {
        style: {
          color: C.text,
          fontSize: len > 150 ? 32 : len > 80 ? 40 : 50,
          fontWeight: 600,
          lineHeight: 1.4,
          marginTop: -28,
          overflow: "hidden",
        },
        children: truncate(quoteText, 200),
      },
    },
  ];

  if (data.text && data.quote) {
    children.push({
      type: "div",
      props: {
        style: { color: C.textMuted, fontSize: 22, marginTop: 22 },
        children: truncate(data.text, 80),
      },
    });
  }

  const f = footerEl(data.source);
  if (f) children.push(f);
  return card(children, highlightColor);
}

function buildBookmarkImage(data: RecordData) {
  const accent = typeAccent.bookmark;
  const children: unknown[] = [header(data, accent)];

  if (data.source) {
    children.push({
      type: "div",
      props: {
        style: {
          display: "flex",
          alignItems: "center",
          marginTop: 36,
          gap: 10,
        },
        children: [
          {
            type: "div",
            props: {
              style: {
                width: 10,
                height: 10,
                borderRadius: 5,
                background: accent,
                opacity: 0.6,
              },
            },
          },
          {
            type: "span",
            props: {
              style: { color: accent, fontSize: 22, fontWeight: 500 },
              children: data.source,
            },
          },
        ],
      },
    });
  }

  const titleLen = (data.text || "").length;
  children.push({
    type: "div",
    props: {
      style: {
        color: C.text,
        fontSize: titleLen > 60 ? 40 : 50,
        fontWeight: 700,
        lineHeight: 1.3,
        marginTop: data.source ? 14 : 36,
        overflow: "hidden",
      },
      children: truncate(data.text || "Untitled Bookmark", 80),
    },
  });

  if (data.quote) {
    children.push({
      type: "div",
      props: {
        style: {
          color: C.textSecondary,
          fontSize: 26,
          lineHeight: 1.5,
          marginTop: 18,
          overflow: "hidden",
        },
        children: truncate(data.quote, 160),
      },
    });
  }

  return card(children, accent);
}

function buildCollectionImage(data: RecordData) {
  const accent = typeAccent.collection;
  const children: unknown[] = [header(data, accent)];

  const iconChildren: unknown[] = [];
  if (data.icon) {
    if (data.icon.startsWith("icon:")) {
      const iconName = data.icon.replace("icon:", "");
      const uri = lucideIconUri(iconName, accent);
      if (uri) {
        iconChildren.push({
          type: "img",
          props: {
            src: uri,
            width: 64,
            height: 64,
            style: { flexShrink: 0 },
          },
        });
      }
    } else {
      iconChildren.push({
        type: "span",
        props: {
          style: { fontSize: 72, lineHeight: 1 },
          children: data.icon,
        },
      });
    }
  }
  iconChildren.push({
    type: "span",
    props: {
      style: {
        color: C.text,
        fontSize: (data.title || "").length > 24 ? 44 : 56,
        fontWeight: 700,
        overflow: "hidden",
        lineHeight: 1.2,
      },
      children: truncate(data.title || "Collection", 36),
    },
  });

  children.push({
    type: "div",
    props: {
      style: {
        display: "flex",
        alignItems: "center",
        gap: 24,
        marginTop: 40,
      },
      children: iconChildren,
    },
  });

  children.push({
    type: "div",
    props: {
      style: {
        color: data.description ? C.textSecondary : C.textMuted,
        fontSize: 28,
        lineHeight: 1.5,
        marginTop: 22,
        overflow: "hidden",
      },
      children: data.description
        ? truncate(data.description, 160)
        : "A collection on Margin",
    },
  });

  children.push({
    type: "div",
    props: {
      style: {
        display: "flex",
        alignItems: "center",
        gap: 12,
        marginTop: "auto",
        paddingTop: 24,
      },
      children: [
        logoElement(24, C.textFaint),
        {
          type: "span",
          props: {
            style: { color: C.textFaint, fontSize: 20 },
            children: "margin.at",
          },
        },
      ].filter(Boolean) as unknown[],
    },
  });

  return card(children, accent);
}

export const GET: APIRoute = async ({ url }) => {
  const uri = url.searchParams.get("uri");
  if (!uri) {
    return new Response("uri parameter required", { status: 400 });
  }

  try {
    const data = await fetchRecordData(uri);
    if (!data) {
      return new Response("Record not found", { status: 404 });
    }

    const fonts = loadFonts();

    let element: unknown;
    switch (data.type) {
      case "collection":
        element = buildCollectionImage(data);
        break;
      case "bookmark":
        element = buildBookmarkImage(data);
        break;
      case "highlight":
        element = buildHighlightImage(data);
        break;
      case "annotation":
      default:
        element = buildAnnotationImage(data);
        break;
    }

    const svg = await satori(element as React.ReactNode, {
      width: 1200,
      height: 630,
      fonts: [
        { name: "Inter", data: fonts.regular, weight: 400, style: "normal" },
        { name: "Inter", data: fonts.bold, weight: 700, style: "normal" },
      ],
      loadAdditionalAsset: async (code: string, segment: string) => {
        if (code === "emoji") {
          const codepoints = [...segment]
            .map((c) => c.codePointAt(0)!.toString(16))
            .join("-");
          const emojiUrl = `https://cdn.jsdelivr.net/gh/twitter/twemoji@14.0.2/assets/svg/${codepoints}.svg`;
          try {
            const res = await fetch(emojiUrl);
            if (res.ok)
              return `data:image/svg+xml,${encodeURIComponent(await res.text())}`;
          } catch {
            // ignore
          }
        }
        return "";
      },
    });

    const resvg = new Resvg(svg, {
      fitTo: { mode: "width", value: 1200 },
    });
    const png = resvg.render().asPng();

    return new Response(new Uint8Array(png), {
      headers: {
        "Content-Type": "image/png",
        "Cache-Control": "public, max-age=86400",
      },
    });
  } catch (e) {
    console.error("[og-image] Error generating image:", e);
    console.error("[og-image] cwd:", process.cwd());
    console.error("[og-image] publicDir:", getPublicDir());
    const BASE_URL = process.env.BASE_URL || "https://margin.at";
    return Response.redirect(`${BASE_URL}/og.png`, 302);
  }
};
