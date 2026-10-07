package model

// Semver of the --json contract. MAJOR = breaking field change,
// MINOR = additive only.
//
// 0.1.0: initial contract.
// 0.2.0: additive — Report gains `degraded` (failed collectors, partial data).
const SchemaVersion = "0.2.0"
