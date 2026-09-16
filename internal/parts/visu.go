package parts

import (
	"encoding/binary"
	"log"

	"go-secondreality/internal/blob"
	"go-secondreality/internal/constants"
	"go-secondreality/internal/shim"
)

const (
	visuMaxObj       = 256
	visuMaxSides     = 16
	visuUnitShr      = 14
	visuRowLen       = constants.ScreenWidth
	visuObjPoolSize  = 64
	visuMatrixPoolSz = visuObjPoolSize * 2
	visuVListPoolSz  = 4096
	visuPVListPoolSz = 2048
	visuNListPoolSz  = 4096
	visuMemPoolSize  = 160000
)

const (
	visuVFUp    = 1
	visuVFDown  = 2
	visuVFLeft  = 4
	visuVFRight = 8
	visuVFNear  = 16
	visuVFFar   = 32
)

const (
	visuFDefault uint16 = 0xf001
	visuFVisible uint16 = 0x0001
	visuF2Side   uint16 = 0x0200
	visuFShade32 uint16 = 0x0C00
	visuFGouraud uint16 = 0x1000
)

type visuAngle uint16
type visuVisfl int16

type visuRMatrix struct {
	m [9]int32
	x int32
	y int32
	z int32
}

type visuVList struct {
	x      int32
	y      int32
	z      int32
	normal int16
	_      int16
}

type visuNList struct {
	x int16
	y int16
	z int16
	_ int16
}

type visuPVList struct {
	x  int16
	y  int16
	vf visuVisfl
	_  [5]int16
}

type visuObject struct {
	flags  uint16
	parent *visuObject
	r0     *visuRMatrix
	r      *visuRMatrix
	pd     []byte
	pdLen  int16
	plnum  int
	pl     [16][]uint16
	vnum   int
	nnum   int
	nnum1  int
	v0     []visuVList
	n0     []visuNList
	v      []visuVList
	n      []visuNList
	pv     []visuPVList
	vf     visuVisfl
	name   string
}

type visuSCo struct {
	o     *visuObject
	dist  int32
	index int
	on    int
}

type visuSScl struct {
	data []byte
}

type visuSCl struct {
	frames int
	ready  int
}

type visuPoly struct {
	sides uint16
	color uint16
	flags uint16
	vxseg []visuVList
	v     [visuMaxSides]struct{ x, y int16 }
	VX    [visuMaxSides]*visuVList
	GR    [visuMaxSides]uint16
}

type visuStream struct {
	p []uint16
}

var (
	visuPolyOversample     uint8
	visuPolyOversample16   uint8 = 16
	visuPolyOversamples    uint8 = 1
	visuPolyOversampleMask uint8

	visuProjClipX = [2]int32{0, constants.ScreenWidth - 1}
	visuProjClipY = [2]int32{0, constants.ScreenHeight - 1}
	visuProjClipZ = [2]int32{256, 1000000000}

	visuProjMulX int32 = 250
	visuProjMulY int32 = 220
	visuProjAddX int32 = 160
	visuProjAddY int32 = 100

	visuProjAspect        uint16 = 256
	visuProjOversampleShr uint16

	visuPoly1    visuPoly
	visuPoly2    visuPoly
	visuPolyCurY uint16

	visuPolyDraw [2048]uint16

	visuObjects     [visuObjPoolSize]visuObject
	visuObjectsIdx  int
	visuMatrices    [visuMatrixPoolSz]visuRMatrix
	visuMatricesIdx int
	visuVListPool   [visuVListPoolSz]visuVList
	visuVListIdx    int
	visuPVListPool  [visuPVListPoolSz]visuPVList
	visuPVListIdx   int
	visuNListPool   [visuNListPoolSz]visuNList
	visuNListIdx    int
	visuMemPool     [visuMemPoolSize]byte
	visuMemPoolIdx  int

	visuScene0 []byte
	visuScenem []byte
	visuCity   int
	visuXit    int

	visuSceneList [64]visuSScl
	visuScl       int
	visuSclp      int

	visuCo    [visuMaxObj]visuSCo
	visuCoNum int

	visuCamObject visuObject
	visuCam       visuRMatrix

	visuOrder    [visuMaxObj]int
	visuOrderNum int

	visuSp []byte

	visuCl          [4]visuSCl
	visuClr         int
	visuClw         int
	visuFirstFrame  int = 1
	visuDeadlock    int
	visuCopperCnt   int
	visuSyncFrame   int
	visuCurrFrame   int
	visuCopperDelay int = 16
	visuRepeat      int
	visuAvgRepeat   int
)

func visuResetMemoryPool() {
	visuObjectsIdx = 0
	visuMatricesIdx = 0
	visuVListIdx = 0
	visuPVListIdx = 0
	visuNListIdx = 0
	visuMemPoolIdx = 0
}

func visuReset() {
	_ = visuEnsureData()
	visuResetMemoryPool()

	visuPolyOversample = 0
	visuPolyOversample16 = 16
	visuPolyOversamples = 1
	visuPolyOversampleMask = 0

	visuProjClipZ[0] = 256
	visuProjClipZ[1] = 1000000000
	visuProjClipX[0] = 0
	visuProjClipX[1] = constants.ScreenWidth - 1
	visuProjClipY[0] = 0
	visuProjClipY[1] = constants.ScreenHeight - 1

	visuProjMulX = 250
	visuProjMulY = 220
	visuProjAddX = 160
	visuProjAddY = 100
	visuProjAspect = 256
	visuProjOversampleShr = 0

	visuPoly1 = visuPoly{}
	visuPoly2 = visuPoly{}
	visuPolyCurY = 0

	for i := range visuPolyDraw {
		visuPolyDraw[i] = 0
	}

	visuScene0 = nil
	visuScenem = nil
	visuCity = 0
	visuXit = 0
	for i := range visuSceneList {
		visuSceneList[i] = visuSScl{}
	}
	visuScl = 0
	visuSclp = 0
	for i := range visuCo {
		visuCo[i] = visuSCo{}
	}
	visuCoNum = 0
	visuCamObject = visuObject{}
	visuCam = visuRMatrix{}
	for i := range visuOrder {
		visuOrder[i] = 0
	}
	visuOrderNum = 0
	visuSp = nil
	for i := range visuCl {
		visuCl[i] = visuSCl{}
	}
	visuClr = 0
	visuClw = 0
	visuFirstFrame = 1
	visuDeadlock = 0
	visuCopperCnt = 0
	visuSyncFrame = 0
	visuCurrFrame = 0
	visuCopperDelay = 16
	visuRepeat = 0
	visuAvgRepeat = 0
}

