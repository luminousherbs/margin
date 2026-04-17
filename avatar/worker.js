export default {
  async fetch(request, env) {
    const stringToColor = (str) => {
      let hash = 0;
      for (let i = 0; i < str.length; i++) {
        hash = str.charCodeAt(i) + ((hash << 5) - hash);
      }
      let color = "#";
      for (let i = 0; i < 3; i++) {
        const value = (hash >> (i * 8)) & 0xff;
        color += ("00" + value.toString(16)).substr(-2);
      }
      return color;
    };

    const url = new URL(request.url);
    const { pathname, searchParams } = url;

    if (!pathname || pathname === "/") {
      return new Response(`This is Margin's avatar service. It fetches avatars directly from the AT Protocol PDS and caches them on Cloudflare.
You can't use this directly unfortunately since all requests are signed and may only originate from the appview.`);
    }

    const size = searchParams.get("size");
    const resizeToTiny = size === "tiny";

    const cache = caches.default;
    let cacheKey = request.url;
    let response = await cache.match(cacheKey);
    if (response) return response;

    const pathParts = pathname.slice(1).split("/");
    if (pathParts.length < 2) {
      return new Response("Bad URL", { status: 400 });
    }

    const [signatureHex, actor] = pathParts;
    const decodedActor = decodeURIComponent(actor);
    const actorBytes = new TextEncoder().encode(decodedActor);

    const key = await crypto.subtle.importKey(
      "raw",
      new TextEncoder().encode(env.AVATAR_SHARED_SECRET),
      { name: "HMAC", hash: "SHA-256" },
      false,
      ["sign", "verify"],
    );

    const sigBytes = Uint8Array.from(
      signatureHex.match(/.{2}/g).map((b) => parseInt(b, 16)),
    );
    const valid = await crypto.subtle.verify("HMAC", key, sigBytes, actorBytes);

    if (!valid) {
      return new Response("Invalid signature", { status: 403 });
    }

    try {
      let avatarUrl = null;
      const marginApiUrl = env.MARGIN_API_URL || "https://margin.at";

      try {
        const marginResponse = await fetch(
          `${marginApiUrl}/api/profile/${decodedActor}`,
        );
        if (marginResponse.ok) {
          const marginProfile = await marginResponse.json();
          if (marginProfile.avatar) {
            if (typeof marginProfile.avatar === "string") {
              avatarUrl = marginProfile.avatar;
            }
          }
        }
      } catch (e) {}

      if (!avatarUrl) {
        try {
          const identityResp = await fetch(
            `https://slingshot.microcosm.blue/xrpc/blue.microcosm.identity.resolveMiniDoc?identifier=${encodeURIComponent(decodedActor)}`,
          );
          if (identityResp.ok) {
            const identity = await identityResp.json();
            const did = identity.did;
            const pds = identity.pds;
            if (did && pds) {
              const profileResp = await fetch(
                `${pds}/xrpc/com.atproto.repo.getRecord?repo=${encodeURIComponent(did)}&collection=app.bsky.actor.profile&rkey=self`,
              );
              if (profileResp.ok) {
                const profileRecord = await profileResp.json();
                const avatarBlob = profileRecord?.value?.avatar;
                const cid = avatarBlob?.ref?.$link;
                if (cid) {
                  avatarUrl = `${pds}/xrpc/com.atproto.sync.getBlob?did=${encodeURIComponent(did)}&cid=${encodeURIComponent(cid)}`;
                }
              }
            }
          }
        } catch (e) {}
      }

      if (!avatarUrl) {
        const bgColor = stringToColor(decodedActor);
        const sizePx = resizeToTiny ? 32 : 128;
        const svg = `<svg width="${sizePx}" height="${sizePx}" viewBox="0 0 ${sizePx} ${sizePx}" xmlns="http://www.w3.org/2000/svg"><rect width="${sizePx}" height="${sizePx}" fill="${bgColor}"/></svg>`;
        const svgData = new TextEncoder().encode(svg);

        response = new Response(svgData, {
          headers: {
            "Content-Type": "image/svg+xml",
            "Cache-Control": "public, max-age=43200",
            "Access-Control-Allow-Origin": "*",
          },
        });
        await cache.put(cacheKey, response.clone());
        return response;
      }

      let avatarResponse;
      if (resizeToTiny) {
        avatarResponse = await fetch(avatarUrl, {
          cf: {
            image: {
              width: 32,
              height: 32,
              fit: "cover",
              format: "webp",
            },
          },
        });
      } else {
        avatarResponse = await fetch(avatarUrl);
      }

      if (!avatarResponse.ok) {
        const bgColor = stringToColor(decodedActor);
        const sizePx = resizeToTiny ? 32 : 128;
        const svg = `<svg width="${sizePx}" height="${sizePx}" viewBox="0 0 ${sizePx} ${sizePx}" xmlns="http://www.w3.org/2000/svg"><rect width="${sizePx}" height="${sizePx}" fill="${bgColor}"/></svg>`;
        return new Response(svg, {
          headers: {
            "Content-Type": "image/svg+xml",
            "Access-Control-Allow-Origin": "*",
          },
        });
      }

      const avatarData = await avatarResponse.arrayBuffer();
      const contentType =
        avatarResponse.headers.get("content-type") || "image/jpeg";

      response = new Response(avatarData, {
        headers: {
          "Content-Type": contentType,
          "Cache-Control": "public, max-age=86400",
          "Access-Control-Allow-Origin": "*",
        },
      });

      await cache.put(cacheKey, response.clone());
      return response;
    } catch (error) {
      return new Response(`error fetching avatar: ${error.message}`, {
        status: 500,
      });
    }
  },
};
