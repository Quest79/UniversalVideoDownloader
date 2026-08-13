"use strict";

const VIDEOISH_SELECTOR = [
  "video",
  "ytd-player",
  ".html5-video-player",
  "[data-testid*='video']",
  "[class*='video-player']",
  "[class*='VideoPlayer']",
  "[class*='videoPlayer']",
  "[aria-label*='video' i]"
].join(",");

function absoluteHref(a) {
  try { return a && a.href ? new URL(a.href, location.href).href : ""; } catch (_) { return ""; }
}

function closestHref(target, patterns) {
  let el = target instanceof Element ? target : null;
  for (let depth = 0; el && depth < 12; depth++, el = el.parentElement) {
    const anchors = [];
    if (el.matches?.("a[href]")) anchors.push(el);
    anchors.push(...el.querySelectorAll?.("a[href]") || []);
    for (const a of anchors) {
      const href = absoluteHref(a);
      if (href && patterns.some((p) => p.test(href))) return href;
    }
  }
  return "";
}

function findVideo(target) {
  if (!(target instanceof Element)) return null;
  if (target.matches("video")) return target;
  const direct = target.closest("video");
  if (direct) return direct;

  let el = target;
  for (let depth = 0; el && depth < 10; depth++, el = el.parentElement) {
    const v = el.querySelector?.("video");
    if (v) return v;
    if (el.matches?.(VIDEOISH_SELECTOR)) return el;
  }
  return null;
}

function isLikelyVideoContext(target) {
  if (findVideo(target)) return true;
  if (!(target instanceof Element)) return false;
  // Player overlays frequently sit above the actual <video>.
  const rect = target.getBoundingClientRect?.();
  if (!rect) return false;
  const videos = [...document.querySelectorAll("video")];
  return videos.some((v) => {
    const r = v.getBoundingClientRect();
    const cx = rect.left + Math.min(rect.width, 10) / 2;
    const cy = rect.top + Math.min(rect.height, 10) / 2;
    return cx >= r.left && cx <= r.right && cy >= r.top && cy <= r.bottom;
  });
}

