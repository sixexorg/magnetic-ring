package crypto

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/sixexorg/magnetic-ring/crypto/secp256k1/bitelliptic"
)

var (
	curve  = bitelliptic.S256()
	params = curve.Params()
	one    = big.NewInt(1)

	ErrInvalidVRF    = errors.New("invalid VRF proof")
	ErrInvalidPubkey = errors.New("invalid secp256k1 public key")
)

// PrivateKey supports evaluating the VRF function.
type PrivateKey interface {
	// Evaluate returns the output of H(f_k(m)) and its proof.
	Evaluate(m []byte) (index [32]byte, proof []byte)
	// Public returns the corresponding public key.
	Public() PublicKey
	// Bytes byte array of private key.
	Bytes() []byte
	// Sign returns the result of the signature for hash.
	Sign(hash []byte) ([]byte, error)

	Hex() string
}

// PublicKey supports verifying output from the VRF function.
type PublicKey interface {
	// ProofToHash verifies the NP-proof supplied by Proof and outputs Index.
	ProofToHash(m, proof []byte) (index [32]byte, err error)
	// Verify verify hash with signature and public key.
	Verify(hash, sig []byte) (bool, error)
	// Bytes byte array of public key.
	Bytes() []byte
	Hex() string
}

// GenerateKey 生成带VRF功能的私钥
func GenerateKey() (PrivateKey, error) {
	priv, err := generateEcdsaKey()
	if err != nil {
		return nil, err
	}

	return &VrfablePrivateKey{priv}, nil
}

// ToPrivateKey 将字节数组转换成带VRF功能的私钥
func ToPrivateKey(d []byte) (PrivateKey, error) {
	priv := new(VrfablePrivateKey)
	ecdsaPriv, err := toECDSA(d)
	if err != nil {
		return nil, err
	}

	priv.PrivateKey = ecdsaPriv
	return priv, nil
}
func HexToPrivateKey(hexStr string) (PrivateKey, error) {
	byteArr, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, err
	}
	return ToPrivateKey(byteArr)
}

// UnmarshalPubkey converts bytes to a secp256k1 public key.
func UnmarshalPubkey(pub []byte) (PublicKey, error) {
	x, y := elliptic.Unmarshal(S256(), pub)
	if x == nil {
		return nil, ErrInvalidPubkey
	}

	pubKey := new(VrfablePublicKey)
	pubKey.PublicKey = &ecdsa.PublicKey{Curve: S256(), X: x, Y: y}
	return pubKey, nil
}

// PrivateKey holds a private VRF key.
type VrfablePrivateKey struct {
	*ecdsa.PrivateKey
}

// Evaluate returns the verifiable unpredictable function evaluated at m
func (k *VrfablePrivateKey) Evaluate(m []byte) (index [32]byte, proof []byte) {
	nilIndex := [32]byte{}
	// Prover chooses r <-- [1,N-1]
	r, _, _, err := generateKeyFromCurve(curve, rand.Reader)
	if err != nil {
		return nilIndex, nil
	}
	ri := new(big.Int).SetBytes(r)

	// H = H1(m)
	Hx, Hy := H1(m)

	// VRF_k(m) = [k]H
	sHx, sHy := curve.ScalarMult(Hx, Hy, k.D.Bytes())

	// vrf := elliptic.Marshal(curve, sHx, sHy) // 65 bytes.
	vrf := curve.Marshal(sHx, sHy) // 65 bytes.

	// G is the base point
	// s = H2(G, H, [k]G, VRF, [r]G, [r]H)
	rGx, rGy := curve.ScalarBaseMult(r)
	rHx, rHy := curve.ScalarMult(Hx, Hy, r)
	var b bytes.Buffer
	b.Write(curve.Marshal(params.Gx, params.Gy))
	b.Write(curve.Marshal(Hx, Hy))
	b.Write(curve.Marshal(k.PublicKey.X, k.PublicKey.Y))
	b.Write(vrf)
	b.Write(curve.Marshal(rGx, rGy))
	b.Write(curve.Marshal(rHx, rHy))
	s := H2(b.Bytes())

	// t = r−s*k mod N
	t := new(big.Int).Sub(ri, new(big.Int).Mul(s, k.D))
	t.Mod(t, params.N)

	// Index = H(vrf)
	index = sha256.Sum256(vrf)

	// Write s, t, and vrf to a proof blob. Also write leading zeros before s and t
	// if needed.
	var buf bytes.Buffer
	buf.Write(make([]byte, 32-len(s.Bytes())))
	buf.Write(s.Bytes())
	buf.Write(make([]byte, 32-len(t.Bytes())))
	buf.Write(t.Bytes())
	buf.Write(vrf)

	return index, buf.Bytes()
}

