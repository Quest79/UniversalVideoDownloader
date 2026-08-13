package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Request struct {
	RequestID   uint64 `json:"request_id,omitempty"`
	Action      string `json:"action"`
	URL         string `json:"url"`
	MediaURL    string `json:"media_url"`
	Option      string `json:"option"`
	UseCookies  bool   `json:"use_cookies"`
	PreferMedia bool   `json:"prefer_media"`
	UserAgent   string `json:"user_agent"`
}

type Option struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Disabled bool   `json:"disabled,omitempty"`
}

type Response struct {
	BackendVersion string   `json:"backend_version"`
	RequestID      uint64   `json:"request_id,omitempty"`
	Event          string   `json:"event,omitempty"`
	DownloadID     string   `json:"download_id,omitempty"`
	DownloadsDir   string   `json:"downloads_dir,omitempty"`
	Path           string   `json:"path,omitempty"`
	LogPath        string   `json:"log_path,omitempty"`
	OK             bool     `json:"ok"`
	Error          string   `json:"error,omitempty"`
	Message        string   `json:"message,omitempty"`
	Title          string   `json:"title,omitempty"`
	Extractor      string   `json:"extractor,omitempty"`
	UsedCookies    bool     `json:"used_cookies,omitempty"`
	ResolvedURL    string   `json:"resolved_url,omitempty"`
	Options        []Option `json:"options,omitempty"`
}

type Format struct {
	FormatID string  `json:"format_id"`
	Ext      string  `json:"ext"`
	VCodec   string  `json:"vcodec"`
	ACodec   string  `json:"acodec"`
	Height   float64 `json:"height"`
	FPS      float64 `json:"fps"`
	TBR      float64 `json:"tbr"`
}

type Info struct {
	ID        string          `json:"id"`
	Title     string          `json:"title"`
	Extractor string          `json:"extractor"`
	Formats   []Format        `json:"formats"`
	Entries   json.RawMessage `json:"entries"`
}

func exeDir() string {
	p, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(p)
}

func tool(name string) string {
	return filepath.Join(exeDir(), name)
}

