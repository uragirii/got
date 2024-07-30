package index

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/uragirii/got/internals"
	"github.com/uragirii/got/internals/git/sha"
	objTree "github.com/uragirii/got/internals/git/tree"
)

type CacheTree struct {
	RelPath       string
	EntryCount    int
	SubTrees      []*CacheTree
	SHA           *sha.SHA
	SubTreesCount int
	IsInvalidated bool
}

func parseCacheTreeArrToTree(cacheTreeArr *[]*CacheTree, startIdx int) (*CacheTree, int, error) {
	if startIdx > len(*cacheTreeArr) {
		return nil, 0, fmt.Errorf("invalid index %d", startIdx)
	}

	cacheTree := (*cacheTreeArr)[startIdx]

	subTrees := make([]*CacheTree, 0, cacheTree.SubTreesCount)

	idx := startIdx + 1

	for range cacheTree.SubTreesCount {
		subTree, newIdx, err := parseCacheTreeArrToTree(cacheTreeArr, idx)

		if err != nil {
			return nil, 0, err
		}

		subTrees = append(subTrees, subTree)

		idx = newIdx
	}

	cacheTree.SubTrees = subTrees

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
				RelPath:       relPath,
				EntryCount:    entryCount,
				SubTreesCount: subTreeCount,
				IsInvalidated: true,
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
			RelPath:       relPath,
			EntryCount:    entryCount,
			SHA:           sha,
			SubTreesCount: subTreeCount,
			IsInvalidated: false,
		}

		cacheTrees = append(cacheTrees, cacheTree)
	}

	cacheTree, _, err := parseCacheTreeArrToTree(&cacheTrees, 0)

	return cacheTree, err
}

func (tree CacheTree) Write(writer io.Writer) (int, error) {
	bytesWritten := 0
	n, _ := writer.Write([]byte(tree.RelPath))
	bytesWritten += n
	n, _ = writer.Write([]byte{0x00})
	bytesWritten += n

	n, _ = writer.Write([]byte(fmt.Sprintf("%d", tree.EntryCount)))
	bytesWritten += n

	n, _ = writer.Write([]byte{' '})
	bytesWritten += n

	n, _ = writer.Write([]byte(fmt.Sprintf("%d", tree.SubTreesCount)))
	bytesWritten += n

	n, _ = writer.Write([]byte{'\n'})
	bytesWritten += n

	if !tree.IsInvalidated {
		n, _ = writer.Write(*tree.SHA.GetBytes())
		bytesWritten += n
	}

	sort.Slice(tree.SubTrees, func(i, j int) bool {
		// Sorted differently
		// @see https://github.com/git/git/blob/557ae147e6cdc9db121269b058c757ac5092f9c9/cache-tree.c#L47
		if len(tree.SubTrees[i].RelPath) < len(tree.SubTrees[j].RelPath) {
			return true
		}
		if len(tree.SubTrees[j].RelPath) < len(tree.SubTrees[i].RelPath) {
			return false
		}

		return tree.SubTrees[j].RelPath < tree.SubTrees[i].RelPath

	})

	for _, subTree := range tree.SubTrees {
		n, _ = subTree.Write(writer)
		bytesWritten += n
	}

	return bytesWritten, nil
}

func (tree *CacheTree) add(splittedFilePath []string) {

	if len(splittedFilePath) == 0 {
		tree.EntryCount = -1
		tree.IsInvalidated = true

		return
	}

	// base path
	if splittedFilePath[0] == "." && tree.RelPath == "" {
		tree.EntryCount = -1
		tree.IsInvalidated = true

		return
	}

	tree.EntryCount = -1
	tree.IsInvalidated = true

	first := splittedFilePath[0]

	for _, subTree := range tree.SubTrees {
		if subTree.RelPath == first {

			subTree.add(splittedFilePath[1:])
			return
		}
	}

	// new subtree
	emptySlice := make([]byte, sha.BYTES_LEN)
	sha, _ := sha.FromByteSlice(&emptySlice)

	subTree := CacheTree{
		RelPath:       first,
		EntryCount:    0,
		SubTrees:      make([]*CacheTree, 0),
		SHA:           sha,
		SubTreesCount: 0,
	}

	tree.SubTrees = append(tree.SubTrees, &subTree)
	tree.SubTreesCount++

	subTree.add(splittedFilePath[1:])
}

func (tree *CacheTree) Hydrate(basePath string, index *Index) (int, error) {

	if !tree.IsInvalidated || tree.EntryCount > 0 {
		tree.IsInvalidated = false
		return tree.EntryCount, nil
	}

	fmt.Println("invalidating ", tree.RelPath)

	gitDir, err := internals.GetGitDir()

	if err != nil {
		return 0, err
	}

	rootDir := path.Join(gitDir, "..")

	dirPath := path.Join(basePath, tree.RelPath)

	items, err := os.ReadDir(dirPath)

	if err != nil {
		return 0, err
	}

	entryCount := 0

	enteries := make([]objTree.TreeEntry, 0, len(items))

	relDirPath, err := filepath.Rel(rootDir, dirPath)

	if err != nil {
		return 0, err
	}

	for _, item := range items {
		if item.IsDir() {
			continue
		}

		// only deal while file items

		filePath := path.Join(relDirPath, item.Name())

		// file is in gitignore possibly
		if !index.Has(filePath) {
			fmt.Println("ignoring", filePath)
			continue
		}

		indexEntry := index.Get(filePath)

		enteries = append(enteries, objTree.TreeEntry{
			Name: item.Name(),
			SHA:  indexEntry.SHA,
			// TODO: Fixme check mode from index file and then check
			Mode: objTree.ModeNormal,
		})
		entryCount++

	}

	if tree.RelPath == "object" {
		fmt.Println("Entrycount", entryCount)
	}

	for _, subTree := range tree.SubTrees {
		subTreeEntryCount, err := subTree.Hydrate(dirPath, index)

		if err != nil {
			return 0, nil
		}
		fmt.Println("\thydrate", subTree.RelPath, subTreeEntryCount)

		entryCount += subTreeEntryCount

		enteries = append(enteries, objTree.TreeEntry{
			Name: subTree.RelPath,
			SHA:  subTree.SHA,
			Mode: objTree.ModeDir,
		})
	}

	gitTree, err := objTree.FromEnteries(enteries)

	if err != nil {
		fmt.Println("err gitTree", entryCount)

		return 0, nil
	}

	tree.EntryCount = entryCount

	tree.SHA = gitTree.SHA

	err = gitTree.WriteToFile()

	if err != nil {
		fmt.Println("err write", entryCount, err)

		return 0, nil
	}

	return entryCount, nil
}
