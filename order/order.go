package order

// Order compares a and b. A negative result means a < b, zero means
// equivalent, and a positive result means a > b. Only the sign is significant.
type Order[T any] func(a, b T) int
