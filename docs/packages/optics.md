# Package `structural/optics`

**Import path:** `github.com/alvadorncorp/gopackages/functional/structural/optics`

`optics` provides one abstraction: a typed `Lens[S,A]` built from explicit
getter and immutable setter functions. `Get`, `Set`, and `Modify` operate on
one focus; free function `Compose` joins nested focuses without reflection,
field-name strings, runtime lookup, or interface registries.

`NewLens` rejects nil getter/setter callbacks. A Lens zero value is invalid.
The caller is responsible for supplying callbacks that satisfy Get-Set,
Set-Get, and Set-Set. The package does not check these laws at runtime. Value
copying alone does not deep-copy maps, slices, pointers, or other references;
setters own the required copy semantics for their source type.

The package exports no Prism, Traversal, Iso, Getter, Setter, or ModifyF and
imports no functional sibling. Domain/modeling packages may publish their own
Lens values while optics remains independent of them.
See [module overview](../overview.md) and [examples](../examples.md).
