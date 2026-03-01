# Bootstrapping arci with arci

arci is a development method backed by tooling. The method works before the tooling exists. This concept captures the bootstrapping strategy: how to build arci using its own approach before the enforcement tooling is in place.

## The bootstrapping problem

You can't use arci to build arci until arci exists. But you can follow arci's method manually and progressively replace manual discipline with real enforcement. The transformation chain (concept → need → requirement → task → deliverable) is a way of thinking about work, not just a set of command-line commands.

## Staged approach

**Stage 0 (proto-arci)** builds the knowledge graph by hand. CLAUDE.md encodes development discipline as agent instructions. Shell-based Claude Code skills approximate core operations (find ready tasks, show task context, produce status overviews). This tests the ontology under real conditions.

**Stage 1 (graph storage and command-line tool)** builds the graph engine and enough command-line surface to replace hand-editing and shell scripts. Once this works, arci manages its own development graph through its own commands.

**Stage 2 (hooks and enforcement)** builds the hook engine. Development discipline becomes enforceable rather than advisory. PreToolUse hooks deny writes to baselined content. Stop hooks enforce deliverable recording. SessionStart hooks inject graph context.

**Stage 3+ (skills, dashboard, extensions)** builds the agent interaction layer, diagnostics dashboard, and extension system. Each stage uses the tooling from previous stages.

## Risk calibration

The risk of skipping manual practice is building arci without ever using arci's approach, discovering problems late, and having to retrofit. The risk of overdoing it is spending so long hand-crafting JSON-LD that code never ships. Each stage should produce just enough graph structure to plan the next stage's work and test the ontology.
