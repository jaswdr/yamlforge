package parser

import "time"

type Config struct {
	App      AppConfig              `yaml:"app"`
	Database DatabaseConfig         `yaml:"database"`
	Server   ServerConfig           `yaml:"server"`
	UI       UIConfig               `yaml:"ui"`
	Models   map[string]ModelConfig `yaml:"models"`
}

type AppConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
}

type DatabaseConfig struct {
	Type       string `yaml:"type"`
	Path       string `yaml:"path"`
	Connection string `yaml:"connection"`
}

type ServerConfig struct {
	Port int        `yaml:"port"`
	Host string     `yaml:"host"`
	CORS CORSConfig `yaml:"cors"`
	Auth AuthConfig `yaml:"auth"`
}

type CORSConfig struct {
	Enabled bool     `yaml:"enabled"`
	Origins []string `yaml:"origins"`
}

type AuthConfig struct {
	Type    string       `yaml:"type"`
	Secret  string       `yaml:"secret"`
	Expires string       `yaml:"expires"`
	Users   []UserConfig `yaml:"users"`
}

type UserConfig struct {
	Username    string                      `yaml:"username"`
	Password    string                      `yaml:"password"`
	Email       string                      `yaml:"email"`
	Role        string                      `yaml:"role"`
	Active      bool                        `yaml:"active"`
	Permissions map[string]EntityPermission `yaml:"permissions"`
}

type EntityPermission struct {
	Read  bool `yaml:"read"`
	Write bool `yaml:"write"`
}

type UIConfig struct {
	Theme  string `yaml:"theme"`
	Title  string `yaml:"title"`
	Logo   string `yaml:"logo"`
	Layout string `yaml:"layout"`
}

type ModelConfig struct {
	Fields      map[string]FieldConfig `yaml:"fields"`
	UI          *UIModelConfig         `yaml:"ui"`
	Permissions *PermissionsConfig     `yaml:"permissions"`
}

type FieldConfig struct {
	Type       string   `yaml:"type"`
	Primary    bool     `yaml:"primary"`
	Required   bool     `yaml:"required"`
	Unique     bool     `yaml:"unique"`
	Min        int      `yaml:"min"`
	Max        int      `yaml:"max"`
	Pattern    string   `yaml:"pattern"`
	Options    []string `yaml:"options"`
	Default    any      `yaml:"default"`
	AutoNow    bool     `yaml:"auto_now"`
	AutoNowAdd bool     `yaml:"auto_now_add"`
	Nullable   bool     `yaml:"nullable"`
	Index      bool     `yaml:"index"`
	To         string   `yaml:"to"`
	OnDelete   string   `yaml:"on_delete"`
	Items      string   `yaml:"items"`
}

type UIModelConfig struct {
	List *UIListConfig `yaml:"list"`
	Form *UIFormConfig `yaml:"form"`
}

type UIListConfig struct {
	Columns    []string `yaml:"columns"`
	Sortable   []string `yaml:"sortable"`
	Searchable []string `yaml:"searchable"`
}

type UIFormConfig struct {
	Fields []string `yaml:"fields"`
}

type PermissionsConfig struct {
	Create string `yaml:"create"`
	Read   string `yaml:"read"`
	Update string `yaml:"update"`
	Delete string `yaml:"delete"`
}

type FieldType string

const (
	FieldTypeText     FieldType = "text"
	FieldTypeNumber   FieldType = "number"
	FieldTypeBoolean  FieldType = "boolean"
	FieldTypeDatetime FieldType = "datetime"
	FieldTypeDate     FieldType = "date"
	FieldTypeTime     FieldType = "time"
	FieldTypeID       FieldType = "id"
	FieldTypeEmail    FieldType = "email"
	FieldTypePassword FieldType = "password"
	FieldTypePhone    FieldType = "phone"
	FieldTypeURL      FieldType = "url"
	FieldTypeSlug     FieldType = "slug"
	FieldTypeEnum     FieldType = "enum"
	FieldTypeColor    FieldType = "color"
	FieldTypeFile     FieldType = "file"
	FieldTypeImage    FieldType = "image"
	FieldTypeMarkdown FieldType = "markdown"
	FieldTypeJSON     FieldType = "json"
	FieldTypeArray    FieldType = "array"
	FieldTypeRelation FieldType = "relation"
	FieldTypeCurrency FieldType = "currency"
	FieldTypeLocation FieldType = "location"
	FieldTypeIP       FieldType = "ip"
	FieldTypeUUID     FieldType = "uuid"
	FieldTypeDuration FieldType = "duration"
)

func (f FieldType) String() string {
	return string(f)
}

