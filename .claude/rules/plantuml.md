---
paths:
  - "**/*.puml"
---

# PlantUML diagram conventions

**Always activate the `c4-modeling` skill with the Skill tool before creating or updating any `.puml` file.** The skill provides the full C4 model reference, macro syntax, validation checklist, and common mistakes guide.

## Diagram location and naming

- All diagrams live in `docs/diagrams/`
- Naming convention: `krav-c4-{level}.puml` (kebab-case)
- The `@startuml` identifier must match the filename without extension: `@startuml krav-c4-context`

## C4 library conventions

- Use the standard PlantUML C4 includes: `!include <C4/C4_{Level}>` where `{Level}` is `Context`, `Container`, or `Component`
- Use C4 element macros: `Person`, `System`, `System_Ext`, `System_Boundary`, `Container`, `Container_Ext`, `ContainerDb`, `ContainerDb_Ext`, `Container_Boundary`, `Component`, `Rel`, `Rel_L`, `Rel_R`
- End every diagram with `SHOW_LEGEND()`

## Diagram structure

Diagrams must follow this ordering:

1. `@startuml {id}` (identifier matches filename)
2. `!include <C4/C4_{Level}>` (C4 library include)
3. `title` (diagram title)
4. Actors and external systems
5. System boundary with internal elements
6. Relationships (`Rel`)
7. `SHOW_LEGEND()`
8. `@enduml`

## Validation and export

- Run `just lint-plantuml` to check syntax
- Run `just export-plantuml` to generate PNG and SVG into `docs/diagrams/png/` and `docs/diagrams/svg/`
- Run `just watch-plantuml` during iterative editing for automatic re-export on save
- Never commit generated PNG/SVG without re-exporting from current `.puml` source

## Styling

- Use default C4-PlantUML styling only
- Do not add `skinparam`, `!theme`, or other style overrides
