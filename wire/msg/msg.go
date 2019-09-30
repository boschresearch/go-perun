// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

// Package msg contains all message types, as well as serialising and
// deserialising logic used in peer communications.
package msg // import "perun.network/go-perun/wire/msg"

import (
	"io"
	"strconv"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
)

// Msg is the top-level abstraction for all messages sent between perun
// nodes.
type Msg interface {
	// Category returns the message's subcategory.
	Category() Category
}

// Encode encodes a message into an io.Writer.
func Encode(msg Msg, writer io.Writer) (err error) {
	// Encode the message category, then encode the message.
	if err = msg.Category().Encode(writer); err == nil {
		switch msg.Category() {
		case Control:
			cmsg, _ := msg.(ControlMsg)
			err = encodeControlMsg(cmsg, writer)
		case Channel:
			err = channelEncodeFun(writer, msg)
		case Peer:
			err = peerEncodeFun(writer, msg)
		default:
			log.Panicf("Encode(): Unhandled message category: %v", msg.Category())
		}
	}

	// Handle both error sources at once.
	if err != nil {
		err = errors.WithMessagef(err, "failed to write message with category %v", msg.Category())
	}
	return
}

// Decode decodes a message from an io.Reader.
func Decode(reader io.Reader) (Msg, error) {
	var cat Category
	if err := cat.Decode(reader); err != nil {
		return nil, errors.WithMessage(err, "failed to decode message category")
	}

	switch cat {
	case Control:
		return decodeControlMsg(reader)
	case Channel:
		return channelDecodeFun(reader)
	case Peer:
		return peerDecodeFun(reader)
	default:
		log.Panicf("Decode(): Unhandled message category: %v", cat)
		panic("This should never happen")
	}
}

var channelDecodeFun func(io.Reader) (Msg, error)

// RegisterChannelDecode register the function that will decode all messages of category ChannelMsg
func RegisterChannelDecode(fun func(io.Reader) (Msg, error)) {
	if channelDecodeFun != nil || fun == nil {
		log.Panic("RegisterChannelDecode called twice or with invalid argument")
	}
	channelDecodeFun = fun
}

var peerDecodeFun func(io.Reader) (Msg, error)

// RegisterPeerDecode register the function that will decode all messages of category PeerMsg
func RegisterPeerDecode(fun func(io.Reader) (Msg, error)) {
	if peerDecodeFun != nil || fun == nil {
		log.Panic("RegisterPeerDecode called twice or with invalid argument")
	}
	peerDecodeFun = fun
}

var channelEncodeFun func(io.Writer, Msg) error

// RegisterChannelEncode register the function that will encode all messages of category ChannelMsg
func RegisterChannelEncode(fun func(io.Writer, Msg) error) {
	if channelEncodeFun != nil || fun == nil {
		log.Panic("RegisterChannelEncode called twice or with invalid argument")
	}
	channelEncodeFun = fun
}

var peerEncodeFun func(io.Writer, Msg) error

// RegisterPeerEncode register the function that will encode all messages of category PeerMsg
func RegisterPeerEncode(fun func(io.Writer, Msg) error) {
	if peerEncodeFun != nil || fun == nil {
		log.Panic("RegisterPeerEncode called twice or with invalid argument")
	}
	peerEncodeFun = fun
}

// Category is an enumeration used for (de)serializing messages and
// identifying a message's subcategory.
type Category uint8

// Enumeration of message categories.
const (
	Control Category = iota
	Peer
	Channel

	// This constant marks the first invalid enum value.
	categoryEnd
)

// String returns the name of a message category, if it is valid, otherwise,
// returns its numerical representation for debugging purposes.
func (c Category) String() string {
	if !c.Valid() {
		return strconv.Itoa(int(c))
	}
	return [...]string{
		"ControlMsg",
		"PeerMsg",
		"ChannelMsg",
	}[c]
}

// Valid checks whether a Category is a valid value.
func (c Category) Valid() bool {
	return c < categoryEnd
}

func (c Category) Encode(writer io.Writer) error {
	if _, err := writer.Write([]byte{byte(c)}); err != nil {
		return errors.Wrap(err, "failed to write category")
	}
	return nil
}

func (c *Category) Decode(reader io.Reader) error {
	buf := make([]byte, 1)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return errors.WithMessage(err, "failed to write category")
	}

	*c = Category(buf[0])
	if !c.Valid() {
		return errors.New("invalid message category encoding: " + c.String())
	}
	return nil
}
