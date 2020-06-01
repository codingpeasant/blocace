package blockchain

import (
	"encoding/hex"
	"sort"
	"testing"
)

var hashStr []string = []string{"69292d123e8278e18e040fe7080898b4f6695413bd8890c851251b6646e4be82", "8841661dc86c2fbc2586f3f658b72713e371d89efae562d848f0ef4329a78280", "4da4d28f757484cb26ff94d94df6154d3676d33e00a0afd5dead650abe42c217", "7494edfee13f844b71cea5735f7566c2e01cca3f3be8746dd43551fc1fb67d0b", "a8af696e9eb5d84d5f504b190c7150e1ec1a0306c2453e1151937d9430dc18d9"}

func TestNewMerkleTree(t *testing.T) {
	expectedRootHash := "7e85ea1a1bc07d4a934661d0b78295617316d6e7363bce5a1e8d9e4557859437"

	var txHashes [][]byte
	for _, hash := range hashStr {
		decodedHash, _ := hex.DecodeString(hash)
		txHashes = append(txHashes, decodedHash)
	}

	merkleTree := NewMerkleTree(txHashes)

	if hex.EncodeToString(merkleTree.RootNode.Data) != expectedRootHash {
		t.Errorf("tree root hash expected: %s, actual: %x", expectedRootHash, merkleTree.RootNode.Data)
	}
}

func TestVerificationPath(t *testing.T) {
	txHashToFindStr := "a8af696e9eb5d84d5f504b190c7150e1ec1a0306c2453e1151937d9430dc18d9"
	txHashToFind, _ := hex.DecodeString(txHashToFindStr)

	var txHashes [][]byte
	for _, hash := range hashStr {
		decodedHash, _ := hex.DecodeString(hash)
		txHashes = append(txHashes, decodedHash)
	}

	merkleTree := NewMerkleTree(txHashes)
	_, index := merkleTree.findNodeByData(txHashToFind)
	if index != 14 {
		t.Errorf("findNodeByData expected: %d, actual: %d", 14, index)
	}

	txHashToVerifyStr := "8841661dc86c2fbc2586f3f658b72713e371d89efae562d848f0ef4329a78280"
	txHashToVerify, _ := hex.DecodeString(txHashToVerifyStr)

	verificationPath := merkleTree.GetVerificationPath(txHashToVerify)

	indices := make([]int, 0, len(verificationPath))
	for index := range verificationPath {
		indices = append(indices, index)
	}

	expectedIndices := []int{0, 2, 3, 9}
	sort.Ints(indices)

	for i := 0; i < len(expectedIndices); i++ {
		if expectedIndices[i] != indices[i] {
			t.Errorf("verificationPath expected: %d, actual: %d", expectedIndices, indices)
		}
	}
}
