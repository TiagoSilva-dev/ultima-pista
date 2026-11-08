import { useId } from 'react'

const INK = '#e8dcc2'

/**
 * Dark, lamp-lit "engraving" scene behind the case paper. The background (lamp,
 * light cone, table) is shared; the object on the table depends on the case.
 */
export function CaseArt({ kind }: { kind: string }) {
  const uid = useId().replace(/:/g, '')
  const glow = `glow-${uid}`
  const cone = `cone-${uid}`
  const table = `table-${uid}`
  const grain = `grain-${uid}`
  const hatch = `hatch-${uid}`

  return (
    <svg viewBox="0 0 400 320" preserveAspectRatio="xMaxYMid slice" className="case-art-svg" aria-hidden="true">
      <defs>
        <radialGradient id={glow} cx="74%" cy="36%" r="70%">
          <stop offset="0" stopColor="#6b5a42" />
          <stop offset="0.35" stopColor="#2e2620" />
          <stop offset="1" stopColor="#100d0b" />
        </radialGradient>
        <linearGradient id={cone} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" stopColor="#ffe7b0" stopOpacity="0.42" />
          <stop offset="1" stopColor="#ffe7b0" stopOpacity="0" />
        </linearGradient>
        <linearGradient id={table} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" stopColor="#4a3c2f" />
          <stop offset="1" stopColor="#120f0d" />
        </linearGradient>
        <pattern id={hatch} width="5" height="5" patternUnits="userSpaceOnUse" patternTransform="rotate(35)">
          <line x1="0" y1="0" x2="0" y2="5" stroke="#000" strokeWidth="1.2" opacity="0.35" />
        </pattern>
        <filter id={grain}>
          <feTurbulence type="fractalNoise" baseFrequency="0.85" numOctaves="2" stitchTiles="stitch" />
          <feColorMatrix values="0 0 0 0 0.9  0 0 0 0 0.85  0 0 0 0 0.75  0 0 0 0.55 0" />
        </filter>
      </defs>

      <rect width="400" height="320" fill={`url(#${glow})`} />
      <rect width="400" height="320" fill={`url(#${hatch})`} />

      {kind === 'glass' || kind === '' ? <Bottles /> : null}
      {kind === 'library' ? <Shelves /> : null}

      {/* table */}
      <path d="M0 246 L400 232 L400 320 L0 320 Z" fill={`url(#${table})`} />
      <path d="M0 246 L400 232" stroke="#8a7458" strokeWidth="1.2" opacity="0.6" />
      <path d="M0 262 L400 250 M0 282 L400 272 M0 302 L400 296" stroke="#000" strokeWidth="1" opacity="0.35" />

      {/* hanging lamp */}
      <path d="M300 0 V64" stroke="#0a0807" strokeWidth="2.5" />
      <path d="M294 64 h12 v10 h-12 z" fill="#0d0b0a" />
      <path d="M252 112 C258 84 276 72 300 72 C324 72 342 84 348 112 Z" fill="#0c0a09" stroke="#3b3128" strokeWidth="1.5" />
      <path d="M262 104 C268 88 282 80 296 79" stroke="#6b5a45" strokeWidth="1.2" fill="none" opacity="0.7" />
      <path d="M252 112 L170 300 L400 300 L400 210 L348 112 Z" fill={`url(#${cone})`} />
      <ellipse cx="300" cy="112" rx="48" ry="6" fill="#fff1cc" />
      <ellipse cx="300" cy="113" rx="30" ry="3.5" fill="#fffaf0" />

      <g fill="none" stroke={INK} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <SceneObject kind={kind} />
      </g>

      <rect width="400" height="320" filter={`url(#${grain})`} opacity="0.18" />
    </svg>
  )
}

function Bottles() {
  return (
    <g fill="#1a1512" stroke="#5a4b3b" strokeWidth="1" opacity="0.85">
      <path d="M338 150 h8 v14 c6 4 8 10 8 18 v54 h-24 v-54 c0-8 2-14 8-18 z" />
      <path d="M362 160 h7 v12 c5 4 7 9 7 16 v48 h-21 v-48 c0-7 2-12 7-16 z" />
      <path d="M384 140 h8 v18 c6 4 8 10 8 18 v60 h-24 v-60 c0-8 2-14 8-18 z" />
      <path d="M320 178 h7 v10 c4 3 6 8 6 14 v34 h-19 v-34 c0-6 2-11 6-14 z" />
    </g>
  )
}

function Shelves() {
  return (
    <g stroke="#5a4b3b" strokeWidth="1" opacity="0.7">
      <path d="M330 130 H400 M330 180 H400" />
      {[334, 342, 352, 358, 368, 378, 386, 394].map((x, i) => (
        <rect key={x} x={x} y={i % 2 ? 140 : 136} width={i % 3 ? 7 : 9} height={i % 2 ? 40 : 44} fill="#1a1512" />
      ))}
    </g>
  )
}

