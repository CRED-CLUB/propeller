package maven

import (
	"context"
	"github.com/CRED-CLUB/propeller/pkg/fs"
	"os"
	"text/template"

	"github.com/CRED-CLUB/propeller/pkg/logger"
)

// Interpolator ...
type Interpolator struct {
	localSettingsTemplate *template.Template
	pomTemplate           *template.Template
	fs                    fs.FileSystem
}

// NewInterpolator ...
func NewInterpolator(fs fs.FileSystem) (*Interpolator, error) {
	localSettingsContents, err := fs.ReadFile("tools/prototool/java/maven/local-settings.xml.hbs")
	if err != nil {
		return nil, err
	}

	pomContents, err := fs.ReadFile("tools/prototool/java/maven/pom.xml.hbs")
	if err != nil {
		return nil, err
	}

	localSettingsTemplate, err := template.New("local_settings_template").Parse(string(localSettingsContents))
	if err != nil {
		return nil, err
	}

	pomTemplate, err := template.New("pom_template").Parse(string(pomContents))
	if err != nil {
		return nil, err
	}

	return &Interpolator{
		localSettingsTemplate: localSettingsTemplate,
		pomTemplate:           pomTemplate,
		fs:                    fs,
	}, nil
}

// InterpolateLocalSettings ...
func (i *Interpolator) InterpolateLocalSettings(ctx context.Context, m *Config, destinationFilePath string) error {
	filePath, err := os.Create(destinationFilePath)
	if err != nil {
		logger.Ctx(ctx).Errorf("error in creating destination file path for %s %v", destinationFilePath, err)
		return err
	}

	m.ArtifactUser = os.Getenv("JAVA_ARTIFACT_USER_NAME_ENV")
	m.ArtifactPassword = os.Getenv("JAVA_ARTIFACT_PASSWORD_ENV")
	m.PassPhrase = os.Getenv("JAVA_PASS_PHRASE")

	err = i.localSettingsTemplate.Execute(filePath, m)
	if err != nil {
		logger.Ctx(ctx).Errorf("error in executing interpolation for %s %v", destinationFilePath, err)
		return err
	}
	return nil
}

// InterpolatePom ...
func (i *Interpolator) InterpolatePom(ctx context.Context, m *Config, destinationFilePath string) error {
	filePath, err := os.Create(destinationFilePath)
	if err != nil {
		logger.Ctx(ctx).Errorf("error in creating destination file path for %s %v", destinationFilePath, err)
		return err
	}

	err = i.pomTemplate.Execute(filePath, m)
	if err != nil {
		logger.Ctx(ctx).Errorf("error in executing interpolation for %s %v", destinationFilePath, err)
		return err
	}
	return nil
}
