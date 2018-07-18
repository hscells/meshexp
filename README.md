# meshexp

[![GoDoc](https://godoc.org/github.com/hscells/meshexp?status.svg)](https://godoc.org/github.com/hscells/meshexp)
[![Go Report Card](https://goreportcard.com/badge/github.com/hscells/meshexp)](https://goreportcard.com/report/github.com/hscells/meshexp)

meshexp is a library for analysing and exploding medical subject headings (MeSH). It comes with what 
is to the best of the authors knowledge the most up to date version of the MeSH ontology published.

An example use of this library is as follows:

```go
tree, err := Default()
if err != nil {
    t.Error(err)
}

tree.Explode("neuralgia, postherpetic")
```

This will return a subsumption of the particular term.

For more information, see the godoc, available by clicking the badge above.