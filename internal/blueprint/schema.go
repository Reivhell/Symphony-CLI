package blueprint

// Blueprint adalah representasi lengkap dari satu file template.yaml
type Blueprint struct {
	SchemaVersion      string           `yaml:"schema_version" json:"schema_version"`
	Name               string           `yaml:"name" json:"name"`
	Version            string           `yaml:"version" json:"version"`
	Author             string           `yaml:"author" json:"author"`
	Description        string           `yaml:"description" json:"description"`
	MinSymphonyVersion string           `yaml:"min_symphony_version" json:"min_symphony_version"`
	Tags               []string         `yaml:"tags" json:"tags"`
	Extends            string           `yaml:"extends" json:"extends"`
	Validations        []ValidationRule `yaml:"validations" json:"validations"`
	Prompts            []Prompt         `yaml:"prompts" json:"prompts"`
	Actions            []Action         `yaml:"actions" json:"actions"`
	CompletionMessage  string           `yaml:"completion_message" json:"completion_message"`
	Plugins            []Plugin         `yaml:"plugins" json:"plugins"`
}

// Prompt merepresentasikan satu pertanyaan interaktif
type Prompt struct {
	ID        string   `yaml:"id" json:"id"`
	Question  string   `yaml:"question" json:"question"`
	Type      string   `yaml:"type" json:"type"` // "input" | "select" | "confirm" | "multiselect"
	Options   []string `yaml:"options" json:"options"`
	Default   any      `yaml:"default" json:"default"`
	DependsOn string   `yaml:"depends_on" json:"depends_on"` // Ekspresi kondisional
}

// Action merepresentasikan satu instruksi yang akan dieksekusi engine
type Action struct {
	Type       string `yaml:"type"` // "render" | "shell" | "ast-inject"
	Source     string `yaml:"source"`
	Target     string `yaml:"target"`
	If         string `yaml:"if"`      // Ekspresi kondisional opsional
	Command    string `yaml:"command"` // Untuk type "shell"
	WorkingDir string `yaml:"working_dir"`
	Strategy   string `yaml:"strategy"` // Untuk type "ast-inject"
	Anchor     string `yaml:"anchor"`
	Content    string `yaml:"content"`
}

// ValidationRule mendefinisikan aturan validasi untuk satu field prompt
type ValidationRule struct {
	Field   string `yaml:"field"`
	Rule    string `yaml:"rule"` // "regex" | "required" | "min_length"
	Pattern string `yaml:"pattern"`
	Message string `yaml:"message"`
}

// Plugin mendefinisikan custom renderer eksternal
type Plugin struct {
	Name       string   `yaml:"name"`
	Executable string   `yaml:"executable"`
	Handles    []string `yaml:"handles"`
}
