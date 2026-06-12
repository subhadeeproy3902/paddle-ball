# tools — asset generators

Scripts that regenerate the committed images in [`../assets/`](../assets). They
need Python + Pillow and a headless Chrome. Outputs (`*.png`, `terminal.html`,
`*_preview.html`) are git-ignored; only these generators are tracked.

## Live terminal screenshot — `assets/screenshot.png`

1. Render a real game frame to truecolor ANSI:
   ```bash
   go test -tags shots ./game/ -run TestShot      # writes game/zz_play.ansi
   ```
2. Convert it to a styled terminal window, then screenshot it:
   ```bash
   python tools/shot.py game/zz_play.ansi "paddle-ball — arcade" tools/terminal.html
   chrome --headless=new --force-device-scale-factor=2 --window-size=950,610 \
     --default-background-color=00000000 --screenshot=assets/screenshot.png tools/terminal.html
   ```

## OG image — `assets/og.png`

`tools/og.html` is the 1200×630 template. Screenshot at 2× and downscale:
```bash
chrome --headless=new --force-device-scale-factor=2 --window-size=1200,630 \
  --screenshot=tools/og-2x.png tools/og.html
python -c "from PIL import Image; Image.open('tools/og-2x.png').resize((1200,630)).save('assets/og.png')"
```

## Icons — `assets/icon-512.png`, `apple-touch-icon.png`, `favicon-32.png`

`tools/icon.html` renders `assets/logo.svg` at 512; downscale the rest with Pillow.

The logo (`assets/logo.svg`) and README banner (`assets/banner.svg`) are
hand-authored vector files — no generator needed.
