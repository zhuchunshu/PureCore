# Dashboard UI Polish — Design Spec

## Problem

The `/dashboard` and all child pages render sidebar + content **directly on the `bg-base-200` background** without a unifying card wrapper. The sidebar is a flat list of text links. Forms use bare `input-bordered` DaisyUI inputs. The layout looks unfinished.

## Design

**Glassmorphism** — glass cards, backdrop blur, subtle borders, layered depth. Everything wrapped in a card shell.

### 1. Layout Card Wrapper (`UserDashboard.vue`)

The entire `flex gap-6` layout (sidebar + `<router-view />`) is now wrapped in a **single glass card** that sits on the `bg-base-200` background:

- Outer card: `bg-base-100/60 backdrop-blur-xl border border-base-300/20 rounded-3xl shadow-xl shadow-black/5 p-6`
- Inside the card: sidebar (left) + `<router-view />` (right) in `flex gap-6`

This gives the whole dashboard a unified elevated surface feel instead of two floating islands.

### 2. Sidebar (`UserDashboard.vue`)

Replace flat link list with a premium sidebar **inside the layout card**:

```
┌──────────────────────┐
│ [Avatar]  User Name  │  ← user info block
│           email@..   │
│                      │
│  ● Overview          │  ← active: bg-primary/12 text-primary font-semibold
│    Profile           │     ring-1 ring-primary/20 shadow-sm
│    Security          │     scale-[1.02]
│    API Keys          │
│    Sessions          │     hover: bg-base-200/60
│    Integrations      │
│                      │
│ ─────────────────    │  ← subtle divider
│    Logout            │  ← text-error/60 hover:text-error
└──────────────────────┘
```

- User block at top: `AvatarInitials` component + name + email, with `bg-base-200/50 rounded-2xl p-3 mb-4`
- Nav items: `rounded-xl px-4 py-2.5 text-sm font-medium`, `transition-all duration-200`
- Active: `bg-primary/12 text-primary font-semibold ring-1 ring-primary/20 shadow-sm scale-[1.02]`
- Logout: divider + `text-error/60 hover:text-error hover:bg-error/5`
- Mobile: Same hidden behavior below md

### 3. Page Header

Move the title + "welcome back" into a **centered header section above the card**, or into the card's top area.

### 4. Profile Form (`UserProfile.vue`)

- **Card header**: `flex items-center gap-3` with `w-10 h-10 rounded-xl bg-primary/10` icon box + title + subtitle
- **Inputs**: Replace `input-bordered` with custom glass inputs:
  - `bg-base-200/50 border border-base-300/20 rounded-xl px-4 py-3 w-full`
  - Focus: `focus:bg-base-100 focus:border-primary/40 focus:ring-2 focus:ring-primary/10`
  - Placeholder: `placeholder:text-base-content/25`
  - `transition-all duration-200`
  - Inputs sit inside a white/glass card section with proper spacing
- **Bio textarea**: Same style, `min-h-[90px]`
- **Save button**: Gradient pill matching `GradientButton` style:
  - `bg-gradient-to-r from-primary to-secondary text-primary-content rounded-xl px-6 py-3 font-semibold shadow-lg shadow-primary/20`
  - `hover:shadow-primary/30 hover:scale-[1.02] transition-all duration-200`
- **Layout**: Two-column grid (`grid-cols-1 sm:grid-cols-2 gap-5`)

### 5. Security Form (`UserSecurity.vue`)

- Same glass card header pattern
- Same custom input styling
- Password visibility toggle per field (`Eye` / `EyeOff` from lucide)
- Strength indicator bar below new-password: 3 segments
  - `<8 chars: all gray`
  - `8+ chars: 1/3 filled red`
  - `10+ chars: 2/3 filled amber`
  - `12+ chars: 3/3 filled green`

### 6. Overview Page (`UserOverview.vue`)

No structural changes needed — already uses `TechCard` which has glass styling. Just needs to render inside the new card wrapper that `UserDashboard.vue` provides.

## Implementation Checklist

| # | File | Changes |
|---|------|---------|
| 1 | `web/src/pages/UserDashboard.vue` | Wrap content area in glass card shell, redesign sidebar with user block + nav items + logout |
| 2 | `web/src/components/dashboard/UserProfile.vue` | Glass card header, custom inputs, gradient save button |
| 3 | `web/src/components/dashboard/UserSecurity.vue` | Glass card header, custom inputs, password toggle, strength bar |

## Verification

1. `cd web && bun run build` — no errors (production SSR + client)
2. Visit `/dashboard` — sidebar + overview content inside glass card on bg-base-200
3. Visit `/dashboard/profile` — custom inputs with focus glow, gradient save button
4. Visit `/dashboard/security` — password toggle works, strength bar fills correctly
5. Mobile viewport — sidebar hidden, content card fills width with proper padding
6. Light/dark modes both look correct
