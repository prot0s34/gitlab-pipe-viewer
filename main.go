// main.go
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/xanzy/go-gitlab"
)

var (
	gitlabClient *gitlab.Client
	token        string
)

func init() {
	token := os.Getenv("GITLAB_PERSONAL_TOKEN")
	if token == "" {
		fmt.Println("Please set GITLAB_PERSONAL_TOKEN environment variable.")
		os.Exit(1)
	}

	gitlabURL := os.Getenv("GITLAB_URL")
	if gitlabURL == "" {
		gitlabURL = "https://gitlab.com"
	}

	// Initialize GitLab client and handle errors
	var err error
	gitlabClient, err = gitlab.NewClient(token, gitlab.WithBaseURL(gitlabURL+"/api/v4"))
	if err != nil {
		fmt.Println("Error creating GitLab client:", err)
		os.Exit(1)
	}

	// Fetch and display some data using the gitlabClient variable
	groups, _, err := gitlabClient.Groups.ListGroups(&gitlab.ListGroupsOptions{})
	if err != nil {
		fmt.Println("Error fetching groups:", err)
		os.Exit(1)
	}

	for _, group := range groups {
		fmt.Println("Group:", group.Name)
	}
}

func main() {
	app := tview.NewApplication()

	if err := app.SetRoot(buildTree(app), true).Run(); err != nil {
		fmt.Println("Error:", err)
	}
}

func buildTree(app *tview.Application) *tview.TreeView {
    root := tview.NewTreeNode("GitLab Pipelines").
        SetColor(tcell.ColorYellow).
        SetSelectable(false)

    tree := tview.NewTreeView().
        SetRoot(root).
        SetCurrentNode(root).
        SetTopLevel(1).
        SetGraphicsColor(tcell.ColorGreen)

    tree.SetSelectedFunc(func(node *tview.TreeNode) {
        // Handle selection logic here
        projectName := node.GetText()
        if strings.HasPrefix(projectName, "Project: ") {
            showPipelines(app, node)
        }
    })

    root.AddChild(buildGroups())

    return tree
}

func buildGroups() *tview.TreeNode {
    root := tview.NewTreeNode("Groups").
        SetColor(tcell.ColorYellow)

    groups, _, err := gitlabClient.Groups.ListGroups(&gitlab.ListGroupsOptions{})
    if err != nil {
        fmt.Println("Error fetching groups:", err)
        return root
    }

    for _, group := range groups {
        groupNode := tview.NewTreeNode("Group: " + group.Name).
            SetColor(tcell.ColorWhite)
        root.AddChild(groupNode)

        projects, _, err := gitlabClient.Groups.ListGroupProjects(group.ID, &gitlab.ListGroupProjectsOptions{})
        if err != nil {
            fmt.Println("Error fetching projects for group", group.Name, ":", err)
            continue
        }

        for _, project := range projects {
            projectNode := tview.NewTreeNode("Project: " + project.Name).
                SetColor(tcell.ColorBlue).
                SetReference(fmt.Sprintf("%d", project.ID)) // Convert project ID to string
            groupNode.AddChild(projectNode)
        }
    }

    return root
}

func showPipelines(app *tview.Application, projectNode *tview.TreeNode) {
    // Extract the project ID from the reference
    projectID, ok := projectNode.GetReference().(string)
    if !ok {
        fmt.Println("Invalid project reference")
        return
    }

    // Fetch and display pipeline information for the selected project
    // Example: https://pkg.go.dev/github.com/xanzy/go-gitlab#PipelinesService.ListProjectPipelines

    projectPipelines, _, err := gitlabClient.Pipelines.ListProjectPipelines(projectID, &gitlab.ListProjectPipelinesOptions{})
    if err != nil {
        fmt.Println("Error fetching pipelines for project", projectID, ":", err)
        return
    }

    // Create a new tview.List to display pipeline information
    pipelineList := tview.NewList().ShowSecondaryText(false)

    for _, pipeline := range projectPipelines {
        // Format pipeline information as a string
        pipelineInfo := fmt.Sprintf("Pipeline ID: %d, Status: %s", pipeline.ID, pipeline.Status)
        // Add the pipeline information to the list
        pipelineList.AddItem(pipelineInfo, "", 0, nil)
    }

    // Set the selected function for the pipeline list
    pipelineList.SetSelectedFunc(func(index int, _ string, _ string, _ rune) {
        // Handle selection logic here if needed
    })

    // Create a new flex container to hold the list
    flex := tview.NewFlex().
        AddItem(pipelineList, 0, 1, false)

    // Set the root of the application to the flex container
    app.SetRoot(flex, true).SetFocus(pipelineList)
}