func visuGetNewObject() *visuObject {
	if visuObjectsIdx >= len(visuObjects) {
		return &visuObjects[len(visuObjects)-1]
	}
	obj := &visuObjects[visuObjectsIdx]
	visuObjectsIdx++
	*obj = visuObject{}
	return obj
}

func visuGetNewRMatrix() *visuRMatrix {
	if visuMatricesIdx >= len(visuMatrices) {
		return &visuMatrices[len(visuMatrices)-1]
	}
	m := &visuMatrices[visuMatricesIdx]
	visuMatricesIdx++
	*m = visuRMatrix{}
	return m
}

func visuAllocVList(count int) []visuVList {
	if count <= 0 {
		return nil
	}
	if visuVListIdx+count <= len(visuVListPool) {
		out := visuVListPool[visuVListIdx : visuVListIdx+count]
		visuVListIdx += count
		return out
	}
	return make([]visuVList, count)
}

func visuAllocPVList(count int) []visuPVList {
	if count <= 0 {
		return nil
	}
	if visuPVListIdx+count <= len(visuPVListPool) {
		out := visuPVListPool[visuPVListIdx : visuPVListIdx+count]
		visuPVListIdx += count
		return out
	}
	return make([]visuPVList, count)
}

func visuAllocNList(count int) []visuNList {
	if count <= 0 {
		return nil
	}
	if visuNListIdx+count <= len(visuNListPool) {
		out := visuNListPool[visuNListIdx : visuNListIdx+count]
		visuNListIdx += count
		return out
	}
	return make([]visuNList, count)
}

func visuReadFile(name string) []byte {
	data, err := blob.ReadAll(name)
	if err != nil {
		log.Printf("visu: readfile %s: %v", name, err)
		return nil
	}
	if len(data) <= len(visuMemPool)-visuMemPoolIdx {
		dst := visuMemPool[visuMemPoolIdx : visuMemPoolIdx+len(data)]
		copy(dst, data)
		visuMemPoolIdx += len(data)
		return dst
	}
	return data
}

func visuLoadObject(name string) *visuObject {
	data := visuReadFile(name)
	if len(data) == 0 {
		return visuGetNewObject()
	}
	o := visuGetNewObject()
	o.flags = visuFDefault
	o.r = visuGetNewRMatrix()
	o.r0 = visuGetNewRMatrix()
	o.vnum = 0
	o.nnum = 0
	o.v0 = nil
	o.n0 = nil
	o.v = nil
	o.n = nil
	o.pv = nil
	o.plnum = 1

	for d := 0; d+8 <= len(data); {
		tag := string(data[d : d+4])
		length := int(binary.LittleEndian.Uint32(data[d+4 : d+8]))
		d0 := d
		d += 8
		if d+length > len(data) {
			break
		}
		payload := data[d : d+length]

		switch tag {
		case "END ":
			return o
		case "VERS":
			if len(payload) >= 2 {
				_ = int16(binary.LittleEndian.Uint16(payload[0:2]))
			}
		case "NAME":
			o.name = visuCString(payload)
		case "VERT":
			if len(payload) >= 4 {
				o.vnum = int(int16(binary.LittleEndian.Uint16(payload[0:2])))
				o.v0 = visuParseVList(payload[4:], o.vnum)
				o.v = visuAllocVList(o.vnum)
				o.pv = visuAllocPVList(o.vnum)
			}
		case "NORM":
			if len(payload) >= 4 {
				o.nnum = int(int16(binary.LittleEndian.Uint16(payload[0:2])))
				o.nnum1 = int(int16(binary.LittleEndian.Uint16(payload[2:4])))
				o.n0 = visuParseNList(payload[4:], o.nnum)
				o.n = visuAllocNList(o.nnum)
			}
		case "POLY":
			o.pd = payload
		default:
			if len(tag) >= 3 && tag[:3] == "ORD" {
				b := 0
				if tag[3] == '0' {
					b = 0
				} else if tag[3] == 'E' {
					b = o.plnum
					o.plnum++
				} else {
					break
				}
				o.pl[b] = visuParseU16(payload)
			}
		}

		d = d0 + length + 8
	}
	return o
}

func visuCString(data []byte) string {
	for i, b := range data {
		if b == 0 {
			return string(data[:i])
		}
	}
	return string(data)
}

func visuParseVList(data []byte, count int) []visuVList {
	out := make([]visuVList, count)
	stride := 16
	for i := 0; i < count; i++ {
		off := i * stride
		if off+stride > len(data) {
			break
		}
		out[i].x = int32(binary.LittleEndian.Uint32(data[off : off+4]))
		out[i].y = int32(binary.LittleEndian.Uint32(data[off+4 : off+8]))
		out[i].z = int32(binary.LittleEndian.Uint32(data[off+8 : off+12]))
		out[i].normal = int16(binary.LittleEndian.Uint16(data[off+12 : off+14]))
	}
	return out
}

func visuParseNList(data []byte, count int) []visuNList {
	out := make([]visuNList, count)
	stride := 8
	for i := 0; i < count; i++ {
		off := i * stride
		if off+stride > len(data) {
			break
		}
		out[i].x = int16(binary.LittleEndian.Uint16(data[off : off+2]))
		out[i].y = int16(binary.LittleEndian.Uint16(data[off+2 : off+4]))
		out[i].z = int16(binary.LittleEndian.Uint16(data[off+4 : off+6]))
	}
	return out
}