function sitePermalink(target) {
  const host = location.hostname.toLowerCase().replace(/^www\./, "");
  const page = location.href;

  const rules = [
    { hosts: ["instagram.com"], patterns: [/instagram\.com\/(?:reel|p|tv)\/[A-Za-z0-9_-]+/i] },
    { hosts: ["facebook.com", "m.facebook.com", "fb.watch"], patterns: [/facebook\.com\/(?:reel|watch|share\/v|[^/]+\/videos)\//i, /facebook\.com\/watch\?v=/i, /fb\.watch\//i] },
    { hosts: ["x.com", "twitter.com"], patterns: [/(?:x|twitter)\.com\/[^/]+\/status\/\d+/i] },
    { hosts: ["reddit.com", "old.reddit.com"], patterns: [/reddit\.com\/r\/[^/]+\/comments\//i, /reddit\.com\/comments\//i] },
    { hosts: ["tiktok.com"], patterns: [/tiktok\.com\/@[^/]+\/video\/\d+/i] },
    { hosts: ["pinterest.com"], patterns: [/pinterest\.[^/]+\/pin\/\d+/i] },
    { hosts: ["tumblr.com"], patterns: [/tumblr\.com\/[^/]+\/\d+/i, /[^/]+\.tumblr\.com\/post\/\d+/i] },
    { hosts: ["9gag.com"], patterns: [/9gag\.com\/gag\//i] },
    { hosts: ["linkedin.com"], patterns: [/linkedin\.com\/(?:feed\/update|posts|events)\//i] },
    { hosts: ["vk.com"], patterns: [/vk\.com\/video/i] },
    { hosts: ["ok.ru"], patterns: [/ok\.ru\/video\//i] },
    { hosts: ["imgur.com"], patterns: [/imgur\.com\/(?:gallery|a)\//i, /imgur\.com\/[A-Za-z0-9]+/i] },
    { hosts: ["streamable.com"], patterns: [/streamable\.com\/[A-Za-z0-9]+/i] },
    { hosts: ["twitch.tv"], patterns: [/twitch\.tv\/videos\/\d+/i, /clips\.twitch\.tv\//i, /twitch\.tv\/[^/]+\/clip\//i] },
    { hosts: ["bilibili.com"], patterns: [/bilibili\.com\/video\//i] },
    { hosts: ["dailymotion.com"], patterns: [/dailymotion\.com\/video\//i] },
    { hosts: ["rumble.com"], patterns: [/rumble\.com\/v[^/]+/i] },
    { hosts: ["nicovideo.jp"], patterns: [/nicovideo\.jp\/watch\//i] },
    { hosts: ["rutube.ru"], patterns: [/rutube\.ru\/video\//i] },
    { hosts: ["flickr.com"], patterns: [/flickr\.com\/photos\/[^/]+\/\d+/i] },
    { hosts: ["odysee.com"], patterns: [/odysee\.com\/@/i, /odysee\.com\/\$/i] },
    { hosts: ["likee.video"], patterns: [/likee\.video\//i] },
    { hosts: ["truthsocial.com"], patterns: [/truthsocial\.com\/@[^/]+\/posts\//i] },
    { hosts: ["substack.com"], patterns: [/substack\.com\//i] }
  ];

  for (const rule of rules) {
    if (rule.hosts.some((h) => host === h || host.endsWith("." + h))) {
      const found = closestHref(target, rule.patterns);
      if (found) return found;
    }
  }

  // These sites normally expose the target video in the page URL itself.
  if (/^(?:youtube\.com|youtu\.be|vimeo\.com|dailymotion\.com|bilibili\.com|rumble\.com|streamable\.com|twitch\.tv|nicovideo\.jp|rutube\.ru|espn\.com|cnn\.com|bbc\.(?:com|co\.uk)|nytimes\.com|foxnews\.com|yahoo\.com)$/i.test(host)) {
    return page;
  }

  return page;
}

function contextPayload(target) {
  const video = findVideo(target);
  const hasVideo = Boolean(video) || isLikelyVideoContext(target);
  let mediaUrl = "";
  if (video instanceof HTMLVideoElement) {
    const src = video.currentSrc || video.src || video.querySelector("source[src]")?.src || "";
    if (/^https?:/i.test(src)) mediaUrl = src;
  }
  const host = location.hostname.toLowerCase().replace(/^www\./, "");
  const siteHandled = [
    "youtube.com", "youtu.be", "instagram.com", "facebook.com", "fb.watch",
    "x.com", "twitter.com", "reddit.com", "tiktok.com", "pinterest.com",
    "linkedin.com", "twitch.tv", "vimeo.com", "dailymotion.com", "bilibili.com",
    "rumble.com", "tumblr.com", "vk.com", "ok.ru", "nicovideo.jp", "streamable.com",
    "imgur.com", "9gag.com", "rutube.ru", "flickr.com", "odysee.com", "likee.video",
    "truthsocial.com", "substack.com", "espn.com", "cnn.com", "bbc.com", "bbc.co.uk",
    "nytimes.com", "foxnews.com", "yahoo.com"
  ].some((h) => host === h || host.endsWith("." + h));
  return {
    type: "uvd-context",
    hasVideo,
    pageUrl: location.href,
    targetUrl: sitePermalink(target),
    mediaUrl,
    preferMedia: Boolean(mediaUrl && !siteHandled)
  };
}

document.addEventListener("contextmenu", (event) => {
  try {
    browser.runtime.sendMessage(contextPayload(event.target));
  } catch (_) {}
}, true);

// In-page status indicator so download actions remain visible even when OS notifications are disabled.
(function setupUvdToasts() {
  let current = null;
  let timer = null;

  function showToast(message, kind) {
    if (current) current.remove();
    if (timer) clearTimeout(timer);

    const el = document.createElement("div");
    el.textContent = String(message || "");
    el.setAttribute("data-uvd-toast", "1");
    Object.assign(el.style, {
      position: "fixed",
      zIndex: "2147483647",
      right: "18px",
      bottom: "18px",
      maxWidth: "520px",
      padding: "12px 16px",
      borderRadius: "10px",
      color: "white",
      background: kind === "error" ? "#b42318" : kind === "success" ? "#067647" : "#25252d",
      font: "600 14px/1.35 system-ui, sans-serif",
      boxShadow: "0 8px 28px rgba(0,0,0,.35)",
      pointerEvents: "none"
    });

    (document.documentElement || document.body).appendChild(el);
    current = el;
    timer = setTimeout(() => {
      if (current === el) current = null;
      el.remove();
    }, kind === "error" ? 9000 : 6000);
  }

  browser.runtime.onMessage.addListener((msg) => {
    if (msg?.type === "uvd-toast") showToast(msg.message, msg.kind || "info");
  });
})();
