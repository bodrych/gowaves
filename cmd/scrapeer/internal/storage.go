package internal

import (
	"fmt"
	"net"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/wavesplatform/gowaves/pkg/crypto"
)

const (
	peerNodePrefix       byte = iota           // Keys to store peers by its IPs
)

var (
	zeroSignature    = crypto.Signature{}
	maxSignature     = crypto.Signature{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
)

type peerKey struct {
	prefix byte
	ip     net.IP
}

func (k peerKey) bytes() []byte {
	buf := make([]byte, 1+net.IPv6len)
	buf[0] = k.prefix
	copy(buf[1:], k.ip.To16())
	return buf
}

func (k *peerKey) fromBytes(data []byte) {
	if l := len(data); l < 1+net.IPv6len {
		panic(fmt.Sprintf("%d is not enough bytes for peerKey", l))
	}
	k.prefix = data[0]
	k.ip = net.IP(data[1 : 1+net.IPv6len])
}

type signatureKey struct {
	prefix    byte
	signature crypto.Signature
}

func (k signatureKey) bytes() []byte {
	buf := make([]byte, 1+crypto.SignatureSize)
	buf[0] = k.prefix
	copy(buf[1:], k.signature[:])
	return buf
}

type storage struct {
	db      *leveldb.DB
	genesis crypto.Signature
}

func NewStorage(path string, genesis crypto.Signature) (*storage, error) {
	wrapError := func(err error) error {
		return errors.Wrap(err, "failed to open storage")
	}
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, wrapError(err)
	}
	s := &storage{db: db, genesis: genesis}

	sn, err := s.db.GetSnapshot()
	if err != nil {
		return nil, wrapError(err)
	}
	defer sn.Release()
	return s, nil
}

func (s *storage) Close() error {
	return s.db.Close()
}

func (s *storage) peer(ip net.IP) (PeerNode, error) {
	peer := PeerNode{}
	sn, err := s.db.GetSnapshot()
	if err != nil {
		return peer, err
	}
	defer sn.Release()
	k := peerKey{prefix: peerNodePrefix, ip: ip.To16()}
	v, err := sn.Get(k.bytes(), nil)
	if err != nil {
		return peer, err
	}
	err = peer.UnmarshalBinary(v)
	if err != nil {
		return peer, err
	}
	return peer, nil
}

func (s *storage) putPeer(ip net.IP, peer PeerNode) error {
	batch := new(leveldb.Batch)
	k := peerKey{prefix: peerNodePrefix, ip: ip.To16()}
	v, err := peer.MarshalBinary()
	if err != nil {
		return err
	}
	batch.Put(k.bytes(), v)
	err = s.db.Write(batch, nil)
	if err != nil {
		return err
	}
	return nil
}

func (s *storage) peers() ([]PeerNode, error) {
	sn, err := s.db.GetSnapshot()
	if err != nil {
		return nil, errors.Wrap(err, "failed to collect peers")
	}
	defer sn.Release()
	st := []byte{peerNodePrefix}
	lm := []byte{peerNodePrefix + 1}
	it := sn.NewIterator(&util.Range{Start: st, Limit: lm}, nil)
	r := make([]PeerNode, 0)
	for it.Next() {
		var v PeerNode
		err = v.UnmarshalBinary(it.Value())
		if err != nil {
			return nil, errors.Wrap(err, "failed to collect peers")
		}
		r = append(r, v)
	}
	it.Release()
	return r, nil
}

func (s *storage) hasPeer(ip net.IP) (bool, error) {
	sn, err := s.db.GetSnapshot()
	if err != nil {
		return false, err
	}
	defer sn.Release()
	k := peerKey{prefix: peerNodePrefix, ip: ip.To16()}
	_, err = sn.Get(k.bytes(), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}