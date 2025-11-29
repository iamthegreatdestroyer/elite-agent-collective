# Icon Guidelines for Elite Agent Collective

## Specifications

| Property | Requirement |
|----------|-------------|
| **Dimensions** | 256x256 pixels |
| **Format** | PNG with transparency |
| **Color Space** | sRGB |
| **File Name** | `icon.png` |

## Design Concept

The icon should represent the Elite Agent Collective's core identity:

### Visual Theme
- **Primary Concept:** Interconnected neural network or constellation pattern
- **Representation:** 40 agents as interconnected nodes
- **Style:** Modern, professional, tech-forward

### Color Palette

| Color | Hex | Usage |
|-------|-----|-------|
| Deep Blue | `#1a1b4b` | Primary background gradient start |
| Purple | `#4a1d96` | Primary background gradient end |
| Gold | `#ffd700` | Accent nodes/highlights |
| White | `#ffffff` | Secondary accent nodes |

### Design Elements

1. **Central Node**
   - Larger, prominent node representing the collective
   - Glowing effect with gold or white accent

2. **Agent Nodes**
   - 8 visible nodes representing 8 tiers
   - Connected with subtle lines
   - Gradient from blue to purple

3. **Connections**
   - Thin lines connecting nodes
   - Suggests neural network or constellation
   - Subtle glow effect

### Visual Style

```
         ★
        /|\
       / | \
      ★--●--★
     /   |   \
    ★    |    ★
     \   |   /
      ★--●--★
         |
         ★

Legend:
● = Central hub (larger, brighter)
★ = Agent nodes (smaller)
- = Connection lines
```

## Do's and Don'ts

### Do's
✅ Keep it simple and recognizable at small sizes
✅ Use the specified color palette
✅ Ensure transparency works on both light and dark backgrounds
✅ Test at 16x16, 32x32, 64x64, 128x128 for visibility
✅ Use vector tools (Figma, Illustrator, Inkscape) then export to PNG

### Don'ts
❌ Don't include text (too small to read at icon sizes)
❌ Don't use too many colors
❌ Don't use copyrighted imagery
❌ Don't add complex gradients that don't scale
❌ Don't use thin lines that disappear at small sizes

## File Placement

Save the final icon as:
```
marketplace/icon.png
```

## Testing

Test the icon in these contexts:
- [ ] GitHub Marketplace listing (256x256)
- [ ] VS Code extension panel (various sizes)
- [ ] Browser favicon (if needed)
- [ ] Dark and light backgrounds

## Creation Tools

Recommended tools for creating the icon:
- **Figma** (free, web-based)
- **Adobe Illustrator** (professional)
- **Inkscape** (free, open-source)
- **Canva** (free with templates)

## Example Process

1. Create a 512x512 artboard (for high-res)
2. Design using vector shapes
3. Apply gradients and effects
4. Export at 256x256 PNG with transparency
5. Test on various backgrounds
6. Optimize file size
