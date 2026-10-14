package types

// Ordered is the constraint every type ordered by Interval must satisfy.
// comparable enables the zero-value (unbounded) check inside Interval without
// reflection; Compare provides the total order. Both version axes
// (GolangciVersion, GoVersion) implement it; comparing intervals or values
// across the two axes is a compile error.
type Ordered[V any] interface {
	comparable
	Compare(V) int
}

// Interval is a closed range [Min, Max] over an ordered type. The zero value
// of Min or Max means unbounded in that direction, so Interval[GolangciVersion]{}
// matches every version. An interval with both bounds set and Min > Max is
// empty (Contains is always false).
type Interval[V Ordered[V]] struct {
	Min V
	Max V
}

// NewInterval returns the closed interval [min, max]. It panics when both
// bounds are set and min > max: an inverted interval is always a programming
// error in curated data, never a runtime condition. Use the struct literal for
// intentionally unbounded ranges.
func NewInterval[V Ordered[V]](min, max V) Interval[V] {
	i := Interval[V]{Min: min, Max: max}
	if i.IsEmpty() {
		panic("types: inverted interval (min > max)")
	}

	return i
}

// IsZeroBound reports whether bound is the zero value (unbounded).
func isZeroBound[V Ordered[V]](bound V) bool {
	var zero V

	return bound == zero
}

// Contains reports whether v lies within the interval. Unbounded sides are
// open: a zero Min imposes no lower bound, a zero Max imposes no upper bound.
func (i Interval[V]) Contains(v V) bool {
	if !isZeroBound(i.Min) && v.Compare(i.Min) < 0 {
		return false
	}

	if !isZeroBound(i.Max) && v.Compare(i.Max) > 0 {
		return false
	}

	return true
}

// IsEmpty reports whether the interval can contain nothing (both bounds set,
// Min > Max).
func (i Interval[V]) IsEmpty() bool {
	if isZeroBound(i.Min) || isZeroBound(i.Max) {
		return false
	}

	return i.Min.Compare(i.Max) > 0
}

// IsUnbounded reports whether neither side of the interval is bounded, i.e.
// the interval matches everything.
func (i Interval[V]) IsUnbounded() bool {
	return isZeroBound(i.Min) && isZeroBound(i.Max)
}

// LowerBound returns the interval's lower bound, or "" when unbounded below.
func (i Interval[V]) String() string {
	switch {
	case i.IsUnbounded():
		return "any"
	case isZeroBound(i.Min):
		return "<=" + i.Max.String()
	case isZeroBound(i.Max):
		return ">=" + i.Min.String()
	default:
		return ">=" + i.Min.String() + ",<=" + i.Max.String()
	}
}

// ParseInterval parses the format Interval.String produces, using parse to
// turn each bound back into V. Round-trips: ParseInterval[V](i.String(), parse)
// equals i for non-empty intervals.
func ParseInterval[V Ordered[V]](s string, parse func(string) (V, error)) (Interval[V], error) {
	if s == "any" {
		return Interval[V]{}, nil
	}

	var interval Interval[V]

	for _, part := range splitIntervalBounds(s) {
		op, value, found := strings.Cut(part, "=")
		if !found {
			return Interval[V]{}, fmt.Errorf("types: invalid interval bound %q in %q", part, s)
		}

		bound, err := parse(value)
		if err != nil {
			return Interval[V]{}, fmt.Errorf("types: invalid interval %q: %w", s, err)
		}

		switch op {
		case ">":
			interval.Min = bound
		case "<":
			interval.Max = bound
		default:
			return Interval[V]{}, fmt.Errorf("types: invalid interval operator %q in %q", op, s)
		}
	}

	interval.MustValidate()

	return interval, nil
}

// MustValidate panics when the interval is inverted.
func (i Interval[V]) MustValidate() {
	if i.IsEmpty() {
		panic("types: inverted interval (min > max)")
	}
}

func splitIntervalBounds(s string) []string {
	return strings.Split(s, ",")
}
