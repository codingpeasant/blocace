package blocace

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
)

// MerkleTree represents a Merkle tree
type MerkleTree struct {
	RootNode *MerkleNode
}

// MerkleNode represents a Merkle tree node
type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
}

// NewMerkleTree creates a new Merkle tree from a sequence of data
func NewMerkleTree(txHashes [][]byte) *MerkleTree {
	var nodes []MerkleNode

	if len(txHashes)%2 != 0 {
		txHashes = append(txHashes, txHashes[len(txHashes)-1])
	}

	for _, txHash := range txHashes {
		node := NewMerkleNode(nil, nil, txHash)
		nodes = append(nodes, *node)
	}

	for len(nodes) > 1 {
		var newLevel []MerkleNode

		if len(nodes)%2 != 0 {
			nodes = append(nodes, nodes[len(nodes)-1])
		}

		for j := 0; j < len(nodes); j += 2 {
			node := NewMerkleNode(&nodes[j], &nodes[j+1], nil)
			newLevel = append(newLevel, *node)
		}

		nodes = newLevel
	}

	mTree := MerkleTree{&nodes[0]}

	return &mTree
}

// NewMerkleNode creates a new Merkle tree node
func NewMerkleNode(left, right *MerkleNode, txHash []byte) *MerkleNode {
	mNode := MerkleNode{}

	if left == nil && right == nil {
		mNode.Data = txHash
	} else {
		prevHashes := append(left.Data, right.Data...)
		hash := crypto.Keccak256(prevHashes)
		mNode.Data = hash[:]
	}

	mNode.Left = left
	mNode.Right = right

	return &mNode
}

// GetVerificationPath finds the necessary transaction hashes for clients to verify if a transaction has been included in the block
func (mt MerkleTree) GetVerificationPath(txToVerify []byte) map[int][]byte {
	nodes, index := mt.findNodeByData(txToVerify)

	if index <= 0 {
		return nil
	}

	var hashPath = make(map[int][]byte)
	cursor := index
	for cursor > 0 {
		if cursor%2 != 0 { // left node
			hashPath[cursor+1] = nodes[cursor+1].Data
		} else {
			hashPath[cursor-1] = nodes[cursor-1].Data
		}
		cursor = (cursor - 1) / 2
	}
	hashPath[cursor] = nodes[cursor].Data // root node

	return hashPath
}

// findNodeByData finds the tree node having the transation hash
func (mt MerkleTree) findNodeByData(hash []byte) ([]*MerkleNode, int) {
	if mt.RootNode != nil {
		var nodeBuffer []*MerkleNode
		var nodes []*MerkleNode
		var nodeIndex int

		nodeBuffer = append(nodeBuffer, mt.RootNode)

		for len(nodeBuffer) != 0 {
			node := nodeBuffer[0]
			nodes = append(nodes, node)
			nodeBuffer = nodeBuffer[1:]

			if bytes.Equal(node.Data, hash) {
				nodeIndex = len(nodes) - 1 // nodeIndex might not be unique - can be overridden after the first one is found
			}

			if node.Left != nil {
				nodeBuffer = append(nodeBuffer, node.Left)
			}
			if node.Right != nil {
				nodeBuffer = append(nodeBuffer, node.Right)
			}
		}

		return nodes, nodeIndex

	}

	return nil, -1
}

// walk traverses the tree in level by level
func (mt MerkleTree) walk() {
	if mt.RootNode != nil {
		var nodeBuffer []*MerkleNode
		nodeBuffer = append(nodeBuffer, mt.RootNode)

		for len(nodeBuffer) != 0 {
			node := nodeBuffer[0]
			nodeBuffer = nodeBuffer[1:]

			fmt.Printf("\nData: %x\n", node.Data)

			if node.Left != nil {
				nodeBuffer = append(nodeBuffer, node.Left)
			}
			if node.Right != nil {
				nodeBuffer = append(nodeBuffer, node.Right)
			}
		}
	}
}