func (f FieldType) IsValid() bool {
	switch f {
	case FieldTypeText, FieldTypeNumber, FieldTypeBoolean, FieldTypeDatetime,
		FieldTypeDate, FieldTypeTime, FieldTypeID, FieldTypeEmail,
		FieldTypePassword, FieldTypePhone, FieldTypeURL, FieldTypeSlug,
		FieldTypeEnum, FieldTypeColor, FieldTypeFile, FieldTypeImage,
		FieldTypeMarkdown, FieldTypeJSON, FieldTypeArray, FieldTypeRelation,
		FieldTypeCurrency, FieldTypeLocation, FieldTypeIP, FieldTypeUUID,
		FieldTypeDuration:
		return true
	}
	return false
}

func (f FieldType) SQLType() string {
	switch f {
	case FieldTypeID, FieldTypeNumber:
		return "INTEGER"
	case FieldTypeText, FieldTypeEmail, FieldTypePassword, FieldTypePhone,
		FieldTypeURL, FieldTypeSlug, FieldTypeEnum, FieldTypeColor,
		FieldTypeMarkdown, FieldTypeJSON, FieldTypeCurrency, FieldTypeIP,
		FieldTypeUUID, FieldTypeDuration:
		return "TEXT"
	case FieldTypeBoolean:
		return "BOOLEAN"
	case FieldTypeDatetime, FieldTypeDate, FieldTypeTime:
		return "DATETIME"
	case FieldTypeFile, FieldTypeImage:
		return "TEXT"
	case FieldTypeArray:
		return "TEXT"
	case FieldTypeRelation:
		return "INTEGER"
	case FieldTypeLocation:
		return "TEXT"
	default:
		return "TEXT"
	}
}

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

type Model struct {
	Name        string
	Fields      []Field
	Permissions Permissions
	UI          UIModel
}

type Field struct {
	Name       string
	Type       FieldType
	Primary    bool
	Required   bool
	Unique     bool
	Min        *int
	Max        *int
	Pattern    string
	Options    []string
	Default    any
	AutoNow    bool
	AutoNowAdd bool
	Nullable   bool
	Index      bool
	RelatedTo  string
	OnDelete   string
	ArrayType  string
}

type Permissions struct {
	Create string
	Read   string
	Update string
	Delete string
}

type UIModel struct {
	List UIList
	Form UIForm
}

type UIList struct {
	Columns    []string
	Sortable   []string
	Searchable []string
}

type UIForm struct {
	Fields []string
}

func DefaultConfig() *Config {
	return &Config{
		App: AppConfig{
			Name:    "My Application",
			Version: "1.0.0",
		},
		Database: DatabaseConfig{
			Type: "sqlite",
			Path: "./data.db",
		},
		Server: ServerConfig{
			Port: 8080,
			Host: "0.0.0.0",
			CORS: CORSConfig{
				Enabled: true,
				Origins: []string{"*"},
			},
			Auth: AuthConfig{
				Type: "none",
			},
		},
		UI: UIConfig{
			Theme:  "light",
			Title:  "My App",
			Layout: "sidebar",
		},
		Models: make(map[string]ModelConfig),
	}
}

type DatabaseType string

const (
	DatabaseSQLite DatabaseType = "sqlite"
)

type AuthType string

const (
	AuthNone  AuthType = "none"
	AuthBasic AuthType = "basic"
	AuthJWT   AuthType = "jwt"
)

type UITheme string

const (
	UIThemeLight UITheme = "light"
	UIThemeDark  UITheme = "dark"
	UIThemeAuto  UITheme = "auto"
)

type UILayout string

const (
	UILayoutSidebar UILayout = "sidebar"
	UILayoutTopbar  UILayout = "topbar"
	UILayoutMinimal UILayout = "minimal"
)

type RelationType string

const (
	RelationCascade  RelationType = "cascade"
	RelationRestrict RelationType = "restrict"
	RelationSetNull  RelationType = "set_null"
)

type RequestContext struct {
	UserID   string
	Role     string
	TenantID string
}

type QueryParams struct {
	Page     int
	PageSize int
	Sort     []SortField
	Filters  []Filter
	Search   string
}

type SortField struct {
	Field string
	Desc  bool
}

type Filter struct {
	Field    string
	Operator string
	Value    any
}

type APIResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
}

type Meta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalCount int64 `json:"total_count"`
	TotalPages int   `json:"total_pages"`
}

type ValidationRule interface {
	Validate(value any) error
}

type RequiredRule struct{}

func (r RequiredRule) Validate(value any) error {
	if value == nil || value == "" {
		return ValidationError{Message: "field is required"}
	}
	return nil
}

type MinLengthRule struct {
	Min int
}

func (r MinLengthRule) Validate(value any) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}
	if len(str) < r.Min {
		return ValidationError{Message: "value is too short"}
	}
	return nil
}

type MaxLengthRule struct {
	Max int
}

func (r MaxLengthRule) Validate(value any) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}
	if len(str) > r.Max {
		return ValidationError{Message: "value is too long"}
	}
	return nil
}

type Timestamp struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
