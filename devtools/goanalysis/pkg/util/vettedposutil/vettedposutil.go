package vettedposutil

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/tools/go/analysis"
)

type VettedPositions struct {
	lock *sync.RWMutex
	// analyzer to set of positions.
	vetted map[string]map[token.Position]struct{}
	used   map[string]map[token.Position]struct{}
}

func NewEmptyVettedPositions() *VettedPositions {
	return &VettedPositions{
		lock:   &sync.RWMutex{},
		vetted: make(map[string]map[token.Position]struct{}),
		used:   make(map[string]map[token.Position]struct{}),
	}
}

func NewVettedPositionsFromFile(pathToFile string) (out *VettedPositions, err error) {
	f, err := os.Open(pathToFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	out = NewEmptyVettedPositions()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		// line is filename:line:column: analyzer1 [analyzer2 ...]
		parts := strings.Split(line, " ")
		if len(parts) < 2 {
			return nil, fmt.Errorf("filename:line:column: analyzer1 [analyzer2 ...]: %v", line)
		}

		analyzers := parts[1:]
		positionWithTrailingColon := parts[0]

		filenameLineColumn := strings.Split(positionWithTrailingColon, ":")
		if len(filenameLineColumn) != 4 {
			return nil, fmt.Errorf("filename:line:column: analyzer1 [analyzer2 ...]: %v", line)
		}

		lineNumber, err := strconv.Atoi(filenameLineColumn[1])
		if err != nil {
			return nil, fmt.Errorf("filename:line:column: analyzer1 [analyzer2 ...]: %v", line)
		}
		columnNumber, err := strconv.Atoi(filenameLineColumn[2])
		if err != nil {
			return nil, fmt.Errorf("filename:line:column: analyzer1 [analyzer2 ...]: %v", line)
		}

		position := token.Position{
			Filename: filenameLineColumn[0],
			Line:     lineNumber,
			Column:   columnNumber,
		}

		for _, analyzer := range analyzers {
			m, ok := out.vetted[analyzer]
			if !ok {
				m = make(map[token.Position]struct{})
				out.vetted[analyzer] = m
			}

			m[position] = struct{}{}
		}
	}
	err = scanner.Err()
	if err != nil {
		return
	}

	return
}

func (p *VettedPositions) CheckAndMarkUsed(analyzer string, position token.Position) bool {
	p.lock.Lock()
	defer p.lock.Unlock()

	m, ok := p.vetted[analyzer]
	if !ok {
		return false
	}

	for suffix := range m {
		if strings.HasSuffix(position.String(), suffix.String()) {

			used, ok := p.used[analyzer]
			if !ok {
				used = make(map[token.Position]struct{})
				p.used[analyzer] = used
			}
			used[suffix] = struct{}{}
			return true
		}
	}

	return false
}

func (p *VettedPositions) Err() error {
	var lines []string
	for analyzer, vetted := range p.vetted {
		used := p.used[analyzer]

		for position := range vetted {
			_, ok := used[position]
			if !ok {
				lines = append(lines, fmt.Sprintf("%v: %v\n", position, analyzer))
			}
		}
	}
	sort.Strings(lines)
	if len(lines) > 0 {
		return fmt.Errorf("unused vetted positions:\n%v", strings.Join(lines, ""))
	}
	return nil
}

func (p *VettedPositions) Report(pass *analysis.Pass, n *ast.File) {
	p.lock.Lock()
	defer p.lock.Unlock()

	filePosition := pass.Fset.Position(n.FileStart)
	var lines []string
	for analyzer, vetted := range p.vetted {
		for vettedPosition := range vetted {
			if strings.HasSuffix(filePosition.Filename, vettedPosition.Filename) {
				used := p.used[analyzer]
				_, ok := used[vettedPosition]
				if !ok {
					lines = append(lines, fmt.Sprintf("  %v: %v\n", vettedPosition, analyzer))
				}

			}
		}
	}
	sort.Strings(lines)
	if len(lines) > 0 {
		pass.Reportf(n.FileStart, "unused vetted positions:\n%v", strings.Join(lines, ""))
	}
}
