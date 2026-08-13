"use strict";

const HOST = "universal_video_downloader_v44";
const ROOT = "uvd-root";
const SLOT_COUNT = 12;
const SLOT_IDS = Array.from({ length: SLOT_COUNT }, (_, i) => `uvd-slot-${i}`);
const SCAN_OPTION = { id: "scan", label: "Scan with local helper (may use site cookies)…" };
const recentContext = new Map();
const probeCache = new Map();
const activeMenu = new Map();

const BACKEND_VERSION = "4.4.0";
const COMPANION_URL = "https://github.com/Quest79/UniversalVideoDownloader/releases/latest";

function notify(title, message) {
  browser.notifications.create({
    type: "basic",
    iconUrl: browser.runtime.getURL("icons/icon-96.png"),
    title,
    message: String(message || "")
  }).catch(() => {});
}

function toast(tabId, message, kind = "info") {
  if (tabId == null) return;
  browser.tabs.sendMessage(tabId, { type: "uvd-toast", message: String(message || ""), kind }).catch(() => {});
}

async function native(message) {
  let response;
  try {
    response = await browser.runtime.sendNativeMessage(HOST, message);
  } catch (error) {
    const msg = String(error?.message || error || "Native helper error");
    if (/not found|No such native application|native.*host|specified native messaging host/i.test(msg)) {
      throw new Error(`The Windows companion is not installed. Download it from ${COMPANION_URL}, extract it, and run INSTALL_HELPER.bat.`);
    }
    throw new Error(msg);
  }
  if (!response) throw new Error("The v4.4 backend returned no response.");
  if (response.backend_version !== BACKEND_VERSION) {
    throw new Error(`Wrong backend version (${response.backend_version || "unknown"}). Run INSTALL_HELPER.bat from the v4.4 bundle.`);
  }
  return response;
}

browser.runtime.onMessage.addListener((msg, sender) => {
  if (msg?.type !== "uvd-context" || sender.tab?.id == null) return;
  if (sender.tab.incognito) {
    recentContext.delete(sender.tab.id);
    activeMenu.delete(sender.tab.id);
    return;
  }
  recentContext.set(sender.tab.id, { ...msg, received: Date.now() });
});

browser.tabs.onRemoved.addListener((tabId) => {
  recentContext.delete(tabId);
  activeMenu.delete(tabId);
});

browser.contextMenus.create({
  id: ROOT,
  title: "Download video with local helper",
  contexts: ["video", "page", "link"],
  visible: true
});

for (let i = 0; i < SLOT_COUNT; i++) {
  browser.contextMenus.create({
    id: SLOT_IDS[i],
    parentId: ROOT,
    title: i === 0 ? SCAN_OPTION.label : "—",
    contexts: ["video", "page", "link"],
    visible: i === 0,
    enabled: i === 0
  });
}

async function updateSlots(options, loading = false) {
  const ops = [];
  for (let i = 0; i < SLOT_COUNT; i++) {
    const option = options?.[i];
    if (loading && i === 0) {
      ops.push(browser.contextMenus.update(SLOT_IDS[i], { visible: true, enabled: false, title: "Scanning available qualities…" }));
    } else if (option) {
      ops.push(browser.contextMenus.update(SLOT_IDS[i], { visible: true, enabled: !option.disabled, title: option.label.slice(0, 120) }));
    } else {
      ops.push(browser.contextMenus.update(SLOT_IDS[i], { visible: false, enabled: false, title: "—" }));
    }
  }
  await Promise.allSettled(ops);
  try { await browser.contextMenus.refresh(); } catch (_) {}
}

function pickTarget(info, tab) {
  if (tab.incognito) return null;
  const ctx = recentContext.get(tab.id);
  const fresh = ctx && Date.now() - ctx.received < 5000;
  const hasVideo = info.mediaType === "video" || info.contexts?.includes("video") || (fresh && ctx.hasVideo);
  if (!hasVideo) return null;
  const pageUrl = secureURL((fresh && (ctx.targetUrl || ctx.pageUrl)) || info.pageUrl || tab.url);
  const mediaUrl = secureURL((fresh && ctx.mediaUrl) || info.srcUrl || "");
  if (!pageUrl && !mediaUrl) return null;
  return {
    url: pageUrl,
    mediaUrl,
    preferMedia: Boolean(fresh && ctx.preferMedia)
  };
}