func visuParseU16(data []byte) []uint16 {
	count := len(data) / 2
	out := make([]uint16, count)
	for i := 0; i < count; i++ {
		off := i * 2
		out[i] = binary.LittleEndian.Uint16(data[off : off+2])
	}
	return out
}

func visuFxDot3Shift14(a0, b0, a1, b1, a2, b2 int32) int32 {
	acc := int64(a0) * int64(b0)
	acc += int64(a1) * int64(b1)
	acc += int64(a2) * int64(b2)
	return int32(acc >> visuUnitShr)
}

func visuM32(m *visuRMatrix, idx int) int32 {
	return m.m[idx]
}

func visuM16(m *visuRMatrix, idx int) int32 {
	return int32(int16(m.m[idx] & 0xFFFF))
}

func visuMulMatricesOverwriteRight(m1 *visuRMatrix, m2 *visuRMatrix) {
	var r [9]int32
	r[0] = visuFxDot3Shift14(m1.m[0], m2.m[0], m1.m[1], m2.m[3], m1.m[2], m2.m[6])
	r[1] = visuFxDot3Shift14(m1.m[0], m2.m[1], m1.m[1], m2.m[4], m1.m[2], m2.m[7])
	r[2] = visuFxDot3Shift14(m1.m[0], m2.m[2], m1.m[1], m2.m[5], m1.m[2], m2.m[8])
	r[3] = visuFxDot3Shift14(m1.m[3], m2.m[0], m1.m[4], m2.m[3], m1.m[5], m2.m[6])
	r[4] = visuFxDot3Shift14(m1.m[3], m2.m[1], m1.m[4], m2.m[4], m1.m[5], m2.m[7])
	r[5] = visuFxDot3Shift14(m1.m[3], m2.m[2], m1.m[4], m2.m[5], m1.m[5], m2.m[8])
	r[6] = visuFxDot3Shift14(m1.m[6], m2.m[0], m1.m[7], m2.m[3], m1.m[8], m2.m[6])
	r[7] = visuFxDot3Shift14(m1.m[6], m2.m[1], m1.m[7], m2.m[4], m1.m[8], m2.m[7])
	r[8] = visuFxDot3Shift14(m1.m[6], m2.m[2], m1.m[7], m2.m[5], m1.m[8], m2.m[8])
	m2.m = r
}

func visuRotateSingleVec(rm *visuRMatrix, src [3]int32) [3]int32 {
	M := rm.m
	x := src[0]
	y := src[1]
	z := src[2]
	return [3]int32{
		visuFxDot3Shift14(x, M[0], y, M[1], z, M[2]),
		visuFxDot3Shift14(x, M[3], y, M[4], z, M[5]),
		visuFxDot3Shift14(x, M[6], y, M[7], z, M[8]),
	}
}

func visuCalcApplyRMatrix(dest, apply *visuRMatrix) {
	visuMulMatricesOverwriteRight(apply, dest)
	tmp := [3]int32{dest.x, dest.y, dest.z}
	tmp = visuRotateSingleVec(apply, tmp)
	dest.x = tmp[0] + apply.x
	dest.y = tmp[1] + apply.y
	dest.z = tmp[2] + apply.z
}

func visuCalcRotate(count int, dest, source []visuVList, matrix *visuRMatrix) {
	if count <= 0 {
		return
	}
	for i := 0; i < count; i++ {
		sx := source[i].x
		sy := source[i].y
		sz := source[i].z
		x := visuFxDot3Shift14(sx, visuM16(matrix, 0), sy, visuM16(matrix, 1), sz, visuM16(matrix, 2)) + matrix.x
		y := visuFxDot3Shift14(sx, visuM16(matrix, 3), sy, visuM16(matrix, 4), sz, visuM16(matrix, 5)) + matrix.y
		z := visuFxDot3Shift14(sx, visuM16(matrix, 6), sy, visuM16(matrix, 7), sz, visuM16(matrix, 8)) + matrix.z
		dest[i].x = x
		dest[i].y = y
		dest[i].z = z
		dest[i].normal = source[i].normal
	}
}

func visuCalcNRotate(count int, dest, source []visuNList, matrix *visuRMatrix) {
	if count <= 0 {
		return
	}
	for i := 0; i < count; i++ {
		sx := int32(int16(source[i].x))
		sy := int32(int16(source[i].y))
		sz := int32(int16(source[i].z))
		x := visuFxDot3Shift14(sx, visuM16(matrix, 0), sy, visuM16(matrix, 1), sz, visuM16(matrix, 2))
		y := visuFxDot3Shift14(sx, visuM16(matrix, 3), sy, visuM16(matrix, 4), sz, visuM16(matrix, 5))
		z := visuFxDot3Shift14(sx, visuM16(matrix, 6), sy, visuM16(matrix, 7), sz, visuM16(matrix, 8))
		dest[i].x = int16(x)
		dest[i].y = int16(y)
		dest[i].z = int16(z)
	}
}

func visuCalcSingleZ(vertexnum int, vertexlist []visuVList, matrix *visuRMatrix) int32 {
	v := vertexlist[vertexnum]
	z := visuFxDot3Shift14(v.x, visuM32(matrix, 6), v.y, visuM32(matrix, 7), v.z, visuM32(matrix, 8))
	z += matrix.z
	return z
}

func visuCalcProject(count int, dest []visuPVList, source []visuVList) int {
	vfAll := 0xFFFF
	for i := 0; i < count; i++ {
		X := source[i].x
		Y := source[i].y
		Z := source[i].z
		vf := 0
		if Z < visuProjClipZ[0] {
			vf |= visuVFNear
			Z = visuProjClipZ[0]
		} else if Z > visuProjClipZ[1] {
			vf |= visuVFFar
		}
		y := int32(int64(visuProjMulY)*int64(Y)/int64(Z)) + visuProjAddY
		if y > visuProjClipY[1] {
			vf |= visuVFDown
		}
		if y < visuProjClipY[0] {
			vf |= visuVFUp
		}
		dest[i].y = int16(y)
		x := int32(int64(visuProjMulX)*int64(X)/int64(Z)) + visuProjAddX
		if x > visuProjClipX[1] {
			vf |= visuVFRight
		}
		if x < visuProjClipX[0] {
			vf |= visuVFLeft
		}
		dest[i].x = int16(x)
		dest[i].vf = visuVisfl(vf)
		vfAll &= vf
	}
	return vfAll
}

