#!/usr/bin/env node

/**
 * OG Image Preview Generator
 *
 * Usage:
 *   node tools/preview-og.mjs                          # generates all sample types
 *   node tools/preview-og.mjs --uri at://did/col/rkey  # fetches real data from running backend
 */

import satori from "satori";
import { Resvg } from "@resvg/resvg-js";
import { readFileSync, writeFileSync, mkdirSync } from "node:fs";
import { join, dirname } from "node:path";
import { fileURLToPath } from "node:url";
import { execSync } from "node:child_process";

const __dirname = dirname(fileURLToPath(import.meta.url));
const ROOT = join(__dirname, "..");

const fontsDir = join(ROOT, "public", "fonts");
const regular = readFileSync(join(fontsDir, "Inter-Regular.ttf"));
const bold = readFileSync(join(fontsDir, "Inter-Bold.ttf"));

let logoDataURI = "";
try {
  const buf = readFileSync(join(ROOT, "public", "logo.svg"));
  logoDataURI = `data:image/svg+xml;base64,${buf.toString("base64")}`;
} catch {}

const outDir = join(ROOT, "tools", "og-preview");
mkdirSync(outDir, { recursive: true });

function truncate(str, max) {
  if (str.length <= max) return str;
  return str.slice(0, max - 3) + "...";
}

function wrapCard(children) {
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
              borderTop: `3px solid ${children.__accent || "#3b82f6"}`,
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

function avatarCircle(author, size = 48) {
  const letter =
    author[0] === "@"
      ? (author[1] || "?").toUpperCase()
      : (author[0] || "?").toUpperCase();
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

const typeColors = {
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

function getTypeColor(type) {
  return typeColors[type] || typeColors.annotation;
}

function typeBadge(type) {
  const labels = {
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

function marginBrand() {
  if (!logoDataURI) return null;
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
          props: { src: logoDataURI, width: 28, height: 24 },
        },
      ],
    },
  };
}

function buildAnnotationImage(data) {
  const children = [];
  const tc = getTypeColor(data.type || "annotation");

  children.push({
    type: "div",
    props: {
      style: { display: "flex", alignItems: "center", width: "100%" },
      children: [
        data.avatarURL
          ? {
              type: "img",
              props: {
                src: data.avatarURL,
                width: 48,
                height: 48,
                style: { borderRadius: 24 },
              },
            }
          : avatarCircle(data.author, 48),
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
            children: [typeBadge(data.type || "annotation")],
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

  const footerChildren = [];
  if (data.source)
    footerChildren.push({
      type: "span",
      props: {
        style: { color: "#71717a", fontSize: 20 },
        children: data.source,
      },
    });
  footerChildren.push(marginBrand());
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

  children.__accent = tc.accent;
  return wrapCard(children);
}

function buildBookmarkImage(data) {
  const children = [];
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

  const authorChildren = [
    avatarCircle(data.author, 36),
    {
      type: "span",
      props: {
        style: { color: "#71717a", fontSize: 20, marginLeft: 12 },
        children: data.author,
      },
    },
  ];
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
            children: authorChildren,
          },
        },
        marginBrand(),
      ],
    },
  });

  children.__accent = tc.accent;
  return wrapCard(children);
}

function buildCollectionImage(data) {
  const children = [];
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

  const authorChildren = [
    avatarCircle(data.author, 36),
    {
      type: "span",
      props: {
        style: { color: "#71717a", fontSize: 20, marginLeft: 12 },
        children: data.author,
      },
    },
  ];
  const footerChildren = [
    {
      type: "div",
      props: {
        style: { display: "flex", alignItems: "center" },
        children: authorChildren,
      },
    },
  ];
  footerChildren.push(marginBrand());

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

  return wrapCard(children);
}

async function renderPNG(element, filename) {
  const svg = await satori(element, {
    width: 1200,
    height: 630,
    fonts: [
      { name: "Inter", data: regular.buffer, weight: 400, style: "normal" },
      { name: "Inter", data: bold.buffer, weight: 700, style: "normal" },
    ],
    loadAdditionalAsset: async (code, segment) => {
      if (code === "emoji") {
        const codepoints = [...segment]
          .map((c) => c.codePointAt(0).toString(16))
          .join("-");
        const url = `https://cdn.jsdelivr.net/gh/twitter/twemoji@14.0.2/assets/svg/${codepoints}.svg`;
        try {
          const res = await fetch(url);
          if (res.ok)
            return `data:image/svg+xml,${encodeURIComponent(await res.text())}`;
        } catch {}
      }
      return "";
    },
  });

  const resvg = new Resvg(svg, { fitTo: { mode: "width", value: 1200 } });
  const png = resvg.render().asPng();
  const out = join(outDir, filename);
  writeFileSync(out, png);
  console.log(`  ✓ ${out}`);
  return out;
}

