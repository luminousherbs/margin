import type { APIRoute } from "astro";
import satori from "satori";
import { Resvg } from "@resvg/resvg-js";
import { readFileSync } from "node:fs";
import { join } from "node:path";

export const prerender = false;

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

async function fetchRecordData(uri: string): Promise<RecordData | null> {
  try {
    const res = await fetch(
      `${API_URL}/api/annotation?uri=${encodeURIComponent(uri)}`,
    );
    if (res.ok) {
      const item = await res.json();
      const author = item.author || item.creator || {};
      const handle = author.handle || "";
      const displayName = author.displayName || handle || "someone";
      const avatarURL = author.avatar || "";
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
          author: handle ? `@${handle}` : "someone",
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
          author: handle ? `@${handle}` : "someone",
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
        author: handle ? `@${handle}` : "someone",
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
      const displayName = author.displayName || handle || "someone";
      const avatarURL = author.avatar || "";

      return {
        type: "collection",
        author: handle ? `@${handle}` : "someone",
        displayName,
        avatarURL,
        text: "",
        quote: "",
        source: "",
        title: item.name || "Collection",
        icon: item.icon || "📁",
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
  return str.slice(0, max - 3) + "...";
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
  bg: "#f4f4f5",
  text: "#18181b",
  textSecondary: "#52525b",
  textMuted: "#a1a1aa",
  textFaint: "#d4d4d8",
  primary: "#3b82f6",
  primaryDark: "#2563eb",
  border: "#e4e4e7",
};

const namedColors: Record<string, string> = {
  yellow: "#facc15",
  green: "#4ade80",
  red: "#f87171",
  blue: "#60a5fa",
};

function resolveHighlightColor(color: string): string {
  if (!color) return "#facc15";
  if (color.startsWith("#")) return color;
  return namedColors[color] || "#facc15";
}

const typeColors: Record<string, string> = {
  annotation: "#3b82f6",
  highlight: "#facc15",
  bookmark: "#22c55e",
  collection: "#3b82f6",
};

function lightTint(hex: string): string {
  const r = parseInt(hex.slice(1, 3), 16);
  const g = parseInt(hex.slice(3, 5), 16);
  const b = parseInt(hex.slice(5, 7), 16);
  const mix = (c: number) => Math.round(c * 0.12 + 255 * 0.88);
  return `rgb(${mix(r)}, ${mix(g)}, ${mix(b)})`;
}

function avatarElement(url: string, name: string, size: number): unknown {
  if (url) {
    return {
      type: "img",
      props: {
        src: url,
        width: size,
        height: size,
        style: { borderRadius: size / 2, flexShrink: 0 },
      },
    };
  }
  const letter =
    name[0] === "@"
      ? name[1]?.toUpperCase() || "?"
      : name[0]?.toUpperCase() || "?";
  return {
    type: "div",
    props: {
      style: {
        width: size,
        height: size,
        borderRadius: size / 2,
        background: "#e4e4e7",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        color: "#71717a",
        fontSize: Math.round(size * 0.45),
        fontWeight: 700,
        flexShrink: 0,
      },
      children: letter,
    },
  };
}

function coloredLogoUri(color: string): string {
  const publicDir = getPublicDir();
  const svg = readFileSync(join(publicDir, "logo.svg"), "utf-8");
  const recolored = svg.replace(/fill="[^"]*"/, `fill="${color}"`);
  return `data:image/svg+xml,${encodeURIComponent(recolored)}`;
}

function logoElement(size: number, color: string): unknown {
  return {
    type: "img",
    props: {
      src: coloredLogoUri(color),
      width: size,
      height: size,
      style: { flexShrink: 0 },
    },
  };
}

const typeLabels: Record<string, string> = {
  annotation: "Annotation",
  highlight: "Highlight",
  bookmark: "Bookmark",
  collection: "Collection",
};

function headerWithBadge(data: RecordData, accentColor: string): unknown {
  return {
    type: "div",
    props: {
      style: { display: "flex", alignItems: "center" },
      children: [
        avatarElement(data.avatarURL, data.displayName, 48),
        {
          type: "div",
          props: {
            style: {
              display: "flex",
              flexDirection: "column",
              marginLeft: 14,
              flex: 1,
            },
            children: [
              {
                type: "span",
                props: {
                  style: {
                    color: C.text,
                    fontSize: 22,
                    fontWeight: 600,
                  },
                  children: data.displayName,
                },
              },
              {
                type: "span",
                props: {
                  style: {
                    color: C.textMuted,
                    fontSize: 17,
                    marginTop: 1,
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
            },
            children: [
              logoElement(24, accentColor),
              {
                type: "span",
                props: {
                  style: {
                    color: C.textFaint,
                    fontSize: 18,
                  },
                  children: "|",
                },
              },
              {
                type: "span",
                props: {
                  style: {
                    color: accentColor,
                    fontSize: 16,
                    fontWeight: 600,
                    textTransform: "uppercase" as const,
                    letterSpacing: 1,
                  },
                  children: typeLabels[data.type] || data.type,
                },
              },
            ],
          },
        },
      ],
    },
  };
}

function footerSource(source?: string): unknown | null {
  if (!source) return null;
  return {
    type: "div",
    props: {
      style: {
        display: "flex",
        alignItems: "center",
        marginTop: "auto",
        paddingTop: 16,
      },
      children: [
        {
          type: "span",
          props: {
            style: { color: C.textMuted, fontSize: 16 },
            children: source,
          },
        },
      ],
    },
  };
}

function wrap(children: unknown[], bg?: string): unknown {
  return {
    type: "div",
    props: {
      style: {
        display: "flex",
        flexDirection: "column",
        width: "100%",
        height: "100%",
        background: bg || C.bg,
        padding: "48px 64px",
        fontFamily: "Inter",
      },
      children,
    },
  };
}

function buildAnnotationImage(data: RecordData) {
  const accent = typeColors.annotation;
  const children: unknown[] = [headerWithBadge(data, accent)];

  if (data.text) {
    const len = data.text.length;
    children.push({
      type: "div",
      props: {
        style: {
          color: C.text,
          fontSize: len > 200 ? 26 : len > 100 ? 30 : 36,
          fontWeight: 500,
          lineHeight: 1.45,
          marginTop: 32,
          overflow: "hidden",
        },
        children: truncate(data.text, 280),
      },
    });
  }

  if (data.quote) {
    children.push({
      type: "div",
      props: {
        style: { display: "flex", marginTop: 24 },
        children: [
          {
            type: "div",
            props: {
              style: {
                width: 4,
                borderRadius: 2,
                background: accent,
                flexShrink: 0,
              },
            },
          },
          {
            type: "div",
            props: {
              style: {
                color: C.textSecondary,
                fontSize: 20,
                lineHeight: 1.6,
                paddingLeft: 20,
                fontStyle: "italic",
                overflow: "hidden",
              },
              children: truncate(data.quote, 200),
            },
          },
        ],
      },
    });
  }

  const footer = footerSource(data.source);
  if (footer) children.push(footer);

  return wrap(children, lightTint(accent));
}

function buildHighlightImage(data: RecordData) {
  const highlightColor = resolveHighlightColor(data.color);
  const bgTint = lightTint(highlightColor);
  const quoteText = data.quote || data.text || "Highlighted passage";
  const len = quoteText.length;

  return {
    type: "div",
    props: {
      style: {
        display: "flex",
        flexDirection: "column",
        width: "100%",
        height: "100%",
        background: bgTint,
        padding: "48px 64px",
        fontFamily: "Inter",
      },
      children: [
        headerWithBadge(data, highlightColor),
        {
          type: "div",
          props: {
            style: {
              color: highlightColor,
              fontSize: 120,
              fontWeight: 700,
              lineHeight: 1,
              marginTop: 28,
            },
            children: "\u201C",
          },
        },
        {
          type: "div",
          props: {
            style: {
              color: C.text,
              fontSize: len > 150 ? 28 : len > 80 ? 34 : 42,
              fontWeight: 600,
              lineHeight: 1.4,
              marginTop: -30,
              overflow: "hidden",
            },
            children: truncate(quoteText, 240),
          },
        },
        data.text && data.quote
          ? {
              type: "div",
              props: {
                style: {
                  color: C.textSecondary,
                  fontSize: 20,
                  marginTop: 20,
                },
                children: truncate(data.text, 80),
              },
            }
          : null,
        {
          type: "div",
          props: {
            style: {
              display: "flex",
              alignItems: "center",
              marginTop: "auto",
              paddingTop: 16,
            },
            children: [
              data.source
                ? {
                    type: "span",
                    props: {
                      style: { color: C.textMuted, fontSize: 16 },
                      children: data.source,
                    },
                  }
                : null,
            ].filter(Boolean),
          },
        },
      ].filter(Boolean),
    },
  };
}

function buildBookmarkImage(data: RecordData) {
  const children: unknown[] = [headerWithBadge(data, typeColors.bookmark)];

  if (data.source) {
    children.push({
      type: "div",
      props: {
        style: {
          color: typeColors.bookmark,
          fontSize: 18,
          marginTop: 32,
        },
        children: data.source,
      },
    });
  }

  const titleLen = (data.text || "").length;
  children.push({
    type: "div",
    props: {
      style: {
        color: C.text,
        fontSize: titleLen > 60 ? 34 : 42,
        fontWeight: 700,
        lineHeight: 1.25,
        marginTop: data.source ? 10 : 32,
        overflow: "hidden",
      },
      children: truncate(data.text || "Untitled Bookmark", 90),
    },
  });

  if (data.quote) {
    children.push({
      type: "div",
      props: {
        style: {
          color: C.textSecondary,
          fontSize: 22,
          lineHeight: 1.5,
          marginTop: 16,
          overflow: "hidden",
        },
        children: truncate(data.quote, 180),
      },
    });
  }

  return wrap(children, lightTint(typeColors.bookmark));
}

function buildCollectionImage(data: RecordData) {
  const children: unknown[] = [headerWithBadge(data, typeColors.collection)];

  children.push({
    type: "div",
    props: {
      style: {
        display: "flex",
        alignItems: "center",
        gap: 20,
        marginTop: 36,
      },
      children: [
        {
          type: "span",
          props: { style: { fontSize: 52 }, children: data.icon },
        },
        {
          type: "span",
          props: {
            style: {
              color: C.text,
              fontSize: 44,
              fontWeight: 700,
              overflow: "hidden",
            },
            children: truncate(data.title, 36),
          },
        },
      ],
    },
  });

  children.push({
    type: "div",
    props: {
      style: {
        color: data.description ? C.textSecondary : C.textMuted,
        fontSize: 24,
        lineHeight: 1.5,
        marginTop: 20,
        overflow: "hidden",
      },
      children: data.description
        ? truncate(data.description, 180)
        : "A collection on Margin",
    },
  });

  return wrap(children, lightTint(typeColors.collection));
}

export const GET: APIRoute = async ({ url }) => {
  const uri = url.searchParams.get("uri");
  if (!uri) {
    return new Response("uri parameter required", { status: 400 });
  }

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
};
