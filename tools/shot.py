#!/usr/bin/env python3
"""Convert a truecolor-ANSI game frame into a styled terminal-window HTML page
ready for a headless-Chrome screenshot. Usage:

    python tools/shot.py <frame.ansi> "<title>" <out.html>
"""
import re, sys, html

ESC = re.compile(r"\x1b\[([0-9;]*)m")


def parse(text):
    cells, color, bold, i = [], None, False, 0
    while i < len(text):
        m = ESC.match(text, i)
        if m:
            parts = (m.group(1) or "0").split(";")
            j = 0
            while j < len(parts):
                p = parts[j]
                if p in ("", "0"):
                    color, bold = None, False
                elif p == "1":
                    bold = True
                elif p == "38" and j + 2 < len(parts) and parts[j + 1] == "2":
                    color = (int(parts[j + 2]), int(parts[j + 3]), int(parts[j + 4]))
                    j += 4
                j += 1
            i = m.end()
        else:
            cells.append((text[i], color, bold))
            i += 1
    return cells


def to_html(cells):
    out, i, n = [], 0, len(cells)
    while i < n:
        col, bold = cells[i][1], cells[i][2]
        j = i + 1
        while j < n and cells[j][1] == col and cells[j][2] == bold:
            j += 1
        seg = html.escape("".join(c[0] for c in cells[i:j]))
        style = ("color:rgb(%d,%d,%d);" % col if col else "") + ("font-weight:700;" if bold else "")
        out.append('<span style="%s">%s</span>' % (style, seg) if style else seg)
        i = j
    return "".join(out)


TEMPLATE = """<!doctype html><html><head><meta charset="utf-8">
<link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;700&display=swap" rel="stylesheet">
<style>
  html,body{margin:0;background:transparent}
  body{display:flex;align-items:center;justify-content:center;min-height:100vh;padding:46px}
  .win{
    background:#181715;border:1px solid #34322d;border-radius:14px;overflow:hidden;
    box-shadow:0 40px 90px -30px rgba(0,0,0,.85), 0 0 0 1px rgba(0,0,0,.4);
  }
  .bar{display:flex;align-items:center;gap:14px;padding:14px 18px;border-bottom:1px solid #2c2a26}
  .dots{display:flex;gap:8px}
  .dots i{width:12px;height:12px;border-radius:50%;display:block;background:#34322d}
  .dots i:first-child{background:#cc785c}
  .title{font-family:'JetBrains Mono',monospace;font-size:13px;color:#6c6a64;letter-spacing:.02em}
  pre{
    margin:0;padding:18px 22px 22px;
    font-family:'JetBrains Mono',ui-monospace,monospace;
    font-size:15px;line-height:1.34;color:#ece6da;
    font-variant-emoji:text;white-space:pre;
  }
</style></head><body>
  <div class="win">
    <div class="bar"><div class="dots"><i></i><i></i><i></i></div><div class="title">@TITLE@</div></div>
    <pre>@CONTENT@</pre>
  </div>
</body></html>"""

ansi_path, title, out_path = sys.argv[1], sys.argv[2], sys.argv[3]
text = open(ansi_path, encoding="utf-8").read().rstrip("\n")
doc = TEMPLATE.replace("@TITLE@", html.escape(title)).replace("@CONTENT@", to_html(parse(text)))
open(out_path, "w", encoding="utf-8").write(doc)
print("wrote", out_path)
