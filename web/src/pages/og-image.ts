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

let logoDataURI: string | null = null;

function getLogoDataURI(): string {
  if (logoDataURI) return logoDataURI;
  try {
    const publicDir = getPublicDir();
    const buf = readFileSync(join(publicDir, "logo.svg"));
    logoDataURI = `data:image/svg+xml;base64,${buf.toString("base64")}`;
  } catch {
    logoDataURI = "";
  }
  return logoDataURI;
}

const API_URL = process.env.API_URL || "http://localhost:8081";

interface RecordData {
  type: "annotation" | "highlight" | "bookmark" | "collection";
  author: string;
  avatarURL: string;
  text: string;
  quote: string;
  source: string;
  title: string;
  icon: string;
  description: string;
}

async function fetchRecordData(uri: string): Promise<RecordData | null> {
  try {
    const res = await fetch(
      `${API_URL}/api/annotation?uri=${encodeURIComponent(uri)}`,
    );
    if (res.ok) {
      const item = await res.json();
      const author = item.author || item.creator || {};
      const handle = author.handle
        ? `@${author.handle}`
        : author.did || "someone";
      const avatarURL = author.avatar || "";
      const targetSource = item.target?.source || item.url || item.source || "";
      const domain = targetSource
        ? (() => {
            try {
              return new URL(targetSource).host;
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
          author: handle,
          avatarURL,
          text: targetTitle,
          quote: selectorText,
          source: domain,
          title: "",
          icon: "",
          description: "",
        };
      }

      if (uri.includes("/at.margin.bookmark/")) {
        return {
          type: "bookmark",
          author: handle,
          avatarURL,
          text: item.title || targetTitle || "Bookmark",
          quote: item.description || bodyText || "",
          source: domain,
          title: "",
          icon: "",
          description: "",
        };
      }

      return {
        type: "annotation",
        author: handle,
        avatarURL,
        text: bodyText,
        quote: selectorText,
        source: domain,
        title: "",
        icon: "",
        description: "",
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
      const handle = author.handle
        ? `@${author.handle}`
        : author.did || "someone";
      const avatarURL = author.avatar || "";

      return {
        type: "collection",
        author: handle,
        avatarURL,
        text: "",
        quote: "",
        source: "",
        title: item.name || "Collection",
        icon: item.icon || "📁",
        description: item.description || "",
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

function avatarElement(url: string, author: string, size: number): unknown {
  if (url) {
    return {
      type: "img",
      props: {
        src: url,
        width: size,
        height: size,
        style: { borderRadius: size / 2 },
      },
    };
  }
  const letter =
    author[0] === "@"
      ? author[1]?.toUpperCase() || "?"
      : author[0]?.toUpperCase() || "?";
  return {
    type: "div",
    props: {
      style: {
        width: size,
        height: size,
        borderRadius: size / 2,
        background: "#3b82f6",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        color: "white",
        fontSize: Math.round(size * 0.45),
        fontWeight: 700,
      },
      children: letter,
    },
  };
}

const typeColors: Record<
  string,
  { accent: string; badge: string; badgeText: string; bar: string }
> = {
  annotation: {
    accent: "#3b82f6",
    badge: "#1e3a8a",
    badgeText: "#60a5fa",
    bar: "#60a5fa",
  },
  highlight: {
    accent: "#eab308",
    badge: "#422006",
    badgeText: "#facc15",
    bar: "#facc15",
  },
  bookmark: {
    accent: "#22c55e",
    badge: "#052e16",
    badgeText: "#4ade80",
    bar: "#4ade80",
  },
};

function getTypeColor(type: string) {
  return typeColors[type] || typeColors.annotation;
}

function typeBadge(type: string): unknown {
  const labels: Record<string, string> = {
    annotation: "Annotation",
    highlight: "Highlight",
    bookmark: "Bookmark",
  };
  const c = getTypeColor(type);
  return {
    type: "div",
    props: {
      style: {
        padding: "6px 16px",
        borderRadius: 99,
        background: c.badge,
        color: c.badgeText,
        fontSize: 16,
        fontWeight: 600,
      },
      children: labels[type] || type,
    },
  };
}

function marginBrand(logo: string): unknown {
  if (!logo) return null;
  return {
    type: "div",
    props: {
      style: {
        display: "flex",
        alignItems: "center",
        marginLeft: "auto",
      },
      children: [
        {
          type: "img",
          props: { src: logo, width: 28, height: 24 },
        },
      ],
    },
  };
}

function buildAnnotationImage(data: RecordData, logo: string) {
  const children: unknown[] = [];
  const tc = getTypeColor(data.type);

  children.push({
    type: "div",
    props: {
      style: { display: "flex", alignItems: "center", width: "100%" },
      children: [
        avatarElement(data.avatarURL, data.author, 48),
        {
          type: "span",
          props: {
            style: { color: "#a1a1aa", fontSize: 22, marginLeft: 14 },
            children: data.author,
          },
        },
        {
          type: "div",
          props: {
            style: { marginLeft: "auto", display: "flex" },
            children: [typeBadge(data.type)],
          },
        },
      ],
    },
  });

  if (data.text) {
    children.push({
      type: "div",
      props: {
        style: {
          color: "#fafafa",
          fontSize: data.text.length > 200 ? 26 : 32,
          lineHeight: 1.45,
          marginTop: 32,
          overflow: "hidden",
        },
        children: truncate(data.text, 300),
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
                background: tc.bar,
                flexShrink: 0,
              },
            },
          },
          {
            type: "div",
            props: {
              style: {
                color: "#a1a1aa",
                fontSize: data.quote.length > 150 ? 22 : 26,
                lineHeight: 1.5,
                paddingLeft: 18,
                fontStyle: "italic",
                overflow: "hidden",
              },
              children: `"${truncate(data.quote, 250)}"`,
            },
          },
        ],
      },
    });
  }

  const footerChildren: unknown[] = [];
  if (data.source) {
    footerChildren.push({
      type: "span",
      props: {
        style: { color: "#71717a", fontSize: 20 },
        children: data.source,
      },
    });
  }
  footerChildren.push(marginBrand(logo));

  children.push({
    type: "div",
    props: {
      style: {
        display: "flex",
        alignItems: "center",
        marginTop: "auto",
        paddingTop: 28,
        borderTop: "1px solid #27272a",
      },
      children: footerChildren,
    },
  });

  return wrapCard(children, tc.accent);
}

function buildBookmarkImage(data: RecordData, logo: string) {
  const children: unknown[] = [];
  const tc = getTypeColor("bookmark");

  children.push({
    type: "div",
    props: {
      style: { display: "flex", alignItems: "center", width: "100%" },
      children: [
        {
          type: "div",
          props: {
            style: { display: "flex", alignItems: "center", gap: 10 },
            children: [
              {
                type: "div",
                props: {
                  style: {
                    width: 36,
                    height: 36,
                    borderRadius: 10,
                    background: "#052e16",
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                  },
                  children: {
                    type: "div",
                    props: {
                      style: {
                        fontSize: 18,
                        color: "#4ade80",
                        fontWeight: 700,
                      },
                      children: "🔗",
                    },
                  },
                },
              },
              {
                type: "span",
                props: {
                  style: { color: "#71717a", fontSize: 20 },
                  children: data.source || "Saved page",
                },
              },
            ],
          },
        },
        {
          type: "div",
          props: {
            style: { marginLeft: "auto", display: "flex" },
            children: [typeBadge("bookmark")],
          },
        },
      ],
    },
  });

  children.push({
    type: "div",
    props: {
      style: {
        color: "#fafafa",
        fontSize: (data.text?.length || 0) > 60 ? 36 : 44,
        fontWeight: 700,
        lineHeight: 1.3,
        marginTop: 36,
        overflow: "hidden",
      },
      children: truncate(data.text || "Untitled Bookmark", 100),
    },
  });

  if (data.quote) {
    children.push({
      type: "div",
      props: {
        style: {
          color: "#a1a1aa",
          fontSize: 24,
          lineHeight: 1.5,
          marginTop: 20,
          overflow: "hidden",
        },
        children: truncate(data.quote, 200),
      },
    });
  }

  children.push({
    type: "div",
    props: {
      style: {
        display: "flex",
        alignItems: "center",
        marginTop: "auto",
        paddingTop: 28,
        borderTop: "1px solid #27272a",
      },
      children: [
        {
          type: "div",
          props: {
            style: { display: "flex", alignItems: "center" },
            children: [
              avatarElement(data.avatarURL, data.author, 36),
              {
                type: "span",
                props: {
                  style: { color: "#71717a", fontSize: 20, marginLeft: 12 },
                  children: data.author,
                },
              },
            ],
          },
        },
        marginBrand(logo),
      ],
    },
  });

  return wrapCard(children, tc.accent);
}

function buildCollectionImage(data: RecordData, logo: string) {
  const children: unknown[] = [];

  children.push({
    type: "div",
    props: {
      style: { display: "flex", alignItems: "center", gap: 18 },
      children: [
        {
          type: "span",
          props: { style: { fontSize: 64 }, children: data.icon },
        },
        {
          type: "span",
          props: {
            style: {
              color: "#fafafa",
              fontSize: 48,
              fontWeight: 700,
              overflow: "hidden",
            },
            children: truncate(data.title, 40),
          },
        },
      ],
    },
  });

  children.push({
    type: "div",
    props: {
      style: {
        color: data.description ? "#a1a1aa" : "#71717a",
        fontSize: 26,
        lineHeight: 1.5,
        marginTop: 24,
        overflow: "hidden",
      },
      children: data.description
        ? truncate(data.description, 200)
        : "A collection on Margin",
    },
  });

  children.push({
    type: "div",
    props: {
      style: {
        display: "flex",
        alignItems: "center",
        marginTop: "auto",
        paddingTop: 28,
        borderTop: "1px solid #27272a",
      },
      children: [
        {
          type: "div",
          props: {
            style: { display: "flex", alignItems: "center" },
            children: [
              avatarElement(data.avatarURL, data.author, 36),
              {
                type: "span",
                props: {
                  style: { color: "#71717a", fontSize: 20, marginLeft: 12 },
                  children: data.author,
                },
              },
            ],
          },
        },
        marginBrand(logo),
      ],
    },
  });

  return wrapCard(children);
}

function wrapCard(children: unknown[], accent: string = "#3b82f6") {
  return {
    type: "div",
    props: {
      style: {
        display: "flex",
        width: "100%",
        height: "100%",
        background: "#09090b",
        padding: 40,
        fontFamily: "Inter",
      },
      children: [
        {
          type: "div",
          props: {
            style: {
              display: "flex",
              flexDirection: "column",
              width: "100%",
              height: "100%",
              padding: "52px 56px",
              border: "1px solid #27272a",
              borderRadius: 24,
              borderTop: `3px solid ${accent}`,
              background: "#18181b",
              overflow: "hidden",
            },
            children,
          },
        },
      ],
    },
  };
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
  const logo = getLogoDataURI();

  const element =
    data.type === "collection"
      ? buildCollectionImage(data, logo)
      : data.type === "bookmark"
        ? buildBookmarkImage(data, logo)
        : buildAnnotationImage(data, logo);

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
        const url = `https://cdn.jsdelivr.net/gh/twitter/twemoji@14.0.2/assets/svg/${codepoints}.svg`;
        try {
          const res = await fetch(url);
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