function SceneObject({ kind }: { kind: string }) {
  switch (kind) {
    case 'library':
      return (
        <>
          <path d="M228 238 h120 v-14 h-120 z" fill="#2a221b" />
          <path d="M236 224 h104 v-14 h-104 z" fill="#2a221b" />
          <path d="M232 210 h112 v-12 h-112 z" fill="#2a221b" />
          <path d="M240 232 h86 M244 218 h80 M240 205 h90" strokeWidth="1" opacity="0.6" />
          <path d="M244 196 c18 -10 34 -10 44 -2 c10 -8 26 -8 44 2 l-4 -36 c-16 -8 -30 -8 -40 0 c-10 -8 -24 -8 -40 0 z" fill="#2c241c" />
          <path d="M288 194 v-34 M254 176 h24 M254 184 h22 M298 176 h24 M298 184 h22" strokeWidth="1" opacity="0.7" />
        </>
      )
    case 'train':
      return (
        <>
          <rect x="226" y="160" width="132" height="76" rx="4" fill="#2a221b" />
          <path d="M226 176 h132" strokeWidth="1.2" />
          <rect x="238" y="186" width="28" height="22" fill="#1a1512" />
          <rect x="278" y="186" width="28" height="22" fill="#1a1512" />
          <rect x="318" y="186" width="28" height="22" fill="#1a1512" />
          <circle cx="252" cy="240" r="10" fill="#1a1512" />
          <circle cx="332" cy="240" r="10" fill="#1a1512" />
          <path d="M212 252 h160" strokeWidth="1.5" />
        </>
      )
    case 'photo':
      return (
        <>
          <path d="M240 236 l8 -88 h96 l8 88 z" fill="#2a221b" />
          <rect x="258" y="160" width="76" height="62" fill="#1a1512" strokeWidth="1.4" />
          <circle cx="283" cy="182" r="8" />
          <circle cx="309" cy="182" r="8" />
          <path d="M268 218 c4 -14 12 -20 15 -20 c3 0 11 6 15 20 M294 218 c4 -14 12 -20 15 -20 c3 0 11 6 15 20" strokeWidth="1.4" />
        </>
      )
    case 'garden':
      return (
        <>
          <path d="M266 236 l-8 -40 h76 l-8 40 z" fill="#2a221b" />
          <path d="M254 196 h84" />
          <path d="M296 194 c0 -30 -4 -46 -18 -60 M296 194 c0 -24 8 -40 22 -52 M296 194 c-4 -20 -18 -28 -32 -28" />
          <path d="M278 134 l-6 -6 l10 0 z M318 142 l4 -8 l4 8 z M264 166 l-8 -2 l6 -6 z" fill={INK} />
          <path d="M270 150 l4 4 M314 156 l-4 4 M286 170 l4 -3" strokeWidth="1" opacity="0.7" />
        </>
      )
    case 'letter':
      return (
        <>
          <path d="M228 238 l18 -66 l118 26 l-18 66 z" fill="#2a221b" />
          <path d="M246 172 l46 52 l72 -26" strokeWidth="1.6" />
          <circle cx="296" cy="222" r="9" fill="#7a2532" stroke="#e8dcc2" strokeWidth="1.4" />
          <path d="M292 222 h8 M296 218 v8" stroke="#e8dcc2" strokeWidth="1" />
        </>
      )
    case 'office':
      return (
        <>
          <path d="M226 240 l14 -74 h88 l-6 74 z" fill="#2e261e" />
          <path d="M250 184 h62 M248 196 h50 M246 208 h58 M244 220 h40" strokeWidth="1" opacity="0.25" />
          <path d="M338 168 l26 66" strokeWidth="3" />
          <path d="M362 230 l3 8 l-6 -3 z" fill={INK} />
        </>
      )
    default:
      // a glass of water
      return (
        <>
          <ellipse cx="300" cy="244" rx="40" ry="5" fill="#000" stroke="none" opacity="0.4" />
          <path d="M266 160 L334 160 L326 242 C326 246 274 246 274 242 Z" fill="#2c261f" fillOpacity="0.55" />
          <ellipse cx="300" cy="160" rx="34" ry="6" />
          <path d="M269 196 C280 202 320 202 331 196" strokeWidth="1.4" />
          <path d="M269.5 197 L274 242 C274 246 326 246 326 242 L330.5 197 C320 202 280 202 269.5 197 Z" fill={INK} fillOpacity="0.16" stroke="none" />
          <path d="M276 172 L282 236" strokeWidth="3" opacity="0.55" />
          <path d="M320 176 L316 232" strokeWidth="1.2" opacity="0.4" />
          <path d="M276 240 C290 244 310 244 324 240" strokeWidth="1.2" opacity="0.6" />
        </>
      )
  }
}