func validHTTPURL(s string) bool {
	u, err := url.Parse(strings.TrimSpace(s))
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

func runYTDLP(ctx context.Context, useCookies bool, userAgent string, args ...string) ([]byte, error) {
	ytdlp := tool("yt-dlp.exe")
	if _, err := os.Stat(ytdlp); err != nil {
		return nil, fmt.Errorf("yt-dlp.exe is missing. Run INSTALL_HELPER.bat again")
	}
	base := []string{
		"--ignore-config",
		"--no-warnings",
		"--no-progress",
		"--no-playlist",
		"--socket-timeout", "15",
		"--extractor-retries", "2",
		"--fragment-retries", "2",
	}
	// Current yt-dlp requires an external JavaScript runtime for full YouTube
	// extraction. Keep Deno beside yt-dlp.exe and tell yt-dlp exactly where it is.
	deno := tool("deno.exe")
	if _, err := os.Stat(deno); err == nil {
		base = append(base, "--js-runtimes", "deno:"+deno)
	}
	if useCookies {
		base = append(base, "--cookies-from-browser", "firefox")
	}
	if strings.TrimSpace(userAgent) != "" {
		base = append(base, "--user-agent", strings.TrimSpace(userAgent))
	}
	base = append(base, args...)
	cmd := exec.CommandContext(ctx, ytdlp, base...)
	cmd.Dir = exeDir()
	out, err := cmd.CombinedOutput()
	if err != nil {
		text := strings.TrimSpace(string(out))
		if len(text) > 1400 {
			text = text[len(text)-1400:]
		}
		if text == "" {
			text = err.Error()
		}
		return nil, errors.New(text)
	}
	return out, nil
}

func extractFirstInfo(raw []byte) (Info, error) {
	var info Info
	if err := json.Unmarshal(raw, &info); err != nil {
		return info, fmt.Errorf("could not parse yt-dlp metadata: %w", err)
	}
	if len(info.Formats) > 0 {
		return info, nil
	}
	if len(info.Entries) > 0 && string(info.Entries) != "null" {
		var entries []json.RawMessage
		if json.Unmarshal(info.Entries, &entries) == nil {
			for _, e := range entries {
				var child Info
				if json.Unmarshal(e, &child) == nil && len(child.Formats) > 0 {
					return child, nil
				}
			}
		}
	}
	return info, nil
}

func stripRangeParams(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return raw
	}
	q := u.Query()
	for _, key := range []string{"bytestart", "byteend"} {
		q.Del(key)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

type probeAttempt struct {
	candidate   string
	cookies     bool
	impersonate bool
	timeout     time.Duration
	referer     string
}

func runProbeAttempt(a probeAttempt, userAgent string) (Info, error) {
	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	defer cancel()
	args := []string{"--dump-single-json", "--skip-download"}
	if a.referer != "" && a.referer != a.candidate {
		args = append(args, "--referer", a.referer)
	}
	if a.impersonate {
		args = append(args, "--impersonate", "chrome")
	}
	out, err := runYTDLP(ctx, a.cookies, userAgent, append(args, a.candidate)...)
	if err != nil {
		return Info{}, err
	}
	info, err := extractFirstInfo(out)
	if err != nil {
		return Info{}, err
	}
	if len(info.Formats) == 0 {
		return Info{}, errors.New("yt-dlp found the page but returned no downloadable formats")
	}
	return info, nil
}

func probeURL(target string, mediaURL string, preferMedia bool, userAgent string) (Info, bool, string, error) {
	target = strings.TrimSpace(target)
	mediaURL = stripRangeParams(strings.TrimSpace(mediaURL))
	pageOK := validHTTPURL(target)
	mediaOK := validHTTPURL(mediaURL)
	if !pageOK && !mediaOK {
		return Info{}, false, "", errors.New("invalid video URL")
	}

	// Prefer the actual post/watch page for known sites because it exposes all
	// qualities. For generic sites prefer the clicked media URL first.
	ordered := []string{}
	if preferMedia && mediaOK {
		ordered = append(ordered, mediaURL)
	}
	if pageOK {
		ordered = append(ordered, target)
	}
	if mediaOK && (len(ordered) == 0 || ordered[0] != mediaURL) && mediaURL != target {
		ordered = append(ordered, mediaURL)
	}

	var errorsSeen []string
	for _, candidate := range ordered {
		isDirect := mediaOK && candidate == mediaURL && candidate != target
		attempts := []probeAttempt{
			{candidate: candidate, cookies: false, impersonate: false, timeout: 30 * time.Second, referer: target},
			{candidate: candidate, cookies: true, impersonate: false, timeout: 45 * time.Second, referer: target},
		}
		// Browser impersonation is most useful for page extractors. Direct CDN
		// URLs generally need the page referer/cookies instead.
		if !isDirect {
			attempts = append(attempts, probeAttempt{candidate: candidate, cookies: false, impersonate: true, timeout: 30 * time.Second, referer: target})
		}

		for _, a := range attempts {
			info, err := runProbeAttempt(a, userAgent)
			if err == nil {
				return info, a.cookies, candidate, nil
			}
			msg := strings.TrimSpace(err.Error())
			if msg != "" {
				if len(msg) > 500 {
					msg = msg[len(msg)-500:]
				}
				errorsSeen = append(errorsSeen, msg)
			}
		}
	}

	if len(errorsSeen) == 0 {
		return Info{}, false, "", errors.New("no downloadable video source was found")
	}
	// Return the most recent extractor error, which is usually the most useful,
	// and persist all attempts for diagnosis without asking the user to guess.
	logText := strings.Join(errorsSeen, "\n\n--- fallback ---\n\n")
	_ = os.WriteFile(filepath.Join(exeDir(), "last-probe.log"), []byte(logText), 0644)
	return Info{}, false, "", errors.New(errorsSeen[len(errorsSeen)-1])
}

func codecPresent(v string) bool {
	v = strings.TrimSpace(strings.ToLower(v))
	return v != "" && v != "none"
}

func buildOptions(info Info) []Option {
	heights := map[int]bool{}
	maxFPS := map[int]int{}
	hasVideo := false
	hasAudio := false
	hasMP4 := false

	for _, f := range info.Formats {
		v := codecPresent(f.VCodec)
		a := codecPresent(f.ACodec)
		if v {
			hasVideo = true
			h := int(f.Height + 0.5)
			if h > 0 {
				heights[h] = true
				if int(f.FPS+0.5) > maxFPS[h] {
					maxFPS[h] = int(f.FPS + 0.5)
				}
			}
			if strings.EqualFold(f.Ext, "mp4") {
				hasMP4 = true
			}
		}
		if a {
			hasAudio = true
		}
	}

	options := []Option{}
	if hasVideo {
		options = append(options, Option{ID: "best", Label: "★ Best quality • video + audio"})
		if hasMP4 {
			options = append(options, Option{ID: "mp4", Label: "Best MP4 • most compatible"})
		}

		hs := make([]int, 0, len(heights))
		for h := range heights {
			hs = append(hs, h)
		}
		sort.Sort(sort.Reverse(sort.IntSlice(hs)))
		// Collapse unusual near-duplicate heights while preserving actual source options.
		seen := map[int]bool{}
		for _, h := range hs {
			bucket := h
			standards := []int{4320, 2160, 1440, 1080, 720, 576, 480, 360, 240, 144}
			for _, s := range standards {
				if h >= s-24 && h <= s+24 {
					bucket = s
					break
				}
			}
			if seen[bucket] {
				continue
			}
			seen[bucket] = true
			fps := maxFPS[h]
			label := fmt.Sprintf("%dp • video + audio", bucket)
			if fps > 30 {
				label = fmt.Sprintf("%dp%d • video + audio", bucket, fps)
			}
			options = append(options, Option{ID: "h:" + strconv.Itoa(bucket), Label: label})
			if len(options) >= 11 {
				break
			}
		}
	}
	if hasAudio && len(options) < 12 {
		options = append(options, Option{ID: "audio", Label: "Audio only • best quality"})
	}
	if len(options) == 0 {
		options = append(options, Option{ID: "best", Label: "★ Best available file"})
	}
	if len(options) > 12 {
		options = options[:12]
	}
	return options
}

func selector(option string) (format string, audioOnly bool, err error) {
	switch option {
	case "best":
		return "bestvideo+bestaudio/best", false, nil
	case "mp4":
		return "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/bestvideo+bestaudio/best", false, nil
	case "audio":
		return "bestaudio/best", true, nil
	default:
		if strings.HasPrefix(option, "h:") {
			n, e := strconv.Atoi(strings.TrimPrefix(option, "h:"))
			if e != nil || n < 100 || n > 10000 {
				return "", false, errors.New("invalid quality")
			}
			s := fmt.Sprintf("bestvideo[height<=%d][ext=mp4]+bestaudio[ext=m4a]/best[height<=%d][ext=mp4]/bestvideo[height<=%d]+bestaudio/best[height<=%d]", n, n, n, n)
			return s, false, nil
		}
	}
	return "", false, errors.New("unknown download option")
}

func downloadsDir() string {
	home := os.Getenv("USERPROFILE")
	if home == "" {
		home, _ = os.UserHomeDir()
	}
	dir := filepath.Join(home, "Downloads")
	_ = os.MkdirAll(dir, 0755)
	return dir
}

func tailFile(path string, max int) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	text := strings.TrimSpace(string(data))
	if len(text) > max {
		text = text[len(text)-max:]
	}
	return text
}

func copyFile(src, dst string) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	_, _ = io.Copy(out, in)
	_ = out.Close()
}

