package v3yaml

import (
	"github.com/bearer/bearer/internal/detectors/openapi/queries"
	"github.com/bearer/bearer/internal/detectors/openapi/reportadder"
	yamlparser "github.com/bearer/bearer/internal/detectors/openapi/yaml"
	"github.com/bearer/bearer/internal/parser"
	"github.com/bearer/bearer/internal/parser/nodeid"
	reporttypes "github.com/bearer/bearer/internal/report"
	"github.com/bearer/bearer/internal/report/operations/operationshelper"
	"github.com/bearer/bearer/internal/report/schema/schemahelper"
	"github.com/bearer/bearer/internal/util/file"
	"github.com/smacker/go-tree-sitter/yaml"
)

var queryParameters = parser.QueryMustCompile(yaml.GetLanguage(), `
(_
	(
      block_mapping_pair
        key:
            (flow_node) @helperName
            (#match? @helperName "^name$")
         value:
            (flow_node) @param_name
	)
	(
      block_mapping_pair
        key:
            (flow_node) @helperSchema
            (#match? @helperSchema "^schema$")
         value:
            (block_node (block_mapping) @param_schema)
	)
)
`)

func ProcessFile(idGenerator nodeid.Generator, file *file.FileInfo, report reporttypes.Report) (bool, error) {
	tree, err := parser.ParseFile(file, file.Path, yaml.GetLanguage())
	if err != nil {
		return false, err
	}
	defer tree.Close()

	nodeIDMap := nodeid.New(tree, idGenerator)
	nodeIDMap.Annotate()

	foundSchemas := make(map[parser.Node]*schemahelper.Schema)
	err = queries.AnnotateV3Paramaters(nodeIDMap, tree, foundSchemas, queryParameters)
	if err != nil {
		return false, err
	}

	err = yamlparser.AnnotateOperationId(nodeIDMap, tree, foundSchemas)
	if err != nil {
		return false, err
	}

	err = yamlparser.AnnotateObjects(nodeIDMap, tree, foundSchemas)
	if err != nil {
		return false, err
	}

	foundPaths := make(map[parser.Node]*operationshelper.Operation)
	err = yamlparser.AnnotatePaths(tree, foundPaths)
	if err != nil {
		return false, err
	}

	reportadder.AddSchema(file, report, foundSchemas, idGenerator)

	return true, err
}
