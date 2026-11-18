# Time as data

A domain rule that depends on "now" receives the already-decided fact as
input (checkout: `Coupon.expired bool`, not `time.Now()` in the domain
package). `context` and the clock stay in the adapter. Don't export a `Clock`
from `workflow`.
