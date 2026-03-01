# Versioning and compatibility

This document describes arci's versioning policy, backward compatibility guarantees, and migration strategies. When teams commit rule configurations to git, they need confidence that upgrading arci won't break their workflows.

## Version numbering

arci uses semantic versioning (SemVer) with the format MAJOR.MINOR.PATCH.

MAJOR version increments indicate breaking changes that may require configuration migration. Examples include removing deprecated features, changing configuration schema in incompatible ways, or altering expression language semantics.

MINOR version increments add features in a backward-compatible manner. Examples include new action types, new expression functions, and new CLI commands.

PATCH version increments make backward-compatible bug fixes. Examples include fixing expression evaluation bugs, correcting documentation, and addressing security vulnerabilities.

## Pre-1.0 versioning

During the 0.x development phase, the API and configuration schema may change more freely. The project documents breaking changes in release notes, but users should expect some instability. The goal is to reach 1.0 with a stable, well-designed API rather than prematurely committing to suboptimal designs.

After 1.0, the compatibility guarantees this document describes apply fully.

## Compatibility guarantees

Pending: define specific compatibility guarantees.

Topics to cover:

### Configuration schema compatibility

- How long the system supports deprecated fields
- Migration tooling for schema changes
- Validation of old configurations against new schemas

### Expression language compatibility

- Adding new operators and functions
- Deprecating operators and functions
- Behavior changes in existing functions

### Action type compatibility

- Adding new action types
- Deprecating action types
- Changing action parameters

### CLI compatibility

- Adding new commands
- Deprecating commands
- Changing command output format

### API compatibility

- Adding new endpoints
- Deprecating endpoints
- Changing request/response schemas

### Extension API compatibility

- Protocol version for extensions
- Backward compatibility for extension authors
- Extension/arci version matrix

## Schema versioning

Pending: document configuration schema versioning.

Topics to cover:

### Schema version field

- Explicit version declaration in configuration
- Default version for legacy configurations
- Version validation on load

### Schema migration

- Automatic migration for minor changes
- Migration tooling for major changes
- Preserving comments and formatting

### Multiple schema support

- Supporting multiple schema versions simultaneously
- Deprecation timeline for old schemas
- Migration warnings in CLI and dashboard

## Extension compatibility

Pending: document extension versioning concerns.

Topics to cover:

### Extension/arci version constraints

- How extensions declare compatible arci versions
- Behavior when constraints aren't met
- Upgrade coordination between extension and arci

### Extension API versions

- Versioned extension protocol
- Backward compatibility for extension authors
- Breaking changes in extension API

### Lockfile compatibility

- Lockfile schema versioning
- Cross-version lockfile handling
- Regeneration requirements after upgrades

## Deprecation policy

Pending: define the deprecation process.

Topics to cover:

### Deprecation timeline

- Minimum support period for deprecated features
- Warning mechanisms during deprecation
- Removal process

### Deprecation communication

- Release notes
- CLI warnings
- Dashboard indicators
- Documentation updates

### Migration assistance

- Migration guides for deprecated features
- Automated migration tooling
- Validation of migrated configurations

## Release process

Pending: document the release process.

Topics to cover:

### Release cadence

- Regular release schedule (if any)
- Security release policy
- Long-term support versions (if any)

### Release artifacts

- GitHub releases with prebuilt binaries
- Go module releases (pkg.go.dev)
- Homebrew formula
- Changelog maintenance

### Testing before release

- Compatibility testing
- Regression testing
- Upgrade testing

## Upgrade workflows

Pending: document upgrade procedures.

Topics to cover:

### Upgrading arci

- Testing configuration against new version
- Reviewing breaking changes
- Rolling back if necessary

### Upgrading extensions

- Using `arci extension upgrade`
- Reviewing extension changelogs
- Testing after extension updates

### Team coordination

- Communicating upgrades to team members
- Synchronizing arci versions across team
- CI/CD considerations

## Known breaking changes

Pending: document breaking changes as they occur.

This section tracks breaking changes across versions, including what changed, why it changed, how to migrate, and which versions the change affects.

---

This document expands as actual releases refine the versioning policy. The goal is predictability: users should understand what upgrading means and have the tools to do it safely.
