package commands

import (
	"context"
	"fmt"

	"github.com/docker/ecs-plugin/pkg/amazon"
	"github.com/docker/ecs-plugin/pkg/compose"
	"github.com/docker/ecs-plugin/pkg/docker"
	"github.com/spf13/cobra"
)

func ComposeCommand(clusteropts *docker.AwsContext) *cobra.Command {
	cmd := &cobra.Command{
		Use: "compose",
	}
	opts := &compose.ProjectOptions{}
	opts.AddFlags(cmd.Flags())

	cmd.AddCommand(
		ConvertCommand(clusteropts, opts),
		UpCommand(clusteropts, opts),
		DownCommand(clusteropts, opts),
	)
	return cmd
}

type upOptions struct {
	loadBalancerArn string
}

func (o upOptions) LoadBalancerArn() *string {
	if o.loadBalancerArn == "" {
		return nil
	}
	return &o.loadBalancerArn
}

func ConvertCommand(clusteropts *docker.AwsContext, projectOpts *compose.ProjectOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use: "convert",
		RunE: compose.WithProject(projectOpts, func(project *compose.Project, args []string) error {
			client, err := amazon.NewClient(clusteropts.Profile, clusteropts.Cluster, clusteropts.Region)
			if err != nil {
				return err
			}
			template, err := client.Convert(context.Background(), project)
			if err != nil {
				return err
			}

			j, err := template.JSON()
			if err != nil {
				fmt.Printf("Failed to generate JSON: %s\n", err)
			} else {
				fmt.Printf("%s\n", string(j))
			}
			return nil
		}),
	}
	return cmd
}

func UpCommand(clusteropts *docker.AwsContext, projectOpts *compose.ProjectOptions) *cobra.Command {
	opts := upOptions{}
	cmd := &cobra.Command{
		Use: "up",
		RunE: compose.WithProject(projectOpts, func(project *compose.Project, args []string) error {
			client, err := amazon.NewClient(clusteropts.Profile, clusteropts.Cluster, clusteropts.Region)
			if err != nil {
				return err
			}
			return client.ComposeUp(context.Background(), project)
		}),
	}
	cmd.Flags().StringVar(&opts.loadBalancerArn, "load-balancer", "", "")
	return cmd
}

type downOptions struct {
	DeleteCluster bool
}

func DownCommand(clusteropts *docker.AwsContext, projectOpts *compose.ProjectOptions) *cobra.Command {
	opts := downOptions{}
	cmd := &cobra.Command{
		Use: "down",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := amazon.NewClient(clusteropts.Profile, clusteropts.Cluster, clusteropts.Region)
			if err != nil {
				return err
			}
			if len(args) == 0 {
				project, err := compose.ProjectFromOptions(projectOpts)
				if err != nil {
					return err
				}
				return client.ComposeDown(context.Background(), project.Name, opts.DeleteCluster)
			}
			// project names passed as parameters
			for _, name := range args {
				err := client.ComposeDown(context.Background(), name, opts.DeleteCluster)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&opts.DeleteCluster, "delete-cluster", false, "Delete cluster")
	return cmd
}
