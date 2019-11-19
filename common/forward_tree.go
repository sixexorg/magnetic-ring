package common

type ForwardTree struct {
	spans    []*Span
	idx      int
	offset   uint64
	eof      bool
	width    uint64
	key      uint64
	original bool
	val      interface{}
}

type Span struct {
	L   uint64
	R   uint64
	Val interface{}
}

type FTreer interface {
	GetHeight() uint64
	GetVal() interface{}
}

func NewForwardTree(width, start, end uint64, ftrees []FTreer) *ForwardTree {
	//ass
	ft := &ForwardTree{
		width:    width,
		spans:    make([]*Span, 0, 20),
		key:      start,
		original: true,
	}
	if start == 0 {
		ft.original = false
	}
	sp := &Span{}
	ref := start
	ending := false
	for k, v := range ftrees {
		h := v.GetHeight()
		val := v.GetVal()
		if k == 0 {
			sp.L = start
			sp.R = h
			if h < start {
				sp.R = start
			}
			sp.Val = val
			ending = true
		} else {
			if h > ref {
				//last ending
				for h > ref {
					sp.R = ref
					ref = ref + width
				}
				l := ref - width
				//if k != 1 {
				spTmp := &Span{}
				DeepCopy(&spTmp, sp)
				ft.spans = append(ft.spans, spTmp)
				//}
				sp.L = l
				sp.Val = val
				ending = true

			} else {
				sp.R = h
				sp.Val = val
				ending = true
			}
		}
	}
	if ending {
		for ref <= end {
			sp.R = ref
			ref = ref + width
		}
		ft.spans = append(ft.spans, sp)
	}
	return ft
}

func (this *ForwardTree) Next() (eof bool) {
	if this.eof {
		return this.eof
	}
	f := true
Loop:
	lt := this.spans[this.idx]
	if f {
		if this.original {
			this.original = false
		} else {
			this.offset = this.offset + this.width
			this.key += this.width
			f = false
		}
	}
	lTmp := lt.L + this.offset
	if lTmp > lt.R {
		interval := lTmp - lt.R
		this.offset = interval
		this.idx++
		if this.idx+1 > len(this.spans) {
			this.eof = true
			return this.eof
		}
		goto Loop
	}
	this.val = this.spans[this.idx].Val
	return false
}

func (this *ForwardTree) Val() interface{} {
	return this.val
}
func (this *ForwardTree) Key() uint64 {
	return this.key
}
