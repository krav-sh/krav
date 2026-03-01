# Phase-gated module lifecycle

Each module in arci tracks its current lifecycle phase independently. Phases gate what work can happen and when a module can advance, ensuring that architecture settles before design begins, design settles before coding begins, and so on.

## The phases

The six phases follow ISO/IEC/IEEE 15288 lifecycle processes:

**Architecture** identifies components, boundaries, and interfaces. The module's decomposition into child modules happens here. Architecture-phase tasks produce design documents, interface specifications, and module boundary definitions.

**Design** defines APIs, data models, and algorithms. The detailed engineering work that turns architectural boundaries into concrete specifications. Design-phase tasks produce API specs, data model definitions, and algorithm descriptions.

**Coding** builds the thing. Code, configuration, and other artifacts that realize the design. Coding-phase tasks produce source files, tests, and build configurations.

**Integration** assembles components and resolves interfaces. It verifies that separately built pieces work together. Integration-phase tasks produce integration test results and interface compliance evidence.

**Verification** tests against requirements. It confirms that what the team built matches what the team specified. Verification-phase tasks execute test cases and record results.

**Validation** confirms stakeholder needs. The final check that the built, verified system actually satisfies the original expectations. Validation-phase tasks involve stakeholder review and acceptance.

## Phase gates

A module advances to the next phase only when gate criteria hold:

- All tasks for the current phase are complete
- No blocking defects (critical or major severity) remain open
- Someone explicitly requests the advancement (not automatic)

Phase advancement creates a natural checkpoint for quality. If coding tasks reveal design problems, those surface as defects that block advancement until the team resolves them.

## Independent phases

Each module's phase is independent. The Graph module can be in coding while the Dashboard module is still in architecture. Cross-module coordination uses task dependencies and baseline policies rather than synchronized phases. Work proceeds where it's ready rather than waiting for the entire project to advance together.

## Phase regression

If a baselined module needs changes, the team regresses its phase. This creates a defect automatically, acknowledging that the change breaks a previously accepted baseline. The module then progresses through phases again with the changes and gets re-baselined at the appropriate gate.
