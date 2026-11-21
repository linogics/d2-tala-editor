# User guide

The editor has four areas: the palette on the left, the canvas in the middle, the code on the right and a side panel with the tabs Properties, Classes and Theme. The top bar holds the theme picker, sketch toggle, examples, grid settings, export buttons, the light/dark switch and links to this guide, the licenses and the GitHub repository.

Drag the border on the left of the side panel or of the code pane to resize it. Double click a border to reset its width. Widths are saved in the browser. A focused border also moves with the left and right arrow keys.

## Code and canvas

The code is the source of truth. Every canvas action becomes a small edit of the D2 text, and every text change triggers a new render.

- While you type, the editor renders a fast preview with one TALA seed. When you stop typing for about 1.5 seconds it renders the full layout.
- If the code has an error, the error is shown inline and the last good diagram stays on the canvas. Canvas edits are blocked until the error is fixed ("Fix the code error first").
- Undo and redo work in the code editor (`Ctrl+Z`, `Ctrl+Shift+Z`). A canvas action, including a group action on many shapes, is one undo step.
- Your code is saved in the browser's local storage and comes back on reload.

## Adding shapes

Drag an item from the palette onto the canvas. Both palette groups are sorted by name.

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

Select a shape and drag the round handle on its right edge onto another shape. This writes `a -> b`. Hold `Alt` when you release to write a plain line `a -- b` instead.

To change an existing connection, select it and use the **source head** and **target head** checkboxes in the Properties panel. They pick the D2 operator:

| source head | target head | Code |
|-------------|-------------|------|
| off | on | `a -> b` |
| on | off | `a <- b` |
| on | on | `a <-> b` |
| off | off | `a -- b` |

Each checkbox sits in a box with that end's arrowhead **Shape** and **Label**. An end without a head greys out its Shape list. A shape set before stays in the code and comes back when you turn the head on again. The Label works on both ends either way. Connections set through a glob such as `(a -> b)[*]` must be changed in the code.

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
| `Delete` | Delete the selection, containers with everything inside |
| `Shift+Delete` | Delete the selection, children of a container move up a level |
| `Alt` + drag the connect handle | Connect with a plain line `--` |
| `Space` + drag, or middle mouse drag | Pan |
| Mouse wheel | Zoom at the cursor |

A burst of arrow key presses is pinned as one edit once you pause.

## Properties panel

Shows the fields of the selected shape or edge: label, shape type, style (fill, stroke, font, opacity, and more), `near`, arrowheads for edges, classes and the pin checkbox. Clearing a field, or picking "(unset)" in a list, removes the key from the code. `near` lists the positions and every shape outside the selected shape's own branch. Edges have head checkboxes, see [Connecting shapes](#connecting-shapes). Text fields such as the label and tooltip take several lines: Shift+Enter adds a line break, Enter applies the value. SQL tables refuse labels with line breaks.

Fields sit in titled groups: General, Arrows, Position, Fill, Stroke, Font, Effects and Extras. Click a title to fold its group. Folds are kept per group name in the browser and apply to the Classes panel too.

With several shapes selected, a field shows a value only when all of them agree, and setting it writes it on every shape. Deleting a container deletes its children, Shift keeps them.

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