type Emitter func(Response)

func startDownload(req Request, emit Emitter) Response {
	target := req.URL
	if !validHTTPURL(target) && validHTTPURL(req.MediaURL) {
		target = req.MediaURL
	}
	if !validHTTPURL(target) {
		return Response{OK: false, Error: "invalid video URL"}
	}

	format, audioOnly, err := selector(req.Option)
	if err != nil {
		return Response{OK: false, Error: err.Error()}
	}

	dir := downloadsDir()
	downloadID := fmt.Sprintf("%d-%d", time.Now().UnixNano(), os.Getpid())
	logPath := filepath.Join(exeDir(), "last-download.log")
	finalPathFile := filepath.Join(exeDir(), "download-"+downloadID+".path")
	_ = os.Remove(finalPathFile)

	args := []string{
		"--ignore-config",
		"--no-warnings",
		"--newline",
		"--no-playlist",
		"--windows-filenames",
		"--trim-filenames", "180",
		"--concurrent-fragments", "4",
		"--ffmpeg-location", exeDir(),
		"--print-to-file", "after_move:filepath", finalPathFile,
		"-f", format,
		"-o", filepath.Join(dir, "%(title).180B [%(id)s].%(ext)s"),
	}
	if _, statErr := os.Stat(tool("deno.exe")); statErr == nil {
		args = append(args, "--js-runtimes", "deno:"+tool("deno.exe"))
	}
	if strings.TrimSpace(req.UserAgent) != "" {
		args = append(args, "--user-agent", strings.TrimSpace(req.UserAgent))
	}
	if req.UseCookies {
		args = append(args, "--cookies-from-browser", "firefox")
	}
	if audioOnly {
		args = append(args, "--extract-audio", "--audio-format", "m4a")
	} else {
		args = append(args, "--merge-output-format", "mp4/mkv")
	}
	args = append(args, target)

	log, e := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if e != nil {
		return Response{OK: false, Error: e.Error(), DownloadsDir: dir, LogPath: logPath}
	}

	cmd := exec.Command(tool("yt-dlp.exe"), args...)
	cmd.Dir = exeDir()
	cmd.Stdout = log
	cmd.Stderr = log
	err = cmd.Run()
	_ = log.Close()

	if err != nil {
		msg := tailFile(logPath, 1600)
		if msg == "" {
			msg = err.Error()
		}
		return Response{OK: false, Error: msg, DownloadsDir: dir, LogPath: logPath}
	}

	finalPath := ""
	if b, readErr := os.ReadFile(finalPathFile); readErr == nil {
		lines := strings.FieldsFunc(strings.TrimSpace(string(b)), func(r rune) bool { return r == '\r' || r == '\n' })
		if len(lines) > 0 {
			finalPath = strings.TrimSpace(lines[len(lines)-1])
		}
	}
	_ = os.Remove(finalPathFile)
	return Response{
		OK: true, Message: "Download finished.", DownloadID: downloadID,
		Path: finalPath, DownloadsDir: dir, LogPath: logPath,
	}
}

