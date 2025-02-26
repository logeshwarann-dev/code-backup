// Code generated by Makefile, DO NOT EDIT.

/*
 * Copyright 2021 ByteDance Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package sse

import (
    `unsafe`

    `github.com/bytedance/sonic/internal/rt`
)

var F_f64toa func(out unsafe.Pointer, val float64) (ret int) 

var S_f64toa uintptr

//go:nosplit
func f64toa(out *byte, val float64) (ret int) {
	return F_f64toa((rt.NoEscape(unsafe.Pointer(out))), val)
}

