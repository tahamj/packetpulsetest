#!/usr/bin/env python3
"""Build the Pingle manual-testing guide from its authored parts.

The guide is AUTHORED as ``parts/*.md`` and this script produces the two
distributed forms:

  * ``MANUAL_TESTING_GUIDE.md``  - the monolith everyone reads and greps
  * ``manual-testing-guide.html`` - the styled edition (sidebar, search,
    light/dark theme, print)

Both outputs are GENERATED. Do not hand-edit them: ``./pingletest.sh docs``
fails when either drifts from ``parts/``.

Modelled on MShop's generator, for the same reason it exists there: a
hand-maintained HTML edition drifts from the Markdown it was copied from, and
a sidebar written by hand stops listing the sections the document has.

Usage:
    python3 build/build_guide.py [--check]

    --check  render into memory and exit non-zero if either committed output is
             stale. This is what the docs suite runs.
"""

from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path

try:
    from markdown_it import MarkdownIt
    from markdown_it.common.utils import escapeHtml
except ImportError:  # pragma: no cover - actionable message beats a traceback
    sys.exit(
        "markdown-it-py is not installed.\n"
        "It is pinned in pingletest/manual-testing/build/requirements.txt and\n"
        "installed into pingletest/.venv by `./pingletest.sh docs`. By hand:\n"
        "    python3 -m venv pingletest/.venv\n"
        "    pingletest/.venv/bin/pip install -r pingletest/manual-testing/build/requirements.txt"
    )

HERE = Path(__file__).resolve().parent
ROOT = HERE.parent
PARTS_DIR = ROOT / "parts"
SHELL = HERE / "shell.html"
OUT_MD = ROOT / "MANUAL_TESTING_GUIDE.md"
OUT_HTML = ROOT / "manual-testing-guide.html"

# GitHub-style alert blocks. The icon/label text is part of the rendered
# contract — the stylesheet colours each variant off the `a-*` class.
ALERTS = {
    "NOTE": ("a-note", "ℹ️ NOTE"),
    "TIP": ("a-tip", "💡 TIP"),
    "IMPORTANT": ("a-important", "📌 IMPORTANT"),
    "WARNING": ("a-warning", "⚠️ WARNING"),
    "CAUTION": ("a-caution", "🛑 CAUTION"),
}
ALERT_RE = re.compile(r"^\[!(" + "|".join(ALERTS) + r")\]\s*(?:\n|$)")

# Matches the version stamp in the front matter, e.g. `v2026.10-PROD-v1`.
VERSION_RE = re.compile(r"`(v\d{4}\.\d{2}-PROD-v\d+)`")


# ── markdown assembly ───────────────────────────────────────────────────

def read_parts() -> str:
    """Concatenate parts/*.md in filename order.

    Filenames carry a numeric prefix so lexical order is document order.
    """
    parts = sorted(PARTS_DIR.glob("*.md"))
    if not parts:
        sys.exit(f"No part files found under {PARTS_DIR}")
    return "".join(p.read_text(encoding="utf-8") for p in parts)


# ── slugs ───────────────────────────────────────────────────────────────

def slugify(text: str) -> str:
    """GitHub-compatible anchor slug.

    Lowercase, drop everything that is not a letter/digit/space/hyphen (which
    removes emoji, em dashes, backticks, dots and parentheses), then collapse
    runs of separators. Leading/trailing hyphens are stripped so a heading that
    opens with an emoji does not produce a leading dash.
    """
    s = text.lower()
    s = re.sub(r"[^a-z0-9 \-]+", "", s)
    s = re.sub(r"[\s\-]+", "-", s)
    return s.strip("-")


def plain_text(token) -> str:
    """Flatten an inline token to its visible text, for slugging."""
    if token is None:
        return ""
    if not token.children:
        return token.content
    out = []
    for child in token.children:
        if child.type in ("text", "code_inline"):
            out.append(child.content)
        elif child.type == "softbreak":
            out.append(" ")
    return "".join(out)


# ── token post-processing ───────────────────────────────────────────────

def mark_alerts(tokens) -> None:
    """Tag `> [!TIP]` blockquotes and strip the marker from their first line."""
    for i, tok in enumerate(tokens):
        if tok.type != "blockquote_open":
            continue
        if i + 2 >= len(tokens):
            continue
        if tokens[i + 1].type != "paragraph_open" or tokens[i + 2].type != "inline":
            continue
        inline = tokens[i + 2]
        m = ALERT_RE.match(inline.content)
        if not m:
            continue
        tok.meta["alert"] = m.group(1)
        inline.content = inline.content[m.end():]
        children = inline.children or []
        # Drop the leading "[!TYPE]" text node and the softbreak after it.
        while children:
            first = children[0]
            if first.type == "text" and first.content.lstrip().startswith("[!"):
                children.pop(0)
                continue
            if first.type == "softbreak":
                children.pop(0)
            break
        inline.children = children


def assign_heading_slugs(tokens) -> list[tuple[int, str, str]]:
    """Stamp a unique slug on every heading; return (level, slug, text)."""
    seen: dict[str, int] = {}
    toc: list[tuple[int, str, str]] = []
    for i, tok in enumerate(tokens):
        if tok.type != "heading_open":
            continue
        text = plain_text(tokens[i + 1] if i + 1 < len(tokens) else None)
        base = slugify(text) or "section"
        if base in seen:
            seen[base] += 1
            slug = f"{base}-{seen[base]}"
        else:
            seen[base] = 0
            slug = base
        tok.meta["slug"] = slug
        toc.append((int(tok.tag[1]), slug, text))
    return toc


