# Running verification

## What

The developer says `run the tests`, `verify the parser module`, `check coverage`, or `run verification for REQ-X`. The agent executes test cases, records results in the graph, identifies gaps, and reports on the verification state of requirements and needs.

## Why

Verification is where the graph proves its worth. Without it, requirements are aspirational statements. With recorded test results linked to test cases linked to requirements, you have evidence that the system does what it claims. The graph can answer "is this requirement verified?" with a concrete yes or no, traced to a specific test execution.

The upward propagation model is also important here: when all test cases for a requirement pass, the requirement can transition to `verified`. When all requirements derived from a need pass verification, the need can transition to `satisfied`. This is how verification results flow up the traceability chain from code-level evidence to stakeholder-level satisfaction.

## What happens in the graph

The agent identifies what to verify. This might be all test cases for a module, test cases for specific requirements, or the full test suite. It runs the tests (via the project's test runner, such as pytest, jest, or cargo test) and captures results.

For each test case, the agent updates `currentResult` (pass, fail, skip, error) and `lastRunAt`. Test cases in `implemented` status transition to `executable` once they run at least once.

The agent then evaluates coverage: which requirements have all test cases passing? Which have failing tests? Which lack test cases entirely? This analysis uses `krav reqcoverage` and `krav tc untested` to surface gaps.

If all test cases for a requirement pass, the agent may propose transitioning the requirement to `verified` status. This transition is not automatic; it requires confirmation that the tests are actually sufficient, not just that they pass.

## Trigger patterns

`Run the tests`, `verify module X`, `check coverage`, `are the parser requirements verified?`, `run the test suite`, `what's the verification status?`

## Graph before

TC-* nodes in `implemented` or `executable` status with `verifies` edges to REQ-* nodes. Test code exists in the codebase.

## Graph after

Updated `currentResult` and `lastRunAt` on TC-* nodes. Possibly REQ-* transitions to `verified`. Coverage gap reports.

## Agent interaction layer

### Skills

The `krav:verify` skill builds this workflow. It runs inside the `krav-verifier` subagent, so its preprocessing and instructions target isolated test execution. Preprocessing loads the module's TC-* nodes, their current results, and the verification method for each. The skill instructions guide the verifier through executing each test case and recording pass/fail/skip results with evidence via `krav tc record-result`.

The skill emphasizes faithful execution: run the tests as specified, report results accurately, don't fix failing tests. If a test case specification is ambiguous, the verifier flags it as a defect rather than interpreting it charitably. The skill also provides instructed commands for coverage analysis queries via `krav reqcoverage` and `krav tc untested`.

### Agents

The `krav-verifier` subagent is the canonical execution context for this workflow. Like the reviewer, it starts with a fresh context window to avoid coding bias. The agent that wrote the test code shouldn't also run and interpret the tests, since it carries assumptions about expected behavior that might mask real failures.

The verifier has access to Read, Grep, Glob, and Bash tools. Bash is broader than the reviewer's because the verifier needs to execute test suites, but the configuration excludes Write and Edit. The verifier records results through `krav` CLI commands, not by editing test files or graph data directly.

### Policies

The `subagent-context` policy fires on SubagentStart when the verifier launches, injecting the module's test cases and their current state as `additionalContext`. This complements the skill's preprocessing with live graph state.

The `review-completion-gate` policy fires on SubagentStop and enforces that the verifier records results for every TC-* node in scope before completing. Without this gate, the verifier could stop after running some tests and leave coverage gaps unreported.

The `graph-integrity` and `cli-auto-approve` policies ensure result recording goes through the CLI without friction.

### Task types

This workflow executes `execute-tests` tasks (the primary verification task type) and `benchmark-performance` tasks (for quantitative verification against threshold requirements). Both are verification-phase tasks that run inside the verifier subagent. The distinction matters because benchmark tasks produce quantitative measurements compared against requirement thresholds, while test execution tasks produce binary pass/fail results.

## Open questions

**How does the agent map test runner output to TC-* nodes?** The test runner reports pass/fail by test function name. The graph has TC-* nodes with `implementation` fields pointing at test files. The agent needs to bridge these: match test runner output (like `test_parser_error_handling PASSED`) to the corresponding TC-* node. This mapping could use paths, test function names, naming conventions, or explicit annotations, but the design does not specify any of these approaches.

**What about tests that aren't in the graph?** Most projects have tests that predate Krav or that developers wrote without graph involvement. Running the test suite produces results for tests that have no TC-* node. Should the agent ignore these, report them as untracked, or offer to create TC-* nodes for them?

**CI integration.** In practice, CI runs the tests, not the developer's Claude Code session. How do CI results get into the graph? An `krav tc import-results` command that parses test runner output? A PostToolUse hook in CI that fires after `npm test`? A GitHub Action that calls Krav? Nobody has designed this plumbing yet.

**Verification vs. regression testing.** "Run the tests" might mean "verify these new requirements" or "make sure nothing broke." The graph cares about both but for different reasons. Verification advances requirement status. Regression testing might reveal new defects against requirements that passed verification earlier. The agent needs to distinguish these cases and report on each.

**Partial verification.** What if 8 of 10 test cases for a requirement pass, and the team already knows about the 2 failing tests? The requirement isn't fully verified, but it's close. The design doesn't have a "partially verified" status. Should the agent track partial progress, or is it binary?

**Performance and load tests.** Some verification methods (analysis) produce quantitative results that need comparison against thresholds, not just pass/fail. "Latency p99 is 47 ms against a 50 ms requirement" is a pass, but the margin matters. Where does quantitative verification evidence live in the graph? The `currentResult` field is an enum (pass/fail/skip/error), not a measurement.
