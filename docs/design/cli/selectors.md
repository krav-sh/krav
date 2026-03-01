# Selectors

Selectors identify one or more policies (and optionally rules within them) for CLI commands to operate on. Any command that accepts a `<selector>` argument uses the syntax described here.

## Exact name matching

The simplest selector is an exact policy name:

```bash
arci hook policy explain security-baseline
```

This matches the single policy named `security-baseline`.

## Glob patterns

Selectors support glob patterns for matching multiple policies:

- `security-*` matches all policies whose names start with `security-`
- `*-injection` matches policies ending with `-injection`
- `*` matches all policies

```bash
# Enable all security policies
arci hook policy enable 'security-*'

# Disable all policies (use with care)
arci hook policy disable '*'
```

## Policy:rule syntax

To target a specific rule within a policy, use the `policy:rule` syntax:

- `security-baseline:blocked-commands` targets the `blocked-commands` rule within the `security-baseline` policy
- `security-baseline:*` targets all rules in the `security-baseline` policy

## Rule-level glob patterns

The rule portion also supports glob patterns:

- `security-baseline:block-*` matches rules starting with `block-` in the `security-baseline` policy
- `*:track-*` matches rules starting with `track-` in any policy

## Examples

```bash
# Explain a specific rule
arci hook policy explain security-baseline:blocked-commands

# Test all rules in a policy
arci hook policy test security-baseline:*

# Test tracking rules across all policies
arci hook policy test '*:track-*' --input @test-input.json
```

Rule-level selectors are useful with `explain` and `test` commands when you want to focus on specific behavior within a policy.

## See also

- [hook](commands/hook.md) — hook commands that accept selectors