# ── renderer ────────────────────────────────────────────────────────────

def build_markdown_renderer() -> MarkdownIt:
    md = MarkdownIt("js-default", {"html": True, "linkify": False, "typographer": False})

    def heading_open(tokens, idx, options, env):
        tok = tokens[idx]
        return f'<{tok.tag} id="{tok.meta["slug"]}">'

    def heading_close(tokens, idx, options, env):
        # The matching open token carries the slug.
        for j in range(idx, -1, -1):
            if tokens[j].type == "heading_open":
                slug = tokens[j].meta["slug"]
                break
        else:  # pragma: no cover - a close without an open is malformed
            slug = ""
        anchor = (
            f'<a class="anchor" href="#{slug}" '
            f'aria-label="Link to this section">#</a>'
        )
        return f"{anchor}</{tokens[idx].tag}>\n"

    def table_open(tokens, idx, options, env):
        return '<div class="tw"><table>'

    def table_close(tokens, idx, options, env):
        return "</table></div>\n"

    def fence(tokens, idx, options, env):
        tok = tokens[idx]
        info = (tok.info or "").strip().split()
        lang = info[0] if info else ""
        if lang == "mermaid":
            # Rendered client-side by mermaid.js; must NOT be escaped into
            # <code>, which is why it gets its own branch.
            return f'<pre class="mermaid">{escapeHtml(tok.content)}</pre>\n'
        return f'<pre class="code"><code>{escapeHtml(tok.content)}</code></pre>\n'

    def blockquote_open(tokens, idx, options, env):
        kind = tokens[idx].meta.get("alert")
        if not kind:
            return "<blockquote>\n"
        cls, label = ALERTS[kind]
        return f'<div class="alert {cls}"><div class="alert-h">{label}</div>'

    def blockquote_close(tokens, idx, options, env):
        for j in range(idx, -1, -1):
            if tokens[j].type == "blockquote_open":
                kind = tokens[j].meta.get("alert")
                break
        else:  # pragma: no cover
            kind = None
        return "</div>\n" if kind else "</blockquote>\n"

    md.renderer.rules.update({
        "heading_open": heading_open,
        "heading_close": heading_close,
        "table_open": table_open,
        "table_close": table_close,
        "fence": fence,
        "blockquote_open": blockquote_open,
        "blockquote_close": blockquote_close,
    })
    return md


def render_html(markdown_text: str) -> str:
    md = build_markdown_renderer()
    tokens = md.parse(markdown_text)
    mark_alerts(tokens)
    toc = assign_heading_slugs(tokens)
    body = md.renderer.render(tokens, md.options, {})
    nav = build_nav(toc)

    version_match = VERSION_RE.search(markdown_text)
    if not version_match:
        sys.exit("Could not find a `vYYYY.MM-PROD-vN` version stamp in the guide.")
    version = version_match.group(1)
    title_version = version.split("-")[0]  # e.g. v2026.09

    shell = SHELL.read_text(encoding="utf-8")
    return (
        shell.replace("{{TITLE_VERSION}}", title_version)
        .replace("{{VERSION}}", version)
        .replace("{{NAV}}", nav)
        .replace("{{BODY}}", body)
    )


def build_nav(toc) -> str:
    """Sidebar links: every h1 becomes .n1, every h2 becomes .n2.

    Generating this from the heading list is what keeps the sidebar complete —
    the hand-maintained edition had drifted to 66 links for 70 headings.
    """
    lines = []
    for level, slug, text in toc:
        if level > 2:
            continue
        cls = "n1" if level == 1 else "n2"
        label = escapeHtml(text)
        lines.append(f'    <a class="{cls}" href="#{slug}">{label}</a>')
    return "\n".join(lines)


# ── entry point ─────────────────────────────────────────────────────────

def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument(
        "--check",
        action="store_true",
        help="exit non-zero if a committed output is stale, writing nothing",
    )
    args = ap.parse_args()

    markdown_text = read_parts()
    html = render_html(markdown_text)

    if args.check:
        stale = []
        if not OUT_MD.exists() or OUT_MD.read_text(encoding="utf-8") != markdown_text:
            stale.append(OUT_MD.name)
        if not OUT_HTML.exists() or OUT_HTML.read_text(encoding="utf-8") != html:
            stale.append(OUT_HTML.name)
        if stale:
            print(
                "STALE: " + ", ".join(stale) + "\n"
                "These files are generated from parts/*.md. Re-run:\n"
                "    pingletest/.venv/bin/python3 pingletest/manual-testing/build/build_guide.py",
                file=sys.stderr,
            )
            return 1
        print("up to date: MANUAL_TESTING_GUIDE.md, manual-testing-guide.html")
        return 0

    OUT_MD.write_text(markdown_text, encoding="utf-8")
    OUT_HTML.write_text(html, encoding="utf-8")
    print(f"wrote {OUT_MD.relative_to(ROOT.parent.parent)}  ({len(markdown_text):,} bytes)")
    print(f"wrote {OUT_HTML.relative_to(ROOT.parent.parent)}  ({len(html):,} bytes)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
