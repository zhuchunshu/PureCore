import { createServer } from 'node:http'
import { request as httpRequest } from 'node:http'
import { request as httpsRequest } from 'node:https'
import { readFileSync, readdirSync, existsSync } from 'node:fs'
import { join, extname, dirname } from 'node:path'
import { createServer as createViteServer } from 'vite'

const __dirname = new URL('.', import.meta.url).pathname
const isProduction = process.env.NODE_ENV === 'production'

// Load .env file if present (fallback for standalone SSR usage)
try {
  const envPath = join(dirname(__dirname), '..', '.env')
  if (existsSync(envPath)) {
    const envContent = readFileSync(envPath, 'utf-8')
    for (const line of envContent.split('\n')) {
      const trimmed = line.trim()
      if (trimmed && !trimmed.startsWith('#')) {
        const eqIndex = trimmed.indexOf('=')
        if (eqIndex > 0) {
          const key = trimmed.slice(0, eqIndex).trim()
          const value = trimmed.slice(eqIndex + 1).trim()
          if (!process.env[key]) {
            process.env[key] = value
          }
        }
      }
    }
  }
} catch (_) {
  // Ignore errors — .env is optional for SSR
}

const PORT = process.env.FRONTEND_PORT || process.env.SSR_PORT || 9001

// Build API proxy target from environment variables
const apiProtocol = process.env.VITE_API_PROTOCOL || 'http'
const apiHost = process.env.VITE_API_HOST || 'localhost'
const apiPort = process.env.VITE_API_PORT || process.env.BACKEND_PORT || '9002'
const apiTarget = `${apiProtocol}://${apiHost}:${apiPort}`

// Helper: proxy an API request to the backend
function proxyApiRequest(req, res) {
  return new Promise((resolve, reject) => {
    const targetUrl = new URL(req.url, apiTarget)
    const requester = apiProtocol === 'https' ? httpsRequest : httpRequest
    const proxyReq = requester(
      targetUrl,
      {
        method: req.method,
        headers: { ...req.headers, host: `${apiHost}:${apiPort}` },
      },
      (proxyRes) => {
        res.writeHead(proxyRes.statusCode, proxyRes.headers)
        proxyRes.pipe(res)
        resolve()
      }
    )
    proxyReq.on('error', (err) => {
      console.error(`[proxy] API request failed: ${err.message}`)
      reject(err)
    })
    if (req.method !== 'GET' && req.method !== 'HEAD') {
      req.pipe(proxyReq)
    } else {
      proxyReq.end()
    }
  })
}

// Load translations for SSR
// Priority: 1) web/public/lang/ (production Docker), 2) ../lang/ (development fallback)
function loadTranslations() {
  const candidates = [
    join(__dirname, 'public', 'lang'),
    join(__dirname, '..', 'lang'),
  ]
  for (const langDir of candidates) {
    try {
      if (!existsSync(langDir)) continue
      const result = {}
      for (const file of readdirSync(langDir)) {
        if (file.endsWith('.json')) {
          const locale = file.replace('.json', '')
          result[locale] = JSON.parse(readFileSync(join(langDir, file), 'utf-8'))
        }
      }
      if (Object.keys(result).length > 0) return result
    } catch (_) {
      // try next candidate
    }
  }
  console.warn('No translation files found — i18n will be unavailable')
  return {}
}
const translations = loadTranslations()

// Load project info from purecore.json
// Priority: 1) alongside server.js (Docker), 2) parent directory (dev)
let projectInfo = null
let seoDefaults = { title: 'PureCore', description: '', keywords: '' }
try {
  const paths = [
    join(__dirname, 'purecore.json'),
    join(__dirname, '..', 'purecore.json'),
  ]
  let found = false
  for (const p of paths) {
    if (existsSync(p)) {
      projectInfo = JSON.parse(readFileSync(p, 'utf-8'))
      found = true
      break
    }
  }
  if (!found) {
    throw new Error('purecore.json not found')
  }
  if (projectInfo) {
    const desc = projectInfo.description || {}
    seoDefaults = {
      title: projectInfo.name || 'PureCore',
      description: desc.en || desc.zh || '',
      keywords: (projectInfo.keywords || []).join(', '),
    }
  }
} catch (err) {
  console.warn('Could not load project info from purecore.json:', err.message)
}

// Load theme configuration: process.env.THEME > theme.config.json > 'sunset'
let configThemeName = process.env.THEME || process.env.VITE_THEME
if (!configThemeName) {
  try {
    const themeConfigPath = join(__dirname, 'theme.config.json')
    if (existsSync(themeConfigPath)) {
      const themeConfig = JSON.parse(readFileSync(themeConfigPath, 'utf-8'))
      configThemeName = themeConfig.theme || 'sunset'
    }
  } catch (err) {
    console.warn('Could not load theme config, using default "sunset":', err.message)
  }
}
if (!configThemeName) configThemeName = 'sunset'

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

/**
 * Print a beautiful startup banner with colored output.
 *
 * Uses only ANSI escape codes (no external dependencies).
 * All values are derived from runtime configuration or package.json.
 */
