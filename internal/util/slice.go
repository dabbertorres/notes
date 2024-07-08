package util

func MapSlice[S ~[]T, T any, O any](slice S, mapper func(T) O) []O {
	out := make([]O, len(slice))
	for i, v := range slice {
		out[i] = mapper(v)
	}
	return out
}

func MapSliceIndexed[S ~[]T, T any, O any](slice S, mapper func(int, T) O) []O {
	out := make([]O, len(slice))
	for i, v := range slice {
		out[i] = mapper(i, v)
	}
	return out
}

// SliceDiff does the same thing as [SliceDiffBy], but is a convenience for values
// that are directly comparable (or otherwise don't need custom comparison logic).
func SliceDiff[S ~[]T, T comparable](before, after S) (additions, deletions S) {
	return SliceDiffBy(before, after, func(lhs, rhs T) bool { return lhs == rhs })
}

// SliceDiffBy returns two slices representing the additions and deletions done on before
// to get after.
// In other words, by adding everything in additions to before, and removing everything in
// deletions from before, the result would be equal to after.
//
// Both before and after should be sorted in the same manner (method and order).
func SliceDiffBy[S ~[]T, T any](before, after S, equal func(T, T) bool) (additions, deletions S) {
	bi, ai := 0, 0

	for bi < len(before) && ai < len(after) {
		if !equal(before[bi], after[ai]) {
			// look ahead for before[bi] in after to see if we can jump ahead
			var found bool
			j := ai
			for ; j < len(after); j++ {
				if equal(before[bi], after[j]) {
					found = true
					break
				}
			}

			if found {
				// everything between ai and j is an addition
				additions = append(additions, after[ai:j]...)

				// we can continue from j in after
				ai = j
			} else {
				// this is a deletion
				deletions = append(deletions, before[bi])

				bi++

				// is it an addition?
				found = false
				j = bi
				for ; j < len(before); j++ {
					if equal(before[j], after[ai]) {
						found = true
						break
					}
				}

				// couldn't find it in before, so yes it is
				if !found {
					additions = append(additions, after[ai])
					ai++
				}
			}
		} else {
			bi++
			ai++
		}
	}

	if ai < len(after) {
		// everything else is added
		additions = append(additions, after[ai:]...)
	}

	if bi < len(before) {
		// everything else is deleted
		deletions = append(deletions, before[bi:]...)
	}

	return
}
