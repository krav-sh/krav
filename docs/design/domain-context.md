# Domain context

## Overview

Every project has a domain: the subject matter that the system operates in, the vocabulary its users speak, the rules and invariants that constrain what the system can do. An agent working on a healthcare system needs to understand HIPAA. An agent working on a trading platform needs to know what a limit order is. An agent deriving requirements for a build tool needs to know how dependency graphs work.

ARCI structures development work but does not manage domain knowledge directly. The agent layer needs domain context to do useful work at every stage: formulating concepts with the right vocabulary, writing needs in stakeholder language, deriving correct requirements, writing code that handles domain edge cases, and reviewing work against rules the specification might not fully capture. Domain context integration is how ARCI connects its engineering lifecycle to the actual subject matter of the project.

## `.arci/DOMAIN.md`

Every ARCI project has a conventional domain reference document at `.arci/DOMAIN.md`. This file is the project's primary domain context source. It describes the problem domain in plain language: what the system does, who it's for, what vocabulary the domain uses, what rules and invariants govern the domain, and what external standards or regulations apply.

The file is human-authored and human-maintained. ARCI does not generate or modify it. The agent reads it whenever domain understanding would improve the quality of a graph operation. Skill and subagent instructions reference it by convention, ensuring the agent loads domain context before formalization, derivation, decomposition, review, and other domain-sensitive workflows.

### Content guidance

`DOMAIN.md` should cover:

The problem domain itself. Describe the space this project operates in, the problem it solves, and the key entities and their relationships. Write this for an intelligent reader who knows nothing about the domain. If the project is a trading system, explain what orders, fills, positions, and instruments are. If it's a build tool, explain what targets, dependencies, and build graphs are.

Domain vocabulary. Terms that have specific meaning in this domain. "Fill" in a trading system is not the same as "fill" in a graphics library. "Target" in a build system is not the same as "target" in a deployment pipeline. Define the terms that an agent needs to use correctly when writing concepts, needs, requirements, and code.

Domain rules and invariants. Facts about the domain that constrain the system. "Traders can only cancel a limit order before the exchange fills it." "A patient can have at most one active prescription for a controlled substance." "Build targets form a DAG; cycles are invalid." These are not requirements (they are not things the system "shall" do); they are truths about the domain that requirements must respect.

Regulatory and compliance context. Standards, regulations, or organizational policies that apply. HIPAA for healthcare. PCI-DSS for payment processing. WCAG for accessibility. SEC regulations for financial systems. These don't need full reproduction; a summary of what applies and references to authoritative sources is sufficient.

Stakeholder landscape. Who uses the system and what do they care about? This overlaps with the stakeholder classes on needs, but `DOMAIN.md` provides the broader picture: what does a day in the life of each stakeholder look like? What are their pain points? What would success look like for them?

### What `DOMAIN.md` is not

It's not a requirements document. It describes the domain, not the system. "Traders can cancel limit orders before fill" is domain context. "The system shall support limit order cancellation" is a requirement. The former informs the latter but doesn't replace it.

It's not a glossary file for a linter or prose checker. If the project uses Vale or a similar tool for documentation linting, the domain vocabulary in `DOMAIN.md` serves a different purpose (agent context, not automated prose checking) even though the content overlaps.

It's not exhaustive. The goal is "enough context for the agent to make good decisions," not "complete domain encyclopedia." A page or two is often sufficient. The file can reference external documents for depth.

## Module-level domain context

While `DOMAIN.md` provides project-wide domain context, individual modules may need additional context specific to their subdomain. A module representing the auth subsystem of a healthcare app needs HIPAA context that the logging module doesn't.

The `domainContext` field on Module nodes specifies additional context sources for the agent to load when working on that module.

### Schema

```json
{
  "@context": "context.jsonld",
  "@id": "MOD-A4F8R2X1",
  "@type": "Module",
  "title": "Payment processing",
  "domainContext": {
    "skills": ["pci-dss-compliance"],
    "documents": [
      "docs/domain/payment-flows.md",
      "docs/domain/card-network-rules.md"
    ]
  }
}
```

The `domainContext` object has two fields:

`skills` is an array of agent skill names. These are Claude Code skills (defined in `.claude/skills/` or provided by plugins) that contain structured instructions and reference material for a specific domain area. When the agent begins work on this module, the listed skills load into context. Skill names follow whatever naming convention the project's Claude Code skills use.

`documents` is an array of filesystem paths, relative to the project root. These are reference documents that the agent should read when working on this module. They might be internal documentation, excerpts from standards, domain model descriptions, or anything else that provides relevant context. ARCI validates paths at load time; missing documents produce warnings, not errors.

Both fields are optional. A module with no `domainContext` inherits context from `DOMAIN.md` only.

### Context loading

The context injection chain for a module is:

1. `.arci/DOMAIN.md` is always loaded. Every agent interaction has access to the project's domain reference.
2. The target module's `domainContext.skills` activate. Skill instructions become available to the agent.
3. The agent reads the target module's `domainContext.documents` into context, making the document contents available for reference.

Ancestor modules' `domainContext` is not automatically inherited. If a parent module references a HIPAA skill and a child module doesn't, the child module won't have HIPAA context unless it also lists it. This is intentional: context is expensive, and loading irrelevant parent context wastes the agent's context window. If a child module needs the same context as its parent, it lists the same sources.

The exception is `DOMAIN.md`, which applies everywhere regardless of module. Project-wide context is always available.

### JSON-LD context update

The schema adds the `domainContext` property to the JSON-LD context document:

```json
{
  "@context": {
    "domainContext": "arci:domainContext"
  }
}
```

