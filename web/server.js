import { createServer } from 'node:http'
import { readFileSync, readdirSync, existsSync } from 'node:fs'
import { join, extname } from 'node:path'
import { createServer as createViteServer } from 'vite'

const __dirname = new URL('.', import.meta.url).pathname
const isProduction = process.env.NODE_ENV === 'production'
const PORT = process.env.SSR_PORT || 9001

// Load translations for SSR from web/public/lang/
const langDir = join(__dirname, '..', 'lang')
const translations = {}
for (const file of readdirSync(langDir)) {
  if (file.endsWith('.json')) {
    const locale = file.replace('.json', '')
    translations[locale] = JSON.parse(readFileSync(join(langDir, file), 'utf-8'))
  }
}

// Load project info from web/package.json under the "purecore" key
let projectInfo = null
try {
  const pkg = JSON.parse(readFileSync(join(__dirname, 'package.json'), 'utf-8'))
  projectInfo = pkg.purecore || null
} catch (err) {
  console.warn('Could not load project info from package.json:', err.message)
}

// Load theme configuration (default from config file)
let configThemeName = 'sunset'
try {
  const themeConfig = JSON.parse(readFileSync(join(__dirname, 'theme.config.json'), 'utf-8'))
  configThemeName = themeConfig.theme || 'sunset'
} catch (err) {
  console.warn('Could not load theme config, using default "sunset":', err.message)
}

// Read theme cookie from request, fallback to config default
function detectTheme(req) {
  const getCookie = (name) => {
    const cookieHeader = req.headers['cookie'] || ''
    const match = cookieHeader.match(new RegExp(`(?:^|; )${name}=([^;]*)`))
    return match ? decodeURIComponent(match[1]) : null
  }
  const cookieTheme = getCookie('purecore-theme')
  return cookieTheme || configThemeName
}

const mimeMap = {
  '.html': 'text/html', '.js': 'application/javascript', '.css': 'text/css',
  '.json': 'application/json', '.png': 'image/png', '.jpg': 'image/jpeg',
  '.jpeg': 'image/jpeg', '.svg': 'image/svg+xml', '.ico': 'image/x-icon',
  '.woff': 'font/woff', '.woff2': 'font/woff2',
}

function getMimeType(filepath) {
  return mimeMap[extname(filepath).toLowerCase()] || 'application/octet-stream'
}

function detectLocale(req) {
  // Read cookie method
  const getCookie = (name) => {
    const cookieHeader = req.headers['cookie'] || ''
    const match = cookieHeader.match(new RegExp(`(?:^|; )${name}=([^;]*)`))
    return match ? decodeURIComponent(match[1]) : null
  }

  // 1. Check explicit cookie (set by client when user switches language)
  const cookieLocale = getCookie('purecore-locale')
  if (cookieLocale === 'zh' || cookieLocale === 'en') return cookieLocale

  // 2. Fallback to Accept-Language header
  const lang = req.headers['accept-language'] || ''
  return lang.startsWith('zh') ? 'zh' : 'en'
}

// ===== PRODUCTION MODE =====
if (isProduction) {
  const clientDist = join(__dirname, 'dist', 'client')
  const template = readFileSync(join(clientDist, 'index.html'), 'utf-8')
  const { render } = await import(join(__dirname, 'dist', 'server', 'entry-server.js'))

  // Find built CSS file
  const assetsDir = join(clientDist, 'assets')
  let cssHref = ''
  if (existsSync(assetsDir)) {
    const cssFile = readdirSync(assetsDir).find(f => f.endsWith('.css'))
    if (cssFile) cssHref = `/assets/${cssFile}`
  }

  Bun.serve({
    port: PORT,
    async fetch(req) {
      const url = new URL(req.url)
      const pathname = url.pathname

      // Serve static assets (anything with a file extension)
      if (pathname !== '/' && pathname.includes('.')) {
        const filePath = join(clientDist, pathname)
        if (existsSync(filePath)) {
          return new Response(Bun.file(filePath), {
            headers: { 'Content-Type': getMimeType(filePath) },
          })
        }
        return new Response('Not Found', { status: 404 })
      }

      // SSR render for all page routes
      try {
        const locale = detectLocale({ headers: Object.fromEntries(req.headers.entries()) })
        const ssrTheme = detectTheme({ headers: Object.fromEntries(req.headers.entries()) })
        const { html, statusCode = 200 } = await render(pathname, { locale, translations, projectInfo })
        const finalHtml = template
          .replace('<html', `<html data-theme="${ssrTheme}"`)
          .replace('<!--ssr-outlet-->', html)
          .replace('<!--preload-links-->', cssHref ? `<link rel="stylesheet" href="${cssHref}" />` : '')
        return new Response(finalHtml, { headers: { 'Content-Type': 'text/html' }, status: statusCode })
      } catch (err) {
        console.error('✗ SSR error:', err)
        return new Response('Internal Server Error', { status: 500 })
      }
    },
  })

  console.log(`✓ SSR server (production) → http://localhost:${PORT}`)

// ===== DEVELOPMENT MODE =====
} else {
  const vite = await createViteServer({
    server: { middlewareMode: true },
    appType: 'custom',
  })

  // Vite dev server with proper CSS handling
  const httpServer = createServer((req, res) => {
    // Let Vite handle all asset requests (JS, CSS, HMR, images, etc.) first
    vite.middlewares(req, res, async () => {
      const url = req.url || '/'

      // Only handle page requests (HTML routes), not asset requests
      // Vite middleware already handled assets; anything reaching here is a page route
      if (url.startsWith('/api/')) {
        res.writeHead(404, { 'Content-Type': 'text/plain' })
        res.end('API not found')
        return
      }

      // Render the app server-side for page routes
      try {
        let template = readFileSync(join(__dirname, 'index.html'), 'utf-8')
        template = await vite.transformIndexHtml(url, template)

        const { render } = await vite.ssrLoadModule('/src/entry-server.js')
        const locale = detectLocale(req)
        const ssrTheme = detectTheme(req)
        const { html, statusCode = 200 } = await render(url, { locale, translations, projectInfo })

        const finalHtml = template
          .replace('<html', `<html data-theme="${ssrTheme}"`)
          .replace('<!--ssr-outlet-->', html)
        res.writeHead(statusCode, { 'Content-Type': 'text/html' })
        res.end(finalHtml)
      } catch (err) {
        vite.ssrFixStacktrace(err)
        console.error('✗ SSR error:', err)
        res.writeHead(500, { 'Content-Type': 'text/plain' })
        res.end(err.stack || 'Internal Server Error')
      }
    })
  })

  httpServer.listen(PORT, () => {
    console.log(`✓ SSR server (development) → http://localhost:${PORT}`)
  })
}
