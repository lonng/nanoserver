// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package i2c allows users to read from and write to a slave I2C device.
package i2c // import "golang.org/x/exp/io/i2c"

import (
	"fmt"

	"golang.org/x/exp/io/i2c/driver"
)

const tenbitMask = 1 << 12

// Device represents an I2C device. Devices must be closed once
// they are no longer in use.
type Device struct {
	conn driver.Conn
}

// TenBit marks an I2C address as a 10-bit address.
func TenBit(addr int) int {
	return addr | tenbitMask
}

// TOOD(jbd): Do we need higher level I2C packet writers and readers?
// TODO(jbd): Support bidirectional communication.

// Read reads len(buf) bytes from the device.
func (d *Device) Read(buf []byte) error {
	// TODO(jbd): Support reading from a register.
	if err := d.conn.Read(buf); err != nil {
		return fmt.Errorf("error reading from device: %v", err)
	}
	return nil
}

// Write writes the buffer to the device. If it is required to write to a
// specific register, the register should be passed as the first byte in the
// given buffer.
func (d *Device) Write(buf []byte) (err error) {
	if err := d.conn.Write(buf); err != nil {
		return fmt.Errorf("error writing to the device: %v", err)
	}
	return nil
}

// Close closes the device and releases the underlying sources.
func (d *Device) Close() error {
	return d.conn.Close()
}

// Open opens a connection to an I2C device.
// All devices must be closed once they are no longer in use.
func Open(o driver.Opener) (*Device, error) {
	conn, err := o.Open()
	if err != nil {
		return nil, err
	}
	return &Device{conn: conn}, nil
}

func resolveAddr(a int) (addr int, tenbit bool) {
	return a & (tenbitMask - 1), a&tenbitMask == tenbitMask
}
