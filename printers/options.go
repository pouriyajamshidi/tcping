package printers

// options contains common display options shared by all printers
type options struct {
	ShowTimestamp     bool
	ShowSourceAddress bool
	ShowFailuresOnly  bool
}

type hasOptions interface {
	options() *options
}

// WithTimestamp enables timestamp display in printer output
func WithTimestamp[T hasOptions]() func(T) {
	return func(p T) {
		p.options().ShowTimestamp = true
	}
}

// WithSourceAddress enables source address display in printer output
func WithSourceAddress[T hasOptions]() func(T) {
	return func(p T) {
		p.options().ShowSourceAddress = true
	}
}

// WithFailuresOnly configures the printer to only show failed probes
func WithFailuresOnly[T hasOptions]() func(T) {
	return func(p T) {
		p.options().ShowFailuresOnly = true
	}
}
