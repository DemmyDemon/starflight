package shaders

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed *.kage
var kagen embed.FS

var shaders map[string]*ebiten.Shader

func addShader(name string, code []byte) error {
	if shaders == nil {
		shaders = make(map[string]*ebiten.Shader)
	}
	shader, err := ebiten.NewShader(code)
	if err != nil {
		return err
	}
	shaders[name] = shader
	return nil
}

func MustLoad() {
	kagefiles, err := kagen.ReadDir(".")
	if err != nil {
		panic(err)
	}
	for _, f := range kagefiles {
		name := f.Name()
		code, err := kagen.ReadFile(name)
		if err != nil {
			panic(err)
		}
		key := strings.TrimSuffix(name, filepath.Ext(name))
		err = addShader(key, code)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Shader %s compile error: %s\n", name, err.Error())
			os.Exit(1)
		}
	}
}

func Get(key string) *ebiten.Shader {
	return shaders[key]
}