func visuCameraAngle(a visuAngle) {
	halfw := visuProjClipX[1] - visuProjAddX
	bx := int32(a) >> 1
	if bx < 8*64 {
		bx = 8 * 64
	}
	if bx >= 16384 {
		bx = 16383
	}
	bx = (bx >> (6 - 1)) &^ 1
	if len(visuAvistan) == 0 {
		return
	}
	av := visuAvistan[uint32(bx)>>1]
	p := int64(halfw) * int64(av)
	visuProjMulX = int32(p >> 8)
	p2 := int64(visuProjMulX) * int64(visuProjAspect)
	visuProjMulY = int32(p2 >> 8)
}

func visuWindow(x1, x2, y1, y2, z1, z2 int32) {
	visuProjClipX[0] = x1
	visuProjClipX[1] = x2
	visuProjAddX = (x1 + x2) >> 1
	visuProjClipY[0] = y1
	visuProjClipY[1] = y2
	visuProjAddY = (y1 + y2) >> 1
	visuProjClipZ[0] = z1
	visuProjClipZ[1] = z2
}

func visuClear255() {
	for y := 0; y < constants.ScreenHeight; y++ {
		row := y * constants.ScreenWidth
		if row+constants.ScreenWidth > len(shim.VRAM) {
			break
		}
		for i := 0; i < constants.ScreenWidth; i++ {
			shim.VRAM[row+i] = 0xFF
		}
	}
}

func visuClearBG(bg []byte) {
	if len(bg) < constants.ScreenSize {
		return
	}
	copy(shim.VRAM[:constants.ScreenSize], bg[:constants.ScreenSize])
}

func visuInit() {
	visuProjClipZ[0] = 256
	visuProjClipZ[1] = 1000000000
	visuProjClipX[0] = 0
	visuProjClipX[1] = constants.ScreenWidth - 1
	visuProjClipY[0] = 0
	visuProjClipY[1] = 199
	visuProjMulX = 250
	visuProjMulY = 220
	visuProjAddX = 160
	visuProjAddY = 100
	visuProjOversampleShr = 0
	visuProjAspect = 225
}

func visuCulledByBackface(n *visuNList, v0 *visuVList) bool {
	dot := int64(n.x)*int64(v0.x) + int64(n.y)*int64(v0.y) + int64(n.z)*int64(v0.z)
	return dot >= 0
}

func visuNormalLightU8(n *visuNList) uint16 {
	const Lx = 12118
	const Ly = 10603
	const Lz = 3030
	acc := int64(n.x)*Lx + int64(n.y)*Ly + int64(n.z)*Lz
	v := int(acc >> 21)
	v += 128
	if v < 0 {
		v = 0
	}
	if v > 255 {
		v = 255
	}
	return uint16(v)
}

func visuCalcLightQuant(flags uint16, n *visuNList) uint16 {
	f := flags & visuFShade32
	if f == 0 {
		return 0
	}
	nl := visuNormalLightU8(n)
	dx := f >> 10
	sh := 6 - dx
	q := nl >> sh
	if q < 1 {
		q = 1
	}
	if q > 30 {
		q = 30
	}
	return q
}

func visuDrawFillNrm(s []uint16) {
	if len(s) < 2 {
		return
	}
	color := uint8(s[0])
	startY := s[1]
	if int(startY) >= len(visuRowTable) {
		return
	}
	di := shim.VRAM[visuRowTable[startY]:]
	idx := 2
	var lx, rx, la, ra int32
	for {
		if idx >= len(s) {
			return
		}
		tag := int16(s[idx])
		idx++
		if tag < 0 {
			return
		}
		if tag != 0 {
			if idx+4 > len(s) {
				return
			}
			xs := uint32(s[idx]) | (uint32(s[idx+1]) << 16)
			idx += 2
			xa := uint32(s[idx]) | (uint32(s[idx+1]) << 16)
			idx += 2
			lx = int32(xs)
			la = int32(xa)
		}
		if idx >= len(s) {
			return
		}
		tag = int16(s[idx])
		idx++
		if tag < 0 {
			return
		}
		if tag != 0 {
			if idx+4 > len(s) {
				return
			}
			xs := uint32(s[idx]) | (uint32(s[idx+1]) << 16)
			idx += 2
			xa := uint32(s[idx]) | (uint32(s[idx+1]) << 16)
			idx += 2
			rx = int32(xs)
			ra = int32(xa)
		}
		if idx >= len(s) {
			return
		}
		cnt := int(s[idx])
		idx++
		if cnt <= 0 {
			return
		}
		for cnt > 0 {
			lx += la
			rx += ra
			xL := int(uint32(lx) >> 16)
			xR := int(uint32(rx) >> 16)
			if xL != xR {
				if xR < xL {
					xL, xR = xR, xL
				}
				if xL < 0 {
					xL = 0
				}
				if xR > visuRowLen {
					xR = visuRowLen
				}
				span := xR - xL
				if span > 0 {
					for i := 0; i < span; i++ {
						di[xL+i] = color
					}
				}
			}
			di = di[visuRowLen:]
			cnt--
		}
	}
}

