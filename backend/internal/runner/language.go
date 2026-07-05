package runner

type Language string

const (
	LangGo     Language = "go"
	LangPython Language = "python"
)

type Spec struct {
	Image    string
	Filename string
	Cmd      []string
}

func buildSpecs(goImage, pythonImage string) map[Language]Spec {
	return map[Language]Spec{
		LangGo: {
			Image:    goImage,
			Filename: "main.go",
			Cmd:      []string{"go", "run", "/workspace/main.go"},
		},
		LangPython: {
			Image:    pythonImage,
			Filename: "main.py",
			Cmd:      []string{"python", "/workspace/main.py"},
		},
	}
}

func ParseLanguage(s string) (Language, bool) {
	switch Language(s) {
	case LangGo, LangPython:
		return Language(s), true
	default:
		return "", false
	}
}