func handle(req Request, emit Emitter) Response {
	switch req.Action {
	case "ping":
		return Response{OK: true, Message: "native helper is running"}
	case "probe":
		info, cookies, resolved, err := probeURL(req.URL, req.MediaURL, req.PreferMedia, req.UserAgent)
		if err != nil {
			return Response{OK: false, Error: err.Error()}
		}
		return Response{
			OK: true, Title: info.Title, Extractor: info.Extractor,
			UsedCookies: cookies, ResolvedURL: resolved, Options: buildOptions(info),
		}
	case "download":
		return startDownload(req, emit)
	default:
		return Response{OK: false, Error: "unknown action"}
	}
}

func readMessage(r io.Reader) (Request, error) {
	var req Request
	var size uint32
	if err := binary.Read(r, binary.LittleEndian, &size); err != nil {
		return req, err
	}
	if size == 0 || size > 8*1024*1024 {
		return req, errors.New("invalid message size")
	}
	data := make([]byte, size)
	if _, err := io.ReadFull(r, data); err != nil {
		return req, err
	}
	if err := json.Unmarshal(data, &req); err != nil {
		return req, err
	}
	return req, nil
}

func writeMessage(w io.Writer, resp Response) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, uint32(len(data))); err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func main() {
	const backendVersion = "4.4.0"
	if len(os.Args) > 1 && os.Args[1] == "--self-test" {
		fmt.Printf("{\"ok\":true,\"helper\":\"uvd-host\",\"version\":\"%s\",\"dir\":%q}\n", backendVersion, exeDir())
		return
	}

	r := bufio.NewReader(os.Stdin)
	req, err := readMessage(r)
	if err != nil {
		_ = writeMessage(os.Stdout, Response{BackendVersion: backendVersion, OK: false, Error: err.Error()})
		return
	}

	emit := func(resp Response) {}
	resp := handle(req, emit)
	resp.BackendVersion = backendVersion
	resp.RequestID = req.RequestID
	_ = writeMessage(os.Stdout, resp)
}
