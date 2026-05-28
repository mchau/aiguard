package snapshot

import (
	"os"
	"path/filepath"
	"strings"
)

// TreeNode represents one file or directory in the snapshot tree.
type TreeNode struct {
	Path     string
	IsDir    bool
	Depth    int
	Children []*TreeNode
}

// Walk builds a tree of files up to maxDepth using the provided matcher.
func Walk(rootDir string, maxDepth int, matcher *Matcher) (*TreeNode, error) {
	root := &TreeNode{Path: ".", IsDir: true, Depth: 0}
	err := walkDir(rootDir, rootDir, root, 0, maxDepth, matcher)
	return root, err
}

func walkDir(baseDir, dir string, node *TreeNode, depth, maxDepth int, matcher *Matcher) error {
	if depth >= maxDepth {
		return nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		rel, _ := filepath.Rel(baseDir, filepath.Join(dir, entry.Name()))
		rel = filepath.ToSlash(rel)

		if matcher.ShouldIgnore(rel) {
			continue
		}
		// Also ignore paths ending in the entry name with trailing slash for dirs
		if entry.IsDir() && matcher.ShouldIgnore(rel+"/") {
			continue
		}

		child := &TreeNode{
			Path:  rel,
			IsDir: entry.IsDir(),
			Depth: depth + 1,
		}
		node.Children = append(node.Children, child)

		if entry.IsDir() {
			if err := walkDir(baseDir, filepath.Join(dir, entry.Name()), child, depth+1, maxDepth, matcher); err != nil {
				return err
			}
		}
	}
	return nil
}

// FlattenFiles returns all non-directory leaf nodes.
func FlattenFiles(root *TreeNode) []string {
	var files []string
	flatten(root, &files)
	return files
}

func flatten(n *TreeNode, out *[]string) {
	if !n.IsDir {
		*out = append(*out, n.Path)
	}
	for _, c := range n.Children {
		flatten(c, out)
	}
}

// RenderTree renders the tree as a markdown code block.
func RenderTree(root *TreeNode) string {
	var sb strings.Builder
	for _, child := range root.Children {
		renderNode(&sb, child, "")
	}
	return sb.String()
}

func renderNode(sb *strings.Builder, n *TreeNode, indent string) {
	name := filepath.Base(n.Path)
	if n.IsDir {
		sb.WriteString(indent + name + "/\n")
		for _, c := range n.Children {
			renderNode(sb, c, indent+"  ")
		}
	} else {
		sb.WriteString(indent + name + "\n")
	}
}
