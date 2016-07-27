/*
 * ISC License (ISC)
 *
 * Copyright (c) 2016, Derek Buitenhuis <derek.buitenhuis at gmail dot com>
 *
 * Permission to use, copy, modify, and/or distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

/*
 * This file contans minimal Go-to-C bindings for use termbox-go that emulate
 * just enough of the termbox C API to allow BXD to link and run. The goal
 * is Windows support for C codebases using termbox.
 *
 * To build and install:
 *     go get github.com/nsf/termbox-go
 *     go build -buildmode=c-archive -o termbox.a termbox-c.go
 *     cp -f termbox.h $(PREFIX)/include/termbox.h
 *     cp -f termbox.a $(PREFIX)/lib/libtermbox.a
 */

package main

/*
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

#define TB_BLACK 1
#define TB_WHITE 8
#define TB_BOLD 0x0100
#define TB_UNDERLINE 0x0200
#define TB_EVENT_KEY 1
#define TB_EVENT_RESIZE 2
#define TB_KEY_ARROW_UP (0xFFFF-18)
#define TB_KEY_ARROW_DOWN (0xFFFF-19)
#define TB_KEY_SPACE 0x20

struct tb_cell {
    uint32_t ch;
    uint16_t fg;
    uint16_t bg;
};

struct tb_event {
    uint8_t type;
    uint8_t mod;
    uint16_t key;
    uint32_t ch;
    int32_t w;
    int32_t h;
    int32_t x;
    int32_t y;
};
*/
import "C"

import (
	termbox "github.com/nsf/termbox-go"
	"unsafe"
)

var cellBuffer *C.struct_tb_cell
var cellSize int
var attrs = map[C.uint16_t]termbox.Attribute{
	0: termbox.ColorDefault,
	1: termbox.ColorBlack,
	8: termbox.ColorWhite,
}
var cattrs = map[termbox.Attribute]C.uint16_t{
	termbox.ColorDefault: 0,
	termbox.ColorBlack:   1,
	termbox.ColorWhite:   2,
}

//export tb_init
func tb_init() C.int {
	cellBuffer = nil
	cellSize = 0

	err := termbox.Init()
	if err != nil {
		return -1
	}
	return 0
}

//export tb_shutdown
func tb_shutdown() {
	termbox.Close()
}

//export tb_width
func tb_width() C.int {
	w, _ := termbox.Size()
	return C.int(w)
}

//export tb_height
func tb_height() C.int {
	_, h := termbox.Size()
	return C.int(h)
}

//export tb_clear
func tb_clear() {
	if cellBuffer != nil && cellSize != 0 {
		C.memset(unsafe.Pointer(cellBuffer), 0, C.size_t(cellSize))
	}
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
}

//export tb_present
func tb_present() {
	w, h := termbox.Size()
	buf := termbox.CellBuffer()
	if cellSize == w*h {
		for i := 0; i < w*h; i++ {
			c_cell := (*C.struct_tb_cell)(unsafe.Pointer(uintptr(unsafe.Pointer(cellBuffer)) + uintptr(C.sizeof_struct_tb_cell*i)))
			buf[i].Fg = termbox.Attribute((*c_cell).fg)
			buf[i].Bg = termbox.Attribute((*c_cell).bg)
			buf[i].Ch = rune(uint32((*c_cell).ch))
		}
	}
	termbox.Flush()
}

//export tb_cell_buffer
func tb_cell_buffer() *C.struct_tb_cell {
	w, h := termbox.Size()
	buf := termbox.CellBuffer()
	if cellSize != len(buf) {
		if cellBuffer != nil {
			C.free(unsafe.Pointer(cellBuffer))
		}
		cellBuffer = (*C.struct_tb_cell)(C.malloc(C.sizeof_struct_tb_cell * C.size_t(w*h)))
		cellSize = w * h
	}
	for i := 0; i < w*h; i++ {
		c_cell := (*C.struct_tb_cell)(unsafe.Pointer(uintptr(unsafe.Pointer(cellBuffer)) + uintptr(C.sizeof_struct_tb_cell*i)))
		(*c_cell).fg = C.uint16_t(buf[i].Fg)
		(*c_cell).bg = C.uint16_t(buf[i].Bg)
		(*c_cell).ch = C.uint32_t(uint32(buf[i].Ch))
	}
	return cellBuffer
}

//export tb_poll_event
func tb_poll_event(ev *C.struct_tb_event) C.int {
	event := termbox.PollEvent()

	if event.Type == termbox.EventKey {
		(*ev)._type = C.TB_EVENT_KEY
	} else if event.Type == termbox.EventResize {
		(*ev)._type = C.TB_EVENT_RESIZE
		return 1
	} else {
		(*ev)._type = 3
		return 1
	}

	if event.Ch == 0 {
		if event.Key == termbox.KeySpace {
			(*ev).key = C.TB_KEY_SPACE
		} else if event.Key == termbox.KeyArrowUp {
			(*ev).key = C.TB_KEY_ARROW_UP
		} else if event.Key == termbox.KeyArrowDown {
			(*ev).key = C.TB_KEY_ARROW_DOWN
		} else {
			(*ev).key = 0
		}
	} else {
		(*ev).key = 0
		(*ev).ch = C.uint32_t(uint32(event.Ch))
	}

	return 1
}

func main() {
}
