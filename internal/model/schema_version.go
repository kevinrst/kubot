package model

// Semver of the --json contract. MAJOR = breaking field change,
// MINOR = additive only.
//
// 0.1.0: initial contract.
// 0.2.0: additive — Report gains `degraded` (failed collectors, partial data).
// 0.3.0: additive — Finding gains `suppressed` + `suppression_reason`.
const SchemaVersion = "0.3.0"