const samples = [
  {
    name: "annotation.png",
    builder: buildAnnotationImage,
    data: {
      type: "annotation",
      author: "@alice.bsky.social",
      avatarURL: "",
      text: "This is a really insightful point about decentralized identity. The AT Protocol's approach to portable accounts changes everything.",
      quote:
        "Users should own their data and be able to move between services without losing their identity or social graph.",
      source: "atproto.com",
    },
  },
  {
    name: "highlight.png",
    builder: buildAnnotationImage,
    data: {
      type: "highlight",
      author: "@bob.bsky.social",
      avatarURL: "",
      text: "",
      quote:
        "The web annotation data model provides a framework for sharing annotations across different platforms, creating an interoperable layer of user-generated metadata on top of existing web content.",
      source: "w3.org",
    },
  },
  {
    name: "bookmark.png",
    builder: buildBookmarkImage,
    data: {
      type: "bookmark",
      author: "@carol.bsky.social",
      avatarURL: "",
      text: "How to Build a Chrome Extension with React and TypeScript",
      quote:
        "A comprehensive guide covering manifest v3, content scripts, popup pages, and background workers.",
      source: "dev.to",
    },
  },
  {
    name: "collection.png",
    builder: buildCollectionImage,
    data: {
      author: "@dave.bsky.social",
      avatarURL: "",
      title: "Web Standards",
      icon: "🌍",
      description:
        "Articles and specs about W3C web standards, accessibility, and the open web platform.",
    },
  },
  {
    name: "collection-minimal.png",
    builder: buildCollectionImage,
    data: {
      author: "@eve.bsky.social",
      avatarURL: "",
      title: "Reading List",
      icon: "📚",
      description: "",
    },
  },
];

const args = process.argv.slice(2);
const uriArg = args.find(
  (a) => a.startsWith("--uri=") || args[args.indexOf("--uri") + 1],
);
const uri = uriArg?.startsWith("--uri=")
  ? uriArg.slice(6)
  : args[args.indexOf("--uri") + 1];

if (uri) {
  const apiURL = process.env.API_URL || "http://localhost:8081";
  console.log(`Fetching ${uri} from ${apiURL}...`);

  let data = null;

  try {
    const res = await fetch(
      `${apiURL}/api/annotation?uri=${encodeURIComponent(uri)}`,
    );
    if (res.ok) {
      const item = await res.json();
      const author = item.author || item.creator || {};
      const handle = author.handle
        ? `@${author.handle}`
        : author.did || "someone";
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
      data = {
        author: handle,
        avatarURL: author.avatar || "",
        text: item.body || item.bodyValue || item.text || item.title || "",
        quote:
          item.target?.selector?.exact ||
          item.selector?.exact ||
          item.description ||
          "",
        source: domain,
      };
    }
  } catch {}

  if (!data) {
    try {
      const res = await fetch(
        `${apiURL}/api/collection?uri=${encodeURIComponent(uri)}`,
      );
      if (res.ok) {
        const item = await res.json();
        const author = item.author || item.creator || {};
        data = {
          type: "collection",
          author: author.handle ? `@${author.handle}` : author.did || "someone",
          avatarURL: author.avatar || "",
          title: item.name || "Collection",
          icon: item.icon || "📁",
          description: item.description || "",
        };
      }
    } catch {}
  }

  if (!data) {
    console.error("Could not fetch record for URI:", uri);
    process.exit(1);
  }

  const element =
    data.type === "collection"
      ? buildCollectionImage(data)
      : buildAnnotationImage(data);
  const file = await renderPNG(element, "live-preview.png");
  tryOpen(file);
} else {
  console.log("Generating OG image previews...\n");
  let lastFile;
  for (const s of samples) {
    lastFile = await renderPNG(s.builder(s.data), s.name);
  }
  console.log(`\nDone! Files in ${outDir}`);
  tryOpen(outDir);
}

function tryOpen(path) {
  try {
    execSync(`open "${path}"`, { stdio: "ignore" });
  } catch {}
}
