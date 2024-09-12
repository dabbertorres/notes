package util

func MapMap[M ~map[K]V, K comparable, V any, OK comparable, OV any](m M, mapper func(K, V) (OK, OV)) map[OK]OV {
	out := make(map[OK]OV, len(m))
	for k, v := range m {
		newK, newV := mapper(k, v)
		out[newK] = newV
	}
	return out
}
