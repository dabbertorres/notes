package util

// Apply calls f with t, and then returns t.
func Apply[T any](t T, f func(T)) T {
	f(t)
	return t
}

// Chain1 returns a new function that when called, returns the result of
// calling fs... in the provided order.
//
// If fs... is empty, then the identity function for F is returned, where
// the return value is equal to the argument.
func Chain1[F ~func(A) A, A any](fs ...F) F {
	if len(fs) == 0 {
		return func(arg A) A { return arg }
	}

	return func(arg A) A {
		for i := 0; i < len(fs); i++ {
			arg = fs[i](arg)
		}
		return arg
	}
}

// ChainReverse1 returns a new function that when called, returns the result of
// calling fs... in the opposite of the provided order.
//
// If fs... is empty, then the identity function for F is returned, where
// the return value is equal to the argument.
func ChainReverse1[F ~func(A) A, A any](fs ...F) F {
	if len(fs) == 0 {
		return func(arg A) A { return arg }
	}

	return func(arg A) A {
		for i := len(fs) - 1; i >= 0; i-- {
			arg = fs[i](arg)
		}
		return arg
	}
}

// AllUntilError1 returns a new function that calls all fs in order until one returns an error,
// and returns it, or if none fail, returns nil.
//
// If fs... is empty, a function that does nothing and returns nil is returned.
func AllUntilError1[F ~func(A) error, A any](fs ...F) F {
	if len(fs) == 0 {
		return func(A) error { return nil }
	}

	return func(a A) error {
		for _, f := range fs {
			if err := f(a); err != nil {
				return err
			}
		}

		return nil
	}
}

// AllUntilError2 returns a new function that calls all fs in order until one returns an error,
// and returns it, or if none fail, returns nil.
//
// If fs... is empty, a function that does nothing and returns nil is returned.
func AllUntilError2[F ~func(A1, A2) error, A1, A2 any](fs ...F) F {
	if len(fs) == 0 {
		return func(A1, A2) error { return nil }
	}

	return func(a1 A1, a2 A2) error {
		for _, f := range fs {
			if err := f(a1, a2); err != nil {
				return err
			}
		}

		return nil
	}
}

func FoldBool[T any](success bool, onSuccess, onFailure T) T {
	if success {
		return onSuccess
	}
	return onFailure
}
