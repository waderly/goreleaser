package nfpm

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/goreleaser/goreleaser/config"
	"github.com/goreleaser/goreleaser/context"
	"github.com/goreleaser/goreleaser/internal/artifact"
	"github.com/goreleaser/goreleaser/internal/testlib"
	"github.com/stretchr/testify/assert"
)

func TestDescription(t *testing.T) {
	assert.NotEmpty(t, Pipe{}.String())
}

func TestRunPipeNoFormats(t *testing.T) {
	var ctx = &context.Context{
		Git: context.GitInfo{
			CurrentTag: "v1.0.0",
		},
		Config:      config.Project{},
		Parallelism: runtime.NumCPU(),
	}
	testlib.AssertSkipped(t, Pipe{}.Run(ctx))
}

func TestRunPipeInvalidFormat(t *testing.T) {
	var ctx = context.New(config.Project{
		ProjectName: "nope",
		NFPM: config.NFPM{
			Bindir:       "/usr/bin",
			NameTemplate: defaultNameTemplate,
			Formats:      []string{"nope"},
			Files:        map[string]string{},
		},
	})
	ctx.Git = context.GitInfo{
		CurrentTag: "v1.2.3",
	}
	for _, goos := range []string{"linux", "darwin"} {
		for _, goarch := range []string{"amd64", "386"} {
			ctx.Artifacts.Add(artifact.Artifact{
				Name:   "mybin",
				Path:   "whatever",
				Goarch: goarch,
				Goos:   goos,
				Type:   artifact.Binary,
			})
		}
	}
	assert.Contains(t, Pipe{}.Run(ctx).Error(), `no packager registered for the format nope`)
}

func TestRunPipe(t *testing.T) {
	folder, err := ioutil.TempDir("", "archivetest")
	assert.NoError(t, err)
	var dist = filepath.Join(folder, "dist")
	assert.NoError(t, os.Mkdir(dist, 0755))
	assert.NoError(t, os.Mkdir(filepath.Join(dist, "mybin"), 0755))
	var binPath = filepath.Join(dist, "mybin", "mybin")
	_, err = os.Create(binPath)
	assert.NoError(t, err)
	var ctx = context.New(config.Project{
		ProjectName: "mybin",
		Dist:        dist,
		NFPM: config.NFPM{
			Bindir:       "/usr/bin",
			NameTemplate: defaultNameTemplate,
			Formats:      []string{"deb", "rpm"},
			Dependencies: []string{"make"},
			Recommends:   []string{"svn"},
			Suggests:     []string{"bzr"},
			Conflicts:    []string{"git"},
			Description:  "Some description",
			License:      "MIT",
			Maintainer:   "me@me",
			Vendor:       "asdf",
			Homepage:     "https://goreleaser.github.io",
			Files: map[string]string{
				"./testdata/testfile.txt": "/usr/share/testfile.txt",
			},
			ConfigFiles: map[string]string{
				"./testdata/testfile.txt": "/etc/nope.conf",
			},
		},
	})
	ctx.Git = context.GitInfo{CurrentTag: "v1.0.0"}
	for _, goos := range []string{"linux", "darwin"} {
		for _, goarch := range []string{"amd64", "386"} {
			ctx.Artifacts.Add(artifact.Artifact{
				Name:   "mybin",
				Path:   binPath,
				Goarch: goarch,
				Goos:   goos,
				Type:   artifact.Binary,
			})
		}
	}
	assert.NoError(t, Pipe{}.Run(ctx))
	assert.Len(t, ctx.Config.NFPM.Files, 1, "should not modify the config file list")
}

func TestInvalidNameTemplate(t *testing.T) {
	var ctx = &context.Context{
		Parallelism: runtime.NumCPU(),
		Artifacts:   artifact.New(),
		Config: config.Project{
			NFPM: config.NFPM{
				NameTemplate: "{{.Foo}",
				Formats:      []string{"deb"},
			},
		},
	}
	ctx.Artifacts.Add(artifact.Artifact{
		Name:   "mybin",
		Goos:   "linux",
		Goarch: "amd64",
		Type:   artifact.Binary,
	})
	assert.Contains(t, Pipe{}.Run(ctx).Error(), `template: {{.Foo}:1: unexpected "}" in operand`)
}

func TestCreateFileDoesntExist(t *testing.T) {
	folder, err := ioutil.TempDir("", "archivetest")
	assert.NoError(t, err)
	var dist = filepath.Join(folder, "dist")
	assert.NoError(t, os.Mkdir(dist, 0755))
	assert.NoError(t, os.Mkdir(filepath.Join(dist, "mybin"), 0755))
	var ctx = context.New(config.Project{
		Dist:        dist,
		ProjectName: "asd",
		NFPM: config.NFPM{
			Formats: []string{"deb", "rpm"},
			Files: map[string]string{
				"testdata/testfile.txt": "/var/lib/test/testfile.txt",
			},
		},
	})
	ctx.Git = context.GitInfo{
		CurrentTag: "v1.2.3",
	}
	ctx.Artifacts.Add(artifact.Artifact{
		Name:   "mybin",
		Path:   filepath.Join(dist, "mybin", "mybin"),
		Goos:   "linux",
		Goarch: "amd64",
		Type:   artifact.Binary,
	})
	assert.Contains(t, Pipe{}.Run(ctx).Error(), `dist/mybin/mybin: file does not exist`)
}

func TestInvalidConfig(t *testing.T) {
	folder, err := ioutil.TempDir("", "archivetest")
	assert.NoError(t, err)
	var dist = filepath.Join(folder, "dist")
	assert.NoError(t, os.Mkdir(dist, 0755))
	assert.NoError(t, os.Mkdir(filepath.Join(dist, "mybin"), 0755))
	var ctx = context.New(config.Project{
		Dist: dist,
		NFPM: config.NFPM{
			Formats: []string{"deb"},
		},
	})
	ctx.Artifacts.Add(artifact.Artifact{
		Name:   "mybin",
		Path:   filepath.Join(dist, "mybin", "mybin"),
		Goos:   "linux",
		Goarch: "amd64",
		Type:   artifact.Binary,
	})
	assert.Contains(t, Pipe{}.Run(ctx).Error(), `invalid nfpm config: package name cannot be empty`)
}

func TestDefault(t *testing.T) {
	var ctx = &context.Context{
		Config: config.Project{
			NFPM: config.NFPM{},
		},
	}
	assert.NoError(t, Pipe{}.Default(ctx))
	assert.Equal(t, "/usr/local/bin", ctx.Config.NFPM.Bindir)
	assert.Equal(t, defaultNameTemplate, ctx.Config.NFPM.NameTemplate)
}

func TestDefaultSet(t *testing.T) {
	var ctx = &context.Context{
		Config: config.Project{
			NFPM: config.NFPM{
				Bindir:       "/bin",
				NameTemplate: "foo",
			},
		},
	}
	assert.NoError(t, Pipe{}.Default(ctx))
	assert.Equal(t, "/bin", ctx.Config.NFPM.Bindir)
	assert.Equal(t, "foo", ctx.Config.NFPM.NameTemplate)
}