func visuDrawFillGrd(s []uint16) {
	if len(s) < 2 {
		return
	}
	_ = s[0]
	startY := s[1]
	if int(startY) >= len(visuRowTable) {
		return
	}
	di := shim.VRAM[visuRowTable[startY]:]
	idx := 2
	var lx, rx, la, ra int32
	var lc, rc uint16
	var lca, rca int16
	for {
		if idx >= len(s) {
			return
		}
		tag := int16(s[idx])
		idx++
		if tag < 0 {
			return
		}
		if tag != 0 {
			if idx+6 > len(s) {
				return
			}
			lc = s[idx]
			lca = int16(s[idx+1])
			idx += 2
			lx = int32(uint32(s[idx]) | (uint32(s[idx+1]) << 16))
			idx += 2
			la = int32(uint32(s[idx]) | (uint32(s[idx+1]) << 16))
			idx += 2
		}
		if idx >= len(s) {
			return
		}
		tag = int16(s[idx])
		idx++
		if tag < 0 {
			return
		}
		if tag != 0 {
			if idx+6 > len(s) {
				return
			}
			rc = s[idx]
			rca = int16(s[idx+1])
			idx += 2
			rx = int32(uint32(s[idx]) | (uint32(s[idx+1]) << 16))
			idx += 2
			ra = int32(uint32(s[idx]) | (uint32(s[idx+1]) << 16))
			idx += 2
		}
		if idx >= len(s) {
			return
		}
		cnt := int(s[idx])
		idx++
		if cnt <= 0 {
			return
		}
		for cnt > 0 {
			lx += la
			rx += ra
			lc = uint16(int32(lc) + int32(lca))
			rc = uint16(int32(rc) + int32(rca))
			xL := int(uint32(lx) >> 16)
			xR := int(uint32(rx) >> 16)
			cL := int((lc >> 8) & 0xFF)
			cR := int((rc >> 8) & 0xFF)
			if xL != xR {
				if xR < xL {
					xL, xR = xR, xL
					cL, cR = cR, cL
				}
				if xL < 0 {
					xL = 0
				}
				if xR > visuRowLen {
					xR = visuRowLen
				}
				span := xR - xL
				if span > 0 {
					diff := cR - cL
					step := diff / span
					rem := diff % span
					acc := 0
					sgn := 1
					if rem < 0 {
						sgn = -1
						rem = -rem
					}
					c := cL
					p := di[xL:]
					for i := 0; i < span; i++ {
						p[i] = byte(c)
						acc += rem
						if acc >= span {
							c += step + sgn
							acc -= span
						} else {
							c += step
						}
					}
				}
			}
			di = di[visuRowLen:]
			cnt--
		}
	}
}

func visuVidDrawFill(stream []uint16) {
	if len(stream) == 0 {
		return
	}
	switch stream[0] {
	case 0:
		visuDrawFillNrm(stream[1:])
	case 1:
		visuDrawFillGrd(stream[1:])
	}
}

func visuEdgeSetupNrm(x0, x1 int16, height int, Xstart *uint32, Xadd *int32) {
	sh := int32(16 - visuPolyOversample)
	xs := int32(x0) << sh
	mask := ^int32(0xFFFF)
	xs = (xs & mask) | 0x8000
	xe := int32(x1) << sh
	xe = (xe & mask) | 0x8000
	pre := int32(0)
	if height > 0 {
		pre = int32((int64(xe-xs) / int64(height)))
	}
	sub := uint8(visuPolyCurY & uint16(visuPolyOversampleMask))
	if sub != 0 {
		xs -= int32(int64(pre) * int64(sub))
	}
	xa := pre << visuPolyOversample
	*Xstart = uint32(xs)
	*Xadd = xa
}

func visuEdgeSetupGrd(x0, x1 int16, height int, c0, c1 uint16, Cstart *uint16, Cadd *int16, Xstart *uint32, Xadd *int32) {
	*Cstart = c0
	ca := int16(0)
	if height > 0 {
		ca = int16((int32(c1) - int32(c0)) / int32(height))
	}
	*Cadd = ca
	sub := uint8(visuPolyCurY & uint16(visuPolyOversampleMask))
	if sub != 0 {
		*Cstart = uint16(int32(*Cstart) - int32(ca)*int32(sub))
	}
	visuEdgeSetupNrm(x0, x1, height, Xstart, Xadd)
}

func visuSwHeader(s *visuStream, color uint16, startY uint16) {
	s.p = append(s.p, color, startY)
}

func visuSwLeft(s *visuStream, present bool, xs uint32, xa int32) {
	if present {
		s.p = append(s.p, 1, uint16(xs&0xFFFF), uint16((xs>>16)&0xFFFF), uint16(uint32(xa)&0xFFFF), uint16(uint32(xa)>>16))
	} else {
		s.p = append(s.p, 0)
	}
}

func visuSwRight(s *visuStream, present bool, xs uint32, xa int32) {
	if present {
		s.p = append(s.p, 1, uint16(xs&0xFFFF), uint16((xs>>16)&0xFFFF), uint16(uint32(xa)&0xFFFF), uint16(uint32(xa)>>16))
	} else {
		s.p = append(s.p, 0)
	}
}

func visuSwCount(s *visuStream, rowsOS uint16) {
	s.p = append(s.p, rowsOS)
}

func visuSwEnd(s *visuStream) {
	s.p = append(s.p, 0xFFFF)
}

