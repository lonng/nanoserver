// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gldriver

import (
	"encoding/binary"
	"image"
	"image/color"
	"image/draw"

	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/gl"
)

type textureImpl struct {
	w    *windowImpl
	id   gl.Texture
	size image.Point
}

func (t *textureImpl) Size() image.Point       { return t.size }
func (t *textureImpl) Bounds() image.Rectangle { return image.Rectangle{Max: t.size} }

func (t *textureImpl) Release() {
	t.w.glctxMu.Lock()
	defer t.w.glctxMu.Unlock()

	t.w.glctx.DeleteTexture(t.id)
	t.id = gl.Texture{}
}

func (t *textureImpl) Upload(dp image.Point, src screen.Buffer, sr image.Rectangle) {
	buf := src.(*bufferImpl)
	buf.preUpload()

	t.w.glctxMu.Lock()
	defer t.w.glctxMu.Unlock()

	// TODO: adjust if dp is outside dst bounds, or r is outside src bounds.
	t.w.glctx.BindTexture(gl.TEXTURE_2D, t.id)
	m := buf.rgba.SubImage(sr).(*image.RGBA)
	b := m.Bounds()
	// TODO check m bounds smaller than t.size
	t.w.glctx.TexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, b.Dx(), b.Dy(), gl.RGBA, gl.UNSIGNED_BYTE, m.Pix)
}

func (t *textureImpl) Fill(dr image.Rectangle, src color.Color, op draw.Op) {
	t.w.glctxMu.Lock()
	defer t.w.glctxMu.Unlock()

	// TODO.
}

var quadCoords = f32Bytes(binary.LittleEndian,
	0, 0, // top left
	1, 0, // top right
	0, 1, // bottom left
	1, 1, // bottom right
)

const textureVertexSrc = `#version 100
uniform mat3 mvp;
uniform mat3 uvp;
attribute vec3 pos;
attribute vec2 inUV;
varying vec2 uv;
void main() {
	vec3 p = pos;
	p.z = 1.0;
	gl_Position = vec4(mvp * p, 1);
	uv = (uvp * vec3(inUV, 1)).xy;
}
`

const textureFragmentSrc = `#version 100
precision mediump float;
varying vec2 uv;
uniform sampler2D sample;
void main() {
	gl_FragColor = texture2D(sample, uv);
}
`

const fillVertexSrc = `#version 100
uniform mat3 mvp;
attribute vec3 pos;
void main() {
	vec3 p = pos;
	p.z = 1.0;
	gl_Position = vec4(mvp * p, 1);
}
`

const fillFragmentSrc = `#version 100
precision mediump float;
uniform vec4 color;
void main() {
	gl_FragColor = color;
}
`
