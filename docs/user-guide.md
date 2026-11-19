# User guide

The editor has four areas: the palette on the left, the canvas in the middle, the code on the right and a side panel with the tabs Properties, Classes and Theme. The top bar holds the theme picker, sketch toggle, examples, grid settings, export buttons and the light/dark switch.

## Code and canvas

The code is the source of truth. Every canvas action becomes a small edit of the D2 text, and every text change triggers a new render.

- While you type, the editor renders a fast preview with one TALA seed. When you stop typing for about 1.5 seconds it renders the full layout.
- If the code has an error, the error is shown inline and the last good diagram stays on the canvas. Canvas edits are blocked until the error is fixed ("Fix the code error first").
- Undo and redo work in the code editor (`Ctrl+Z`, `Ctrl+Shift+Z`). A canvas action, including a group action on many shapes, is one undo step.
- Your code is saved in the browser's local storage and comes back on reload.

## Adding shapes

Drag an item from the palette onto the canvas.

- Dropped on empty space, the shape is added at the top level and pinned where you dropped it.
- Dropped on a container, it becomes a child of that container.
- Dropped in a grid or sequence diagram, it is added as the next cell or actor, without a position.

## Moving shapes

Drag a shape to move it.

- **Pin.** Dropping it in its own parent writes `top` and `left`, so TALA keeps it there.
- **Move into a container.** Dropping it on another container moves the shape's declaration into that container. References to it in the code are updated.
- **Move out.** Dropping it on empty space moves it to the top level.
- **Reorder.** Inside a grid or a sequence diagram, dropping it between siblings moves its declaration. A marker line shows where it will go.
- **Sequence messages.** Drag a message up or down to reorder it.

With grid snapping on (magnet button), shapes snap to the grid. The number next to it is the grid size.

## Connecting shapes

Select a shape and drag the round handle on its right edge onto another shape. This writes `a -> b`.

## Selecting

| Action | How |
|--------|-----|
| Select one shape or edge | Click it |
| Add or remove a shape | `Shift`, `Ctrl` or `Cmd` + click |
| Box select | Drag on empty space (hold a modifier to add to the selection) |
| Select all shapes | `Ctrl+A` / `Cmd+A` |
| Clear selection | `Esc` |

Selected shapes are highlighted in the code too.

## Keyboard and mouse

| Key | Action |
|-----|--------|
| Arrow keys | Nudge the selected shapes by 1 px |
| `Shift` + arrow keys | Nudge by one grid step |
| `Delete` | Delete the selection |
| `Space` + drag, or middle mouse drag | Pan |
| Mouse wheel | Zoom at the cursor |

A burst of arrow key presses is pinned as one edit once you pause.

## Properties panel

Shows the fields of the selected shape or edge: label, shape type, style (fill, stroke, font, opacity, and more), `near`, arrowheads for edges, classes and the pin checkbox. Clearing a field removes the key from the code.

With several shapes selected, a field shows a value only when all of them agree, and setting it writes it on every shape. Deleting a container deletes its children.

With nothing selected, "Pin all" writes `top` and `left` for every shape that can be pinned, freezing the current layout.

## Classes panel

Create D2 classes, rename them (all uses are updated) and edit their style fields. Assign classes to shapes from the Properties panel. Several classes are written as a D2 array.

## Theme panel

Overrides single theme colors (`theme-overrides` or `dark-theme-overrides` in `vars.d2-config`). Each slot is labeled with what it paints by default, for example N1 is text and B1 is borders and connection lines. Explicit styles on a shape win over the theme. The panel edits the set for the current app mode. Dark overrides only apply when the diagram sets a dark theme. "Reset all" removes the overrides.

The theme picker in the top bar sets `theme-id`. The pencil button toggles sketch style.

## Export

- **.d2** downloads the current code.
- **.svg** runs a full render without the preview seed, so the result matches the `d2` CLI.

## Performance tips

TALA is slow on large diagrams that have no fixed positions. Pin shapes, or use "Pin all" once you like the layout, and renders get much faster. See [perf.md](perf.md).
