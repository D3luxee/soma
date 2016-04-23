/*-
Copyright (c) 2016, Jörg Pernfuß <code.jpe@gmail.com>
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

* Redistributions of source code must retain the above copyright notice, this
  list of conditions and the following disclaimer.

* Redistributions in binary form must reproduce the above copyright notice,
  this list of conditions and the following disclaimer in the documentation
  and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

// Package somaauth implements the common key exchange and
// authentication bits between SOMA service and client.
package somaauth

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"hash"
	"math/big"
	"net"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/nacl/box"

	"github.com/dchest/blake2b"
	"github.com/satori/go.uuid"
)

type KexRequest struct {
	Public               string    `json:"public_key"`
	Request              uuid.UUID `json:"request,omitempty"`
	InitializationVector string    `json:"initialization_vector"`
	Token                string    `json:"password_token,omitempty"`
	// unexported private fields
	private  string    `json:"-"`
	peer     string    `json:"-"`
	sourceIP net.IP    `json:"-"`
	count    uint      `json:"-"`
	time     time.Time `json:"-"`
}

// KexExpirySeconds can be set to regulate how fast an open KexRequest
// expires
var KexExpirySeconds uint64 = 60

// NewKexRequest returns a KexRequest with a set random
// InitializationVector and new generated random keypair
func NewKexRequest() *KexRequest {
	var (
		err error
		bIV []byte
		k   KexRequest
	)

	k = KexRequest{}

	// generate the IV from 192bit random
	bIV = make([]byte, 24)
	if _, err = rand.Read(bIV); err != nil {
		return nil
	}
	k.InitializationVector = hex.EncodeToString(bIV)

	if err = k.GenerateNewKeypair(); err != nil {
		return nil
	}
	return &k
}

// GenerateNewKeypair generate a new public,private Keypair
func (k *KexRequest) GenerateNewKeypair() error {
	var (
		err                   error
		bRandom, bSecret      []byte
		hBlake                hash.Hash
		publicKey, privateKey *[32]byte
	)

	// generate keypair, starting with 1024bit random as
	// noted in DJB's docs
	bRandom = make([]byte, 128)
	if _, err = rand.Read(bRandom); err != nil {
		return err
	}

	// hash 1k down to 256bit
	hBlake = blake2b.New256()
	hBlake.Write(bRandom)
	bSecret = hBlake.Sum(nil)
	// generate keys
	if publicKey, privateKey, err = box.GenerateKey(
		bytes.NewReader(bSecret),
	); err != nil {
		return err
	}

	// set keys
	k.Public = hex.EncodeToString(publicKey[:])
	k.private = hex.EncodeToString(privateKey[:])
	return nil
}

// GenerateNewRequestID generate a new UUIDv4 for this Kex
func (k *KexRequest) GenerateNewRequestID() {
	k.Request = uuid.NewV4()
}

// IsExpired returns true if the Kex-Exchange is more than
// KexExpirySeconds seconds old
func (k *KexRequest) IsExpired() bool {
	return time.Now().UTC().After(k.time.UTC().Add(
		time.Duration(KexExpirySeconds) * time.Second))
}

// IsSameSource returns true if the paramter IP address is the
// same as the one recorded in the KexRequest
func (k *KexRequest) IsSameSource(ip net.IP) bool {
	return k.sourceIP.Equal(ip)
}

// NextNonce returns the next nonce to use. Nonces are built by
// interpreting the IV as a positive integer number and adding the
// count of requested nonces; thus implementing a simple non-repeating
// counter. The IV itself is never used as a nonce.
// Returns nil on error
func (k *KexRequest) NextNonce() *[24]byte {
	var (
		ib    []byte
		iv    *big.Int
		nonce *[24]byte
		err   error
	)

	k.count += 1
	if ib, err = hex.DecodeString(k.InitializationVector); err != nil {
		// hex decode failed
		return nil
	}

	// initialize
	iv = big.NewInt(0)
	// convert k.IV to Int by setting ib bytes as value
	iv.SetBytes(ib)
	// ensure the resulting number is positive
	iv.Abs(iv)
	// add the current counter value on top
	iv.Add(iv, big.NewInt(int64(k.count)))
	// check the resulting number still fits *[24]byte
	if len(iv.Bytes()) != 24 {
		return nil
	}

	// build nonce
	nonce = &[24]byte{}
	copy(nonce[:], iv.Bytes()[0:24])
	return nonce
}

// PeerKey returns the public key of the kex peer, or nil if it has
// not been set yet.
func (k *KexRequest) PeerKey() *[32]byte {
	var (
		pk   []byte
		err  error
		peer *[32]byte
	)

	// k.peer has not been set yet
	if k.peer == "" {
		return nil
	}

	// k.peer was set incorrect
	if pk, err = hex.DecodeString(k.peer); err != nil {
		return nil
	}

	// how?!
	if len(pk) != 32 {
		return nil
	}

	// return public key bytes
	peer = &[32]byte{}
	copy(peer[:], pk[0:32])
	return peer
}

// PrivateKey returns our private key for this kex, or nil if it has
// not been set yet.
func (k *KexRequest) PrivateKey() *[32]byte {
	var (
		pk      []byte
		err     error
		private *[32]byte
	)

	// k.private has not been set yet
	if k.private == "" {
		return nil
	}

	// k.private was set incorrect
	if pk, err = hex.DecodeString(k.private); err != nil {
		return nil
	}

	// again: how?
	if len(pk) != 32 {
		return nil
	}

	// return private key bytes
	private = &[32]byte{}
	copy(private[:], pk[0:32])
	return private
}

// PublicKey returns our public key for this key exchange, or nil if
// it has not been set yet.
func (k *KexRequest) PublicKey() *[32]byte {
	var (
		pk     []byte
		err    error
		public *[32]byte
	)

	// k.Public has not been set yet
	if k.Public == "" {
		return nil
	}

	// k.Public was set incorrect
	if pk, err = hex.DecodeString(k.Public); err != nil {
		return nil
	}

	// you know it.
	if len(pk) != 32 {
		return nil
	}

	// return public key bytes
	public = &[32]byte{}
	copy(public[:], pk[0:32])
	return public
}

// SetPeerKey sets the kex peer public key
func (k *KexRequest) SetPeerKey(pk *[32]byte) {
	k.peer = hex.EncodeToString(pk[:])
}

// SetRequestUUID sets the UUID of this Kex from a string
func (k *KexRequest) SetRequestUUID(s string) error {
	var err error

	if k.Request, err = uuid.FromString(s); err != nil {
		return err
	}

	return nil
}

// SetTimeUTC records the current time within the Kex
func (k *KexRequest) SetTimeUTC() {
	k.time = time.Now().UTC()
}

// SetIPAddress records the client's IP address
func (k *KexRequest) SetIPAddress(r *http.Request) {
	k.sourceIP = net.ParseIP(extractAddress(r.RemoteAddr))
}

// extractAddress extracts the IP address part of the IP:port string
// set as net/http.Request.RemoteAddr. It handles IPv4 cases like
// 192.0.2.1:48467 and IPv6 cases like [2001:db8::1%lo0]:48467
func extractAddress(str string) string {
	var addr string

	switch {
	case strings.Contains(str, `]`):
		// IPv6 address [2001:db8::1%lo0]:48467
		addr = strings.Split(str, `]`)[0]
		addr = strings.Split(addr, `%`)[0]
		addr = strings.TrimLeft(addr, `[`)
	default:
		// IPv4 address 192.0.2.1:48467
		addr = strings.Split(str, `:`)[0]
	}
	return addr
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
