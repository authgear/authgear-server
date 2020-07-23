package newinteraction

type Result interface {
	result()
}

type ResultCommit struct {
	PreserveGraph *Graph
}

func (ResultCommit) result() {}

type ResultSave struct {
	Graph *Graph
}

func (ResultSave) result() {}
