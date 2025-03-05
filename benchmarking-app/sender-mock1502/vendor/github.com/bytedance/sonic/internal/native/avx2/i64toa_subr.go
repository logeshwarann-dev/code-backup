// +build !noasm !appengine
// Code generated by asm2asm, DO NOT EDIT.

package avx2

import (
	`github.com/bytedance/sonic/loader`
)

const (
    _entry__i64toa = 64
)

const (
    _stack__i64toa = 8
)

const (
    _size__i64toa = 2272
)

var (
    _pcsp__i64toa = [][2]uint32{
        {0x1, 0},
        {0xae, 8},
        {0xaf, 0},
        {0x201, 8},
        {0x202, 0},
        {0x287, 8},
        {0x288, 0},
        {0x456, 8},
        {0x457, 0},
        {0x4e2, 8},
        {0x4e3, 0},
        {0x610, 8},
        {0x611, 0},
        {0x771, 8},
        {0x772, 0},
        {0x8d9, 8},
        {0x8e0, 0},
    }
)

var _cfunc_i64toa = []loader.CFunc{
    {"_i64toa_entry", 0,  _entry__i64toa, 0, nil},
    {"_i64toa", _entry__i64toa, _size__i64toa, _stack__i64toa, _pcsp__i64toa},
}
