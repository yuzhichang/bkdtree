package bkdtree

import (
	"encoding/binary"
	"io"
	"math"
)

type KdTreeExtNodeInfo struct {
	offset    uint64 //offset in file. It's a leaf if less than idxBegin, otherwise an intra node.
	numPoints uint64 //number of points of subtree rooted at this node
}

type KdTreeExtIntraNode struct {
	splitDim    uint32
	numStrips   uint32
	splitValues []uint64
	children    []KdTreeExtNodeInfo
}

func (n *KdTreeExtIntraNode) Read(r io.Reader) (err error) {
	//According to https://golang.org/pkg/encoding/binary/#Read,
	//"Data must be a pointer to a fixed-size value or a slice of fixed-size values."
	//Slice shall be adjusted to the expected length before calling binary.Read().
	//However this is not needed for binary.Write().
	err = binary.Read(r, binary.BigEndian, &n.splitDim)
	if err != nil {
		return
	}
	err = binary.Read(r, binary.BigEndian, &n.numStrips)
	if err != nil {
		return
	}
	n.splitValues = make([]uint64, n.numStrips-1)
	err = binary.Read(r, binary.BigEndian, &n.numStrips)
	if err != nil {
		return
	}
	n.children = make([]KdTreeExtNodeInfo, n.numStrips)
	err = binary.Read(r, binary.BigEndian, &n.children)
	return
}

type KdTreeExtMeta struct {
	numDims     uint32
	bytesPerDim uint32
	idxBegin    uint64
	numPoints   uint64 //the current number of points. Deleting points could trigger rebuilding the tree.
}

var KdMetaSize int64 = 8 * 3

type BkdTree struct {
	bkdCap      int // N in the paper. len(trees) shall be no larger than math.log2(bkdCap/t0mCap)
	t0mCap      int // M in the paper, the capacity of in-memory buffer
	numDims     int // number of point dimensions
	bytesPerDim int // number of bytes of each encoded dimension
	pointSize   int
	dir         string //directory of files which hold the persisted kdtrees
	prefix      string //prefix of file names
	numPoints   int
	t0m         []Point         // T0M in the paper, in-memory buffer.
	trees       []KdTreeExtMeta //persisted kdtrees
}

//NewBkdTree creates a BKDTree
func NewBkdTree(bkdCap, t0mCap, numDims, bytesPerDim int, dir, prefix string) (bkd *BkdTree) {
	treesCap := int(math.Log2(float64(bkdCap / t0mCap)))
	bkd = &BkdTree{
		bkdCap:      bkdCap,
		t0mCap:      t0mCap,
		numDims:     numDims,
		bytesPerDim: bytesPerDim,
		pointSize:   numDims*bytesPerDim + 8,
		dir:         dir,
		prefix:      prefix,
		t0m:         make([]Point, 0, t0mCap),
		trees:       make([]KdTreeExtMeta, 0, treesCap),
	}
	for i := 0; i < treesCap; i++ {
		kd := KdTreeExtMeta{
			numDims:     uint32(bkd.numDims),
			bytesPerDim: uint32(bkd.bytesPerDim),
			idxBegin:    0,
			numPoints:   0,
		}
		bkd.trees = append(bkd.trees, kd)
	}
	return
}
