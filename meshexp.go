package meshexp

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"os"
	"strings"
)

// TreeReference is a reference to a MeSH term with the corresponding location in the tree.
type TreeReference struct {
	MedicalSubjectHeading string
	TreeLocation          []string
}

// Node is an element of the tree containing MeSH terms. It contains a Reference to a MeSH term, and any Children it
// may have.
type Node struct {
	Reference TreeReference
	Children  Tree
	Depth     int64
}

// Tree is used to represent the structure of MeSH.
type Tree map[string]Node

// MeSHTree is the structure of MeSH terms. It contains the tree structure of the ontology, as well as a mapping of
// heading to location in the tree for fast look up.
type MeSHTree struct {
	Tree      Tree
	Locations map[string][][]string
}

// New loads a MeSH tree from a file.
func New(meshTreeFilepath string) (*MeSHTree, error) {
	file, err := os.Open(meshTreeFilepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return MeSHTreeFromReader(file)
}

// Default loads the default MeSH tree file.
func Default() (*MeSHTree, error) {
	return MeSHTreeFromReader(strings.NewReader(mtrees2019))
}

// MeSHTreeFromReader loads a MeSH tree from any reader.
func MeSHTreeFromReader(reader io.Reader) (*MeSHTree, error) {
	tree := MeSHTree{
		Tree:      make(Tree),
		Locations: make(map[string][][]string),
	}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		ref, err := treeReferenceFromString(scanner.Text())
		if err != nil {
			return nil, err
		}
		// Add the first layer.
		if _, ok := tree.Tree[ref.TreeLocation[0]]; !ok {
			tree.Tree[ref.TreeLocation[0]] = Node{
				Reference: ref,
				Children:  make(Tree),
			}
		} else {
			// Add a child node to an existing node.
			tree.Tree[ref.TreeLocation[0]].addChild(ref.TreeLocation[1:], ref, 0)
		}

		// Remember the location for this heading.
		normalisedHeading := strings.ToLower(ref.MedicalSubjectHeading)
		tree.Locations[normalisedHeading] = append(tree.Locations[normalisedHeading], ref.TreeLocation)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &tree, nil
}

// Explode extracts specific MeSH terms from a given MeSH term (the terms indented beneath it in the tree structure).
func (t MeSHTree) Explode(term string) (terms []string) {
	if locations, ok := t.Locations[strings.ToLower(term)]; ok {
		for _, location := range locations {
			terms = append(terms, t.Tree.At(location).Terms()...)
		}
	}
	return
}

// Depth extracts the depth at which the term appears in the ontology.
func (t MeSHTree) Depth(term string) int64 {
	if locations, ok := t.Locations[strings.ToLower(term)]; ok {
		for _, location := range locations {
			return int64(len(location))
		}
	}
	return 0
}

// Contains indicates if a term is contained within the MeSH ontology.
func (t MeSHTree) Contains(term string) bool {
	if _, ok := t.Locations[strings.ToLower(term)]; ok {
		return true
	}
	return false
}

// Parents finds the parents for a particular mesh term.
// This method may return >1 term as one term may map to more than one parent of different names, for example when a
// a term is both a symptom of a disease and the description of a disease.
// see: https://meshb.nlm.nih.gov/record/ui?ui=D051474
func (t MeSHTree) Parents(term string) (parents []string) {
	if locations, ok := t.Locations[strings.ToLower(term)]; ok {
		for _, location := range locations {
			if len(location) > 2 {
				// Get the locations of all the parents.
				loc := location[:len(location)-1]
				p := t.Tree.At(location[:len(location)-2])
				for _, parent := range p {

					// Test that the locations are the same.
					eq := true
					for i, v := range loc {
						if v != parent.Reference.TreeLocation[i] {
							eq = false
							break
						}
					}

					if eq {
						parents = append(parents, parent.Reference.MedicalSubjectHeading)
					}
				}
			}
		}
	}
	return
}

func (t MeSHTree) Reference(term string) []TreeReference {
	var references []TreeReference
	if locations, ok := t.Locations[strings.ToLower(term)]; ok {
		for _, location := range locations {
			ref, err := treeReferenceFromString(fmt.Sprintf("%s;%s", term, strings.Join(location, ".")))
			if err != nil {
				panic(err)
			}
			references = append(references, ref)
		}
		return references
	}
	return nil
}

// Terms extracts the Medical Subject Headings from a tree and all of the children of that tree.
func (t Tree) Terms() (terms []string) {
	for _, node := range t {
		terms = append(terms, node.Reference.MedicalSubjectHeading)
		terms = append(terms, node.Children.Terms()...)
	}
	return
}

// At gets the part of the tree at the specified location.
func (t Tree) At(location []string) Tree {
	if len(location) == 0 {
		return t
	}
	if node, ok := t[location[0]]; ok {
		return node.Children.At(location[1:])
	}
	return Tree{}
}

// addChild adds a TreeReference at the specified location in the tree.
func (n Node) addChild(location []string, ref TreeReference, depth int64) {
	if innerNode, ok := n.Children[location[0]]; ok {
		innerNode.addChild(location[1:], ref, depth+1)
	} else {
		n.Children[location[0]] = Node{
			Reference: ref,
			Children:  make(Tree),
			Depth:     depth,
		}
	}
}

// treeReferenceFromString creates a TreeReference from a string.
func treeReferenceFromString(text string) (TreeReference, error) {
	parts := strings.Split(text, ";")
	if len(parts) != 2 {
		return TreeReference{}, errors.New(fmt.Sprintf("malformed tree reference %v", text))
	}

	locations := strings.Split(parts[1], ".")

	return TreeReference{
		MedicalSubjectHeading: parts[0],
		TreeLocation:          locations,
	}, nil
}