function secureURL(value) {
  try {
    const parsed = new URL(String(value || ""));
    return parsed.protocol === "https:" ? parsed.href : "";
  } catch (_) {
    return "";
  }
}

function probeCacheKey(target) {
  return JSON.stringify([target.url || "", target.mediaUrl || "", Boolean(target.preferMedia)]);
}

browser.contextMenus.onShown.addListener(async (info, tab) => {
  if (!tab?.id) return;
  const target = pickTarget(info, tab);
  await browser.contextMenus.update(ROOT, { visible: Boolean(target) });
  if (!target) {
    try { await browser.contextMenus.refresh(); } catch (_) {}
    return;
  }

  const cacheKey = probeCacheKey(target);
  const current = activeMenu.get(tab.id);
  if (current?.cacheKey === cacheKey && current.scanning) {
    await updateSlots([], true);
    return;
  }

  const cached = probeCache.get(cacheKey);
  if (cached && Date.now() - cached.time < 90_000) {
    activeMenu.set(tab.id, { target, probe: cached.probe, cacheKey, scanning: false });
    await updateSlots(cached.probe.options || []);
    return;
  }

  activeMenu.set(tab.id, { target, probe: null, cacheKey, scanning: false });
  await updateSlots([SCAN_OPTION]);
});

browser.contextMenus.onClicked.addListener(async (info, tab) => {
  if (!tab?.id || tab.incognito || !SLOT_IDS.includes(String(info.menuItemId))) return;
  const state = activeMenu.get(tab.id);
  if (!state?.target) {
    toast(tab.id, "Download option expired. Right-click the video again.", "error");
    return;
  }

  const index = SLOT_IDS.indexOf(String(info.menuItemId));
  if (!state.probe) {
    if (index !== 0 || state.scanning) return;
    state.scanning = true;
    await updateSlots([], true);
    toast(tab.id, "Scanning the selected video with the local helper…", "info");
    try {
      const probe = await native({
        action: "probe",
        url: state.target.url,
        media_url: state.target.mediaUrl,
        prefer_media: state.target.preferMedia
      });
      if (!probe?.ok) throw new Error(probe?.error || "No downloadable video was found.");
      probeCache.set(state.cacheKey, { time: Date.now(), probe });
      activeMenu.set(tab.id, { ...state, probe, scanning: false });
      await updateSlots(probe.options || []);
      toast(tab.id, "Scan complete. Right-click the video again and choose a quality.", "success");
    } catch (error) {
      activeMenu.set(tab.id, { ...state, scanning: false });
      await updateSlots([SCAN_OPTION]);
      const detail = String(error?.message || error || "No downloadable formats found").replace(/\s+/g, " ").slice(0, 108);
      toast(tab.id, `Video scan failed: ${detail}`, "error");
    }
    return;
  }

  if (!state.probe.options) {
    toast(tab.id, "Download options expired. Scan the video again.", "error");
    return;
  }

  const option = state.probe.options[index];
  if (!option || option.disabled || option.id === "error") return;

  const destination = "C:\\Users\\<you>\\Downloads";
  toast(tab.id, `Downloading ${option.label}… Keep Firefox open.`, "info");
  notify("Universal Video Downloader", `Downloading ${option.label}…`);

  try {
    const result = await native({
      action: "download",
      url: state.probe.resolved_url || state.target.url,
      media_url: state.target.mediaUrl,
      option: option.id,
      use_cookies: Boolean(state.probe.used_cookies)
    });

    if (!result?.ok) throw new Error(result?.error || "Download failed.");
    const text = result.path ? `Finished: ${result.path}` : `Finished. Saved in ${result.downloads_dir || destination}`;
    notify("Download finished", text);
    toast(tab.id, text, "success");
  } catch (error) {
    const msg = String(error?.message || error || "Download failed");
    notify("Download failed", msg);
    toast(tab.id, `Download failed: ${msg}`, "error");
  }
});