func visuPolyNrmBuild(P *visuPoly, out []uint16) int {
	top := int(P.v[0].y)
	bottom := top
	topi := 0
	for i := 1; i < int(P.sides); i++ {
		y := int(P.v[i].y)
		if y < top {
			top = y
			topi = i
		}
		if y > bottom {
			bottom = y
		}
	}
	visuPolyCurY = uint16(top)
	S := visuStream{p: out[:0]}
	visuSwHeader(&S, P.color, uint16(top>>visuPolyOversample))
	if top == bottom {
		visuSwEnd(&S)
		return len(S.p)
	}
	li, ri := topi, topi
	var lxs, rxs uint32
	var lxa, rxa int32
	var lh, rh uint16
	for {
		if P.v[li].y == int16(visuPolyCurY) {
			prev := li - 1
			if prev < 0 {
				prev = int(P.sides) - 1
			}
			dy := int(P.v[prev].y) - int(P.v[li].y)
			for dy == 0 {
				li = prev
				prev = li - 1
				if prev < 0 {
					prev = int(P.sides) - 1
				}
				dy = int(P.v[prev].y) - int(P.v[li].y)
				if li == topi {
					break
				}
			}
			if dy <= 0 {
				visuSwEnd(&S)
				return len(S.p)
			}
			lh = uint16(dy)
			visuEdgeSetupNrm(P.v[li].x, P.v[prev].x, dy, &lxs, &lxa)
			li = prev
			visuSwLeft(&S, true, lxs, lxa)
		} else {
			visuSwLeft(&S, false, 0, 0)
		}
		if P.v[ri].y == int16(visuPolyCurY) {
			next := ri + 1
			if next == int(P.sides) {
				next = 0
			}
			dy := int(P.v[next].y) - int(P.v[ri].y)
			for dy == 0 {
				ri = next
				next++
				if next == int(P.sides) {
					next = 0
				}
				dy = int(P.v[next].y) - int(P.v[ri].y)
				if ri == topi {
					break
				}
			}
			if dy <= 0 {
				visuSwEnd(&S)
				return len(S.p)
			}
			rh = uint16(dy)
			visuEdgeSetupNrm(P.v[ri].x, P.v[next].x, dy, &rxs, &rxa)
			ri = next
			visuSwRight(&S, true, rxs, rxa)
		} else {
			visuSwRight(&S, false, 0, 0)
		}
		shorter := lh
		if rh < shorter {
			shorter = rh
		}
		rowsOS := uint16(shorter >> visuPolyOversample)
		visuSwCount(&S, rowsOS)
		lh -= shorter
		rh -= shorter
		visuPolyCurY += shorter
		if int(visuPolyCurY) >= bottom {
			break
		}
	}
	visuSwEnd(&S)
	return len(S.p)
}

func visuPolyGrdBuild(P *visuPoly, out []uint16) int {
	top := int(P.v[0].y)
	bottom := top
	topi := 0
	for i := 1; i < int(P.sides); i++ {
		y := int(P.v[i].y)
		if y < top {
			top = y
			topi = i
		}
		if y > bottom {
			bottom = y
		}
	}
	visuPolyCurY = uint16(top)
	S := visuStream{p: out[:0]}
	visuSwHeader(&S, P.color, uint16(top>>visuPolyOversample))
	if top == bottom {
		visuSwEnd(&S)
		return len(S.p)
	}
	li, ri := topi, topi
	var lxs, rxs uint32
	var lxa, rxa int32
	var lc, rc uint16
	var lca, rca int16
	var lh, rh uint16
	for {
		if P.v[li].y == int16(visuPolyCurY) {
			prev := li - 1
			if prev < 0 {
				prev = int(P.sides) - 1
			}
			dy := int(P.v[prev].y) - int(P.v[li].y)
			for dy == 0 {
				li = prev
				prev = li - 1
				if prev < 0 {
					prev = int(P.sides) - 1
				}
				dy = int(P.v[prev].y) - int(P.v[li].y)
				if li == topi {
					break
				}
			}
			if dy <= 0 {
				visuSwEnd(&S)
				return len(S.p)
			}
			lh = uint16(dy)
			visuEdgeSetupGrd(P.v[li].x, P.v[prev].x, dy, P.GR[li], P.GR[prev], &lc, &lca, &lxs, &lxa)
			S.p = append(S.p, 1, lc, uint16(lca), uint16(lxs&0xFFFF), uint16((lxs>>16)&0xFFFF), uint16(uint32(lxa)&0xFFFF), uint16(uint32(lxa)>>16))
			li = prev
		} else {
			S.p = append(S.p, 0)
		}
		if P.v[ri].y == int16(visuPolyCurY) {
			next := ri + 1
			if next == int(P.sides) {
				next = 0
			}
			dy := int(P.v[next].y) - int(P.v[ri].y)
			for dy == 0 {
				ri = next
				next++
				if next == int(P.sides) {
					next = 0
				}
				dy = int(P.v[next].y) - int(P.v[ri].y)
				if ri == topi {
					break
				}
			}
			if dy <= 0 {
				visuSwEnd(&S)
				return len(S.p)
			}
			rh = uint16(dy)
			visuEdgeSetupGrd(P.v[ri].x, P.v[next].x, dy, P.GR[ri], P.GR[next], &rc, &rca, &rxs, &rxa)
			S.p = append(S.p, 1, rc, uint16(rca), uint16(rxs&0xFFFF), uint16((rxs>>16)&0xFFFF), uint16(uint32(rxa)&0xFFFF), uint16(uint32(rxa)>>16))
			ri = next
		} else {
			S.p = append(S.p, 0)
		}
		shorter := lh
		if rh < shorter {
			shorter = rh
		}
		rowsOS := uint16(shorter >> visuPolyOversample)
		visuSwCount(&S, rowsOS)
		lh -= shorter
		rh -= shorter
		visuPolyCurY += shorter
		if int(visuPolyCurY) >= bottom {
			break
		}
	}
	visuSwEnd(&S)
	return len(S.p)
}

func visuZClipIntersectXY(farV, nearV *visuVList, clipZ int32, outX, outY *int16) {
	dz := int64(farV.z) - int64(nearV.z)
	if dz == 0 {
		dz = 1
	}
	m := int64(farV.z) - int64(clipZ)
	dx := int64(nearV.x) - int64(farV.x)
	dy := int64(nearV.y) - int64(farV.y)
	Xi := int64(farV.x) + (dx*m)/dz
	Yi := int64(farV.y) + (dy*m)/dz
	x := int32((int64(visuProjMulX)*Xi)/int64(clipZ)) + visuProjAddX
	y := int32((int64(visuProjMulY)*Yi)/int64(clipZ)) + visuProjAddY
	*outX = int16(x)
	*outY = int16(y)
}

