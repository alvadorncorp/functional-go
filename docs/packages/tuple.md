# Package `structural/tuple`

**Import path:** `github.com/alvadorncorp/gopackages/functional/structural/tuple`

`tuple` carries two or three heterogeneous values through intermediate
operations without reflection or ad-hoc local structs. Tuple2 exposes First and
Second; Tuple3 adds Third. Zero values are valid and payload immutability is
shallow: pointer, slice, and map payloads preserve their ordinary Go aliases.

Construct with Of2/Of3. `FromValues2` is a `Deprecated:` synonym of Of2: it
keeps working with unchanged behavior and stays exported in v1, but new code
should call Of2
([ADR-014](../adr/functional/adr-014-v1-surface-hygiene.md)). MapFirst,
MapSecond, MapFirst3, MapSecond3, and MapThird3 may change one component type;
Swap reverses Tuple2. The package intentionally has no Pair/Triple aliases,
Tuple4+, rotations, generic equality, or String formatting.

The package is a stdlib-only leaf and is the carrier used by iterator.Zip and
iterator.Enumerate. See [module overview](../overview.md) and
[examples](../examples.md).