function printBanner() {
  const c = {
    reset:   '\x1b[0m',
    bold:    '\x1b[1m',
    dim:     '\x1b[2m',
    cyan:    '\x1b[36m',
    blue:    '\x1b[34m',
    magenta: '\x1b[35m',
    yellow:  '\x1b[33m',
    green:   '\x1b[32m',
    white:   '\x1b[37m',
    gray:    '\x1b[90m',
  }

  const name    = projectInfo?.name || 'PureCore'
  const version = projectInfo?.version ? ` v${projectInfo.version}` : ''
  const type    = projectInfo?.release_type || ''
  const tag     = type ? ` [${type}]` : ''
  const env     = isProduction ? 'production' : 'development'
  const envColor = isProduction ? c.green : c.yellow
  const theme   = configThemeName || 'sunset'

  const boxWidth = 52
  const top    = '╔' + '═'.repeat(boxWidth - 2) + '╗'
  const bottom = '╚' + '═'.repeat(boxWidth - 2) + '╝'
  const sep    = '╟' + '─'.repeat(boxWidth - 2) + '╢'

  function pad(label, value, valueColor) {
    const labelLen = label.length + 1 // +1 for the space
    const valStr = String(value)
    const full = ` ${label} ${c.dim}${valStr}${c.reset}`
    return full
  }

  const lines = [
    '',
    `${c.cyan}${c.bold}${top}${c.reset}`,
    `${c.cyan}║${c.reset}   ${c.bold}${c.magenta}${name}${c.reset}${c.gray}${version}${tag}${c.reset}${' '.repeat(Math.max(0, boxWidth - 10 - name.length - version.length - tag.length - 2))}${c.cyan}║${c.reset}`,
    `${c.cyan}║${c.reset}   ${c.gray}🚀  PureCore Client Server${' '.repeat(boxWidth - 40)}${c.cyan}║${c.reset}`,
    `${c.cyan}║${c.reset}${' '.repeat(boxWidth - 2)}${c.cyan}║${c.reset}`,
    `${c.cyan}${sep}${c.reset}`,
    `${c.cyan}║${c.reset}   ${c.bold}Local${c.reset}    ${c.green}http://localhost:${PORT}${' '.repeat(boxWidth - 22 - String(PORT).length)}${c.cyan}║${c.reset}`,
    `${c.cyan}║${c.reset}   ${c.bold}API${c.reset}      ${c.gray}→ ${apiTarget}${' '.repeat(Math.max(0, boxWidth - 16 - apiTarget.length))}${c.cyan}║${c.reset}`,
    `${c.cyan}║${c.reset}${' '.repeat(boxWidth - 2)}${c.cyan}║${c.reset}`,
    `${c.cyan}║${c.reset}   ${c.bold}Theme${c.reset}    ${theme}${' '.repeat(boxWidth - 17 - theme.length)}${c.cyan}║${c.reset}`,
    `${c.cyan}║${c.reset}   ${c.bold}Mode${c.reset}     ${envColor}${env}${' '.repeat(boxWidth - 17 - env.length)}${c.cyan}║${c.reset}`,
    `${c.cyan}${bottom}${c.reset}`,
    '',
  ]

  console.log(lines.join('\n'))
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

      // Proxy API requests to backend
      if (pathname.startsWith('/api/')) {
        try {
          const targetUrl = new URL(req.url, apiTarget)
          // Build clean headers for forwarding. Remove hop-by-hop headers
          // (host, connection, etc.) and set the correct target host.
          const proxyHeaders = {}
          for (const [key, value] of req.headers.entries()) {
            const lower = key.toLowerCase()
            if (['host', 'connection', 'keep-alive', 'transfer-encoding', 'te', 'trailer', 'upgrade'].includes(lower)) {
              continue
            }
            proxyHeaders[key] = value
          }
          proxyHeaders['host'] = `${apiHost}:${apiPort}`
          const response = await fetch(targetUrl, {
            method: req.method,
            headers: proxyHeaders,
            body: req.method !== 'GET' && req.method !== 'HEAD' ? await req.arrayBuffer() : undefined,
          })
          const respHeaders = {}
          response.headers.forEach((value, key) => { respHeaders[key] = value })
          return new Response(response.body, {
            status: response.status,
            headers: respHeaders,
          })
        } catch (err) {
          console.error('[proxy] API proxy error:', err.message)
          return new Response('Backend unreachable', { status: 502 })
        }
      }

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
        const { html, statusCode = 200, pageTitle, pageDescription } = await render(pathname, { locale, translations, projectInfo })
        const finalHtml = template
          .replace('<html', `<html data-theme="${ssrTheme}"`)
          .replace('<!--ssr-outlet-->', html)
          .replace('<!--seo-title-->', pageTitle || seoDefaults.title)
          .replace('<!--seo-description-->', pageDescription || seoDefaults.description)
          .replace('<!--seo-keywords-->', seoDefaults.keywords)
          .replace('<!--preload-links-->', cssHref ? `<link rel="stylesheet" href="${cssHref}" />` : '')
        return new Response(finalHtml, { headers: { 'Content-Type': 'text/html' }, status: statusCode })
      } catch (err) {
        console.error('✗ SSR error:', err)
        return new Response('Internal Server Error', { status: 500 })
      }
    },
  })

  printBanner()

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

      // Proxy API requests to backend
      if (url.startsWith('/api/')) {
        try {
          await proxyApiRequest(req, res)
        } catch {
          if (!res.headersSent) {
            res.writeHead(502, { 'Content-Type': 'text/plain' })
            res.end('Backend unreachable')
          }
        }
        return
      }

      // Render the app server-side for page routes
      try {
        let template = readFileSync(join(__dirname, 'index.html'), 'utf-8')
        template = await vite.transformIndexHtml(url, template)

        const { render } = await vite.ssrLoadModule('/src/entry-server.js')
        const locale = detectLocale(req)
        const ssrTheme = detectTheme(req)
        const { html, statusCode = 200, pageTitle, pageDescription } = await render(url, { locale, translations, projectInfo })

        const finalHtml = template
          .replace('<html', `<html data-theme="${ssrTheme}"`)
          .replace('<!--ssr-outlet-->', html)
          .replace('<!--seo-title-->', pageTitle || seoDefaults.title)
          .replace('<!--seo-description-->', pageDescription || seoDefaults.description)
          .replace('<!--seo-keywords-->', seoDefaults.keywords)
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
    printBanner()
  })
}