It's a datatype property (not an object property) because its value is a structured literal, not a reference to another graph node.

## Integration with workflows

Domain context affects most workflows, but some more than others.

During concept exploration, `DOMAIN.md` provides the vocabulary and rules the agent uses to reason about design options. A concept about error handling in a compiler needs the agent to understand what syntax trees, token spans, and diagnostic levels are. The domain document provides this baseline.

During formalization and derivation, domain context shapes the needs and requirements the agent produces. Stakeholder language comes from the domain. Verification criteria reference domain invariants. An agent that doesn't understand the domain writes generic requirements; one that does writes requirements that reflect how the system actually needs to behave.

During task decomposition, module-level domain context helps the agent produce tasks with the right detail level and technical approach. A payment processing module with PCI-DSS context produces tasks that include security review and compliance verification. Without that context, the agent might skip those steps.

During review, domain context is essential. The agent reviewing auth code against HIPAA requirements needs HIPAA context to evaluate whether the code is adequate. The review skill instructions should explicitly load module domain context before beginning evaluation.

During verification, domain context informs what edge cases matter. Test cases for a trading system need to cover scenarios that only make sense if you understand how markets work (partial fills, after-hours orders, circuit breakers). The agent writing test cases needs domain context to identify these scenarios.

## Examples

### OSS parsing library

```markdown
<!-- .arci/DOMAIN.md -->
# Domain: parser construction

This project builds a parsing library. The domain is formal language
theory and compiler construction, applied to building practical parsers
for programming languages and data formats.

## Key concepts

A parser transforms a sequence of characters (source text) into a
structured representation (typically an abstract syntax tree or AST).
This happens in stages: lexing (characters to tokens), parsing (tokens
to tree), and optionally semantic analysis (tree annotation and
validation).

A grammar defines the language being parsed. Grammars are typically
expressed as production rules in a notation like BNF or PEG. The
grammar determines what inputs are valid and how they decompose into
structure...

## Domain vocabulary

- **Token**: An atomic unit of syntax (keyword, identifier, literal,
  operator, punctuation)
- **Span**: A range in the source text, identified by byte offset and
  length
- **AST**: Abstract syntax tree; the parsed structure with source
  locations stripped
- **CST**: Concrete syntax tree; preserves all source tokens including
  whitespace and comments
...
```

No module-level domain context needed for a single-domain library.

### Healthcare API with payment processing

```markdown
<!-- .arci/DOMAIN.md -->
# Domain: healthcare practice management

This system manages patient scheduling, clinical records, billing,
and insurance claims for small medical practices...
```

```json
{"@id": "MOD-B1LL1NG1", "@type": "Module", "title": "Billing",
 "domainContext": {
   "skills": ["healthcare-billing"],
   "documents": ["docs/domain/insurance-claim-lifecycle.md",
                  "docs/domain/cpt-code-reference.md"]}}
```

```json
{"@id": "MOD-P4YM3NT1", "@type": "Module", "title": "Payment processing",
 "domainContext": {
   "skills": ["pci-dss-compliance"],
   "documents": ["docs/domain/payment-flows.md"]}}
```

```json
{"@id": "MOD-CL1N1C41", "@type": "Module", "title": "Clinical records",
 "domainContext": {
   "skills": ["hipaa-compliance"],
   "documents": ["docs/domain/phi-handling-rules.md",
                  "docs/domain/clinical-terminology.md"]}}
```

### Cross-platform mobile app

```markdown
<!-- .arci/DOMAIN.md -->
# Domain: personal finance tracking

This app helps individuals track spending, set budgets, and understand
their financial habits...
```

```json
{"@id": "MOD-10SP1ATF", "@type": "Module", "title": "iOS app",
 "domainContext": {
   "skills": ["ios-platform"],
   "documents": ["docs/domain/apple-hig-guidelines.md"]}}
```

```json
{"@id": "MOD-4NDRO1D1", "@type": "Module", "title": "Android app",
 "domainContext": {
   "skills": ["android-platform"],
   "documents": ["docs/domain/material-design-guidelines.md"]}}
```

Platform skills here aren't strictly "domain" (they're platform knowledge), but the mechanism works for any context the agent needs when working on a specific module. The distinction between "domain context" and "platform context" is academic from the agent's perspective, as it's all reference material that improves decision quality.

## Open questions

**Should `arci init` create a stub `DOMAIN.md`?** Probably yes, with a comment template guiding the developer to fill in the sections. An empty domain document is a missed opportunity; a template-based one prompts the developer to think about domain context from the start.

**Validation of document paths.** When should ARCI validate that paths in `domainContext.documents` actually exist? At graph load time? At context injection time? The fail-open principle suggests validation should warn, not error. A missing document shouldn't block the agent from working.

**Context size management.** Domain documents and skills consume context window. For modules with extensive domain context (multiple skills, large documents), the injected context might crowd out the task-specific context the agent needs. Should there be a size budget? Should the agent summarize rather than inject full documents when context is tight? This is a general context engineering problem, not specific to domain context, but it matters here.

**Relationship to concepts.** Domain concepts (CON-* nodes with `conceptType: domain` or similar) could hold the same information as `DOMAIN.md` sections. Should there be a migration path from `DOMAIN.md` prose to structured concept nodes? Or are these intentionally separate: `DOMAIN.md` is reference context, concepts are design thinking?

**Shared skills across modules.** If multiple modules reference the same skill, the system loads it once per session (not per module). But the agent might work on multiple modules in a session. Should skill loading be session-scoped (load all skills for all modules the agent touches) or task-scoped (load only the current module's skills)?
