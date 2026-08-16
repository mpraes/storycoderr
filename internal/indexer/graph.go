package indexer

import (
	"context"
	"database/sql"
	"strings"

	"storycode/internal/analyzers/fastapi"
	"storycode/internal/analyzers/python"
	"storycode/internal/repository"
)

func persistGraph(ctx context.Context, tx *sql.Tx, opts Options, repoID, runID string, files []indexedFile, changedPython bool, kind string) error {
	if kind == "incremental" && !changedPython {
		return skipGraph(tx, repoID, runID, opts)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return extractAndSave(tx, opts, repoID, runID, files)
}

func extractAndSave(tx *sql.Tx, opts Options, repoID, runID string, files []indexedFile) error {
	pyFiles := pythonSources(files)
	report(opts.Out, "Extracting symbols...")
	extracted := python.Extract(pyFiles)
	report(opts.Out, "Detecting FastAPI routes...")
	routes := detectLocatedRoutes(pyFiles)
	report(opts.Out, "Extracting calls...")
	calls := python.ExtractCalls(pyFiles)
	report(opts.Out, "Completing index...")
	return saveGraph(tx, repoID, runID, extracted, calls, routes)
}

func skipGraph(tx *sql.Tx, repoID, runID string, opts Options) error {
	report(opts.Out, "Extracting symbols...")
	report(opts.Out, "Detecting FastAPI routes...")
	report(opts.Out, "Extracting calls...")
	report(opts.Out, "Completing index...")
	report(opts.Out, "Index kind: incremental")
	return repository.TouchGraph(tx, repoID, runID)
}

func pythonSources(files []indexedFile) []python.File {
	var out []python.File
	for _, file := range files {
		if file.ReadError != "" || !isPython(file.Path) {
			continue
		}
		out = append(out, python.File{Path: file.Path, Source: file.Content})
	}
	return out
}

type locatedRoute struct {
	File  string
	Route fastapi.Route
}

func detectLocatedRoutes(files []python.File) []locatedRoute {
	var out []locatedRoute
	for _, file := range files {
		for _, route := range fastapi.DetectRoutes(fastapi.File{Path: file.Path, Source: file.Source}) {
			out = append(out, locatedRoute{File: file.Path, Route: route})
		}
	}
	return out
}

func saveGraph(tx *sql.Tx, repoID, runID string, extracted python.Result, calls python.CallResult, routes []locatedRoute) error {
	if err := repository.ClearPythonGraph(tx, repoID); err != nil {
		return err
	}
	if err := insertSymbols(tx, repoID, runID, extracted.Symbols); err != nil {
		return err
	}
	ids, err := repository.LoadSymbolIDs(tx, repoID)
	if err != nil {
		return err
	}
	if err := insertRelations(tx, repoID, runID, extracted, calls.Relations, ids); err != nil {
		return err
	}
	hits, err := repository.LoadSymbolHits(tx, repoID)
	if err != nil {
		return err
	}
	return insertRoutes(tx, repoID, runID, routes, hits)
}

func insertSymbols(tx *sql.Tx, repoID, runID string, symbols []python.Symbol) error {
	for _, sym := range symbols {
		if err := insertOneSymbol(tx, repoID, runID, sym); err != nil {
			return err
		}
	}
	return nil
}

func insertOneSymbol(tx *sql.Tx, repoID, runID string, sym python.Symbol) error {
	id, err := repository.NewID()
	if err != nil {
		return err
	}
	fileID, err := repository.FileID(tx, repoID, sym.SourceFile)
	if err != nil {
		return err
	}
	return repository.InsertSymbol(tx, repository.SymbolWrite{
		ID:            id,
		RepositoryID:  repoID,
		SourceFileID:  fileID,
		QualifiedName: sym.QualifiedName,
		DisplayName:   sym.DisplayName,
		Kind:          string(sym.Kind),
		StartLine:     sym.StartLine,
		EndLine:       sym.EndLine,
		SemanticHash:  sym.SemanticHash,
		Confidence:    "high",
		LastSeenRunID: runID,
	})
}

func insertRelations(tx *sql.Tx, repoID, runID string, extracted python.Result, rels []python.CallRelation, ids map[string]string) error {
	fileIDs := symbolFileIDs(tx, repoID, extracted)
	for _, rel := range rels {
		if err := insertOneRelation(tx, repoID, runID, rel, ids, fileIDs); err != nil {
			return err
		}
	}
	return nil
}

func symbolFileIDs(tx *sql.Tx, repoID string, extracted python.Result) map[string]string {
	out := map[string]string{}
	for _, sym := range extracted.Symbols {
		id, err := repository.FileID(tx, repoID, sym.SourceFile)
		if err != nil {
			continue
		}
		out[sym.QualifiedName] = id
	}
	return out
}

func insertOneRelation(tx *sql.Tx, repoID, runID string, rel python.CallRelation, ids, fileIDs map[string]string) error {
	fromID := ids[rel.FromSymbol]
	if fromID == "" {
		return nil
	}
	id, err := repository.NewID()
	if err != nil {
		return err
	}
	return repository.InsertRelation(tx, repository.RelationWrite{
		ID:            id,
		RepositoryID:  repoID,
		FromSymbolID:  fromID,
		ToSymbolID:    ids[rel.ToSymbol],
		ToExternalRef: rel.ExternalRef,
		Kind:          rel.Kind,
		SourceFileID:  fileIDs[rel.FromSymbol],
		Line:          rel.Line,
		Confidence:    rel.Confidence,
		LastSeenRunID: runID,
	})
}

func insertRoutes(tx *sql.Tx, repoID, runID string, routes []locatedRoute, hits []repository.SymbolHit) error {
	for _, located := range routes {
		if err := insertOneRoute(tx, repoID, runID, located, hits); err != nil {
			return err
		}
	}
	return nil
}

func insertOneRoute(tx *sql.Tx, repoID, runID string, located locatedRoute, hits []repository.SymbolHit) error {
	id, err := repository.NewID()
	if err != nil {
		return err
	}
	route := located.Route
	return repository.UpsertEntryPoint(tx, repository.EntryWrite{
		ID:            id,
		RepositoryID:  repoID,
		HandlerID:     handlerID(located.File, route.HandlerSymbol, hits),
		Kind:          "http",
		Key:           route.EntryPointKey,
		Label:         route.Method + " " + route.Path,
		Method:        route.Method,
		Path:          route.Path,
		Framework:     route.Framework,
		Confidence:    route.Confidence,
		LastSeenRunID: runID,
	})
}

func handlerID(file, handler string, hits []repository.SymbolHit) string {
	for _, hit := range hits {
		if hit.Path != file {
			continue
		}
		if hit.DisplayName == handler || strings.HasSuffix(hit.QualifiedName, "."+handler) {
			return hit.ID
		}
	}
	return ""
}
