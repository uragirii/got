package index

import (
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/uragirii/got/internals/git/sha"
)

type CacheTree struct {
	relPath       string
	entryCount    int
	subTrees      []*CacheTree
	sha           *sha.SHA
	subTreesCount int
	isInvalidated bool
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

		//An entry can be in an invalidated state and is represented by having
		//a negative number in the entry_count field. In this case, there is no
		//object name and the next entry starts immediately after the newline.
		//When writing an invalid entry, -1 should always be used as entry_count.
		if entryCount < 0 {
			cacheTree := &CacheTree{
				relPath:       relPath,
				entryCount:    entryCount,
				subTreesCount: subTreeCount,
				isInvalidated: true,
			}

			cacheTrees = append(cacheTrees, cacheTree)
			continue
		}

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
			isInvalidated: false,
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

	if !tree.isInvalidated {
		n, _ = writer.Write(*tree.sha.GetBytes())
		bytesWritten += n
	}

	sort.Slice(tree.subTrees, func(i, j int) bool {
		// Sorted differently
		// @see https://github.com/git/git/blob/557ae147e6cdc9db121269b058c757ac5092f9c9/cache-tree.c#L47
		if len(tree.subTrees[i].relPath) < len(tree.subTrees[j].relPath) {
			return true
		}
		if len(tree.subTrees[j].relPath) < len(tree.subTrees[i].relPath) {
			return false
		}

		return tree.subTrees[j].relPath < tree.subTrees[i].relPath

	})

	for _, subTree := range tree.subTrees {
		n, _ = subTree.Write(writer)
		bytesWritten += n
	}

	return bytesWritten, nil
}

func (tree *CacheTree) add(splittedFilePath []string) {

	if len(splittedFilePath) == 0 {
		tree.entryCount = -1
		tree.isInvalidated = true

		return
	}

	// base path
	if splittedFilePath[0] == "." && tree.relPath == "" {
		tree.entryCount = -1
		tree.isInvalidated = true

		return
	}

	tree.entryCount = -1
	tree.isInvalidated = true

	first := splittedFilePath[0]

	for _, subTree := range tree.subTrees {
		if subTree.relPath == first {

			subTree.add(splittedFilePath[1:])
			return
		}
	}

	// new subtree
	emptySlice := make([]byte, sha.BYTES_LEN)
	sha, _ := sha.FromByteSlice(&emptySlice)

	subTree := CacheTree{
		relPath:       first,
		entryCount:    0,
		subTrees:      make([]*CacheTree, 0),
		sha:           sha,
		subTreesCount: 0,
	}

	tree.subTrees = append(tree.subTrees, &subTree)
	tree.subTreesCount++

	subTree.add(splittedFilePath[1:])

}
