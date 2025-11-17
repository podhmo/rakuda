package bindingparse

// Validator is the interface that wraps the basic Validate method.
type Validator interface {
	Validate() error
}

// WithValidation is a higher-order function that takes a parser for a type T
// and returns a new parser. The returned parser first decodes the raw string
// into a value of type T, and then, if the decoding is successful, it calls
// the Validate() method on the decoded value.
//
// The type parameter T is constrained to types that implement the Validator interface.
func WithValidation[T Validator](parse func(s string) (T, error)) func(s string) (T, error) {
	return func(s string) (T, error) {
		val, err := parse(s)
		if err != nil {
			var zero T
			return zero, err
		}
		if err := val.Validate(); err != nil {
			var zero T
			return zero, err
		}
		return val, nil
	}
}
