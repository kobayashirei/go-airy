package version

// Project metadata
const (
	// Name is the project name
	Name = "Airy"
	
	// Version is the current version
	Version = "v0.1.0"
	
	// Author is the project author
	Author = "Rei"
	
	// GitHub is the GitHub username
	GitHub = "kobayashirei"
	
	// Website is the project website
	Website = "iqwq.com"
	
	// Repository is the Go module path
	Repository = "github.com/kobayashirei/airy"
)

// Info returns all version information as a map
func Info() map[string]string {
	return map[string]string{
		"project":    Name,
		"version":    Version,
		"author":     Author,
		"github":     GitHub,
		"website":    Website,
		"repository": Repository,
	}
}
