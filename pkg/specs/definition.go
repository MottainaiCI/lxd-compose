package specs

type LxdCEnvironment struct {
	Version string `json:"version,omitempty" yaml:"version,omitempty"`

	TemplateEngine LxdCTemplateEngine `json:"template_engine,omitempty" yaml:"template_engine,omitempty"`

	Projects LxdCProject `json:"projects" yaml:"projects"`
}

type LxdCTemplateEngine struct {
	Engine string `json:"engine" yaml:"engine"`
	Opts   string `json:"opts,omitempty" yaml:"opts,omitempty"`
}

type LxdCProject struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	IncludeGroupDirs []string `json:"include_groups_dir,omitempty" yaml:"include_groups_dir,omitempty"`

	IncludeEnvFiles []string `json:"include_env_files,omitempty" yaml:"include_env_files,omitempty"`

	Environments []map[string]string `json:"envs,omitempty" yaml:"envs,omitempty"`

	Groups []LxdCGroup `json:"groups" yaml:"groups"`
}

type LxdCGroup struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Connection  string `json:"connection,omitempty" yaml:"connection,omitempty"`

	CommonProfiles []string `json:"common_profiles,omitempty" yaml:"common_profiles,omitempty"`
	Ephemeral      bool     `json:"ephemeral,omitempty" yaml:"ephemeral,omitempty"`

	ImageFetchOptions []string `json:"image_fetch_opts,omitempty" yaml:"image_fetch_opts,omitempty"`

	LxdCNode []LxdCNode `json:"nodes" yaml:"nodes"`
}

type LxdCNode struct {
	Name              string `json:"name" yaml:"name"`
	ImageSource       string `json:"image_source" yaml:"image_source"`
	ImageRemoteServer string `json:"image_remote_server,omitempty" yaml:"image_remote_server,omitempty"`

	Labels []string `json:"labels,omitempty" yaml:"labels,omitempty"`

	SourceDir string `json:"source_dir,omitempty" yaml:"source_dir,omitempty"`

	BootstrapCommand []string `json:"bootstrap_commands,omitempty" yaml:"bootstrap_commands,omitempty"`
	SyncPostCommands []string `json:"sync_post_commands,omitempty" yaml:"sync_post_commands,omitempty"`

	ConfigTemplates []LxdCConfigTemplate `json:"config_templates,omitempty" yaml:"config_templates,omitempty"`
	SyncResources   []LxdCSyncResource   `json:"sync_resources,omitempty" yaml:"sync_resources,omitempty"`
}

type LxdCConfigTemplate struct {
	Source      string `json:"source" yaml:"source"`
	Destination string `json:"dst" yaml:"dst"`
}

type LxdCSyncResource struct {
	Source      string `json:"source" yaml:"source"`
	Destination string `json:"dst" yaml:"dst"`
	Recursive   bool   `json:"recursive,omitempty" yaml:"recursive,omitempty"`
}
