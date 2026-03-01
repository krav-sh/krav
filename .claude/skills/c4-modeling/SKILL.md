---
name: c4-modeling
description: >-
  Create and update C4 architecture diagrams using C4-PlantUML.
  Use when creating new C4 diagrams, updating existing architecture diagrams,
  modeling system context, containers, or components, or when asked to draw,
  diagram, or visualize architecture with PlantUML.
stage-classification: permanent
rationale: "Domain-general architecture diagramming practice; independent of arci bootstrapping stages"
---

# Create and update C4 architecture diagrams

Generate correct C4 architecture diagrams using C4-PlantUML, following the C4 model specification and project conventions.

## Instructions

### Step 1: Determine the diagram level

Identify which C4 diagram level the user needs. If not specified, ask:

- **System Context (Level 1)**: The system in its environment, showing people, external systems, and relationships. Start here for new projects.
- **Container (Level 2)**: Zooms into one system to show containers (apps, data stores, services) and their interactions. The most commonly needed diagram type.
- **Component (Level 3)**: Zooms into one container to show its internal components. Scope is always a single container.
- **Dynamic**: Runtime behavior and interaction flows for a specific use case.
- **Deployment**: Maps containers to infrastructure nodes and environments.
- **System Landscape**: Enterprise-wide view of all software systems in an organization.

Level 1 and Level 2 provide the most value for most teams. Level 4 (Code) is rarely needed and should be auto-generated from IDEs if used at all.

See [references/c4-model-reference.md](references/c4-model-reference.md) for the full C4 abstraction hierarchy and diagram-level details.

### Step 2: Identify elements and relationships

For the chosen diagram level, identify:

1. **People**: Who interacts with the system or container? Use `Person()` for internal users, `Person_Ext()` for external users.
2. **Systems**: What software systems are in scope? Use `System()` for internal, `System_Ext()` for external.
3. **Containers** (Level 2+): What separately deployable units exist? Apps, data stores, message queues. A C4 "container" is a deployable unit and does not refer to a Docker container.
4. **Components** (Level 3): What are the major structural building blocks inside a single container?
5. **Relationships**: What are the interactions? Label with specific intent (like `Reads customer data` instead of `Uses`) and include the protocol for inter-process communication (`HTTPS`, `SQL`, `gRPC`).

See [references/c4-model-reference.md](references/c4-model-reference.md) for notation rules and element metadata requirements.

### Step 3: Read existing diagrams for consistency

Before writing or modifying a diagram, read diagrams at adjacent levels in `docs/diagrams/` to maintain consistency. The existing diagrams are `arci-c4-context.puml` and `arci-c4-container.puml`.

- Elements shared across levels must use the same alias, label, and technology
- External systems in a container diagram must match those in the context diagram
- Components in a component diagram must sit inside a container that appears in the container diagram

### Step 4: Write the diagram

Generate the `.puml` file following these project conventions.

**File location and naming:**

- All diagrams live in `docs/diagrams/`
- Naming convention: `arci-c4-{level}.puml` (kebab-case)
- The `@startuml` identifier must match the filename without extension

**Required structure (follow this exact ordering):**

```text
1. @startuml {id}               identifier matches filename
2. !include <C4/C4_{Level}>     stdlib import
3. title                        "[Diagram type] for [scope]"
4. Actors and external systems  Person, System_Ext, etc.
5. System boundary with internals
6. Relationships                Rel, Rel_L, Rel_R, etc.
7. SHOW_LEGEND()
8. @enduml
```

**Styling rules:**

- Use default C4-PlantUML styling only
- Do not add `skinparam`, `!theme`, or other style overrides

See [references/c4-plantuml-syntax.md](references/c4-plantuml-syntax.md) for the complete macro reference.

### Step 5: Apply element metadata

Every element must include:

- **Name**: Short, descriptive label
- **Type**: Technology or role (like `Python`, `PostgreSQL`, or `SPA`)
- **Description**: One sentence explaining what the element does

Every relationship must include:

- **Label**: Specific intent describing what happens (like `Sends order confirmations` instead of `Calls`)
- **Technology** (for inter-process communication): Protocol or API (`HTTPS`, `gRPC`, `SQL`)

Use unidirectional arrows. Use `BiRel()` only when both directions carry distinct, independent traffic.

### Step 6: Check the diagram

Run through the validation checklist before presenting the diagram:

1. Does the title follow `[Diagram type] for [scope]`?
2. Does every element have a name, type, and description?
3. Does every relationship have a label describing intent?
4. Do inter-process relationships include technology/protocol?
5. Are arrows unidirectional (unless bidirectional is intentional)?
6. Is `SHOW_LEGEND()` present?
7. Does the scope match the level? (Context = whole system, Container = one system, Component = one container)
8. Are there orphan elements with no relationships?

See [references/diagram-review-checklist.md](references/diagram-review-checklist.md) for the full validation checklist.

See [references/common-mistakes.md](references/common-mistakes.md) to verify the diagram avoids known C4 modeling pitfalls.

### Step 7: Lint and export

After writing the diagram file:

1. Run `just lint-plantuml` to check syntax.
2. Run `just export-plantuml` to generate PNG and SVG into `docs/diagrams/png/` and `docs/diagrams/svg/`.
3. If syntax errors occur, read the error output, fix the `.puml` file, and re-run.

## Updating existing diagrams

When modifying an existing diagram:

1. Read the current `.puml` file.
2. Read diagrams at adjacent levels for consistency (Step 3).
3. Make targeted changes. Do not rewrite the entire diagram unless the user requests it.
4. Run validation (Step 6) and linting (Step 7) after changes.

## Modeling guidance for common patterns

- **Microservices**: Model as containers within a system boundary. If one team owns a set of microservices, group them in a single system boundary.
- **Message queues and topics**: Model individual topics as separate containers, not a single message broker container.
- **Event-driven architectures**: Show event flows as relationships between producing and consuming containers, with the topic/queue as an intermediary container.
- **External systems**: Always use `_Ext()` variants. Do not show the internal structure of external systems.

## Error handling

- If `just lint-plantuml` fails, read the error output and fix syntax issues. Common issues: missing `@enduml`, unclosed boundaries, typos in macro names.
- If `just export-plantuml` fails, tell the user the export failed and suggest running `just lint-plantuml` first to isolate syntax problems.
- If the user requests a diagram type not covered by C4 (like an ER diagram or flowchart), explain that C4 covers system architecture views and suggest the appropriate UML diagram type instead.
