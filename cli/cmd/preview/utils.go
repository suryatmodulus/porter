package preview

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/porter-dev/porter/api/types"
	"github.com/porter-dev/porter/cli/cmd/config"
)

type Source struct {
	Name          string
	Repo          string
	Version       string
	IsApplication bool
	SourceValues  map[string]interface{}
}

type Target struct {
	AppName   string
	Project   uint
	Cluster   uint
	Namespace string
}

func GetSource(input map[string]interface{}) (*Source, error) {
	output := &Source{}

	// first read from env vars
	output.Name = os.Getenv("PORTER_SOURCE_NAME")
	output.Repo = os.Getenv("PORTER_SOURCE_REPO")
	output.Version = os.Getenv("PORTER_SOURCE_VERSION")

	// next, check for values in the YAML file
	if output.Name == "" {
		if name, ok := input["name"]; ok {
			nameVal, ok := name.(string)
			if !ok {
				return nil, fmt.Errorf("invalid name provided")
			}
			output.Name = nameVal
		}
	}

	if output.Name == "" {
		return nil, fmt.Errorf("source name required")
	}

	if output.Repo == "" {
		if repo, ok := input["repo"]; ok {
			repoVal, ok := repo.(string)
			if !ok {
				return nil, fmt.Errorf("invalid repo provided")
			}
			output.Repo = repoVal
		}
	}

	if output.Version == "" {
		if version, ok := input["version"]; ok {
			versionVal, ok := version.(string)
			if !ok {
				return nil, fmt.Errorf("invalid version provided")
			}
			output.Version = versionVal
		}
	}

	// lastly, just put in the defaults
	if output.Version == "" {
		output.Version = "latest"
	}

	output.IsApplication = output.Repo == "https://charts.getporter.dev"

	if output.Repo == "" {
		output.Repo = "https://charts.getporter.dev"

		values, err := existsInRepo(output.Name, output.Version, output.Repo)

		if err == nil {
			// found in "https://charts.getporter.dev"
			output.SourceValues = values
			output.IsApplication = true
			return output, nil
		}

		output.Repo = "https://chart-addons.getporter.dev"

		values, err = existsInRepo(output.Name, output.Version, output.Repo)

		if err == nil {
			// found in https://chart-addons.getporter.dev
			output.SourceValues = values
			return output, nil
		}

		return nil, fmt.Errorf("source does not exist in any repo")
	} else {
		// we look in the passed-in repo
		values, err := existsInRepo(output.Name, output.Version, output.Repo)

		if err == nil {
			output.SourceValues = values
			return output, nil
		}
	}

	return nil, fmt.Errorf("source '%s' does not exist in repo '%s'", output.Name, output.Repo)
}

func GetTarget(input map[string]interface{}) (*Target, error) {
	output := &Target{}

	// first read from env vars
	if projectEnv := os.Getenv("PORTER_PROJECT"); projectEnv != "" {
		project, err := strconv.Atoi(projectEnv)
		if err != nil {
			return nil, err
		}
		output.Project = uint(project)
	}

	if clusterEnv := os.Getenv("PORTER_CLUSTER"); clusterEnv != "" {
		cluster, err := strconv.Atoi(clusterEnv)
		if err != nil {
			return nil, err
		}
		output.Cluster = uint(cluster)
	}

	output.Namespace = os.Getenv("PORTER_NAMESPACE")

	// next, check for values in the YAML file
	if output.Project == 0 {
		if project, ok := input["project"]; ok {
			projectVal, ok := project.(uint)
			if !ok {
				return nil, fmt.Errorf("project value must be an integer")
			}
			output.Project = projectVal
		}
	}

	if output.Cluster == 0 {
		if cluster, ok := input["cluster"]; ok {
			clusterVal, ok := cluster.(uint)
			if !ok {
				return nil, fmt.Errorf("cluster value must be an integer")
			}
			output.Cluster = clusterVal
		}
	}

	if output.Namespace == "" {
		if namespace, ok := input["namespace"]; ok {
			namespaceVal, ok := namespace.(string)
			if !ok {
				return nil, fmt.Errorf("invalid namespace provided")
			}
			output.Namespace = namespaceVal
		}
	}

	if appName, ok := input["app_name"]; ok {
		appNameVal, ok := appName.(string)
		if !ok {
			return nil, fmt.Errorf("invalid app_name provided")
		}
		output.AppName = appNameVal
	}

	// lastly, just put in the defaults
	if output.Project == 0 {
		output.Project = config.GetCLIConfig().Project
	}
	if output.Cluster == 0 {
		output.Cluster = config.GetCLIConfig().Cluster
	}
	if output.Namespace == "" {
		output.Namespace = "default"
	}

	return output, nil
}

func existsInRepo(name, version, url string) (map[string]interface{}, error) {
	chart, err := config.GetAPIClient().GetTemplate(
		context.Background(),
		name, version,
		&types.GetTemplateRequest{
			TemplateGetBaseRequest: types.TemplateGetBaseRequest{
				RepoURL: url,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return chart.Values, nil
}
