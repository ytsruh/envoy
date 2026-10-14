# Envoy website (Zola)

Static landing page. Built with [Zola](https://www.getzola.org/), styled with Tailwind (CDN for now).

## Prereqs

```bash
brew install zola
# or: cargo install zola
```

## Dev

```bash
make website-serve   # http://127.0.0.1:1111
```

## Build

```bash
make website-build   # output in website/public/
```

`website/public/` is gitignored. Future docs will live under `website/content/docs/`.
Tailwind is currently the v3 browser CDN (ported 1:1 from the Go `home.templ`); migrate to the Tailwind v4 standalone CLI when the site needs production CSS.
