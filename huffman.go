// huffman.go
// Package main provides functionality for constructing a Huffman tree, generating a code table,
// and handling serialization for LZ77-like compression algorithms. It defines the Node structure
// for the Huffman tree, a priority queue for tree construction, and methods for creating and
// managing Huffman codes.

package main

import (
	"container/heap"
	"fmt"
	"io"
)

// Node represents a node in the Huffman tree.
// It can be either a leaf node containing a byte value or an internal node with child nodes.
type Node struct {
	value       byte  // The byte value (only for leaf nodes).
	freq        int   // Frequency of the byte or combined frequency for internal nodes.
	Left, Right *Node // Child nodes (nil for leaf nodes).
	isLeaf      bool  // Indicates whether the node is a leaf.
	id          int   // Unique identifier for graphviz representation.
}

// NewNode creates a new Node instance.
// Parameters:
// - id: Unique identifier for the node.
// - val: Byte value (used only for leaf nodes).
// - freq: Frequency of the byte or combined frequency for internal nodes.
// - l: Left child node (nil for leaf nodes).
// - r: Right child node (nil for leaf nodes).
func NewNode(id int, val byte, freq int, l, r *Node) *Node {
	return &Node{
		id:     id,
		value:  val,
		freq:   freq,
		Left:   l,
		Right:  r,
		isLeaf: l == nil && r == nil,
	}
}

// DumpGraphviz writes the Graphviz representation of the Huffman tree to the provided writer.
func (n *Node) DumpGraphviz(w io.Writer) {
	w.Write([]byte("Digraph g {\n"))
	w.Write([]byte(n.getGraphviz()))
	w.Write([]byte("}\n"))
}

// getGraphviz recursively generates the Graphviz representation of the Huffman tree.
func (n *Node) getGraphviz() string {
	repr := fmt.Sprintf("\t%d[label=\"value=%d freq=%d\"]\n", n.id, n.value, n.freq)
	if n.Left != nil {
		repr += fmt.Sprintf("\t%d -> %d[label=\"0\"]\n", n.id, n.Left.id)
		repr += n.Left.getGraphviz()
	}
	if n.Right != nil {
		repr += fmt.Sprintf("\t%d -> %d[label=\"1\"]\n", n.id, n.Right.id)
		repr += n.Right.getGraphviz()
	}
	return repr
}

// PriorityQueue implements heap.Interface and holds Nodes.
type PriorityQueue []*Node

func (pq PriorityQueue) Len() int            { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool  { return pq[i].freq < pq[j].freq }
func (pq PriorityQueue) Swap(i, j int)       { pq[i], pq[j] = pq[j], pq[i] }
func (pq *PriorityQueue) Push(x interface{}) { *pq = append(*pq, x.(*Node)) }
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	x := old[n-1]
	*pq = old[0 : n-1]
	return x
}

// RemoveEmpty removes nodes with zero frequency from the priority queue.
func (pq *PriorityQueue) RemoveEmpty() PriorityQueue {
	notEmptyPQ := make(PriorityQueue, 0, 256)
	for _, node := range *pq {
		if node.freq != 0 {
			notEmptyPQ = append(notEmptyPQ, node)
		}
	}
	return notEmptyPQ
}

// constructHuffmanTree creates a Huffman tree based on the frequencies of bytes in the Values.
// It returns the root node of the Huffman tree.
func constructHuffmanTree(values []Value) *Node {
	freqs := make(PriorityQueue, 256)
	var idCounter int // Unique ID counter for nodes.

	// Initialize the frequency of each byte to zero.
	for i := 0; i < 256; i++ {
		freqs[i] = &Node{
			value:  byte(i),
			freq:   0,
			isLeaf: true,
			id:     idCounter,
		}
		idCounter++
	}

	// Calculate frequencies based on the Values.
	for _, v := range values {
		if v.IsLiteral {
			freqs[v.GetLiteralBinary()].freq += 1
		} else {
			for _, b := range v.GetPointerBinary() {
				freqs[b].freq += 1
			}
		}
	}

	// Remove nodes with zero frequency.
	freqs = freqs.RemoveEmpty()

	// Initialize the heap.
	heap.Init(&freqs)

	// Build the Huffman tree.
	for freqs.Len() > 1 {
		// Pop two nodes with the smallest frequencies.
		right := heap.Pop(&freqs).(*Node)
		left := heap.Pop(&freqs).(*Node)

		// Create a new internal node with these two nodes as children.
		newNode := NewNode(idCounter, 0, left.freq+right.freq, left, right)
		idCounter++

		// Push the new node back into the heap.
		heap.Push(&freqs, newNode)
	}

	// The remaining node is the root of the Huffman tree.
	root := heap.Pop(&freqs).(*Node)
	return root
}

// Code represents a binary code with its associated bit length.
// It is used as a key in the valTable map for reverse lookup during deserialization.
type Code struct {
	c    uint64 // The binary code value.
	bits byte   // The number of bits in the code.
}

// String returns a string representation of the Code.
func (c Code) String() string {
	return fmt.Sprintf("%08b(%d)\n", c.c, c.bits)
}

// addBit appends a single bit to a Code struct, shifting existing bits to make room.
// Parameters:
// - c: The current Code struct.
// - bit: The bit to append (true for 1, false for 0).
// Returns:
// - The updated Code struct with the new bit appended.
func addBit(c Code, bit bool) Code {
	var b uint64
	if bit {
		b = 1
	}
	return Code{
		c:    (c.c << 1) | b,
		bits: c.bits + 1,
	}
}

// CodeTable maps byte values to their corresponding binary codes.
// It is used by BinaryWriter to serialize data and by BinaryReader to deserialize data.
type CodeTable map[byte]Code

// createCodeTable generates a CodeTable from the Huffman tree.
// It traverses the tree to assign binary codes to each byte value based on their position in the tree.
// Parameters:
// - root: The root node of the Huffman tree.
// - prefix: The current binary code prefix during traversal.
// Returns:
// - A CodeTable mapping byte values to their binary codes.
func createCodeTable(root *Node, prefix Code) CodeTable {
	codeTable := make(CodeTable)
	if root.isLeaf {
		codeTable[root.value] = prefix
		return codeTable
	}
	if root.Left != nil {
		leftTable := createCodeTable(root.Left, addBit(prefix, false))
		codeTable = mergeTables(codeTable, leftTable)
	}
	if root.Right != nil {
		rightTable := createCodeTable(root.Right, addBit(prefix, true))
		codeTable = mergeTables(codeTable, rightTable)
	}
	return codeTable
}

// mergeTables merges two CodeTables into one.
// If there are overlapping keys, the entries from table 'b' will overwrite those in 'a'.
func mergeTables(a, b CodeTable) CodeTable {
	for k, v := range b {
		a[k] = v
	}
	return a
}