// Public returns the corresponding public key as bytes.
func (k *VrfablePrivateKey) Public() PublicKey {
	return &VrfablePublicKey{&k.PublicKey}
}

func (k *VrfablePrivateKey) Bytes() []byte {
	priv := k.PrivateKey
	return paddedBigBytes(priv.D, priv.Params().BitSize/8)
}

func (k *VrfablePrivateKey) Sign(hash []byte) ([]byte, error) {
	return Sign(k.PrivateKey, hash)
}
func (k *VrfablePrivateKey) Hex() string {
	return hex.EncodeToString(k.Bytes())
}

func (k *VrfablePrivateKey) Clear(hash []byte) ([]byte, error) {
	fmt.Printf("%s\n", k.D)
	return Sign(k.PrivateKey, hash)
}

// PublicKey holds a public VRF key.
type VrfablePublicKey struct {
	*ecdsa.PublicKey
}

func (pk *VrfablePublicKey) Verify(hash, sig []byte) (bool, error) {
	return Verify(pk.PublicKey, hash, sig)
}

func (pk *VrfablePublicKey) Bytes() []byte {
	pub := pk.PublicKey
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}

	return elliptic.Marshal(curve, pub.X, pub.Y)
}
func (pk *VrfablePublicKey) Hex() string {
	return hex.EncodeToString(pk.Bytes())
}

// ProofToHash asserts that proof is correct for m and outputs index.
func (pk *VrfablePublicKey) ProofToHash(m, proof []byte) (index [32]byte, err error) {
	nilIndex := [32]byte{}
	// verifier checks that s == H2(m, [t]G + [s]([k]G), [t]H1(m) + [s]VRF_k(m))
	if got, want := len(proof), 64+65; got != want {
		return nilIndex, ErrInvalidVRF
	}

	// Parse proof into s, t, and vrf.
	s := proof[0:32]
	t := proof[32:64]
	vrf := proof[64 : 64+65]

	// uHx, uHy := elliptic.Unmarshal(curve, vrf)
	uHx, uHy := curve.Unmarshal(vrf) //////???
	if uHx == nil {
		return nilIndex, ErrInvalidVRF
	}

	// [t]G + [s]([k]G) = [t+ks]G
	tGx, tGy := curve.ScalarBaseMult(t)
	ksGx, ksGy := curve.ScalarMult(pk.X, pk.Y, s)
	tksGx, tksGy := curve.Add(tGx, tGy, ksGx, ksGy)

	// H = H1(m)
	// [t]H + [s]VRF = [t+ks]H
	Hx, Hy := H1(m)
	tHx, tHy := curve.ScalarMult(Hx, Hy, t)
	sHx, sHy := curve.ScalarMult(uHx, uHy, s)
	tksHx, tksHy := curve.Add(tHx, tHy, sHx, sHy)

	//   H2(G, H, [k]G, VRF, [t]G + [s]([k]G), [t]H + [s]VRF)
	// = H2(G, H, [k]G, VRF, [t+ks]G, [t+ks]H)
	// = H2(G, H, [k]G, VRF, [r]G, [r]H)
	var b bytes.Buffer
	b.Write(curve.Marshal(params.Gx, params.Gy))
	b.Write(curve.Marshal(Hx, Hy))
	b.Write(curve.Marshal(pk.X, pk.Y))
	b.Write(vrf)
	b.Write(curve.Marshal(tksGx, tksGy))
	b.Write(curve.Marshal(tksHx, tksHy))
	h2 := H2(b.Bytes())

	// Left pad h2 with zeros if needed. This will ensure that h2 is padded
	// the same way s is.
	var buf bytes.Buffer
	buf.Write(make([]byte, 32-len(h2.Bytes())))
	buf.Write(h2.Bytes())

	if !hmac.Equal(s, buf.Bytes()) {
		return nilIndex, ErrInvalidVRF
	}
	return sha256.Sum256(vrf), nil
}

