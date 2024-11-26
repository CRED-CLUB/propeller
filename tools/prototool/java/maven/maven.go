package maven

import (
	"context"
	"fmt"
	"os"

	"github.com/CRED-CLUB/propeller/pkg/fs"
	"github.com/CRED-CLUB/propeller/pkg/process"

	"github.com/CRED-CLUB/propeller/pkg/logger"
)

// Config for maven
type Config struct {
	GroupID          string
	ArtifactID       string
	Name             string
	Version          string
	ReleaseURL       string
	ReleaseID        string
	ArtifactUser     string
	ArtifactPassword string
	PrivateKey       string
	PassPhrase       string
}

// Maven ...
type Maven struct {
	fs           fs.FileSystem
	interpolator *Interpolator
}

// NewMaven ...
func NewMaven(fs fs.FileSystem) (*Maven, error) {
	i, err := NewInterpolator(fs)
	if err != nil {
		return nil, err
	}
	return &Maven{
		fs:           fs,
		interpolator: i,
	}, nil
}

// Init maven
func (m *Maven) Init(ctx context.Context, version string) error {
	err := m.fs.CreateDir(fmt.Sprintf("%s/%s", "gen", "java"), os.FileMode(755), true)
	if err != nil {
		return err
	}

	mavenCfg := m.buildMavenConfig(ctx, version)
	err = m.interpolator.InterpolateLocalSettings(ctx, mavenCfg, fmt.Sprintf("%s/%s/%s", "gen", "java", "local-settings.xml"))
	if err != nil {
		return err
	}

	err = m.interpolator.InterpolatePom(ctx, mavenCfg, fmt.Sprintf("%s/%s/%s", "gen", "java", "pom.xml"))
	if err != nil {
		return err
	}

	return nil
}

// Build maven
func (m *Maven) Build(ctx context.Context) error {
	err := m.fs.Cd(fmt.Sprintf("%s/%s", "gen", "java"))
	if err != nil {
		logger.Ctx(ctx).Errorf("build failed %s", err)
		return err
	}

	_, err = process.Execute("mvn", "clean", "install")
	if err == nil {
		logger.Ctx(ctx).Info("Successfully compiled Java project")
		return err
	}
	return nil
}

// Publish maven
func (m *Maven) Publish(ctx context.Context) error {
	o, err := process.Execute("mvn", "clean", "deploy", "-s", "local-settings.xml")
	if err != nil {
		logger.Ctx(ctx).Errorf("failed to execute mvn deploy: %s output %s", err, o)
		return err
	}
	return err
}

func (m *Maven) buildMavenConfig(ctx context.Context, version string) *Config {
	mavenCfg := &Config{
		GroupID:          "io.github.abhishekvrshny",
		ArtifactID:       "propeller",
		Name:             "propeller",
		Version:          version,
		ReleaseURL:       "https://s01.oss.sonatype.org/content/repositories/releases/",
		ReleaseID:        "releases",
		ArtifactUser:     "",
		ArtifactPassword: "",
		PrivateKey:       "",
		PassPhrase:       "",
	}
	return mavenCfg
}
