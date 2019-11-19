package dpoa

import "math"

func CalcStellar(total float64) (m, n int) {
	n = int(math.Sqrt(total))
	if n % 2 == 0{
		n = n -1
	}
	m = int(total/float64(n))
	return
}
