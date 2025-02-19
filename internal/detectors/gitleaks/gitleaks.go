package gitleaks

import (
	_ "embed"
	"log"
	"strings"

	"github.com/bearer/bearer/internal/detectors/types"
	"github.com/bearer/bearer/internal/parser/nodeid"
	"github.com/bearer/bearer/internal/report"
	"github.com/bearer/bearer/internal/report/secret"
	"github.com/bearer/bearer/internal/report/source"
	"github.com/bearer/bearer/internal/util/file"
	"github.com/pelletier/go-toml"
	"github.com/zricethezav/gitleaks/v8/config"
	"github.com/zricethezav/gitleaks/v8/detect"
)

//go:embed gitlab_config.toml
var RawConfig []byte

type detector struct {
	gitleaksDetector *detect.Detector
	idGenerator      nodeid.Generator
}

func New(idGenerator nodeid.Generator) types.Detector {
	var vc config.ViperConfig
	toml.Unmarshal(RawConfig, &vc) //nolint:all,errcheck
	cfg, err := vc.Translate()
	if err != nil {
		log.Fatal(err)
	}

	gitleaksDetector := detect.NewDetector(cfg)

	return &detector{
		gitleaksDetector: gitleaksDetector,
		idGenerator:      idGenerator,
	}
}

func (detector *detector) AcceptDir(dir *file.Path) (bool, error) {
	return true, nil
}

func (detector *detector) ProcessFile(file *file.FileInfo, dir *file.Path, report report.Report) (bool, error) {
	findings, err := detector.gitleaksDetector.DetectFiles(file.Path.AbsolutePath)

	if err != nil {
		return false, err
	}

	for _, finding := range findings {
		text := strings.TrimPrefix(finding.Line, "\n")
		report.AddSecretLeak(secret.Secret{
			Description: finding.Description,
		}, source.Source{
			Filename:          file.Path.RelativePath,
			StartLineNumber:   &finding.StartLine,
			StartColumnNumber: &finding.StartColumn,
			EndLineNumber:     &finding.EndLine,
			EndColumnNumber:   &finding.EndColumn,
			Text:              &text,
		})
	}

	return false, nil
}