func H1(m []byte) (x, y *big.Int) {
	h := sha512.New()
	var i uint32
	byteLen := (params.BitSize + 7) >> 3
	for x == nil && i < 100 {
		// TODO: Use a NIST specified DRBG.
		h.Reset()
		binary.Write(h, binary.BigEndian, i)
		h.Write(m)
		r := []byte{2} // Set point encoding to "compressed", y=0.
		r = h.Sum(r)
		x, y = Unmarshal(curve, r[:byteLen+1])
		i++
	}

	return
}

// H2 hashes to an integer [1,N-1]
func H2(m []byte) *big.Int {
	// NIST SP 800-90A § A.5.1: Simple discard method.
	byteLen := (params.BitSize + 7) >> 3
	h := sha512.New()
	for i := uint32(0); ; i++ {
		// TODO: Use a NIST specified DRBG.
		h.Reset()
		binary.Write(h, binary.BigEndian, i)
		h.Write(m)
		b := h.Sum(nil)
		k := new(big.Int).SetBytes(b[:byteLen])
		if k.Cmp(new(big.Int).Sub(params.N, one)) == -1 {
			return k.Add(k, one)
		}
	}
}

// Unmarshal a compressed point in the form specified in section 4.3.6 of ANSI X9.62.
func Unmarshal(curve elliptic.Curve, data []byte) (x, y *big.Int) {
	byteLen := (curve.Params().BitSize + 7) >> 3
	if (data[0] &^ 1) != 2 {
		return // unrecognized point encoding
	}
	if len(data) != 1+byteLen {
		return
	}

	// Based on Routine 2.2.4 in NIST Mathematical routines paper
	params := curve.Params()
	tx := new(big.Int).SetBytes(data[1 : 1+byteLen])
	y2 := y2(params, tx)
	sqrt := defaultSqrt
	ty := sqrt(y2, params.P)
	if ty == nil {
		return // "y^2" is not a square: invalid point
	}
	var y2c big.Int
	y2c.Mul(ty, ty).Mod(&y2c, params.P)
	if y2c.Cmp(y2) != 0 {
		return // sqrt(y2)^2 != y2: invalid point
	}
	if ty.Bit(0) != uint(data[0]&1) {
		ty.Sub(params.P, ty)
	}

	x, y = tx, ty // valid point: return it
	return
}

// Use the curve equation to calculate y² given x.
// only applies to curves of the form y² = x³ - 3x + b.
func y2(curve *elliptic.CurveParams, x *big.Int) *big.Int {

	// y² = x³ - 3x + b
	//x3 := new(big.Int).Mul(x, x)
	//x3.Mul(x3, x)
	//
	//threeX := new(big.Int).Lsh(x, 1)
	//threeX.Add(threeX, x)
	//
	//x3.Sub(x3, threeX)
	//x3.Add(x3, curve.B)
	//x3.Mod(x3, curve.P)

	// change to bitelliptic : y² = x³ + b
	x3 := new(big.Int).Mul(x, x)
	x3.Mul(x3, x)

	x3.Add(x3, curve.B)
	x3.Mod(x3, curve.P)
	return x3
}

func defaultSqrt(x, p *big.Int) *big.Int {
	var r big.Int
	if nil == r.ModSqrt(x, p) {
		return nil // x is not a square
	}
	return &r
}