func visuClipNearReproject(in *visuPoly, out *visuPoly) uint8 {
	di := 0
	clipZ := visuProjClipZ[0]
	var newOr uint8
	for i := 0; i < int(in.sides); i++ {
		j := i + 1
		if j == int(in.sides) {
			j = 0
		}
		Vi := in.VX[i]
		Vj := in.VX[j]
		zi := Vi.z
		zj := Vj.z
		Ii := zi >= clipZ
		Ij := zj >= clipZ
		if Ii {
			out.v[di] = in.v[i]
			out.GR[di] = in.GR[i]
			out.VX[di] = in.VX[i]
			di++
		}
		if Ii != Ij {
			var Xc, Yc int16
			if zi >= zj {
				visuZClipIntersectXY(Vi, Vj, clipZ, &Xc, &Yc)
			} else {
				visuZClipIntersectXY(Vj, Vi, clipZ, &Xc, &Yc)
			}
			gI := in.GR[i]
			if in.flags&visuFGouraud != 0 {
				dz := int64(Vi.z) - int64(Vj.z)
				if dz != 0 {
					m := int64(Vi.z) - int64(clipZ)
					dg := int32(in.GR[j]) - int32(in.GR[i])
					gI = uint16(int32(in.GR[i]) + int32((int64(dg)*m)/dz))
				}
			}
			out.v[di].x = Xc
			out.v[di].y = Yc
			out.GR[di] = gI
			out.VX[di] = in.VX[i]
			if Yc < int16(visuProjClipY[0]) {
				newOr |= visuVFUp
			}
			if Yc > int16(visuProjClipY[1]) {
				newOr |= visuVFDown
			}
			if Xc < int16(visuProjClipX[0]) {
				newOr |= visuVFLeft
			}
			if Xc > int16(visuProjClipX[1]) {
				newOr |= visuVFRight
			}
			di++
		}
	}
	out.sides = uint16(di)
	out.vxseg = in.vxseg
	out.flags = in.flags
	out.color = in.color
	return newOr
}

func visuOutUp(x, y int16) bool    { return y < int16(visuProjClipY[0]) }
func visuOutDown(x, y int16) bool  { return y > int16(visuProjClipY[1]) }
func visuOutLeft(x, y int16) bool  { return x < int16(visuProjClipX[0]) }
func visuOutRight(x, y int16) bool { return x > int16(visuProjClipX[1]) }

func visuSideClip(in *visuPoly, out *visuPoly, isOut func(int16, int16) bool, horiz bool, clipv int32) {
	di := 0
	for i := 0; i < int(in.sides); i++ {
		j := i + 1
		if j == int(in.sides) {
			j = 0
		}
		xi := in.v[i].x
		yi := in.v[i].y
		xj := in.v[j].x
		yj := in.v[j].y
		oi := isOut(xi, yi)
		oj := isOut(xj, yj)
		if !oi {
			out.v[di] = in.v[i]
			out.GR[di] = in.GR[i]
			out.VX[di] = in.VX[i]
			di++
		}
		if oi != oj {
			var Xc, Yc int16
			if horiz {
				dy := int32(yj) - int32(yi)
				dx := int32(xj) - int32(xi)
				dist := int32(clipv) - int32(yi)
				if dy != 0 {
					slope := (int64(dist) << 14) / int64(dy)
					Xc = int16(int32(xi) + int32((int64(dx)*slope)>>14))
				} else {
					Xc = xi
				}
				Yc = int16(clipv)
			} else {
				dx := int32(xj) - int32(xi)
				dy := int32(yj) - int32(yi)
				dist := int32(clipv) - int32(xi)
				if dx != 0 {
					slope := (int64(dist) << 14) / int64(dx)
					Yc = int16(int32(yi) + int32((int64(dy)*slope)>>14))
				} else {
					Yc = yi
				}
				Xc = int16(clipv)
			}
			g0 := in.GR[i]
			g1 := in.GR[j]
			gI := g0
			den := int32(0)
			if horiz {
				den = int32(yj) - int32(yi)
			} else {
				den = int32(xj) - int32(xi)
			}
			if den != 0 {
				num := int32(0)
				if horiz {
					num = int32(clipv) - int32(yi)
				} else {
					num = int32(clipv) - int32(xi)
				}
				slopeG := (int64(num) << 14) / int64(den)
				dg := int32(g1) - int32(g0)
				gI = uint16(int32(g0) + int32((int64(dg)*slopeG)>>14))
			}
			out.v[di].x = Xc
			out.v[di].y = Yc
			out.GR[di] = gI
			out.VX[di] = in.VX[i]
			di++
		}
	}
	out.sides = uint16(di)
	out.vxseg = in.vxseg
	out.flags = in.flags
	out.color = in.color
}

func visuNewClip(P **visuPoly, tmp *visuPoly, visAnd, visOr uint8) {
	in := *P
	out := tmp
	if in.sides < 3 {
		in.sides = 0
		*P = in
		return
	}
	if visAnd&visuVFFar != 0 {
		in.sides = 0
		*P = in
		return
	}
	if visOr&visuVFNear != 0 {
		inserted := visuClipNearReproject(in, out)
		in, out = out, in
		if in.sides < 3 {
			in.sides = 0
			*P = in
			return
		}
		visOr = (visOr | inserted) & 0xFF
	} else {
		visOr &= (visuVFUp | visuVFDown | visuVFLeft | visuVFRight)
	}
	if visOr&visuVFUp != 0 {
		visuSideClip(in, out, func(x, y int16) bool { return visuOutUp(x, y) }, true, visuProjClipY[0])
		in, out = out, in
		if in.sides < 3 {
			in.sides = 0
			*P = in
			return
		}
	}
	if visOr&visuVFDown != 0 {
		visuSideClip(in, out, func(x, y int16) bool { return visuOutDown(x, y) }, true, visuProjClipY[1])
		in, out = out, in
		if in.sides < 3 {
			in.sides = 0
			*P = in
			return
		}
	}
	if visOr&visuVFLeft != 0 {
		visuSideClip(in, out, func(x, y int16) bool { return visuOutLeft(x, y) }, false, visuProjClipX[0])
		in, out = out, in
		if in.sides < 3 {
			in.sides = 0
			*P = in
			return
		}
	}
	if visOr&visuVFRight != 0 {
		visuSideClip(in, out, func(x, y int16) bool { return visuOutRight(x, y) }, false, visuProjClipX[1])
		in = out
		if in.sides < 3 {
			in.sides = 0
			*P = in
			return
		}
	}
	*P = in
}

