# Banner Guidelines for Elite Agent Collective

## Specifications

| Property | Requirement |
|----------|-------------|
| **Dimensions** | 1280x640 pixels |
| **Aspect Ratio** | 2:1 |
| **Format** | PNG |
| **Color Space** | sRGB |
| **File Name** | `banner.png` |

## Design Concept

The banner should serve as a professional header image for the GitHub Marketplace listing.

### Required Elements

1. **Title:** "Elite Agent Collective"
2. **Tagline:** "40 Specialized AI Agents for GitHub Copilot"
3. **Visual Element:** Tier visualization or agent constellation

### Layout

```
┌────────────────────────────────────────────────────────────────┐
│                                                                │
│    ┌──────────────────────────────────────────────────────┐   │
│    │                                                      │   │
│    │         ELITE AGENT COLLECTIVE                       │   │
│    │                                                      │   │
│    │    40 Specialized AI Agents for GitHub Copilot       │   │
│    │                                                      │   │
│    └──────────────────────────────────────────────────────┘   │
│                                                                │
│    ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐              │
│    │Tier 1│ │Tier 2│ │Tier 3│ │Tier 4│ │ ...  │              │
│    └──────┘ └──────┘ └──────┘ └──────┘ └──────┘              │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

### Color Palette

| Color | Hex | Usage |
|-------|-----|-------|
| Deep Blue | `#0d1117` | Background base |
| Purple Gradient | `#1a1b4b` to `#4a1d96` | Gradient overlay |
| Gold | `#ffd700` | Accent highlights |
| White | `#ffffff` | Primary text |
| Light Gray | `#c9d1d9` | Secondary text |

### Typography

| Element | Font | Size | Weight |
|---------|------|------|--------|
| Title | Sans-serif (Inter, Roboto, SF Pro) | 72-96px | Bold |
| Tagline | Sans-serif | 32-40px | Regular |
| Tier Labels | Sans-serif | 18-24px | Medium |

## Design Elements

### Background
- Gradient from deep blue to purple
- Subtle neural network pattern overlay (optional)
- Stars or node patterns representing agents

### Tier Visualization Options

**Option A: Linear Tiers**
```
Tier 1 → Tier 2 → Tier 3 → ... → Tier 8
```

**Option B: Constellation/Network**
```
    ★ ─ ★
   /     \
  ★       ★
   \     /
    ★ ─ ★
```

**Option C: Circular/Radial**
```
       Tier 1
    Tier 8   Tier 2
   Tier 7  ●  Tier 3
    Tier 6   Tier 4
       Tier 5
```

### Agent Representation
- Small icons or nodes for each tier
- Color-coded by tier level
- Connected with subtle lines

## Do's and Don'ts

### Do's
✅ Keep text readable and accessible
✅ Use high contrast for text
✅ Maintain breathing room around elements
✅ Test on both mobile and desktop
✅ Use the specified color palette
✅ Include the GitHub Copilot context

### Don'ts
❌ Don't overcrowd with too many elements
❌ Don't use thin fonts that are hard to read
❌ Don't include small text that won't be visible
❌ Don't use copyrighted images or logos
❌ Don't add QR codes or complex graphics

## File Placement

Save the final banner as:
```
marketplace/banner.png
```

## Testing

Test the banner in these contexts:
- [ ] GitHub Marketplace listing page
- [ ] Mobile view (smaller screens)
- [ ] Repository social preview
- [ ] Dark and light mode contexts

## Creation Tools

Recommended tools for creating the banner:
- **Figma** (free, web-based)
- **Canva** (free with templates)
- **Adobe Photoshop/Illustrator** (professional)
- **GIMP** (free, open-source)

## Example Process

1. Create a 1280x640 artboard
2. Apply background gradient
3. Add title and tagline text
4. Create tier visualization
5. Add subtle decorative elements
6. Export as PNG
7. Test in various contexts
8. Optimize file size (< 1MB recommended)

## Accessibility

- Ensure 4.5:1 contrast ratio for text
- Don't rely solely on color to convey information
- Test with color blindness simulators
