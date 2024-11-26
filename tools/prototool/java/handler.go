package java

import (
	"context"
	"github.com/CRED-CLUB/propeller/pkg/fs"

	"github.com/CRED-CLUB/propeller/pkg/logger"
	"github.com/CRED-CLUB/propeller/tools/prototool/java/maven"
)

// Handler ...
type Handler struct {
	maven *maven.Maven
}

// NewHandler ...
func NewHandler() (*Handler, error) {
	mvn, err := maven.NewMaven(&fs.LocalFileSystem{})
	if err != nil {
		return nil, err
	}
	return &Handler{mvn}, nil
}

// BuildAndPublish ...
func (h *Handler) BuildAndPublish(ctx context.Context, v string) error {
	err := h.maven.Init(ctx, v)
	if err != nil {
		logger.Ctx(ctx).Error(err.Error())
		return err
	}
	err = h.maven.Build(ctx)
	if err != nil {
		logger.Ctx(ctx).Error(err.Error())
		return err
	}
	err = h.maven.Publish(ctx)
	if err != nil {
		logger.Ctx(ctx).Error(err.Error())
		return err
	}
	return nil
}
