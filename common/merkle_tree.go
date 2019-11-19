package common

import "crypto/sha256"

// param hashes will be used as workspace
func ComputeMerkleRoot(hashes []Hash) Hash {
	if len(hashes) == 0 {
		return Hash{}
	}
	sha := sha256.New()
	var temp Hash
	for len(hashes) != 1 {
		n := len(hashes) / 2
		for i := 0; i < n; i++ {
			sha.Reset()
			sha.Write(hashes[2*i][:])
			sha.Write(hashes[2*i+1][:])
			sha.Sum(temp[:0])
			sha.Reset()
			sha.Write(temp[:])
			sha.Sum(hashes[i][:0])
		}
		if len(hashes) == 2*n+1 {
			sha.Reset()
			sha.Write(hashes[2*n][:])
			sha.Write(hashes[2*n][:])

			sha.Sum(temp[:0])
			sha.Reset()
			sha.Write(temp[:])
			sha.Sum(hashes[n][:0])

			hashes = hashes[:n+1]
		} else {
			hashes = hashes[:n]
		}
	}

	return hashes[0]
}
