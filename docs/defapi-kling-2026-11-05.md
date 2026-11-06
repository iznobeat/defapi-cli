# Kling 4.0 support — 2026-11-05

Added `defapi video kling` from the official model page:
<https://defapi.org/api/model/en/kuaishou/kling-4.0>

- Endpoint: `POST /api/kling/gen`, model `kling-4.0`, polled via `/api/task/query`;
  the finished video is `data.result.video`, which the existing `pollVideo` reads.
- Modes: `reference-to-video` (default; `image_urls` max 10, `video_urls` max 5,
  `keep_original_sound`) and `keyframe-to-video` (`first_frame_url` required,
  `keyframe_urls` max 8, `last_frame_url`; total frames capped at
  `min(duration + 1, 10)`). The CLI picks keyframe mode when any frame flag is set.
- `duration` 3–30 (default 5), `resolution` 720p/1080p (default 720p),
  `aspect_ratio` auto/16:9/9:16/1:1/21:9 (sent only when `--ratio` is given).
- Pricing listed: $0.4275 per output second. Video references are billed for
  `duration + max(input video duration, duration)` seconds. Failed tasks are free.
- Not implemented: `elements`, `voices`, and `callback_url` (structured inputs
  that do not fit simple flags).

Caveat: the page itself says "Local catalog code only; public API model ID and
pricing are unconfirmed." The model is not marked Coming Soon. Request building is
tested offline only, and no authenticated generation request was made. Run a
paid smoke test before treating this as verified end to end.