func visuDrawPolyList(L []uint16, D []byte, V []visuVList, PV []visuPVList, N []visuNList, F uint16) {
	visuPolyOversample = uint8(visuProjOversampleShr)
	visuPolyOversample16 = uint8(16 - visuPolyOversample)
	visuPolyOversamples = uint8(1 << visuPolyOversample)
	visuPolyOversampleMask = visuPolyOversamples - 1
	if (F & visuFVisible) == 0 {
		return
	}
	if len(L) < 2 {
		return
	}
	lp := 2
	for lp < len(L) {
		off := L[lp]
		lp++
		if off == 0 {
			break
		}
		oi := int(off)
		if oi+6 > len(D) {
			continue
		}
		hdr := binary.LittleEndian.Uint16(D[oi : oi+2])
		sides := uint16(hdr & 0x00FF)
		flags := (hdr & (F | 0x0F00)) | (F & visuFVisible)
		colorw := binary.LittleEndian.Uint16(D[oi+2 : oi+4])
		if colorw == 0xFFFF {
			continue
		}
		color8 := uint8(colorw & 0x00FF)
		normalIdx := binary.LittleEndian.Uint16(D[oi+4 : oi+6])
		if (flags & visuF2Side) == 0 {
			v0i := binary.LittleEndian.Uint16(D[oi+6 : oi+8])
			if int(v0i) < len(V) && int(normalIdx) < len(N) {
				if visuCulledByBackface(&N[normalIdx], &V[v0i]) {
					continue
				}
			}
		}
		visuPoly1.sides = sides
		visuPoly1.flags = flags
		visuPoly1.vxseg = V
		anded := uint8(0xFF)
		orred := uint8(0x00)
		for i := 0; i < int(sides); i++ {
			vi := binary.LittleEndian.Uint16(D[oi+6+2*i : oi+8+2*i])
			if int(vi) >= len(PV) || int(vi) >= len(V) {
				continue
			}
			pvi := &PV[vi]
			visuPoly1.v[i].x = pvi.x
			visuPoly1.v[i].y = pvi.y
			visuPoly1.VX[i] = &V[vi]
			vf := uint8(pvi.vf & 0xFF)
			anded &= vf
			orred |= vf
		}
		if flags&visuFGouraud != 0 {
			visuPoly1.color = uint16(color8)
			for i := 0; i < int(sides); i++ {
				nv := uint16(visuPoly1.VX[i].normal)
				if int(nv) < len(N) {
					q := visuCalcLightQuant(flags, &N[nv])
					visuPoly1.GR[i] = uint16(((visuPoly1.color + q) & 0xFF) << 8)
				}
			}
		} else {
			q := uint16(0)
			if int(normalIdx) < len(N) {
				q = visuCalcLightQuant(flags, &N[normalIdx])
			}
			visuPoly1.color = uint16(color8+uint8(q)) & 0xFF
		}
		if anded != 0 {
			continue
		}
		cur := &visuPoly1
		tmp := &visuPoly2
		if orred != 0 {
			visuNewClip(&cur, tmp, anded, orred)
			if cur.sides < 3 {
				continue
			}
		}
		if flags&visuFGouraud != 0 {
			visuPolyDraw[0] = 1
			n := visuPolyGrdBuild(cur, visuPolyDraw[1:])
			visuVidDrawFill(visuPolyDraw[:n+1])
		} else {
			visuPolyDraw[0] = 0
			n := visuPolyNrmBuild(cur, visuPolyDraw[1:])
			visuVidDrawFill(visuPolyDraw[:n+1])
		}
	}
}

func visuDrawObject(o *visuObject) {
	if o == nil || (o.flags&visuFVisible) == 0 {
		return
	}
	visuCalcRotate(o.vnum, o.v, o.v0, o.r)
	if o.flags&visuFGouraud != 0 {
		visuCalcNRotate(o.nnum, o.n, o.n0, o.r)
	} else {
		visuCalcNRotate(o.nnum1, o.n, o.n0, o.r)
	}
	o.vf = visuVisfl(visuCalcProject(o.vnum, o.pv, o.v))
	if o.vf != 0 {
		return
	}
	a := 0
	al := int32(0x7fffffff)
	for b := 1; b < o.plnum; b++ {
		if len(o.pl[b]) < 2 {
			continue
		}
		c := o.pl[b][1]
		if int(c) >= len(o.v) {
			continue
		}
		bl := o.v[c].z
		if bl < al {
			al = bl
			a = b
		}
	}
	visuDrawPolyList(o.pl[a], o.pd, o.v, o.pv, o.n, o.flags)
}

func visuLsGet(f uint8) int32 {
	var l int32
	switch f & 3 {
	case 0:
		l = 0
	case 1:
		if len(visuSp) == 0 {
			return 0
		}
		l = int32(int8(visuSp[0]))
		visuSp = visuSp[1:]
	case 2:
		if len(visuSp) < 2 {
			return 0
		}
		l = int32(visuSp[0]) | (int32(int8(visuSp[1])) << 8)
		visuSp = visuSp[2:]
	case 3:
		if len(visuSp) < 4 {
			return 0
		}
		l = int32(visuSp[0]) | (int32(visuSp[1]) << 8) | (int32(visuSp[2]) << 16) | (int32(int8(visuSp[3])) << 24)
		visuSp = visuSp[4:]
	}
	return l
}

func visuResetScene() {
	if visuSclp < 0 || visuSclp >= visuScl {
		visuSclp = 0
	}
	visuSp = visuSceneList[visuSclp].data
	for i := 0; i < visuCoNum; i++ {
		if visuCo[i].o != nil {
			*visuCo[i].o.r = visuRMatrix{}
			*visuCo[i].o.r0 = visuRMatrix{}
		}
	}
	visuSclp++
	if visuSclp >= visuScl {
		visuSclp = 0
	}
}
