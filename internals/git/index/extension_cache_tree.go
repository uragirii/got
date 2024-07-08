package index

import (
	"fmt"
	"io"
	"strconv"

	"github.com/uragirii/got/internals/git/sha"
)

type CacheTree struct {
	relPath       string
	entryCount    int
	subTrees      []*CacheTree
	sha           *sha.SHA
	subTreesCount int
}

func parseCacheTreeArrToTree(cacheTreeArr *[]*CacheTree, startIdx int) (*CacheTree, int, error) {
	if startIdx > len(*cacheTreeArr) {
		return nil, 0, fmt.Errorf("invalid index %d", startIdx)
	}

	cacheTree := (*cacheTreeArr)[startIdx]

	subTrees := make([]*CacheTree, 0, cacheTree.subTreesCount)

	idx := startIdx + 1

	for range cacheTree.subTreesCount {
		subTree, newIdx, err := parseCacheTreeArrToTree(cacheTreeArr, idx)

		if err != nil {
			return nil, 0, err
		}

		subTrees = append(subTrees, subTree)

		idx = newIdx
	}

	cacheTree.subTrees = subTrees

	return cacheTree, idx, nil
}

func newCacheTree(treeContents *[]byte) (*CacheTree, error) {
	// tree, _, err := parseCacheTreeEntry(treeContents, 1)

	var cacheTrees []*CacheTree

	for idx := 0; idx < len(*treeContents)-sha.BYTES_LEN; {
		startIdx := idx

		for ; (*treeContents)[idx] != 0x00; idx++ {
		}

		relPath := string((*treeContents)[startIdx:idx])

		idx++

		startIdx = idx

		for ; (*treeContents)[idx] != ' '; idx++ {
		}

		entryCount, err := strconv.Atoi(string((*treeContents)[startIdx:idx]))

		if err != nil {
			return nil, err
		}

		idx++

		startIdx = idx

		for ; (*treeContents)[idx] != '\n'; idx++ {
		}

		subTreeCount, err := strconv.Atoi(string((*treeContents)[startIdx:idx]))

		if err != nil {
			return nil, err
		}

		idx++

		shaSlice := (*treeContents)[idx : idx+sha.BYTES_LEN]

		idx += sha.BYTES_LEN

		sha, err := sha.FromByteSlice(&shaSlice)

		if err != nil {
			return nil, err
		}

		cacheTree := &CacheTree{
			relPath:       relPath,
			entryCount:    entryCount,
			sha:           sha,
			subTreesCount: subTreeCount,
		}

		cacheTrees = append(cacheTrees, cacheTree)
	}

	cacheTree, _, err := parseCacheTreeArrToTree(&cacheTrees, 0)

	return cacheTree, err
}

func (tree CacheTree) Write(writer io.Writer) (int, error) {
	bytesWritten := 0
	n, _ := writer.Write([]byte(tree.relPath))
	bytesWritten += n
	n, _ = writer.Write([]byte{0x00})
	bytesWritten += n

	n, _ = writer.Write([]byte(fmt.Sprintf("%d", tree.entryCount)))
	bytesWritten += n

	n, _ = writer.Write([]byte{' '})
	bytesWritten += n

	n, _ = writer.Write([]byte(fmt.Sprintf("%d", tree.subTreesCount)))
	bytesWritten += n

	n, _ = writer.Write([]byte{'\n'})
	bytesWritten += n

	n, _ = writer.Write(*tree.sha.GetBytes())
	bytesWritten += n

	for _, subTree := range tree.subTrees {
		n, _ = subTree.Write(writer)
		bytesWritten += n
	}

	return bytesWritten, nil

}
