# Design System Tokens

This document defines the design system tokens used for generating CSS variables in loko documentation.

## Color Palette

### Primary Colors
- `primary`: #2563eb (Blue)
- `primary-dark`: #1e40af (Dark Blue)
- `primary-light`: #dbeafe (Light Blue)

### Semantic Colors
- `success`: #10b981 (Green)
- `warning`: #f59e0b (Amber)
- `error`: #ef4444 (Red)

### Neutral Colors
- `text`: #1f2937 (Dark Gray)
- `text-light`: #6b7280 (Medium Gray)
- `bg`: #ffffff (White)
- `bg-alt`: #f9fafb (Off White)
- `border`: #e5e7eb (Light Gray)

### Dark Mode Overrides
- `text` (dark): #f3f4f6 (Light Gray)
- `text-light` (dark): #d1d5db (Medium Gray)
- `bg` (dark): #111827 (Very Dark)
- `bg-alt` (dark): #1f2937 (Dark Gray)
- `border` (dark): #374151 (Dark Gray)
- `primary-light` (dark): #1e3a8a (Dark Blue)

## Typography

### Font Families
- `font-family`: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif
- `font-mono`: "Menlo", "Monaco", "Courier New", monospace

### Font Sizes
- `h1`: 2.25rem (36px)
- `h2`: 1.875rem (30px)
- `h3`: 1.25rem (20px)
- `h4`: 1rem (16px)
- `body`: 1rem (16px)
- `small`: 0.875rem (14px)
- `xs`: 0.75rem (12px)

### Font Weights
- `normal`: 400
- `medium`: 500
- `semibold`: 600
- `bold`: 700

## Spacing Scale

- `xs`: 0.25rem (4px)
- `sm`: 0.5rem (8px)
- `md`: 1rem (16px)
- `lg`: 1.5rem (24px)
- `xl`: 2rem (32px)
- `2xl`: 3rem (48px)

## Border & Radius

- `border-radius`: 0.375rem (6px)

## Shadows

- `shadow-sm`: 0 1px 2px 0 rgba(0, 0, 0, 0.05)
- `shadow-md`: 0 4px 6px -1px rgba(0, 0, 0, 0.1)
- `shadow-lg`: 0 10px 15px -3px rgba(0, 0, 0, 0.1)

## Responsive Breakpoints

- `sm`: 640px
- `md`: 768px
- `lg`: 1024px
- `xl`: 1280px
- `2xl`: 1536px

## Component Styles

### Buttons
- `padding`: var(--spacing-sm) var(--spacing-md)
- `background`: var(--color-primary)
- `color`: white
- `border-radius`: var(--border-radius)
- `font-weight`: 500
- `transition`: background-color 0.2s

### Cards
- `padding`: var(--spacing-lg)
- `background`: var(--color-bg-alt)
- `border`: 1px solid var(--color-border)
- `border-radius`: var(--border-radius)
- `box-shadow`: var(--shadow-sm)

### Tags
- `padding`: var(--spacing-xs) var(--spacing-md)
- `background`: var(--color-primary-light)
- `color`: var(--color-primary)
- `border-radius`: var(--border-radius)
- `font-size`: 0.75rem
- `font-weight`: 500
- `text-transform`: uppercase
- `letter-spacing`: 0.5px

## Line Heights

- `tight`: 1.25
- `normal`: 1.5
- `relaxed`: 1.625
- `loose`: 2

## Letter Spacing

- `tight`: -0.02em
- `normal`: 0em
- `wide`: 0.02em
- `wider`: 0.05em
- `widest`: 0.1em
