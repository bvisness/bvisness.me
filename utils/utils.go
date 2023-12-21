package utils

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func Must1[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func Must2[T1, T2 any](v1 T1, v2 T2, err error) (T1, T2) {
	if err != nil {
		panic(err)
	}
	return v1, v2
}
